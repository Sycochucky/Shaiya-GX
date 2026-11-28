package app

import (
	"math"
	"math/rand/v2"

	"github.com/Spelunkern/Shaiya-GX/internal/clientlog"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/movespeed"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

type mobActor struct {
	entity        *game.Entity
	actor         *scene.CreatureActor
	modelLoaded   bool
	ground        mobGroundCache
	pos           render.V3
	position      remotePosition
	moveSpeedTier uint8
	yaw, heading  float64
	idleAt        float64
	attackTarget  uint32
	attackImpact  *mobAttackImpact
	incomingHits  int
	deathAt       float64
	deathFrozen   bool
	deathSound    bool
	deathNotified bool
	dead          bool
}

// Track enters as packets arrive so a move or attack in the same batch has
// the original position even before the model has been built.
func (a *App) trackMob(e *game.Entity) *mobActor {
	if a.wv == nil || e == nil || e.Kind != game.KindMob {
		return nil
	}
	if a.wv.mobs == nil {
		a.wv.mobs = map[uint32]*mobActor{}
	}
	m := a.wv.mobs[e.ID]
	if m == nil || m.entity != e {
		if m != nil {
			a.finishMobAttack(m)
		}
		yaw := rand.Float64() * (2 * math.Pi)
		m = &mobActor{entity: e, yaw: yaw, heading: yaw}
		a.wv.mobs[e.ID] = m
		a.updateMobPosition(m)
	}
	return m
}

func (a *App) loadMobActor(m *mobActor) {
	if m.modelLoaded || a.assets == nil {
		return
	}
	m.modelLoaded = true
	actor, err := a.assets.BuildMonster(m.entity.TypeID)
	if err != nil {
		clientlog.Warnf("mob %d type %d: %v", m.entity.ID, m.entity.TypeID, err)
	}
	m.actor = actor
	m.placeActor()
}

func (a *App) updateMobPosition(m *mobActor) {
	e := m.entity
	newMove := m.position.initialized && m.position.moveSerial != e.MoveSerial
	reset := m.position.initialized && m.position.resetSerial != e.PositionResetSerial
	if e.Dead && !reset && m.position.initialized {
		return
	}
	pos := m.position.update(e, a.clock)
	if m.position.duration > 0 {
		delta := m.position.to.Sub(pos)
		if newMove {
			m.turnToward(delta)
		}
		if !e.MobHeightKnown {
			// 0x0603 supplies a destination. Read the native pace on move
			// orders and whenever a speed packet changes the active tier.
			tier := uint8(2)
			if e.MobSpeedKnown {
				tier = movespeed.MobMoveTier(e.MobMoveSpeed)
			}
			if newMove || tier != m.moveSpeedTier {
				m.moveSpeedTier = tier
				m.position.from, m.position.start = pos, a.clock
				m.position.duration = math.Hypot(float64(delta[0]), float64(delta[2])) / a.legacyMobMoveSpeed(e, tier)
			}
		}
	}
	m.pos = m.ground.resolve(a.wv.world, pos, e.MobHeightKnown, e.PositionResetSerial)
	if reset {
		m.placeActor()
	}
}

func (a *App) legacyMobMoveSpeed(e *game.Entity, tier uint8) float64 {
	if a.assets == nil {
		return 2
	}
	def, err := a.assets.MonsterDefinition(e.TypeID)
	if err != nil {
		return 2
	}
	step, ms := def.NormalStep, def.NormalTimeMS
	if e.Motion != 0 || step == 0 || ms == 0 {
		step, ms = def.ChaseStep, def.ChaseTimeMS
	}
	if speed := movespeed.MobUnitsPerSecond(int(step), ms, tier); speed > 0 {
		return speed
	}
	return 2
}

func (m *mobActor) turnToward(dir render.V3) {
	if dir[0]*dir[0]+dir[2]*dir[2] > .000001 {
		m.heading = math.Atan2(float64(dir[0]), float64(dir[2]))
	}
}

func (m *mobActor) placeActor() {
	if m.actor != nil {
		sin, cos := math.Sincos(m.yaw)
		m.actor.SetTransform(render.Basis(m.pos, render.V3{-float32(sin), 0, -float32(cos)}, render.V3{0, 1, 0}))
	}
}

// Keep the run cycle across short gaps between chase steps. World checks mob
// movement every 50 ms and receives player movement samples every 100 ms.
const mobChaseAnimationGrace = .15

func (m *mobActor) locomotionAction(now float64) int {
	p := m.position
	if p.duration <= 0 || p.from[0] == p.to[0] && p.from[2] == p.to[2] {
		return 6
	}
	arrival := p.start + p.duration
	if m.entity.Motion != 0 && now < arrival+mobChaseAnimationGrace {
		return 1
	}
	if now < arrival {
		return 0
	}
	return 6
}

func (a *App) prepareMonsters() {
	effectBudget := 2
	v := a.wv
	v.mobPoseSamples.Reset()
	effects := a.clientSettings().SkillEffects
	a.updatePlayerAnimationRates()
	for id, m := range v.mobs {
		if e := a.state.Entities[id]; e == nil || e.Kind != game.KindMob {
			a.finishMobAttack(m)
			delete(v.mobs, id)
		}
	}
	for _, e := range a.state.Entities {
		m := a.trackMob(e)
		if m == nil {
			continue
		}
		a.loadMobActor(m)
		if m.actor == nil {
			continue
		}
		if m.deathFrozen && !e.Dead {
			m.position, m.ground = remotePosition{}, mobGroundCache{}
			m.deathFrozen, m.deathSound, m.deathNotified = false, false, false
			m.idleAt = 0
		}
		a.updateMobPosition(m)

		action := 6
		death := e.Dead && m.incomingHits == 0
		if e.Dead {
			if !m.deathFrozen {
				m.freezeCorpse(e, a.clock)
			}
		} else {
			if m.actor.Attacking() && m.attackTarget != 0 && m.attackTarget != e.ID {
				if target, ok := a.effectPosition(m.attackTarget); ok {
					m.turnToward(target.Sub(m.pos))
				}
			}
			turn := math.Remainder(m.heading-m.yaw, 2*math.Pi)
			m.yaw += turn * min(1, max(0, a.frameDT)*4.5)
			m.placeActor()
			action = m.locomotionAction(a.clock)
			if action != 6 || m.actor.Attacking() {
				m.idleAt = 0
			} else if m.idleAt == 0 {
				m.idleAt = a.clock + 5
			} else if a.clock >= m.idleAt {
				m.idleAt = a.clock + float64(3+rand.IntN(3))
				if rand.IntN(2) == 0 {
					m.actor.PlayIdleVariation()
				}
			}
		}
		if death {
			a.presentMobDeath(m)
		} else {
			if err := m.actor.SetAction(action); err != nil {
				a.assets.Missing[err.Error()] = true
			}
			m.actor.CorpseFade = m.updateCorpse(false, a.clock)
		}
		a.setMobAnimationRate(e, m.actor)
		m.actor.Update(float32(a.frameDT))
		if impact := m.attackImpact; impact != nil && (m.actor.Time >= impact.contact || !m.actor.Attacking()) {
			a.finishMobAttack(m)
		}
		if effects {
			v.mobPoseSamples.Sample(m.actor.Character)
		}
		m.actor.PrepareEffects(float32(a.frameDT), effects, !death, &effectBudget)
	}
}

func (a *App) presentMobDeath(m *mobActor) {
	if !m.deathFrozen {
		m.freezeCorpse(m.entity, a.clock)
	}
	a.finishMobAttack(m)
	if !m.dead {
		if err := m.actor.SetAction(5); err != nil {
			a.assets.Missing[err.Error()] = true
		}
	}
	if m.deathSound {
		m.deathSound = false
		a.playSound(m.actor.Sound(3), m.pos[0], m.pos[2])
	}
	m.actor.CorpseFade = m.updateCorpse(true, a.clock)
}

// updatePlayerAnimationRates runs with the monsters: remote characters take
// their rates from the speed indices the server broadcasts (0x051C).
func (a *App) updatePlayerAnimationRates() {
	for id, r := range a.wv.actors {
		e := a.state.Entities[id]
		if e == nil || e.Speed == nil || r.character == nil || e.Shape == nil {
			continue
		}
		base, have := a.weaponAttackTime(e.Shape.Equipment[5][0], e.Shape.Equipment[5][1])
		r.character.SetSpeedTiers(e.Speed.Move, e.Speed.Attack, base, have)
	}
}

func (a *App) drawMonsters() {
	s := a.scene3d
	v := a.wv
	frustum := render.ViewFrustum(s.View, s.VisibilityProjection())
	effects := a.clientSettings().SkillEffects
	for id, e := range a.state.Entities {
		if e.Kind != game.KindMob {
			continue
		}
		if cached := v.mobs[id]; cached != nil && cached.actor != nil {
			mob := cached.actor
			if mob.CorpseFade < 1 && mob.Visible(frustum, s.ClipPlane) {
				if !effects {
					v.mobPoseSamples.Sample(mob.Character)
				}
				if effects && !s.Shadow && mob.HasDrawEffects() {
					// Keep transparent effects in their original position
					// relative to the surrounding opaque geometry.
					a.flushCharacterBatch()
					mob.Draw(a.skinner, s)
					mob.DrawEffects(s)
				} else {
					v.crowdCharacters = append(v.crowdCharacters, mob.Character)
				}
			}
		}
	}
	a.flushCharacterBatch()
}

// Phoenix's deliberate alternative to native corpse sinking: hold the death
// pose for 3.5 seconds, then discard increasing pixel coverage over 1.5s.
func corpseFade(age float64) float32 { return float32(min(1, max(0, (age-3.5)/1.5))) }

func (m *mobActor) updateCorpse(dead bool, now float64) float32 {
	if dead && !m.dead {
		m.deathAt = now
	}
	m.dead = dead
	if !dead {
		return 0
	}
	return corpseFade(now - m.deathAt)
}

// Hold the last rendered position until the corpse is removed. A GX stop
// packet, when present, has already placed the actor at the server's drop point.
func (m *mobActor) freezeCorpse(e *game.Entity, now float64) {
	m.position = remotePosition{initialized: true, pos: m.pos, from: m.pos, to: m.pos, moveSerial: e.MoveSerial, resetSerial: e.PositionResetSerial, lastArrival: now}
	m.deathFrozen = true
	m.attackTarget = 0
}
