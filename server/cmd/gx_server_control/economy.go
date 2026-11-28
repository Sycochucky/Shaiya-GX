package main

import (
	"context"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
)

func (d *dashboard) economySettings(ctx context.Context) fyne.CanvasObject {
	fields := []*widget.Entry{widget.NewEntry(), widget.NewEntry(), widget.NewEntry(), widget.NewEntry(), widget.NewEntry()}
	status := widget.NewLabel("Loading economy settings…")
	status.Wrapping = fyne.TextWrapWord
	save := widget.NewButton("Save economy settings", nil)
	save.Disable()
	form := widget.NewForm(
		widget.NewFormItem("Weapon enchant chance scale (%)", fields[0]),
		widget.NewFormItem("Armor enchant chance scale (%)", fields[1]),
		widget.NewFormItem("NPC buy prices (%)", fields[2]),
		widget.NewFormItem("NPC sell prices (%)", fields[3]),
		widget.NewFormItem("Reroll fee (gold)", fields[4]),
	)
	save.OnTapped = func() {
		var values [5]uint32
		for i, field := range fields {
			v, err := strconv.ParseUint(strings.TrimSpace(field.Text), 10, 32)
			if err != nil {
				status.SetText("Enter a whole number in each field.")
				return
			}
			values[i] = uint32(v)
		}
		next := config.Economy{WeaponEnchantPercent: values[0], ArmorEnchantPercent: values[1],
			VendorBuyPercent: values[2], VendorSellPercent: values[3], RerollGold: values[4]}
		if err := next.Validate(); err != nil {
			status.SetText(err.Error())
			return
		}
		save.Disable()
		go func() {
			err := d.manager.SaveEconomy(ctx, next)
			fyne.Do(func() {
				if ctx.Err() != nil {
					return
				}
				save.Enable()
				if err != nil {
					status.SetText(err.Error())
					return
				}
				status.SetText("Saved. New quotes use these values.")
			})
		}()
	}
	go func() {
		current, err := d.manager.Economy()
		fyne.Do(func() {
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				status.SetText(err.Error())
				return
			}
			for i, value := range [...]uint32{current.WeaponEnchantPercent, current.ArmorEnchantPercent,
				current.VendorBuyPercent, current.VendorSellPercent, current.RerollGold} {
				fields[i].SetText(strconv.FormatUint(uint64(value), 10))
			}
			status.SetText("100% keeps the authored rates and prices. Custom prices and fees require the matched client. Guild-house buy discounts retain their own pricing.")
			save.Enable()
		})
	}()
	return container.NewVBox(form, save, status)
}
