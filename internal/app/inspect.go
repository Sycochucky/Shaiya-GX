package app

import (
	"log"

	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

type inspectView struct {
	open         bool
	target       uint32
	deadline     float64
	items        [proto.InspectSlots]proto.InventoryItem
	shape        proto.UserShape
	preview      *scene.Character
	previewBuilt bool
	appearance   scene.Appearance
}

func (a *App) inspectTarget(id uint32) *game.Entity {
	if a.state == nil || id == 0 || id == a.state.SelfID {
		return nil
	}
	e := a.state.Entities[id]
	if e == nil || e.Kind != game.KindUser || e.Shape == nil || (e.IsGM || e.Concealed) && !a.state.Self.IsGM {
		return nil
	}
	if own, ok := a.selfSlot(); !ok || own.Race/2 != e.Shape.Race/2 {
		return nil
	}
	return e
}

func (a *App) inspectInRange(e *game.Entity) bool {
	if e == nil || a.wv == nil {
		return false
	}
	pos := a.wv.visualPosition()
	dx, dy, dz := pos[0]-e.X, pos[1]-e.Y, pos[2]-e.Z
	return dx*dx+dy*dy+dz*dz <= proto.InspectRange*proto.InspectRange
}

func (a *App) requestInspect(id uint32) {
	e := a.inspectTarget(id)
	if e == nil || a.inspectUI == nil || a.world == nil {
		return
	}
	if !a.inspectInRange(e) {
		a.systemMessage(17, 233, nil)
		return
	}
	if err := a.world.Send(proto.InspectPacket(id)); err != nil {
		a.backToLoginID(12168)
		return
	}
	a.inspect = inspectView{target: id, deadline: a.clock + 5}
}

func (a *App) confirmedInspect(p []byte) {
	v := &a.inspect
	if v.target == 0 || v.open || a.clock >= v.deadline {
		return
	}
	target, items, err := proto.DecodeInspectEquipment(p)
	if err != nil || target != v.target {
		return
	}
	e := a.inspectTarget(target)
	if !a.inspectInRange(e) {
		a.inspect = inspectView{}
		return
	}
	v.open, v.items, v.shape = true, items, *e.Shape
}

func (a *App) validateInspect() {
	v := &a.inspect
	if v.target != 0 && ((!v.open && a.clock >= v.deadline) || !a.inspectInRange(a.inspectTarget(v.target))) {
		if a.inspectUI != nil && a.windowDrag.position == &a.inspectUI.Position {
			a.finishWindowDrag()
		}
		a.inspect = inspectView{}
	}
}

func (a *App) updateInspect(in *platform.Input, w, h int) bool {
	if !a.inspect.open {
		return false
	}
	panel := a.inspectUI
	if panel.Update(in, w, h) {
		a.inspect = inspectView{}
		return true
	}
	x, y := panel.Origin(w, h)
	return panelPointer(in, x, y, ui.InspectWidth, ui.InspectHeight)
}

func (a *App) drawInspect(w, h int) {
	a.validateInspect()
	v := &a.inspect
	if !v.open {
		return
	}
	x, y := a.inspectUI.Origin(w, h)
	a.clearTooltipUnderPanel(x, y, ui.InspectWidth, ui.InspectHeight)
	a.inspectUI.Draw(a.sprites, a.font, w, h, v.shape.Name)

	if !v.previewBuilt {
		a.buildInspectPreview()
	}
	if v.preview != nil {
		a.drawEquipmentActor(v.preview, v.appearance, x+60, y+51, 125, 203, w, h)
	}
	for i, item := range v.items {
		if item.Type == 0 {
			continue
		}
		def := a.itemsItem(item.Type, item.TypeID)
		if def == nil {
			continue
		}
		cx, cy, size := ui.InspectCell(uint8(i))
		cx, cy = x+cx, y+cy
		a.drawItemIcon(def, cx, cy, size, white, &item)
		if a.lastMouseX >= cx && a.lastMouseX < cx+size && a.lastMouseY >= cy && a.lastMouseY < cy+size {
			a.tooltipContent = a.itemTooltip(def, &item, item.Count, "")
			a.tooltipComparison = tooltipContent{}
		}
	}
}

func (a *App) buildInspectPreview() {
	v := &a.inspect
	v.previewBuilt = true
	slot := proto.CharacterSlot{Race: v.shape.Race, Gender: v.shape.Gender, Class: v.shape.Class,
		Face: v.shape.Face, Hair: v.shape.Hair, Height: v.shape.Height, Cloak: v.items[7].Gems}
	for i := range v.shape.Equipment {
		slot.EquipType[i], slot.EquipTypeID[i] = v.items[i].Type, v.items[i].TypeID
	}
	ap := a.appearance(&slot)
	ap.WeaponEnchant = weaponEnchantFromCraft(v.items[5].Craft)
	ap.OffhandEnchant = weaponEnchantFromCraft(v.items[6].Craft)
	v.appearance = ap
	var err error
	v.preview, err = scene.BuildCharacter(a.assets, ap, "000_normal")
	if err != nil {
		log.Printf("inspect preview: %v", err)
	}
}
