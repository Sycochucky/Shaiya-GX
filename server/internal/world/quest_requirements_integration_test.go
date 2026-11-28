//go:build integration

package world

import (
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
)

func TestQuestConfiguredStartGuardsCostAndRepeatInterval(t *testing.T) {
	s, add := forgeRuleSession(t)
	s.char.Cg, s.char.Og, s.char.Ig = 2, 3, 4
	s.player = &ConnectedPlayer{character: *s.char}
	partner := &ConnectedPlayer{character: db.Character{Class: 5}}
	item := add(db.Item{Bag: 1, Slot: 10, Type: 29, TypeID: 7, Count: 2})
	db.MustDefs().Items[db.ItemKey{Type: 29, TypeID: 7}] = &db.ItemDef{Type: 29, TypeID: 7, Count: 5}

	defs, links := questDefs, questItemLinks
	t.Cleanup(func() { questDefs, questItemLinks = defs, links })
	questDefs = []questDef{{
		ID: 1, StartType: 4, StartItemType: 29, StartItemID: 7, Faction: 6,
		Male: true, MaxLevel: 60, Jobs: [6]uint8{1},
		CG: 2, OG: 3, IG: 4, RequireParty: true, PartyJobs: [6]uint8{1, 0, 0, 0, 0, 1},
		TickStart: 360, TickReceive: 1,
	}}
	questItemLinks = map[[2]uint8]questItemLink{{29, 7}: {Start: []uint16{1}}}
	request := proto.QuestStartPacket(7, 1)
	s.handleUseItem(proto.UseItemPacket(1, 10))
	s.handleQuestStart(request)
	if s.findQuest(1) != nil || s.items[0].Count != 2 {
		t.Fatal("solo request started a party quest or charged its letter")
	}
	s.player.party = &party{members: []*ConnectedPlayer{s.player, partner}}
	partner.character.Class = 4
	s.handleQuestStart(request)
	if s.findQuest(1) != nil || s.items[0].Count != 2 {
		t.Fatal("party with wrong class started the quest")
	}
	partner.character.Class = 5
	for _, stat := range []*uint8{&s.char.Cg, &s.char.Og, &s.char.Ig} {
		*stat--
		s.handleQuestStart(request)
		if s.findQuest(1) != nil || s.items[0].Count != 2 {
			t.Fatal("unmet character flag charged the letter or started the quest")
		}
		*stat++
	}
	s.char.Gender = 1
	s.handleQuestStart(request)
	if s.findQuest(1) != nil || s.items[0].Count != 2 {
		t.Fatal("wrong sex started the quest")
	}
	s.char.Gender = 0
	s.handleQuestStart(request)
	if s.findQuest(1) == nil || s.items[0].Count != 1 {
		t.Fatal("qualifying party did not start the quest and consume one letter")
	}
	if !s.player.hasQuest(1) {
		t.Fatal("new quest was not eligible immediately after commit")
	}
	s.handleQuestAbandon(proto.QuestAbandonPacket(1))
	if s.player.hasQuest(1) {
		t.Fatal("abandoned quest stayed eligible after commit")
	}
	if err := s.loadQuests(); err != nil {
		t.Fatal(err)
	}
	s.handleUseItem(proto.UseItemPacket(1, 10))
	s.handleQuestStart(request)
	if s.findQuest(1) != nil || s.items[0].Count != 1 {
		t.Fatal("reconnect bypassed start interval or consumed the letter")
	}
	pool, _ := db.Pool(db.GameData)
	if _, err := pool.Exec("UPDATE GX_QuestStartTerms SET AvailableAt=0 WHERE CharID=? AND QuestID=1", s.char.ID); err != nil {
		t.Fatal(err)
	}
	s.handleQuestStart(request)
	if s.findQuest(1) == nil || findItemByUID(s.items, item.UID) >= 0 {
		t.Fatal("quest did not reopen once the start interval elapsed")
	}
}

