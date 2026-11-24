package world

import (
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/attackspeed"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/skilltiming"
	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

const (
	weaponSlot            = 5
	autoAttackRangeGrace  = 2.0
	autoAttackCancelRange = 80.0
	attackSuccessNormal   = 0
	attackResultCritical  = 1
	attackResultMiss      = 3
	deathTypeMob          = 0x02
)

type weaponCombat struct {
	ItemType, TypeID uint8
	AttackTime       uint8
	Range            float32
	Delay            time.Duration
}

// weaponAttackDelay takes the native zero-based attack index, not index+1.
func weaponAttackDelay(tier, weaponBase uint8) (time.Duration, bool) {
	return skilltiming.IntervalFor(tier, weaponBase, true)
}

func equippedWeaponCombat(items []db.Item) (weaponCombat, bool) {
	w := equippedItem(items, weaponSlot)
	if w == nil {
		return weaponCombat{}, false
	}
	d := db.MustDefs().Item(w.Type, w.TypeID)
	if d == nil {
		return weaponCombat{}, false
	}
	if d.AttackTime == 0 {
		return weaponCombat{}, false
	}
	at := d.AttackTime
	delay, ok := weaponAttackDelay(attackspeed.Tier(at, 0), at)
	if !ok {
		return weaponCombat{}, false
	}
	r := float32(d.Range)
	if r <= 0 {
		r = 2
	}
	return weaponCombat{d.Type, d.TypeID, at, r, delay}, true
}

func sendNear(mapID uint16, x, z, radius float32, frame []byte, label string) {
	for _, p := range PlayersNear(mapID, x, z, radius) {
		p.Send(frame, label)
	}
}

// autoAttackSwing resolves one swing; false cancels the auto-attack.
func (s *Session) autoAttackSwing(target uint32) bool {
	ch := s.char
	if ch == nil {
		return false
	}
	s.cancelVehicleCast()
	if s.cast != nil || s.dash != nil {
		return false
	}
	if s.ridingVehicle() {
		s.log.Printf("GX Server World cancelled auto attack while mounted character_id=%d mob_global_id=%d", ch.ID, target)
		return false
	}
	if s.cannotAct() {
		return false // CUser::GetAttackSpeed is 255 while stunned or asleep
	}
	if s.safeZoneProtection() == safeBoth {
		s.sendSkillMessage(0, proto.MsgCannotAttack)
		return false
	}
	wc, ok := equippedWeaponCombat(s.items)
	if !ok {
		s.log.Printf("GX Server World cancelled auto attack without valid weapon character_id=%d mob_global_id=%d", ch.ID, target)
		return false
	}
	stats := StatsFor(ch, s.equipment())
	// CUser::GetAttackSpeedNow: the swing index moves with gems, passives
	// and buffs (0..9 over g_ConstUserAttackSpeed).
	wc.Delay, _ = weaponAttackDelay(attackspeed.Tier(wc.AttackTime, stats.AttackSpeed), wc.AttackTime)
	if !s.lastActionAt.IsZero() && time.Since(s.lastActionAt) < max(wc.Delay, s.ownClip()) {
		return true
	}
	mob, ok := MobByID(ch.MapID, target)
	if !ok || mob.CurrentHP == 0 {
		s.log.Printf("GX Server World cancelled auto attack unknown or dead mob character_id=%d mob_global_id=%d", ch.ID, target)
		return false
	}
	// CUser::EnableAttackRange: weapon reach + mob size + 2.3, plus the
	// movement allowance (at least 2 units).
	size := float32(0)
	if md := db.MustDefs().Mobs[mob.MobID]; md != nil {
		size = float32(md.Size)
	}
	reach := wc.Range + float32(stats.Range) + size + 2.3 + autoAttackRangeGrace
	distance := DistanceXZ(ch.PosX, ch.PosZ, mob.PosX, mob.PosZ)
	if distance > reach {
		if distance > autoAttackCancelRange {
			s.log.Printf("GX Server World cancelled auto attack target too far character_id=%d mob_global_id=%d distance=%.2f", ch.ID, target, distance)
			return false
		}
		return true
	}
	def := db.MustDefs().Mobs[mob.MobID]
	if def == nil {
		return false
	}
	if mobFriendly(def, s.faction) {
		s.sendSkillMessage(0, 317)
		return false
	}
	kind := attackMelee
	if stats.RangedWeapon {
		kind = attackRanged
	}
	if stats.Disabled[kind] > 0 {
		return true // the category is disabled (abilities 32-34); wait it out
	}
	if !s.attackMovementAllowed() {
		return false
	}
	s.commitAttackStop(time.Now())
	s.breakCombatForm()
	s.breakStealth()
	s.lastActionAt, s.lastCombatAt = time.Now(), time.Now()
	if speed := wakeMob(ch.MapID, target); speed != nil {
		sendNear(ch.MapID, mob.PosX, mob.PosZ, BroadcastRadius, speed, "0x0609 mob speed (woke)")
	}
	mc, _ := MobCombatByID(ch.MapID, target)
	// 0x0503 result is 0 normal, 1 critical, 3 miss.
	if !playerHits(&stats, &mc, kind) {
		sendNear(ch.MapID, mob.PosX, mob.PosZ, CombatRadius,
			buildCharacterMobAttack(ch.ID, mob.ID, attackResultMiss, 0), "0x0503 character mob auto attack (miss)")
		return true
	}
	damage, crit := playerDamage(&stats, &mc, kind)
	result, ok := DamageMobFrom(ch.MapID, target, uint32(damage), ch.ID)
	if !ok {
		return false
	}
	AddMobAggro(ch.MapID, target, ch.ID, hitAggro(&stats, 0))
	code := uint8(attackSuccessNormal)
	if crit {
		code = attackResultCritical
	}
	s.log.Printf("GX Server World auto attack mob_global_id=%d mob_id=%d kind=%d damage=%d crit=%t hp=%d/%d killed=%t",
		result.Mob.ID, result.Mob.MobID, kind, result.DamageHP, crit, result.Mob.CurrentHP, result.Mob.MaxHP, result.Killed)
	sendNear(ch.MapID, result.Mob.PosX, result.Mob.PosZ, CombatRadius,
		buildCharacterMobAttack(ch.ID, result.Mob.ID, code, result.DamageHP), "0x0503 character mob auto attack")
	alive := s.resolveMobDamage(&result)
	if alive {
		s.gemCastsOnMob(result.Mob.ID, weaponGemSkills(s.items)) // AttackItemOption(CMob)
	}
	return alive
}

// resolveMobDamage handles the aftermath of any damage to a mob: death,
// drops and exp, or an HP sync. Returns whether the mob is still alive.
func (s *Session) resolveMobDamage(r *MobDamageResult) bool {
	ch := s.char
	m := &r.Mob
	if r.Speed != nil {
		sendNear(ch.MapID, m.PosX, m.PosZ, BroadcastRadius, r.Speed, "0x0609 mob speed (woke)")
	}
	if !r.Killed {
		sendNear(ch.MapID, m.PosX, m.PosZ, CombatRadius, buildMobCurrentHP(m.ID, m.CurrentHP), "0x0610 mob current hp after attack")
		s.send(buildMobTargetHP(m), "0x0305 mob target hp after attack")
		return true
	}
	sendMobDeath(ch.MapID, m, mobDeathCredit(m, ch.ID), "0x0606 mob death")
	rewardKill(s.player, ch.ID, ch.MapID, m, s.logger)
	return false
}

func sendMobDeath(mapID uint16, m *MapMob, killerID uint32, label string) {
	if stop := m.deathPositionPacket(); stop != nil {
		sendNear(mapID, m.PosX, m.PosZ, CombatRadius, stop, "mob death position")
	}
	sendNear(mapID, m.PosX, m.PosZ, CombatRadius, buildMobDeath(m.ID, killerID), label)
}

// applyIncomingMobDamage applies one queued mob attack; true when it
// killed. Slot attacks with a skill (Attacktype = a level-100 Skills row)
// go through CMob::AttackUser(skill); others are CMob::Attack.
func (s *Session) applyIncomingMobDamage(d pendingDamage) bool {
	if d.Guard != nil {
		return s.applyGuardStrike(*d.Guard)
	}
	if s.invulnerable() || (s.player != nil && d.First && s.player.noMobTarget.Load()) {
		return false
	}
	ch := s.char
	if ch == nil {
		return false
	}
	equip := s.equipment()
	stats := StatsFor(ch, equip)
	if stats.CurrentHP == 0 {
		return false
	}
	s.lastCombatAt = time.Now()
	def := db.MustDefs().Mobs[d.MobID]
	if def == nil {
		return false
	}
	mc, ok := MobCombatByID(ch.MapID, d.MobGlobalID)
	if !ok {
		mc = mobCombatFor(def)
	}
	slot := &def.Attacks[min(d.Slot, 2)]
	skillID := d.SkillID
	if skillID == 0 {
		skillID = slot.Type
	}
	if skillID != 0 {
		if sk := db.MustDefs().NPCSkill(skillID); sk != nil {
			return s.applyMobSkill(d, def, &mc, sk, &stats)
		}
	}
	// CMob::AttackSuccess / CMob::AttackDamage: normal attacks are melee.
	var damage uint32
	result := uint8(attackSuccessNormal)
	if s.blocksAttack(attackMelee) || !mobHits(&mc, &stats, attackMelee) {
		result = attackResultMiss
	} else {
		duelHitBy(s.player, 0) // CMob::Attack -> PvPChkLeaveByAttack
		v, crit := mobDamage(&mc, slot, &stats)
		damage = uint32(max(v, 0))
		if stats.Absorption > 0 {
			damage = max(satSub(damage, stats.Absorption), 1)
		}
		if crit {
			result = attackResultCritical
		}
	}
	// CMob::Attack: a reflected hit spares the player and goes back to
	// the mob (0x0605 then carries no damage).
	reflected := result != attackResultMiss && s.reflects(attackMelee, def)
	killed := false
	shown := damage
	if reflected {
		shown = 0
	} else {
		killed = s.takeMobDamage(d, damage, &stats)
		shown = s.assisted[0] // Health Assistant moves part of it to SP / MP
	}
	sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius,
		buildMobAttackCharacter(result, d.MobGlobalID, ch.ID, uint16(min(shown, 65535))), "0x0605 mob attack character")
	if reflected {
		s.reflectToMob(d.MobGlobalID, [3]uint16{uint16(min(damage, 65535)), 0, 0})
	}
	if result != attackResultMiss && !killed {
		s.gemCastsOnMob(d.MobGlobalID, armourGemSkills(s.items)) // BeAttackedItemOption(CMob)
	}
	s.log.Printf("GX Server World mob attack mob_global_id=%d mob_id=%d slot=%d result=%d damage=%d killed=%t",
		d.MobGlobalID, d.MobID, d.Slot, result, damage, killed)
	s.afterMobDamage(d, killed)
	return killed
}

