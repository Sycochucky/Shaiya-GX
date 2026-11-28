package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/mapgeometry"
	"math"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/movespeed"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

// Native 0x43bfa0: base 5 (0x409250), offsets -2/-1/0/+1/+2,
// and 255 blocks movement. Unknown indices retain the native base speed.
func movementSpeed(s *proto.CharacterSpeed) float32 {
	if s == nil {
		return runSpeed
	}
	return movespeed.Units(s.Move)
}

// Native movement replaces the normal speed with the mount's speed. The base
// tier is known to be 7; some special variants use 8 but their tier is not
// carried by Rapture's mount-shape packet.
func playerMovementSpeed(s *proto.CharacterSpeed, mounted bool) float32 {
	if s != nil && s.Move == 255 {
		return 0
	}
	if mounted {
		return 7
	}
	return movementSpeed(s)
}

// place keeps the complete camera offset, including elevation and user zoom.
func (v *worldView) place(pos render.V3) {
	low, high := v.world.PlayerHeightLimits()
	if math.IsNaN(float64(pos[1])) {
		pos[1] = v.pos[1]
	}
	pos[1] = min(max(pos[1], low), high)
	if v.pos[0] != pos[0] || v.pos[2] != pos[2] {
		v.movementRemainder = [2]float64{}
	}
	v.eye = v.eye.Add(pos.Sub(v.pos))
	v.pos = pos
}

const visualCorrectionDuration float32 = .18

func (v *worldView) visualPosition() render.V3 {
	if v.visualCorrectionLeft == 0 {
		return v.pos
	}
	return v.world.GroundPlayer(v.pos, v.visualCorrectionOffset, false)
}

func (v *worldView) clearVisualCorrection() {
	v.visualCorrectionOffset, v.visualCorrectionLeft = render.V3{}, 0
}

// Apply the server position immediately for simulation. Only the visible
// model can converge from its previous position after a small ground update.
func (v *worldView) reconcilePosition(pos render.V3, smooth bool) {
	offset := v.visualPosition().Sub(pos)
	v.place(pos)
	v.clearVisualCorrection()
	if !smooth || math.Abs(float64(offset[1])) > .05 {
		return
	}
	offset[1] = 0
	if length := offset.Len(); length < .001 || length > .45 {
		return
	}
	// A correction at a wall or step must display the legal position.
	if v.world.GroundPlayer(pos, offset, false).Sub(pos.Add(offset)).Len() > .025 {
		return
	}
	v.visualCorrectionOffset, v.visualCorrectionLeft = offset, visualCorrectionDuration
}

func (v *worldView) advanceVisualCorrection(dt float32) {
	if v.visualCorrectionLeft <= dt {
		v.clearVisualCorrection()
		return
	}
	remaining := v.visualCorrectionLeft - dt
	v.visualCorrectionOffset = v.visualCorrectionOffset.Scale(remaining / v.visualCorrectionLeft)
	v.visualCorrectionLeft = remaining
}

func (v *worldView) movePacket(motion uint8) []byte {
	direction := uint16(math.Atan2(float64(v.facing[0]), float64(v.facing[2])) / (2 * math.Pi) * 65536)
	return proto.MovePacket(motion, direction, v.pos[0], v.pos[1], v.pos[2])
}

func (v *worldView) step(delta render.V3) bool {
	pos := v.pos.Add(delta)
	// Keep the fraction lost when each frame is rounded to the world position.
	// At high frame rates, discarding that fraction every frame can make a
	// sustained run cover more distance than the server's elapsed-time budget.
	x := float64(v.pos[0]) + float64(delta[0]) + v.movementRemainder[0]
	z := float64(v.pos[2]) + float64(delta[2]) + v.movementRemainder[1]
	targetX, targetZ := float32(x), float32(z)
	pos[0], pos[2] = targetX, targetZ
	delta[0], delta[2] = targetX-v.pos[0], targetZ-v.pos[2]
	groundSolved := false
	if v.world != nil && v.world.HasTerrain() {
		size := float32(v.world.WLD.MapSize)
		pos[0], pos[2] = min(max(pos[0], 0), size), min(max(pos[2], 0), size)
		if floor, deep := v.world.WaterFloor(pos[0], pos[2]); deep {
			if v.world.SwimmingAt(v.pos[0], v.pos[1], v.pos[2]) {
				pos[1] = min(max(pos[1], floor), scene.SwimSurfaceY)
			} else {
				// Falling into deep water takes time; a ledge must not teleport
				// the player or camera directly to the swimming level.
				pos[1] = max(v.pos[1], scene.SwimEntryY)
			}
		} else if v.world.HasPlayerCollision() {
			pos = v.world.GroundPlayerDT(v.pos, delta, v.stepDT, v.jumping)
			groundSolved = true
		} else if !v.jumping {
			pos[1] = v.world.Height(pos[0], pos[2])
		}
	}
	if v.world != nil {
		if !v.world.HasTerrain() && v.world.HasPlayerCollision() {
			pos = v.world.GroundPlayerDT(v.pos, delta, v.stepDT, v.jumping)
			groundSolved = true
		}
		if !groundSolved {
			pos, _ = v.world.MovePlayer(v.pos, pos.Sub(v.pos))
		}
		if v.world.HasTerrain() {
			size := float32(v.world.WLD.MapSize)
			pos[0], pos[2] = min(max(pos[0], 0), size), min(max(pos[2], 0), size)
		}
	}
	moved := pos != v.pos
	v.place(pos)
	v.movementRemainder = [2]float64{}
	if pos[0] == targetX {
		v.movementRemainder[0] = x - float64(targetX)
	}
	if pos[2] == targetZ {
		v.movementRemainder[1] = z - float64(targetZ)
	}
	return moved
}

