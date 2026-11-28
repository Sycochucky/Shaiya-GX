package config

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	MobEXPKey    = "RAPTURE_MOB_EXP_PERCENT"
	QuestEXPKey  = "RAPTURE_QUEST_EXP_PERCENT"
	PvPKillKey   = "RAPTURE_PVP_KILL_MULTIPLIER"
	MobDropKey   = "RAPTURE_MOB_DROP_PERCENT"
	MobGoldKey   = "RAPTURE_MOB_GOLD_PERCENT"
	WorldDropKey = "RAPTURE_WORLD_DROP_PERCENT"
)

type RewardRates struct {
	MobEXPPercent    uint32
	QuestEXPPercent  uint32
	PvPKills         uint32
	MobDropPercent   uint32
	MobGoldPercent   uint32
	MobGoldChance    uint32
	WorldDropPercent uint32
}

func DefaultRewardRates() RewardRates {
	return RewardRates{
		MobEXPPercent: 100, QuestEXPPercent: 100, PvPKills: 1,
		MobDropPercent: 100, MobGoldPercent: 100, MobGoldChance: 30, WorldDropPercent: 100,
	}
}

// RuntimeValue gives explicitly inherited values precedence over the local file.
func RuntimeValue(settings map[string]string, key string) string {
	if value, set := os.LookupEnv(key); set {
		return value
	}
	return settings[key]
}

// ParseRewardRates accepts whole percentages (100 = 1Ã—) and whole PvP multipliers.
func ParseRewardRates(settings map[string]string) (RewardRates, error) {
	rates := DefaultRewardRates()
	for _, field := range []struct {
		key      string
		min, max uint32
		target   *uint32
	}{
		{MobEXPKey, 0, 10000, &rates.MobEXPPercent},
		{QuestEXPKey, 0, 10000, &rates.QuestEXPPercent},
		{PvPKillKey, 1, 100, &rates.PvPKills},
		{MobDropKey, 0, 10000, &rates.MobDropPercent},
		{MobGoldKey, 0, 10000, &rates.MobGoldPercent},
		{WorldDropKey, 0, 10000, &rates.WorldDropPercent},
	} {
		text := settings[field.key]
		if text == "" {
			continue
		}
		value, err := strconv.ParseUint(text, 10, 32)
		if err != nil || uint32(value) < field.min || uint32(value) > field.max {
			return RewardRates{}, fmt.Errorf("%s must be a whole number from %d to %d", field.key, field.min, field.max)
		}
		*field.target = uint32(value)
	}
	return rates, nil
}

func ReadRewardRates(settings map[string]string) (RewardRates, error) {
	values := make(map[string]string, len(settings)+6)
	maps.Copy(values, settings)
	maps.Copy(values, InheritedRewardRates())
	return ParseRewardRates(values)
}

// ScaleEXP applies a percentage after native reward rounding, with saturation.
func ScaleEXP(amount, percent uint32) uint32 {
	return uint32(min(uint64(amount)*uint64(percent)/100, math.MaxUint32))
}

// InheritedRewardRates records the launch environment before loading runtime.env.
func InheritedRewardRates() map[string]string {
	values := make(map[string]string, 6)
	for _, key := range []string{MobEXPKey, QuestEXPKey, PvPKillKey, MobDropKey, MobGoldKey, WorldDropKey} {
		if value, ok := os.LookupEnv(key); ok {
			values[key] = value
		}
	}
	return values
}

// SaveRewardRates validates and saves local settings, honoring the receiving
// process's launch environment. The returned rates can immediately go live.
func SaveRewardRates(root string, settings, inherited map[string]string) (RewardRates, error) {
	values := make(map[string]string, len(settings)+len(inherited))
	maps.Copy(values, settings)
	maps.Copy(values, inherited)
	rates, err := ParseRewardRates(values)
	if err != nil {
		return RewardRates{}, err
	}
	return rates, saveRewardRates(root, rates, inherited)
}

func saveRewardRates(root string, rates RewardRates, inherited map[string]string) error {
	if _, err := ReadRuntimeEnv(root); err != nil {
		return err
	}
	path := RuntimeEnvPath(root)
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
		{MobEXPKey, rates.MobEXPPercent},
		{QuestEXPKey, rates.QuestEXPPercent},
		{PvPKillKey, rates.PvPKills},
		{MobDropKey, rates.MobDropPercent},
		{MobGoldKey, rates.MobGoldPercent},
		{WorldDropKey, rates.WorldDropPercent},
	} {
		if _, overridden := inherited[item.key]; overridden {
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
	return replaceFile(file.Name(), path)
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
