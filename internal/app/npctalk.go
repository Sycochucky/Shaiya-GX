package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

// Every NPC opens its conversation window first (Game.exe 0x48e050 case 100);
// the service button then opens the NPC's own window and "talk" its quest.

// npcServiceLabel is FUN_0050f790 / FUN_0050f8e0: the service button text.
func (a *App) npcServiceLabel(e *game.Entity, def sdata.NPC) string {
	id := 0x14b8 // OK
	switch e.NpcType {
	case 1:
		id = 0x14ae // trade
		switch def.MerchantType {
		case 8:
			id = 0x14f3
		case 9:
			id = 0x14f5
		}
	case 2:
		id = 0x14b4
	case 3:
		id = 5161 // Blacksmith: GX replaces the native "repair weapon" with the forge menu
	case 4:
		id = 0x14b7
	case 6:
		id = 0x14b3
	case 11:
		id = 0x14b5
	}
	return a.msg.Get(id)
}

// npcHasService reports whether the NPC has a window of its own in GX.
func (a *App) npcHasService(e *game.Entity) bool {
	switch e.NpcType {
	case 1, 2, 3, 6, 11:
		return true
	case 4:
		return a.pvpRewardUI != nil
	}
	return false
}

// Game.exe 0x4d58ce closes NPC windows when the 3D distance exceeds 3.
const npcWindowRange = 3

func (a *App) checkNPCWindowRange() {
	id := a.shop.npcID
	if id == 0 || a.state == nil || a.wv == nil {
		return
	}
	e := a.state.Entities[id]
	if e != nil && a.wv.mapID == a.state.MapID && !a.state.Self.Dead {
		delta := a.visualEntityPosition(e).Sub(a.wv.pos)
		if delta.Dot(delta) <= npcWindowRange*npcWindowRange {
			return
		}
	}

	// Discard unsent UI actions. Keep in-flight transactions so their
	// acknowledgements can still update inventory, gold or position.
	a.shop.quantity, a.shop.dragging = nil, false
	a.npcQuest = npcQuestView{}
	if !a.gate.pending {
		a.gate = gatekeeperView{}
	}
	if a.auction.npc == id {
		a.auction = auctionView{}
	}
	if a.bank.npcID == id {
		a.bank.selected, a.bank.dragging = false, false
	}
	a.house.npc, a.house.confirm = 0, houseConfirmNone
	a.pvpReward.confirm = 0
	for _, message := range []*ui.Message{a.gateConfirm, a.gemConfirm, a.auctionConfirm, a.house.confirmBox, a.pvpRewardMessage} {
		if message != nil {
			message.Visible = false
		}
	}
	a.closeNPC()
}

func (a *App) closeNPC() {
	a.shop.npcID, a.shop.talk, a.shop.questMode = 0, false, false
	a.resetBlacksmith()
	if a.world != nil {
		if err := a.world.Send(proto.NpcTalkPacket(0)); err != nil {
			a.backToLoginID(12011)
		}
	}
}

// enterNPCService leaves the conversation for the NPC's own window.
func (a *App) enterNPCService(e *game.Entity) {
	a.shop.talk = false
	a.inventory.servicePosition = ui.WindowPosition{}
	if !a.npcHasService(e) {
		a.closeNPC()
		return
	}
	if (a.merchantOpen() || a.warehouseOpen()) && !a.inventory.open {
		a.toggleInventory()
	}
	if e.NpcType == 4 {
		a.openPvPRewards()
	}
	if e.NpcType == 2 && a.gateUI != nil {
		a.gateUI.ResetWelcomeScroll()
	}
}

func (a *App) updateNPCTalk(in *platform.Input, w, h int, e *game.Entity) bool {
	def := a.npcDefs[sdata.NPCKey{Type: e.NpcType, ID: e.TypeID}]
	if len(def.StartQuests)+len(def.EndQuests) > 0 && !a.questDataLoaded && a.arc != nil {
		_ = a.loadQuests()
	}
	entries := a.npcQuestEntries(e)
	x, y := a.npcTalkUI.Origin(w, h)
	captured := panelPointer(in, x, y, ui.NPCTalkWidth, ui.NPCTalkHeight)
	closed, service, quest := a.npcTalkUI.Update(in, w, h, len(entries) > 0)
	switch {
	case closed || in.HasKey(platform.KeyEscape):
		a.closeNPC()
	case service:
		a.enterNPCService(e)
	case quest && len(entries) > 0:
		// FUN_004cf800 takes the first applicable quest of the NPC.
		a.shop.talk, a.shop.questMode = false, true
		a.npcQuest = npcQuestView{npc: e.ID, selected: entries[0].id, ending: entries[0].ending}
		a.npcQuestUI.ResetDescriptionScroll()
	}
	return captured
}

func (a *App) drawNPCTalk(w, h int, e *game.Entity) {
	def := a.npcDefs[sdata.NPCKey{Type: e.NpcType, ID: e.TypeID}]
	canQuest := len(a.npcQuestEntries(e)) > 0
	a.npcTalkUI.Draw(a.sprites, a.font, w, h, def.Name, questText(def.Welcome), a.npcServiceLabel(e, def), a.msg.Get(0x14ad), canQuest)
	if portrait := a.npcPortrait(def); portrait != nil {
		x, y := a.npcTalkUI.Origin(w, h)
		a.sprites.Draw(portrait, float32(x+7), float32(y+7), 64, 64, 0, 0, 1, 1, white)
	}
}
