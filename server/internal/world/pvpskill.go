package world

import (
	"github.com/Spelunkern/Shaiya-GX/internal/skillrange"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Hostile skills on users follow ps_game.exe:
//
//   - 0x0511 (a skill on a user) with TypeEffect 1 or 3: CUser::SkillAttack
//     0x00451d90 / SkillDebuff 0x00452bf0 pick TargetType 3 (one hostile
//     user, CUser::ChkTargetUser 0x0044e140: !IsTeam and attackable, else
//     result 11), 6 (around the caster, SkillAttackRange 0x00451f80 /
//     SkillDebuffRange 0x00452c50) or 7 (around the target user,
//     SkillAttackRangeSelect 0x00452560 / SkillDebuffRangeSelect
//     0x00453050); an area announces 0x0511 result 4 and hits the hostile
//     users within ApplyRange first, then the mobs.
//   - 0x0517 TargetType 7 (around a mob, 0x00450f70 / 0x00451570) hits the
//     hostile users around the mob too; 0x0517 TargetType 6 is mobs only.
//   - CUser::AttackUser(skill) 0x0044f4c0 on each user: a miss sends 0x0511
//     (0x0512 for a drain) or, for later and area hits, 0x0513, result 3;
//     a hit runs ChkMirror, AttackDamage (the weapon wears on a critical),
//     CUser::Damage, 0x0511 / 0x0512 (drain: the caster heals the damage)
//     / 0x0513 to the view, then the skill's AddDamage with 0x0514, a
//     mirrored hit back with 0x0515, and a debuff (TypeEffect 3) stays on
//     a target that survived. Area hits skip drains; detail 22 repeats
//     MultiAttack times; detail 23 then sacrifices the caster.
//
// Not modeled: the HP-ability shields
// (abilities 7/8) and state resistances in AttackSuccess, the clone of
// skill 100.

const (
	opUserSkillDrain packet.Opcode = 0x0512
	opUserSkillOther packet.Opcode = 0x0513
	opUserSkillExtra packet.Opcode = 0x0514
	opUserSkillPvP   packet.Opcode = packet.SkillUseNoTarget // 0x0511
)

// userSkillCenter resolves a 0x0511 hostile skill's target user; ok false
// is ChkTargetUser's refusal.
func (s *Session) userSkillCenter(def *db.SkillDef, target uint32) (*ConnectedPlayer, bool) {
	if def.TargetType == targetTypeEnemiesNearCaster {
		return nil, true
	}
	p := s.hostileUser(target)
	return p, p != nil
}

// hostileUsersWithin lists the alive hostile users within r of (x, z).
func (s *Session) hostileUsersWithin(x, z, r float32) []*ConnectedPlayer {
	var out []*ConnectedPlayer
	for _, p := range PlayersOnMap(s.char.MapID) {
		if p == s.player || !p.IsAlive() || p.immune.Load() || s.isTeam(p) || !s.pvpWindowAllows(p) {
			continue
		}
		px, pz := p.PositionXZ()
		if DistanceXZ(x, z, px, pz) <= r {
			out = append(out, p)
		}
	}
	return out
}

// userSkillAttack runs an attack or debuff skill sent as 0x0511.
func (s *Session) userSkillAttack(k db.CharacterSkill, def *db.SkillDef, target uint32, castPos *[2]float32, stats CharacterStats) {
	ch := s.char
	if !def.MobSkill && def.TypeEffect == typeEffectAttack && (def.TypeDetail < detailAttack || def.TypeDetail > detailDrain) {
		s.log.Printf("GX Server World rejected skill use: attack detail %d not modeled skill_id=%d", def.TypeDetail, k.SkillID)
		return
	}
	if stats.Disabled[min(int(def.Kind), 2)] > 0 {
		return
	}
	center, ok := s.userSkillCenter(def, target)
	if !ok {
		s.send(buildSkillUseResult(opUserSkillPvP, skillResultNotTarget, ch.ID, 0, k.SkillID, k.Level, 0, 0, 0), "0x0511 skill target refused")
		return
	}
	if center != nil {
		if !s.userSkillReach(center, def, castPos != nil) {
			s.sendUserSkillRangeFailure(k, center)
			return
		}
	}
	if center != nil {
		c := center.Snapshot()
		if !s.beginStinger(def, [3]float32{c.PosX, c.PosY, c.PosZ}) {
			return
		}
	}
	sp, mp := s.skillCosts(def, &stats)
	s.breakCombatForm()
	s.breakStealth()
	s.startCooldown(k, def, castPos != nil)
	s.spendDelta(mp, sp)
	s.persist()
	s.send(buildSkillUsedSpMp(sp, mp), "0x050C skill used sp/mp")
	s.lastCombatAt = time.Now()
	s.log.Printf("GX Server World pvp skill use skill_id=%d level=%d effect=%d detail=%d target_type=%d",
		k.SkillID, k.Level, def.TypeEffect, def.TypeDetail, def.TargetType)
	if def.TargetType == targetTypeEnemiesNearCaster || def.TargetType == targetTypeEnemiesNearTarget {
		x, y, z, announced := ch.PosX, ch.PosY, ch.PosZ, uint32(0)
		if center != nil {
			c := center.Snapshot()
			x, y, z, announced = c.PosX, c.PosY, c.PosZ, center.CharacterID
		}
		sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius,
			buildSkillUseResult(opUserSkillPvP, resultAnnounce, ch.ID, announced, k.SkillID, k.Level, 0, 0, 0), "0x0511 area skill")
		if def.TypeDetail == detailDrain {
			return
		}
		s.areaUserHits(def, k, x, y, z)
		for _, id := range s.skillMobsWithin(x, y, z, float32(def.ApplyRange)) {
			s.skillStrike(id, def, k, false)
		}
		s.afterSkillStrikes(def)
		return
	}
	if def.TypeDetail == detailMulti {
		sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius,
			buildSkillUseResult(opUserSkillPvP, resultAnnounce, ch.ID, center.CharacterID, k.SkillID, k.Level, 0, 0, 0), "0x0511 multi-hit skill")
		s.queueSkillHits(center, def, false, def.MultiAttack)
		return
	}
	s.queueSkillHits(center, def, true, 1)
	s.afterSkillStrikes(def)
}