// applyMobSkill sends the cast with its animation slot (0x060B or 0x060C)
// and any subsequent results without a slot (0x060D).
func (s *Session) applyMobSkill(d pendingDamage, md *db.MobDef, mc *mobCombat, sk *db.SkillDef, stats *CharacterStats) bool {
	ch := s.char
	mob, _ := MobByID(ch.MapID, d.MobGlobalID)
	hits := 1
	if sk.TypeDetail == detailMulti {
		hits = int(sk.MultiAttack)
	}
	killed := false
	for i := 0; i < hits && !killed; i++ {
		first := d.First && sk.TypeDetail != detailMulti
		drain := first && sk.TypeDetail == detailDrain
		var damage [3]int32
		result := uint8(attackSuccessNormal)
		if s.blocksAttack(int(sk.Kind)) || s.immuneToLapisSkill(sk) || s.blocksStatusSkill(sk) || !mobSkillHits(mc, sk, stats) {
			result = attackResultMiss
		} else {
			duelHitBy(s.player, 0)
			var crit bool
			damage, crit = mobSkillDamage(mc, md, sk, stats, mob.CurrentHP)
			if crit {
				result = attackResultCritical
			}
		}
		if damage[0] > 0 || damage[1] > 0 || damage[2] > 0 {
			s.breakCombatForm()
			s.breakStealth()
		}
		hp := uint32(max(damage[0], 0))
		if hp > 0 && stats.Absorption > 0 {
			hp = max(satSub(hp, stats.Absorption), 1)
		}
		// CMob::AttackUser: a landed skill goes through CUser::ChkMirror for its
		// category; a mirrored hit spares the player and hurts the mob instead.
		mirrored := result != attackResultMiss && s.reflects(min(int(sk.Kind), 2), md)
		var back [3]uint16
		if mirrored {
			back = [3]uint16{uint16(min(hp, 65535)), uint16(min(max(damage[1], 0), 65535)), uint16(min(max(damage[2], 0), 65535))}
			hp, damage[1], damage[2] = 0, 0, 0
		}
		if damage[1] > 0 || damage[2] > 0 {
			s.spendDelta(uint32(max(damage[2], 0)), uint32(max(damage[1], 0)))
			if hp == 0 {
				s.persist()
			}
		}
		if !mirrored {
			killed = s.takeMobDamage(d, hp, stats)
		}
		if hp > 0 {
			hp = s.assisted[0]
			damage[1], damage[2] = damage[1]+int32(s.assisted[1]), damage[2]+int32(s.assisted[2])
		}
		words := [3]uint16{uint16(min(hp, 65535)), uint16(min(max(damage[1], 0), 65535)), uint16(min(max(damage[2], 0), 65535))}
		var frame []byte
		switch {
		case drain:
			var healed [3]uint16
			if mirrored {
				// the mob is not healed by a mirrored drain
			} else if r, gain, ok := HealMob(ch.MapID, d.MobGlobalID, hp); ok {
				healed[0] = uint16(min(gain, 65535))
				sendNear(ch.MapID, r.PosX, r.PosZ, CombatRadius, buildMobCurrentHP(r.ID, r.CurrentHP), "0x0610 mob hp after drain")
			}
			frame = buildMobSkillPacket(0x060C, result, d.MobGlobalID, ch.ID, sk, words, &healed, d.Slot)
		case i == 0 && (d.First || d.StartsCast):
			frame = buildMobSkillPacket(0x060B, result, d.MobGlobalID, ch.ID, sk, words, nil, d.Slot)
		default:
			frame = buildMobSkillPacket(0x060D, result, d.MobGlobalID, ch.ID, sk, words, nil)
		}
		sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius, frame, "mob skill on character")
		if mirrored {
			s.reflectToMob(d.MobGlobalID, back)
		}
		s.log.Printf("GX Server World mob skill mob_global_id=%d mob_id=%d skill_id=%d result=%d damage=%d killed=%t",
			d.MobGlobalID, d.MobID, sk.SkillID, result, hp, killed)
		if !killed && !drain && result != attackResultMiss && (sk.AddDamageHP != 0 || sk.AddDamageSP != 0 || sk.AddDamageMP != 0) {
			add := extraSkillDamage(sk, 100+mobLevelPenalty(stats.Level, mc.Level))
			extra := [3]uint16{uint16(add[0]), uint16(add[1]), uint16(add[2])}
			addMirrored := mirrored || s.reflects(min(int(sk.Kind), 2), md)
			shownAdd := extra
			if addMirrored {
				shownAdd = [3]uint16{}
			} else {
				if add[1] != 0 || add[2] != 0 {
					s.spendDelta(add[2], add[1])
					if add[0] == 0 {
						s.persist()
					}
				}
				killed = s.takeMobDamage(d, add[0], stats)
			}
			w := packet.New(0x060E)
			w.U32(d.MobGlobalID)
			w.U32(ch.ID)
			w.U16(sk.SkillID)
			w.U8(sk.SkillLevel)
			w.U16(shownAdd[0])
			w.U16(shownAdd[1])
			w.U16(shownAdd[2])
			sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius, w.MustFinish(), "0x060E mob skill extra damage")
			if addMirrored {
				s.reflectToMob(d.MobGlobalID, extra)
			}
		}
		// CMob::AttackUser(skill): a debuff stays on the player.
		if !killed && result != attackResultMiss && sk.TypeEffect == typeEffectDebuff {
			s.applyMobDebuff(sk, d.MobGlobalID)
		}
		if !killed && result != attackResultMiss {
			s.gemCastsOnMob(d.MobGlobalID, armourGemSkills(s.items)) // BeAttackedItemOption(CMob)
		}
		refreshed := s.stats()
		stats = &refreshed
	}
	// A sacrifice skill (detail 23) pays the caster's whole HP after the blow
	// resolves. For a mob ("Self Destruction") that means dying with no credit,
	// drop or experience, exactly like the player's afterSkillStrikes.
	if sk.TypeDetail == detailSacrifice {
		s.eraseMob(d.MobGlobalID)
	}
	s.afterMobDamage(d, killed)
	return killed
}

