package config

import (
	"math"
	"testing"
)

func TestRewardRatesPreserveDefaultAndValidateBounds(t *testing.T) {
	defaults, err := ParseRewardRates(nil)
	if err != nil || defaults != DefaultRewardRates() || defaults.MobDropPercent != 100 ||
		defaults.MobGoldPercent != 100 || defaults.WorldDropPercent != 100 {
		t.Fatalf("defaults = %+v, %v", defaults, err)
	}
	rates, err := ParseRewardRates(map[string]string{
		MobEXPKey: "250", QuestEXPKey: "50", PvPKillKey: "4",
		MobDropKey: "200", MobGoldKey: "0", WorldDropKey: "300",
	})
	if err != nil || ScaleEXP(13, rates.MobEXPPercent) != 32 || ScaleEXP(13, rates.QuestEXPPercent) != 6 ||
		rates.PvPKills != 4 || rates.MobDropPercent != 200 || rates.MobGoldPercent != 0 || rates.WorldDropPercent != 300 {
		t.Fatalf("custom rates = %+v, %v", rates, err)
	}
	for _, input := range []struct{ key, value string }{
		{MobEXPKey, "10001"}, {QuestEXPKey, "-1"}, {PvPKillKey, "0"}, {PvPKillKey, "101"},
		{MobEXPKey, "2.5"}, {MobDropKey, "10001"}, {MobDropKey, "-1"},
		{MobGoldKey, "-1"}, {MobGoldKey, "10001"}, {WorldDropKey, "1.5"}, {WorldDropKey, "10001"},
	} {
		if _, err := ParseRewardRates(map[string]string{input.key: input.value}); err == nil {
			t.Fatalf("accepted invalid rate %s=%s", input.key, input.value)
		}
	}
	if got := ScaleEXP(math.MaxUint32, 10000); got != math.MaxUint32 {
		t.Fatalf("overflow scaled to %d", got)
	}
}