// areaUserHits queues visible hostile users inside the 3D area radius.
func (s *Session) areaUserHits(def *db.SkillDef, k db.CharacterSkill, x, y, z float32) {
	n := uint8(1)
	if def.TypeDetail == detailMulti {
		n = def.MultiAttack
	}
	for _, p := range s.hostileUsersWithin(x, z, float32(def.ApplyRange)) {
		if s.safeZoneBlocksUser(p) {
			continue
		}
		c := p.Snapshot()
		if s.skillAreaContains(skillrange.Position{x, y, z}, skillrange.Position{c.PosX, c.PosY, c.PosZ}, float32(def.ApplyRange)) {
			s.queueSkillHits(p, def, false, n)
		}
	}
}

func (s *Session) queueSkillHits(p *ConnectedPlayer, def *db.SkillDef, first bool, n uint8) {
	ch := s.char
	stats := s.stats()
	caster := currentDamagePools(&stats)
	for i := uint8(0); i < n; i++ {
		p.QueueUserHit(pendingUserHit{AttackerID: ch.ID, Stats: stats, Kind: int(def.Kind),
			Skill: def, First: first, Caster: caster})
	}
}

// applyUserSkillHit is the target side of CUser::AttackUser(skill); true
// when this character died.
func (s *Session) applyUserSkillHit(h pendingUserHit, stats CharacterStats) bool {
	ch := s.char
	sk := h.Skill
	attacker := FindPlayerByID(h.AttackerID)
	drain := h.First && sk.TypeDetail == detailDrain
	op := opUserSkillOther
	switch {
	case drain:
		op = opUserSkillDrain
	case h.First:
		op = opUserSkillPvP
	}
	send := func(result uint8, dealt, healed [3]uint16, label string) {
		var frame []byte
		if op == opUserSkillDrain {
			frame = buildDrainPacket(op, result, h.AttackerID, ch.ID, sk.SkillID, sk.SkillLevel, dealt, healed)
		} else {
			frame = buildSkillUseResult(op, result, h.AttackerID, ch.ID, sk.SkillID, sk.SkillLevel, dealt[0], dealt[1], dealt[2])
		}
		sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius, frame, label)
	}
	lapisDebuff := h.ItemCast && sk.TypeEffect == typeEffectDebuff
	s.cancelMove()
	kind := min(int(sk.Kind), 2)
	if s.blocksAttack(kind) || s.immuneToLapisSkill(sk) || s.blocksStatusSkill(sk) || !userSkillHits(&h.Stats, &stats, sk) {
		if !lapisDebuff {
			send(attackResultMiss, [3]uint16{}, [3]uint16{}, "pvp skill (miss)")
		}
		return false
	}
	duelHitBy(s.player, h.AttackerID) // PvPChkLeaveByAttack
	mirrored := s.reflects(kind, nil)
	damage, crit := userSkillDamage(&h.Stats, &stats, sk, h.Caster)
	words := func(d [3]int32) [3]uint16 {
		var out [3]uint16
		for i, v := range d {
			out[i] = uint16(min(max(v, 0), 65535))
		}
		return out
	}
	dealt := words(damage)
	hp := uint32(dealt[0])
	if stats.Absorption > 0 && hp > 0 {
		hp = max(satSub(hp, stats.Absorption), 1)
		dealt[0] = uint16(hp) // Damage display, drain and reflection use the absorbed hit.
	}
	code := uint8(attackSuccessNormal)
	if crit {
		code = attackResultCritical
	}
	killed := false
	if !mirrored {
		killed = s.takeUserSkillDamage(h.AttackerID, [3]uint32{hp, uint32(dealt[1]), uint32(dealt[2])})
	}
	shown := dealt
	if mirrored {
		shown = [3]uint16{}
	} else if hp > 0 { // Health Assistant moved part of the HP damage to SP / MP
		shown[0] = uint16(min(s.assisted[0], 65535))
		shown[1] = uint16(min(uint32(shown[1])+s.assisted[1], 65535))
		shown[2] = uint16(min(uint32(shown[2])+s.assisted[2], 65535))
	}
	drainQueued := false
	if drain && !mirrored && attacker != nil && dealt != ([3]uint16{}) {
		resolved := &pendingDrain{
			source: h.AttackerID, target: ch.ID, mapID: ch.MapID, x: ch.PosX, z: ch.PosZ,
			skill: sk, result: code, dealt: shown, recovery: dealt,
		}
		drainQueued = attacker.QueueUserHit(pendingUserHit{Drain: resolved})
	}
	if !lapisDebuff && !drainQueued {
		send(code, shown, [3]uint16{}, "pvp skill")
	}
	if !drain && !killed && (sk.AddDamageHP != 0 || sk.AddDamageSP != 0 || sk.AddDamageMP != 0) {
		add := extraSkillDamage(sk, 100)
		shownAdd, addMirrored := add, mirrored || s.reflects(kind, nil)
		if addMirrored {
			shownAdd = [3]uint32{}
		} else {
			killed = s.takeUserSkillDamage(h.AttackerID, add)
		}
		sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius,
			buildUserSkillExtra(h.AttackerID, ch.ID, sk.SkillID, sk.SkillLevel, shownAdd), "0x0514 pvp skill extra damage")
		if addMirrored {
			s.mirrorDamage(h.AttackerID, 0, add)
		}
	}
	if mirrored && attacker != nil {
		attacker.QueueUserHit(pendingUserHit{AttackerID: ch.ID, Reflected: true,
			Damage: [3]uint32{uint32(dealt[0]), uint32(dealt[1]), uint32(dealt[2])}})
	}
	if !killed && sk.TypeEffect == typeEffectDebuff {
		if _, ok := s.addEffect(sk, 0); ok {
			s.buffs[len(s.buffs)-1].CasterUser = h.AttackerID
			if lapisDebuff {
				send(code, shown, [3]uint16{}, "pvp skill")
			}
		}
	}
	s.log.Printf("GX Server World pvp skill hit attacker_id=%d target_id=%d skill_id=%d damage=%v crit=%t mirrored=%t killed=%t",
		h.AttackerID, ch.ID, sk.SkillID, dealt, crit, mirrored, killed)
	s.afterUserDamage(h.AttackerID, killed)
	if !killed {
		s.itemOptionsAfterUserHit(h)
	}
	return killed
}

