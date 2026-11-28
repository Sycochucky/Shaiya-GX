package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/clientlog"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

type skillProjectile struct {
	target     uint32
	sequence   int
	impact     string
	sound      string
	pos        render.V3
	age, speed float32
	delay      float32
	instance   *scene.Effect
	arrived    bool
	// follow keeps mesh bodies (the Choir ring) attached to the flying
	// effect instead of leaving a trail of world-space copies.
	follow bool
}
type skillImpact struct {
	fixedWorld       *render.M4 // Native area announcements capture their release location.
	authoredLifetime bool
	positionOnly     bool // world-aligned spawn anchor, without actor rotation
	target           uint32
	age              float32
	instance         *scene.Effect
}

// Native magic mode 2 (0x57e320 -> 0x450650 -> 0x40e760): start at
// caster logical Y+0.7 (feet+1.6), wait 12 keys at default 30/s, then
// speed=initial distance*5. Time already spent in release is not waited twice.
// HP and other gameplay outcomes have already been applied from the server.
func (a *App) startSkillProjectile(r proto.SkillResult, file, impact, sound string) {
	if len(a.wv.projectiles) >= 32 {
		a.playEntitySound(r.Target, sound)
		return
	}
	source, ok := a.effectPosition(r.Source)
	if !ok {
		return
	}
	if _, ok := a.effectPosition(r.Target); !ok {
		return
	}
	x := a.loadSkillEffect(file, int(r.Level)-1)
	if x == nil {
		a.playEntitySound(r.Target, sound)
		return
	}
	elapsed := float32(0)
	if actor := a.playerActor(r.Source); actor != nil && actor.ActionActive() {
		elapsed = actor.Time
	}
	p := skillProjectile{target: r.Target, sequence: int(r.Level) - 1, impact: impact, sound: sound, pos: source.Add(render.V3{0, 1.6, 0}), instance: x, delay: .4, age: elapsed, follow: r.ID >= 164 && r.ID <= 167}
	target, _ := a.effectPosition(r.Target)
	target[1] += .9
	flightElapsed := max(0, elapsed-p.delay)
	if flightElapsed > 0 {
		distance := target.Sub(p.pos).Len()
		dir := p.move(target, flightElapsed)
		if p.arrived {
			// At distance*5 the authored flight takes 0.2s. A late result
			// must not replay the launch and the whole flight after release.
			travel := float32(0)
			if p.speed > 0 {
				travel = distance / p.speed
			}
			a.playEntitySound(r.Target, p.sound)
			a.startSkillImpactAt(r.Target, impact, p.sequence, max(0, flightElapsed-travel))
			return
		}
		p.setTransform(render.Basis(p.pos, dir, render.V3{0, 1, 0}))
		p.instance.Update(flightElapsed)
	}
	a.wv.projectiles = append(a.wv.projectiles, p)
}
func (a *App) loadSkillEffect(file string, sequence int) *scene.Effect {
	if file == "" {
		return nil
	}
	x, err := scene.LoadEffect(a.assets, file, sequence)
	v := a.wv
	if v.effectErrors == nil {
		v.effectErrors = map[string]bool{}
	}
	if !v.effectErrors[file] && (err != nil || x.Skipped > 0) {
		if err != nil {
			clientlog.Warnf("skill effect %s: %v", file, err)
		} else {
			clientlog.Debugf("skill effect %s: %d unsupported emitters omitted", file, x.Skipped)
		}
		v.effectErrors[file] = true
	}
	return x
}
func (a *App) effectPosition(id uint32) (render.V3, bool) {
	if id == a.state.SelfID {
		return a.wv.pos, true
	}
	if e := a.state.Entities[id]; e != nil {
		return a.visualEntityPosition(e), true
	}
	return render.V3{}, false
}
func (a *App) effectTargetMatrix(id uint32) (render.M4, bool) {
	if c := a.playerActor(id); c != nil {
		// Native 0x417ae0 subtracts .9 from the logical origin. Character.World
		// is already at the feet here; subtracting it again buries ground FX.
		return c.World, true
	}
	if c := a.wv.mobs[id]; c != nil && c.actor != nil {
		return c.actor.World, true
	}
	if p, ok := a.effectPosition(id); ok {
		return render.Basis(p, render.V3{0, 0, 1}, render.V3{0, 1, 0}), true
	}
	return render.M4{}, false
}
func (a *App) updateResolvedEffects() {
	v := a.wv
	initialImpacts := len(v.impacts)
	a.updatePlayerFeedback()
	a.updateSkillImpactCues()
	dt := float32(min(max(a.frameDT, 0), .25))
	a.updateSkillReleaseEffects(dt)
	a.updateSkillAreaShowers(dt)
	a.updatePhysicalProjectiles(dt)
	kept := v.projectiles[:0]
	for _, p := range v.projectiles {
		p.age += dt
		target, ok := a.effectPosition(p.target)
		if !ok || p.age > 10 {
			p.arrived = true
			p.instance.Stop()
		}
		if !p.arrived && p.age >= p.delay {
			target[1] += .9 // both users and mobs are stored at their feet
			dir := p.move(target, dt)
			if p.arrived {
				p.instance.Stop()
				if p.follow {
					p.instance.FinishMeteor()
				}
				a.playEntitySound(p.target, p.sound)
				a.startSkillImpact(p.target, p.impact, p.sequence)
			}
			p.setTransform(render.Basis(p.pos, dir, render.V3{0, 1, 0}))
		}
		if p.age < p.delay && !p.arrived {
			kept = append(kept, p)
			continue
		}
		p.instance.Update(min(dt, max(0, p.age-p.delay)))
		if p.instance.Done() && p.arrived {
			continue
		}
		kept = append(kept, p)
	}
	clear(v.projectiles[len(kept):])
	v.projectiles = kept
	for i := range v.impacts {
		p := &v.impacts[i]
		advance := dt
		if i >= initialImpacts {
			advance = 0
		}
		p.age += advance
		if m, ok := a.impactTransform(*p); ok {
			p.instance.SetTransform(m)
		} else {
			p.instance.Stop()
		}
		if !p.authoredLifetime && p.age >= 5 {
			p.instance.Stop()
		} // Bound lingering authored loops in this first delivery.
		v.effectUpdates = append(v.effectUpdates, scene.EffectUpdate{Instance: p.instance, Delta: advance})
	}
	v.updateQueuedEffects()
	alive := v.impacts[:0]
	for _, p := range v.impacts {
		if p.instance.Done() {
			continue
		}
		alive = append(alive, p)
	}
	clear(v.impacts[len(alive):])
	v.impacts = alive
}

