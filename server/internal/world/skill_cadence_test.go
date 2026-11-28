package world

import (
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/skilltiming"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
	"testing"
	"time"
)

func cadenceSession(t *testing.T, ready uint8) *Session {
	// These tests pin the table ceiling; animation-bound cadence has its own test.
	skilltiming.AnimationBound = false
	t.Cleanup(func() { skilltiming.AnimationBound = true })
	s, d, _ := spellCastRangeFixture(t)
	s.player = RegisterPlayer(*s.char, nil, false, s.faction, s.out, s.logger)
	t.Cleanup(s.player.Unregister)
	d.TypeEffect, d.TypeDetail, d.TargetType = typeEffectHeal, detailHeal, targetTypeSelf
	d.ReadyTime, d.ResetTimeSeconds, d.HealHP = ready, 1, 5
	second := *d
	second.SkillID++
	db.MustDefs().Skills[db.SkillKey{SkillID: second.SkillID, Level: 1}] = &second
	s.skills = append(s.skills, db.CharacterSkill{SkillID: second.SkillID, Level: 1, Number: 2})
	weapon := *db.MustDefs().Item(1, 1)
	weapon.AttackTime = 5
	db.MustDefs().Items[db.ItemKey{Type: 1, TypeID: 1}] = &weapon
	s.items = []db.Item{{Bag: 0, Slot: weaponSlot, Type: 1, TypeID: 1, Count: 1}}
	s.char.CurrentHP = 1
	return s
}
func TestSharedCadenceRejectsAlternatingSpellAndBasicFlood(t *testing.T) {
	s := cadenceSession(t, 0)
	cast := func(slot uint8) {
		s.handleUseSkill(packet.SkillUseNoTarget, proto.UseSkillPacket(slot, s.char.ID, true))
	}
	before := time.Now()
	cast(1)
	if s.char.CurrentHP <= 1 || s.lastActionAt.Before(before) {
		t.Fatal("accepted skill did not reserve shared action clock")
	}
	hp := s.char.CurrentHP
	held := time.Now().Add(time.Hour)
	s.lastActionAt = held
	for i := range 100 {
		cast(uint8(i%2 + 1))
		if !s.autoAttackSwing(987654) || !s.userAttackSwing(987654) {
			t.Fatal("basic attack reached target lookup during shared recovery")
		}
	}
	if s.char.CurrentHP != hp || s.lastActionAt != held {
		t.Fatal("alternating actions bypassed or extended recovery")
	}
	interval, _ := skilltiming.Interval(s.speeds().attack)
	s.lastActionAt = time.Now().Add(-interval)
	cast(2)
	if s.char.CurrentHP <= hp {
		t.Fatal("different skill did not resume after native recovery")
	}
}
func TestSharedCombatCadenceUsesEveryNativeTierAndLiveHaste(t *testing.T) {
	s := cadenceSession(t, 0)
	now := time.Now()
	s.lastActionAt = now
	for tier := uint8(0); tier <= 9; tier++ {
		s.buffs = []ActiveBuff{{Effect: typeEffectBuff, Abilities: [3]buffAbility{{Type: 16, Value: int32(tier) - 5}}}}
		interval, ok := skilltiming.Interval(tier)
		if !ok || s.speeds().attack != tier || s.combatActionReady(now.Add(interval-time.Nanosecond)) || !s.combatActionReady(now.Add(interval)) {
			t.Fatal("wrong shared recovery", tier, interval, s.speeds())
		}
	}
	s.buffs = nil
	if s.combatActionReady(now.Add(time.Second)) {
		t.Fatal("normal-speed recovery too short")
	}
	s.buffs = []ActiveBuff{{Effect: typeEffectBuff, Abilities: [3]buffAbility{{Type: 16, Value: 4}}}}
	if !s.combatActionReady(now.Add(time.Second)) {
		t.Fatal("live haste did not accelerate recovery")
	}
	s.buffs = []ActiveBuff{{Effect: typeEffectDebuff, Detail: detailStun}}
	if s.combatActionReady(now.Add(time.Hour)) {
		t.Fatal("blocked attack speed allowed a spell")
	}
}
func TestPreparedReleaseDoesNotChargeASecondAttackInterval(t *testing.T) {
	s := cadenceSession(t, 1)
	known := &s.skills[0]
	def := *db.MustDefs().Skill(known.SkillID, known.Level)
	def.SkillID = 29 // Magic Arrow has a generated clip for its instant rank.
	known.SkillID = def.SkillID
	db.MustDefs().Skills[db.SkillKey{SkillID: def.SkillID, Level: known.Level}] = &def
	s.handleUseSkill(packet.SkillUseNoTarget, proto.UseSkillPacket(1, s.char.ID, true))
	if s.cast == nil || s.lastActionAt.IsZero() {
		t.Fatal("prepared cast not started")
	}
	start := s.lastActionAt
	s.cast.readyAt = time.Now().Add(-time.Second)
	s.processPendingCast()
	if s.cast != nil || s.char.CurrentHP <= 1 || s.lastActionAt != start || s.ownClip() != 0 {
		t.Fatal("prepared release restarted attack recovery")
	}
	interval, _ := skilltiming.Interval(s.speeds().attack)
	if !s.combatActionReady(start.Add(interval)) {
		t.Fatal("prepared cast charged two intervals")
	}

	instant := def
	instant.SkillLevel, instant.ReadyTime = 3, 0
	known.Level = instant.SkillLevel
	db.MustDefs().Skills[db.SkillKey{SkillID: instant.SkillID, Level: instant.SkillLevel}] = &instant
	s.lastActionAt = time.Now().Add(-2 * interval)
	s.handleUseSkill(packet.SkillUseNoTarget, proto.UseSkillPacket(1, s.char.ID, true))
	if s.cast != nil || s.ownClip() != 733*time.Millisecond {
		t.Fatalf("instant rank should apply Magic Arrow's clip, got %v", s.ownClip())
	}
}

