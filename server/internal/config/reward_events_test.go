package config

import (
	"testing"
	"time"
)

func TestRewardEventPersistenceAndExpiry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	now := time.Date(2026, 11, 26, 9, 0, 0, 0, time.UTC)
	gold := uint32(250)
	canUse := func(id uint16) bool { return id == 1 || id == 2 }
	preset := RewardPreset{Name: "Dungeon Night", Maps: []uint16{1}, Rates: MapRewardOverrides{MonsterGold: &gold}}
	if _, err := SaveRewardEvent(root, RewardEventEdit{Preset: &preset}, now, canUse); err != nil {
		t.Fatal(err)
	}
	start := now.Add(time.Hour)
	if _, err := SaveRewardEvent(root, RewardEventEdit{StartPreset: preset.Name, StartsAt: start, Minutes: 120}, now, canUse); err != nil {
		t.Fatal(err)
	}
	state, err := ReadRewardEvents(root)
	if err != nil {
		t.Fatal(err)
	}
	event := state.Scheduled
	if event == nil || event.Running(now) || !event.Running(start) ||
		event.Running(start.Add(2*time.Hour)) ||
		!event.Preset.Affects(1) || event.Preset.Affects(2) {
		t.Fatalf("unexpected schedule: %+v", event)
	}
	if _, err := SaveRewardEvent(root, RewardEventEdit{StartPreset: preset.Name, Minutes: 60}, now, canUse); err == nil {
		t.Fatal("overlapping event accepted")
	}
	if result, err := SaveRewardEvent(root, RewardEventEdit{End: true}, start, canUse); err != nil || result.Scheduled != nil || len(result.Presets) != 1 {
		t.Fatalf("end should retain the preset and clear the active event: %+v, %v", result, err)
	}
}
