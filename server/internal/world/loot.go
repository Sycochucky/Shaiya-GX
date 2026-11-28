package world

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Universal loot (a GX system that replaces the original item divisions): an
// item a party or raid earns goes straight into one member's bag. Members in
// reach of the dead monster (same map, partyRewardRange) take turns at random:
// the first item goes to anyone, then only to members who have not received
// one this round, and when everyone in reach has one a new round starts.
//
// The group remembers each member while it exists. Whoever joins a round that
// already started waits for the next one; whoever leaves or disconnects keeps
// their place for lootGrace, so leaving and returning cannot be used to get
// served again and an honest disconnection costs nothing.
const lootGrace = 10 * time.Minute

type lootEntry struct {
	served bool
	left   time.Time // zero while the member is in the group
}

// lootJoin records a member entering (partyMu held).
func (pt *party) lootJoin(id uint32, now time.Time) {
	if pt.loot == nil {
		pt.loot = map[uint32]*lootEntry{}
	}
	if e, ok := pt.loot[id]; ok && !e.left.IsZero() && now.Sub(e.left) <= lootGrace {
		e.left = time.Time{} // back in time: the same place as before
		return
	}
	served := false
	for other, e := range pt.loot {
		if other != id && e.left.IsZero() && e.served {
			served = true // the round already began: wait for the next
		}
	}
	pt.loot[id] = &lootEntry{served: served}
}

// lootLeave records a member leaving or disconnecting (partyMu held).
func (pt *party) lootLeave(id uint32, now time.Time) {
	if e := pt.loot[id]; e != nil {
		e.left = now
	}
}

func (pt *party) lootEntryOf(id uint32) *lootEntry {
	if pt.loot == nil {
		pt.loot = map[uint32]*lootEntry{}
	}
	e := pt.loot[id]
	if e == nil {
		e = &lootEntry{}
		pt.loot[id] = e
	}
	return e
}

// lootPick chooses who receives one item among members able to carry it
// (partyMu held). reach is everyone in range, fits the subset with room.
func (pt *party) lootPick(reach, fits []*ConnectedPlayer, now time.Time) *ConnectedPlayer {
	for id, e := range pt.loot {
		if !e.left.IsZero() && now.Sub(e.left) > lootGrace {
			delete(pt.loot, id) // forgotten after the grace period
		}
	}
	if len(fits) == 0 {
		return nil
	}
	var open []*ConnectedPlayer
	for _, m := range fits {
		if !pt.lootEntryOf(m.CharacterID).served {
			open = append(open, m)
		}
	}
	if len(open) == 0 {
		waiting := false
		for _, m := range reach {
			waiting = waiting || !pt.lootEntryOf(m.CharacterID).served
		}
		if !waiting { // everyone in reach has received: a new round
			for _, m := range reach {
				pt.lootEntryOf(m.CharacterID).served = false
			}
		}
		open = fits // someone with a full bag keeps their turn; the rest are served in the meantime
	}
	pick := open[rand.IntN(len(open))]
	pt.lootEntryOf(pick.CharacterID).served = true
	return pick
}

// lootFits reports whether the item fits in the member's bags.
func lootFits(p *ConnectedPlayer, tpl *db.Item) bool {
	p.mu.Lock()
	items := append([]db.Item(nil), p.items...)
	p.mu.Unlock()
	_, left := planGiveItem(&items, tpl, 1, maxStackOf(tpl.Type, tpl.TypeID))
	return left == 0
}

// distributeLoot picks the recipient for an item earned by the group of
// members (already limited to the monster's reach); nil sends it to the ground.
func distributeLoot(members []*ConnectedPlayer, tpl *db.Item, now time.Time) *ConnectedPlayer {
	if len(members) == 0 {
		return nil
	}
	var fits []*ConnectedPlayer
	for _, m := range members {
		if lootFits(m, tpl) {
			fits = append(fits, m)
		}
	}
	partyMu.Lock()
	defer partyMu.Unlock()
	pt := partyOf(members[0])
	if pt == nil {
		return nil
	}
	return pt.lootPick(members, fits, now)
}

