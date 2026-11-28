package app

import (
	"math"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
)

func TestRemotePositionInterpolatesAndRetargets(t *testing.T) {
	e := &game.Entity{X: 10, Y: 3, Z: 20}
	var p remotePosition
	if got := p.update(e, 0); got != (render.V3{10, 3, 20}) {
		t.Fatal(got)
	}
	e.X, e.MoveSerial = 14, 1
	if got := p.update(e, 1); got[0] != 10 {
		t.Fatal(got)
	}
	if got := p.update(e, 1.25); got[0] != 12 {
		t.Fatal(got)
	}
	e.X, e.MoveSerial = 16, 2
	if got := p.update(e, 1.25); got[0] != 12 {
		t.Fatal(got)
	}
	if got := p.update(e, 1.375); got[0] != 14 {
		t.Fatal(got)
	}
	if got := p.update(e, 1.5); got[0] != 16 {
		t.Fatal(got)
	}
	if e.X != 16 {
		t.Fatal("presentation changed server state")
	}
}

func TestMonsterMovementUsesConfirmedPacketsForVisualInterpolation(t *testing.T) {
	s := game.New()
	spawn := proto.NewWriter(proto.MobEnter)
	spawn.U32(42)
	spawn.Bool(false)
	spawn.U16(1)
	spawn.F32(10)
	spawn.F32(20)
	if _, err := s.Apply(spawn.Bytes()); err != nil {
		t.Fatal(err)
	}
	e := s.Entities[42]
	var p remotePosition
	p.update(e, 0)
	move := proto.NewWriter(proto.MobMove)
	move.U32(42)
	move.U8(1)
	move.F32(14)
	move.F32(20)
	if _, err := s.Apply(move.Bytes()); err != nil {
		t.Fatal(err)
	}
	if e.MoveSerial != 1 || e.X != 14 {
		t.Fatal("server movement not applied")
	}
	if got := p.update(e, 1); got[0] != 10 {
		t.Fatal(got)
	}
	if got := p.update(e, 1.25); got[0] != 12 {
		t.Fatal(got)
	}
	if e.X != 14 {
		t.Fatal("visual movement changed authoritative position")
	}
	if _, err := s.Apply(move.Bytes()[:5]); err == nil || e.MoveSerial != 1 {
		t.Fatal("truncated movement changed state")
	}
}

func TestRemotePositionSnapsOnResetOrLargeJump(t *testing.T) {
	e := &game.Entity{}
	var p remotePosition
	p.update(e, 0)
	e.X, e.MoveSerial = 21, 1
	if got := p.update(e, 1); got[0] != 21 {
		t.Fatal(got)
	}
	e.Y, e.MoveSerial = 9, 2
	if got := p.update(e, 2); got[1] != 9 {
		t.Fatal(got)
	}
	e.X, e.PositionResetSerial = 22, 1
	if got := p.update(e, 3); got[0] != 22 {
		t.Fatal(got)
	}
}

func TestMobPositionUsesServerDurationThroughCorners(t *testing.T) {
	e := &game.Entity{Kind: game.KindMob, MobHeightKnown: true}
	var p remotePosition
	p.update(e, 0)
	e.X, e.MoveSerial, e.MobMoveDurationMS = 2, 1, 200
	p.update(e, 1)
	if got := p.update(e, 1.1); got != (render.V3{1, 0, 0}) {
		t.Fatalf("server timing ignored: %v", got)
	}
	p.update(e, 1.2)
	e.Z, e.MoveSerial, e.MobMoveDurationMS = 1, 2, 100
	p.update(e, 1.2)
	if got := p.update(e, 1.25); got != (render.V3{2, 0, .5}) {
		t.Fatalf("corner interpolation cut or paused: %v", got)
	}
	if e.X != 2 || e.Z != 1 {
		t.Fatal("presentation changed authoritative position")
	}
}

func TestMobChaseAnimationSurvivesShortGapsBetweenMovementPackets(t *testing.T) {
	a := inputApp()
	e := &game.Entity{Kind: game.KindMob, MobHeightKnown: true}
	mob := a.trackMob(e)
	check := func(now float64, want int) {
		t.Helper()
		a.clock = now
		a.updateMobPosition(mob)
		if got := mob.locomotionAction(now); got != want {
			t.Fatalf("at %.3fs: action = %d, want %d", now, got, want)
		}
	}
	check(0, 6)
	e.Motion, e.MobMoveDurationMS = 1, 25
	for i := range 8 {
		now := 1 + float64(i)*.1
		e.X += .15
		e.MoveSerial++
		check(now, 1)
		check(now+.02, 1)
		check(now+.09, 1)
	}
	check(1.9, 6)

	e.Motion, e.MobMoveDurationMS = 0, 80
	e.X, e.MoveSerial = 2, e.MoveSerial+1
	check(2, 0)
	check(2.081, 6)

	e.Motion, e.MobMoveDurationMS = 1, 30
	e.X, e.MoveSerial = 3, e.MoveSerial+1
	check(3, 1)
	e.MobMoveDurationMS = 0
	e.MoveSerial++
	e.PositionResetSerial++
	check(3.01, 6)
}

