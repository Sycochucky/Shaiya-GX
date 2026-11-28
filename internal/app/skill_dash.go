package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"time"
)

func (a *App) skillPhysicsTime() float64 {
	return a.state.PhysicsTime + time.Since(a.physicsClockAt).Seconds()
}
func (a *App) receiveSkillDash(d proto.SkillDash) {
	v := a.wv
	if v == nil {
		return
	}
	if d.Stopped {
		delete(v.dashes, d.Source)
		pos := render.V3(d.To)
		if d.Source == a.state.SelfID {
			v.place(pos)
			a.state.Self.X, a.state.Self.Y, a.state.Self.Z = pos[0], pos[1], pos[2]
		} else if e := a.state.Entities[d.Source]; e != nil {
			e.X, e.Y, e.Z = pos[0], pos[1], pos[2]
		}
		if actor := a.playerActor(d.Source); actor != nil && actor.PlayingAction(101) {
			actor.ReleaseAction()
		}
		return
	}
	if v.dashes == nil {
		v.dashes = map[uint32]proto.SkillDash{}
	}
	if a.physicsClockAt.IsZero() {
		a.state.PhysicsTime = d.Started
		a.physicsClockAt = time.Now()
	}
	v.dashes[d.Source] = d
	if actor := a.playerActor(d.Source); actor != nil {
		actor.PlaySkillAction(56, 101)
	}
	if d.Source == a.state.SelfID {
		v.travel, v.autoRun, v.moving = false, false, false
		v.evading = false
		v.requestedSkill = nil
		v.requestedSkills = nil
		v.jumping = false
	}
}
func (a *App) updateSkillDashes() bool {
	v := a.wv
	active := false
	for id, d := range v.dashes {
		elapsed := a.skillPhysicsTime() - d.Started
		dead := a.state.Self.Dead
		if id != a.state.SelfID {
			e := a.state.Entities[id]
			if e == nil {
				delete(v.dashes, id)
				continue
			}
			dead = e.Dead
		}
		if dead || elapsed >= d.Duration() {
			if !dead && id == a.state.SelfID {
				v.place(render.V3(d.To))
			}
			delete(v.dashes, id)
			if actor := a.playerActor(id); actor != nil && actor.PlayingAction(101) {
				actor.ReleaseAction()
			}
			continue
		}
		pos := render.V3(d.Position(elapsed))
		actor := a.playerActor(id)
		if actor != nil && !actor.PlayingAction(101) {
			actor.PlaySkillAction(56, 101)
		}
		if actor != nil && actor.PlayingAction(101) {
			actor.SetActionFrame(d.Frame(elapsed))
		}
		if id == a.state.SelfID {
			active = true
			v.place(pos)
			a.state.Self.X, a.state.Self.Y, a.state.Self.Z = pos[0], pos[1], pos[2]
			v.facing = render.V3(d.To).Sub(render.V3(d.From))
			v.facing[1] = 0
			v.facing = v.facing.Norm()
			v.attackMoveUntil = max(v.attackMoveUntil, a.clock+.05)
		} else {
			e := a.state.Entities[id]
			e.X, e.Y, e.Z = pos[0], pos[1], pos[2]
		}
	}
	return active
}
