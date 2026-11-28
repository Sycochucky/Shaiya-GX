package app

import (
	"fmt"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
)

func (a *App) confirmEconomy(quote proto.EconomySettings, message string, send func()) {
	if a.gemConfirm == nil || a.gemConfirm.Visible {
		return
	}
	a.economyConfirm = func() {
		if a.state == nil || a.world == nil || a.state.Self.Dead ||
			a.state.Economy.Revision != quote.Revision || a.shop.pending || a.inventory.pending {
			a.questMessage("Prices changed. Review the current quote and try again.")
			return
		}
		send()
	}
	a.gemConfirm.OpenFull(message, true, a.font)
}

func (a *App) merchantBuyQuote(index uint8) (uint32, bool) {
	e := a.state.Entities[a.shop.npcID]
	if e == nil {
		return 0, false
	}
	def := a.npcDefs[sdata.NPCKey{Type: e.NpcType, ID: e.TypeID}]
	if int(index) >= len(def.Items) {
		return 0, false
	}
	entry := def.Items[index]
	item := a.itemsItem(entry.Type, entry.TypeID)
	if item == nil {
		return 0, false
	}
	return proto.ScaleVendorUnitPrice(uint32(max(item.BuyPrice, 0)), a.state.Economy.BuyPercent), true
}

func priceConfirmation(action string, count uint16, perUnit uint32) string {
	return fmt.Sprintf("%s %d item(s) for %d gold?", action, count, uint64(count)*uint64(perUnit))
}
