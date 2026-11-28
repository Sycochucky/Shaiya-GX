package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
)

type RewardPreset struct {
	Name     string
	Maps     []uint16
	Rates    MapRewardOverrides
	Announce bool
}

func (p RewardPreset) Validate() error {
	if p.Name != strings.TrimSpace(p.Name) || p.Name == "" || len(p.Name) > 60 || strings.IndexFunc(p.Name, unicode.IsControl) >= 0 {
		return errors.New("enter a preset name from 1 to 60 bytes without control characters")
	}
	if p.Rates.Empty() {
		return errors.New("set at least one reward rate")
	}
	if len(p.Maps) > 100 {
		return errors.New("select up to 100 maps")
	}
	seen := make(map[uint16]bool, len(p.Maps))
	for _, id := range p.Maps {
		if seen[id] {
			return fmt.Errorf("duplicate map %d", id)
		}
		seen[id] = true
	}
	return p.Rates.Validate()
}

func (p RewardPreset) Affects(mapID uint16) bool {
	return len(p.Maps) == 0 || slices.Contains(p.Maps, mapID)
}

type TimedRewardEvent struct {
	Preset           RewardPreset
	StartsAt, EndsAt time.Time
}

func (e *TimedRewardEvent) Running(now time.Time) bool {
	return e != nil && !now.Before(e.StartsAt) && now.Before(e.EndsAt)
}

type RewardEvents struct {
	Presets   map[string]RewardPreset
	Scheduled *TimedRewardEvent
}

type RewardEventEdit struct {
	Preset      *RewardPreset
	StartPreset string
	StartsAt    time.Time
	Minutes     int
	End         bool
}

func rewardEventsPath(root string) string { return filepath.Join(root, ".local", "reward-events.json") }

func ReadRewardEvents(root string) (RewardEvents, error) {
	state := RewardEvents{Presets: map[string]RewardPreset{}}
	raw, err := os.ReadFile(rewardEventsPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return state, fmt.Errorf("read reward events: %w", err)
	}
	if state.Presets == nil {
		state.Presets = map[string]RewardPreset{}
	}
	for name, preset := range state.Presets {
		if name != preset.Name {
			return state, fmt.Errorf("preset name %q does not match its key", name)
		}
		if err := preset.Validate(); err != nil {
			return state, fmt.Errorf("preset %q: %w", name, err)
		}
	}
	if event := state.Scheduled; event != nil {
		if err := event.Preset.Validate(); err != nil {
			return state, fmt.Errorf("scheduled preset: %w", err)
		}
		if !event.EndsAt.After(event.StartsAt) || event.EndsAt.Sub(event.StartsAt) > 7*24*time.Hour {
			return state, errors.New("invalid reward event time range")
		}
	}
	return state, nil
}

func SaveRewardEvent(root string, edit RewardEventEdit, now time.Time, mapAvailable func(uint16) bool) (RewardEvents, error) {
	state, err := ReadRewardEvents(root)
	if err != nil {
		return state, err
	}
	validateMaps := func(preset RewardPreset) error {
		for _, id := range preset.Maps {
			if !mapAvailable(id) {
				return fmt.Errorf("map %d is unavailable", id)
			}
		}
		return nil
	}
	actions := 0
	if edit.Preset != nil {
		actions++
	}
	if edit.StartPreset != "" {
		actions++
	}
	if edit.End {
		actions++
	}
	if actions != 1 {
		return state, errors.New("select exactly one event action")
	}
	switch {
	case edit.Preset != nil:
		preset := *edit.Preset
		if err := preset.Validate(); err != nil {
			return state, err
		}
		if err := validateMaps(preset); err != nil {
			return state, err
		}
		if _, exists := state.Presets[preset.Name]; !exists && len(state.Presets) >= 100 {
			return state, errors.New("reward event preset limit reached")
		}
		state.Presets[preset.Name] = preset
	case edit.End:
		state.Scheduled = nil
	default:
		preset, ok := state.Presets[edit.StartPreset]
		if !ok {
			return state, errors.New("save the preset before scheduling it")
		}
		if err := validateMaps(preset); err != nil {
			return state, err
		}
		if state.Scheduled != nil && now.Before(state.Scheduled.EndsAt) {
			return state, errors.New("end or cancel the existing event before scheduling another")
		}
		start := edit.StartsAt
		if start.IsZero() {
			start = now
		}
		if start.Before(now.Add(-time.Minute)) || start.After(now.Add(365*24*time.Hour)) || edit.Minutes < 1 || edit.Minutes > 10080 {
			return state, errors.New("enter a start time within the next year and a duration of 1 to 10080 minutes")
		}
		state.Scheduled = &TimedRewardEvent{preset, start.UTC(), start.Add(time.Duration(edit.Minutes) * time.Minute).UTC()}
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return state, err
	}
	return state, writeRuntimeFile(rewardEventsPath(root), string(raw)+"\n")
}
