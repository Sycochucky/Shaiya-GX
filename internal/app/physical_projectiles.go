package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/effect"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
	"math"
)

type physicalProjectile struct {
	cue              actionClock
	result           proto.AttackResult
	part             *scene.Part
	weapon           uint8
	pos, destination render.V3
	speed, age       float32
	numberHeight     float32
	released         bool
	swingQueued      bool // the cast or auto-attack already owns the release cue
	hitSound         string
	impact           string
	impactSound      string
	sequence         int
	mob              *mobActor
	showNumber       bool
}

func rangedFamily(weapon uint8) uint8 {
	if f := effect.WeaponFamily(weapon); f != 0 {
		weapon = f
	}
	if weapon == 11 || weapon == 13 || weapon == 14 {
		return weapon
	}
	return 0
}
func physicalReleaseTime(weapon, index uint8) float32 {
	phase := float32(30)
	if weapon == 11 {
		phase = 15
	}
	rate := [10]float32{15, 18, 21, 24, 27, 30, 34, 38, 42, 45}
	return phase / rate[min(index, 9)]
}
func (a *App) startPhysicalProjectile(r proto.AttackResult, weapon uint8, actor *scene.Character) bool {
	kind := rangedFamily(weapon)
	if kind == 0 || a.assets == nil || len(a.wv.physicalProjectiles) >= 128 || r.Result > 1 && r.Result != 3 {
		return false
	}
	destination, ok := a.effectPosition(r.Target)
	if !ok {
		return false
	}
	var part *scene.Part
	if kind == 11 {
		part = actor.PrimaryWeaponPart()
	} else {
		part, _ = a.assets.ArrowModel()
	}
	if part == nil {
		return false
	}
	index := uint8(5)
	if r.Source == a.state.SelfID && a.state.Self.Speed != nil {
		index = a.state.Self.Speed.Attack
	} else if e := a.state.Entities[r.Source]; e != nil && e.Speed != nil {
		index = e.Speed.Attack
	}
	_, hit := scene.WeaponSounds(weapon, r.Result == 1, a.attackVariant)
	a.wv.physicalProjectiles = append(a.wv.physicalProjectiles, physicalProjectile{
		numberHeight: a.combatNumberHeight(r.Target), cue: actionClock{actor, actor.ActionID(), physicalReleaseTime(kind, index)},
		result: r, part: part, weapon: kind, destination: destination.Add(render.V3{0, .9, 0}), hitSound: hit,
		mob: a.waitForMobHit(r.Target, r.HP, r.Result), showNumber: a.showsAttackResult(r),
	})
	return true
}
func physicalFlightMatrix(pos, direction render.V3, weapon uint8) render.M4 {
	up := render.V3{0, 1, 0}
	if math.Abs(float64(direction.Norm().Dot(up))) > .99 {
		up = render.V3{1, 0, 0}
	}
	// Original 3DO arrow points down local -Z; retain the source row-vector basis.
	basis := render.Basis(pos, direction.Scale(-1), up)
	if weapon == 11 {
		rz := render.QuatMatrix([4]float32{0, 0, .70710677, .70710677}, [3]float32{})
		ry := render.QuatMatrix([4]float32{0, .70710677, 0, .70710677}, [3]float32{})
		return rz.Mul(ry).Mul(basis)
	}
	return basis
}
func (p *physicalProjectile) move(destination render.V3, dt float32) bool {
	p.destination = destination
	delta := destination.Sub(p.pos)
	distance := delta.Len()
	if p.speed == 0 {
		p.speed = distance * 5
	}
	step := max(0, dt) * p.speed
	if distance <= step+.005 {
		p.pos = destination
		return true
	}
	p.pos = p.pos.Add(delta.Norm().Scale(step))
	return false
}
func (a *App) finishPhysicalProjectile(p *physicalProjectile) {
	p.cue.actor.RestoreThrownWeapon(p.cue.action)
	r := p.result
	if p.showNumber || a.showsAttackResult(r) {
		if _, ok := a.effectPosition(r.Target); ok {
			a.showCombatNumber(r.Target, r.HP, r.Result == 1, r.Result == 3, combatDamage)
		} else {
			a.emitCombatNumberAt(p.destination.Add(render.V3{0, p.numberHeight - .9, 0}), uint32(r.HP), r.Result == 1, r.Result == 3, combatDamage)
		}
	}
	if r.HP > 0 && r.Result <= 1 {
		a.playSound(p.hitSound, p.destination[0], p.destination[2])
		a.playCharacterDamageVoice(r.Target, r.Result, r.HP)
	}
	if r.Result <= 1 {
		a.playSound(p.impactSound, p.destination[0], p.destination[2])
		if p.impact != "" {
			a.startSkillImpact(r.Target, p.impact, p.sequence)
		}
	}
	a.finishMobHit(p.mob)
}
func (a *App) updatePhysicalProjectiles(dt float32) {
	kept := a.wv.physicalProjectiles[:0]
	for _, p := range a.wv.physicalProjectiles {
		p.age += max(0, dt)
		if !p.released {
			due, alive := p.cue.ready(a.playerActor(p.result.Source))
			if !alive {
				a.finishPhysicalProjectile(&p)
				continue
			}
			if !due {
				kept = append(kept, p)
				continue
			}
			p.released = true
			p.pos, _ = a.effectPosition(p.result.Source)
			p.pos[1] += 1.2
			if m, ok := p.cue.actor.PrimaryWeaponMatrix(); ok {
				p.pos = render.V3{m[12], m[13], m[14]}
			}
			if p.weapon == 11 {
				p.cue.actor.HideThrownWeapon(p.cue.action)
			}
			if !p.swingQueued {
				swing, _ := scene.WeaponSounds(p.weapon, p.result.Result == 1, 0)
				a.playSound(swing, p.pos[0], p.pos[2])
			}
		}
		destination := p.destination
		if target, ok := a.effectPosition(p.result.Target); ok {
			destination = target.Add(render.V3{0, .9, 0})
		}
		if p.move(destination, dt) || p.age > 10 {
			a.finishPhysicalProjectile(&p)
			continue
		}
		kept = append(kept, p)
	}
	clear(a.wv.physicalProjectiles[len(kept):])
	a.wv.physicalProjectiles = kept
}
func (a *App) drawPhysicalProjectiles() {
	if len(a.wv.physicalProjectiles) == 0 {
		return
	}
	a.scene3d.Begin()
	for _, p := range a.wv.physicalProjectiles {
		var matrix render.M4
		if p.released {
			matrix = physicalFlightMatrix(p.pos, p.destination.Sub(p.pos), p.weapon)
		} else {
			// Native bows show their nocked arrow in the same weapon attachment space.
			if p.weapon != 13 {
				continue
			}
			var ok bool
			matrix, ok = p.cue.actor.PrimaryWeaponMatrix()
			if !ok {
				continue
			}
		}
		a.scene3d.DrawModel(p.part.Mesh, p.part.Tex, matrix, -1, p.part.Alpha)
	}
}

func (a *App) queuePhysicalSkill(r proto.SkillResult, d sdata.Skill, secondary bool) bool {
	if r.Level == 100 || r.Target == 0 || r.Target == r.Source || r.Result > 3 || (d.Attack != 1 && d.Attack != 2) || (d.Effect != 1 && d.Effect != 3 && d.Effect != 5) {
		return false
	}
	actor := a.playerActor(r.Source)
	if actor == nil {
		return false
	}
	hit := proto.AttackResult{Result: r.Result, Source: r.Source, Target: r.Target, HP: r.HP, SP: r.SP, MP: r.MP}
	if !a.startPhysicalProjectile(hit, a.playerWeapon(r.Source), actor) {
		return false
	}
	p := &a.wv.physicalProjectiles[len(a.wv.physicalProjectiles)-1]
	p.impact, p.sequence = effect.ImpactLibrary(r.ID), int(r.Level)-1
	if secondary {
		p.impact, p.sequence = skillSecondaryLibrary(r, d)
		p.swingQueued = true
	} else {
		p.impactSound = skillResultEffectSound(r, d)
	}
	a.wv.physicalSkillQueued = true
	return true
}
