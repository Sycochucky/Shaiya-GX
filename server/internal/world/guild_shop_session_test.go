package world

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

func TestGuildShopPurchaseAllowsAccountReconnect(t *testing.T) {
	s, _ := forgeRuleSession(t)
	account := testAccount(t)
	s.accountID, s.sessionID = uint32(account.UserUID), 123
	if login, err := db.TryGameLogin(account.UserID, "local", s.sessionID, "127.0.0.1"); err != nil || login.Status != 0 {
		t.Fatalf("initial login failed: %+v %v", login, err)
	}
	if !s.acquireAccount(s.accountID, s.sessionID) {
		t.Fatal("could not acquire account")
	}
	t.Cleanup(func() {
		if s.accountID != 0 {
			releaseAccount(s.accountID, s.sessionID)
		}
	})
	if login, err := db.TryGameLogin(account.UserID, "local", 124, "127.0.0.1"); err != nil || login.Status != 1 {
		t.Fatalf("running session should block a second login: %+v %v", login, err)
	}

	zone := s.char.MapID
	oldZone, hadZone := mapConfig.Zones[zone]
	oldNPCs := npcDefs[0]
	t.Cleanup(func() {
		npcDefs[0] = oldNPCs
		if hadZone {
			mapConfig.Zones[zone] = oldZone
		} else {
			delete(mapConfig.Zones, zone)
		}
		destroyInstance(zone)
	})
	mapConfig.Zones[zone] = ZoneConfig{CreateType: guildHouseCreateType}
	npcDefs[0] = []npcDef{{Type: npcTypeShop, TypeID: 1, ShopType: 4, ShopItems: [][2]uint8{{25, 43}}}}
	AddEntities(zone, MapEntities{NPCs: []MapNpc{{ID: 998001, Type: npcTypeShop, TypeID: 1}}})
	s.player = RegisterPlayer(*s.char, s.items, false, FactionLight, s.out, s.logger, s)
	s.player.guild.guild = &guild{Rank: 1}

	w := packet.New(packet.ShopBuy)
	w.U32(998001)
	w.U8(0)
	w.Count(1)
	s.handleShopBuy(w.MustFinish()[2:])
	if len(s.items) != 1 || s.items[0].TypeID != 43 {
		t.Fatalf("purchase did not complete: %+v", s.items)
	}

	s.close()
	s.accountID = 0
	if login, err := db.TryGameLogin(account.UserID, "local", 124, "127.0.0.1"); err != nil || login.Status != 0 {
		t.Fatalf("account should be able to log in after disconnect: %+v %v", login, err)
	}
}
