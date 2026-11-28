package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func (d *dashboard) rewardEvents(ctx context.Context) fyne.CanvasObject {
	preset := widget.NewSelect(nil, nil)
	preset.PlaceHolder = "Load saved preset"
	name := widget.NewEntry()
	maps := widget.NewEntry()
	maps.SetPlaceHolder("All maps, or IDs separated by commas")
	announce := widget.NewCheck("Announce start and end in affected maps", nil)
	exp, loot, gold, eventLoot := widget.NewEntry(), widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	inputs := []*widget.Entry{exp, loot, gold, eventLoot}
	for _, input := range inputs {
		input.SetPlaceHolder("Use the current map rate")
	}
	form := widget.NewForm(
		widget.NewFormItem("Name", name), widget.NewFormItem("Map IDs", maps),
		widget.NewFormItem("Monster EXP (%)", exp), widget.NewFormItem("Monster loot (%)", loot),
		widget.NewFormItem("Monster gold (%)", gold), widget.NewFormItem("Event loot (%)", eventLoot),
	)
	status := widget.NewLabel("Loading reward events…")
	status.Wrapping = fyne.TextWrapWord
	feedback := widget.NewLabel("")
	feedback.Wrapping = fyne.TextWrapWord
	minutes := widget.NewEntry()
	minutes.SetText("120")
	startAt := widget.NewEntry()
	startAt.SetPlaceHolder("Now, or YYYY-MM-DD HH:MM")
	scheduleForm := widget.NewForm(widget.NewFormItem("Duration (minutes)", minutes),
		widget.NewFormItem("Start (local time)", startAt))
	save := widget.NewButton("Save preset", nil)
	start := widget.NewButton("Schedule preset", nil)
	end := widget.NewButton("End or cancel", nil)
	var state config.RewardEvents
	ready, busy := false, true
	updateStatus := func() {
		if !ready {
			return
		}
		now := time.Now()
		active := state.Scheduled
		switch {
		case active == nil:
			status.SetText("No scheduled boost.")
		case active.Running(now):
			remaining := active.EndsAt.Sub(now).Round(time.Second)
			status.SetText(fmt.Sprintf("%s active · %s · %s left", active.Preset.Name, rewardScope(active.Preset.Maps), remaining))
		case now.Before(active.StartsAt):
			status.SetText(fmt.Sprintf("%s scheduled · %s · %s", active.Preset.Name, rewardScope(active.Preset.Maps), active.StartsAt.Local().Format("Jan 2 15:04")))
		default:
			status.SetText(active.Preset.Name + " ended. Map rates restored.")
		}
		setEnabled(save, !busy)
		setEnabled(start, !busy && preset.Selected != "")
		setEnabled(end, !busy && active != nil && now.Before(active.EndsAt))
	}
	refresh := func(updated config.RewardEvents) {
		state = updated
		names := make([]string, 0, len(state.Presets))
		for key := range state.Presets {
			names = append(names, key)
		}
		sort.Strings(names)
		preset.SetOptions(names)
		ready = true
		updateStatus()
	}
	preset.OnChanged = func(selected string) {
		if p, ok := state.Presets[selected]; ok {
			name.SetText(p.Name)
			maps.SetText(rewardMapIDs(p.Maps))
			announce.SetChecked(p.Announce)
			for i, rate := range [...]*uint32{p.Rates.MonsterEXP, p.Rates.MonsterLoot, p.Rates.MonsterGold, p.Rates.EventLoot} {
				text := ""
				if rate != nil {
					text = strconv.FormatUint(uint64(*rate), 10)
				}
				inputs[i].SetText(text)
			}
		}
		updateStatus()
	}
	take := func(edit config.RewardEventEdit) {
		busy = true
		updateStatus()
		feedback.SetText("Saving…")
		go func() {
			err := d.manager.SaveRewardEvent(ctx, edit)
			var updated config.RewardEvents
			if err == nil {
				updated, err = d.manager.RewardEvents()
			}
			fyne.Do(func() {
				if ctx.Err() != nil {
					return
				}
				busy = false
				if err != nil {
					feedback.SetText(err.Error())
					updateStatus()
					return
				}
				refresh(updated)
				if edit.Preset != nil {
					preset.SetSelected(edit.Preset.Name)
				}
				feedback.SetText("Saved.")
			})
		}()
	}
	save.OnTapped = func() {
		ids, err := parseRewardMapIDs(maps.Text)
		if err != nil {
			feedback.SetText(err.Error())
			return
		}
		rates := [4]*uint32{}
		for i, input := range inputs {
			if raw := strings.TrimSpace(input.Text); raw != "" {
				num, err := strconv.ParseUint(raw, 10, 32)
				if err != nil || num > 10000 {
					feedback.SetText("Enter whole percentages from 0 to 10000.")
					return
				}
				value := uint32(num)
				rates[i] = &value
			}
		}
		p := config.RewardPreset{Name: strings.TrimSpace(name.Text), Maps: ids, Announce: announce.Checked,
			Rates: config.MapRewardOverrides{MonsterEXP: rates[0], MonsterLoot: rates[1], MonsterGold: rates[2], EventLoot: rates[3]}}
		take(config.RewardEventEdit{Preset: &p})
	}
	start.OnTapped = func() {
		duration, err := strconv.Atoi(strings.TrimSpace(minutes.Text))
		if err != nil {
			feedback.SetText("Enter a whole number of minutes.")
			return
		}
		var at time.Time
		if raw := strings.TrimSpace(startAt.Text); raw != "" {
			at, err = time.ParseInLocation("2006-01-02 15:04", raw, time.Local)
			if err != nil {
				feedback.SetText("Use YYYY-MM-DD HH:MM for the start time.")
				return
			}
		}
		take(config.RewardEventEdit{StartPreset: preset.Selected, StartsAt: at, Minutes: duration})
	}
	end.OnTapped = func() { take(config.RewardEventEdit{End: true}) }
	save.Disable()
	start.Disable()
	end.Disable()
	go func() {
		initial, err := d.manager.RewardEvents()
		fyne.Do(func() {
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				status.SetText(err.Error())
				return
			}
			busy = false
			refresh(initial)
		})
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fyne.Do(func() {
					if ctx.Err() == nil {
						updateStatus()
					}
				})
			}
		}
	}()
	return container.NewVScroll(container.NewVBox(status, preset, form, announce, save, widget.NewSeparator(),
		scheduleForm, container.NewHBox(start, end), feedback))
}

func parseRewardMapIDs(raw string) ([]uint16, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var ids []uint16
	for part := range strings.SplitSeq(raw, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 16)
		if err != nil {
			return nil, fmt.Errorf("enter map IDs separated by commas")
		}
		ids = append(ids, uint16(id))
	}
	return ids, nil
}

func rewardMapIDs(ids []uint16) string {
	labels := make([]string, len(ids))
	for i, id := range ids {
		labels[i] = strconv.Itoa(int(id))
	}
	return strings.Join(labels, ", ")
}

func rewardScope(ids []uint16) string {
	if len(ids) == 0 {
		return "All maps"
	}
	return "Map IDs " + rewardMapIDs(ids)
}
