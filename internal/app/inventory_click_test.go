package app

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"

	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func TestInventoryEquipmentTypesAndNoZeroPrice(t *testing.T) {
	for _, typ := range []uint8{1, 10, 15, 45, 58, 65} {
		if got := inventoryEquipSlots(typ); len(got) != 1 || got[0] != 5 {
			t.Fatal(typ, got)
		}
	}
	for typ, want := range map[uint8]uint8{16: 0, 17: 1, 18: 2, 20: 3, 21: 4, 34: 6, 24: 7, 23: 8, 42: 13, 66: 0, 92: 4} {
		if got := inventoryEquipSlots(typ); len(got) != 1 || got[0] != want {
			t.Fatal(typ, got)
		}
	}
	for _, typ := range []uint8{0, 25, 95, 100, 255} {
		if len(inventoryEquipSlots(typ)) != 0 {
			t.Fatal("consumable became equipment", typ)
		}
	}
	if itemSellingPrice(0) != "" || itemSellingPrice(17) != "Sell: 17 gold" {
		t.Fatal("selling price presentation")
	}
	if debuffFlash(.25) != float32(63)/255 || debuffFlash(1.25) != float32(63)/255 {
		t.Fatal("debuff pulse must use elapsed seconds")
	}
}

func TestInventoryComparisonByEquipmentFamily(t *testing.T) {
	a := inputApp()
	a.items = make(map[sdata.ItemKey]*sdata.Item)
	for _, tc := range []struct {
		name                        string
		bagType, equippedType, slot uint8
	}{
		{"weapon", 1, 5, 5},
		{"helmet", 16, 31, 0},
		{"armor", 17, 32, 1},
		{"leggings", 18, 33, 2},
		{"gloves", 20, 35, 3},
		{"boots", 21, 36, 4},
		{"shield", 24, 39, 7},
		{"cloak", 19, 34, 6},
		{"necklace", 23, 38, 8},
		{"ring", 22, 37, 9},
		{"bracelet", 40, 41, 11},
		{"mount", 42, 42, 13},
	} {
		a.items[sdata.ItemKey{Type: tc.equippedType, TypeID: 1}] = &sdata.Item{Type: tc.equippedType, TypeID: 1, Name: tc.name}
		a.state.Inventory[[2]uint8{0, tc.slot}] = proto.InventoryItem{Type: tc.equippedType, TypeID: 1, Count: 1}
		tip := a.inventoryComparison(1, tc.bagType)
		if tc.bagType == 42 {
			if len(tip.tooltip) != 0 {
				t.Fatal("mounts must use a single tooltip")
			}
			continue
		}
		if len(tip.tooltip) == 0 || tip.tooltip[0].text != tc.name {
			t.Errorf("%s: missing equipped tooltip: %v", tc.name, tip.tooltip)
		}
	}
}

func TestInventoryPressOnlyLiftsIconAfterPointerMoves(t *testing.T) {
	for _, composing := range []bool{false, true} {
		a := inputApp()
		a.inventory.open, a.inventory.bag, a.chatOpen = true, 1, composing
		key := [2]uint8{1, 0}
		item := proto.InventoryItem{Bag: 1, Slot: 0, Type: 25, TypeID: 1, Count: 2}
		a.state.Inventory[key] = item
		x, y := a.inventoryUI.Origin(1024, 768)
		cx, cy, _, _ := ui.InventoryCell(1, 0)
		x, y = x+cx+10, y+cy+10
		update := func(in platform.Input) {
			if composing {
				a.updateChatItemDrag(&in, 1024, 768)
			} else {
				a.worldControls(&in, 1024, 768)
			}
		}
		press := platform.Input{MouseX: x, MouseY: y, Pressed: true, MouseDown: true}
		for range 2 {
			update(press)
			update(platform.Input{MouseX: x + 3, MouseY: y - 3, MouseDown: true})
			if a.inventory.dragging || a.inventory.selected || !a.inventory.pressedItem {
				t.Fatalf("composing=%v: press or jitter lifted the item", composing)
			}
			update(platform.Input{MouseX: x + 3, MouseY: y - 3, Clicked: true})
			if a.inventory.dragging || a.inventory.pressedItem || a.inventory.selected {
				t.Fatalf("composing=%v: stale pointer gesture after click", composing)
			}
		}
		update(press)
		update(platform.Input{MouseX: x - 4, MouseY: y, MouseDown: true})
		if !a.inventory.dragging || a.inventory.source != key {
			t.Fatalf("composing=%v: held movement did not lift the item", composing)
		}
		update(platform.Input{MouseX: x, MouseY: y, Clicked: true})
		if a.inventory.dragging || a.inventory.selected || a.state.Inventory[key] != item {
			t.Fatalf("composing=%v: returning a drag to its source changed inventory", composing)
		}
	}
}