func (a *App) drawResolvedEffects() {
	a.drawPhysicalProjectiles()
	if !a.clientSettings().SkillEffects {
		return
	}
	v := a.wv
	for _, area := range v.areaShowers {
		for _, p := range area.particles {
			v.visibleEffects = append(v.visibleEffects, p.instance)
		}
	}
	for _, r := range v.releaseEffects {
		v.visibleEffects = append(v.visibleEffects, r.instance)
	}
	for _, p := range v.projectiles {
		if p.age >= p.delay || p.arrived {
			v.visibleEffects = append(v.visibleEffects, p.instance)
		}
	}
	for _, p := range v.impacts {
		v.visibleEffects = append(v.visibleEffects, p.instance)
	}
	v.drawQueuedEffects(a.scene3d)
}

func (p *skillProjectile) move(target render.V3, dt float32) render.V3 {
	delta := target.Sub(p.pos)
	distance := delta.Len()
	if p.speed == 0 {
		p.speed = distance * 5
	}
	step := p.speed * min(dt, max(0, p.age-p.delay))
	if distance <= step+.005 {
		p.pos = target
		p.arrived = true
	} else {
		p.pos = p.pos.Add(delta.Norm().Scale(step))
	}
	return delta.Norm()
}

func (a *App) impactTransform(p skillImpact) (render.M4, bool) {
	if p.fixedWorld != nil {
		return *p.fixedWorld, true
	}
	if !p.positionOnly {
		return a.effectTargetMatrix(p.target)
	}
	point, ok := a.effectPosition(p.target)
	if e := a.state.Entities[p.target]; e != nil && !e.MobHeightKnown {
		point = a.skillTargetPosition(e)
	}
	matrix := render.Identity()
	matrix[12], matrix[13], matrix[14] = point[0], point[1], point[2]
	return matrix, ok
}

func (p *skillProjectile) setTransform(m render.M4) {
	if p.follow {
		p.instance.SetMeteorTransform(m)
	} else {
		p.instance.SetTransform(m)
	}
}
