package world

import (
	"math"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
)

func speedDefs(t *testing.T) {
	t.Helper()
	old := db.MustDefs()
	t.Cleanup(func() { db.SetDefsForTest(old) })
	db.SetDefsForTest(&db.Defs{Items: map[db.ItemKey]*db.ItemDef{
		{Type: 21, TypeID: 1}:   {Type: 21, TypeID: 1, Slot: 6},
		{Type: 30, TypeID: 106}: {Type: 30, TypeID: 106, Slot: 1, Speed: 1, AttackTime: 1},
		{Type: 30, TypeID: 107}: {Type: 30, TypeID: 107, Slot: 1, Speed: 2},
		{Type: 23, TypeID: 57}:  {Type: 23, TypeID: 57, Speed: 1},
	}})
}

func TestSonicEquipmentBuffsAndOriginalMaximum(t *testing.T) {
	speedDefs(t)
	s, _ := gmWireSession(t, db.Character{ID: 990060, CurrentHP: 100}, false)
	check := func(want uint8) {
		t.Helper()
		if sp := s.speeds(); sp.move != want {
			t.Fatalf("speed %+v expected %d", sp, want)
		}
	}
	check(2)
	s.items = []db.Item{{Bag: 1, Type: 21, TypeID: 1, Gems: [6]uint8{106}}}
	check(2)
	s.items[0].Bag = 0
	check(3)
	s.items[0].Gems[0] = 107
	check(4)
	s.items = append(s.items, db.Item{Bag: 0, Type: 23, TypeID: 57})
	check(4)
	// Ability 17 is a signed offset, shared by spells and item-cast buffs.
	haste := &db.SkillDef{SkillID: 114, SkillLevel: 1, TypeEffect: typeEffectBuff, TypeDetail: 82, KeepSeconds: 60}
	haste.AbilityType[0], haste.AbilityValue[0] = 17, 2
	if _, ok := s.addEffect(haste, 0, s.char.ID); !ok {
		t.Fatal("haste refused")
	}
	check(4)
	s.items = nil
	check(4)
	s.removeEffects(func(*ActiveBuff) bool { return true }, "test expiry")
	check(2)
	slow := *haste
	slow.SkillID = 999
	slow.TypeEffect = typeEffectDebuff
	slow.AbilityValue[0] = 1
	if _, ok := s.addEffect(&slow, 0); !ok {
		t.Fatal("slow refused")
	}
	check(1)
	s.removeEffects(func(*ActiveBuff) bool { return true }, "test expiry")
	check(2)
}

func TestServerSpeedSnapshotAndChangePublication(t *testing.T) {
	speedDefs(t)
	s, wire := gmWireSession(t, db.Character{ID: 990061, MapID: 60093, CurrentHP: 100}, false)
	s.player = RegisterPlayer(*s.char, nil, false, 0, s.out, nil)
	t.Cleanup(s.player.Unregister)
	next := func(want uint8) {
		t.Helper()
		select {
		case p := <-wire:
			sp, err := proto.DecodeCharacterSpeed(p)
			if err != nil || sp.Move != want || sp.ID != s.char.ID {
				t.Fatalf("speed %v: %v", sp, err)
			}
		case <-time.After(time.Second):
			t.Fatal("missing speed")
		}
	}
	s.syncPlayerSpeed(true)
	next(2)
	s.syncPlayerSpeed(false)
	select {
	case p := <-wire:
		t.Fatalf("unchanged speed resent %x", p)
	default:
	}
	s.items = []db.Item{{Bag: 0, Slot: 4, Type: 21, TypeID: 1, Gems: [6]uint8{106}}}
	s.sendChangedPools(s.maxPools(), "Sonic linking")
	next(3)
	if frame := <-wire; proto.Op(frame) != proto.CharacterAdditionalStats {
		t.Fatalf("missing additional stats: %x", frame)
	}
	state := game.New()
	state.SelfID = 990062
	for _, p := range s.player.enterFrames() {
		if _, err := state.Apply(p[2:]); err != nil {
			t.Fatal(err)
		}
	}
	if e := state.Entities[s.char.ID]; e == nil || e.Speed == nil || e.Speed.Move != 3 {
		t.Fatal("new viewer missed Sonic", e)
	}
	s.items = nil
	s.syncPlayerSpeed(false)
	next(2)
}

