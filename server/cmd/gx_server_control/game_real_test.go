package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	gxsession "github.com/Spelunkern/Shaiya-GX/internal/session"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func TestGamePathDetectsPortableClient(t *testing.T) {
	parent := t.TempDir()
	server := filepath.Join(parent, "Server")
	client := filepath.Join(parent, "Client")
	if err := os.MkdirAll(client, 0700); err != nil {
		t.Fatal(err)
	}
	name := "Shaiya GX"
	if filepath.Separator == '\\' {
		name += ".exe"
	}
	want := filepath.Join(client, name)
	if err := os.WriteFile(want, []byte("discovery fixture"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(gamePathFile(server)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gamePathFile(server), []byte(filepath.Join(parent, "missing", name)), 0600); err != nil {
		t.Fatal(err)
	}
	if got := gamePath(server); got != want {
		t.Fatalf("did not find the paired Client after a stale saved path: %q", got)
	}
}

func TestRunGameRealClientEntersDisposableWorld(t *testing.T) {
	root := os.Getenv("GX_REAL_GAME_SERVER_FIXTURE")
	if root == "" {
		t.Skip("requires a paired native Windows client and disposable server")
	}
	marker, err := os.ReadFile(filepath.Join(root, ".gx-disposable-test"))
	if err != nil || strings.TrimSpace(string(marker)) != "gx-run-game-disposable-v1" {
		t.Fatal("requires a runtime created specifically for the Run Game probe")
	}
	path := gamePath(root)
	if path == "" || filepath.Dir(path) != filepath.Join(filepath.Dir(root), "Client") {
		t.Fatalf("automatic discovery did not select the paired Client: %q", path)
	}
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	a, err := net.Listen("tcp", "127.0.0.1:0")
	check(err)
	b, err := net.Listen("tcp", "127.0.0.1:0")
	check(err)
	loginAddr, worldAddr := a.Addr().String(), b.Addr().String()
	worldPort := b.Addr().(*net.TCPAddr).Port
	a.Close()
	b.Close()
	if strings.HasSuffix(loginAddr, ":30800") || worldPort == 30810 {
		t.Fatal("reserved a default server port")
	}
	t.Setenv("GX_TLS", "0")
	t.Setenv("RAPTURE_PUBLIC_IP", "127.0.0.1")
	t.Setenv("RAPTURE_LOGIN_LISTEN", loginAddr)
	t.Setenv("RAPTURE_WORLD_LISTEN", worldAddr)
	t.Setenv("RAPTURE_DB_PATH", filepath.Join(root, "runtime", "rapture.sqlite"))
	t.Setenv("RAPTURE_DB_SEED", filepath.Join(root, "data", "rapture.sqlite"))
	t.Setenv("RAPTURE_DATA_DIR", filepath.Join(root, "data"))
	env := config.RuntimeEnvPath(root)
	check(os.MkdirAll(filepath.Dir(env), 0700))
	check(os.WriteFile(env, []byte(fmt.Sprintf("RAPTURE_LOGIN_LISTEN=%s\nRAPTURE_WORLD_LISTEN=%s\nRAPTURE_PUBLIC_IP=127.0.0.1\n", loginAddr, worldAddr)), 0600))
	m := control.New(root)
	ctx := t.Context()
	t.Cleanup(func() {
		if err := m.Apply(context.WithoutCancel(ctx), control.ForceStop, control.Services(), nil); err != nil {
			t.Error(err)
		}
	})
	check(m.Apply(ctx, control.Start, control.Services(), nil))
	status := m.Snapshot(ctx)
	for service, expected := range map[control.Service]string{control.Login: loginAddr, control.World: worldAddr} {
		if got := status[service]; got.State != control.Ready || got.Processes != 1 || got.PID == 0 || got.Endpoint != expected {
			t.Fatalf("unexpected disposable service state for %s: %+v", service, got)
		}
	}
	// Simulate reopening from both the original folder and an updated copy.
	reopened := control.New(root).Snapshot(ctx)
	if reopened[control.Login].PID != status[control.Login].PID || reopened[control.World].PID != status[control.World].PID {
		t.Fatal("reopen did not reattach to the original services")
	}
	updateCopy := control.New(t.TempDir())
	owner, err := updateCopy.ConflictingInstallation(ctx)
	check(err)
	if owner == nil || !strings.EqualFold(owner.Root(), root) {
		t.Fatal("a fresh dashboard did not discover the verified running installation")
	}
	if err := updateCopy.Apply(ctx, control.Start, control.Services(), nil); err == nil ||
		!strings.Contains(err.Error(), "already owns these ports") {
		t.Fatalf("an update copy attempted a duplicate start: %v", err)
	}
	for service, local := range updateCopy.Snapshot(ctx) {
		if local.Processes != 0 {
			t.Fatalf("update copy has an unexpected %s process: %+v", service, local)
		}
	}
	t.Log("Same-folder reopen recovered the original PIDs. A fresh copy discovered the owner and refused duplicate startup.")
	var secret [18]byte
	_, err = rand.Read(secret[:])
	check(err)
	password := hex.EncodeToString(secret[:])
	suffix := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
	user, character := "GXRun"+suffix, "GXWorld"+suffix
	_, err = m.Operator(ctx, operatorwire.Request{Action: "create", Name: user, Password: password})
	check(err)

	// Create a throwaway character via the actual login and world protocol.
	login, err := gxsession.Dial(loginAddr, user, password)
	check(err)
	defer login.Close()
	ip, err := login.SelectServer(login.Servers[0].ID, proto.ClientVersion)
	check(err)
	if !ip.IsLoopback() {
		t.Fatal("disposable Login advertised a non-loopback address")
	}
	world, err := login.DialWorld(ip, worldPort)
	check(err)
	defer world.Close()
	wait := func(op proto.Opcode) []byte {
		t.Helper()
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		for {
			select {
			case packet, ok := <-world.Packets:
				if !ok {
					t.Fatalf("fixture character setup disconnected: %v", world.Err())
				}
				if proto.Op(packet) == op {
					return packet
				}
			case <-timer.C:
				t.Fatalf("fixture character setup timed out on %04x", op)
			}
		}
	}
	wait(proto.AccountFaction)
	check(world.Send(proto.FactionPacket(0)))
	wait(proto.AccountFaction)
	check(world.Send(proto.CreateCharacterPacket(proto.NewCharacter{Name: character, Hair: 1, Face: 1, Height: 1})))
	reply, err := proto.DecodeU8(wait(proto.CreateCharacter))
	check(err)
	if reply != 0 {
		t.Fatalf("fixture character creation returned %d", reply)
	}
	check(world.Shutdown(10 * time.Second))

	cmd, err := gameCommand(path, m.Snapshot(ctx))
	check(err)
	screenshot := filepath.Join(filepath.Dir(root), "run-game-world-"+suffix+".png")
	if _, err := os.Stat(screenshot); err == nil {
		t.Fatalf("choose a fresh screenshot path: %s", screenshot)
	}
	// The game's own offscreen mode adds automation only. Use the exact
	// binary, directory, Login address, and World port from Run Game.
	cmd.Args = append(cmd.Args, "-autoplay", user, "-screenshot", screenshot, "-screenshot-screen", "world")
	profile := t.TempDir()
	appData, localData := filepath.Join(profile, "AppData"), filepath.Join(profile, "LocalAppData")
	check(os.Mkdir(appData, 0700))
	check(os.Mkdir(localData, 0700))
	cmd.Env = append(os.Environ(), "SHAIYA_PASSWORD="+password, "APPDATA="+appData, "LOCALAPPDATA="+localData)
	check(cmd.Start())
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			path := filepath.Join(cmd.Dir, "logs", "client.log")
			log, _ := os.ReadFile(path)
			note := strings.ReplaceAll(string(log[max(0, len(log)-2800):]), password, "[redacted]")
			t.Fatalf("native graphical client exited: %v\n%s", err, note)
		}
	case <-time.After(100 * time.Second):
		cmd.Process.Kill()
		<-done
		t.Fatal("graphical client did not complete its World capture")
	}
	image, err := os.Stat(screenshot)
	check(err)
	if image.Size() < 1024 {
		t.Fatal("World capture was empty")
	}
	worldlog, err := os.ReadFile(filepath.Join(root, "logs", "gx_world.log"))
	check(err)
	if !strings.Contains(string(worldlog), "/"+character) || !strings.Contains(string(worldlog), "selected character id=") {
		t.Fatal("disposable World did not log the throwaway character")
	}
	after := m.Snapshot(ctx)
	if after[control.Login].PID != status[control.Login].PID || after[control.World].PID != status[control.World].PID {
		t.Fatal("Run Game changed a service process")
	}
	t.Logf("Real Shaiya GX entered disposable World. Login %s, World %s, PIDs %d/%d, capture %d bytes.",
		loginAddr, worldAddr, status[control.Login].PID, status[control.World].PID, image.Size())
	check(m.Apply(ctx, control.Stop, control.Services(), nil))
	for service, info := range m.Snapshot(ctx) {
		if info.Processes != 0 || info.State != control.Stopped {
			t.Fatalf("disposable %s remained active: %+v", service, info)
		}
	}
}
