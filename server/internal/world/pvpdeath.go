package world

import (
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/pvprank"
	"math"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Death by a user follows CUser::Death 0x004586a0:
//
//   - CDamage (Add 0x004b8b60 from CUser::Damage, GetMax 0x004b8c40, Reset
//     at every death) keeps the HP damage taken per attacker; the reward
//     goes to the attacker with the most damage (the first one on a tie),
//     not to the last hit. A mob or the user itself at the top gives
//     nothing.
//   - Normal PvP (PvPGetDeathType 3; duel deaths 4 and 5 are handled
//     separately) credits kills in GX. A solo killer gets KillPCFrom.
//     In a party, the first eligible member gets KillPCFrom and the others
//     get KillPCTo.
//   - KillPCFrom 0x0045b050: KillPCTo for the killer, then the dead user's
//     deaths + 1 (up to g_tDeathInfo[15] 200000) with 0x020E index 1.
//     KillPCTo 0x0045aed0: QuestUpdateKillPC, kills + 1 (up to
//     g_tKillInfo[15] 200000; the NEW_SERVER doubling is off in the EP4.5
//     ps_game.ini), 0x020E index 0, 0x0522 (id, kills) to the view.
//   - PvP deaths preserve inventory and gold.
//
//// GX extends both counters to the client's final threshold (1000000),
// with all 31 client-table reward stages, as explicitly requested.
// Country blessing and duel counters are handled by killcount.go, duel.go and
// guildduel.go. Native redeemable reward stages are in pvpreward.go.

const (
	damageByUser = 1
	damageByMob  = 2
	pvpKillCap   = pvprank.CounterCap
	pvpDeathCap  = pvprank.CounterCap
	pvpExpFactor = 1.2
)

// damageEntry is one CDamage node.
type damageEntry struct {
	Type   uint8
	ID     uint32
	Amount uint32
}

// addDamage is CDamage::Add.
func (s *Session) addDamage(kind uint8, id, amount uint32) {
	for i := range s.damageLedger {
		if e := &s.damageLedger[i]; e.Type == kind && e.ID == id {
			e.Amount += amount
			return
		}
	}
	s.damageLedger = append(s.damageLedger, damageEntry{kind, id, amount})
}

// topDamage is CDamage::GetMax: the first entry with the highest damage.
func (s *Session) topDamage() (damageEntry, bool) {
	if len(s.damageLedger) == 0 {
		return damageEntry{}, false
	}
	top := s.damageLedger[0]
	for _, e := range s.damageLedger[1:] {
		if top.Amount < e.Amount {
			top = e
		}
	}
	return top, true
}

// pvpExp is CUser::GetExp: the dead user's level x 50 x ((dead - killer +
// 1) x 0.2 + 1), at least 1, and 0 when the dead user is 10+ levels above.
func pvpExp(deadLevel, killerLevel uint16) uint32 {
	diff := int32(deadLevel) - int32(killerLevel)
	if diff > 9 {
		return 0
	}
	v := int64(math.Trunc(float64(int32(deadLevel)*50) * (float64(diff+1)*0.2 + 1)))
	return uint32(max(v, 1))
}

// pvpKill is a kill credited to a user's session (KillPCTo).
type pvpKill struct{ VictimID uint32 }

func (p *ConnectedPlayer) QueuePvPKill(k pvpKill) {
	p.mailMu.Lock()
	p.pvpKillCount = min(satAdd(p.pvpKillCount, 1), pvpKillCap)
	p.mailMu.Unlock()
}
func (p *ConnectedPlayer) takePvPKillCount() uint32 {
	p.mailMu.Lock()
	defer p.mailMu.Unlock()
	n := p.pvpKillCount
	p.pvpKillCount = 0
	return n
}

// QueueItemGrant hands an item straight to a user (ItemGet).
func (p *ConnectedPlayer) QueueItemGrant(it db.Item) bool {
	p.mailMu.Lock()
	if len(p.itemGrants) >= maxPendingItemGrants {
		p.mailMu.Unlock()
		return false
	}
	p.itemGrants = append(p.itemGrants, it)
	p.mailMu.Unlock()
	return true
}

func (p *ConnectedPlayer) TakeItemGrants() []db.Item {
	p.mailMu.Lock()
	defer p.mailMu.Unlock()
	g := p.itemGrants
	p.itemGrants = nil
	return g
}

// tickPvP applies the kills and items other sessions credited to this one.
func (s *Session) tickPvP() {
	if s.player == nil || s.char == nil {
		return
	}
	s.killPCToCount(s.player.takePvPKillCount())
	for _, it := range s.player.TakeItemGrants() {
		adds, left, err := s.giveItem(&it, it.Count, maxStackOf(it.Type, it.TypeID))
		if err != nil {
			s.log.Printf("GX Server World pvp item grant failed: %v", err)
		}
		for _, a := range adds {
			s.send(a, "0x0205 pvp item")
		}
		if left > 0 {
			// No room: the item lands at the user's feet, theirs alone.
			it.Count = left
			s.placeGroundItem(it, s.char.ID, true)
		}
		if s.player != nil {
			s.player.UpdateItems(s.items)
		}
	}
}

// killPCTo is CUser::KillPCTo on this (the killing) user.
func (s *Session) killPCTo() { s.killPCToCount(1) }
func (s *Session) killPCToCount(count uint32) {
	if count == 0 {
		return
	}
	ch := s.char
	for range min(count, uint32(255*len(s.quests))) {
		s.questKillPC()
	}
	if ch.Kills >= pvpKillCap {
		return
	}
	remaining := pvpKillCap - ch.Kills
	added := uint32(min(uint64(count)*uint64(rewardRates.PvPKills), uint64(remaining)))
	ch.Kills += added
	s.persist()
	addKillCountBatch(s.faction, min(count, remaining), s.log) // Actual kills for the killer's country.
	s.send(buildKillCountUpdate(0, ch.Kills), "0x020E kills")
	w := packet.New(0x0522)
	w.U32(ch.ID)
	w.U32(ch.Kills)
	frame := w.MustFinish()
	s.send(frame, "0x0522 kill count")
	s.sendToViewers(frame, "0x0522 kill count (view)")
}

// questKillPC is CUser::QuestUpdateKillPC: the first active quest that
// ends on player kills counts one.
func (s *Session) questKillPC() {
	for _, q := range s.quests {
		if q.def.EndType != 3 {
			continue
		}
		if q.counters[0] == 255 {
			return
		}
		s.bumpQuestCounter(q, 0)
		if q.def.EndNpcType == 0 && q.def.PvpKills <= q.counters[0] {
			s.completeQuest(q, 0, true, false)
		}
		return
	}
}

// deathByUser runs CUser::Death's reward part on this (the dead) user.
func (s *Session) deathByUser() {
	ch := s.char
	defer func() { s.damageLedger = nil }() // CDamage::Reset
	top, ok := s.topDamage()
	if !ok || top.Type != damageByUser || top.ID == ch.ID {
		return
	}
	killer := FindPlayerByID(top.ID)
	if killer == nil {
		return
	}
	if s.duelDeath(killer) || s.guildDuelDeath(killer) {
		return // PvPGetDeathType 4/5: PvPChkLoseDeath, no rewards
	}
	partyMu.Lock()
	pt := partyOf(killer)
	var members []*ConnectedPlayer
	if pt != nil {
		members = append(members, pt.members...)
	}
	partyMu.Unlock()
	// Killing a player gives no experience: only the kill itself is credited.
	if len(members) == 0 {
		s.killPCFrom(killer)
		s.notifyKill(killer)
		return
	}
	var near []*ConnectedPlayer
	for _, m := range members {
		mx, mz := m.PositionXZ()
		if m.MapID() != ch.MapID || DistanceXZ(mx, mz, ch.PosX, ch.PosZ) > partyRewardRange {
			continue
		}
		near = append(near, m)
	}
	for i, m := range near {
		if i == 0 {
			s.killPCFrom(m)
		} else {
			m.QueuePvPKill(pvpKill{VictimID: ch.ID})
		}
		s.notifyKill(m)
	}
}

// notifyKill tells a user who gets the credit for this user's death which player
// it killed; the client shows it in red (SysMsg 8).
func (s *Session) notifyKill(killer *ConnectedPlayer) {
	w := packet.New(packet.Opcode(proto.GXKillNotice))
	w.FixedString(s.char.Name, 21)
	killer.Send(w.MustFinish(), "GX kill notice")
}

// killPCFrom is CUser::KillPCFrom: the killer's KillPCTo, then this
// user's death count.
func (s *Session) killPCFrom(killer *ConnectedPlayer) {
	ch := s.char
	killer.QueuePvPKill(pvpKill{VictimID: ch.ID})
	if ch.Deaths >= pvpDeathCap {
		return
	}
	ch.Deaths++
	s.persist()
	s.send(buildKillCountUpdate(1, ch.Deaths), "0x020E deaths")
}

// placeGroundItem drops an item at this user's feet (drop type 3).
func (s *Session) placeGroundItem(it db.Item, owner uint32, noParty bool) {
	ch := s.char
	x, z := dropPosition(ch.PosX, ch.PosZ)
	mi := MapItem{ID: nextMapItemID(), OwnerID: owner, NoParty: noParty, DroppedAt: time.Now(), PosX: x, PosZ: z, Item: it}
	addMapItem(ch.MapID, mi)
	dispatchItemEnter(ch.MapID, &mi, "0x0401 pvp item drop")
}
