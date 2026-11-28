package app

import (
	"fmt"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
)

type composeView struct {
	rune, item                   [2]uint8
	pending, success, refreshing bool
	refresh                      gemRefresh
}

func (a *App) sendCompose(runeKey, itemKey [2]uint8) bool {
	if a.state == nil || a.world == nil || a.state.Self.Dead || a.shop.pending || a.inventory.pending || a.state.Exchange.Active || a.state.Duel.Phase != game.DuelNone || runeKey == itemKey {
		return false
	}
	rune, r := a.state.Inventory[runeKey]
	item, i := a.state.Inventory[itemKey]
	if !r || !i || rune.Count == 0 || item.Count != 1 || runeKey[0] == 0 {
		return false
	}
	request := proto.ComposePacket(runeKey[0], runeKey[1], itemKey[0], itemKey[1])
	quote := a.state.Economy
	if quote.Revision != 0 && quote.RerollGold > 0 {
		a.confirmEconomy(quote, fmt.Sprintf("Reroll this item for %d gold and one rune?", quote.RerollGold), func() {
			if a.state.Inventory[runeKey] == rune && a.state.Inventory[itemKey] == item && a.nearBlacksmith() {
				a.sendComposePacket(runeKey, itemKey, proto.EconomyActionPacket(quote.Revision, quote.RerollGold, request))
			}
		})
		return true
	}
	return a.sendComposePacket(runeKey, itemKey, request)
}

func (a *App) sendComposePacket(runeKey, itemKey [2]uint8, request []byte) bool {
	a.compose = composeView{rune: runeKey, item: itemKey, pending: true}
	a.shop.pending, a.shop.deadline = true, a.clock+10
	if err := a.world.Send(request); err != nil {
		a.backToLoginID(12035)
		return false
	}
	return true
}

func (a *App) confirmedCompose(p []byte) {
	v := &a.compose
	if !v.pending {
		return
	}
	switch proto.Op(p) {
	case proto.ItemCompose:
		r, err := proto.DecodeComposeResult(p)
		if err != nil || v.refreshing || v.item != [2]uint8{r.Bag, r.Slot} {
			return
		}
		v.success, v.refreshing = r.Result == 0, true
		v.refresh.start([][2]uint8{v.rune})
		a.shop.deadline = a.clock + 10
		// The rune consumption packet follows 0x0806. Its ordered self-move
		// confirms the final slot on success as well as on refusal.
		if err := a.world.Send(proto.ItemMovePacket(v.rune[0], v.rune[1], v.rune[0], v.rune[1])); err != nil {
			a.backToLoginID(12036)
		}
	case proto.ItemBagToBag:
		if v.refreshing && v.refresh.acknowledge(p) && len(v.refresh.pending) == 0 {
			v.pending, a.shop.pending = false, false
			a.syncBlacksmithItems()
			a.forgeResult(v.success)
			if !v.success {
				a.questMessageID(12037)
			}
		}
	}
}
