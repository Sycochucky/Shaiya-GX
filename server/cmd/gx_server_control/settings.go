package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func (d *dashboard) settingsView() fyne.CanvasObject {
	values, overridden, err := d.manager.RewardSettings()
	if err != nil {
		return widget.NewLabel("Could not load settings: " + err.Error())
	}
	fields := []struct {
		label string
		key   string
		hint  string
		entry *widget.Entry
	}{
		{"Monster EXP (%)", config.MobEXPKey, "100 = 1×. Applies to solo, group and offline monster kill EXP.", widget.NewEntry()},
		{"Quest EXP (%)", config.QuestEXPKey, "100 = 1×. Applies to quest completion EXP.", widget.NewEntry()},
		{"PvP rank kills (×)", config.PvPKillKey, "1 = normal. Applies to rank and rank milestone progress.", widget.NewEntry()},
		{"Monster loot chance (%)", config.MobDropKey, "100 = normal. 200 = double the chance per monster loot slot.", widget.NewEntry()},
		{"Monster gold payout (%)", config.MobGoldKey, "Scales the gold amount before a party split. 0 turns off payouts.", widget.NewEntry()},
		{"Global event loot chance (%)", config.WorldDropKey, "Scales the chance for each eligible global drop rule.", widget.NewEntry()},
	}
	items := make([]*widget.FormItem, 0, len(fields))
	editable := false
	for _, field := range fields {
		field.entry.SetText(values[field.key])
		hint := field.hint
		if overridden[field.key] {
			field.entry.Disable()
			hint = "Set by the launch environment."
		} else {
			editable = true
		}
		items = append(items, &widget.FormItem{Text: field.label, Widget: field.entry, HintText: hint})
	}
	message := widget.NewLabel("Rates apply to new rewards immediately when World is running.")
	message.Wrapping = fyne.TextWrapWord
	save := widget.NewButton("Save rates", nil)
	save.Importance = widget.HighImportance
	setEnabled(save, editable)
	defaults := container.NewVScroll(container.NewPadded(container.NewVBox(
		widget.NewForm(items...), message, container.NewHBox(save),
	)))
	save.OnTapped = func() {
		save.Disable()
		message.SetText("Saving rates…")
		values := make(map[string]string, len(fields))
		for _, field := range fields {
			values[field.key] = strings.TrimSpace(field.entry.Text)
		}
		go func() {
			err := d.manager.SaveRewardSettings(d.ctx, values)
			if d.ctx.Err() != nil {
				return
			}
			fyne.Do(func() {
				save.Enable()
				if err != nil {
					message.SetText(err.Error())
					return
				}
				message.SetText("Rates saved.")
			})
		}()
	}
	return container.NewAppTabs(
		container.NewTabItem("World defaults", defaults),
		container.NewTabItem("Per-map rates", d.mapRewardSettings()),
		container.NewTabItem("Map controls", d.mapTuningSettings(d.ctx)),
		container.NewTabItem("Event presets", d.rewardEvents(d.ctx)),
		container.NewTabItem("Event drops", d.scopedDropSettings(d.ctx)),
		container.NewTabItem("Economy", d.economySettings(d.ctx)),
	)
}