func TestQuestConfiguredRewardChecksTimeAttributesAndItemsAtChoice(t *testing.T) {
	s, add := forgeRuleSession(t)
	add(db.Item{Bag: 1, Slot: 10, Type: 1, TypeID: 1, Count: 5})
	def := &questDef{ID: 1, EndType: 1, UserSelect: 2,
		FarmItems: [3]questItem{{Type: 1, TypeID: 1, Count: 2}},
		Results: [3]questResult{
			{Money: 100},
			{NeedItemID: 1, NeedItemCount: 4, NeedTime: 20, NeedHG: 3, NeedVG: -2, NeedOG: 4, Money: 200},
		},
	}
	q := &activeQuest{ID: 1, def: def, startAt: time.Now()}
	s.quests = []*activeQuest{q}
	if err := db.AddCharacterQuest(s.char.ID, db.CharQuest{QuestID: 1}); err != nil {
		t.Fatal(err)
	}
	s.char.Hg, s.char.Vg, s.char.Og = 3, -2, 4
	s.handleQuestEnd(proto.QuestEndPacket(0, 1))
	if q.success == 0 {
		t.Fatal("the baseline reward should be available")
	}
	pick := proto.QuestChoosePacket(1, 1)
	s.handleQuestChoose(pick)
	if s.findQuest(1) == nil || s.items[0].Count != 5 {
		t.Fatal("reward ignored its elapsed time")
	}
	q.startAt = q.startAt.Add(-time.Minute)
	for _, stat := range []*int16{&s.char.Hg, &s.char.Vg} {
		*stat--
		s.handleQuestChoose(pick)
		if s.findQuest(1) == nil || s.items[0].Count != 5 {
			t.Fatal("reward ignored HG/VG")
		}
		*stat++
	}
	s.char.Og--
	s.handleQuestChoose(pick)
	if s.findQuest(1) == nil || s.items[0].Count != 5 {
		t.Fatal("reward ignored OG")
	}
	s.char.Og++
	for range 2 {
		if _, err := s.consumeOne(0); err != nil {
			t.Fatal(err)
		}
	}
	s.handleQuestChoose(pick)
	if s.findQuest(1) == nil || s.items[0].Count != 3 {
		t.Fatal("reward ignored item removal after opening the choice")
	}
	add(db.Item{Bag: 1, Slot: 11, Type: 1, TypeID: 1, Count: 2})
	money := s.char.Money
	s.handleQuestChoose(pick)
	if s.findQuest(1) != nil || s.items[0].Count != 1 || s.char.Money != money+200 {
		t.Fatal("eligible variant failed to persist its selected quantity and reward")
	}
}

func TestQuestTimedAutomaticRewardRetriesWhenReady(t *testing.T) {
	s, add := forgeRuleSession(t)
	add(db.Item{Bag: 1, Slot: 10, Type: 1, TypeID: 1, Count: 1})
	def := &questDef{ID: 1, EndType: 1, ResultType: 1,
		FarmItems: [3]questItem{{Type: 1, TypeID: 1, Count: 1}},
		Results:   [3]questResult{{Money: 100, NeedTime: 20}},
	}
	s.startQuest(def, 1, 0)
	q := s.findQuest(1)
	if q == nil {
		t.Fatal("quest did not start")
	}
	s.questTimers()
	if s.findQuest(1) == nil || len(s.items) != 1 {
		t.Fatal("automatic reward bypassed NeedTime")
	}
	q.startAt = q.startAt.Add(-time.Minute)
	s.questTimers()
	if s.findQuest(1) != nil || len(s.items) != 0 || len(s.finishedQuests) != 1 || s.finishedQuests[0].Success != 1 {
		t.Fatal("automatic item quest was not retried after NeedTime")
	}
}

func TestQuestTickKeepRestoresElapsedTimeAndExpiresBeforeReward(t *testing.T) {
	s, _ := forgeRuleSession(t)
	defs := questDefs
	t.Cleanup(func() { questDefs = defs })
	questDefs = []questDef{{ID: 1, EndType: 2, MobCounts: [2]uint8{1}, TickKeep: 8, MaxMinutes: 60,
		Results: [3]questResult{{Money: 500}},
	}}
	started := time.Now().Add(-9 * time.Second).Truncate(time.Millisecond)
	if err := db.CommitQuestStart(s.char.ID, db.CharQuest{QuestID: 1, StartedAt: started.UnixMilli(), Counts: [3]uint8{1}}, nil, nil, 0); err != nil {
		t.Fatal(err)
	}
	if err := s.loadQuests(); err != nil {
		t.Fatal(err)
	}
	q := s.findQuest(1)
	if q == nil || !q.startAt.Equal(started) || !q.endAt.Equal(started.Add(8*time.Second)) {
		t.Fatalf("lost exact start or shorter TickKeep deadline: %+v", q)
	}
	money := s.char.Money
	s.handleQuestEnd(proto.QuestEndPacket(0, 1))
	if s.findQuest(1) != nil || s.char.Money != money || len(s.finishedQuests) != 1 || s.finishedQuests[0].Success != 0 {
		t.Fatal("expired quest delivered a reward")
	}
}