func TestSpeedBudgetAppliesEveryTierAndNeverTrustsPackets(t *testing.T) {
	speedDefs(t)
	for _, bonus := range []int32{-2, -1, 0, 1, 2, 100} {
		s, now := playerPhysicsFixture(t)
		s.buffs = []ActiveBuff{{Effect: typeEffectBuff, Detail: 82, Abilities: [3]buffAbility{{Type: 17, Value: max(bonus, 0)}}}}
		if bonus < 0 {
			s.buffs[0].Effect = typeEffectDebuff
			s.buffs[0].Abilities[0].Value = -bonus
		}
		start := s.char.PosZ
		for range 10 {
			now = now.Add(100 * time.Millisecond)
			_, _ = s.validatePlayerMovement(8, .05, 29, 1, 0, now)
		}
		limit := s.movementSpeed() + .2
		if delta := s.char.PosZ - start; math.Abs(float64(delta-limit)) > .015 {
			t.Fatalf("bonus %d: distance %g expected %g", bonus, delta, limit)
		}
		old := s.char.PosZ
		for range 100 {
			_, _ = s.validatePlayerMovement(8, .05, 29, 1, 0, now)
		}
		if s.char.PosZ > old+.01 {
			t.Fatal("flood accelerated")
		}
	}
	s, now := playerPhysicsFixture(t)
	s.buffs = []ActiveBuff{{Effect: typeEffectDebuff, Detail: detailStop}}
	old := s.movement.pos
	now = now.Add(time.Second)
	_, _ = s.validatePlayerMovement(8, .05, 29, 1, 0, now)
	if s.movement.pos != old {
		t.Fatal("blocked player used banked movement", s.movement.pos, old)
	}
}

