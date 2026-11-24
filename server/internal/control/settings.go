package control

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

// RewardSettings returns the editable local values and tracks inherited overrides.
func (m *Manager) RewardSettings() (map[string]string, map[string]bool, error) {
	values, err := config.ReadRuntimeEnv(m.root)
	if err != nil {
		return nil, nil, err
	}
	keys := []string{config.MobEXPKey, config.QuestEXPKey, config.PvPKillKey}
	defaults := config.DefaultRewardRates()
	fallbacks := []uint32{defaults.MobEXPPercent, defaults.QuestEXPPercent, defaults.PvPKills}
	effective := make(map[string]string, len(keys))
	override := make(map[string]bool, len(keys))
	for i, key := range keys {
		_, override[key] = os.LookupEnv(key)
		effective[key] = config.RuntimeValue(values, key)
		if effective[key] == "" {
			effective[key] = strconv.FormatUint(uint64(fallbacks[i]), 10)
		}
	}
	return effective, override, nil
}

func (m *Manager) SaveRewardSettings(ctx context.Context, mob, quest, kills string) error {
	rates, err := config.ParseRewardRates(mob, quest, kills)
	if err != nil {
		return err
	}
	release, err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer release()
	if _, err := config.ReadRuntimeEnv(m.root); err != nil {
		return err
	}
	path := config.RuntimeEnvPath(m.root)
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	content := string(raw)
	eol := "\n"
	if strings.Contains(content, "\r\n") {
		eol = "\r\n"
	}
	for _, item := range []struct {
		key string
		val uint32
	}{
		{config.MobEXPKey, rates.MobEXPPercent},
		{config.QuestEXPKey, rates.QuestEXPPercent},
		{config.PvPKillKey, rates.PvPKills},
	} {
		if _, overridden := os.LookupEnv(item.key); overridden {
			continue
		}
		content = replaceSetting(content, item.key, fmt.Sprint(item.val), eol)
	}
	if content == string(raw) {
		return nil
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".runtime-env-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.WriteString(content)
	if writeErr == nil {
		writeErr = file.Sync()
	}
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

// Update the last assignment, which is the one the runtime parser uses.
func replaceSetting(content, key, value, eol string) string {
	lines := strings.Split(content, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		name, _, found := strings.Cut(lines[i], "=")
		if found && strings.TrimSpace(name) == key {
			suffix := ""
			if strings.HasSuffix(lines[i], "\r") {
				suffix = "\r"
			}
			lines[i] = name + "=" + value + suffix
			return strings.Join(lines, "\n")
		}
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += eol
	}
	return content + key + "=" + value + eol
}
