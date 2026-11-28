package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

// Only the actor animation is predicted. Results, impacts and resources still
// require server packets. An acknowledgment must not restart this same action.
type requestedSkillPresentation struct {
	key     sdata.SkillKey
	target  uint32
	actor   *scene.Character
	action  uint64
	started float64
}

// predictInstantSkillPose is off: see requestedSkillPresentation.
const predictInstantSkillPose = false

func (a *App) requestedSkillPresentation(packet []byte) {
	if a.wv == nil || a.state.Self.Mounted || len(packet) != 7 || proto.Op(packet) != proto.SkillUse && proto.Op(packet) != proto.SkillUseNoTarget {
		return
	}
	r := proto.NewReader(packet)
	number, target := r.U8(), r.U32()
	k, ok := a.state.Skills[number]
	if !ok {
		return
	}
	key := sdata.SkillKey{ID: k.ID, Level: k.Level}
	d, ok := a.skillDefs[key]
	if !ok || !a.validSkillTarget(d, target) {
		return
	}
	actor := a.wv.self
	if actor == nil {
		return
	}
	// The original starts the animation, cast and impact effects together when
	// the server's result arrives (0x57bdc0); predicting the pose at send time
	// left the effects trailing it. Timed casts start on the cast-start packet.
	if d.ReadyTime > 0 || k.ID == 56 || !predictInstantSkillPose {
		return
	}
	p := &requestedSkillPresentation{key: key, target: target, actor: actor, started: a.clock}
	a.attackVariant++
	played := actor.PlaySkillAction(k.ID, scene.SkillAction(d.Animation, d.Attack, d.Effect, a.playerWeapon(a.state.SelfID), a.attackVariant))
	if played {
		p.action = actor.ActionID()
		a.wv.requestedSkill = p
		if d.Target < 4 && d.Detail != 22 {
			if len(a.wv.requestedSkills) >= 32 {
				a.wv.requestedSkills = a.wv.requestedSkills[1:]
			}
			a.wv.requestedSkills = append(a.wv.requestedSkills, p)
		}
	}
}

func (a *App) updateRequestedSkillPresentation() {
	kept := a.wv.requestedSkills[:0]
	for _, p := range a.wv.requestedSkills {
		if p.actor == a.wv.self && a.clock-p.started < 15 {
			kept = append(kept, p)
		}
	}
	clear(a.wv.requestedSkills[len(kept):])
	a.wv.requestedSkills = kept
	p := a.wv.requestedSkill
	if p == nil {
		return
	}
	if p.actor != a.wv.self || p.actor.ActionID() != p.action || a.clock-p.started > 15 {
		a.wv.requestedSkill = nil
		return
	}
}

func (a *App) matchingRequestedSkill(r proto.SkillResult) bool {
	p := a.wv.requestedSkill
	return p != nil && r.Source == a.state.SelfID && p.key == (sdata.SkillKey{ID: r.ID, Level: r.Level}) && (p.target == r.Target || r.Result > 4) && p.actor == a.wv.self && p.actor.ActionID() == p.action && a.clock-p.started < 15
}

// Older instant acknowledgments must not rewind a newer combo animation.
// Area/multi-hit skills retain their existing announcement path.
func (a *App) takeInstantPrediction(r proto.SkillResult) (*requestedSkillPresentation, bool) {
	if r.Source != a.state.SelfID {
		return nil, false
	}
	for i, p := range a.wv.requestedSkills {
		if p.key != (sdata.SkillKey{ID: r.ID, Level: r.Level}) || p.target != r.Target && r.Result <= 4 && !(r.Result == 3 && r.Target == 0) {
			continue
		}
		a.wv.requestedSkills = append(a.wv.requestedSkills[:i], a.wv.requestedSkills[i+1:]...)
		current := p == a.wv.requestedSkill && p.actor == a.wv.self && p.actor != nil && p.actor.ActionID() == p.action
		return p, current
	}
	return nil, false
}
func (a *App) rejectInstantPrediction(id uint16) {
	if a.wv == nil {
		return
	}
	for _, p := range a.wv.requestedSkills {
		if p.key.ID == id {
			_, current := a.takeInstantPrediction(proto.SkillResult{Source: a.state.SelfID, ID: id, Level: p.key.Level, Result: 5})
			if current {
				a.wv.releaseSkillAction()
				a.wv.requestedSkill = nil
			}
			return
		}
	}
}