// buildMobSkillPacket: result, mob id, target id, a byte the client reads
// before the skill id (0x060B/0x060C only), skill id, level, HP/SP/MP
// damage and, for 0x060C, the healed HP/SP/MP.
func buildMobSkillPacket(op packet.Opcode, result uint8, mobID, target uint32, sk *db.SkillDef, damage [3]uint16, healed *[3]uint16, attackSlot ...uint8) []byte {
	w := packet.New(op)
	w.U8(result)
	w.U32(mobID)
	w.U32(target)
	if op != 0x060D {
		slot := uint8(0)
		if len(attackSlot) > 0 {
			slot = min(attackSlot[0], 2)
		}
		w.U8(slot)
	}
	w.U16(sk.SkillID)
	w.U8(sk.SkillLevel)
	for _, v := range damage {
		w.U16(v)
	}
	if healed != nil {
		for _, v := range healed {
			w.U16(v)
		}
	}
	return w.MustFinish()
}

// takeMobDamage lowers the character's HP; true when it died.
func (s *Session) takeMobDamage(d pendingDamage, damage uint32, stats *CharacterStats) bool {
	s.assisted = [3]uint32{damage}
	if s.invulnerable() {
		return false
	}
	ch := s.char
	cur := s.stats()
	if cur.CurrentHP == 0 || damage == 0 {
		return false
	}
	s.cancelMove() // CMob::Attack -> StatusCastingStop -> CancleMove
	s.breakStealth()
	s.cancelTravelCast()
	damage = s.absorbHealthAssist(damage)
	if damage > 0 {
		s.addDamage(damageByMob, d.MobGlobalID, damage)
	}
	newEffective := satSub(cur.CurrentHP, damage)
	ch.CurrentHP = satSub(ch.CurrentHP, cur.CurrentHP-newEffective)
	killed := newEffective == 0
	s.persist()
	s.cancelVehicleCast()
	if !killed {
		s.wake()
	}
	return killed
}

