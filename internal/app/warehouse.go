package app

import (
	"fmt"

	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

type warehouseView struct {
	npcID                       uint32
	page                        int
	source, hover               uint8
	selected, dragging, hovered bool
	guild, guildPending         bool
	sourceMove, destMove        [2]uint8
}

func (a *App) warehouseBag() uint8 {
	if a.bank.guild {
		return proto.GuildBankBag
	}
	return proto.WarehouseBag
}

func (a *App) warehouseItems() map[uint8]proto.InventoryItem {
	if a.bank.guild {
		return a.state.Guild.Bank
	}
	return a.state.Warehouse
}

func (a *App) guildBankAllowed(source, destination [2]uint8) bool {
	g := &a.state.Guild
	if g.ID == 0 || g.Rank == 0 || source == destination {
		return false
	}
	if source[0] == proto.GuildBankBag {
		if g.Rank > 2 {
			a.questMessageID(12206)
			return false
		}
		item, ok := g.Bank[source[1]]
		if !ok {
			return false
		}
		if source[1] >= 40 || destination[0] == proto.GuildBankBag && destination[1] >= 40 {
			a.questMessageID(12207)
			return false
		}
		if destination[0] != proto.GuildBankBag && item.Type != 94 {
			def := a.itemsItem(item.Type, item.TypeID)
			if def == nil {
				return false
			}
			fee := uint32(float32(uint32(item.Count)*def.SellPrice) * .05)
			if a.state.Details.Money < fee {
				a.questMessageID(12208)
				return false
			}
		}
		return true
	}
	if destination[0] != proto.GuildBankBag || source[0] == 0 {
		return false
	}
	if g.Rank == 9 {
		a.questMessageID(12209)
		return false
	}
	if _, occupied := g.Bank[destination[1]]; occupied && g.Rank > 2 {
		a.questMessageID(12210)
		return false
	}
	item, ok := a.state.Inventory[source]
	if !ok {
		return false
	}
	if item.Type == 27 || item.Type == 28 || item.Type == 29 {
		a.questMessageID(12211)
		return false
	}
	def := a.itemsItem(item.Type, item.TypeID)
	if def == nil || def.ReqOG != 0 {
		a.questMessageID(12212)
		return false
	}
	return true
}

// warehouseCallSkill is the Warehouse Call buff (ability 39): while it lasts
// the warehouse can be opened from anywhere, and it cannot be cancelled.
const warehouseCallSkill = 228

const warehouseCallDescription = "Opens your personal warehouse without a banker. Click the Warehouse Call buff to reopen it while the effect lasts."

func (a *App) hasWarehouseCall() bool {
	if a.state == nil {
		return false
	}
	for _, b := range a.state.Buffs {
		if b.Skill == warehouseCallSkill {
			return true
		}
	}
	return false
}

// remoteWarehouseOpen: the buff's window, shown while no NPC is being talked to.
func (a *App) remoteWarehouseOpen() bool {
	if a.remoteBank && (!a.hasWarehouseCall() || a.shop.npcID != 0 || a.state.Self.Dead) {
		a.remoteBank = false
	}
	return a.remoteBank && a.bankUI != nil
}

// openRemoteWarehouse shows the window from the buff, with the bag beside it.
func (a *App) openRemoteWarehouse() {
	if a.bankUI == nil || !a.hasWarehouseCall() || a.shop.npcID != 0 || a.state.Self.Dead {
		return
	}
	if !a.remoteBank {
		a.bank = warehouseView{}
		a.inventory.servicePosition = ui.WindowPosition{}
	}
	a.remoteBank = true
	if !a.inventory.open {
		a.toggleInventory()
	}
}

func (a *App) warehouseOpen() bool {
	if a.remoteWarehouseOpen() {
		return true
	}
	e := a.state.Entities[a.shop.npcID]
	return a.bankUI != nil && e != nil && e.NpcType == 6 && !a.shop.overlay() && !a.state.Self.Dead
}
func (a *App) closeWarehouse() {
	a.bank.selected, a.bank.dragging = false, false
	if a.remoteBank {
		a.remoteBank = false // no NPC conversation to end
		return
	}
	if err := a.world.Send(proto.NpcTalkPacket(0)); err != nil {
		a.backToLoginID(12204)
		return
	}
	a.shop.npcID = 0
}
func (a *App) updateWarehouse(in *platform.Input, w, h int) bool {
	v := &a.bank
	if v.npcID != a.shop.npcID {
		*v = warehouseView{npcID: a.shop.npcID}
		if !a.inventory.open {
			a.toggleInventory()
		}
	}
	x, y := a.bankUI.Origin(w, h)
	captured := panelPointer(in, x, y, ui.WarehouseWidth, ui.WarehouseHeight)
	// The bank NPC of a house map is the guild warehouse and every other one the
	// personal storage: there is no selector between the two.
	v.guild = !a.remoteBank && houseMaps[a.state.MapID]
	page, close := a.bankUI.Update(in, w, h, v.page)
	v.page = page
	if close || in.HasKey(platform.KeyEscape) {
		a.closeWarehouse()
		return true
	}
	v.hover, v.hovered = ui.WarehouseHit(in.MouseX-x, in.MouseY-y, v.page)
	if (!v.guild && !a.state.WarehouseReady) || a.inventory.pending || a.shop.pending {
		return captured
	}
	if in.Pressed && v.hovered && !v.selected {
		if _, ok := a.warehouseItems()[v.hover]; ok {
			v.source, v.selected, v.dragging = v.hover, true, true
		}
	}
	if in.Clicked {
		wasDrag := v.dragging
		v.dragging = false
		if tab := ui.WarehouseTab(in.MouseX-x, in.MouseY-y); wasDrag && v.selected && tab >= 0 {
			if tab != int(v.source)/ui.WarehousePageSize {
				if slot, ok := firstFreeWarehouseSlot(a.warehouseItems(), tab); ok {
					bag := a.warehouseBag()
					v.page = tab
					a.moveWarehouseItem([2]uint8{bag, v.source}, [2]uint8{bag, slot})
				} else {
					a.questMessageID(12213)
				}
			}
			v.selected = false
			return true
		}
		if v.hovered && v.selected {
			if v.source != v.hover {
				bag := a.warehouseBag()
				a.moveWarehouseItem([2]uint8{bag, v.source}, [2]uint8{bag, v.hover})
				v.selected = false
			} else if !wasDrag {
				v.selected = false
			}
		} else if wasDrag {
			v.selected = false
		}
	}
	return captured
}
func firstFreeWarehouseSlot(items map[uint8]proto.InventoryItem, page int) (uint8, bool) {
	for i := range ui.WarehousePageSize {
		slot := uint8(page*ui.WarehousePageSize + i)
		if _, occupied := items[slot]; !occupied {
			return slot, true
		}
	}
	return 0, false
}

func (a *App) moveWarehouseItem(source, destination [2]uint8) {
	if !a.warehouseOpen() || (!a.bank.guild && !a.state.WarehouseReady) || a.inventory.pending || a.shop.pending {
		return
	}
	if a.bank.guild {
		if !a.guildBankAllowed(source, destination) {
			return
		}
		a.bank.guildPending = true
		a.bank.sourceMove, a.bank.destMove = source, destination
	}
	if err := a.world.Send(proto.ItemMovePacket(source[0], source[1], destination[0], destination[1])); err != nil {
		a.backToLoginID(12205)
		return
	}
	a.inventory.pending, a.inventory.deadline = true, a.clock+10
	a.inventory.selected, a.inventory.dragging = false, false
	a.bank.selected, a.bank.dragging = false, false
}

// Cross-panel release belongs to its source before the destination handles it.
func (a *App) updateWarehouseDrag(in *platform.Input, w, h int) bool {
	v := &a.bank
	if !a.warehouseOpen() {
		v.selected, v.dragging = false, false
		return false
	}
	if v.dragging && (in.HasKey(platform.KeyEscape) || (!in.MouseDown && !in.Clicked)) {
		v.dragging, v.selected = false, false
		return true
	}
	bx, by := a.bankUI.Origin(w, h)
	ix, iy := a.inventoryUI.Origin(w, h)
	overInventory := a.inventory.open && pointInside(in, ix, iy, ui.InventoryWidth, ui.InventoryHeight)
	if in.Clicked && v.selected && overInventory {
		key, ok := ui.InventoryHit(in.MouseX-ix, in.MouseY-iy, a.inventory.bag)
		// Dropped on a bag tab: the item goes to that bag's first empty slot
		// (0x4f5e31) and the window stays on the bag it shows, as the inventory
		// does for its own items.
		if tab := ui.InventoryBagTab(in.MouseX-ix, in.MouseY-iy); !ok && tab >= 0 {
			bag := uint8(tab + 1)
			for slot := uint8(0); slot < ui.InventoryBagSlots; slot++ {
				if _, occupied := a.state.Inventory[[2]uint8{bag, slot}]; !occupied {
					a.moveWarehouseItem([2]uint8{a.warehouseBag(), v.source}, [2]uint8{bag, slot})
					return true
				}
			}
			v.selected, v.dragging = false, false
			return true
		}
		if ok && key[0] != 0 {
			a.moveWarehouseItem([2]uint8{a.warehouseBag(), v.source}, key)
		} else {
			v.selected, v.dragging = false, false
		}
		return true
	}
	if in.Clicked && a.inventory.selected && !overInventory {
		slot, ok := ui.WarehouseHit(in.MouseX-bx, in.MouseY-by, v.page)
		if ok && a.inventory.source[0] != 0 {
			a.moveWarehouseItem(a.inventory.source, [2]uint8{a.warehouseBag(), slot})
			return true
		}
	}
	return false
}
func (a *App) drawWarehouse(w, h int, portrait *render.Texture) {
	v := &a.bank
	x, y := a.bankUI.Origin(w, h)
	a.clearTooltipUnderPanel(x, y, ui.WarehouseWidth, ui.WarehouseHeight)
	a.bankUI.Draw(a.sprites, w, h, v.page, portrait)
	for slot, it := range a.warehouseItems() {
		cx, cy, ok := ui.WarehouseCell(v.page, slot)
		if !ok {
			continue
		}
		cx, cy = cx+x, cy+y
		if v.selected && slot == v.source {
			a.sprites.Draw(a.white, float32(cx), float32(cy), 32, 32, 0, 0, 1, 1, [4]float32{1, .8, .2, .3})
		}
		if d := a.itemsItem(it.Type, it.TypeID); d != nil {
			a.drawItemIcon(d, cx, cy, 32, white, &it)
		}
		if a.itemCount(it.Type, it.TypeID, it.Count) > 0 {
			text := fmt.Sprint(it.Count)
			a.font.Draw(a.sprites, text, cx+32-a.font.Width(text), cy+18, white, 32)
		}
	}
	if v.dragging {
		return
	}
	if hover, hovered := ui.WarehouseHit(a.lastMouseX-x, a.lastMouseY-y, v.page); hovered {
		if it, ok := a.warehouseItems()[hover]; ok {
			if d := a.itemsItem(it.Type, it.TypeID); d != nil {
				var feeText string
				if v.guild {
					feeText = a.tr(12362)
					if it.Type == 94 {
						feeText = a.tr(12363)
					}
				}
				a.drawItemDetails(d, &it, it.Count, min(x+ui.WarehouseWidth+4, max(0, w-304)), min(y+140, max(0, h-110)), w, h, feeText)
			}
		}
	}
}

func (a *App) drawWarehouseDrag() {
	v := &a.bank
	if !v.dragging || !v.selected || !a.warehouseOpen() {
		return
	}
	if it, ok := a.warehouseItems()[v.source]; ok {
		if d := a.itemsItem(it.Type, it.TypeID); d != nil {
			a.drawItemIcon(d, a.lastMouseX-16, a.lastMouseY-16, 32, [4]float32{1, 1, 1, .85}, &it)
		}
	}
}
