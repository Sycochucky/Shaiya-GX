package world

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

type economyState struct {
	config.Economy
	revision uint32
}

var economySettings atomic.Pointer[economyState]
var economyEditMu sync.RWMutex

func activeEconomy() economyState {
	if state := economySettings.Load(); state != nil {
		return *state
	}
	return economyState{Economy: config.DefaultEconomy(), revision: 1}
}

func (e economyState) packet(rejected bool) []byte {
	payload := proto.EconomySettingsPacket(proto.EconomySettings{
		Revision: e.revision, BuyPercent: e.VendorBuyPercent, SellPercent: e.VendorSellPercent,
		RerollGold: e.RerollGold, Rejected: rejected,
	})
	frame := packet.New(packet.Opcode(proto.GXEconomySettings))
	frame.Bytes(payload[2:])
	return frame.MustFinish()
}

func SaveEconomy(req operatorwire.Request, log *logx.Logger) error {
	if req.Economy == nil {
		return fmt.Errorf("missing economy settings")
	}
	e := *req.Economy
	economyEditMu.Lock()
	old := activeEconomy()
	if err := config.SaveEconomy(".", e); err != nil {
		economyEditMu.Unlock()
		return err
	}
	next := economyState{Economy: e, revision: max(old.revision+1, 1)}
	economySettings.Store(&next)
	// Publish the broadcast before allowing another update to publish.
	for _, player := range AllPlayers() {
		player.Send(next.packet(false), "economy settings")
	}
	economyEditMu.Unlock()
	log.Event("Applied economy settings", "activity", "staff.settings")
	return nil
}

func (s *Session) economyAction(op packet.Opcode, p []byte) {
	economyEditMu.RLock()
	defer economyEditMu.RUnlock()
	state := activeEconomy()
	if op == packet.Opcode(proto.GXEconomyAction) {
		revision, amount, body, err := proto.DecodeEconomyAction(p)
		if err != nil || revision != state.revision {
			s.send(state.packet(true), "economy settings changed")
			return
		}
		s.economyConfirmedAmount = &amount
		defer func() { s.economyConfirmedAmount = nil }()
		p = body
		op = packet.Opcode(proto.Op(body))
	}
	switch op {
	case packet.ShopBuy:
		if (state.VendorBuyPercent != 100 && !s.inGuildHouseZone()) && s.economyConfirmedAmount == nil {
			s.send(state.packet(true), "economy confirmation required")
			return
		}
		s.handleShopBuy(p)
	case packet.ShopSell:
		if state.VendorSellPercent != 100 && s.economyConfirmedAmount == nil {
			s.send(state.packet(true), "economy confirmation required")
			return
		}
		s.handleShopSell(p)
	case packet.ItemCompose:
		if state.RerollGold > 0 && s.economyConfirmedAmount == nil {
			s.send(state.packet(true), "economy confirmation required")
			return
		}
		s.handleItemCompose(p)
	}
}

func (s *Session) economyAmountMatches(amount uint32) bool {
	if s.economyConfirmedAmount == nil || *s.economyConfirmedAmount == amount {
		return true
	}
	s.send(activeEconomy().packet(true), "economy quote mismatch")
	return false
}

func scaleEnchantChance(chance, percent uint32) uint32 {
	return uint32(min(uint64(chance)*uint64(percent)/100, 1000000))
}