// takeUserSkillDamage is CUser::Damage with SP and MP damage too.
func (s *Session) takeUserSkillDamage(attackerID uint32, d [3]uint32) bool {
	if s.invulnerable() {
		return false
	}
	if d[0] > 0 || d[1] > 0 || d[2] > 0 {
		s.breakCombatForm()
		s.cancelTravelCast()
		s.breakStealth()
	}
	if d[1] > 0 || d[2] > 0 {
		s.spendDelta(d[2], d[1])
	}
	if d[0] == 0 {
		s.persist()
		return false
	}
	return s.takeUserDamage(attackerID, d[0])
}

// pendingDrain carries a landed PvP hit to the caster, whose session owns the
// pools and can put the actual recovery into the native 0x0512 result.
type pendingDrain struct {
	source, target  uint32
	mapID           uint16
	x, z            float32
	skill           *db.SkillDef
	result          uint8
	dealt, recovery [3]uint16
}

func (d *pendingDrain) broadcast(healed [3]uint16) {
	frame := buildDrainPacket(opUserSkillDrain, d.result, d.source, d.target, d.skill.SkillID, d.skill.SkillLevel, d.dealt, healed)
	sendNear(d.mapID, d.x, d.z, CombatRadius, frame, "0x0512 pvp drain")
}

// applyDrainHeal is CUser::Heal on the drain's caster.
func (s *Session) applyDrainHeal(d *pendingDrain) {
	var healed [3]uint16
	if !s.dead && s.char != nil && s.char.CurrentHP > 0 {
		hp, mp, sp := s.healDelta(uint32(d.recovery[0]), uint32(d.recovery[2]), uint32(d.recovery[1]))
		healed = [3]uint16{uint16(hp), uint16(sp), uint16(mp)}
		if healed != ([3]uint16{}) {
			s.persist()
		}
	}
	d.broadcast(healed)
	if healed != ([3]uint16{}) {
		st := s.stats()
		s.send(buildCurrentHitpoints(&st), "0x0521 current hp after drain")
	}
}

// buildUserSkillExtra (0x0514): caster, target, skill, level, the
// AddDamage HP/SP/MP.
func buildUserSkillExtra(caster, target uint32, skillID uint16, level uint8, add [3]uint32) []byte {
	w := packet.New(opUserSkillExtra)
	w.U32(caster)
	w.U32(target)
	w.U16(skillID)
	w.U8(level)
	for _, v := range add {
		w.U16(uint16(v))
	}
	return w.MustFinish()
}