func (s *Session) afterMobDamage(d pendingDamage, killed bool) {
	ch := s.char
	refreshed := s.stats()
	s.send(buildCurrentHitpoints(&refreshed), "0x0521 current hp after mob attack")
	if !killed {
		return
	}
	s.damageLedger = nil // CDamage::Reset; deaths only count by users (KillPCFrom)
	state, death := buildCharacterDeathState(ch.ID), buildCharacterDeath(ch.ID, deathTypeMob, d.MobGlobalID)
	for _, p := range PlayersNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius) {
		p.Send(state, "0x051C character death state")
		p.Send(death, "0x0504 character death")
	}
	ClearMobAggro(ch.MapID, d.MobGlobalID)
	if s.player != nil {
		s.player.SetAlive(false)
		duelLeave(s.player, 3, true) // PvPChkLeaveByDeath
		guildDuelLeave(s.player, nil, 3)
	}
	s.log.Printf("GX Server World character died character_id=%d killer_mob_global_id=%d mob_id=%d", ch.ID, d.MobGlobalID, d.MobID)
}

// persist saves the active character (usp_Save_Char_Info_E) and refreshes
// the shared snapshot.
func (s *Session) persist() {
	if s.operation != nil {
		return
	}
	if s.char == nil {
		return
	}
	if err := db.SaveCharacter(s.char); err != nil {
		s.persistenceFailed = true
		s.logErr("save character", err)
		if s.conn != nil {
			netx.Abort(s.conn)
		}
		return
	}
	s.syncPlayer()
	s.partyPools()
}

