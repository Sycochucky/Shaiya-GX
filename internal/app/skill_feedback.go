package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/effect"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

// Skill damage is presentation only. HP and resources follow server snapshots.
// Internal outcome kinds; never shown as text.
const (
	lblDamage = iota
	lblMiss
	lblHealing
	lblReflected
)

func (a *App) confirmedSkillFeedback(op proto.Opcode) {
	if a.wv != nil {
		a.wv.physicalSkillQueued = false
	}
	if a.state == nil {
		return
	}
	var source, target uint32
	var value uint16
	critical := false
	label := lblDamage
	switch op {
	case proto.SkillUse, proto.SkillUseNoTarget, proto.SkillUserDrain, proto.SkillMobDrain, proto.SkillPartyMember, proto.SkillMobOther, proto.MobSkillUse, proto.MobSkillDrain, proto.MobSkillOther:
		r := a.state.LastSkill
		if r.Result > 3 || r.Target == 0 {
			return
		}
		source, target, value = r.Source, r.Target, r.HP
		critical = r.Result == 1
		visible := source == a.state.SelfID || target == a.state.SelfID || target == a.state.Target
		if visible && r.Result < 3 {
			a.showDrainRecovery(r.Source, r.HealedHP, r.HealedSP, r.HealedMP, critical)
		}
		d, ok := a.skillDefinition(r.ID, r.Level)
		if a.wv != nil && ok {
			secondary := op == proto.SkillPartyMember || op == proto.SkillMobOther
			if a.queuePhysicalSkill(r, d, secondary) || a.deliverMeleeSkillHit(r, d, secondary) {
				return
			}
			if d.Effect == 1 || d.Effect == 3 || d.Effect == 5 {
				a.playCharacterDamageVoice(r.Target, r.Result, r.HP)
			}
		}
		if r.Result == 3 {
			label = lblMiss
		} else if ok && d.Effect == 2 {
			label = lblHealing
		}
	case proto.SkillUserExtraDamage, proto.SkillExtraDamage:
		r := a.state.LastSkillExtra
		source, target, value = r.Source, r.Target, r.HP
	case proto.SkillUserReflectedDamage, proto.SkillMobReflectedDamage:
		r := a.state.LastReflectedDamage
		source, target, value = r.Source, r.Target, r.HP
		label = lblReflected
	case proto.MobPeriodicDamage, proto.SkillKeep:
		r := a.state.LastPeriodicEffect
		target, value = r.Target, r.HP
		// A repeating skill plays its own effect on every target it reaches,
		// selected or not.
		if effect.PeriodicHit(r.Skill) && a.wv != nil {
			if d, ok := a.skillDefinition(r.Skill, r.Level); ok {
				a.startSkillImpact(r.Target, effect.AttributeDamageLibrary, effect.AttributeDamageSequence(d.Attribute))
			}
		}
		if target != a.state.SelfID && target != a.state.Target {
			return
		}
		if op == proto.SkillKeep {
			d, ok := a.skillDefinition(r.Skill, r.Level)
			if !ok {
				return
			}
			switch d.Effect {
			case 4:
				label = lblHealing
				if d.Detail == 84 { // Fire Thorn is a buff that hurts
					label = lblDamage
				}
			case 2:
				label = lblHealing
			case 1, 3:
				label = lblDamage
			default:
				return
			}
		}
	default:
		return
	}
	if source != a.state.SelfID && target != a.state.SelfID && target != a.state.Target {
		return
	}
	if label != lblMiss && value == 0 {
		return
	}
	kind := combatDamage
	if label == lblHealing {
		kind = combatHealHP
	}
	a.showCombatNumber(target, value, critical, label == lblMiss, kind)
}

// A landed direct melee skill shares weapon contact, reaction and death
// handling with a basic strike. Area recipient packets carry the damage but
// do not start another weapon impact at each target.
func (a *App) deliverMeleeSkillHit(r proto.SkillResult, d sdata.Skill, secondary bool) bool {
	if r.Level == 0 || r.Level == 100 || r.Target == 0 || r.Target == r.Source || r.Result > 1 || r.HP == 0 ||
		(d.Attack != 1 && d.Attack != 2) || (d.Effect != 1 && d.Effect != 3 && d.Effect != 5) {
		return false
	}
	weapon := a.playerWeapon(r.Source)
	if rangedFamily(weapon) != 0 {
		return false
	}
	var actor *scene.Character
	var hit string
	if !secondary {
		actor = a.playerActor(r.Source)
		slot := scene.SkillAction(d.Animation, d.Attack, d.Effect, weapon, a.attackVariant)
		if !actor.PlayingAction(slot) {
			actor = nil
		}
		_, hit = scene.WeaponSounds(weapon, r.Result == 1, a.attackVariant)
	}
	a.deliverAttackResult(proto.AttackResult{Source: r.Source, Target: r.Target, HP: r.HP, SP: r.SP, MP: r.MP, Result: r.Result}, actor, hit)
	return true
}