func TestEarlySkillRequestIsHeldAndRunsWhenRecoveryEnds(t *testing.T) {
	s := cadenceSession(t, 0)
	cast := func(slot uint8) {
		s.handleUseSkill(packet.SkillUseNoTarget, proto.UseSkillPacket(slot, s.char.ID, true))
	}
	interval, _ := skilltiming.Interval(s.speeds().attack)
	s.lastActionAt = time.Now().Add(-interval + 50*time.Millisecond)
	hp := s.char.CurrentHP
	cast(1)
	cast(2) // the latest early selection replaces the earlier one
	if s.char.CurrentHP != hp || s.pendingSkill == nil {
		t.Fatal("early request was not held")
	}
	s.processPendingSkill()
	if s.char.CurrentHP != hp || s.pendingSkill == nil {
		t.Fatal("held request ran before recovery ended")
	}
	s.lastActionAt = time.Now().Add(-interval)
	s.processPendingSkill()
	if s.char.CurrentHP <= hp || s.pendingSkill != nil {
		t.Fatal("held request did not run when recovery ended")
	}
}

// Animation-bound cadence: the next action is due when the basic-attack clip at
// the weapon-relative animation rate ends, never later than the table.
func TestAnimationBoundIntervalWakesAtExactDeadline(t *testing.T) {
	s := cadenceSession(t, 0)
	skilltiming.AnimationBound = true
	old := db.MustDefs()
	t.Cleanup(func() { db.SetDefsForTest(old) })
	db.SetDefsForTest(&db.Defs{Items: map[db.ItemKey]*db.ItemDef{{Type: 1, TypeID: 1}: {Type: 1, TypeID: 1, AttackTime: 5}}})
	s.items = []db.Item{{Bag: 0, Slot: 5, Type: 1, TypeID: 1}}
	s.buffs = []ActiveBuff{{Effect: typeEffectBuff, Abilities: [3]buffAbility{{Type: 16, Value: 4}}}}
	interval, ok := s.actionInterval()
	if !ok || interval.Milliseconds() != 666 {
		t.Fatal("tier 9 with a base 5 weapon must follow the 45 fps clip", interval)
	}
	now := time.Now()
	s.lastActionAt = now.Add(-interval + 30*time.Millisecond)
	s.pendingSkill = &pendingSkillUse{expires: now.Add(time.Second)}
	if w := s.nextWake(); w > 35*time.Millisecond || w < minWake {
		t.Fatal("the loop would sleep past the held skill's deadline", w)
	}
	s.pendingSkill = nil
	if w := s.nextWake(); w != loopTick {
		t.Fatal("idle loop must keep the regular tick", w)
	}
}

// A skill's own animation is a floor for the next action, whatever the table
// interval says; a later action stamp makes it stale.
func TestOwnSkillClipHoldsTheNextAction(t *testing.T) {
	s := cadenceSession(t, 0)
	interval, _ := skilltiming.Interval(s.speeds().attack)
	if interval >= time.Second {
		t.Skip("table interval already covers the clip")
	}
	s.lastActionAt = time.Now().Add(-interval - 50*time.Millisecond)
	if !s.combatActionReady(time.Now()) {
		t.Fatal("recovered interval blocked without a clip floor")
	}
	s.clipFloorAt, s.clipFloor = s.lastActionAt, time.Second
	if s.combatActionReady(time.Now()) {
		t.Fatal("next action started before the skill's clip ended")
	}
	s.lastActionAt = s.lastActionAt.Add(-time.Second)
	if !s.combatActionReady(time.Now()) {
		t.Fatal("clip floor stayed after the clip ended")
	}
	s.lastActionAt = time.Now().Add(-interval - 50*time.Millisecond)
	if !s.combatActionReady(time.Now()) {
		t.Fatal("a stale floor from an earlier action blocked a later one")
	}
	if skilltiming.ClipFloor(55) != time.Second {
		t.Fatal("Rising Crash clip floor missing from the generated table")
	}
}