// --- experience ---

// applyExpGain returns the levels gained and experience actually credited.
func applyExpGain(ch *db.Character, gained uint32) (levels uint16, credited uint32) {
	defs := db.MustDefs()
	limit := defs.LevelCap
	if ch.Level >= limit || gained == 0 {
		return 0, 0
	}
	before := ch.Exp
	ch.Exp = min(satAdd(ch.Exp, gained), defs.Exp[limit-1])
	for ch.Level < limit && ch.Exp >= defs.Exp[ch.Level] {
		levelUp(ch)
		levels++
	}
	if levels > 0 {
		ch.MaxHP, ch.MaxMP, ch.MaxSP = basePools(ch)
		ch.CurrentHP, ch.CurrentMP, ch.CurrentSP = ch.MaxHP, ch.MaxMP, ch.MaxSP
	}
	return levels, satSub(ch.Exp, before)
}

type experienceCredit struct{ base, perfect uint32 }

// CUser::AddExpFromUser (0x458240) grants an equal extra amount to a normal
// seven-member party when the base award alone is below the next threshold.
// Process any level crossed by the bonus immediately using the GX curve.
func applyExperienceAward(ch *db.Character, amount uint32, perfect bool) experienceCredit {
	total := amount
	if perfect {
		defs := db.MustDefs()
		if ch.Level < defs.LevelCap && uint64(ch.Exp)+uint64(amount) < uint64(defs.Exp[ch.Level]) {
			total = satAdd(amount, amount)
		}
	}
	_, credited := applyExpGain(ch, total)
	base := min(amount, credited)
	return experienceCredit{base, credited - base}
}

