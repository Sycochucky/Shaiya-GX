package world

import (
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/attackspeed"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// Normal attacks between users follow ps_game.exe:
//
//   - PacketPC 0x0502 and CUser::UpdateAttack 0x0048ca50 pick the target:
//     a user on the zone, alive, not the attacker, and hostile
//     (CUser::IsTeam 0x0044bdf0 false). EnableAttackRange 0x0044bf90 is the
//     move allowance + the attacker's range + 2.3 (no target size).
//   - CUser::AttackUser 0x00456870: the category is ranged for weapon
//     types 11, 13 and 14 and does nothing while disabled; a miss sends
//     0x0502 result 3; a hit runs ChkMirror on the target, AttackDamage,
//     then CUser::Damage on the target, or, when mirrored,
//     0x0502 with no damage followed by Damage on the attacker and 0x0515.
//   - GX combat formulas are shared by PvP and PvE. PvP is independent
//     of player levels. See combatformula.go.
//   - Death(target, 1, attacker) announces 0x0504 type 1.

const (
	deathTypeUser   = 0x01
	userAttackRange = 2.3
)

// pendingUserHit is a user's normal attack (or a mirrored one coming back)
// to resolve in the target's session.
type pendingUserHit struct {
	AttackerID uint32
	Stats      CharacterStats // the attacker's, at the swing
	Kind       int
	Reflected  bool // a mirrored hit: Damage is final
	Execute    bool // GM /kill: the target dies at once; the kill is credited to AttackerID
	Fixed      bool // fixed skill damage (Fire Thorn): Damage is final, Skill names it
	Trap       bool // a trap sprang on this user: the damage is worked out here (TrabDamage), Skill names it
	Damage     [3]uint32
	Skill      *db.SkillDef  // a hostile skill (pvpskill.go), nil for a normal attack
	First      bool          // the skill's selected target (0x0511/0x0512), else 0x0513
	Caster     damagePools   // the caster's HP/SP/MP (DamageType 3)
	Drain      *pendingDrain // a landed drain to finish in the caster's session
	ItemCast   bool          // cast by a gem (m_bItemAttack): casts no gems itself
	GemAttack  bool          // run AttackItemOption on AttackerID (the target hit)
}

func (p *ConnectedPlayer) QueueUserHit(h pendingUserHit) bool {
	p.mailMu.Lock()
	if p.combatMailClosed {
		p.mailMu.Unlock()
		return false
	}
	if len(p.userHits) >= maxPendingCombat {
		p.mailMu.Unlock()
		p.mailboxOverflow()
		return false
	}
	p.userHits = append(p.userHits, h)
	p.mailMu.Unlock()
	return true
}

func (p *ConnectedPlayer) TakeUserHits() []pendingUserHit {
	p.mailMu.Lock()
	defer p.mailMu.Unlock()
	h := p.userHits
	p.userHits = nil
	return h
}

func (p *ConnectedPlayer) closeUserHits() {
	p.mailMu.Lock()
	p.combatMailClosed = true
	hits := p.userHits
	p.userHits = nil
	p.mailMu.Unlock()
	for _, h := range hits {
		if h.Drain != nil {
			h.Drain.broadcast([3]uint16{})
		}
	}
}

// isTeam is CUser::IsTeam without duels (none exist yet): the user
// itself, anyone on a WarType P map, or the same country.
// isTeam is CUser::IsTeam: the user itself, else not the duel opponent,
// then a WarType P map or the same country.
func (s *Session) isTeam(p *ConnectedPlayer) bool {
	if p == s.player {
		return true
	}
	if s.player != nil && (duelFoes(s.player, p) || guildDuelFoes(s.player, p)) {
		return false
	}
	if z, ok := mapConfig.Zones[realMap(s.char.MapID)]; ok && z.WarType == "P" {
		return true
	}
	return p.Faction() == s.faction
}

// hostileUser finds an attackable user for a normal attack.
func (s *Session) hostileUser(id uint32) *ConnectedPlayer {
	if s.char == nil || id == s.char.ID {
		return nil
	}
	p := FindPlayerByID(id)
	if p == nil || p.MapID() != s.char.MapID || !p.IsAlive() || !s.player.canSee(p) || p.immune.Load() || s.isTeam(p) || !s.pvpWindowAllows(p) {
		return nil
	}
	return p
}

// userAttackSwing is one UpdateAttack step on a user target; false ends
// the auto-attack.
func (s *Session) userAttackSwing(target uint32) bool {
	s.cancelVehicleCast()
	ch := s.char
	if ch == nil || s.ridingVehicle() || s.cast != nil || s.dash != nil || s.cannotAct() {
		return false
	}
	wc, ok := equippedWeaponCombat(s.items)
	if !ok {
		return false
	}
	stats := StatsFor(ch, s.equipment())
	wc.Delay, _ = weaponAttackDelay(attackspeed.Tier(wc.AttackTime, stats.AttackSpeed), wc.AttackTime)
	if !s.lastActionAt.IsZero() && time.Since(s.lastActionAt) < max(wc.Delay, s.ownClip()) {
		return true
	}
	p := s.hostileUser(target)
	if p == nil {
		return false
	}
	if s.safeZoneBlocksUser(p) {
		s.sendSkillMessage(0, proto.MsgCannotAttack)
		return false
	}
	x, z := p.PositionXZ()
	distance := DistanceXZ(ch.PosX, ch.PosZ, x, z)
	if distance > wc.Range+float32(stats.Range)+userAttackRange+autoAttackRangeGrace {
		if distance > autoAttackCancelRange {
			return false
		}
		return true
	}
	kind := attackMelee
	if stats.RangedWeapon {
		kind = attackRanged
	}
	if stats.Disabled[kind] > 0 {
		return true
	}
	if !s.attackMovementAllowed() {
		return false
	}
	s.commitAttackStop(time.Now())
	s.breakCombatForm()
	s.breakStealth()
	s.lastActionAt, s.lastCombatAt = time.Now(), time.Now()
	p.QueueUserHit(pendingUserHit{AttackerID: ch.ID, Stats: stats, Kind: kind})
	return true
}

// tickUserHits resolves the user attacks queued on this session.
func (s *Session) tickUserHits() {
	if s.player == nil {
		return
	}
	for _, h := range s.player.TakeUserHits() {
		if h.Drain != nil {
			s.applyDrainHeal(h.Drain)
			continue
		}
		if s.dead {
			continue
		}
		if s.applyUserHit(h) {
			s.dead, s.autoAttack, s.cast = true, false, nil
			s.lastActionAt = time.Time{}
			s.clearBuffsOnDeath()
		}
	}
}

// applyUserHit is the target side of CUser::AttackUser; true when this
// character died.
func (s *Session) applyUserHit(h pendingUserHit) bool {
	if h.Execute {
		return s.applyExecute(h)
	}
	if s.invulnerable() {
		return false
	}
	ch := s.char
	if ch == nil || s.safeZoneProtection().protects(s.faction) {
		return false
	}
	stats := s.stats()
	if stats.CurrentHP == 0 {
		return false
	}
	s.lastCombatAt = time.Now()
	if h.GemAttack {
		s.applyGemAttack(h.AttackerID)
		return false
	}
	if h.Trap && h.Skill != nil {
		return s.applyTrapHit(h)
	}
	if h.Fixed && h.Skill != nil {
		return s.applyFixedHit(h)
	}
	if h.Skill != nil {
		return s.applyUserSkillHit(h, stats)
	}
	attacker := FindPlayerByID(h.AttackerID)
	if h.Reflected {
		killed := s.takeUserSkillDamage(h.AttackerID, h.Damage)
		w := packet.New(0x0515)
		w.U32(h.AttackerID) // the mirror's owner
		w.U32(ch.ID)
		for _, v := range h.Damage {
			w.U16(uint16(min(v, 65535)))
		}
		sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius, w.MustFinish(), "0x0515 mirrored hit")
		s.afterUserDamage(h.AttackerID, killed)
		return killed
	}
	s.cancelMove() // StatusCastingStop on the target
	if s.blocksAttack(h.Kind) || !userHits(&h.Stats, &stats, h.Kind) {
		sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius,
			buildCharacterUserAttack(h.AttackerID, ch.ID, attackResultMiss, 0), "0x0502 user attack (miss)")
		return false
	}
	duelHitBy(s.player, h.AttackerID) // PvPChkLeaveByAttack
	mirrored := s.reflects(h.Kind, nil)
	value, crit := userDamage(&h.Stats, &stats, h.Kind)
	damage := uint32(max(value, 0))
	if stats.Absorption > 0 {
		damage = max(satSub(damage, stats.Absorption), 1)
	}
	code := uint8(attackSuccessNormal)
	if crit {
		code = attackResultCritical
	}
	killed, shown := false, damage
	if mirrored {
		shown = 0
	} else {
		killed = s.takeUserDamage(h.AttackerID, damage)
		shown = s.assisted[0]
	}
	sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius,
		buildCharacterUserAttack(h.AttackerID, ch.ID, code, uint16(min(shown, 65535))), "0x0502 user attack")
	if mirrored && attacker != nil {
		attacker.QueueUserHit(pendingUserHit{AttackerID: ch.ID, Reflected: true, Damage: [3]uint32{damage, 0, 0}})
	}
	s.log.Printf("GX Server World user attack attacker_id=%d target_id=%d kind=%d damage=%d crit=%t mirrored=%t killed=%t",
		h.AttackerID, ch.ID, h.Kind, damage, crit, mirrored, killed)
	s.afterUserDamage(h.AttackerID, killed)
	if !killed {
		s.itemOptionsAfterUserHit(h)
	}
	return killed
}

