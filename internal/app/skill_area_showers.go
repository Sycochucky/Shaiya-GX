package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/effect"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
	"math/rand/v2"
)

func areaShowerSkill(id uint16) bool {
	_, ok := effect.SkillMeteor(id)
	return ok
}

type skillAreaParticle struct {
	pos, velocity render.V3
	instance      *scene.Effect
	grounded      bool
	groundAge     float32
}
type skillAreaShower struct {
	center, side, forward render.V3
	radius                int
	remaining, credit     float32
	profile               effect.Meteor
	file                  string
	particles             []skillAreaParticle
}

// Game.exe 0x450780/0x430530 creates a fixed world shower. 0x430a80
// emits five falling sequence-zero instances per second for five seconds,
// at signed integer offsets within ApplyRange, then sequence one on terrain.
func (a *App) startSkillAreaShower(r proto.SkillResult, d sdata.Skill, file string) {
	profile, enabled := effect.SkillMeteor(r.ID)
	if !enabled || len(a.wv.areaShowers) >= 16 {
		return
	}
	m, ok := a.impactTransform(skillImpact{target: r.Target, positionOnly: true})
	if !ok {
		return
	}
	center := render.V3{m[12], m[13], m[14]}
	source, ok := a.effectPosition(r.Source)
	if !ok {
		source = center
	}
	dir := center.Sub(source)
	dir[1] = 0
	dir = dir.Norm()
	if dir.Len() == 0 {
		dir = render.V3{0, 0, 1}
		if actor := a.playerActor(r.Source); actor != nil {
			dir = render.V3{actor.World[8], 0, actor.World[10]}.Norm()
		}
		if dir.Len() == 0 {
			dir = render.V3{0, 0, 1}
		}
	}
	side := render.V3{-dir[2], 0, dir[0]}
	shower := skillAreaShower{center: center, side: side, forward: dir.Scale(-1), radius: max(1, int(d.ApplyRange)), remaining: profile.Duration, profile: profile, file: file}
	a.wv.areaShowers = append(a.wv.areaShowers, shower)
}
func (a *App) updateSkillAreaShowers(dt float32) {
	rows := a.wv.areaShowers[:0]
	for _, shower := range a.wv.areaShowers {
		emitDT := min(dt, max(0, shower.remaining))
		shower.remaining -= dt
		shower.credit += emitDT * shower.profile.Rate
		for shower.credit >= 1 && len(shower.particles) < shower.profile.MaxLive {
			shower.credit--
			offset := func() float32 {
				v := float32(rand.IntN(shower.radius))
				if rand.IntN(2) == 0 {
					v = -v
				}
				return v
			}
			pos := shower.center.Add(render.V3{offset(), shower.profile.Height, offset()}).Sub(shower.side.Scale(shower.profile.Lateral))
			if x := a.loadSkillEffect(shower.file, shower.profile.FallingSequence); x != nil {
				shower.particles = append(shower.particles, skillAreaParticle{pos: pos, velocity: shower.side.Scale(shower.profile.Lateral).Add(render.V3{0, -shower.profile.Height, 0}), instance: x})
			}
		}
		alive := shower.particles[:0]
		for _, p := range shower.particles {
			if !p.grounded {
				previous := p.pos
				p.pos = p.pos.Add(p.velocity.Scale(dt))
				ground := func(point render.V3) float32 {
					if a.wv.world != nil && a.wv.world.HasTerrain() {
						return a.wv.world.Height(point[0], point[2])
					}
					return shower.center[1]
				}
				if p.pos[1] <= ground(p.pos) {
					p.pos = meteorContact(previous, p.pos, ground)
					floor := ground(p.pos)
					p.instance.FinishMeteor()
					p.grounded = true
					a.cameraShake(.5) // 0x430a80: every shard that lands renews a half-second shake
					if len(a.wv.impacts) < 32 {
						if x := a.loadSkillEffect(shower.file, shower.profile.ImpactSequence); x != nil {
							m := render.Identity()
							m[12], m[13], m[14] = p.pos[0], floor+shower.profile.ImpactHeight, p.pos[2]
							x.SetTransform(m)
							a.wv.impacts = append(a.wv.impacts, skillImpact{instance: x, fixedWorld: &m})
						}
					}
				}
			} else {
				p.groundAge += dt
			}
			if p.groundAge > shower.profile.Drain || p.grounded && p.instance.Done() {
				continue
			}
			m := render.Basis(p.pos, shower.forward, render.V3{0, 1, 0})
			p.instance.SetMeteorTransform(m)
			p.instance.Update(dt)
			alive = append(alive, p)
		}
		clear(shower.particles[len(alive):])
		shower.particles = alive
		if shower.remaining > 0 || len(alive) > 0 {
			rows = append(rows, shower)
		}
	}
	clear(a.wv.areaShowers[len(rows):])
	a.wv.areaShowers = rows
}

// Resolve the segment's first terrain crossing, so a long frame cannot carry
// a projectile beyond its diagonal landing point (Phoenix rain controller).
func meteorContact(from, to render.V3, ground func(render.V3) float32) render.V3 {
	low, high := float32(0), float32(1)
	delta := to.Sub(from)
	for range 10 {
		f := (low + high) * .5
		p := from.Add(delta.Scale(f))
		if p[1] > ground(p) {
			low = f
		} else {
			high = f
		}
	}
	p := from.Add(delta.Scale(high))
	p[1] = ground(p)
	return p
}