// expBounds returns the cumulative thresholds used by the character's bar.
func expBounds(ch *db.Character) (prev, next uint32) {
	defs := db.MustDefs()
	if defs.LevelCap != 0 && ch.Level >= defs.LevelCap {
		capExp := defs.Exp[defs.LevelCap-1]
		return capExp, capExp
	}
	return defs.Exp[ch.Level-1], defs.Exp[ch.Level]
}

// rewardKill hands out a kill's loot and exp: drops on the ground, gold
// split across the party (or dropped for a solo killer) and each share of
// exp queued to its recipient's session. An offline DoT caster gets the
// solo exp straight in the database.
func rewardKill(killer *ConnectedPlayer, killerID uint32, mapID uint16, m *MapMob, log logPrinter) {
	if m.event != nil && m.event.principal {
		killer = eventCreditedPlayer(killer, mapID, m)
		killerID = 0
		if killer != nil {
			killerID = killer.CharacterID
		}
	}
	eventMobDestroyed(killer, mapID, m, log)
	var party []*ConnectedPlayer
	if killer != nil {
		party = partyMembersNear(killer, mapID, m.PosX, m.PosZ)
	}
	// The loot belongs to whoever dealt the most damage (its whole party counts
	// together), not to the killing blow or the aggro holder.
	lootBy, lootByID := killer, killerID
	if top := lootOwner(m); top != nil {
		lootBy, lootByID = top, top.CharacterID
	}
	var lootParty []*ConnectedPlayer
	if lootBy != nil {
		lootParty = partyMembersNear(lootBy, mapID, m.PosX, m.PosZ)
	}
	var rewards []killReward
	var perfect bool
	base := uint16(min(m.Exp, 65535))
	if killer != nil {
		rewards, perfect = killRewards(killer, mapID, m.Level, base)
	}
	for _, it := range rollMobDrops(mapID, m, lootByID, killerID, lootParty, log) {
		dispatchItemEnter(mapID, &it, "0x0401 mob drop")
	}
	questRecipients := party
	if len(questRecipients) == 0 && killer != nil {
		questRecipients = []*ConnectedPlayer{killer}
	}
	if def := questItemDrop(m, questRecipients); def != nil {
		x, y, z := mobDropPosition(m)
		mi := MapItem{ID: nextMapItemID(), OwnerID: killerID, DroppedAt: time.Now(), PosX: x, PosY: y, PosZ: z,
			Item: db.Item{Type: def.Type, TypeID: def.TypeID, Quality: def.Quality, Count: 1}}
		addMapItem(mapID, mi)
		dispatchItemEnter(mapID, &mi, "0x0401 quest item drop")
	}
	if killer == nil {
		if base > 0 {
			grantOfflineKillExp(killerID, m.MobID, m.Level, uint32(base), log)
		}
		return
	}
	for _, reward := range rewards {
		reward.member.QueueQuestKill(m.MobID)
		if amount := config.ScaleEXP(reward.exp, rewardRates.MobEXPPercent); amount > 0 {
			reward.member.QueueExp(amount, perfect)
		}
	}
}

