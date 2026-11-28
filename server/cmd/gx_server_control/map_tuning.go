package main

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func (d *dashboard) mapTuningSettings(ctx context.Context) fyne.CanvasObject {
	selector := widget.NewSelect(nil, nil)
	selector.PlaceHolder = "Select map"
	selector.Disable()
	stats := widget.NewLabel("Choose a map to view activity.")
	stats.Wrapping = fyne.TextWrapWord
	feedback := widget.NewLabel("")
	feedback.Wrapping = fyne.TextWrapWord
	minLevel, maxLevel, minimumParty, respawn := widget.NewEntry(), widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	instanceLimit, reentryDelay := widget.NewEntry(), widget.NewEntry()
	for _, input := range []*widget.Entry{minLevel, maxLevel, minimumParty, respawn, instanceLimit, reentryDelay} {
		input.SetPlaceHolder("Use map default")
	}
	mobs := widget.NewMultiLineEntry()
	mobs.SetPlaceHolder("Monster ID=seconds, one per line")
	bosses := widget.NewMultiLineEntry()
	bosses.SetPlaceHolder("Monster ID=HH:MM-HH:MM, server local time")
	faction := widget.NewSelect([]string{"Map default", "Light", "Fury"}, nil)
	faction.SetSelected("Map default")
	closed := widget.NewCheck("Close to new entries", nil)
	drain := widget.NewCheck("Drain party instances", nil)
	pvp := widget.NewEntry()
	pvp.SetPlaceHolder("Use map default, or HH:MM-HH:MM")
	pvpAnnounce := widget.NewCheck("Announce PvP window changes in this map", nil)
	save := widget.NewButton("Save map controls", nil)
	update := widget.NewButton("Refresh activity", nil)
	save.Disable()
	update.Disable()
	var tuning map[uint16]config.MapTuning
	idsByChoice := map[string]uint16{}
	activityBusy := false
	readActivity := func(label string, id uint16) {
		if activityBusy {
			return
		}
		activityBusy = true
		update.Disable()
		go func() {
			res, err := d.manager.MapActivity(ctx, id)
			fyne.Do(func() {
				activityBusy = false
				update.Enable()
				if ctx.Err() != nil || selector.Selected != label {
					return
				}
				if err != nil {
					stats.SetText(err.Error())
					return
				}
				text := fmt.Sprintf("Online %d · Instances %d\nSince World started · Monster kills %d · PvP defeats %d · Drops %d · Items collected %d · Gold gained %d",
					res.Players, res.Instances, res.MonsterKills, res.PvPKills, res.Drops, res.ItemsCollected, res.GoldGained)
				for _, boss := range res.BossRespawns {
					detail := "no respawn timer"
					if boss.Alive {
						detail = "spawned"
					} else if !boss.Next.IsZero() {
						detail = "eligible in " + max(time.Until(boss.Next), 0).Round(time.Second).String()
					}
					text += fmt.Sprintf("\nMonster %d · %s", boss.MonsterID, detail)
				}
				stats.SetText(text)
			})
		}()
	}
	update.OnTapped = func() {
		if id, ok := idsByChoice[selector.Selected]; ok {
			readActivity(selector.Selected, id)
		}
	}
	selector.OnChanged = func(label string) {
		id, ok := idsByChoice[label]
		if !ok {
			return
		}
		value := tuning[id]
		minLevel.SetText(mapNumber(value.MinLevel))
		maxLevel.SetText(mapNumber(value.MaxLevel))
		minimumParty.SetText(mapNumber(value.MinimumParty))
		respawn.SetText(mapNumber(value.RespawnPercent))
		instanceLimit.SetText(mapNumber(value.InstanceMinutes))
		reentryDelay.SetText(mapNumber(value.ReentryMinutes))
		f := "Map default"
		if value.Faction != nil {
			f = []string{"Light", "Fury"}[*value.Faction]
		}
		faction.SetSelected(f)
		closed.SetChecked(value.Closed)
		drain.SetChecked(value.Drain)
		pvp.SetText("")
		if w := value.PvPWindow; w != nil {
			pvp.SetText(fmt.Sprintf("%02d:%02d-%02d:%02d", w.StartMinute/60, w.StartMinute%60, w.EndMinute/60, w.EndMinute%60))
		}
		pvpAnnounce.SetChecked(value.PvPAnnouncements)
		lines := make([]string, 0, len(value.MobRespawnSeconds))
		for id, seconds := range value.MobRespawnSeconds {
			lines = append(lines, fmt.Sprintf("%d=%d", id, seconds))
		}
		slices.Sort(lines)
		mobs.SetText(strings.Join(lines, "\n"))
		lines = lines[:0]
		for id, window := range value.BossWindows {
			lines = append(lines, fmt.Sprintf("%d=%02d:%02d-%02d:%02d", id,
				window.StartMinute/60, window.StartMinute%60, window.EndMinute/60, window.EndMinute%60))
		}
		slices.Sort(lines)
		bosses.SetText(strings.Join(lines, "\n"))
		save.Enable()
		feedback.SetText("")
		readActivity(label, id)
	}
	save.OnTapped = func() {
		id, ok := idsByChoice[selector.Selected]
		if !ok {
			return
		}
		lo, loErr := parseMapNumber[uint16](minLevel.Text, 16)
		hi, hiErr := parseMapNumber[uint16](maxLevel.Text, 16)
		party, partyErr := parseMapNumber[uint8](minimumParty.Text, 8)
		delay, delayErr := parseMapNumber[uint32](respawn.Text, 32)
		limit, limitErr := parseMapNumber[uint16](instanceLimit.Text, 16)
		reentry, reentryErr := parseMapNumber[uint16](reentryDelay.Text, 16)
		if loErr != nil || hiErr != nil || partyErr != nil || delayErr != nil || limitErr != nil || reentryErr != nil {
			feedback.SetText("Enter whole numbers or leave fields blank for map defaults.")
			return
		}
		value := config.MapTuning{MinLevel: lo, MaxLevel: hi, MinimumParty: party, RespawnPercent: delay,
			Closed: closed.Checked, Drain: drain.Checked, InstanceMinutes: limit, ReentryMinutes: reentry, MobRespawnSeconds: map[uint16]uint32{}}
		switch faction.Selected {
		case "Light":
			v := uint8(0)
			value.Faction = &v
		case "Fury":
			v := uint8(1)
			value.Faction = &v
		}
		for line := range strings.SplitSeq(mobs.Text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			key, raw, found := strings.Cut(line, "=")
			monster, keyErr := strconv.ParseUint(strings.TrimSpace(key), 10, 16)
			seconds, valueErr := strconv.ParseUint(strings.TrimSpace(raw), 10, 32)
			_, duplicate := value.MobRespawnSeconds[uint16(monster)]
			if !found || keyErr != nil || valueErr != nil || duplicate {
				feedback.SetText("Enter one unique monster ID=seconds on each line.")
				return
			}
			value.MobRespawnSeconds[uint16(monster)] = uint32(seconds)
		}
		if raw := strings.TrimSpace(pvp.Text); raw != "" {
			window, err := parseDailyWindow(raw)
			if err != nil {
				feedback.SetText(err.Error())
				return
			}
			value.PvPWindow = &window
		}
		value.PvPAnnouncements = pvpAnnounce.Checked
		value.BossWindows = map[uint16]config.DailyWindow{}
		for line := range strings.SplitSeq(bosses.Text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			key, hours, found := strings.Cut(line, "=")
			mob, idErr := strconv.ParseUint(strings.TrimSpace(key), 10, 16)
			window, windowErr := parseDailyWindow(hours)
			_, duplicate := value.BossWindows[uint16(mob)]
			if !found || idErr != nil || windowErr != nil || duplicate {
				feedback.SetText("Use one unique monster ID=HH:MM-HH:MM per line for boss windows.")
				return
			}
			value.BossWindows[uint16(mob)] = window
		}
		selector.Disable()
		save.Disable()
		feedback.SetText("Saving map controls…")
		go func() {
			err := d.manager.SaveMapTuning(ctx, id, value)
			fyne.Do(func() {
				if ctx.Err() != nil {
					return
				}
				selector.Enable()
				save.Enable()
				if err != nil {
					feedback.SetText(err.Error())
					return
				}
				tuning[id] = value
				feedback.SetText("Saved. Entry rules apply to new arrivals and respawn changes to newly started timers.")
			})
		}()
	}
	go func() {
		ids, err := d.manager.RewardMaps()
		var saved map[uint16]config.MapTuning
		if err == nil {
			saved, err = d.manager.MapTuning()
		}
		fyne.Do(func() {
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				stats.SetText(err.Error())
				return
			}
			tuning = saved
			slices.Sort(ids)
			options := make([]string, 0, len(ids))
			for _, id := range ids {
				text := fmt.Sprintf("Map %d", id)
				options = append(options, text)
				idsByChoice[text] = id
			}
			selector.SetOptions(options)
			selector.Enable()
		})
	}()
	body := widget.NewForm(widget.NewFormItem("Minimum level", minLevel), widget.NewFormItem("Maximum level", maxLevel),
		widget.NewFormItem("Faction", faction), widget.NewFormItem("Minimum party", minimumParty),
		widget.NewFormItem("Party instance limit (minutes)", instanceLimit), widget.NewFormItem("Dungeon re-entry (minutes)", reentryDelay),
		widget.NewFormItem("Respawn time (%)", respawn), widget.NewFormItem("Monster respawn times", mobs),
		widget.NewFormItem("Boss respawn windows", bosses),
		widget.NewFormItem("PvP window (server local time)", pvp))
	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fyne.Do(func() {
					if ctx.Err() == nil && selector.Selected != "" && !activityBusy {
						update.OnTapped()
					}
				})
			}
		}
	}()
	return container.NewVScroll(container.NewVBox(selector, stats, update, widget.NewSeparator(),
		closed, drain, body, pvpAnnounce, save, feedback))
}

func mapNumber[T ~uint8 | ~uint16 | ~uint32](value *T) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*value), 10)
}

func parseMapNumber[T ~uint8 | ~uint16 | ~uint32](text string, bits int) (*T, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	number, err := strconv.ParseUint(text, 10, bits)
	if err != nil {
		return nil, err
	}
	value := T(number)
	return &value, nil
}

func parseDailyWindow(raw string) (config.DailyWindow, error) {
	left, right, ok := strings.Cut(strings.TrimSpace(raw), "-")
	start, startErr := time.Parse("15:04", left)
	end, endErr := time.Parse("15:04", right)
	result := config.DailyWindow{StartMinute: uint16(start.Hour()*60 + start.Minute()),
		EndMinute: uint16(end.Hour()*60 + end.Minute())}
	if !ok || startErr != nil || endErr != nil || !result.Valid() {
		return config.DailyWindow{}, fmt.Errorf("use HH:MM-HH:MM with different start and end times")
	}
	return result, nil
}
