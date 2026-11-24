package world

import (
	"errors"
	"math/rand/v2"
	"slices"
	"sort"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/npcquest"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Quests from NpcQuest.SData (EP4/5 layout, documented by Parsec and
// recovered from ps_game.exe by Eternity: CUser::QuestEnableStart
// 0x0047f080, QuestStartRequest 0x0047f220, QuestFindItem 0x0047f8e0,
// the end/reward/progress routines). Packets: 0x0901 active list, 0x0906
// finished list, 0x0902 start (npc id, quest id), 0x0903 end / result,
// 0x0905 counter, 0x0907 reward choice, 0x0908 abandon.

type questItem struct{ Type, TypeID, Count uint8 }

type questResult struct {
	NeedMobID                 uint16 // counter selector for mob quests
	NeedMobCount              uint8
	NeedItemID, NeedItemCount uint8 // counter selector for item quests
	NeedTime                  uint32
	NeedHG                    uint16
	NeedVG                    int16
	NeedOG                    uint8
	Exp, Money                uint32
	ItemType, ItemTypeID      uint8
	NextQuest                 uint16
}

type questDef struct {
	ID                  uint16
	Name                string
	MinLevel, MaxLevel  uint16
	Faction             uint8
	Male, Female        bool
	Jobs                [6]uint8
	HG                  uint16 // also an alternative prerequisite quest
	VG                  int16  // likewise
	CG, OG, IG          uint8
	PrevQuest           uint16
	RequireParty        bool
	PartyJobs           [6]uint8
	MinMinutes          uint32
	MaxMinutes          uint32
	TickStart, TickKeep uint32
	TickReceive         uint32 // nonzero: repeatable
	StartType           uint8
	StartNpcType        uint8
	StartNpcID          uint16
	StartItemType       uint8
	StartItemID         uint8
	StartItems          [3]questItem // given on start
	EndType             uint8        // 0 npc, 1 items, 2 mobs, 3 player kills
	EndNpcType          uint8
	EndNpcID            uint16
	FarmItems           [3]questItem
	PvpKills            uint8
	MobIDs              [2]uint16
	MobCounts           [2]uint8
	ResultType          uint8
	UserSelect          uint8
	Results             [3]questResult
}

type questItemLink struct{ Start, End []uint16 }

var (
	questDefs      []questDef
	questItemLinks = map[[2]uint8]questItemLink{}
	mobQuestLinks  = map[uint16][]uint16{}
)

// sqliteQuest follows the native field order declared by QuestDetails.
func sqliteQuest(row npcquest.Row) questDef {
	i := 0
	next := func() any { v := row[i]; i++; return v }
	word := func() uint16 { return uint16(next().(int64)) }
	byteValue := func() uint8 { return uint8(next().(int64)) }
	dword := func() uint32 { return uint32(next().(int64)) }
	item := func() questItem { return questItem{byteValue(), byteValue(), byteValue()} }
	word() // the original one-based row is the lookup key
	q := questDef{ID: word(), Name: next().(string)}
	q.MinLevel, q.MaxLevel = word(), word()
	q.Faction = byteValue()
	byteValue() // reserved catalog byte
	q.Male, q.Female = byteValue() != 0, byteValue() != 0
	for i := range q.Jobs {
		q.Jobs[i] = byteValue()
	}
	q.HG, q.VG = word(), int16(word())
	q.CG, q.OG, q.IG = byteValue(), byteValue(), byteValue()
	q.PrevQuest = word()
	q.RequireParty = byteValue() != 0
	for i := range q.PartyJobs {
		q.PartyJobs[i] = byteValue()
	}
	q.MinMinutes, q.MaxMinutes = dword(), dword()
	q.TickStart, q.TickKeep, q.TickReceive = dword(), dword(), dword()
	q.StartType, q.StartNpcType, q.StartNpcID = byteValue(), byteValue(), word()
	q.StartItemType, q.StartItemID = byteValue(), byteValue()
	for i := range q.StartItems {
		q.StartItems[i] = item()
	}
	q.EndType, q.EndNpcType, q.EndNpcID = byteValue(), byteValue(), word()
	for i := range q.FarmItems {
		q.FarmItems[i] = item()
	}
	q.PvpKills = byteValue()
	for i := range q.MobIDs {
		q.MobIDs[i], q.MobCounts[i] = word(), byteValue()
	}
	q.ResultType, q.UserSelect = byteValue(), byteValue()
	for i := range q.Results {
		r := &q.Results[i]
		r.NeedMobID, r.NeedMobCount = word(), byteValue()
		r.NeedItemID, r.NeedItemCount = byteValue(), byteValue()
		r.NeedTime = dword()
		r.NeedHG, r.NeedVG, r.NeedOG = word(), int16(word()), byteValue()
		r.Exp, r.Money = dword(), dword()
		r.ItemType, r.ItemTypeID, r.NextQuest = byteValue(), byteValue(), word()
	}
	return q
}

// findQuestDef is GetQuest: by row (1-based), not the embedded id.
func findQuestDef(id uint16) *questDef {
	if id == 0 || int(id) > len(questDefs) {
		return nil
	}
	return &questDefs[id-1]
}

// buildMobQuestLinks maps mobs to the mob-kill quests that count them,
// ascending by minimum level (MobQuestLinks::build).
func buildMobQuestLinks() {
	links := map[uint16][]uint16{}
	for i := range questDefs {
		q := &questDefs[i]
		if q.EndType != 2 || q.MobIDs[0] == 0 {
			continue
		}
		for _, mob := range q.MobIDs {
			if mob == 0 {
				continue
			}
			id := uint16(i + 1)
			list := links[mob]
			dup := false
			for _, e := range list {
				dup = dup || e == id
			}
			if !dup {
				links[mob] = append(list, id)
			}
		}
	}
	for mob, list := range links {
		sort.SliceStable(list, func(a, b int) bool {
			return findQuestDef(list[a]).MinLevel < findQuestDef(list[b]).MinLevel
		})
		links[mob] = list
	}
	mobQuestLinks = links
}

// activeQuest is a quest in progress.
type activeQuest struct {
	ID        uint16
	def       *questDef
	counters  [3]uint8
	startAt   time.Time
	endAt     time.Time // timed quests
	successAt time.Time // MinMinutes elapsed
	success   uint8     // 1: waiting for the reward choice
	rewardNPC uint32
}

func (q *activeQuest) delayMinutes() int16 {
	if !q.endAt.IsZero() {
		return int16(min(max(time.Until(q.endAt), 0)/time.Minute, 32767))
	}
	return int16(min(max(time.Since(q.startAt), 0)/time.Minute, 32767))
}

func (q *activeQuest) row() db.CharQuest {
	r := db.CharQuest{QuestID: q.ID, Delay: q.delayMinutes(), Counts: q.counters}
	if q.def.usesElapsedSeconds() && !q.startAt.IsZero() {
		r.StartedAt = q.startAt.UnixMilli()
	}
	return r
}

// loadQuests restores the character's quests as PacketUserDBSetInfo does.
func (s *Session) loadQuests() error {
	s.quests, s.finishedQuests = nil, nil
	s.lastQuestItem = questItemUse{}
	rows, err := db.CharacterQuests(s.char.ID)
	if err != nil {
		return err
	}
	now := time.Now()
	var recovered []db.CharQuest
	for _, r := range rows {
		def := findQuestDef(r.QuestID)
		if def == nil {
			continue
		}
		q := &activeQuest{ID: r.QuestID, def: def, counters: r.Counts}
		delay := time.Duration(max(r.Delay, 0)) * time.Minute
		limit := npcquest.Deadline(def.MaxMinutes, def.TickKeep)
		switch {
		case r.StartedAt > 0 && def.usesElapsedSeconds():
			q.startAt = time.UnixMilli(r.StartedAt)
			if limit > 0 {
				q.endAt = q.startAt.Add(limit)
			}
		case def.MaxMinutes != 0:
			q.startAt = now.Add(delay - npcquest.Deadline(def.MaxMinutes, 0))
			q.endAt = q.startAt.Add(limit)
		default:
			q.startAt = now.Add(-delay)
			if limit > 0 {
				q.endAt = q.startAt.Add(limit)
			}
		}
		if def.MinMinutes != 0 {
			q.successAt = q.startAt.Add(npcquest.Deadline(def.MinMinutes, 0))
		}
		if r.StartedAt == 0 && def.usesElapsedSeconds() {
			recovered = append(recovered, q.row())
		}
		s.quests = append(s.quests, q)
	}
	if err := db.SaveCharacterQuests(s.char.ID, recovered); err != nil {
		return err
	}
	fin, err := db.FinishedQuests(s.char.ID)
	if err != nil {
		return err
	}
	s.finishedQuests = fin
	return nil
}

// questListPackets are 0x0901 (id, minutes left, counters) and 0x0906
// (id, success), up to 255 entries each.
func (s *Session) questListPackets() [][]byte {
	var out [][]byte
	for start := 0; start == 0 || start < len(s.quests); start += 255 {
		end := min(start+255, len(s.quests))
		w := packet.New(packet.QuestList)
		w.U8(uint8(end - start))
		for _, q := range s.quests[start:end] {
			w.U16(q.ID)
			left := uint16(0)
			if !q.endAt.IsZero() {
				left = uint16(min(max(time.Until(q.endAt), 0)/time.Minute, 65535))
			}
			w.U16(left)
			w.Bytes(q.counters[:])
		}
		out = append(out, w.MustFinish())
	}
	for start := 0; start == 0 || start < len(s.finishedQuests); start += 255 {
		end := min(start+255, len(s.finishedQuests))
		w := packet.New(packet.QuestFinishedList)
		w.U8(uint8(end - start))
		for _, f := range s.finishedQuests[start:end] {
			w.U16(f.QuestID)
			w.U8(f.Success)
		}
		out = append(out, w.MustFinish())
	}
	return out
}

// questSnapshot captures active timers and counters on exit.
func (s *Session) questSnapshot() []db.CharQuest {
	rows := make([]db.CharQuest, 0, len(s.quests))
	for _, q := range s.quests {
		rows = append(rows, q.row())
	}
	return rows
}

func (s *Session) findQuest(id uint16) *activeQuest {
	for _, q := range s.quests {
		if q.ID == id {
			return q
		}
	}
	return nil
}

func (s *Session) questFinished(id uint16, onlySuccess bool) bool {
	for _, f := range s.finishedQuests {
		if f.QuestID == id && (!onlySuccess || f.Success != 0) {
			return true
		}
	}
	return false
}

// questPrerequisites: PrevQuest, HG or VG finished successfully (any).
func (s *Session) questPrerequisites(q *questDef) bool {
	ids := []uint16{q.PrevQuest, q.HG, uint16(q.VG)}
	if ids[0] == 0 && ids[1] == 0 && ids[2] == 0 {
		return true
	}
	for _, id := range ids {
		if id != 0 && s.questFinished(id, true) {
			return true
		}
	}
	return false
}

// questEnableStart is CUser::QuestEnableStart: level range, race
// (Faction), job and prerequisites.
func (s *Session) questEnableStart(q *questDef) bool {
	ch := s.char
	if ch.Level < q.MinLevel || ch.Level > q.MaxLevel || ch.Class >= 6 || q.Jobs[ch.Class] == 0 {
		return false
	}
	if !raceAllowed(q.Faction, ch.Race) {
		return false
	}
	if (ch.Gender == 0 && !q.Male) || (ch.Gender == 1 && !q.Female) {
		return false
	}
	return s.questPrerequisites(q)
}

// raceAllowed is the Country/Faction race rule shared with items.
func raceAllowed(country, race uint8) bool {
	switch country {
	case 0, 1:
		return race == country
	case 2:
		return race == 0 || race == 1
	case 3:
		return race == 3
	case 4:
		return race == 2
	case 5:
		return race == 2 || race == 3
	}
	return true
}

func (s *Session) npcMatches(npcID uint32, npcType uint8, typeID uint16) (MapNpc, bool) {
	npc, ok := npcOnMap(s.char.MapID, npcID)
	return npc, ok && s.questNPCInReach(npc) && npc.Type == npcType && npc.TypeID == typeID
}

// Use the same authoritative interaction radius as other NPC services.
const questNPCReach = 30.0

func (s *Session) questNPCInReach(npc MapNpc) bool {
	return DistanceXZ(s.char.PosX, s.char.PosZ, npc.PosX, npc.PosZ) < questNPCReach
}

type questItemUse struct {
	item       [2]uint8
	deadline   time.Time
	pendingUID int64
}

// Native ItemUse (0x4655e0) opens the same window for types 27–29. Type 29
// consumes one unit before the client can request the quest.
func (s *Session) useQuestItem(index int) bool {
	it := s.items[index]
	if it.Bag < InventoryFirstBag || it.Bag > InventoryLastBag || s.dead || s.char.CurrentHP == 0 {
		return false
	}
	use := questItemUse{item: [2]uint8{it.Type, it.TypeID}, deadline: time.Now().Add(5 * time.Minute)}
	if it.Type == 29 && questItemNeedsAtomicStart(it.Type, it.TypeID) {
		use.pendingUID = it.UID
	} else if it.Type == 29 {
		remaining, err := s.consumeOne(index)
		if err != nil {
			s.logErr("consume quest item", err)
			return false
		}
		if s.player != nil {
			s.player.UpdateItems(s.items)
		}
		s.send(buildRemoveItem(&remaining, remaining.Count == 0), "quest item used")
		s.broadcastNear(buildUseItemResult(s.char.ID, &remaining), "0x050A quest item used", false)
	}
	s.lastQuestItem = use
	return true
}

// handleQuestStart (0x0902): NPC or item context, followed by the quest row.
func (s *Session) handleQuestStart(p []byte) {
	if len(p) < 8 || s.char == nil {
		return
	}
	npcID, _ := packet.U32(p[2:])
	id, _ := packet.U16(p[6:])
	q := findQuestDef(id)
	if q == nil || q.StartType > 4 {
		return
	}
	if q.StartType <= 2 {
		npc, ok := npcOnMap(s.char.MapID, npcID)
		if !ok || !s.questNPCInReach(npc) {
			return
		}
		if (q.StartType != 0 || q.StartNpcType != 0) && (npc.Type != q.StartNpcType || npc.TypeID != q.StartNpcID) {
			return
		}
		if q.StartType == 0 && !s.questPrerequisites(q) {
			return
		}
	} else if npcID != uint32(q.StartItemID) {
		return
	}
	if q.StartType == 4 && (s.lastQuestItem.item != [2]uint8{q.StartItemType, q.StartItemID} || !time.Now().Before(s.lastQuestItem.deadline)) {
		return
	}
	if s.findQuest(id) != nil || s.questFinished(id, false) || !s.questEnableStart(q) {
		s.log.Printf("GX Server World quest %d refused (active, finished or requirements)", id)
		return
	}
	if reason := s.questStartRestriction(q); reason != "" {
		s.questFeedback(reason)
		return
	}
	// The native acknowledgement carries an NPC only for types 1 and 2.
	context := npcID
	if q.StartType == 0 || q.StartType >= 3 {
		context = 0
	}
	s.startQuest(q, id, context)
}

func (s *Session) startQuest(q *questDef, id uint16, npcID uint32) {
	now := time.Now()
	aq := &activeQuest{ID: id, def: q, startAt: now}
	limit := npcquest.Deadline(q.MaxMinutes, q.TickKeep)
	if limit > 0 {
		aq.endAt = now.Add(limit)
	}
	if q.MinMinutes != 0 {
		aq.successAt = now.Add(npcquest.Deadline(q.MinMinutes, 0))
	}
	after := slices.Clone(s.items)
	var removed *db.Item
	pendingUID := int64(0)
	if q.StartType == 4 {
		pendingUID = s.lastQuestItem.pendingUID
	}
	if q.StartType == 2 || q.StartType == 3 || pendingUID != 0 {
		i := slices.IndexFunc(after, func(it db.Item) bool {
			return it.Bag >= InventoryFirstBag && it.Bag <= InventoryLastBag && it.Count > 0 &&
				it.Type == q.StartItemType && it.TypeID == q.StartItemID && (pendingUID == 0 || it.UID == pendingUID)
		})
		if i < 0 {
			return
		}
		if q.StartType == 2 || pendingUID != 0 {
			after[i].Count--
			item := after[i]
			removed = &item
			if item.Count == 0 {
				after = slices.Delete(after, i, i+1)
			}
		}
	}
	var grants [][]byte
	for _, it := range q.StartItems {
		if it.Type == 0 {
			continue
		}
		def := db.MustDefs().Item(it.Type, it.TypeID)
		if def == nil {
			return
		}
		tpl := db.Item{Type: it.Type, TypeID: it.TypeID, Quality: def.Quality, MakeType: "Q"}
		adds, left := planGiveItem(&after, &tpl, uint16(max(it.Count, 1)), def.MaxStack())
		if left != 0 {
			s.send(buildAddItemError(1), "quest start inventory full")
			return
		}
		grants = append(grants, adds...)
	}
	row := aq.row()
	row.Delay = int16(min(limit/time.Minute, 32767))
	if err := db.CommitQuestStart(s.char.ID, row, s.items, after, time.Duration(q.TickStart)*time.Second); err != nil {
		if errors.Is(err, db.ErrQuestStartTerm) {
			s.questFeedback("This quest's start interval has not elapsed.")
		} else {
			s.logErr("quest start transaction", err)
		}
		return
	}
	s.items = after
	s.quests = append(s.quests, aq)
	if s.player != nil {
		s.player.UpdateItems(s.items)
	}
	if removed != nil {
		s.send(buildRemoveItem(removed, removed.Count == 0), "quest start cost")
	}
	if pendingUID != 0 {
		s.lastQuestItem = questItemUse{}
		s.broadcastNear(buildUseItemResult(s.char.ID, removed), "0x050A quest item used", false)
	}
	w := packet.New(packet.QuestStart)
	w.U32(npcID)
	w.U16(id)
	s.send(w.MustFinish(), "0x0902 quest start")
	for _, frame := range grants {
		s.send(frame, "quest start item")
	}
	s.log.Printf("GX Server World quest %d %q started", id, q.Name)
}

// questItemTotals counts actual items in carried bags.
func questItemTotals(d *questDef, items []db.Item) (totals [3]int) {
	for _, it := range items {
		if it.Bag < InventoryFirstBag || it.Bag > InventoryLastBag {
			continue
		}
		for i, f := range d.FarmItems {
			if f.Type != 0 && it.Type == f.Type && it.TypeID == f.TypeID {
				totals[i] += int(it.Count)
				break
			}
		}
	}
	return totals
}

func planQuestItems(d *questDef, r questResult, original []db.Item) (items []db.Item, packets [][]byte, ok bool) {
	totals := questItemTotals(d, original)
	var remaining [3]int
	for i, f := range d.FarmItems {
		remaining[i] = int(f.Count)
		if r.NeedItemID == uint8(i+1) {
			remaining[i] = max(remaining[i], int(r.NeedItemCount))
		}
		if totals[i] < remaining[i] {
			return nil, nil, false
		}
	}
	items = slices.Clone(original)
	for i := len(items) - 1; i >= 0; i-- {
		it := items[i]
		if it.Bag < InventoryFirstBag || it.Bag > InventoryLastBag {
			continue
		}
		for k, f := range d.FarmItems {
			if f.Type == 0 || it.Type != f.Type || it.TypeID != f.TypeID {
				continue
			}
			take := remaining[k]
			if isQuestItemType(it.Type) {
				take = int(it.Count)
			}
			take = min(take, int(it.Count))
			if take <= 0 {
				break
			}
			remaining[k] -= take
			res := it
			res.Count -= uint16(take)
			items[i].Count = res.Count
			packets = append(packets, buildRemoveItem(&res, res.Count == 0))
			if res.Count == 0 {
				items = append(items[:i], items[i+1:]...)
			}
			break
		}
	}
	return items, packets, true
}

// handleQuestEnd (0x0903): npc id, quest id (quest_end_request).
func (s *Session) handleQuestEnd(p []byte) {
	if len(p) < 8 || s.char == nil {
		return
	}
	npcID, _ := packet.U32(p[2:])
	id, _ := packet.U16(p[6:])
	q := s.findQuest(id)
	if q == nil {
		return
	}
	d := q.def
	if d.EndType > 3 || (d.EndType == 0 && d.EndNpcType == 0) {
		return
	}
	if d.EndNpcType != 0 || npcID != 0 {
		if _, ok := s.npcMatches(npcID, d.EndNpcType, d.EndNpcID); !ok {
			return
		}
	}
	ctx := npcID
	if d.EndType >= 2 {
		ctx = 0
	}
	q.rewardNPC = npcID
	s.completeQuest(q, ctx, true, true)
}

// completeQuest is quest_requests_complete.
func (s *Session) completeQuest(q *activeQuest, ctx uint32, success, requested bool) {
	d := q.def
	if !success {
		if !s.removeQuest(q, 0) {
			return
		}
		w := packet.New(packet.QuestEnd)
		w.U32(ctx)
		w.U16(q.ID)
		w.Bytes(make([]byte, 14))
		s.send(w.MustFinish(), "0x0903 quest failed")
		s.log.Printf("GX Server World quest %d failed", q.ID)
		return
	}
	now := time.Now()
	if q.expired(now) {
		s.completeQuest(q, ctx, false, requested)
		return
	}
	totals, ready := s.questObjectives(q)
	if !ready {
		return
	}
	if now.Before(q.successAt) {
		if requested {
			s.questFeedback("The minimum quest time has not elapsed.")
		}
		return
	}
	eligible := func(i int) bool {
		return s.questRewardEligible(q, d.Results[i], totals, now)
	}
	switch {
	case d.ResultType == 1:
		if eligible(0) {
			s.awardQuest(q, 0, ctx)
			return
		}
	case d.UserSelect != 0:
		for i := 0; i < min(int(d.UserSelect), len(d.Results)); i++ {
			if eligible(i) {
				q.success = 1
				w := packet.New(packet.QuestChoose)
				w.U16(q.ID)
				s.send(w.MustFinish(), "0x0907 quest reward choice")
				return
			}
		}
	default:
		for i := len(d.Results) - 1; i >= 0; i-- {
			if (i == 0 || d.Results[i].hasRequirements()) && eligible(i) {
				s.awardQuest(q, uint8(i), ctx)
				return
			}
		}
	}
	if requested {
		s.questFeedback("No quest reward is available yet. Check the reward requirements.")
	}
}

// handleQuestChoose (0x0907): quest id, reward index.
func (s *Session) handleQuestChoose(p []byte) {
	if len(p) < 5 || s.char == nil {
		return
	}
	id, _ := packet.U16(p[2:])
	choice := p[4]
	q := s.findQuest(id)
	if q == nil || q.success == 0 || choice >= q.def.UserSelect || choice >= 3 {
		return
	}
	if q.rewardNPC != 0 {
		if _, ok := s.npcMatches(q.rewardNPC, q.def.EndNpcType, q.def.EndNpcID); !ok {
			return
		}
	}
	now := time.Now()
	if q.expired(now) {
		s.completeQuest(q, 0, false, true)
		return
	}
	totals, ready := s.questObjectives(q)
	if !ready || !s.questRewardEligible(q, q.def.Results[choice], totals, now) {
		s.questFeedback("You do not meet the requirements for this reward.")
		return
	}
	s.awardQuest(q, choice, 0)
}

// handleQuestAbandon (0x0908): quest id.
func (s *Session) handleQuestAbandon(p []byte) {
	if len(p) < 4 || s.char == nil {
		return
	}
	id, _ := packet.U16(p[2:])
	if q := s.findQuest(id); q != nil {
		s.completeQuest(q, 0, false, true)
	}
}

// awardQuest sends the credited EXP, authored gold and item through 0x0903.
func (s *Session) awardQuest(q *activeQuest, variant uint8, ctx uint32) {
	if s.char == nil || q == nil || s.findQuest(q.ID) != q || variant >= 3 {
		return
	}
	res := q.def.Results[variant]
	exp := config.ScaleEXP(res.Exp, rewardRates.QuestEXPPercent)
	change := db.CharacterChange{Before: *s.char, After: *s.char, ItemsBefore: s.items, ItemsAfter: append([]db.Item(nil), s.items...)}
	var packets [][]byte
	if q.def.EndType == 1 {
		var ok bool
		change.ItemsAfter, packets, ok = planQuestItems(q.def, res, s.items)
		if !ok {
			return
		}
	}
	var reward *db.Item
	if res.ItemType != 0 {
		def := db.MustDefs().Item(res.ItemType, res.ItemTypeID)
		if def == nil {
			s.log.Printf("GX Server World quest %d has invalid reward definition", q.ID)
			return
		}
		tpl := db.Item{Type: res.ItemType, TypeID: res.ItemTypeID, Quality: def.Quality, MakeType: "Q"}
		adds, left := planGiveItem(&change.ItemsAfter, &tpl, 1, def.MaxStack())
		if left != 0 {
			s.send(buildAddItemError(1), "quest reward inventory full")
			return
		}
		packets = append(packets, adds...)
		for i := len(change.ItemsAfter) - 1; i >= 0; i-- {
			it := &change.ItemsAfter[i]
			if it.Type == tpl.Type && it.TypeID == tpl.TypeID {
				reward = it
				break
			}
		}
	}
	change.After.Money = uint32(min(uint64(change.After.Money)+uint64(res.Money), db.MoneyCap))
	credited := applyExperienceAward(&change.After, exp, s.perfectParty())
	levels := change.After.Level - change.Before.Level
	if levels > 0 {
		st := StatsFor(&change.After, s.equipment())
		change.After.CurrentHP, change.After.CurrentMP, change.After.CurrentSP = st.MaxHP, st.MaxMP, st.MaxSP
	}
	completion := db.QuestCompletion{CharacterID: s.char.ID, QuestID: q.ID, Repeatable: q.def.TickReceive != 0}
	if err := db.CommitGameplay([]db.CharacterChange{change}, &completion); err != nil {
		s.log.Printf("GX Server World quest %d transaction failed: %v", q.ID, err)
		return
	}
	// Publishing is deliberately after commit: failures retain the active quest,
	// required items, character stats, gold and reward-choice state unchanged.
	*s.char = change.After
	s.items = change.ItemsAfter
	s.removeQuestFromMemory(q, 1)
	s.syncPlayer()
	s.partyPools()
	if s.player != nil {
		s.player.UpdateItems(s.items)
	}
	for _, p := range packets {
		s.send(p, "quest committed item")
	}
	if res.Money != 0 {
		s.send(buildSetMoney(s.char.Money), "0x0213 quest gold")
	}
	w := packet.New(packet.QuestEnd)
	w.U32(ctx)
	w.U16(q.ID)
	w.U8(1)
	w.U8(variant)
	w.U32(credited.base)
	w.U32(res.Money)
	if reward != nil {
		w.Bytes([]byte{reward.Bag, reward.Slot, reward.Type, reward.TypeID})
	} else {
		w.Bytes(make([]byte, 4))
	}
	s.send(w.MustFinish(), "0x0903 quest completed")
	s.notifyExpGain(credited, levels)
	s.log.Printf("GX Server World quest %d committed variant %d exp %d money %d", q.ID, variant, credited.base+credited.perfect, res.Money)
}

// removeQuest is QuestRuntime::remove: repeatable quests (TickReceive)
// are deleted, others finish with their success flag.
func (s *Session) removeQuest(q *activeQuest, success uint8) bool {
	var err error
	if q.def.TickReceive != 0 {
		err = db.DeleteCharacterQuest(s.char.ID, q.ID)
	} else {
		err = db.FinishCharacterQuest(s.char.ID, q.ID, success)
	}
	if err != nil {
		s.logErr("finish quest", err)
		return false
	}
	s.removeQuestFromMemory(q, success)
	return true
}

func (s *Session) removeQuestFromMemory(q *activeQuest, success uint8) {
	for i, active := range s.quests {
		if active == q {
			s.quests = append(s.quests[:i], s.quests[i+1:]...)
			break
		}
	}
	if q.def.TickReceive == 0 {
		s.finishedQuests = append(s.finishedQuests, db.FinishedQuest{QuestID: q.ID, Success: success})
	}
}

func (s *Session) bumpQuestCounter(q *activeQuest, slot int) {
	if q.counters[slot] == 255 {
		return
	}
	q.counters[slot]++
	if err := db.SaveCharacterQuest(s.char.ID, q.row()); err != nil {
		q.counters[slot]--
		s.logErr("save quest counter", err)
		return
	}
	w := packet.New(packet.QuestCounter)
	w.U16(q.ID)
	w.U8(uint8(slot))
	w.U8(q.counters[slot])
	s.send(w.MustFinish(), "0x0905 quest counter")
}

// questKill is update_mob_kill_quests for one kill of mobID.
func (s *Session) questKill(mobID uint16) { s.questKillCount(mobID, 1) }
func (s *Session) questKillCount(mobID uint16, count uint8) bool {
	if s.char == nil {
		return true
	}
	type update struct {
		quest    *activeQuest
		counters [3]uint8
	}
	var changes []update
	var rows []db.CharQuest
	var packets [][]byte
	now := time.Now()
	for _, id := range mobQuestLinks[mobID] {
		q := s.findQuest(id)
		if q == nil || q.def.EndType != 2 || (q.def.MinMinutes != 0 && now.After(q.successAt)) {
			continue
		}
		planned := *q
		changed := false
		for slot := 0; slot < 2; slot++ {
			if q.def.MobIDs[slot] != mobID {
				continue
			}
			value := uint8(min(uint16(planned.counters[slot])+uint16(count), 255))
			if value == planned.counters[slot] {
				continue
			}
			planned.counters[slot] = value
			changed = true
			w := packet.New(packet.QuestCounter)
			w.U16(q.ID)
			w.U8(uint8(slot))
			w.U8(value)
			packets = append(packets, w.MustFinish())
		}
		if changed {
			rows = append(rows, planned.row())
			changes = append(changes, update{q, planned.counters})
		}
	}
	if err := db.SaveCharacterQuests(s.char.ID, rows); err != nil {
		s.logErr("quest kill transaction", err)
		return false
	}
	for _, change := range changes {
		change.quest.counters = change.counters
	}
	for _, frame := range packets {
		s.send(frame, "0x0905 committed quest counter")
	}
	for _, change := range changes {
		q := change.quest
		if s.findQuest(q.ID) == q && q.def.EndNpcType == 0 && q.counters[0] >= q.def.MobCounts[0] && q.counters[1] >= q.def.MobCounts[1] {
			s.completeQuest(q, 0, true, false)
		}
	}
	return true
}

// questNeeds: an item that starts a quest, or one an active quest ends with.
func (s *Session) questNeeds(t, id uint8) bool {
	link, ok := questItemLinks[[2]uint8{t, id}]
	if !ok {
		return false
	}
	if len(link.Start) > 0 {
		return true
	}
	for _, qid := range link.End {
		if s.findQuest(qid) != nil {
			return true
		}
	}
	return false
}

// questPickup is quest_requests_progress_pick_item.
func (s *Session) questPickup(t, id uint8) {
	link, ok := questItemLinks[[2]uint8{t, id}]
	if !ok || s.char == nil {
		return
	}
	for _, qid := range link.End {
		q := s.findQuest(qid)
		if q == nil || q.def.EndType != 1 {
			continue
		}
		for slot, f := range q.def.FarmItems {
			if f.Type != t || f.TypeID != id {
				continue
			}
			s.bumpQuestCounter(q, slot)
			if q.def.EndNpcType == 0 && q.success == 0 {
				s.completeQuest(q, 0, true, false)
			}
			break
		}
	}
}

// questTimers fails timed quests whose time ran out.
func (s *Session) questTimers() {
	now := time.Now()
	for _, q := range append([]*activeQuest(nil), s.quests...) {
		if q.expired(now) {
			s.completeQuest(q, 0, false, false)
		} else if q.retryAutomatically() {
			s.completeQuest(q, 0, true, false)
		}
	}
}

// questItemDrop is mob_loot's quest item: Mobs.QuestItemID drops (chance =
// the item's Quality per cent) when it starts a quest or any recipient has
// a quest it ends.
func questItemDrop(mob *MapMob, recipients []*ConnectedPlayer) *db.ItemDef {
	if mob.def == nil || mob.def.QuestItemID <= 0 {
		return nil
	}
	key := db.MobQuestItem(mob.def.QuestItemID)
	t, id := key.Type, key.TypeID
	def := db.MustDefs().Item(t, id)
	if def == nil {
		return nil
	}
	link := questItemLinks[[2]uint8{t, id}]
	eligible := len(link.Start) > 0
	for _, p := range recipients {
		for _, qid := range link.End {
			eligible = eligible || p.hasQuest(qid)
		}
	}
	if eligible && rand.IntN(100) < int(def.Quality) {
		return def
	}
	return nil
}

// syncQuestIDs publishes the active quest ids for other sessions' loot.
func (s *Session) syncQuestIDs() {
	if s.player == nil {
		return
	}
	ids := make([]uint16, 0, len(s.quests))
	for _, q := range s.quests {
		ids = append(ids, q.ID)
	}
	s.player.setQuests(ids)
}