// Active skill IDs are unique within each skill catalog. Each EXP ability
// contributes percentage points to the total applied to monster rewards.
func (s *Session) experienceBonus() uint32 {
	var bonus uint32
	for _, b := range s.buffs {
		for _, a := range b.Abilities {
			if a.Type == 35 && a.Value > 0 {
				bonus = satAdd(bonus, uint32(a.Value))
			}
		}
	}
	return bonus
}

func (s *Session) notifyExpGain(credited experienceCredit, levels uint16) {
	ch := s.char
	if credited.base > 0 || credited.perfect > 0 {
		s.send(buildExpGain(credited), "0x0207 exp gained")
	}
	if levels > 0 {
		// The level-up notice goes to the character (0x0508) and to the
		// players around it (0x051E), within 60 units.
		s.send(buildLevelUp(packet.LevelUpSelf, ch), "0x0508 level up")
		others := buildLevelUp(packet.LevelUpOther, ch)
		for _, o := range PlayersNearExcluding(ch.MapID, ch.ID, ch.PosX, ch.PosZ, 60) {
			o.Send(others, "0x051E level up (nearby)")
		}
		s.sendStatPackets()
		s.partyLevel()
		s.guildLevelChanged()
		s.enforceMapLevel()
	}
}

// buildExpGain (0x0207): base experience and the extra Perfect Party EXP.
func buildExpGain(credited experienceCredit) []byte {
	w := packet.New(packet.ExpGain)
	w.U32(credited.base)
	w.U32(credited.perfect)
	return w.MustFinish()
}

// buildLevelUp (0x0508 self / 0x051E others): character id, level, stat
// points, skill points, exp at the level's start and for the next level.
func buildLevelUp(op packet.Opcode, ch *db.Character) []byte {
	prev, next := expBounds(ch)
	w := packet.New(op)
	w.U32(ch.ID)
	w.U16(ch.Level)
	w.U16(ch.StatPoint)
	w.U16(ch.SkillPoint)
	w.U32(prev)
	w.U32(next)
	return w.MustFinish()
}

// grantOfflineKillExp credits a DoT kill to a caster who went offline.
func grantOfflineKillExp(characterID uint32, mobID, mobLevel uint16, mobExp uint32, log logPrinter) {
	ch, err := db.CharacterByID(characterID)
	if err != nil {
		return
	}
	scaled := uint32(float64(mobExperience(mobLevel, uint16(min(mobExp, 65535)), ch.Level)) * 1.2)
	levels, credited := applyExpGain(&ch, config.ScaleEXP(scaled, rewardRates.MobEXPPercent))
	if credited == 0 && levels == 0 {
		return
	}
	if err := db.SaveCharacter(&ch); err != nil {
		log.Printf("GX Server World offline exp save failed character_id=%d: %v", characterID, err)
		return
	}
	log.Printf("GX Server World offline exp character_id=%d mob_id=%d gained=%d level=%d exp=%d", characterID, mobID, credited, ch.Level, ch.Exp)
}

type logPrinter interface {
	Printf(format string, args ...any)
}
