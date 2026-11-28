package app

import (
	"cmp"

	"github.com/Spelunkern/Shaiya-GX/internal/movespeed"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

type runTrail struct {
	source   uint32
	instance *scene.Effect
}

// Game.exe 0x41c2fa..0x41c375: effect library 10 (Crun.EFT), faction
// sequence 0/1, one shot every >100 ms while running above normal on foot.
func runTrailEnabled(speed *proto.CharacterSpeed, running, mounted, dead, concealed, swimming bool) bool {
	return speed != nil && speed.Move > movespeed.Normal && speed.Move <= movespeed.Maximum && running && !mounted && !dead && !concealed && !swimming
}

func (a *App) updateRunEffects() {
	v := a.wv
	alive := v.runEffects[:0]
	for _, e := range v.runEffects {
		visible := e.source == a.state.SelfID && !a.state.Self.Concealed
		if p := a.state.Entities[e.source]; p != nil && !p.Concealed {
			visible = true
		}
		if !visible {
			continue
		}
		e.instance.Update(float32(a.frameDT))
		if !e.instance.Done() {
			alive = append(alive, e)
		}
	}
	clear(v.runEffects[len(alive):])
	v.runEffects = alive
	if !a.clientSettings().SkillEffects {
		return
	}
	if v.runEffectAt == nil {
		v.runEffectAt = map[uint32]float64{}
	}
	for id := range v.runEffectAt {
		if id != a.state.SelfID && a.state.Entities[id] == nil {
			delete(v.runEffectAt, id)
		}
	}
	spawn := func(id uint32, faction uint8, pos, facing render.V3) {
		if last, ok := v.runEffectAt[id]; ok && a.clock-last <= .1 {
			return
		}
		v.runEffectAt[id] = a.clock
		// Bound particle work for crowded views; never replay a backlog after lag.
		if len(v.runEffects) >= 512 {
			return
		}
		x := a.loadSkillEffect("Effect/Crun.EFT", int(min(faction, 1)))
		if x == nil {
			return
		}
		// Native 0x450340 subtracts .9 from the logical actor origin;
		// GX already stores feet, so no extra vertical offset belongs here.
		x.SetTransform(render.Basis(pos, facing, render.V3{0, 1, 0}))
		v.runEffects = append(v.runEffects, runTrail{id, x})
	}
	self := &a.state.Self
	if runTrailEnabled(self.Speed, v.moving && !v.walking && v.motion&0x3f != 2, self.Mounted, self.Dead || self.Sitting, self.Concealed, v.world.SwimmingAt(v.pos[0], v.pos[1], v.pos[2])) {
		facing := v.visualFacing
		if facing.Len() == 0 {
			facing = v.facing
		}
		spawn(a.state.SelfID, a.state.Faction.Faction, v.pos, facing)
	}
	for id, r := range v.actors {
		e := a.state.Entities[id]
		if e == nil || r.character == nil || e.Shape == nil {
			continue
		}
		running := e.Motion&0x3f != 0 && e.Motion&0x3f != 2
		pos := a.visualEntityPosition(e)
		if runTrailEnabled(e.Speed, running, e.Mounted, e.Dead || e.Sitting, e.Concealed, v.world.SwimmingAt(pos[0], pos[1], pos[2])) {
			spawn(id, e.Shape.Faction, pos, cmp.Or(r.facing, r.heading.vector(e)))
		}
	}
}

func (a *App) drawRunEffects() {
	if !a.clientSettings().SkillEffects {
		return
	}
	for _, e := range a.wv.runEffects {
		e.instance.Draw(a.scene3d)
	}
}
