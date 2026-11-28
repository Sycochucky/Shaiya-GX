//go:build integration

package app

import (
	"math"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/data/archive"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
	"github.com/Spelunkern/Shaiya-GX/internal/testworld"
)

// Map 15 has no terrain heightmap; a click on its floor must still select a
// destination on the authored dungeon floor.
func TestDungeonClickMoveSelectsFloorGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("GL and game resources")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	const W, H = 1600, 900
	win, e := platform.Open(platform.Config{Width: W, Height: H, Hidden: true})
	if e != nil {
		t.Fatal(e)
	}
	defer win.Close()
	if e = render.Init(false); e != nil {
		t.Fatal(e)
	}
	arc, e := archive.OpenDirectory(os.Getenv("SHAIYA_DATA"))
	if e != nil {
		t.Fatal(e)
	}
	defer arc.Close()
	a, e := New(Config{}, arc)
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	a.state = game.New()
	a.state.Phase = game.PhaseInWorld
	a.state.MapID, a.state.SelfID = 15, 1
	a.state.Slots[0] = proto.CharacterSlot{ID: 1, Race: 0, Class: 0, Height: 2}
	a.state.Self.X, a.state.Self.Y, a.state.Self.Z = 175, 7.6, 26
	a.screen = screenWorld
	a.enterWorld()
	a.frameDT = 1.0 / 60
	for range 30 {
		a.Draw(W, H)
	}
	if a.wv.world.HasTerrain() {
		t.Fatal("map 15 should be a dungeon")
	}
	groundClick := &platform.Input{MouseX: W / 2, MouseY: H/2 + 120, Clicked: true}
	a.state.Self.Sitting = true
	a.moveToGround(groundClick, W, H)
	if a.wv.travel || a.wv.travelEffect != nil || a.wv.sitPending {
		t.Fatal("a floor click started travel while seated")
	}
	a.state.Self.Sitting = false
	a.moveToGround(groundClick, W, H)
	d := a.wv.destination
	if !a.wv.travel || d[1] > a.wv.pos[1]+.5 || d.Sub(a.wv.pos).Len() > 40 {
		t.Fatalf("destination %v travel %v from %v", d, a.wv.travel, a.wv.pos)
	}
	if a.wv.travelEffect == nil || a.wv.travelEffect.Origin != d {
		t.Fatal("point.EFT is missing or anchored off the selected floor")
	}
	for range 120 {
		a.updateTravelEffect(.05)
	}
	if a.wv.travelEffect.Done() {
		t.Fatal("point.EFT expired before arrival")
	}
	// Render the actual destination marker through the dungeon camera. This
	// checks its authored emitters, not just the presence of an effect pointer.
	gl.Viewport(0, 0, W, H)
	gl.ClearColor(0, 0, 0, 1)
	gl.ClearDepth(1)
	gl.DepthMask(true)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	a.scene3d.Begin()
	a.wv.travelEffect.Draw(a.scene3d)
	a.scene3d.End()
	pixels := make([]byte, W*H*4)
	gl.ReadPixels(0, 0, W, H, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))
	visible := 0
	for i := 0; i < len(pixels); i += 4 {
		if pixels[i] != 0 || pixels[i+1] != 0 || pixels[i+2] != 0 {
			visible++
		}
	}
	if visible < 10 {
		t.Fatalf("dungeon movement marker invisible: %d pixels", visible)
	}
	t.Logf("destination=%v, point.EFT=%d visible pixels after6seconds", d, visible)
	a.wv.travel = false
	a.updateTravelEffect(.016)
	if a.wv.travelEffect != nil {
		t.Fatal("arrival/cancellation retained the movement marker")
	}
	// The existing heightmap path must retain its own exact destination.
	a.releaseWorld()
	a.state.MapID = 1
	a.state.Self.X, a.state.Self.Y, a.state.Self.Z = 578, 77.705, 1762
	a.enterWorld()
	for range 30 {
		a.Draw(W, H)
	}
	v := a.wv
	if !v.world.HasTerrain() {
		t.Fatal("field fixture has no heightmap")
	}
	goal := v.pos.Add(v.facing.Scale(3))
	goal[1] = v.world.Height(goal[0], goal[2])
	x, y, _, projected := projectLabel(goal, a.scene3d.View.Mul(a.scene3d.Proj), W, H)
	if !projected {
		t.Fatal("field fixture point is outside camera")
	}
	click := &platform.Input{MouseX: x, MouseY: y, Clicked: true}
	a.state.Target = 7
	a.keepChasing(a.state.Target)
	recovery := a.clock + 1
	v.attackMoveUntil = recovery
	a.gameMenu.Settings.UseMouse = false
	a.moveToGround(click, W, H)
	if v.travel || v.travelEffect != nil || !v.attackSent {
		t.Fatal("terrain click affected movement or combat with Use Mouse disabled")
	}
	a.gameMenu.Settings.UseMouse = true
	a.updateCombat(click, W, H)
	if !v.travel || v.destination.Sub(goal).Len() > .1 || v.travelEffect == nil || v.travelEffect.Origin != v.destination {
		t.Fatal("field click behavior changed", goal, v.destination)
	}
	if v.attackSent || v.engagedID != 0 || v.approachTarget != 0 || v.attackMoveUntil != recovery || a.state.Target != 7 {
		t.Fatal("ground click did not abandon the attack while keeping its target selected")
	}
	a.state.LastAttack = proto.AttackResult{Source: a.state.SelfID, Target: 7}
	a.confirmedAnimation(proto.NewWriter(proto.CharacterMobAutoAttack).Bytes())
	if !v.travel || v.attackMoveUntil != recovery {
		t.Fatal("late swing stopped ground travel")
	}
}

