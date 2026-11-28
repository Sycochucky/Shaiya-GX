package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// MapRewardOverrides contains absolute percentages. A nil field inherits the
// world rate. A pointer to zero switches the reward off.
type MapRewardOverrides struct {
	QuestEXP    *uint32 `json:"quest_exp,omitempty"`
	PvPKills    *uint32 `json:"pvp_kills,omitempty"`
	GoldChance  *uint32 `json:"gold_chance,omitempty"`
	MonsterEXP  *uint32 `json:"monster_exp,omitempty"`
	MonsterLoot *uint32 `json:"monster_loot,omitempty"`
	MonsterGold *uint32 `json:"monster_gold,omitempty"`
	EventLoot   *uint32 `json:"event_loot,omitempty"`
}

func (o MapRewardOverrides) Empty() bool {
	return o.MonsterEXP == nil && o.MonsterLoot == nil && o.MonsterGold == nil && o.EventLoot == nil &&
		o.QuestEXP == nil && o.PvPKills == nil && o.GoldChance == nil
}

func (o MapRewardOverrides) Validate() error {
	for _, field := range []struct {
		name  string
		value *uint32
	}{
		{"Quest EXP", o.QuestEXP},
		{"Monster EXP", o.MonsterEXP},
		{"Monster loot", o.MonsterLoot},
		{"Monster gold", o.MonsterGold},
		{"Global event loot", o.EventLoot},
	} {
		if field.value != nil && *field.value > 10000 {
			return fmt.Errorf("%s must be a whole percentage from 0 to 10000", field.name)
		}
	}
	if o.PvPKills != nil && *o.PvPKills > 100 {
		return errors.New("PvP rank kills must be a whole multiplier from 0 to 100")
	}
	if o.GoldChance != nil && *o.GoldChance > 100 {
		return errors.New("gold drop chance must be a whole percentage from 0 to 100")
	}
	return nil
}

func (o MapRewardOverrides) Apply(base RewardRates) RewardRates {
	pairs := [...]struct {
		override *uint32
		value    *uint32
	}{
		{o.QuestEXP, &base.QuestEXPPercent},
		{o.PvPKills, &base.PvPKills},
		{o.GoldChance, &base.MobGoldChance},
		{o.MonsterEXP, &base.MobEXPPercent},
		{o.MonsterLoot, &base.MobDropPercent},
		{o.MonsterGold, &base.MobGoldPercent},
		{o.EventLoot, &base.WorldDropPercent},
	}
	for _, pair := range pairs {
		if pair.override != nil {
			*pair.value = *pair.override
		}
	}
	return base
}

type MapRewards map[uint16]MapRewardOverrides

func mapRewardsPath(root string) string { return filepath.Join(root, ".local", "map-rewards.json") }

func ReadMapRewards(root string) (MapRewards, error) {
	raw, err := os.ReadFile(mapRewardsPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return MapRewards{}, nil
	}
	if err != nil {
		return nil, err
	}
	var rewards MapRewards
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&rewards); err != nil {
		return nil, fmt.Errorf("read map rewards: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF || rewards == nil {
		return nil, errors.New("map rewards must be one JSON object")
	}
	for id, override := range rewards {
		if err := override.Validate(); err != nil {
			return nil, fmt.Errorf("map %d: %w", id, err)
		}
	}
	return rewards, nil
}

// SaveMapReward changes one map and preserves settings for every other map.
// The caller serializes edits for this installation.
func SaveMapReward(root string, id uint16, override MapRewardOverrides) (MapRewards, error) {
	if err := override.Validate(); err != nil {
		return nil, err
	}
	rewards, err := ReadMapRewards(root)
	if err != nil {
		return nil, err
	}
	if override.Empty() {
		delete(rewards, id)
	} else {
		rewards[id] = override
	}
	raw, err := json.MarshalIndent(rewards, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := writeRuntimeFile(mapRewardsPath(root), string(raw)+"\n"); err != nil {
		return nil, err
	}
	return rewards, nil
}