// Apply gravity after walking over deep water. The entry threshold stops the
// fall and lets buoyancy take over; all vertical movement is sent to Rapture.
func (v *worldView) settleIntoWater(dt float32) bool {
	if v.world == nil {
		return false
	}
	_, deep := v.world.WaterFloor(v.pos[0], v.pos[2])
	if !deep || v.pos[1] <= scene.SwimSurfaceY {
		v.waterFallSpeed = 0
		return false
	}
	v.waterFallSpeed = min(v.waterFallSpeed+9.8*dt, 30)
	pos := v.pos
	pos[1] = max(scene.SwimEntryY, pos[1]-v.waterFallSpeed*dt)
	if pos[1] <= scene.SwimSurfaceY {
		pos[1] = scene.SwimEntryY
		v.waterFallSpeed = 0
	}
	if pos == v.pos {
		return false
	}
	pos, hit := v.world.MovePlayer(v.pos, pos.Sub(v.pos))
	if hit {
		v.waterFallSpeed = 0
	}
	changed := pos != v.pos
	if changed {
		v.place(pos)
	}
	return changed
}

// Phoenix rises gently near the surface; a deeply submerged position stays put.
func (v *worldView) floatAtSurface(dt float32) bool {
	if v.world == nil || !v.world.SwimmingAt(v.pos[0], v.pos[1], v.pos[2]) || v.pos[1] < scene.SwimEntryY || v.pos[1] >= scene.SwimSurfaceY {
		return false
	}
	pos := v.pos
	pos[1] = min(scene.SwimSurfaceY, pos[1]+1.05*dt)
	if pos == v.pos {
		return false
	}
	pos, hit := v.world.MovePlayer(v.pos, pos.Sub(v.pos))
	if hit {
		v.waterFallSpeed = 0
	}
	changed := pos != v.pos
	if changed {
		v.place(pos)
	}
	return changed
}

// Swimming forward follows camera pitch; side steps remain horizontal.
func swimDirection(horizontal render.V3, view render.V3, forward, backward bool) render.V3 {
	y := view.Norm()[1]
	if y >= -0.34 && y <= 0.02 {
		y = 0
	}
	if forward {
		horizontal[1] += y
	}
	if backward {
		horizontal[1] -= y
	}
	return horizontal.Norm()
}

func (v *worldView) riseInWater(dt float32) bool {
	if v.world == nil || !v.world.SwimmingAt(v.pos[0], v.pos[1], v.pos[2]) {
		return false
	}
	pos := v.pos
	pos[1] = min(scene.SwimSurfaceY, pos[1]+1.8*dt)
	if pos == v.pos {
		return false
	}
	pos, hit := v.world.MovePlayer(v.pos, pos.Sub(v.pos))
	if hit {
		v.waterFallSpeed = 0
	}
	changed := pos != v.pos
	if changed {
		v.place(pos)
	}
	return changed
}

