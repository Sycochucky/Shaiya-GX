package app

import (
	"math"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

// remotePosition smooths only presentation. Entity coordinates remain the
// server snapshot for gameplay; large jumps and explicit respawns snap.
type remotePosition struct {
	initialized                  bool
	snapped                      bool
	pos, from, to                render.V3
	moveSerial, resetSerial      uint64
	start, duration, lastArrival float64
}

func (r *remotePosition) update(e *game.Entity, now float64) render.V3 {
	r.snapped = false
	target := render.V3{e.X, e.Y, e.Z}
	if !r.initialized {
		r.initialized, r.snapped = true, true
		r.pos, r.from, r.to = target, target, target
		r.moveSerial, r.resetSerial = e.MoveSerial, e.PositionResetSerial
		r.lastArrival = now
		return r.pos
	}
	if r.duration > 0 {
		t := float32(min(max((now-r.start)/r.duration, 0), 1))
		r.pos = r.from.Add(r.to.Sub(r.from).Scale(t))
	}
	if e.PositionResetSerial != r.resetSerial {
		r.snapped = true
		r.pos, r.from, r.to = target, target, target
		r.duration = 0
		r.resetSerial, r.moveSerial = e.PositionResetSerial, e.MoveSerial
		r.lastArrival = now
		return r.pos
	}
	if e.MoveSerial != r.moveSerial {
		delta := target.Sub(r.pos)
		snap := delta.Dot(delta) > 400 || math.Abs(float64(delta[1])) > 8
		if e.Kind == game.KindMob {
			snap = !e.MobHeightKnown && delta[0]*delta[0]+delta[2]*delta[2] >= 400
		}
		if snap {
			r.snapped = true
			r.pos, r.from, r.to = target, target, target
			r.duration = 0
		} else {
			r.from, r.to = r.pos, target
			r.start = now
			r.duration = min(max(now-r.lastArrival, .1), .5)
			if e.Kind == game.KindMob && e.MobHeightKnown && e.MobMoveDurationMS > 0 {
				r.duration = float64(e.MobMoveDurationMS) / 1000
			}
		}
		r.moveSerial = e.MoveSerial
		r.lastArrival = now
	}
	return r.pos
}

// remoteHeading smooths horizontal actor yaw across 10 Hz movement updates
// (matching native CCharacter::Update 0x41a59b..0x41a6a4) so camera-steered
// turns rotate smoothly along the shortest arc instead of snapping per packet.
type remoteHeading struct {
	initialized                  bool
	yaw, from, delta, target     float64
	moveSerial, resetSerial      uint64
	start, duration, lastArrival float64
}

func entityForward(dir uint16) render.V3 {
	s, c := math.Sincos(float64(dir) * (2 * math.Pi / 65536))
	return render.V3{float32(s), 0, float32(c)}
}

func yawVector(yaw float64) render.V3 {
	s, c := math.Sincos(yaw)
	return render.V3{float32(s), 0, float32(c)}
}

func vectorYaw(dir render.V3, fallback float64) (float64, bool) {
	if dir[0]*dir[0]+dir[2]*dir[2] <= 1e-12 {
		return fallback, false
	}
	return math.Atan2(float64(dir[0]), float64(dir[2])), true
}

func shortestAngleDelta(from, to float64) float64 {
	return math.Remainder(to-from, 2*math.Pi)
}

func (h *remoteHeading) update(wantDir render.V3, e *game.Entity, snap bool, now float64) render.V3 {
	want, ok := vectorYaw(wantDir, h.yaw)
	if !ok {
		want, _ = vectorYaw(entityForward(e.Dir), h.yaw)
	}
	if !h.initialized || snap || e.PositionResetSerial != h.resetSerial {
		h.initialized = true
		h.yaw, h.from, h.target = want, want, want
		h.delta, h.duration = 0, 0
		h.moveSerial, h.resetSerial = e.MoveSerial, e.PositionResetSerial
		h.lastArrival = now
		return yawVector(h.yaw)
	}
	if h.duration > 0 {
		t := min(max((now-h.start)/h.duration, 0), 1)
		h.yaw = math.Remainder(h.from+h.delta*t, 2*math.Pi)
		if t >= 1 {
			h.yaw = h.target
			h.duration = 0
		}
	}
	newPacket := e.MoveSerial != h.moveSerial
	retarget := math.Abs(shortestAngleDelta(h.target, want)) > 1e-4
	if newPacket || retarget {
		dur := 0.1
		if newPacket {
			if gap := now - h.lastArrival; gap > 0.1 && gap <= 0.25 {
				dur = gap
			}
			h.moveSerial = e.MoveSerial
			h.lastArrival = now
		}
		delta := shortestAngleDelta(h.yaw, want)
		if math.Abs(delta) <= 1e-4 {
			h.yaw, h.from, h.target = want, want, want
			h.delta, h.duration = 0, 0
		} else {
			h.from, h.delta, h.target = h.yaw, delta, want
			h.start, h.duration = now, dur
		}
	}
	return yawVector(h.yaw)
}

func (h *remoteHeading) snap(dir render.V3, e *game.Entity, now float64) render.V3 {
	yaw, ok := vectorYaw(dir, h.yaw)
	if !ok && !h.initialized {
		yaw, _ = vectorYaw(entityForward(e.Dir), 0)
	}
	h.initialized = true
	h.yaw, h.from, h.target = yaw, yaw, yaw
	h.delta, h.duration = 0, 0
	h.moveSerial, h.resetSerial = e.MoveSerial, e.PositionResetSerial
	h.lastArrival = now
	return yawVector(h.yaw)
}

func (h *remoteHeading) vector(e *game.Entity) render.V3 {
	if h.initialized {
		return yawVector(h.yaw)
	}
	return entityForward(e.Dir)
}

// Patrol packets are sparse steps, not player position samples. MoveSpeed is
// milliseconds per MoveDistance in NpcQuest, matching the server patrol clock.
func (r *remotePosition) updateNPC(e *game.Entity, now float64, def sdata.NPC) render.V3 {
	if !r.initialized || e.PositionResetSerial != r.resetSerial {
		return r.update(e, now)
	}
	if r.duration > 0 {
		t := float32(min(max((now-r.start)/r.duration, 0), 1))
		r.pos = r.from.Add(r.to.Sub(r.from).Scale(t))
	}
	if e.MoveSerial != r.moveSerial {
		target := render.V3{e.X, e.Y, e.Z}
		delta := target.Sub(r.pos)
		delta[1] = 0
		speed := float32(2)
		if def.MoveDistance > 0 && def.MoveSpeed > 0 {
			speed = float32(def.MoveDistance) * 1000 / float32(def.MoveSpeed)
		}
		if e.Motion == 1 {
			speed *= 4
			if def.MoveDistance <= 0 || def.MoveSpeed <= 0 {
				speed = 8
			}
		}
		r.from, r.to, r.start = r.pos, target, now
		r.duration = float64(delta.Len() / max(speed, .1))
		if r.duration == 0 {
			r.pos, r.from = target, target
		}
		r.moveSerial, r.lastArrival = e.MoveSerial, now
	}
	return r.pos
}

// visualEntityPosition keeps picking and effects attached to the rendered actor.
func (a *App) visualEntityPosition(e *game.Entity) render.V3 {
	if a.wv != nil {
		if e.Kind == game.KindUser {
			if actor := a.wv.actors[e.ID]; actor != nil && actor.entity == e && actor.position.initialized {
				return actor.position.pos
			}
		} else if e.Kind == game.KindMob {
			if actor := a.wv.mobs[e.ID]; actor != nil && actor.entity == e && actor.position.initialized {
				return actor.pos
			}
		} else if e.Kind == game.KindNpc {
			if actor := a.wv.npcs[e.ID]; actor != nil && actor.entity == e && actor.position.initialized {
				return actor.position.pos
			}
		}
	}
	return render.V3{e.X, e.Y, e.Z}
}
