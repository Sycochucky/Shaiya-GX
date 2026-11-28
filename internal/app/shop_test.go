package app

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func TestNDoesNotSelectNearbyNPC(t *testing.T) {
	a := &App{state: game.New(), wv: &worldView{}, npcDefs: map[sdata.NPCKey]sdata.NPC{}}
	a.state.Entities[10] = &game.Entity{ID: 10, Kind: game.KindNpc, X: 5}
	for _, key := range []rune{'n', 'N'} {
		if a.updateShop(&platform.Input{Chars: []rune{key}}, 1024, 768) ||
			a.shop.npcID != 0 || a.wv.approachTarget != 0 {
			t.Fatalf("%c activated NPC interaction", key)
		}
	}
}

func TestNPCWindowsCloseAtNativeDistance(t *testing.T) {
	for _, tc := range []struct {
		name        string
		kind        uint8
		talk, quest bool
	}{
		{name: "conversation", kind: 1, talk: true},
		{name: "merchant", kind: 1},
		{name: "gatekeeper", kind: 2},
		{name: "blacksmith", kind: 3},
		{name: "reward", kind: 4},
		{name: "warehouse", kind: 6},
		{name: "quest", kind: 1, quest: true},
		{name: "guild manager", kind: 11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &App{state: game.New(), wv: &worldView{}, gateConfirm: &ui.Message{Visible: true}}
			e := &game.Entity{ID: 42, Kind: game.KindNpc, NpcType: tc.kind}
			a.state.Entities[e.ID] = e
			a.shop = shopView{npcID: e.ID, talk: tc.talk, questMode: tc.quest, quantity: &tradeQuantity{}}
			a.gate = gatekeeperView{npc: e.ID, pending: true}
			a.shop.pending = true
			a.inventory.open = true
			a.wv.pos = render.V3{3, 0, 0}
			a.checkNPCWindowRange()
			if a.shop.npcID != e.ID || a.shop.quantity == nil || !a.gateConfirm.Visible {
				t.Fatal("NPC window closed at the allowed boundary")
			}
			a.wv.pos[1] = 1
			a.checkNPCWindowRange()
			if a.shop.npcID != 0 || a.shop.talk || a.shop.questMode || a.shop.quantity != nil || a.gateConfirm.Visible {
				t.Fatal("an out-of-range NPC window or confirmation remained open")
			}
			if !a.shop.pending || !a.gate.pending || !a.inventory.open {
				t.Fatal("closing NPC discarded a pending response or closed the inventory")
			}
		})
	}
}

func TestNPCWindowRangeTracksMovingNPCAndLeavesRemoteWarehouseOpen(t *testing.T) {
	a := &App{state: game.New(), wv: &worldView{}}
	e := &game.Entity{ID: 42, Kind: game.KindNpc}
	a.state.Entities[e.ID] = e
	a.shop.npcID = e.ID
	a.checkNPCWindowRange()
	if a.shop.npcID != e.ID {
		t.Fatal("nearby NPC window closed")
	}
	e.Z = 4
	a.checkNPCWindowRange()
	if a.shop.npcID != 0 {
		t.Fatal("window remained open after NPC moved away")
	}
	a.remoteBank = true
	a.checkNPCWindowRange()
	if !a.remoteBank {
		t.Fatal("NPC range closed remote warehouse")
	}
}

func TestInventoryDocksAndDragsWithoutSavingServicePosition(t *testing.T) {
	for _, npcType := range []uint8{1, 6} {
		a := inputApp()
		a.bankUI, _ = ui.NewWarehouse(inputTextures{})
		a.inventoryUI.Position.Move(740, 190, 1024, 768, ui.InventoryWidth, ui.InventoryHeight)
		saved := a.inventoryUI.Position
		e := &game.Entity{ID: 42, Kind: game.KindNpc, NpcType: npcType}
		a.state.Entities[e.ID] = e
		a.shop.npcID, a.shop.talk = e.ID, true
		a.enterNPCService(e)
		a.placeNewPersonalPanels(0, 1024, 768)
		if !a.inventory.open {
			t.Fatalf("service type %d did not open inventory", npcType)
		}
		position := &a.marketUI.Position
		origin := a.marketUI.Origin
		width, height := ui.MarketWidth, ui.MarketHeight
		if npcType == 6 {
			position, origin = &a.bankUI.Position, a.bankUI.Origin
			width, height = ui.WarehouseWidth, ui.WarehouseHeight
		}
		for _, x := range []int{32, 600} {
			position.Move(x, 110, 1024, 768, width, height)
			sx, sy := origin(1024, 768)
			ix, iy := a.inventoryUI.Origin(1024, 768)
			if ix+ui.InventoryWidth != sx || iy != sy || a.inventoryUI.Position != saved {
				t.Fatalf("service type %d: invalid docking or saved position changed", npcType)
			}
		}
		ix, iy := a.inventoryUI.Origin(1024, 768)
		a.worldControls(&platform.Input{MouseX: ix + 40, MouseY: iy + 15, Pressed: true, MouseDown: true}, 1024, 768)
		a.worldControls(&platform.Input{MouseX: ix + 70, MouseY: iy + 45, MouseDown: true}, 1024, 768)
		a.worldControls(&platform.Input{MouseX: ix + 70, MouseY: iy + 45, Clicked: true}, 1024, 768)
		position.Move(400, 90, 1024, 768, width, height)
		if x, y := a.inventoryUI.Origin(1024, 768); x != ix+30 || y != iy+30 || a.inventoryUI.Position != saved {
			t.Fatalf("service type %d: drag did not move inventory independently: %d,%d", npcType, x, y)
		}
		a.closeNPC()
		x, y := a.inventoryUI.Origin(1024, 768)
		if x != 740 || y != 190 {
			t.Fatalf("service type %d: original position not restored: %d,%d", npcType, x, y)
		}
		a.shop.npcID, a.shop.talk = e.ID, true
		a.enterNPCService(e)
		sx, sy := origin(1024, 768)
		if ix, iy := a.inventoryUI.Origin(1024, 768); ix+ui.InventoryWidth != sx || iy != sy {
			t.Fatalf("service type %d: inventory did not re-anchor on re-entry", npcType)
		}
	}
}
