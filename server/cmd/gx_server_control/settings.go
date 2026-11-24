package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func (d *dashboard) openSettings() {
	values, overridden, err := d.manager.RewardSettings()
	if err != nil {
		dialog.ShowError(err, d.window)
		return
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
	message := widget.NewLabel("Saved values apply the next time World starts.")
	message.Wrapping = fyne.TextWrapWord
	save := widget.NewButton("Save rates", nil)
	save.Importance = widget.HighImportance
	setEnabled(save, editable)
	settings := dialog.NewCustom("Server Settings", "Close",
		container.NewVBox(widget.NewForm(items...), message, container.NewHBox(save)), d.window)
	settings.Resize(fyne.NewSize(635, 350))
	save.OnTapped = func() {
		save.Disable()
		message.SetText("Saving rates…")
		mob, quest, kills := strings.TrimSpace(fields[0].entry.Text), strings.TrimSpace(fields[1].entry.Text), strings.TrimSpace(fields[2].entry.Text)
		go func() {
			err := d.manager.SaveRewardSettings(d.ctx, mob, quest, kills)
			if d.ctx.Err() != nil {
				return
			}
			fyne.Do(func() {
				save.Enable()
				if err != nil {
					message.SetText(err.Error())
					return
				}
				message.SetText("Rates saved. Restart World to apply them.")
			})
		}()
	}
	settings.Show()
}
