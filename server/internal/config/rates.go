package config

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

const (
	MobEXPKey   = "RAPTURE_MOB_EXP_PERCENT"
	QuestEXPKey = "RAPTURE_QUEST_EXP_PERCENT"
	PvPKillKey  = "RAPTURE_PVP_KILL_MULTIPLIER"
)

type RewardRates struct {
	MobEXPPercent   uint32
	QuestEXPPercent uint32
	PvPKills        uint32
}

func DefaultRewardRates() RewardRates {
	return RewardRates{MobEXPPercent: 100, QuestEXPPercent: 100, PvPKills: 1}
}

// RuntimeValue gives explicitly inherited values precedence over the local file.
func RuntimeValue(settings map[string]string, key string) string {
	if value, set := os.LookupEnv(key); set {
		return value
	}
	return settings[key]
}

// ParseRewardRates accepts whole percentages (100 = 1×) and whole PvP multipliers.
func ParseRewardRates(mob, quest, kills string) (RewardRates, error) {
	rates := DefaultRewardRates()
	for _, field := range []struct {
		key, text string
		min, max  uint32
		target    *uint32
	}{
		{MobEXPKey, mob, 0, 10000, &rates.MobEXPPercent},
		{QuestEXPKey, quest, 0, 10000, &rates.QuestEXPPercent},
		{PvPKillKey, kills, 1, 100, &rates.PvPKills},
	} {
		if field.text == "" {
			continue
		}
		value, err := strconv.ParseUint(field.text, 10, 32)
		if err != nil || uint32(value) < field.min || uint32(value) > field.max {
			return RewardRates{}, fmt.Errorf("%s must be a whole number from %d to %d", field.key, field.min, field.max)
		}
		*field.target = uint32(value)
	}
	return rates, nil
}

func ReadRewardRates(settings map[string]string) (RewardRates, error) {
	return ParseRewardRates(RuntimeValue(settings, MobEXPKey), RuntimeValue(settings, QuestEXPKey), RuntimeValue(settings, PvPKillKey))
}

// ScaleEXP applies a percentage after native reward rounding, with saturation.
func ScaleEXP(amount, percent uint32) uint32 {
	return uint32(min(uint64(amount)*uint64(percent)/100, math.MaxUint32))
}
