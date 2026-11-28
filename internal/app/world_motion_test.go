package app

import (
	"bufio"
	"math"
	"net"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/session"
)

func TestCameraKeepsOrbitAndZoomAcrossElevationChanges(t *testing.T) {
	v := &worldView{pos: render.V3{500, 10, 1700}, eye: render.V3{497, 14, 1695}}
	v.rotateEye(.6, .2)
	offset := v.eye.Sub(v.pos)
	for _, pos := range []render.V3{{501, 12, 1701}, {503, 35, 1705}, {504, 2, 1704}, {800, 78, 1600}} {
		v.place(pos)
		if got := v.eye.Sub(v.pos).Sub(offset).Len(); got > .001 {
			t.Fatalf("elevation change lost the camera orbit/zoom: %v", v.eye.Sub(v.pos))
		}
	}
}

func TestMountedMovementSpeed(t *testing.T) {
	if got := playerMovementSpeed(&proto.CharacterSpeed{Move: 0}, true); got != 7 {
		t.Fatalf("mounted speed = %v, want 7", got)
	}
	if got := playerMovementSpeed(&proto.CharacterSpeed{Move: 0}, false); got != 3 {
		t.Fatalf("unmounted speed = %v, want 3", got)
	}
}

func TestSwimmingDirectionFollowsPitchOnlyForForwardAndBack(t *testing.T) {
	horizontal := render.V3{0, 0, 1}
	down := render.V3{0, -1, 1}
	if d := swimDirection(horizontal, down, true, false); d[1] >= 0 {
		t.Fatalf("forward should dive: %v", d)
	}
	if d := swimDirection(horizontal, down, false, true); d[1] <= 0 {
		t.Fatalf("backward should rise: %v", d)
	}
	if d := swimDirection(horizontal, down, false, false); d[1] != 0 {
		t.Fatalf("strafe should stay level: %v", d)
	}
	if d := swimDirection(horizontal, render.V3{0, -.1, 1}, true, false); d[1] != 0 {
		t.Fatalf("small downward pitch should stay level: %v", d)
	}
}

func TestServerPositionReconcilesBeforeNextPredictionFrame(t *testing.T) {
	a := inputApp()
	a.state.MapID, a.state.SelfID = 1, 10
	a.wv = &worldView{mapID: 1, pos: render.V3{10, 4, 20}, eye: render.V3{10, 6, 15}, moving: true}
	packets := make(chan []byte, 1)
	a.world = &session.World{Packets: packets}
	w := proto.NewWriter(proto.CharacterMove)
	w.U32(10)
	w.U8(0)
	w.U16(16384)
	w.F32(30)
	w.F32(50)
	w.F32(60)
	packets <- w.Bytes()
	a.pumpWorld()
	a.updateWorld(&platform.Input{}, .016)
	if a.wv.pos != (render.V3{30, 50, 60}) || a.state.Self.X != 30 || a.wv.moving {
		t.Fatal("prediction overwrote the server correction", a.wv.pos, a.state.Self)
	}
	if a.wv.eye != (render.V3{30, 52, 55}) || math.Abs(float64(a.wv.facing[0]-1)) > .001 {
		t.Fatal("correction lost camera offset or facing")
	}
}

func TestSmallGroundCorrectionChangesSimulationImmediatelyAndBlendsModel(t *testing.T) {
	a := inputApp()
	a.state.MapID, a.state.SelfID = 1, 10
	old := render.V3{10, 4, 20}
	corrected := render.V3{9.7, 4, 20}
	a.wv = &worldView{mapID: 1, pos: old, eye: render.V3{10, 6, 15}, moving: true}
	w := proto.NewWriter(proto.GXPlayerCorrection)
	w.U32(10)
	w.F32(corrected[0])
	w.F32(corrected[1])
	w.F32(corrected[2])
	w.U8(1)
	w.U16(0)
	w.U8(0)
	w.F32(0)
	w.F32(0)
	w.F32(0)
	a.applyWorldPacket(w.Bytes())
	v := a.wv
	if v.pos != corrected || a.state.Self.X != corrected[0] || v.visualPosition() != old {
		t.Fatalf("correction misplaced simulation or model: physical %v, visual %v", v.pos, v.visualPosition())
	}
	r := proto.NewReader(v.movePacket(1))
	r.U8()
	r.U16()
	if x := r.F32(); x != corrected[0] {
		t.Fatalf("movement packet used visual position: x = %g", x)
	}
	v.advanceVisualCorrection(visualCorrectionDuration / 2)
	if got := v.visualPosition()[0]; math.Abs(float64(got-9.85)) > .0001 {
		t.Fatalf("model did not converge halfway: x = %g", got)
	}
	v.advanceVisualCorrection(visualCorrectionDuration / 2)
	if v.visualPosition() != corrected {
		t.Fatal("model did not settle at the authoritative position")
	}
	v.reconcilePosition(render.V3{4, 4, 20}, true)
	if v.visualPosition() != v.pos {
		t.Fatal("large correction was blended")
	}
}

