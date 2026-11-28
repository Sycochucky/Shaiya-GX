package control

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func TestValidateNetworkSettings(t *testing.T) {
	valid := []NetworkSettings{
		{LoginListen: "127.0.0.1:30800", WorldListen: "127.0.0.1:30810", PublicIP: "127.0.0.1"},
		{LoginListen: "0.0.0.0:30800", WorldListen: "0.0.0.0:30810", PublicIP: "203.0.113.20"},
		{LoginListen: "192.168.1.20:30800", WorldListen: "192.168.1.20:30810", PublicIP: "192.168.1.20"},
	}
	for _, settings := range valid {
		if err := validateNetworkSettings(settings); err != nil {
			t.Fatalf("valid settings %+v: %v", settings, err)
		}
	}
	invalid := []NetworkSettings{
		{LoginListen: "localhost:30800", WorldListen: "127.0.0.1:30810", PublicIP: "127.0.0.1"},
		{LoginListen: "0.0.0.0:30800", WorldListen: "0.0.0.0:30810", PublicIP: "127.0.0.1"},
		{LoginListen: "127.0.0.1:30800", WorldListen: "127.0.0.1:30800", PublicIP: "127.0.0.1"},
		{LoginListen: "127.0.0.1:0", WorldListen: "127.0.0.1:30810", PublicIP: "127.0.0.1"},
	}
	for _, settings := range invalid {
		if err := validateNetworkSettings(settings); err == nil {
			t.Fatalf("invalid settings accepted: %+v", settings)
		}
	}
}

func TestSuggestPortsSkipsUnavailableAndKeepsPairDistinct(t *testing.T) {
	blocked := map[int]bool{30800: true, 30810: true, 30811: true}
	login, world, err := suggestPorts(func(port int) bool { return !blocked[port] })
	if err != nil {
		t.Fatal(err)
	}
	if login != 30801 || world != 30812 || login == world {
		t.Fatalf("ports = %d, %d", login, world)
	}
}

func TestSaveRuntimeSettingsPreservesUnrelatedValues(t *testing.T) {
	root := t.TempDir()
	path := config.RuntimeEnvPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	original := "# keep this\nRAPTURE_DB_PATH=runtime/custom.sqlite\nRAPTURE_LOGIN_LISTEN=127.0.0.1:30000\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	err := config.SaveRuntimeSettings(root, []config.RuntimeSetting{
		{Key: config.LoginListenKey, Value: "0.0.0.0:30800"},
		{Key: config.WorldListenKey, Value: "0.0.0.0:30810"},
		{Key: config.PublicIPKey, Value: "203.0.113.20"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"# keep this", "RAPTURE_DB_PATH=runtime/custom.sqlite", "RAPTURE_LOGIN_LISTEN=0.0.0.0:30800", "RAPTURE_WORLD_LISTEN=0.0.0.0:30810", "RAPTURE_PUBLIC_IP=203.0.113.20"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in %q", want, text)
		}
	}
}