// takeUserDamage is CUser::Damage's HP part; true when it killed.
func (s *Session) takeUserDamage(attackerID, damage uint32) bool {
	s.assisted = [3]uint32{damage}
	if s.invulnerable() {
		return false
	}
	ch := s.char
	cur := s.stats()
	if cur.CurrentHP == 0 || damage == 0 {
		return false
	}
	s.breakStealth()
	s.cancelTravelCast()
	damage = s.absorbHealthAssist(damage)
	if damage > 0 {
		s.addDamage(damageByUser, attackerID, damage) // CDamage::Add in CUser::Damage
	}
	after := satSub(cur.CurrentHP, damage)
	ch.CurrentHP = satSub(ch.CurrentHP, cur.CurrentHP-after)
	killed := after == 0
	s.persist()
	s.cancelVehicleCast()
	if !killed {
		s.wake()
	}
	return killed
}

func (s *Session) afterUserDamage(killerID uint32, killed bool) {
	ch := s.char
	refreshed := s.stats()
	s.send(buildCurrentHitpoints(&refreshed), "0x0521 current hp after user attack")
	if !killed {
		return
	}
	s.deathByUser()
	defer duelLeave(s.player, 3, true)     // PvPChkLeaveByDeath
	defer guildDuelLeave(s.player, nil, 3) // its guild part
	state, death := buildCharacterDeathState(ch.ID), buildCharacterDeath(ch.ID, deathTypeUser, killerID)
	for _, p := range PlayersNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius) {
		p.Send(state, "0x051C character death state")
		p.Send(death, "0x0504 character death (by user)")
	}
	if s.player != nil {
		s.player.SetAlive(false)
	}
	s.log.Printf("GX Server World character died character_id=%d killer_character_id=%d", ch.ID, killerID)
}

func buildCharacterUserAttack(attackerID, targetID uint32, result uint8, damage uint16) []byte {
	w := packet.New(packet.CharacterUserAutoAttack)
	w.U8(result)
	w.U32(attackerID)
	w.U32(targetID)
	w.U16(damage)
	w.U16(0)
	w.U16(0)
	return w.MustFinish()
}
