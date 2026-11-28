package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func (d *dashboard) mapRewardSettings() fyne.CanvasObject {
	selector := widget.NewSelect(nil, nil)
	selector.PlaceHolder = "Select map"
	selector.Disable()
	message := widget.NewLabel("Loading maps…")
	message.Wrapping = fyne.TextWrapWord
	save := widget.NewButton("Save map rates", nil)
	save.Importance = widget.HighImportance
	save.Disable()
	inherit := widget.NewButton("Use world defaults for all", nil)
	inherit.Disable()
	fields := []struct {
		name, key string
		input     *widget.Entry
	}{
		{"Monster EXP (%)", config.MobEXPKey, widget.NewEntry()},
		{"Monster loot chance (%)", config.MobDropKey, widget.NewEntry()},
		{"Monster gold payout (%)", config.MobGoldKey, widget.NewEntry()},
		{"Global event loot chance (%)", config.WorldDropKey, widget.NewEntry()},
		{"Quest EXP at turn-in (%)", config.QuestEXPKey, widget.NewEntry()},
		{"PvP rank kills (×)", config.PvPKillKey, widget.NewEntry()},
		{"Gold drop chance (%)", "", widget.NewEntry()},
	}
	form := widget.NewForm()
	for _, field := range fields {
		field.input.Disable()
		form.Append(field.name, field.input)
	}
	populate := func(override config.MapRewardOverrides) {
		values := [...]*uint32{override.MonsterEXP, override.MonsterLoot, override.MonsterGold, override.EventLoot, override.QuestEXP, override.PvPKills, override.GoldChance}
		for i, field := range fields {
			value := ""
			if values[i] != nil {
				value = strconv.FormatUint(uint64(*values[i]), 10)
			}
			field.input.SetText(value)
		}
	}
	mapsByLabel := map[string]uint16{}
	var overrides config.MapRewards
	selector.OnChanged = func(label string) {
		id, ok := mapsByLabel[label]
		if !ok {
			return
		}
		populate(overrides[id])
		for _, field := range fields {
			field.input.Enable()
		}
		inherit.Enable()
		save.Enable()
		message.SetText("Blank fields follow world defaults. Saved rates apply to the next kill.")
	}
	inherit.OnTapped = func() { populate(config.MapRewardOverrides{}) }
	save.OnTapped = func() {
		id, ok := mapsByLabel[selector.Selected]
		if !ok {
			return
		}
		parsed := [7]*uint32{}
		for i, field := range fields {
			raw := strings.TrimSpace(field.input.Text)
			if raw == "" {
				continue
			}
			number, err := strconv.ParseUint(raw, 10, 32)
			limit := uint64(10000)
			if i >= 5 {
				limit = 100
			}
			if err != nil || number > limit {
				message.SetText(fmt.Sprintf("%s must be blank or a whole number from 0 to %d.", field.name, limit))
				return
			}
			value := uint32(number)
			parsed[i] = &value
		}
		override := config.MapRewardOverrides{MonsterEXP: parsed[0], MonsterLoot: parsed[1], MonsterGold: parsed[2], EventLoot: parsed[3],
			QuestEXP: parsed[4], PvPKills: parsed[5], GoldChance: parsed[6]}
		selector.Disable()
		inherit.Disable()
		save.Disable()
		for _, field := range fields {
			field.input.Disable()
		}
		message.SetText("Saving map rates…")
		go func() {
			err := d.manager.SaveMapReward(d.ctx, id, override)
			fyne.Do(func() {
				if d.ctx.Err() != nil {
					return
				}
				selector.Enable()
				inherit.Enable()
				save.Enable()
				for _, field := range fields {
					field.input.Enable()
				}
				if err != nil {
					message.SetText(err.Error())
					return
				}
				if override.Empty() {
					delete(overrides, id)
				} else {
					overrides[id] = override
				}
				populate(override)
				message.SetText(fmt.Sprintf("Map %d rates saved.", id))
			})
		}()
	}
	go func() {
		ids, err := d.manager.RewardMaps()
		var stored config.MapRewards
		var defaults map[string]string
		if err == nil {
			stored, err = d.manager.MapRewards()
		}
		if err == nil {
			defaults, _, err = d.manager.RewardSettings()
		}
		fyne.Do(func() {
			if d.ctx.Err() != nil {
				return
			}
			if err != nil {
				message.SetText(err.Error())
				return
			}
			if len(ids) == 0 {
				message.SetText("No playable maps found in this installation.")
				return
			}
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
			labels := make([]string, 0, len(ids))
			for _, id := range ids {
				label := fmt.Sprintf("Map %d", id)
				mapsByLabel[label] = id
				labels = append(labels, label)
			}
			overrides = stored
			for _, field := range fields {
				value := defaults[field.key] + "%"
				if field.key == config.PvPKillKey {
					value = defaults[field.key] + "×"
				} else if field.key == "" {
					value = "30%"
				}
				field.input.SetPlaceHolder("Inherit world default (" + value + ")")
			}
			selector.SetOptions(labels)
			selector.Enable()
			message.SetText("Select a map to edit its rates.")
		})
	}()
	return container.NewVScroll(container.NewVBox(selector, form, message, container.NewHBox(save, inherit)))
}
