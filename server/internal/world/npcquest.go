package world

import (
	"github.com/Spelunkern/Shaiya-GX/internal/catalogfiles"
	"path/filepath"
	"sync"

	"github.com/Spelunkern/Shaiya-GX/internal/npcquest"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// NpcQuest.SData holds the NPC definitions (DBInitializeConstNpcQuest
// 0x004caf80; format recovered by Eternity's npcs.cpp): 13 lists, one
// per NPC type, indexed by row (type id - 1). Type 1 NPCs are shops with
// their item list; type 2 are gatekeepers with three destinations.

const (
	npcTypeShop       = 1
	npcTypeGatekeeper = 2
	npcTypeCount      = 13
	gatekeeperReach   = 30.0
)

type npcGate struct {
	MapID   uint16
	X, Y, Z float32
	Money   uint32
}

type npcDef struct {
	Type                   uint8
	TypeID                 uint16
	Country                uint8 // 0 light, 1 fury, 2 both
	Name                   string
	ShopType               uint8
	ShopItems              [][2]uint8
	StartQuests, EndQuests []uint16
	// stNpcInfo +0x08 / +0x0C: patrol step and step interval (ms).
	MoveDistance, MoveSpeed int32
	Gates                   [3]npcGate
}

var npcDefs [npcTypeCount][]npcDef

func loadNpcQuest(log *logx.Logger) error {
	doc, err := catalogfiles.Read(filepath.Join(gameDataDir(), catalogfiles.NPCQuest), catalogfiles.ParseNPCQuest)
	if err != nil {
		return err
	}
	npcs, quests, itemLinks := buildNPCQuest(doc.Catalog)
	npcDefs, questDefs, questItemLinks = npcs, quests, itemLinks
	buildMobQuestLinks()
	log.Printf("GX Server World loaded NpcQuest.SData npcs=%d shops=%d gatekeepers=%d item_links=%d quests=%d",
		len(doc.Catalog[npcquest.NPCs]), len(npcs[0]), len(npcs[1]), len(itemLinks), len(quests))
	return nil
}

func buildNPCQuest(c npcquest.Catalog) ([npcTypeCount][]npcDef, []questDef, map[[2]uint8]questItemLink) {
	var npcs [npcTypeCount][]npcDef
	for _, row := range c[npcquest.NPCs] {
		group := row[0].(int64) - 1
		n := npcDef{Type: uint8(row[2].(int64)), TypeID: uint16(row[3].(int64)),
			MoveDistance: int32(row[5].(int64)), MoveSpeed: int32(row[6].(int64)),
			Name: row[8].(string), ShopType: uint8(row[9].(int64)), Country: 2}
		if country := row[7].(int64); country == 1 || country == 2 {
			n.Country = uint8(country - 1)
		}
		npcs[group] = append(npcs[group], n)
	}
	for _, row := range c[npcquest.Products] {
		n := &npcs[row[0].(int64)-1][row[1].(int64)-1]
		n.ShopItems = append(n.ShopItems, [2]uint8{uint8(row[3].(int64)), uint8(row[4].(int64))})
	}
	for _, row := range c[npcquest.Gates] {
		n := &npcs[row[0].(int64)-1][row[1].(int64)-1]
		n.Gates[row[2].(int64)-1] = npcGate{
			MapID: uint16(row[3].(int64)), X: float32(row[4].(float64)),
			Y: float32(row[5].(float64) + float64(float32(0.9))), Z: float32(row[6].(float64)),
			Money: uint32(row[7].(int64)),
		}
	}
	for _, row := range c[npcquest.NPCLinks] {
		n := &npcs[row[0].(int64)-1][row[1].(int64)-1]
		quest := uint16(row[4].(int64))
		if row[2].(int64) == 0 {
			n.StartQuests = append(n.StartQuests, quest)
		} else {
			n.EndQuests = append(n.EndQuests, quest)
		}
	}
	itemLinks := make(map[[2]uint8]questItemLink)
	for _, row := range c[npcquest.ItemLinks] {
		key := [2]uint8{uint8(row[0].(int64)), uint8(row[1].(int64))}
		links := itemLinks[key]
		quest := uint16(row[4].(int64))
		if row[2].(int64) == 0 {
			links.Start = append(links.Start, quest)
		} else {
			links.End = append(links.End, quest)
		}
		itemLinks[key] = links
	}
	quests := make([]questDef, len(c[npcquest.Quests]))
	for i, row := range c[npcquest.Quests] {
		quests[i] = sqliteQuest(row)
	}
	return npcs, quests, itemLinks
}

// findNpcDef is GetNpc (0x0041d640): by type and row.
func findNpcDef(npcType uint8, typeID uint16) *npcDef {
	if npcquest.RetiredType(npcType) || npcType < 1 || npcType > npcTypeCount || typeID < 1 || int(typeID) > len(npcDefs[npcType-1]) {
		return nil
	}
	return &npcDefs[npcType-1][typeID-1]
}

func npcOnMap(mapID uint16, id uint32) (MapNpc, bool) {
	for _, n := range EntitiesOnMap(mapID).NPCs {
		if n.ID == id {
			return n, true
		}
	}
	return MapNpc{}, false
}

func buildShopBuyResult(result, bag, slot, itemType, typeID uint8, count uint16, money uint32) []byte {
	w := packet.New(packet.ShopBuy)
	w.Bytes([]byte{result, bag, slot, itemType, typeID})
	w.Count(count)
	w.U32(money)
	return w.MustFinish()
}

// handleShopBuy (0x0702): npc id, item index, count; CUser::PacketShop
// (0x0046af10). Result 0 bought, 1 not enough gold, 2 no room. A stack
// must take the whole purchase; otherwise the first free slot is used.
func (s *Session) handleShopBuy(p []byte) { s.inventoryOperation(func() { s.planShopBuy(p) }) }
func (s *Session) planShopBuy(p []byte) {
	if len(p) < 8 || s.char == nil {
		return
	}
	ch := s.char
	r := packet.NewReader(p)
	npcID, index, count := r.U32(), r.U8(), r.Count()
	if r.Err() != nil {
		return
	}
	npc, ok := npcOnMap(ch.MapID, npcID)
	if !ok || npc.Type != npcTypeShop {
		return
	}
	def := findNpcDef(npc.Type, npc.TypeID)
	if def == nil || def.ShopType == proto.AuctionBoard || int(index) >= len(def.ShopItems) {
		return
	}
	entry := def.ShopItems[index]
	item := db.MustDefs().Item(entry[0], entry[1])
	if item == nil || count == 0 || count > item.MaxStack() {
		return
	}
	unitPrice := proto.ScaleVendorUnitPrice(uint32(max(item.Buy, 0)), activeEconomy().VendorBuyPercent)
	if s.inGuildHouseZone() {
		// CUser::PacketShop: inside a house map the guild's NPC levels set the price.
		g, rate, result, detail := s.houseShopRate(def.ShopType)
		if g == nil {
			s.houseNpcRefusal(guildBankRefused, guildBankRanked)
			return
		}
		if result != 0 {
			s.houseNpcRefusal(result, detail)
			return
		}
		unitPrice = houseRound(float64(max(item.Buy, 0)) * rate)
	}
	if !s.economyAmountMatches(unitPrice) {
		return
	}
	price := uint64(unitPrice) * uint64(count)
	if uint64(ch.Money) < price {
		s.send(buildShopBuyResult(1, 0, 0, item.Type, item.TypeID, count, 0), "0x0702 shop buy (no gold)")
		return
	}
	var placed *db.Item
	if item.MaxStack() > 1 {
		for i := range s.items {
			it := &s.items[i]
			if it.Bag >= 1 && it.Bag <= InventoryLastBag && it.Type == item.Type && it.TypeID == item.TypeID &&
				uint16(it.Count)+uint16(count) <= uint16(item.MaxStack()) {
				it.Count += count
				if err := s.saveInventoryItem(it); err != nil {
					s.log.Printf("GX Server World shop stack failed: %v", err)
					return
				}
				placed = it
				break
			}
		}
	}
	if placed == nil {
		free := freeInventorySlots(s.items)
		if len(free) == 0 {
			s.send(buildShopBuyResult(2, 0, 0, item.Type, item.TypeID, 0, 0), "0x0702 shop buy (no room)")
			return
		}
		n := uint16(1)
		if item.MaxStack() > 1 {
			n = count
		}
		it := db.Item{Bag: free[0][0], Slot: free[0][1], Type: item.Type, TypeID: item.TypeID, Quality: item.Quality,
			Count: n, MakeType: "B"}
		if err := s.addInventoryItem(&it); err != nil {
			s.log.Printf("GX Server World shop item failed: %v", err)
			return
		}
		s.items = append(s.items, it)
		placed = &s.items[len(s.items)-1]
	}
	ch.Money -= uint32(price)
	s.persist()
	s.log.Printf("GX Server World bought %d/%d x%d for %d at npc %d", item.Type, item.TypeID, count, price, npc.TypeID)
	s.send(buildShopBuyResult(0, placed.Bag, placed.Slot, placed.Type, placed.TypeID, placed.Count, ch.Money), "0x0702 shop buy")
}

func buildGatekeeperResult(ok bool, money uint32) []byte {
	w := packet.New(packet.Gatekeeper)
	w.Bool(ok)
	w.U32(money)
	return w.MustFinish()
}

// handleGatekeeper (0x020C): npc id, destination 0-2. A gatekeeper of the
// character's faction (or neutral) within 30 units moves it for the
// destination's fee.
func (s *Session) handleGatekeeper(p []byte) {
	if len(p) < 7 || s.char == nil {
		return
	}
	ch := s.char
	npcID, _ := packet.U32(p[2:])
	gate := p[6]
	fail := func(why string) {
		s.log.Printf("GX Server World gatekeeper refused: %s", why)
		s.send(buildGatekeeperResult(false, ch.Money), "0x020C gatekeeper (refused)")
	}
	npc, ok := npcOnMap(ch.MapID, npcID)
	if gate > 2 || !ok || npc.Type != npcTypeGatekeeper || DistanceXZ(npc.PosX, npc.PosZ, ch.PosX, ch.PosZ) >= gatekeeperReach {
		fail("no gatekeeper in reach")
		return
	}
	def := findNpcDef(npc.Type, npc.TypeID)
	if def == nil || (def.Country != 2 && def.Country != s.faction) {
		fail("wrong faction")
		return
	}
	g := def.Gates[gate]
	if g.MapID == 0 && g.X == 0 && g.Z == 0 {
		fail("empty destination")
		return
	}
	if !s.canEnterMap(g.MapID, ch.Level) || !hasGrid(g.MapID) {
		fail("no gold or unknown map")
		return
	}
	if s.movement.geometry != nil {
		geometry, err := s.playerGeometry(g.MapID)
		if err != nil {
			fail("destination physics unavailable")
			return
		}
		if _, ok := geometry.Spawn([3]float32{g.X, g.Y, g.Z}); !ok {
			fail("destination has no physical floor")
			return
		}
	}
	fee := g.Money
	if s.inGuildHouseZone() {
		guildMu.Lock()
		guildOf := s.player.guild.guild
		if guildOf == nil {
			guildMu.Unlock()
			s.houseNpcRefusal(guildBankRefused, guildBankRanked)
			return
		}
		rate := guildOf.gateRate()
		guildMu.Unlock()
		fee = houseRound(float64(fee) * rate)
	}
	if ch.Money < fee {
		fail("no gold")
		return
	}
	ch.Money -= fee
	s.persist()
	s.send(buildGatekeeperResult(true, ch.Money), "0x020C gatekeeper")
	s.log.Printf("GX Server World gatekeeper to map %d for %d", g.MapID, g.Money)
	s.applyTeleport(g.MapID, g.X, g.Y, g.Z, "gatekeeper")
}

// NPC usage (CUser::PacketNpc 0x00467510): 0x0E04 records the NPC a
// player is talking to; an NPC in use stops patrolling.
var (
	npcUseMu sync.Mutex
	npcUsers = map[uint32]int{}
)

func (s *Session) useNpc(id uint32) {
	if s.char != nil && id != 0 {
		if _, ok := npcOnMap(s.char.MapID, id); !ok {
			return
		}
	}
	if id == s.usingNpc {
		s.auctionContact = s.auctionNPC(id)
		return
	}
	npcUseMu.Lock()
	if s.usingNpc != 0 {
		if npcUsers[s.usingNpc]--; npcUsers[s.usingNpc] <= 0 {
			delete(npcUsers, s.usingNpc)
		}
	}
	if id != 0 {
		npcUsers[id]++
	}
	npcUseMu.Unlock()
	s.usingNpc = id
	s.auctionContact = s.auctionNPC(id)
}

func npcInUse(id uint32) bool {
	npcUseMu.Lock()
	defer npcUseMu.Unlock()
	return npcUsers[id] > 0
}