// Only packets about this character may reconcile local prediction. A map
// change will build a new view from State instead of moving the old camera.
func (a *App) confirmedPosition(p []byte) {
	if proto.Op(p) == proto.GXPhysicsClock {
		a.physicsClockAt = time.Now()
		if a.wv != nil && a.wv.world != nil && a.state.PhysicsMapID == a.wv.mapID {
			a.wv.world.SyncPhysicsClock(a.state.PhysicsTime)
		}
		return
	}
	v := a.wv
	if v == nil || v.mapID != a.state.MapID {
		return
	}
	switch proto.Op(p) {
	case proto.GXSkillFailure:
		a.rejectInstantPrediction(a.state.LastSkillFailure.Skill)
	case proto.GXLadder:
		result, err := proto.DecodeLadderState(p)
		if err == nil && result.ID == a.state.SelfID && result.Mode == mapgeometry.LadderDetach {
			v.ladder.Active = false
			v.ladderFinishing = false
			v.ladderRetryAt = a.clock + .35
		}
	case proto.GXSkillDash:
		if a.state.LastSkillDash.Source == a.state.SelfID {
			v.clearVisualCorrection()
		}
		a.receiveSkillDash(a.state.LastSkillDash)
	case proto.GXPlayerCorrection:
		if proto.NewReader(p).U32() != a.state.SelfID {
			return
		}
		_, dashing := v.dashes[a.state.SelfID]
		if dashing && v.self != nil && v.self.PlayingAction(101) {
			v.self.ReleaseAction()
		}
		delete(v.dashes, a.state.SelfID)
		correction := a.state.LastPlayerCorrection
		if correction.ID != a.state.SelfID {
			return
		}
		smooth := v.moving && !v.jumping && !correction.Jumping && !dashing && !v.ladder.Active && !v.evading &&
			!a.state.Self.Dead && !a.state.Self.Sitting && a.clock >= v.attackMoveUntil &&
			(correction.Motion&0x3f != 0 || v.walking)
		v.reconcilePosition(render.V3{correction.X, correction.Y, correction.Z}, smooth)
		v.jumping, v.jumpFall, v.groundFall = correction.Jumping, correction.JumpFall, correction.Fall
		v.slope.Frames = correction.SlopeFrames
		v.motion = correction.Motion
		v.waterFallSpeed = 0
		v.sendT = .1
	case proto.CharacterMove, proto.CharacterMapTeleport, proto.CharacterResurrect:
		if proto.NewReader(p).U32() != a.state.SelfID {
			return
		}
		if proto.Op(p) != proto.CharacterMove {
			v.sitPending, v.sitAttack = false, 0
			if v.self != nil && (v.self.PlayingAction(10) || v.self.PlayingAction(11)) {
				v.self.ReleaseAction()
			}
			delete(v.dashes, a.state.SelfID)
			v.ladder.Active = false
			v.ladderFinishing = false
			v.ladderRetryAt = a.clock + .35
		}
		v.slope.Frames = 0
		v.clearVisualCorrection()
		v.place(render.V3{a.state.Self.X, a.state.Self.Y, a.state.Self.Z})
		v.combatHeading = render.V3{}
		v.approachTarget = 0
		v.snapRenderedCamera()
		v.moving, v.turning, v.motion, v.sendT = false, false, 0, 0
		v.waterDirty = false
		v.waterFallSpeed = 0
		v.jumping, v.jumpPending = false, false
		v.jumpActive, v.jumpMoved, v.landingAction = false, false, landingAction{}
		v.groundFall = 0
		v.evading = false
		v.travel = false
		if v.self != nil && v.self.PlayingAction(8) {
			v.self.ReleaseAction()
		}
		if proto.Op(p) != proto.CharacterMove {
			v.autoRun = false
		}
		if proto.Op(p) == proto.CharacterMove {
			v.motion = a.state.Self.Motion
			v.moving = v.motion&0x3f != 0
			angle := float64(a.state.Self.Dir) / 65536 * 2 * math.Pi
			v.facing = render.V3{float32(math.Sin(angle)), 0, float32(math.Cos(angle))}
		}
	}
}

func (v *worldView) turnFacing(yaw float64) {
	c, s := float32(math.Cos(yaw)), float32(math.Sin(yaw))
	f := v.facing
	v.facing = render.V3{f[0]*c - f[2]*s, 0, f[0]*s + f[2]*c}
}

// EP5.4 0x43e8ad and 0x43f2aa share the side and backward step tiers.
// A mount adds 2 to the base backward step at 0x43f341.
func directionalStepSpeed(s *proto.CharacterSpeed, mounted bool) float32 {
	tier := movespeed.Normal
	if s != nil {
		tier = s.Move
	}
	if tier == movespeed.Blocked {
		return 0
	}
	var speed float32
	if tier <= movespeed.Maximum {
		speed = .5 + float32(tier)*.5
	}
	if mounted {
		speed += 2
	}
	return speed
}

// Native gravity also runs after walking off model surfaces, not only jumping.
func (v *worldView) settleOnGround(dt float32) bool {
	if v.world == nil || !v.world.HasPlayerCollision() || dt <= 0 || v.world.SwimmingAt(v.pos[0], v.pos[1], v.pos[2]) {
		return false
	}
	if _, deep := v.world.WaterFloor(v.pos[0], v.pos[2]); deep {
		return false
	}
	origin := v.pos
	v.place(v.world.SlidePlayer(v.pos, dt, &v.slope))
	if v.groundFall == 0 {
		v.groundFall = -2.98
	}
	v.groundFall = max(-20, v.groundFall*(1+2.8*dt))
	desired := v.pos
	desired[1] += v.groundFall * dt
	if v.world.HasTerrain() {
		desired[1] = max(desired[1], v.world.Height(desired[0], desired[2]))
	}
	pos, hit := v.world.MovePlayer(v.pos, desired.Sub(v.pos))
	if hit || v.world.HasTerrain() && pos[1] <= v.world.Height(pos[0], pos[2])+.001 {
		v.groundFall = -2.98
	}
	pos = v.world.StandingPosition(pos)
	changed := pos.Sub(origin).Len() > 1e-5
	if changed {
		v.place(pos)
	}
	return changed
}
