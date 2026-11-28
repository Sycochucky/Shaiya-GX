package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

type gameProbe struct {
	Directory string
	Arguments []string
}

func init() {
	if output := os.Getenv("GX_TEST_GAME_PROBE_OUTPUT"); output != "" {
		dir, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}
		data, err := json.Marshal(gameProbe{dir, os.Args[1:]})
		if err != nil || os.WriteFile(output, data, 0600) != nil {
			os.Exit(2)
		}
		os.Exit(0)
	}
}

func TestGameCommandStartsClientFromItsFolder(t *testing.T) {
	root := filepath.Join(t.TempDir(), "GX Client With Spaces")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	current, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	testImage, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "Shaiya GX.exe")
	if err := os.WriteFile(path, testImage, 0700); err != nil {
		t.Fatal(err)
	}
	status := map[control.Service]control.Status{
		control.Login: {State: control.Ready, Endpoint: "0.0.0.0:43912"},
		control.World: {State: control.Ready, Endpoint: "[::]:43925"},
	}
	output := filepath.Join(root, "invocation.json")
	cmd, err := gameCommand(path, status)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Env = append(os.Environ(), "GX_TEST_GAME_PROBE_OUTPUT="+output)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var got gameProbe
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	realDir, err := filepath.EvalSymlinks(got.Directory)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-login", "127.0.0.1:43912", "-world-port", "43925"}
	if realDir != realRoot || !reflect.DeepEqual(got.Arguments, want) {
		t.Fatalf("unexpected client invocation: %+v", got)
	}
	status[control.World] = control.Status{State: control.Stopped}
	if cmd, err := gameCommand(path, status); err == nil || cmd != nil {
		t.Fatal("constructed a game command while World was stopped")
	}
}

func TestGameReadinessWaitsForClientMarker(t *testing.T) {
	ready := make(chan struct{})
	startup := &gameStartupLog{ready: ready}
	_, _ = startup.Write([]byte("Graphics initialized\nClient ready frame"))
	select {
	case <-ready:
		t.Fatal("partial startup output signaled readiness")
	default:
	}
	_, _ = startup.Write([]byte("buffer=1024x768 vsync=true\n"))
	select {
	case <-ready:
	default:
		t.Fatal("readiness split across writes was missed")
	}
	_, _ = startup.Write([]byte("Client ready framebuffer=1024x768 vsync=true\n"))
}
