package app

import (
	"slices"

	"github.com/Spelunkern/Shaiya-GX/internal/attackspeed"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
)

type landingAction struct {
	target, selection uint32
	packet            []byte // nil for a basic attack
	skill             sdata.SkillKey
}

func (a *App) attackMovementAllowed() bool {
	v := a.wv
	return !v.jumpActive && !v.jumpPending && !v.jumping && !v.self.PlayingAction(8) &&
		(v.ladder.Active || v.world.PlayerSupported(v.pos) || v.world.SwimmingAt(v.pos[0], v.pos[1], v.pos[2]))
}

// An explicit combat order given in the air always waits for the landing and
// then goes out (with the usual short delay); it is never dropped.
func (a *App) deferAirAction(target uint32, packet []byte) bool {
	v := a.wv
	v.landingAction = landingAction{}
	if a.attackMovementAllowed() {
		return false
	}
	if a.state.Self.Mounted {
		return true
	}
	q := landingAction{target: target, selection: a.state.Target, packet: slices.Clone(packet)}
	if packet != nil && proto.Op(packet) != proto.GXEvolutionCast {
		k, ok := a.state.Skills[proto.NewReader(packet).U8()]
		if !ok {
			return true
		}
		q.skill = sdata.SkillKey{ID: k.ID, Level: k.Level}
	}
	a.skills.queuedSkill = nil
	v.landingAction = q
	return true
}

// Run after sending the final movement sample so World sees the landing
// before the combat request. Revalidate through the usual action paths.
func (a *App) updateLandingAction(in *platform.Input) {
	v := a.wv
	q := v.landingAction
	if q.target == 0 && q.packet == nil {
		return
	}
	if a.state.Self.Dead || a.state.Self.Mounted || in.HasKey(platform.KeyEscape) ||
		(q.target != 0 && q.target != a.state.SelfID && q.selection != a.state.Target) {
		v.landingAction = landingAction{}
		return
	}
	if a.world == nil || in.SuppressAutoRun || !a.attackMovementAllowed() {
		return
	}
	if q.packet == nil {
		if a.clock >= v.attackMoveUntil {
			v.landingAction = landingAction{}
			a.attackTarget()
		}
		return
	}
	if proto.Op(q.packet) == proto.GXEvolutionCast {
		v.landingAction = landingAction{}
		if d, ok := a.evolutionCastDefinition(q.packet); ok && a.validSkillTarget(d, q.target) && a.skillResourcesAvailable(d) &&
			!a.approachEvolutionCast(q.packet, q.target, d.AttackRange, d) {
			a.sendEvolutionCast(q.packet, q.target)
		}
		return
	}
	k, ok := a.state.Skills[proto.NewReader(q.packet).U8()]
	d := a.skillDefs[q.skill]
	if !ok || (sdata.SkillKey{ID: k.ID, Level: k.Level}) != q.skill || a.skills.cooldowns[k.ID] > a.clock ||
		!a.validSkillTarget(d, q.target) || !a.skillResourcesAvailable(d) {
		v.landingAction = landingAction{}
		return
	}
	if a.skillActionReady(k.ID, d) {
		v.landingAction = landingAction{}
		a.dispatchShortcut(q.packet)
	}
}

func (a *App) stopForAttack() bool {
	v := a.wv
	motion := uint8(0)
	if v.jumping {
		motion = 4 | proto.MotionFlag80
	}
	if v.moving || v.motion != motion {
		if err := a.world.Send(v.movePacket(motion)); err != nil {
			a.backToLoginID(12009)
			return false
		}
	}
	v.motion, v.sendT = motion, .1
	v.moving, v.travel, v.autoRun = false, false, false
	v.attackMoveUntil = a.clock + float64(attackspeed.MovementPauseMilliseconds)/1000
	return true
}

func (a *App) confirmAttackStop() {
	v := a.wv
	v.attackMoveUntil = max(v.attackMoveUntil, a.clock+float64(attackspeed.MovementPauseMilliseconds)/1000)
	v.travel, v.autoRun = false, false
}

// Call immediately before sending any skill, after any approach.
func (a *App) prepareSkillMovement(p []byte) bool {
	if len(p) != 7 || proto.Op(p) != proto.SkillUse && proto.Op(p) != proto.SkillUseNoTarget {
		return true
	}
	if a.state.Self.Mounted || !a.attackMovementAllowed() {
		return false
	}
	k, ok := a.state.Skills[proto.NewReader(p).U8()]
	if !ok {
		return true
	}
	d := a.skillDefs[sdata.SkillKey{ID: k.ID, Level: k.Level}]
	if !a.skillActionReady(k.ID, d) {
		a.queueSkill(p)
		return false
	}

	r := proto.NewReader(p)
	r.U8()
	if !a.skillRequestReach(d, r.U32()) {
		return false
	}
	if !a.stopForAttack() {
		return false
	}
	a.skills.queuedSkill = nil
	a.markCombatAction()
	a.skills.actionFloor = a.skillClipFloor(d)
	a.skills.lastSkillID, a.skills.lastSkillAt = k.ID, a.clock
	a.markRevolverSent(p)
	return true
}
