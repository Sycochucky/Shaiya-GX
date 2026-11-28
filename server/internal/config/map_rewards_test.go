package config

import "testing"

func TestMapRewardOverridesRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	gold, event, disabled := uint32(250), uint32(200), uint32(0)
	base := DefaultRewardRates()
	if _, err := SaveMapReward(root, 1, MapRewardOverrides{MonsterGold: &gold, EventLoot: &event}); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveMapReward(root, 2, MapRewardOverrides{MonsterLoot: &disabled}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := ReadMapRewards(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded[1].Apply(base); got.MobGoldPercent != 250 || got.WorldDropPercent != 200 || got.MobEXPPercent != base.MobEXPPercent {
		t.Fatalf("Map 1 rates = %+v", got)
	}
	if got := reloaded[2].Apply(base); got.MobDropPercent != 0 || got.MobGoldPercent != base.MobGoldPercent || got.WorldDropPercent != base.WorldDropPercent {
		t.Fatalf("Map 2 rates = %+v", got)
	}
	if got := reloaded[3].Apply(base); got != base {
		t.Fatalf("Map without an override = %+v", got)
	}
	invalid := uint32(10001)
	if _, err := SaveMapReward(root, 2, MapRewardOverrides{MonsterLoot: &invalid}); err == nil {
		t.Fatal("out-of-range rate accepted")
	}
	reset, err := SaveMapReward(root, 1, MapRewardOverrides{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reset[1]; ok || reset[2].MonsterLoot == nil || *reset[2].MonsterLoot != 0 {
		t.Fatalf("reset changed more than map 1: %+v", reset)
	}
}
