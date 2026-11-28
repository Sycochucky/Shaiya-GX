package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type MapTuning struct {
	InstanceMinutes, ReentryMinutes *uint16                `json:",omitempty"`
	PvPWindow                       *DailyWindow           `json:",omitempty"`
	PvPAnnouncements                bool                   `json:",omitempty"`
	BossWindows                     map[uint16]DailyWindow `json:",omitempty"`
	MinLevel, MaxLevel              *uint16                `json:",omitempty"`
	Faction                         *uint8                 `json:",omitempty"`
	MinimumParty                    *uint8                 `json:",omitempty"`
	Closed, Drain                   bool                   `json:",omitempty"`
	RespawnPercent                  *uint32                `json:",omitempty"`
	MobRespawnSeconds               map[uint16]uint32      `json:",omitempty"`
}

func (t MapTuning) AdmissionOverride() bool {
	return t.MinLevel != nil || t.MaxLevel != nil || t.Faction != nil || t.MinimumParty != nil || t.Closed
}

func (t MapTuning) empty() bool {
	return !t.AdmissionOverride() && !t.Drain && t.RespawnPercent == nil && len(t.MobRespawnSeconds) == 0 && len(t.BossWindows) == 0 && t.PvPWindow == nil && t.InstanceMinutes == nil && t.ReentryMinutes == nil
}

func (t MapTuning) Validate() error {
	if t.MinLevel != nil && t.MaxLevel != nil && *t.MaxLevel != 0 && *t.MinLevel > *t.MaxLevel {
		return errors.New("minimum level must be no higher than the maximum")
	}
	if t.Faction != nil && *t.Faction > 1 {
		return errors.New("faction must be Light or Fury")
	}
	if t.MinimumParty != nil && *t.MinimumParty > 30 {
		return errors.New("minimum party size must be 0 to 30")
	}
	if t.RespawnPercent != nil && (*t.RespawnPercent == 0 || *t.RespawnPercent > 1000) {
		return errors.New("respawn time must be 1 to 1000 percent")
	}
	if t.InstanceMinutes != nil && (*t.InstanceMinutes < 1 || *t.InstanceMinutes > 1440) {
		return errors.New("instance limit must be 1 to 1440 minutes")
	}
	if t.ReentryMinutes != nil && *t.ReentryMinutes > 1440 {
		return errors.New("re-entry delay must be 0 to 1440 minutes")
	}
	if t.PvPWindow != nil && !t.PvPWindow.Valid() {
		return errors.New("PvP needs different start and end times in HH:MM")
	}
	if t.PvPAnnouncements && t.PvPWindow == nil {
		return errors.New("choose a PvP window before enabling announcements")
	}
	if len(t.BossWindows) > 100 {
		return errors.New("select up to 100 boss respawn windows per map")
	}
	for id, window := range t.BossWindows {
		if id == 0 || !window.Valid() {
			return fmt.Errorf("monster %d needs different start and end times in HH:MM", id)
		}
	}
	if len(t.MobRespawnSeconds) > 1000 {
		return errors.New("select up to 1000 monster overrides per map")
	}
	for id, seconds := range t.MobRespawnSeconds {
		if id == 0 || seconds < 1 || seconds > 604800 {
			return fmt.Errorf("monster %d respawn time must be 1 to 604800 seconds", id)
		}
	}
	return nil
}

func tuningPath(root string) string { return filepath.Join(root, ".local", "map-controls.json") }

func ReadMapTuning(root string) (map[uint16]MapTuning, error) {
	raw, err := os.ReadFile(tuningPath(root))
	data := map[uint16]MapTuning{}
	if errors.Is(err, os.ErrNotExist) {
		return data, nil
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("map controls must contain a JSON object")
	}
	for id, t := range data {
		if err := t.Validate(); err != nil {
			return nil, fmt.Errorf("map %d: %w", id, err)
		}
	}
	return data, nil
}

func SaveMapTuning(root string, mapID uint16, value MapTuning) (map[uint16]MapTuning, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	all, err := ReadMapTuning(root)
	if err != nil {
		return nil, err
	}
	if value.empty() {
		delete(all, mapID)
	} else {
		all[mapID] = value
	}
	raw, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return nil, err
	}
	return all, writeRuntimeFile(tuningPath(root), string(raw)+"\n")
}