// This test requires a disposable account positioned in map15 by the N2 runner.
func TestDungeonClickMoveAgainstIsolatedGX(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_DUNGEON_CLICK") != "1" || os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires disposable dungeon N2 fixture")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	const W, H = 1600, 900
	win, err := platform.Open(platform.Config{Width: W, Height: H, Hidden: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err = render.Init(false); err != nil {
		t.Fatal(err)
	}
	arc, err := archive.OpenDirectory(os.Getenv("SHAIYA_DATA"))
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	st, world, _, _ := testworld.Enter(t)
	if st.MapID != 15 {
		t.Fatal("fixture is not in dungeon15", st.MapID)
	}
	a, err := New(Config{}, arc)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	a.state, a.world, a.screen = st, world, screenWorld
	a.enterWorld()
	a.frameDT = 1.0 / 60
	for range 30 {
		a.Draw(W, H)
	}
	a.pumpWorld()
	v := a.wv
	before := v.pos
	v.autoRun = true
	a.updateCharacters(&platform.Input{Keys: []platform.Key{platform.KeyEnter}}, W, H, 1.0/60)
	if !a.chatOpen || !v.autoRun || v.pos == before {
		t.Fatal("opening chat interrupted auto-run")
	}
	time.Sleep(time.Second / 60)
	a.updateCharacters(&platform.Input{Chars: []rune{'o', 'w'}, Down: "W", Space: true}, W, H, 1.0/60)
	if a.chatInput.Text != "ow" || a.menuOpen || v.jumping || !v.autoRun {
		t.Fatal("chat typing leaked into world controls")
	}
	time.Sleep(time.Second / 60)
	before = v.pos
	vp := a.scene3d.View.Mul(a.scene3d.Proj)
	var click platform.Input
	for y := H/2 + 100; y < H/2+220; y += 12 {
		candidate := platform.Input{MouseX: W / 2, MouseY: y, Clicked: true}
		origin, direction, ok := clickRay(candidate.MouseX, y, W, H, vp)
		if !ok || a.pickMob(&candidate, W, H) != 0 || a.pickNPC(&candidate, W, H) != 0 || a.pickGroundModel(&candidate, W, H) != 0 {
			continue
		}
		goal, ok := v.world.RayFloor(origin, direction, clickMoveDistance)
		if ok && goal.Sub(before).Len() > .5 && goal.Sub(before).Len() < 30 {
			click = candidate
			break
		}
	}
	if !click.Clicked {
		t.Fatal("no clear dungeon floor under the cursor")
	}
	press := click
	press.Pressed, press.MouseDown, press.Clicked = true, true, false
	a.updateCharacters(&press, W, H, 0)
	a.updateCharacters(&click, W, H, 0)
	goal := v.destination
	if !a.chatOpen || !v.travel || v.autoRun || v.travelEffect == nil {
		t.Fatalf("click did not start travel with chat open: chat=%t travel=%t auto-run=%t target=%d marker=%t", a.chatOpen, v.travel, v.autoRun, st.Target, v.travelEffect != nil)
	}
	if distance := goal.Sub(before).Len(); distance < .5 || distance > 30 {
		t.Fatal("fixture selected unsuitable distance", before, goal)
	}
	last := time.Now()
	deadline := last.Add(10 * time.Second)
	for v.travel && time.Now().Before(deadline) {
		now := time.Now()
		dt := now.Sub(last).Seconds()
		last = now
		a.frameDT = dt
		a.pumpWorld()
		if a.wv != v {
			t.Fatal("left dungeon during click movement")
		}
		a.updateCharacters(&platform.Input{}, W, H, dt)
		a.updateTravelEffect(float32(dt))
		time.Sleep(time.Second / 60)
	}
	if v.travel || math.Hypot(float64(v.pos[0]-goal[0]), float64(v.pos[2]-goal[2])) > .1 {
		t.Fatal("did not reach clicked floor", before, goal, v.pos)
	}
	a.updateTravelEffect(.016)
	if v.travelEffect != nil {
		t.Fatal("arrival did not remove point.EFT")
	}
	// Give the last authoritative checks time to arrive before the assertion.
	for range 20 {
		a.pumpWorld()
		time.Sleep(time.Second / 60)
	}
	if st.LastPlayerCorrection.ID != 0 {
		t.Fatal("server rejected click movement", st.LastPlayerCorrection)
	}
	t.Logf("map15 click movement accepted: %v -> %v (goal%v); point.EFT cleared on arrival", before, v.pos, goal)
	if !a.chatOpen || a.chatInput.Text != "ow" {
		t.Fatal("click travel closed chat or lost the draft")
	}
	a.updateCharacters(&platform.Input{Keys: []platform.Key{platform.KeyEscape}}, W, H, 0)
	// A manual movement command cancels both travel and its marker.
	v.travel = true
	v.destination = v.pos.Add(v.facing.Scale(2))
	a.startTravelEffect()
	a.updateWorld(&platform.Input{Down: "S"}, .016)
	a.updateTravelEffect(.016)
	if v.travel || v.travelEffect != nil {
		t.Fatal("manual input did not cancel dungeon travel")
	}
}