// lootGrant is an item queued for a member's session.
type lootGrant struct {
	item    db.Item
	mapID   uint16
	x, y, z float32
}

func (p *ConnectedPlayer) QueueLoot(g lootGrant) {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	p.lootQueue = append(p.lootQueue, g)
}

func (p *ConnectedPlayer) takeLoot() []lootGrant {
	p.rewardMu.Lock()
	defer p.rewardMu.Unlock()
	out := p.lootQueue
	p.lootQueue = nil
	return out
}

// applyLoot delivers queued group loot to this session's bags; what no longer
// fits (the bag filled meanwhile) falls to the ground where the monster died.
func (s *Session) applyLoot(p *ConnectedPlayer) {
	for _, g := range p.takeLoot() {
		adds, left, err := s.giveItem(&g.item, g.item.Count, maxStackOf(g.item.Type, g.item.TypeID))
		if err != nil || len(adds) == 0 || left > 0 {
			if err == nil && len(adds) > 0 { // part of a stack fitted
				for _, a := range adds {
					s.send(a, "0x0205 group loot")
				}
				g.item.Count = left
			}
			mi := MapItem{ID: nextMapItemID(), OwnerID: s.char.ID, FromMob: true, DroppedAt: time.Now(), PosX: g.x, PosY: g.y, PosZ: g.z, Item: g.item}
			addMapItem(g.mapID, mi)
			dispatchItemEnter(g.mapID, &mi, "0x0401 group loot overflow")
			continue
		}
		for _, a := range adds {
			s.send(a, "0x0205 group loot")
		}
		activityForMap(realMap(g.mapID)).pickups.Add(uint64(g.item.Count))
		s.activity("loot.pickup", fmt.Sprintf("%s received %s through party loot", s.char.Name, itemActivityName(g.item.Type, g.item.TypeID)),
			"item_type", g.item.Type, "item_id", g.item.TypeID, "quantity", g.item.Count)
		s.questPickup(g.item.Type, g.item.TypeID)
		s.announceLoot(g.item)
	}
}

// announceLoot tells the rest of the group who received the item (SysMsg 61).
func (s *Session) announceLoot(it db.Item) {
	if s.player == nil {
		return
	}
	partyMu.Lock()
	var others []*ConnectedPlayer
	if pt := partyOf(s.player); pt != nil {
		for _, m := range pt.members {
			if m != s.player {
				others = append(others, m)
			}
		}
	}
	partyMu.Unlock()
	if len(others) == 0 {
		return
	}
	w := packet.New(packet.Opcode(proto.GXLootNotice))
	w.U32(s.char.ID)
	w.U8(it.Type)
	w.U8(it.TypeID)
	broadcast(others, w.MustFinish(), "GX group loot notice", s.logger)
}

// lootOwner is who earns a monster's loot: the group (or lone player) that
// dealt the most damage to it, not the last hit nor the aggro holder. Damage
// by members of one party counts together. It returns the biggest damage
// dealer of the winning side; nil when nobody online dealt any.
func lootOwner(m *MapMob) *ConnectedPlayer {
	type side struct {
		total uint64
		best  eventDamageEntry
		first int
	}
	sides := map[any]*side{}
	order := 0
	for _, d := range m.lootDamage {
		p := FindPlayerByID(d.ID)
		if p == nil {
			continue
		}
		var key any = d.ID
		partyMu.Lock()
		if pt := partyOf(p); pt != nil {
			key = pt
		}
		partyMu.Unlock()
		sd := sides[key]
		if sd == nil {
			sd = &side{first: order}
			sides[key] = sd
			order++
		}
		sd.total += d.Amount
		if d.Amount > sd.best.Amount {
			sd.best = d
		}
	}
	var win *side
	for _, sd := range sides {
		if win == nil || sd.total > win.total || sd.total == win.total && sd.first < win.first {
			win = sd
		}
	}
	if win == nil {
		return nil
	}
	return FindPlayerByID(win.best.ID)
}
