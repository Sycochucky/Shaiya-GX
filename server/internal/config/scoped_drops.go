package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/Spelunkern/Shaiya-GX/internal/worlddrops"
)

type ScopedDropRule struct {
	ID                 uint32
	Name               string
	Maps               []uint16
	MinLevel, MaxLevel uint16
	Chance             uint64
	Grade              uint16
	ItemType, ItemID   uint8
	DailyLimit         uint16
	Announce           bool
}

func (rule ScopedDropRule) Validate() error {
	if rule.ID == 0 || len(rule.Name) > 80 || rule.MinLevel == 0 || rule.MaxLevel < rule.MinLevel || rule.Chance > worlddrops.RollRange {
		return errors.New("enter a rule ID, a valid level range, and a chance from 0% to 100%")
	}
	if rule.Grade > 1024 || (rule.Grade != 0) == (rule.ItemType != 0) {
		return errors.New("specify one loot grade or one item type and ID")
	}
	if rule.Grade != 0 && rule.ItemID != 0 {
		return errors.New("item ID must be blank when a grade is selected")
	}
	if len(rule.Maps) > 100 {
		return errors.New("select up to 100 maps")
	}
	seen := map[uint16]bool{}
	for _, id := range rule.Maps {
		if seen[id] {
			return fmt.Errorf("duplicate map %d", id)
		}
		seen[id] = true
	}
	return nil
}

func (rule ScopedDropRule) Eligible(mapID, mobLevel uint16) bool {
	return (len(rule.Maps) == 0 || slices.Contains(rule.Maps, mapID)) &&
		mobLevel >= rule.MinLevel && mobLevel <= rule.MaxLevel
}

type ScopedDropEdit struct {
	Rule     *ScopedDropRule
	DeleteID uint32
}

func dropRulesPath(root string) string { return filepath.Join(root, ".local", "event-drop-rules.json") }

func ReadScopedDrops(root string) ([]ScopedDropRule, error) {
	raw, err := os.ReadFile(dropRulesPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rules []ScopedDropRule
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil, err
	}
	var previous uint32
	for _, rule := range rules {
		if rule.ID <= previous {
			return nil, errors.New("event drop rules must use unique, ordered IDs")
		}
		if err := rule.Validate(); err != nil {
			return nil, fmt.Errorf("event rule %d: %w", rule.ID, err)
		}
		previous = rule.ID
	}
	return rules, nil
}

func SaveScopedDrop(root string, edit ScopedDropEdit) ([]ScopedDropRule, error) {
	if (edit.Rule == nil) == (edit.DeleteID == 0) {
		return nil, errors.New("select exactly one drop rule action")
	}
	rules, err := ReadScopedDrops(root)
	if err != nil {
		return nil, err
	}
	id := edit.DeleteID
	if edit.Rule != nil {
		if err := edit.Rule.Validate(); err != nil {
			return nil, err
		}
		id = edit.Rule.ID
	}
	i, exists := slices.BinarySearchFunc(rules, id, func(rule ScopedDropRule, id uint32) int {
		switch {
		case rule.ID < id:
			return -1
		case rule.ID > id:
			return 1
		default:
			return 0
		}
	})
	switch {
	case edit.Rule == nil && exists:
		rules = slices.Delete(rules, i, i+1)
	case edit.Rule != nil && exists:
		rules[i] = *edit.Rule
	case edit.Rule != nil:
		if len(rules) >= 100 {
			return nil, errors.New("event drop rule limit reached")
		}
		rules = slices.Insert(rules, i, *edit.Rule)
	}
	raw, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return nil, err
	}
	return rules, writeRuntimeFile(dropRulesPath(root), string(raw)+"\n")
}