func TestOtherCharacterTeleportDoesNotStopOurMovement(t *testing.T) {
	a := inputApp()
	a.state.MapID, a.state.SelfID = 1, 10
	a.wv = &worldView{mapID: 1, pos: render.V3{10, 4, 20}, eye: render.V3{10, 6, 15}, moving: true}
	before := *a.wv
	w := proto.NewWriter(proto.CharacterMapTeleport)
	w.U32(20)
	w.U16(2)
	w.F32(300)
	w.F32(50)
	w.F32(600)
	if _, err := a.state.Apply(w.Bytes()); err != nil {
		t.Fatal(err)
	}
	a.confirmedPosition(w.Bytes())
	if a.wv.pos != before.pos || a.wv.eye != before.eye || !a.wv.moving {
		t.Fatal("another character's teleport interrupted the local view")
	}
}

func TestRunningKeepsSubFramePositionPrecision(t *testing.T) {
	v := worldView{pos: render.V3{1051, 2, 410}}
	const fps, seconds = 240, 80
	for range fps * seconds {
		v.step(render.V3{5.0 / fps, 0, 0})
	}
	if got := v.pos[0]; got != 1051+5*seconds {
		t.Fatalf("continuous 5 u/s run ended at %v, want %v", got, 1051+5*seconds)
	}
	v.step(render.V3{5.0 / fps, 0, 0})
	if v.movementRemainder[0] == 0 {
		t.Fatal("sample did not exercise a fractional world-coordinate step")
	}
	v.place(render.V3{300, 2, 400})
	if v.movementRemainder != [2]float64{} {
		t.Fatal("server relocation kept a fraction from the previous movement")
	}
}

func TestNativeMovementSpeedIncludesBlockedAndUnknownIndices(t *testing.T) {
	for index, want := range map[uint8]float32{0: 3, 1: 4, 2: 5, 3: 6, 4: 7, 255: 0, 17: 5} {
		if got := movementSpeed(&proto.CharacterSpeed{Move: index}); got != want {
			t.Errorf("speed %d: got %f want %f", index, got, want)
		}
	}
	if movementSpeed(nil) != 5 {
		t.Fatal("missing speed should retain the native initial rate")
	}
}

func TestImmobilizedInputCannotMoveOrInterruptTheCamera(t *testing.T) {
	a := inputApp()
	a.state.Self.Speed = &proto.CharacterSpeed{Move: 255}
	a.wv = &worldView{pos: render.V3{10, 4, 20}, eye: render.V3{10, 6, 15}}
	a.updateWorld(&platform.Input{Down: "W"}, .5)
	if a.wv.pos != (render.V3{10, 4, 20}) || a.wv.eye != (render.V3{10, 6, 15}) || a.wv.moving {
		t.Fatal("blocked movement predicted a step")
	}
	a.updateWorld(&platform.Input{Down: ">"}, .1)
	if a.wv.pos != (render.V3{10, 4, 20}) || a.wv.eye == (render.V3{10, 6, 15}) {
		t.Fatal("immobilization also blocked camera orbit")
	}
}

