package app

import (
	"math/rand/v2"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

func (a *App) playerActor(id uint32) *scene.Character {
	if a.wv == nil {
		return nil
	}
	if id == a.state.SelfID {
		return a.wv.self
	}
	if actor := a.wv.actors[id]; actor != nil {
		return actor.character
	}
	return nil
}

func (a *App) playerWeapon(id uint32) uint8 {
	if id == a.state.SelfID {
		return a.state.Inventory[[2]uint8{0, 5}].Type
	}
	if e := a.state.Entities[id]; e != nil && e.Shape != nil {
		return e.Shape.Equipment[5][0]
	}
	return 0
}

// confirmedAnimation presents network outcomes; it never changes health.
func (a *App) confirmedAnimation(p []byte) {
	op := proto.Op(p)
	v := a.wv
	if v == nil {
		return
	}
	switch op {
	case proto.GXDrowningDamage:
		r := a.state.LastDrowningDamage
		a.showCombatNumberAmount(r.Target, r.Amount, false, false, combatDamage)
	case proto.GXFallDamage:
		r := a.state.LastFallDamage
		a.showCombatNumberAmount(r.Target, r.Amount, false, false, combatDamage)
	case proto.ChangeAppearance:
		// 0x56b840: ch_maready02.wav (range 10) and effect type 209
		// (DA_target_commercialB.EFT) on the character that changed.
		if m, err := proto.DecodeAppearanceChange(p); err == nil {
			a.appearanceChanged(m.ID)
		}
	case proto.CharacterMotion:
		m := a.state.LastMotion
		if m.ID == a.state.SelfID && (m.Motion == proto.MotionSitDown || m.Motion == proto.MotionStandUp) {
			v.sitPending = false
		}
		if m.Motion == proto.MotionSitDown || m.Motion == proto.MotionStandUp {
			if actor := a.playerActor(m.ID); actor != nil && !actor.PlayingAction(8) {
				action := 10
				if m.Motion == proto.MotionStandUp {
					action = 11
				}
				actor.PlayAction(action, false)
			}
		}
	case proto.SkillCastStarted:
		r := a.state.LastCast
		// The server may require preparation even when local SData says
		// instant. Its cast-start supersedes the predicted release and clock.
		if r.Source == a.state.SelfID {
			v.requestedSkill = nil
			v.requestedSkills = nil
		}

		if actor := a.playerActor(r.Source); actor != nil && !actor.PlayingAction(8) {
			a.loadSkills()
			d, _ := a.skillDefinition(r.ID, r.Level)
			a.faceSkillTarget(r.Source, r.Target)
			actor.PlayCast(scene.SkillCastAction(d.Animation, d.Effect), float32(d.ReadyTime)*.25)
		}
	case proto.GXSkillFailure:
		r := a.state.LastSkillFailure
		if r.Message == proto.MsgCannotUseSkill && a.state.LastCast.Source == a.state.SelfID &&
			a.state.LastCast.ID == r.Skill && v.self != nil {
			v.self.CancelCast()
		}
	case proto.SkillUse, proto.SkillUseNoTarget, proto.SkillUserDrain, proto.SkillMobDrain:
		r := a.state.LastSkill
		if r.Source == a.state.SelfID && a.state.Self.Mounted {
			return
		}
		if r.Source == a.state.SelfID && r.Result <= 4 && !(r.Result == 3 && r.Target == 0) {
			a.confirmAttackStop()
		}
		if prediction, current := a.takeInstantPrediction(r); prediction != nil {
			if current {
				v.requestedSkill = nil
				if r.Result > 4 || r.Result == 3 && r.Target == 0 {
					v.releaseSkillAction()
				}
			}
			return
		}
		predicted := a.matchingRequestedSkill(r)
		if r.Source == a.state.SelfID && v.requestedSkill != nil && !predicted {
			// An older area/multi-hit reply cannot replace a newer predicted combo.
			return
		}
		if r.Source == a.state.SelfID {
			v.requestedSkill = nil
		}
		if predicted {
			if r.Result > 4 || r.Result == 3 && r.Target == 0 {
				a.wv.releaseSkillAction()
			} else {
				return
			}
		}
		if actor := a.playerActor(r.Source); actor != nil {
			actor.CancelCast()
		}
		if r.Result > 4 || r.Result == 3 && r.Target == 0 {
			return
		}
		if actor := a.playerActor(r.Source); actor != nil && !actor.PlayingAction(8) {
			a.loadSkills()
			d, ok := a.skillDefs[sdata.SkillKey{ID: r.ID, Level: r.Level}]
			if !ok {
				return
			}
			if r.ID == 56 && v.dashes[r.Source].Source != 0 {
				return
			}
			weapon := a.playerWeapon(r.Source)
			a.faceSkillTarget(r.Source, r.Target)
			a.attackVariant++
			actor.PlaySkillAction(r.ID, scene.SkillAction(d.Animation, d.Attack, d.Effect, weapon, a.attackVariant))
		}
	case proto.CharacterMove:
		if actor := a.playerActor(proto.NewReader(p).U32()); actor != nil {
			if move, err := proto.DecodeUserMove(p); err == nil && proto.CastInterruptingMotion(move.Motion) {
				actor.CancelCast()
			}
			if move, err := proto.DecodeUserMove(p); err == nil && move.Motion&0x3f != 0 {
				cancelGesture(actor)
			}
		}
	case proto.MapNpcAttackPlayer, proto.MapNpcAttackMob:
		if hit, err := proto.DecodeNpcAttack(p); err == nil && (hit.Target == a.state.SelfID || hit.Target == a.state.Target) {
			a.showCombatNumber(hit.Target, hit.HP, false, false, combatDamage)
		}
	case proto.MobEnter, proto.GXMobEnter:
		a.trackMob(a.state.Entities[proto.NewReader(p).U32()])
	case proto.MobMove, proto.GXMobMove, proto.MobSpeed, proto.TargetMobGetState:
		if mob := a.trackMob(a.state.Entities[proto.NewReader(p).U32()]); mob != nil {
			a.updateMobPosition(mob)
		}
	case proto.MobLeave:
		id := proto.NewReader(p).U32()
		if mob := v.mobs[id]; mob != nil {
			a.finishMobAttack(mob)
			delete(v.mobs, id)
		}
	case proto.MobDeath:
		if mob := a.trackMob(a.state.Entities[proto.NewReader(p).U32()]); mob != nil {
			if !mob.deathNotified {
				mob.deathNotified, mob.deathSound = true, true
			}
		}
	case proto.MobAttackCharacter:
		a.startMobBasicAttack(a.state.LastAttack)
	case proto.CharacterMobAutoAttack, proto.CharacterUserAutoAttack:
		result := a.state.LastAttack
		if a.state.SelfID != 0 && result.Source == a.state.SelfID {
			a.markCombatAction()
			// Apply a swing already in flight after disengaging without
			// replacing the current pose or pausing ongoing movement.
			if (v.travel || v.moving) && v.approachTarget == 0 && !v.attackSent {
				a.deliverAttackResult(result, nil, "")
				return
			}
		}
		ranged := false
		if actor := a.playerActor(result.Source); actor != nil {
			if result.Result == 1 || rand.IntN(10) < 3 {
				a.playCharacterVoice(result.Source, voiceAttack)
			}
			typ := a.playerWeapon(result.Source)
			if result.Source == a.state.SelfID {
				a.confirmAttackStop()
				a.faceSkillTarget(result.Source, result.Target)
			}
			a.attackVariant++
			if !actor.PlayingAction(8) && !a.playFormAttack(result.Source, 0) && actor.PlayBasicAttack(typ, a.attackVariant) {
				ranged = a.startPhysicalProjectile(result, typ, actor)
				if ranged {
					// The string/shot cue follows the animation, not the missile's release,
					// so a replaced action or a fast attack speed cannot swallow it.
					a.wv.physicalProjectiles[len(a.wv.physicalProjectiles)-1].swingQueued = true
				}
				a.queueAttackSounds(result, typ, actor, ranged)
			}
		}
		if !ranged {
			a.deliverAttackResult(result, a.playerActor(result.Source), "")
		}
	case proto.CharacterDeath:
		id := proto.NewReader(p).U32()
		if c := a.playerActor(id); c != nil {
			c.PlayAction(9, true)
		}
	case proto.CharacterResurrect:
		id := proto.NewReader(p).U32()
		if c := a.playerActor(id); c != nil {
			c.ReleaseAction()
		}
	}
}

func (a *App) faceSkillTarget(source, target uint32) {
	if a.wv == nil || source == target {
		return
	}
	goal, ok := a.effectPosition(target)
	if !ok {
		return
	}
	if source == a.state.SelfID {
		dir := goal.Sub(a.wv.pos)
		dir[1] = 0
		if dir.Len() > .001 {
			a.wv.faceCombatTarget(dir.Norm())
		}
	} else if actor := a.wv.actors[source]; actor != nil && actor.character != nil {
		pos := a.visualEntityPosition(actor.entity)
		dir := goal.Sub(pos)
		dir[1] = 0
		if dir.Len() > .001 {
			actor.facing = dir.Norm()
		}
	}
}

// Facing is predicted on input, before any attack/cast confirmation. Publish
// the stationary direction before the action so peers see the same heading.
func (a *App) faceRequestedTarget(target uint32) bool {
	if a.wv == nil {
		return true
	}
	previous := a.wv.facing
	a.faceSkillTarget(a.state.SelfID, target)
	if a.world != nil && previous.Sub(a.wv.facing).Len() > .0001 {
		if err := a.world.Send(a.wv.movePacket(0)); err != nil {
			a.backToLoginID(12034)
			return false
		}
	}
	return true
}

type mobAttackImpact struct {
	result     proto.AttackResult
	target     *game.Entity
	contact    float32
	showNumber bool
}

func (a *App) startMobBasicAttack(result proto.AttackResult) {
	impact := mobAttackImpact{result: result, target: a.state.Entities[result.Target], showNumber: a.showsAttackResult(result)}
	if mob := a.trackMob(a.state.Entities[result.Source]); mob != nil && !mob.entity.Dead {
		a.loadMobActor(mob)
		if mob.actor != nil {
			action := mob.actor.BasicAttackAction()
			if a.playMobAttack(mob, result.Target, action) ||
				action == 3 && mob.actor.Definition.AttackAnimations[0] == 1 && a.playMobAttack(mob, result.Target, 2) {
				if impact.contact = mob.actor.AttackContactTime(); impact.contact > 0 {
					mob.attackImpact = &impact
					return
				}
			}
		}
	}
	a.presentMobAttack(impact)
}

func (a *App) playMobAttack(m *mobActor, target uint32, action int) bool {
	a.finishMobAttack(m)
	if !m.actor.PlayOnce(action) {
		return false
	}
	m.attackTarget = target
	a.playSound(m.actor.Sound(action-2), m.pos[0], m.pos[2])
	return true
}

func (a *App) finishMobAttack(m *mobActor) {
	if impact := m.attackImpact; impact != nil {
		m.attackImpact = nil
		a.presentMobAttack(*impact)
	}
}

func (a *App) presentMobAttack(impact mobAttackImpact) {
	r := impact.result
	if r.Target != a.state.SelfID && a.state.Entities[r.Target] != impact.target {
		return
	}
	a.playMobHitSound(r.Target, r.Result)
	a.playCharacterDamageVoice(r.Target, r.Result, r.HP)
	if impact.showNumber {
		a.showCombatNumber(r.Target, r.HP, r.Result == 1, r.Result == 3, combatDamage)
	}
}

func (a *App) showsAttackResult(r proto.AttackResult) bool {
	return r.Source == a.state.SelfID || r.Target == a.state.SelfID || r.Target == a.state.Target
}

// A confirmed hit retains its target's visual life until the animation or
// projectile reaches contact. The entity remains dead for targeting and HP.
func (a *App) waitForMobHit(target uint32, hp uint16, result uint8) *mobActor {
	if hp == 0 || result > 1 {
		return nil
	}
	m := a.trackMob(a.state.Entities[target])
	if m == nil || m.entity.Dead {
		return nil
	}
	m.incomingHits++
	return m
}

func (a *App) finishMobHit(m *mobActor) {
	if m == nil {
		return
	}
	m.incomingHits--
	if m.incomingHits == 0 && m.entity.Dead && m.actor != nil && a.state.Entities[m.entity.ID] == m.entity {
		a.presentMobDeath(m)
	} else {
		a.reactToMobHit(m)
	}
}

func (a *App) reactToMobHit(m *mobActor) {
	if m != nil && m.actor != nil && !m.entity.Dead && a.state.Entities[m.entity.ID] == m.entity && m.locomotionAction(a.clock) == 6 {
		m.actor.PlayOnce(7)
	}
}

// deliverAttackResult mirrors Game.exe 0x418030/0x41b300: the confirmed blow is
// stored on the attacker and applied when the swing crosses its contact
// marker. The number, the pain voice and the target's hit reaction all happen
// there. With two markers the damage is split in halves (first gets the
// larger share), one application per marker. Without a marker it applies now.
func (a *App) deliverAttackResult(result proto.AttackResult, actor *scene.Character, hitSound string) {
	v := a.wv
	showNumber := a.showsAttackResult(result)
	target := a.state.Entities[result.Target]
	apply := func(hp uint16, pending bool) func() {
		var mob *mobActor
		if pending {
			mob = a.waitForMobHit(result.Target, hp, result.Result)
		}
		return func() {
			if pending {
				defer a.finishMobHit(mob)
			}
			if result.Target != a.state.SelfID && a.state.Entities[result.Target] != target {
				return
			}
			if hp > 0 {
				if result.Result <= 1 {
					a.playEntitySound(result.Target, hitSound)
				}
				a.playCharacterDamageVoice(result.Target, result.Result, hp)
				if !pending && result.Result <= 1 {
					a.reactToMobHit(v.mobs[result.Target])
				}
			}
			if showNumber {
				a.showCombatNumber(result.Target, hp, result.Result == 1, result.Result == 3, combatDamage)
			}
		}
	}
	if actor != nil && len(v.pendingNumbers) < 62 {
		first, second := actor.ContactTimes()
		switch {
		case first > 0 && second > first && second > actor.Time:
			half := result.HP / 2
			firstPending := first > actor.Time
			firstContact := apply(result.HP-half, firstPending)
			if firstPending {
				v.pendingNumbers = append(v.pendingNumbers, pendingNumber{actionClock{actor, actor.ActionID(), first}, firstContact})
			} else {
				firstContact()
			}
			// The blow can hold a mob's death only until its first contact.
			v.pendingNumbers = append(v.pendingNumbers, pendingNumber{actionClock{actor, actor.ActionID(), second}, apply(half, false)})
			return
		case first > actor.Time:
			v.pendingNumbers = append(v.pendingNumbers, pendingNumber{actionClock{actor, actor.ActionID(), first}, apply(result.HP, true)})
			return
		}
	}
	apply(result.HP, false)()
}
