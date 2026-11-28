package world

import (
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/skillchance"
	"github.com/Spelunkern/Shaiya-GX/internal/skilltiming"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

const (
	skillResultNormal = 0

	targetTypeSelf              = 2
	targetTypeEnemiesNearCaster = 6
	targetTypeEnemiesNearTarget = 7
)

// ActiveBuff is a self-buff on the session's character: stat bonuses
// and/or a heal-over-time.
type ActiveBuff struct {
	ID                     uint32
	MapCurse               bool
	SkillID                uint16
	SkillLevel             uint8
	EndsAt                 time.Time
	Effect, Detail         uint8
	Attr                   uint8 // AttrType, for the element buffs 94/95
	DefenceType            uint8
	BlockLeft              uint32 // detail 85 blocks left / 86 chance
	CasterMob              uint32 // the mob that applied a debuff
	Form                   formLook
	CasterUser             uint32 // the user that applied a debuff (its damage counts as theirs)
	prevTick               time.Time
	Abilities              [3]buffAbility
	HealHP, HealSP, HealMP uint32
	nextTickAt             time.Time // zero: no heal component
}

type pendingCast struct {
	targetMobID uint32
	number      uint8
	op          packet.Opcode
	readyAt     time.Time
	targetPos   *[2]float32 // Non-nil marks an accepted prepared cast; distance is checked only at launch.
	corpse      corpseTarget
	evolution   uint32 // granting form generation for a temporary monster skill
}

type buffAbility struct {
	Type  uint8
	Value int32
}

// addBuffs accumulates numeric abilities. Elemental application/removal is
// event-driven in elementstate.go; replaying the remaining list is not native.
func addBuffs(e EquipmentStats, buffs []ActiveBuff) EquipmentStats {
	for _, b := range buffs {
		debuff := b.Effect == typeEffectDebuff
		for i, a := range b.Abilities {
			v := a.Value
			if (b.Detail == 91 && i == 0) || debuff && (a.Type < 32 || a.Type > 34) {
				v = -v
			}
			applyAbility(&e, b.Effect, a.Type, skillchance.Modifier(a.Type, v))
		}
	}
	return e
}

func newBuff(id uint32, def *db.SkillDef, now, ends time.Time, log logPrinter) ActiveBuff {
	b := ActiveBuff{ID: id, SkillID: def.SkillID, SkillLevel: def.SkillLevel, EndsAt: ends, Effect: def.TypeEffect,
		Detail: uint8(min(def.TypeDetail, 255)), Attr: def.AttrType,
		HealHP: uint32(max(def.TimeHealHP, 0)), HealSP: uint32(max(def.TimeHealSP, 0)), HealMP: uint32(max(def.TimeHealMP, 0))}
	switch formOf(def) {
	case proto.FormDegeneration:
		b.Form = formLook{kind: proto.FormDegeneration, mob: def.ChangeLevel}
	case proto.FormEvolution:
		b.Form = formLook{kind: proto.FormEvolution, mob: def.ChangeLevel, skill: def.SkillID, rank: def.SkillLevel}
	}
	if b.Effect == typeEffectBuff && b.Detail == detailPeriodicHeal && b.HealHP|b.HealSP|b.HealMP != 0 {
		b.nextTickAt = now.Add(periodicHealTick)
	}
	b.DefenceType, b.BlockLeft, b.prevTick = def.DefenceType, uint32(def.DefenceValue), now
	if def.TypeEffect == typeEffectBuff && def.TypeDetail == detailFireThorn {
		b.nextTickAt = now.Add(fireThornTick)
	}
	if def.TypeEffect == typeEffectBuff && def.TypeDetail == detailDetection {
		b.nextTickAt = now // the first sweep is immediate
	}
	if def.TypeDetail == 82 || def.TypeDetail == 91 || def.TypeDetail == detailAbilityDown || def.TypeDetail == detailAbilityDown2 || isPotential(def) {
		for i, t := range def.AbilityType {
			b.Abilities[i] = buffAbility{t, int32(def.AbilityValue[i])}
			if !applyAbility(&EquipmentStats{}, def.TypeEffect, t, 0) {
				log.Printf("GX Server World buff ability_type=%d not modeled (value=%d) skill_id=%d", t, def.AbilityValue[i], def.SkillID)
			}
		}
	}
	return b
}

func buildCharacterActiveBuffs(buffs []ActiveBuff) []byte {
	w := packet.New(packet.CharacterActiveBuffs)
	buffs = buffs[:min(len(buffs), 255)]
	w.U8(uint8(len(buffs)))
	for _, b := range buffs {
		w.U32(b.ID)
		w.U16(b.SkillID)
		w.U8(b.SkillLevel)
		if b.MapCurse {
			w.U32(0)
		} else {
			w.U32(uint32(max(time.Until(b.EndsAt), 0) / time.Second))
		}
	}
	return w.MustFinish()
}

func buildSkillUseResult(op packet.Opcode, result uint8, sender, target uint32, skillID uint16, level uint8, hp, sp, mp uint16) []byte {
	w := packet.New(op)
	w.U8(result)
	w.U32(sender)
	w.U32(target)
	w.U16(skillID)
	w.U8(level)
	w.U16(hp)
	w.U16(sp)
	w.U16(mp)
	return w.MustFinish()
}

func buildSkillUsedSpMp(sp, mp uint32) []byte {
	w := packet.New(packet.SkillUsedSpMp)
	w.U32(sp)
	w.U32(mp)
	return w.MustFinish()
}

func (s *Session) knownSkill(number uint8) (db.CharacterSkill, *db.SkillDef, bool) {
	for _, k := range s.skills {
		if k.Number == number {
			def := db.MustDefs().Skill(k.SkillID, k.Level)
			return k, def, def != nil
		}
	}
	return db.CharacterSkill{}, nil, false
}

func (s *Session) skillReady(id uint16) bool {
	at, ok := s.skillCooldowns[id]
	return !ok || !at.After(time.Now())
}

// previousSkillCooling: PrevSkillID names the skill that must be cooling down
// for this one to be usable. SysMsg 315 is the client's "A previous skill is
// required".
func (s *Session) previousSkillCooling(def *db.SkillDef) bool {
	return def.PrevSkillID == 0 || !s.skillReady(def.PrevSkillID)
}

const msgPreviousSkillRequired = 315

// msgUnavailableEquipment is the native text for skill-use result 7 (client
// 0x5764d0 -> SysMsg 0x13a).
const msgUnavailableEquipment = 314

// handleUseSkill (0x0517 / 0x0511): bar number, target id. The ack echoes
// the request opcode.
func (s *Session) handleUseSkill(op packet.Opcode, p []byte) {
	if len(p) < 7 || s.char == nil {
		return
	}
	s.cancelMove()
	ch := s.char
	number := p[2]
	if op == packet.SkillUseNoTarget && number == 0xff {
		id, _ := packet.U32(p[3:])
		s.cancelBuff(id)
		return
	}
	target, _ := packet.U32(p[3:])
	s.pendingSkill = nil
	s.cancelVehicleCast()
	if s.ridingVehicle() || s.cast != nil || s.cannotAct() || s.evolutionGrant != 0 {
		s.log.Printf("GX Server World rejected skill use (mounted, casting, stunned or asleep) character_id=%d", ch.ID)
		return
	}
	k, def, ok := s.knownSkill(number)
	if !ok {
		s.log.Printf("GX Server World rejected skill use: no skill in slot number=%d", number)
		return
	}
	s.beginSkill(op, k, def, target, 0, p)
}

func (s *Session) beginSkill(op packet.Opcode, k db.CharacterSkill, def *db.SkillDef, target, evolution uint32, p []byte) {
	if !s.attackMovementAllowed() {
		return
	}
	ch := s.char
	number := k.Number
	if def.TargetType == proto.SkillTargetRaid && !proto.SupportsRaidSkillTarget(def.TypeEffect) {
		s.sendSkillMessage(k.SkillID, 317)
		return
	}
	if !def.MobSkill && !s.skillWeaponAllowed(def) {
		s.sendSkillMessage(k.SkillID, msgUnavailableEquipment)
		return
	}
	s.autoAttack = false
	if !s.skillActionReady(def, time.Now()) {
		if p == nil || !s.queueSpinCombo(def, op, p) {
			if evolution == 0 && s.holdSkillUse(def, op, p, time.Now()) {
				return
			}
			s.sendSkillMessage(k.SkillID, 215)
		}
		return
	}
	if kind := formOf(def); kind == proto.FormMonsterCopy || kind == proto.FormDisguise {
		if _, ok := s.copyFormTarget(def, target, false); !ok || !s.canAddEffect(def, ch.ID) {
			s.rejectCopyForm(op, k)
			return
		}
	} else {
		target = s.friendlySkillRecipient(def, target)
	}
	castTime := time.Duration(def.ReadyTime) * 250 * time.Millisecond
	if castTime == 0 {
		s.resolveSkillDefinition(target, k, def, op, nil)
		return
	}
	if s.safeZoneBlocksSkill(def, op, target) {
		s.rejectSafeZoneSkill(k, false)
		return
	}
	if !s.definitionReady(def) {
		s.sendSkillMessage(k.SkillID, 215)
		return
	}
	if !def.MobSkill && !s.previousSkillCooling(def) {
		s.sendSkillMessage(k.SkillID, msgPreviousSkillRequired)
		return
	}
	stats := s.stats()
	if s.cannotAct() || stats.Disabled[min(int(def.Kind), 2)] > 0 {
		s.log.Printf("GX Server World rejected skill use: category disabled skill_id=%d", k.SkillID)
		return
	}
	if sp, mp := s.skillCosts(def, &stats); stats.CurrentMP < mp || stats.CurrentSP < sp {
		s.log.Printf("GX Server World rejected skill use: not enough MP/SP skill_id=%d", k.SkillID)
		return
	}
	castTarget := target
	if def.TypeEffect == typeEffectHeal || def.TypeEffect == typeEffectBuff {
		castTarget = ch.ID
		if def.TargetType == targetTypeOther || def.TargetType == targetTypeOtherAlt {
			targets, _, ok := s.friendlyTargets(def, target)
			if !ok {
				s.sendFriendlySkillFailure(k, def, target)
				return
			}
			castTarget = targets[0].CharacterID
		}
	}
	var pos *[2]float32
	var corpse corpseTarget
	if def.TypeEffect == typeEffectOnce {
		p, ok := s.userSkillTarget(def, target)
		if def.TypeDetail == detailKillSoul {
			corpse = s.soulBurnCorpse(p)
			if corpse.player == nil {
				s.rejectSoulBurn(k)
				return
			}
		}
		if !ok {
			return
		}
		if !s.inUserSkillRange(p, def) {
			s.sendUserSkillRangeFailure(k, p)
			return
		}
		castTarget = p.CharacterID
	} else if op == opUserSkillPvP && (def.TypeEffect == typeEffectAttack || def.TypeEffect == typeEffectDebuff) {
		// A hostile skill on a user: ChkTargetUser before the cast.
		if def.TargetType != targetTypeEnemiesNearCaster {
			p := s.hostileUser(target)
			if p == nil {
				s.send(buildSkillUseResult(packet.SkillUseNoTarget, skillResultNotTarget, ch.ID, 0, k.SkillID, k.Level, 0, 0, 0),
					"0x0511 skill target refused")
				return
			}
			if !s.inUserSkillRange(p, def) {
				s.sendUserSkillRangeFailure(k, p)
				return
			}
			x, z := p.PositionXZ()
			castTarget, pos = p.CharacterID, &[2]float32{x, z}
		}
	} else if (def.TypeEffect == typeEffectAttack || def.TypeEffect == typeEffectDebuff) && castTarget != ch.ID && def.TargetType != targetTypeEnemiesNearCaster {
		if def.TypeEffect == typeEffectDebuff && def.TypeDetail == detailDegeneration {
			s.sendSkillMessage(k.SkillID, 317) // it cannot be cast upon a monster
			return
		}
		mob, ok := MobByID(ch.MapID, target)
		if !ok {
			return
		}
		if mobFriendly(db.MustDefs().Mobs[mob.MobID], s.faction) {
			s.sendSkillMessage(k.SkillID, 317)
			return
		}
		if !s.inSkillRange(&mob, def, &stats) {
			s.sendMobSkillRangeFailure(k, &mob)
			return
		}
		pos = &[2]float32{mob.PosX, mob.PosZ}
	}
	if def.TypeEffect == typeEffectAttack || def.TypeEffect == typeEffectDebuff {
		s.breakCombatForm()
		s.breakStealth()
	}
	s.send(packet.New(packet.AutoAttackStop).MustFinish(), "0x0212 auto attack stop (casting)")
	w := packet.New(packet.SkillCastStarted)
	w.U32(ch.ID)
	w.U32(castTarget)
	w.U16(k.SkillID)
	w.U8(k.Level)
	s.broadcastNear(w.MustFinish(), "0x0510 skill cast started", true)
	if pos == nil {
		pos = &[2]float32{}
	}
	now := time.Now()
	s.lastActionAt = now
	s.cast = &pendingCast{targetMobID: target, number: number, op: op, readyAt: now.Add(castTime), targetPos: pos, corpse: corpse, evolution: evolution}
	s.autoAttack = false
	s.syncPlayerSpeed(false)
	s.log.Printf("GX Server World skill cast started skill_id=%d level=%d cast_time=%s", k.SkillID, k.Level, castTime)
}

func (s *Session) processPendingCast() {
	c := s.cast
	if c == nil || c.readyAt.After(time.Now()) {
		return
	}
	s.cast = nil
	s.syncPlayerSpeed(false)
	if c.evolution != 0 {
		if def := s.evolutionSkill(c.evolution, c.number); def != nil {
			s.resolveSkillDefinition(c.targetMobID, db.CharacterSkill{SkillID: def.SkillID, Level: def.SkillLevel, Number: c.number}, def, c.op, c)
		}
		return
	}
	s.resolveSkill(c.targetMobID, c.number, c.op, c)
}

func (s *Session) startCooldown(k db.CharacterSkill, def *db.SkillDef, prepared ...bool) {
	now := time.Now()
	if len(prepared) == 0 || !prepared[0] {
		s.lastActionAt = now
		s.clipFloorAt, s.clipFloor = now, 0
		if def.ReadyTime == 0 && !def.MobSkill && def.TypeEffect != 0 {
			s.clipFloor = skilltiming.ClipFloor(k.SkillID)
		}
	}
	s.pendingCombo = nil
	s.lastSkillID, s.lastSkillAt = k.SkillID, s.lastActionAt
	s.commitAttackStop(now)
	if def.MobSkill {
		if s.evolutionReadyAt == nil {
			s.evolutionReadyAt = map[uint16]time.Time{}
		}
		s.evolutionReadyAt[k.SkillID] = now.Add(time.Duration(def.ResetTimeSeconds) * time.Second)
		s.lastSkillID = 0
		s.sendEvolutionBar()
		return
	}
	reuse := skilltiming.ReuseSeconds(def.ResetTimeSeconds)
	if reuse > 0 && !isPotential(def) {
		if cut := s.cooldownCut(); cut > 0 {
			reuse = reuse - min(reuse, cut)
			s.sendSkillCooldown(k.SkillID, reuse, reuse) // a shortened timer starts with a full bar
		}
	}
	if reuse == 0 {
		delete(s.skillCooldowns, k.SkillID)
	} else {
		s.skillCooldowns[k.SkillID] = now.Add(time.Duration(reuse) * time.Second)
	}
	k.Delay = int16(min(reuse, 32767))
	if err := db.SaveCharacterSkill(s.char.ID, k); err != nil {
		s.log.Printf("GX Server World failed to save skill cooldown: %v", err)
	}
}

func (s *Session) resolveSkill(target uint32, number uint8, op packet.Opcode, cast *pendingCast) {
	ch := s.char
	if ch == nil {
		return
	}
	k, def, ok := s.knownSkill(number)
	if !ok {
		return
	}
	if s.evolutionGrant != 0 {
		return
	}
	s.resolveSkillDefinition(target, k, def, op, cast)
}

func (s *Session) resolveSkillDefinition(target uint32, k db.CharacterSkill, def *db.SkillDef, op packet.Opcode, cast *pendingCast) {
	if s.safeZoneBlocksSkill(def, op, target) {
		s.rejectSafeZoneSkill(k, cast != nil)
		return
	}
	var castPos *[2]float32
	if cast != nil {
		castPos = cast.targetPos
	}
	if isPotential(def) {
		return // triggers by itself (tickPotentials); it cannot be cast
	}
	if !s.definitionReady(def) {
		s.sendSkillMessage(k.SkillID, 215)
		return
	}
	if !def.MobSkill && !s.previousSkillCooling(def) {
		s.sendSkillMessage(k.SkillID, msgPreviousSkillRequired)
		return
	}
	if !s.attackMovementAllowed() {
		s.log.Printf("GX Server World skill release refused: airborne skill_id=%d prepared=%t", k.SkillID, castPos != nil)
		return
	}
	stats := s.stats()
	if s.cannotAct() || stats.Disabled[min(int(def.Kind), 2)] > 0 {
		s.log.Printf("GX Server World rejected skill use: category disabled skill_id=%d", k.SkillID)
		return
	}
	if sp, mp := s.skillCosts(def, &stats); stats.CurrentMP < mp || stats.CurrentSP < sp {
		s.log.Printf("GX Server World rejected skill use: not enough MP/SP skill_id=%d", k.SkillID)
		return
	}
	if kind := formOf(def); kind == proto.FormMonsterCopy || kind == proto.FormDisguise {
		s.castCopiedForm(op, k, def, target, stats, castPos != nil)
		return
	}
	switch def.TypeEffect {
	case typeEffectBuff, typeEffectHeal:
		s.friendlySkill(k, def, target, stats, castPos != nil)
	case typeEffectAttack, typeEffectDebuff:
		if op == opUserSkillPvP {
			s.userSkillAttack(k, def, target, castPos, stats)
		} else if def.TypeEffect == typeEffectDebuff && def.TypeDetail == detailDegeneration {
			s.sendSkillMessage(k.SkillID, 317) // it cannot be cast upon a monster
		} else {
			s.skillAttack(k, def, target, castPos, stats)
		}
	case typeEffectOnce:
		s.skillOnce(k, def, target, stats, cast)
	case typeEffectField:
		s.skillField(k, def, stats, castPos != nil)
	default:
		s.log.Printf("GX Server World rejected skill use: effect %d not usable skill_id=%d", def.TypeEffect, k.SkillID)
	}
}

func (s *Session) sendBuffStats(why string) {
	s.publishPlayerEffects()
	st := s.stats()
	s.send(buildAdditionalStats(&st), "0x0526 additional stats "+why)
	s.send(buildMaxHPMPSP(s.char, &st), "0x051F max hp/mp/sp "+why)
	s.send(buildCurrentHitpoints(&st), "0x0521 current hp/mp/sp "+why)
}

// processBuffs ticks heal-over-time and debuff damage, then expires
// applied skills (CUser::ChkKeepSkill / ChkEndTimeSkill).
func (s *Session) processBuffs() {
	s.processBuffsAt(time.Now())
}

func (s *Session) processBuffsAt(now time.Time) {
	if len(s.buffs) == 0 || s.char == nil || s.dead {
		return
	}
	if s.tickDebuffs(now) {
		return
	}
	for i := range s.buffs {
		b := &s.buffs[i]
		if b.Effect == typeEffectBuff && b.Detail == detailPeriodicHeal {
			s.tickPeriodicHeal(b, now)
		}
	}
	s.removeEffects(func(b *ActiveBuff) bool {
		if b.EndsAt.After(now) {
			return false
		}
		s.log.Printf("GX Server World applied skill expired skill_id=%d level=%d", b.SkillID, b.SkillLevel)
		return true
	}, "after expiry")
}

// persistBuffs snapshots remaining buff time (frozen while offline).
func (s *Session) persistBuffs() {
	if s.char == nil {
		return
	}
	if err := db.SaveAppliedBuffs(s.char.ID, s.buffSnapshot()); err != nil {
		s.log.Printf("GX Server World failed to save buffs: %v", err)
	}
}

func (s *Session) buffSnapshot() []db.AppliedBuff {
	return s.buffSnapshotExcept(nil)
}

func (s *Session) buffSnapshotExcept(exclude func(*ActiveBuff) bool) []db.AppliedBuff {
	var rows []db.AppliedBuff
	for _, b := range s.buffs {
		if b.MapCurse || b.Form.kind != proto.FormNone && b.Form.kind != proto.FormDegeneration || exclude != nil && exclude(&b) {
			continue
		}
		rows = append(rows, db.AppliedBuff{SkillID: b.SkillID, Level: b.SkillLevel, LeftSeconds: uint32(max(time.Until(b.EndsAt), 0) / time.Second)})
	}
	return rows
}

func (s *Session) persistSkillDelays() {
	if s.char != nil {
		s.logErr("save skill delays", db.SaveSkillDelays(s.char.ID, s.skillDelaySnapshot()))
	}
}

// skillDelaySnapshot captures changes to remaining cooldown.
func (s *Session) skillDelaySnapshot() []db.CharacterSkill {
	var changed []db.CharacterSkill
	for _, k := range s.skills {
		left := int16(0)
		if at, ok := s.skillCooldowns[k.SkillID]; ok {
			left = int16(min(max(time.Until(at), 0)/time.Second, 32767))
		}
		if left != k.Delay {
			k.Delay = left
			changed = append(changed, k)
		}
	}
	return changed
}

func restoreBuffs(charID uint32, log logPrinter) ([]ActiveBuff, error) {
	rows, err := db.AppliedBuffs(charID)
	if err != nil {
		return nil, err
	}
	var out []ActiveBuff
	now := time.Now()
	for _, r := range rows {
		def := db.MustDefs().Skill(r.SkillID, r.Level)
		if def == nil || def.TypeEffect != typeEffectBuff && !isPotential(def) || formOf(def) != proto.FormNone {
			continue
		}
		out = append(out, newBuff(uint32(len(out)+1), def, now, now.Add(time.Duration(r.LeftSeconds)*time.Second), log))
	}
	return out, nil
}

// skillWeaponAllowed follows CUser::EnableUseSkillUser: a skill needs an
// equipped weapon (error 7 otherwise) and, unless it is a shield skill, the
// weapon's type must be listed in the skill's NeedWeapon flags. Skills that
// list no weapon at all (item-granted buffs) are not restricted.
func (s *Session) skillWeaponAllowed(def *db.SkillDef) bool {
	listed := def.NeedShield
	for _, ok := range def.NeedWeapon {
		listed = listed || ok
	}
	if !listed {
		return true
	}
	w := equippedItem(s.items, weaponSlot)
	if w == nil || !IsWeaponItemType(w.Type) {
		return false
	}
	if def.NeedShield {
		return true
	}
	rt := int(realType(w.Type))
	return rt >= 1 && rt <= len(def.NeedWeapon) && def.NeedWeapon[rt-1]
}