func TestRemoteWalkInferenceExpiresAndDistinguishesRunStop(t *testing.T) {
	r := &remoteActor{moveX: 10, moveZ: 20}
	e := &game.Entity{X: 10, Z: 20}
	if w, m := r.observeMovement(e, 1); w || m {
		t.Fatal("stationary actor moves")
	}
	e.MoveSerial++
	e.X = 10.5
	if w, m := r.observeMovement(e, 2); !w || !m {
		t.Fatal("position change in walk mode not recognized")
	}
	if w, m := r.observeMovement(e, 2.8); w || m {
		t.Fatal("walk animation did not expire")
	}
	e.MoveSerial++
	e.Motion = 1
	e.X = 11
	if w, m := r.observeMovement(e, 3); w || !m {
		t.Fatal("run motion lost")
	}
	e.MoveSerial++
	e.Motion = 0
	e.X = 11.2
	if w, m := r.observeMovement(e, 3.1); w || m {
		t.Fatal("run stop misread as walk")
	}
	e.MoveSerial++
	e.X = 11.5
	if w, m := r.observeMovement(e, 3.6); !w || !m {
		t.Fatal("walk after run not recognized")
	}
}

func TestStationaryTurnSendsPeriodicAndFinalMovePackets(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	sent := make(chan []byte, 32)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		for {
			p, err := netx.ReadFrame(r)
			if err != nil {
				return
			}
			sent <- p
		}
	}()

	c, err := netx.Dial(ln.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	w := session.NewWorld(c)
	t.Cleanup(func() { w.Close() })

	recvMove := func() (motion uint8, dir uint16, pos render.V3) {
		t.Helper()
		select {
		case p := <-sent:
			if proto.Op(p) != proto.CharacterMove {
				t.Fatalf("opcode = 0x%04X, want 0x0501", proto.Op(p))
			}
			rd := proto.NewReader(p)
			motion = rd.U8()
			dir = rd.U16()
			pos = render.V3{rd.F32(), rd.F32(), rd.F32()}
			return motion, dir, pos
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for 0x0501 turn packet")
			return 0, 0, render.V3{}
		}
	}
	assertQuiet := func() {
		t.Helper()
		select {
		case p := <-sent:
			t.Fatalf("unexpected packet while not turning: 0x%04X", proto.Op(p))
		case <-time.After(25 * time.Millisecond):
		}
	}

	a := inputApp()
	a.world = w
	origin := render.V3{10, 4, 20}
	a.wv = &worldView{pos: origin, eye: render.V3{10, 6, 15}, facing: render.V3{0, 0, 1}}

	// Idle frames, opposing A+D, dead state and immobilized state must send nothing.
	a.updateWorld(&platform.Input{}, .2)
	a.updateWorld(&platform.Input{Down: "AD"}, .2)
	a.state.Self.Dead = true
	a.updateWorld(&platform.Input{Down: "A"}, .2)
	a.state.Self.Dead = false
	a.state.Self.Speed = &proto.CharacterSpeed{Move: 255}
	a.updateWorld(&platform.Input{Down: "A"}, .2)
	a.state.Self.Speed = nil
	assertQuiet()

	// Holding A while standing sends stationary 0x0501 updates every 100 ms,
	// and observers keep standing posture without triggering walk/run animations.
	observer := &game.Entity{X: origin[0], Y: origin[1], Z: origin[2]}
	actor := &remoteActor{moveX: origin[0], moveZ: origin[2]}

	var lastDir uint16
	for i := range 3 {
		a.updateWorld(&platform.Input{Down: "A"}, .1)
		motion, dir, pos := recvMove()
		if motion != 0 || pos != origin || dir == lastDir {
			t.Fatalf("step %d: motion=%d dir=%d (prev %d) pos=%v", i, motion, dir, lastDir, pos)
		}
		lastDir = dir
		observer.Dir = dir
		observer.MoveSerial++
		if walking, moving := actor.observeMovement(observer, float64(i+1)*.1); walking || moving {
			t.Fatal("stationary turn triggered walk/run animation on observer")
		}
	}

	// Releasing D after a partial <100 ms turn flushes the final heading once, then goes quiet.
	a.updateWorld(&platform.Input{Down: "D"}, .04)
	a.updateWorld(&platform.Input{}, 1.0/60)
	motion, finalDir, pos := recvMove()
	wantDir := uint16(math.Atan2(float64(a.wv.facing[0]), float64(a.wv.facing[2])) / (2 * math.Pi) * 65536)
	if motion != 0 || pos != origin || finalDir != wantDir {
		t.Fatalf("release packet: motion=%d dir=%d (want %d) pos=%v", motion, finalDir, wantDir, pos)
	}
	a.updateWorld(&platform.Input{}, .2)
	assertQuiet()
}