func TestSeedSonicHasteCandyCooldownExpiryAndRelog(t *testing.T) {
	s, add := forgeRuleSession(t)
	defs, err := db.ReloadDefsForTest()
	if err != nil {
		t.Fatal(err)
	}
	for id, want := range map[uint8]uint8{106: 1, 107: 2} {
		if d := defs.Item(30, id); d == nil || d.Speed != want {
			t.Fatalf("seed Sonic %d: %+v", id, d)
		}
	}
	candy := defs.Item(25, 177)
	if candy == nil || candy.Range != 210 || candy.AttackTime != 1 {
		t.Fatal("Apple Candy skill mapping", candy)
	}
	haste := defs.Skill(114, 1)
	if haste == nil || haste.AbilityType[0] != 17 || haste.AbilityValue[0] != 2 {
		t.Fatal("seed haste", haste)
	}
	s.itemCooldowns = map[uint8]time.Time{}
	s.isAdmin, s.gmImmune = true, true // A protected GM can still consume a self buff.
	it := add(db.Item{Bag: 1, Slot: 0, Type: 25, TypeID: 177, Count: 3})
	s.handleUseItem(proto.UseItemPacket(1, 0))
	check := func(want uint8) {
		t.Helper()
		if sp := s.speeds(); sp.move != want {
			t.Fatalf("move=%d expected=%d", sp.move, want)
		}
	}
	check(3)
	if s.items[0].Count != 2 || len(s.buffs) != 1 || s.buffs[0].CasterUser != s.char.ID {
		t.Fatal("candy did not consume/apply self buff", s.items, s.buffs)
	}
	s.handleUseItem(proto.UseItemPacket(1, 0))
	if s.items[0].Count != 2 {
		t.Fatal("cooldown consumed candy")
	}
	s.persistBuffs()
	t.Cleanup(func() { db.SaveAppliedBuffs(s.char.ID, nil) })
	s.buffs, err = restoreBuffs(s.char.ID, s.log)
	if err != nil {
		t.Fatal(err)
	}
	check(3)
	s.gmImmune = false
	if _, ok := s.addEffect(haste, 0, s.char.ID); !ok {
		t.Fatal("spell failed")
	}
	check(4)
	// Removing the spell leaves the candy's tier; expiring the candy restores normal.
	s.removeEffects(func(b *ActiveBuff) bool { return b.SkillID == 114 }, "spell expiry")
	check(3)
	s.buffs[0].EndsAt = time.Now().Add(-time.Second)
	s.processBuffs()
	check(2)
	saved, err := db.CharacterItems(s.char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if i := findItemByUID(saved, it.UID); i < 0 || saved[i].Count != 2 {
		t.Fatal("candy quantity not persisted", saved)
	}
	// A weaker same-skill candy is refused before consumption or cooldown.
	sk := *defs.Skill(210, 1)
	sk.SkillLevel = 2
	if _, ok := s.addEffect(&sk, 0, s.char.ID); !ok {
		t.Fatal("stronger buff refused")
	}
	delete(s.itemCooldowns, candy.ReqIg)
	s.handleUseItem(proto.UseItemPacket(1, 0))
	if s.items[0].Count != 2 || len(s.itemCooldowns) != 0 || s.buffs[0].SkillLevel != 2 {
		t.Fatal("refused weaker candy mutated state")
	}
}

func TestSonicLinkAndExtractionRefreshEquippedSpeed(t *testing.T) {
	s, add := forgeRuleSession(t)
	defs, err := db.ReloadDefsForTest()
	if err != nil {
		t.Fatal(err)
	}
	boots := defs.Item(21, 53)
	if !lapisiaFits(defs.Item(30, 106), realType(boots.Type)) || lapisiaFits(defs.Item(30, 106), 20) {
		t.Fatal("native Sonic target flags")
	}
	add(db.Item{Bag: 0, Slot: 4, Type: 21, TypeID: 53, Count: 1})
	add(db.Item{Bag: 1, Slot: 0, Type: 30, TypeID: 106, Count: 2})
	link, extract := defs.Item(100, 192), defs.Item(100, 191)
	if !validForgeHammer(link, specialGemSuccessUp) || link.ReqVg != 100 ||
		!validForgeHammer(extract, specialGemExtract) || extractionChance(extract) != 100 {
		t.Fatal("stock GM hammers must guarantee a valid attempt", link, extract)
	}
	add(db.Item{Bag: 1, Slot: 1, Type: 100, TypeID: 192, Count: 1})
	s.syncPlayerSpeed(true)
	s.handleGemAdd(proto.GemAddHammerPacket(1, 0, 0, 4, 1, 1))
	if sp := s.speeds(); sp.move != 3 || s.lastSpeeds.move != 3 || s.items[0].Gems[0] != 106 {
		t.Fatal("equipped linking did not publish haste", sp, s.lastSpeeds, s.items)
	}
	add(db.Item{Bag: 1, Slot: 2, Type: 100, TypeID: 191, Count: 1})
	s.handleGemRemove(proto.GemRemoveHammerPacket([2]uint8{0, 4}, [2]uint8{1, 2}, 0))
	if sp := s.speeds(); sp.move != 2 || s.lastSpeeds.move != 2 || s.items[0].Gems[0] != 0 {
		t.Fatal("extraction did not remove/publish haste", sp, s.lastSpeeds, s.items)
	}
	saved, err := db.CharacterItems(s.char.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range saved {
		if it.Bag == 0 && it.Slot == 4 && it.Gems[0] != 0 {
			t.Fatal("extraction not persisted")
		}
	}
}

func TestReplacementBuffPublishesLowerSpeedImmediately(t *testing.T) {
	speedDefs(t)
	s, _ := gmWireSession(t, db.Character{ID: 990063, CurrentHP: 100}, false)
	def := &db.SkillDef{SkillID: 60000, SkillLevel: 4, TypeEffect: typeEffectBuff, TypeDetail: 82, KeepSeconds: 60}
	def.AbilityType[0], def.AbilityValue[0] = 17, 2
	s.addEffect(def, 0, s.char.ID)
	def.SkillLevel, def.AbilityValue[0] = 5, 1
	s.addEffect(def, 0, s.char.ID)
	if sp := s.speeds(); sp.move != 3 || s.lastSpeeds.move != 3 {
		t.Fatal("replacement compared against already-removed buff", sp, s.lastSpeeds)
	}
}

// Numerical tolerance must not become a per-packet distance allowance. A
// stream of individually tiny moves still spends the same server-time credit.
func TestSubTolerancePacketFloodCannotAccumulateFreeMovement(t *testing.T) {
	speedDefs(t)
	s, now := playerPhysicsFixture(t)
	start := s.char.PosZ
	for range 5000 {
		_, _ = s.validatePlayerMovement(s.char.PosX, .05, s.char.PosZ+.004, 1, 0, now)
	}
	if delta := s.char.PosZ - start; delta > .201 {
		t.Fatal("small packets bypassed budget", delta)
	}
	now = now.Add(100 * time.Millisecond)
	for range 5000 {
		_, _ = s.validatePlayerMovement(s.char.PosX, .05, s.char.PosZ+.004, 1, 0, now)
	}
	if delta := s.char.PosZ - start; delta > .701 {
		t.Fatal("small packets exceeded elapsed server time", delta)
	}
}

func TestRebirthRepublishesEquipmentSpeedAfterDeathState(t *testing.T) {
	speedDefs(t)
	s, wire := gmWireSession(t, db.Character{ID: 990063, MapID: 60093, CurrentHP: 100}, true)
	base := s.speeds()
	s.items = []db.Item{
		{Bag: 0, Slot: 4, Type: 21, TypeID: 1, Gems: [6]uint8{106}},
		{Bag: 0, Slot: 8, Type: 23, TypeID: 57},
	}
	state := game.New()
	state.SelfID = s.char.ID
	apply := func(frame []byte) {
		t.Helper()
		if _, err := state.Apply(frame); err != nil {
			t.Fatal(err)
		}
	}
	s.syncPlayerSpeed(true)
	apply(waitWorldPacket(t, wire, proto.CharacterDeathState))
	want := *state.Self.Speed
	if want.Move != 4 || want.Attack == base.attack {
		t.Fatalf("fixture did not apply both gear and lapis: %+v", want)
	}

	s.dead = true
	apply(buildCharacterDeathState(s.char.ID)[2:])
	apply(buildCharacterDeath(s.char.ID, deathTypeMob, 0)[2:])
	s.rebirth(s.char.MapID, s.char.PosX, s.char.PosY, s.char.PosZ, rebirthKindBind, true, 0)
	apply(waitWorldPacket(t, wire, proto.CharacterResurrect))
	apply(waitWorldPacket(t, wire, proto.CharacterDeathState))
	if state.Self.Dead || *state.Self.Speed != want {
		t.Fatalf("revived with incorrect speed: %+v, want %+v", state.Self.Speed, want)
	}
}
