package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func TestSaveRatesPreservesOtherSettingsAndEnvironmentPrecedence(t *testing.T) {
	for _, key := range []string{config.MobEXPKey, config.QuestEXPKey, config.PvPKillKey} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv(config.QuestEXPKey, "50")
	root := t.TempDir()
	manager := New(root)
	path := config.RuntimeEnvPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	original := "# Local overrides\r\nRAPTURE_LOGIN_LISTEN=127.0.0.1:39080\r\n" + config.QuestEXPKey + "=100\r\n" +
		config.MobEXPKey + "=150\r\n" + config.MobEXPKey + "=200"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	if err := manager.SaveRewardSettings(t.Context(), "300", "75", "3"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	want := "# Local overrides\r\nRAPTURE_LOGIN_LISTEN=127.0.0.1:39080\r\n" + config.QuestEXPKey + "=100\r\n" +
		config.MobEXPKey + "=150\r\n" + config.MobEXPKey + "=300\r\n" + config.PvPKillKey + "=3\r\n"
	if err != nil || string(data) != want {
		t.Fatalf("input file content was changed unexpectedly: %v", err)
	}
	values, overrides, err := manager.RewardSettings()
	if err != nil || values[config.MobEXPKey] != "300" || values[config.QuestEXPKey] != "50" || values[config.PvPKillKey] != "3" || !overrides[config.QuestEXPKey] {
		t.Fatalf("settings: %v %v, %v", values, overrides, err)
	}
	if err := manager.SaveRewardSettings(t.Context(), "invalid", "50", "1"); err == nil {
		t.Fatal("accepted invalid rate")
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != want {
		t.Fatalf("failed save changed the file: %v", err)
	}
}