func TestRemoteHeadingInterpolatesCirclesAndShortestArc(t *testing.T) {
	dirU16 := func(deg float64) uint16 {
		return uint16(math.Round(math.Mod(math.Mod(deg, 360)+360, 360) / 360 * 65536))
	}
	yawDeg := func(v render.V3) float64 {
		return math.Atan2(float64(v[0]), float64(v[2])) * 180 / math.Pi
	}
	assertAngle := func(got render.V3, wantDeg float64) {
		t.Helper()
		if diff := math.Abs(math.Remainder(yawDeg(got)-wantDeg, 360)); diff > 0.25 {
			t.Fatalf("yaw = %.2f deg, want %.2f deg (diff %.2f)", yawDeg(got), wantDeg, diff)
		}
	}

	e := &game.Entity{Dir: dirU16(90)}
	var h remoteHeading
	assertAngle(h.update(entityForward(e.Dir), e, true, 0), 90)

	// Continuous 10 Hz circle steps (45 deg every 100 ms, crossing +-180 deg).
	steps := []float64{135, 180, -135, -90}
	prev := 90.0
	for i, target := range steps {
		now := 1.0 + float64(i)*0.1
		origDir := dirU16(target)
		e.Dir = origDir
		e.MoveSerial++
		assertAngle(h.update(entityForward(e.Dir), e, false, now), prev)
		mid := prev + math.Remainder(target-prev, 360)*0.5
		assertAngle(h.update(entityForward(e.Dir), e, false, now+0.05), mid)
		if e.Dir != origDir {
			t.Fatal("heading presentation mutated authoritative entity direction")
		}
		prev = target
	}
	assertAngle(h.update(entityForward(e.Dir), e, false, 1.4), -90)

	// Crossing the 350 deg -> 10 deg uint16 wraparound must take the +20 deg arc through 0 deg.
	e.Dir, e.MoveSerial = dirU16(350), e.MoveSerial+1
	h.snap(entityForward(e.Dir), e, 2.0)
	e.Dir, e.MoveSerial = dirU16(10), e.MoveSerial+1
	assertAngle(h.update(entityForward(e.Dir), e, false, 2.1), -10)
	mid := h.update(entityForward(e.Dir), e, false, 2.15)
	if mid[2] < 0.99 || math.Abs(float64(mid[0])) > 0.02 {
		t.Fatalf("expected shortest-arc midpoint near +Z, got %v", mid)
	}
	assertAngle(h.update(entityForward(e.Dir), e, false, 2.2), 10)
}

func TestRemoteHeadingSnapsOnResetTeleportAndCombatOverride(t *testing.T) {
	e := &game.Entity{}
	var p remotePosition
	var h remoteHeading
	p.update(e, 0)
	h.update(entityForward(e.Dir), e, p.snapped, 0)

	// Large teleport jump (> 20 units) snaps both position and heading immediately.
	e.X, e.Dir, e.MoveSerial = 25, 16384, 1
	p.update(e, 0.1)
	if !p.snapped {
		t.Fatal("expected position snap on large teleport")
	}
	if got := h.update(entityForward(e.Dir), e, p.snapped, 0.1); got[0] < 0.99 {
		t.Fatalf("expected immediate +X heading on teleport, got %v", got)
	}

	// Normal step keeps p.snapped false so heading interpolates instead of snapping.
	e.X, e.Dir, e.MoveSerial = 26, 32768, 2
	p.update(e, 0.2)
	if p.snapped {
		t.Fatal("unexpected snap on normal movement step")
	}
	if got := h.update(entityForward(e.Dir), e, p.snapped, 0.2); got[0] < 0.99 {
		t.Fatalf("heading snapped on normal step: %v", got)
	}

	// Explicit PositionResetSerial snaps immediately.
	e.Dir, e.PositionResetSerial = 49152, 1
	p.update(e, 0.25)
	if got := h.update(entityForward(e.Dir), e, p.snapped, 0.25); got[0] > -0.99 {
		t.Fatalf("expected immediate -X heading on PositionResetSerial, got %v", got)
	}

	// Combat/dash override snaps immediately and blends smoothly back when released.
	if got := h.snap(render.V3{0, 0, 1}, e, 0.3); got[2] < 0.99 {
		t.Fatalf("expected immediate +Z combat snap, got %v", got)
	}
	if got := h.update(entityForward(e.Dir), e, false, 0.4); got[2] < 0.99 {
		t.Fatalf("expected smooth release from combat facing, got %v", got)
	}
	mid := h.update(entityForward(e.Dir), e, false, 0.45)
	if mid[0] > -0.70 || mid[2] < 0.70 {
		t.Fatalf("expected midpoint between +Z and -X, got %v", mid)
	}
	if got := h.update(entityForward(e.Dir), e, false, 0.5); got[0] > -0.99 {
		t.Fatalf("expected settled -X heading after release, got %v", got)
	}
}
