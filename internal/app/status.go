package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/attackspeed"
	"github.com/Spelunkern/Shaiya-GX/internal/movespeed"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/pvprank"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

type statusView struct {
	open, pending bool
	deadline      float64
}

func (a *App) toggleStatus() {
	a.playUISound("ui_window.ogg")
	a.status.open = !a.status.open
	if a.statusUI != nil {
		a.statusUI.Clear()
	}
	if a.status.open {
		a.skills.assign = nil
	}
}

func (a *App) statusState() ui.StatusState {
	c, _ := a.selfSlot()
	attack, move := "-", "-"
	if sp := a.state.Self.Speed; sp != nil {
		attack = a.msg.Get(attackspeed.Message(sp.Attack))
		move = a.msg.Get(movespeed.Message(sp.Move))
	}
	if v := a.state.Self.Vehicle; v.State >= proto.VehicleDriver && (a.state.Self.Speed == nil || a.state.Self.Speed.Move != 255) {
		id := 5279
		if v.Speed == 5 {
			id = 5280
		}
		move = a.msg.Get(id)
	}
	return ui.StatusState{Title: pvprank.Title(a.state.Details.Kills, a.msg), AttackSpeed: attack, MoveSpeed: move, Class: ui.ClassName(a.msg, a.state.Faction.Faction == 0, c.Class), Element: a.statusArmorElement(),
		Base: proto.StatPoints{c.Str, c.Dex, c.Rec, c.Int, c.Wis, c.Luc}, Points: a.state.Details.StatPoint, Details: a.state.Details,
		HP: a.state.Pools.HP, MP: a.state.Pools.MP, SP: a.state.Pools.SP, Stats: a.state.Stats, Known: a.state.StatsKnown, Busy: a.status.pending || a.state.Self.Dead}
}

func (a *App) updateStatus(in *platform.Input, w, h int) bool {
	for _, r := range in.Chars {
		if r == 't' || r == 'T' {
			a.toggleStatus()
			return true
		}
	}
	if !a.status.open || a.statusUI == nil {
		return false
	}
	x, y := a.statusUI.Origin(w, h)
	width := ui.StatusWidth
	if a.statusUI.Expanded {
		width += 132
	}
	captured := panelPointer(in, x, y, width, ui.StatusHeight)
	apply, close := a.statusUI.Update(in, w, h, a.statusState(), a.frameDT)
	if close || in.HasKey(platform.KeyEscape) {
		a.status.open = false
		a.statusUI.Clear()
		return true
	}
	if apply && !a.status.pending {
		d := a.statusUI.Draft
		if d.Total() > 0 && d.Total() <= uint32(a.state.Details.StatPoint) && !a.state.Self.Dead {
			if err := a.world.Send(proto.StatPointsPacket(d)); err != nil {
				a.backToLoginID(12193)
				return true
			}
			a.status.pending, a.status.deadline = true, a.clock+10
		}
		// Apply folds the sheet at once, sending only a non-empty draft (0x5037f0).
		a.statusUI.Clear()
	}
	return captured
}

func (a *App) drawStatus(w, h int) {
	if a.status.open && a.statusUI != nil {
		x, y := a.statusUI.Origin(w, h)
		width := ui.StatusWidth
		if a.statusUI.Expanded {
			width += 132
		}
		a.clearTooltipUnderPanel(x, y, width, ui.StatusHeight)
		a.statusUI.Draw(a.sprites, a.font, w, h, a.statusState())
		if description := a.msg.Get(a.statusUI.AttributeMessageAt(a.lastMouseX, a.lastMouseY, w, h)); description != "" {
			a.tooltipContent = tooltipContent{tooltip: []itemDetailLine{{description, white}}}
		}
	}
}

// The native Stat field describes the equipped upper armor, including its active lapis.
func (a *App) statusArmorElement() string {
	var value uint8
	if it, ok := a.state.Inventory[[2]uint8{0, 1}]; ok {
		value = a.effectiveItemElement(a.itemsItem(it.Type, it.TypeID), &it)
	}
	return elementName(value, a.msg)
}
