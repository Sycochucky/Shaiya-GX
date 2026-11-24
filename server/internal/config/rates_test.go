package config

import (
	"math"
	"testing"
)

func TestRewardRatesPreserveDefaultAndValidateBounds(t *testing.T) {
	defaults, err := ParseRewardRates("", "", "")
	if err != nil || defaults != DefaultRewardRates() {
		t.Fatalf("defaults = %+v, %v", defaults, err)
	}
	rates, err := ParseRewardRates("250", "50", "4")
	if err != nil || ScaleEXP(13, rates.MobEXPPercent) != 32 || ScaleEXP(13, rates.QuestEXPPercent) != 6 || rates.PvPKills != 4 {
		t.Fatalf("custom rates = %+v, %v", rates, err)
	}
	for _, input := range [][3]string{{"10001", "", ""}, {"", "-1", ""}, {"", "", "0"}, {"", "", "101"}, {"2.5", "", ""}} {
		if _, err := ParseRewardRates(input[0], input[1], input[2]); err == nil {
			t.Fatalf("accepted invalid rate: %v", input)
		}
	}
	if got := ScaleEXP(math.MaxUint32, 10000); got != math.MaxUint32 {
		t.Fatalf("overflow scaled to %d", got)
	}
}
