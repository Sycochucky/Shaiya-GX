package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/internal/worlddrops"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func (d *dashboard) scopedDropSettings(ctx context.Context) fyne.CanvasObject {
	selectRule := widget.NewSelect(nil, nil)
	selectRule.PlaceHolder = "Load a saved rule"
	id, name, mapIDs := widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	mapIDs.SetPlaceHolder("All maps, or IDs separated by commas")
	lo, hi, chance := widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	grade, itemType, itemID := widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	limit := widget.NewEntry()
	limit.SetPlaceHolder("Unlimited")
	announce := widget.NewCheck("Announce rare drops in the affected map", nil)
	status := widget.NewLabel("Loading event drop rules…")
	status.Wrapping = fyne.TextWrapWord
	save := widget.NewButton("Save rule", nil)
	remove := widget.NewButton("Remove selected rule", nil)
	save.Disable()
	remove.Disable()
	rows := make(map[string]config.ScopedDropRule)
	var busy bool
	applyRows := func(rules []config.ScopedDropRule) {
		clear(rows)
		labels := make([]string, 0, len(rules))
		for _, rule := range rules {
			label := fmt.Sprintf("%d · %s", rule.ID, rule.Name)
			labels = append(labels, label)
			rows[label] = rule
		}
		selectRule.SetOptions(labels)
		save.Enable()
		setEnabled(remove, selectRule.Selected != "")
	}
	selectRule.OnChanged = func(label string) {
		rule, exists := rows[label]
		if !exists || busy {
			return
		}
		id.SetText(strconv.FormatUint(uint64(rule.ID), 10))
		name.SetText(rule.Name)
		mapIDs.SetText(rewardMapIDs(rule.Maps))
		lo.SetText(mapNumber(&rule.MinLevel))
		hi.SetText(mapNumber(&rule.MaxLevel))
		chance.SetText(worlddrops.FormatChance(rule.Chance))
		grade.SetText("")
		itemType.SetText("")
		itemID.SetText("")
		if rule.Grade != 0 {
			grade.SetText(mapNumber(&rule.Grade))
		} else {
			itemType.SetText(mapNumber(&rule.ItemType))
			itemID.SetText(mapNumber(&rule.ItemID))
		}
		limit.SetText("")
		if rule.DailyLimit != 0 {
			limit.SetText(mapNumber(&rule.DailyLimit))
		}
		announce.SetChecked(rule.Announce)
		remove.Enable()
	}
	submit := func(edit config.ScopedDropEdit) {
		busy = true
		save.Disable()
		remove.Disable()
		selectRule.Disable()
		status.SetText("Saving rule…")
		go func() {
			err := d.manager.SaveScopedDrop(ctx, edit)
			var rules []config.ScopedDropRule
			if err == nil {
				rules, err = d.manager.ScopedDropRules()
			}
			fyne.Do(func() {
				if ctx.Err() != nil {
					return
				}
				busy = false
				selectRule.Enable()
				save.Enable()
				if err != nil {
					setEnabled(remove, selectRule.Selected != "")
					status.SetText(err.Error())
					return
				}
				selectRule.ClearSelected()
				applyRows(rules)
				status.SetText("Saved. Rules apply to the next eligible kill.")
			})
		}()
	}
	save.OnTapped = func() {
		number, idErr := strconv.ParseUint(strings.TrimSpace(id.Text), 10, 32)
		minimum, minErr := strconv.ParseUint(strings.TrimSpace(lo.Text), 10, 16)
		maximum, maxErr := strconv.ParseUint(strings.TrimSpace(hi.Text), 10, 16)
		probability, chanceErr := worlddrops.ParseChance(chance.Text)
		lootGrade, gradeErr := parseMapNumber[uint16](grade.Text, 16)
		typ, typErr := parseMapNumber[uint8](itemType.Text, 8)
		item, itemErr := parseMapNumber[uint8](itemID.Text, 8)
		daily, limitErr := parseMapNumber[uint16](limit.Text, 16)
		if idErr != nil || minErr != nil || maxErr != nil || gradeErr != nil ||
			typErr != nil || itemErr != nil || limitErr != nil {
			status.SetText("Use whole numbers for IDs, levels, grade, item type and daily limit.")
			return
		}
		if chanceErr != nil {
			status.SetText(chanceErr.Error())
			return
		}
		maps, mapsErr := parseRewardMapIDs(mapIDs.Text)
		if mapsErr != nil {
			status.SetText(mapsErr.Error())
			return
		}
		rule := config.ScopedDropRule{ID: uint32(number), Name: strings.TrimSpace(name.Text), Maps: maps,
			MinLevel: uint16(minimum), MaxLevel: uint16(maximum), Chance: probability, Announce: announce.Checked}
		if lootGrade != nil {
			rule.Grade = *lootGrade
		}
		if typ != nil {
			rule.ItemType = *typ
		}
		if item != nil {
			rule.ItemID = *item
		} else if typ != nil {
			status.SetText("Enter the item ID as well as its type.")
			return
		}
		if daily != nil {
			rule.DailyLimit = *daily
		}
		if err := rule.Validate(); err != nil {
			status.SetText(err.Error())
			return
		}
		submit(config.ScopedDropEdit{Rule: &rule})
	}
	remove.OnTapped = func() {
		rule, exists := rows[selectRule.Selected]
		if exists {
			dialog.ShowConfirm("Remove event rule", fmt.Sprintf("Remove rule %d · %s?", rule.ID, rule.Name), func(ok bool) {
				if ok && !busy && ctx.Err() == nil {
					submit(config.ScopedDropEdit{DeleteID: rule.ID})
				}
			}, d.window)
		}
	}
	go func() {
		rules, err := d.manager.ScopedDropRules()
		fyne.Do(func() {
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				status.SetText(err.Error())
				return
			}
			applyRows(rules)
			status.SetText("These rules add to the installed event catalog. Daily limits reset at midnight UTC.")
		})
	}()
	form := widget.NewForm(widget.NewFormItem("Rule ID", id), widget.NewFormItem("Name", name),
		widget.NewFormItem("Map IDs", mapIDs), widget.NewFormItem("Minimum monster level", lo),
		widget.NewFormItem("Maximum monster level", hi), widget.NewFormItem("Base drop chance (%)", chance),
		widget.NewFormItem("Loot grade", grade), widget.NewFormItem("Item type", itemType),
		widget.NewFormItem("Item ID", itemID), widget.NewFormItem("Daily limit per account", limit))
	return container.NewVScroll(container.NewVBox(selectRule, form, announce, container.NewHBox(save, remove), status))
}
