package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func init() {
	if expected := os.Getenv("GX_TEST_HANG_IMAGE"); expected != "" {
		path, _ := os.Executable()
		if strings.EqualFold(filepath.Clean(path), expected) {
			time.Sleep(time.Minute)
			os.Exit(0)
		}
	}
}

func copyFixtureFile(t *testing.T, source, dest string) {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(dest, data, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeNativeFixtureEnv(t *testing.T, root string, overrides string) {
	t.Helper()
	a, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	path := config.RuntimeEnvPath(root)
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte(fmt.Sprintf("RAPTURE_LOGIN_LISTEN=%s\nRAPTURE_WORLD_LISTEN=%s\n%s", a.Addr(), b.Addr(), overrides)), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestUnreadyLaunchIsReaped(t *testing.T) {
	root := t.TempDir()
	writeNativeFixtureEnv(t, root, "")
	m := New(root)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	copyFixtureFile(t, self, m.binary(Login))
	t.Setenv("GX_TEST_HANG_IMAGE", m.binary(Login))
	m.startTimeout = 420 * time.Millisecond
	err = m.Apply(t.Context(), Start, []Service{Login}, nil)
	if err == nil || !strings.Contains(err.Error(), "did not report ready") {
		t.Fatalf("unexpected startup outcome: %v", err)
	}
	if status := New(root).Snapshot(t.Context())[Login]; status.State != Stopped || status.Processes != 0 {
		t.Fatalf("unready child remained: %+v", status)
	}
	t.Log("Timed-out child was reaped; fresh manager found no service.")
}

func TestIncompleteInstallationRefusesLaunchBeforeProgress(t *testing.T) {
	root := t.TempDir()
	writeNativeFixtureEnv(t, root, "")
	m := New(root)
	if m.CheckInstallation() == nil {
		t.Fatal("empty folder passed the dashboard installation check")
	}
	for _, action := range []Action{Start, Restart} {
		err := m.Apply(t.Context(), action, Services(), func(Service, Status) {
			t.Fatalf("%s advanced even though the package is incomplete", action)
		})
		if err == nil || !strings.Contains(err.Error(), "incomplete GX installation") || !strings.Contains(err.Error(), "gx_world") {
			t.Fatalf("%s error = %v", action, err)
		}
	}
}

func TestNativeLifecycleAndOperator(t *testing.T) {
	root := os.Getenv("GX_NATIVE_OPERATOR_FIXTURE")
	if root == "" {
		t.Skip("native binaries and disposable assets required")
	}
	ctx := t.Context()
	validate := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	writeNativeFixtureEnv(t, root, "")
	m := New(root)
	t.Cleanup(func() {
		if err := m.Apply(context.Background(), ForceStop, Services(), nil); err != nil {
			t.Error(err)
		}
	})
	act := func(manager *Manager, action Action, targets ...Service) {
		t.Helper()
		validate(manager.Apply(ctx, action, targets, nil))
	}
	ready := func(manager *Manager, service Service) int {
		t.Helper()
		state := manager.Snapshot(ctx)[service]
		if state.State != Ready || state.Processes != 1 {
			t.Fatalf("%s not ready: %+v", service, state)
		}
		return state.PID
	}
	stopped := func(manager *Manager, service Service) {
		t.Helper()
		state := manager.Snapshot(ctx)[service]
		if state.State != Stopped || state.Processes != 0 {
			t.Fatalf("%s still present: %+v", service, state)
		}
	}
	op := func(manager *Manager, req operatorwire.Request) operatorwire.Response {
		t.Helper()
		res, err := manager.Operator(ctx, req)
		validate(err)
		return res
	}
	rejected := func(manager *Manager, req operatorwire.Request) {
		t.Helper()
		if _, err := manager.Operator(ctx, req); err == nil {
			t.Fatalf("operator unexpectedly accepted %s", req.Action)
		}
	}
	started := make(chan struct{}, 1)
	done := make(chan error, 1)
	go func() {
		done <- m.Apply(ctx, Start, Services(), func(_ Service, state Status) {
			if state.State == Starting {
				select {
				case started <- struct{}{}:
				default:
				}
			}
		})
	}()
	<-started
	overlap := New(root).Apply(ctx, Start, Services(), nil)
	validate(<-done)
	if overlap != nil && !strings.Contains(overlap.Error(), "acquire installation control") {
		t.Fatal(overlap)
	}
	login, world := ready(m, Login), ready(m, World)
	m = New(root)
	act(m, Start, Services()...)
	if ready(m, Login) != login || ready(m, World) != world {
		t.Fatal("attach/idempotent start restarted a service")
	}
	op(m, operatorwire.Request{Action: "accounts"})
	rejected(m, operatorwire.Request{Action: "delete"})

	otherRoot := t.TempDir()
	other := New(otherRoot)
	copyFixtureFile(t, m.binary(Login), other.binary(Login))
	copyFixtureFile(t, filepath.Join(root, "build", exeName(Service("gx_console"))), filepath.Join(otherRoot, "build", exeName(Service("gx_console"))))
	copyFixtureFile(t, filepath.Join(root, "data", "rapture.sqlite"), filepath.Join(otherRoot, "data", "rapture.sqlite"))
	shared := filepath.Join(root, "runtime", "rapture.sqlite")
	writeNativeFixtureEnv(t, otherRoot, "RAPTURE_DB_PATH="+shared+"\n")
	rejected(other, operatorwire.Request{Action: "password", AccountID: 1})
	writeNativeFixtureEnv(t, otherRoot, "")
	act(other, Start, Login)
	t.Cleanup(func() {
		if err := other.Apply(context.Background(), ForceStop, []Service{Login}, nil); err != nil {
			t.Error(err)
		}
	})
	foreign := ready(other, Login)
	if err := terminateOwned(foreign, m.binary(Login)); err == nil {
		t.Fatal("foreign executable accepted for termination")
	}
	if ready(other, Login) != foreign {
		t.Fatal("foreign login was touched")
	}
	t.Log("Concurrent Start All, reattach, repeated Start All, live account lookup, invalid delete refusal, shared-runtime guard and foreign PID isolation passed.")

	act(m, Restart, World)
	next := ready(m, World)
	if next == world || ready(m, Login) != login {
		t.Fatal("World restart did not preserve Login")
	}
	world = next
	descriptor := filepath.Join(root, "runtime", "control", "gx_world.json")
	validate(terminateOwned(world, m.binary(World)))
	deadline := time.Now().Add(6 * time.Second)
	for m.Snapshot(ctx)[World].State != Stopped && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	stopped(New(root), World)
	if _, err := os.Stat(descriptor); err != nil {
		t.Fatalf("crash did not leave a stale descriptor: %v", err)
	}
	act(m, Start, Services()...)
	next = ready(m, World)
	if next == world || ready(m, Login) != login {
		t.Fatal("crash recovery duplicated Login or reused World")
	}
	validate(os.Rename(descriptor, descriptor+".test"))
	state := m.Snapshot(ctx)[World]
	if state.State != Unknown || state.Processes != 1 {
		t.Fatalf("missing descriptor status: %+v", state)
	}
	if err := m.Apply(ctx, Start, []Service{World}, nil); err == nil {
		t.Fatal("start accepted missing control identity")
	}
	if err := m.Apply(ctx, Stop, []Service{World}, nil); err == nil {
		t.Fatal("graceful stop silently accepted missing control identity")
	}
	if state := m.Snapshot(ctx)[World]; state.PID != next || state.Processes != 1 {
		t.Fatalf("unresponsive instance changed: %+v", state)
	}
	act(m, ForceStop, World)
	stopped(m, World)
	if ready(m, Login) != login || ready(other, Login) != foreign {
		t.Fatal("force stop targeted another service")
	}
	act(m, Stop, Services()...)
	act(m, Stop, Services()...)
	stopped(m, Login)
	stopped(m, World)
	if ready(other, Login) != foreign {
		t.Fatal("stop targeted foreign installation")
	}
	act(other, Stop, Login)
	t.Log("Restart, crash/stale descriptor recovery, unresponsive duplicate refusal, explicit force stop, and repeated Stop passed.")

	const name = "GXProbeOperator"
	op(m, operatorwire.Request{Action: "create", Name: name, Password: "initial test password"})
	rejected(m, operatorwire.Request{Action: "create", Name: name, Password: "do not replace existing test account"})
	list := op(m, operatorwire.Request{Action: "accounts", Query: name}).Accounts
	if len(list) != 1 {
		t.Fatalf("account search count %d", len(list))
	}
	account := list[0]
	uid := account.ID
	change := operatorwire.Request{Action: "account", AccountID: uid, GM: true, Points: 1150, ExpectedAccount: &account}
	op(m, change)
	rejected(m, change)
	found := op(m, operatorwire.Request{Action: "details", AccountID: uid}).Account
	if !found.GM || found.Points != 1150 {
		t.Fatal("GM and points did not persist")
	}
	rejected(m, operatorwire.Request{Action: "account", AccountID: uid, Points: 1 << 31, ExpectedAccount: found})
	op(m, operatorwire.Request{Action: "password", AccountID: uid, Password: "replacement test password"})

	t.Setenv("RAPTURE_DB_PATH", shared)
	t.Setenv("RAPTURE_DATA_DIR", filepath.Join(root, "data"))
	t.Setenv("RAPTURE_DB_SEED", filepath.Join(root, "data", "rapture.sqlite"))
	row, err := db.AccountByUID(int32(uid))
	validate(err)
	old, err := db.TryGameLogin(name, "initial test password", 1, "127.0.0.1")
	validate(err)
	updated, err := db.TryGameLogin(name, "replacement test password", 2, "127.0.0.1")
	validate(err)
	if old.Status == 0 || updated.Status != 0 || !row.Admin || row.Point != 1150 {
		t.Fatal("real login did not reflect new credentials, GM, and points")
	}
	char, err := db.CreateCharacter(row, db.NewCharacter{Name: "GXProbeChar", Mode: 1, Str: 10, Dex: 11, HP: 100, MP: 100, SP: 100})
	validate(err)
	validate(db.Close())

	req := operatorwire.Request{Action: "stats", AccountID: uid, CharacterID: char}
	current := op(m, req).Stats
	nextStats := *current
	nextStats.Strength = 312
	nextStats.Gold = 3_000_000_001
	req.Action = "save-stats"
	req.Stats = &nextStats
	req.ExpectedStats = current
	op(m, req)
	rejected(m, req)
	changed := op(m, operatorwire.Request{Action: "stats", AccountID: uid, CharacterID: char}).Stats
	if *changed != nextStats {
		t.Fatalf("character stats not round-tripped")
	}
	invalid := nextStats
	invalid.Strength = 32001
	rejected(m, operatorwire.Request{Action: "save-stats", AccountID: uid, CharacterID: char, Stats: &invalid, ExpectedStats: changed})
	req = operatorwire.Request{Action: "delete", AccountID: uid, CharacterID: char, ConfirmName: "confirm mismatch"}
	rejected(m, req)
	req.ConfirmName = "GXProbeChar"
	op(m, req)
	deleted := op(m, operatorwire.Request{Action: "details", AccountID: uid}).Characters
	if len(deleted) != 1 || !deleted[0].Deleted {
		t.Fatal("native soft deletion not visible")
	}
	req.Action = "restore"
	op(m, req)
	if got := op(m, operatorwire.Request{Action: "stats", AccountID: uid, CharacterID: char}).Stats; *got != nextStats {
		t.Fatal("restore lost saved character stats")
	}
	rejected(m, operatorwire.Request{Action: "stats", AccountID: uid + 1, CharacterID: char})
	t.Log("Account create, duplicate rejection, real password change, GM, points, character edits, stale/invalid values, delete/restore and ownership checks passed.")
	proof, _ := json.Marshal(map[string]any{"account": name, "character": "GXProbeChar", "account_id": uid, "character_id": char})
	_ = os.WriteFile(filepath.Join(root, "operator-test-proof.json"), proof, 0600)
}
