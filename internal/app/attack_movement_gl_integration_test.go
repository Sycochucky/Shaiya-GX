//go:build integration

package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/archive"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"os"
	"runtime"
	"testing"
)

func TestAttackPauseStopsPredictedMovementGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL and original assets")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	win, err := platform.Open(platform.Config{Width: 128, Height: 128, Hidden: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err = render.Init(true); err != nil {
		t.Fatal(err)
	}
	arc, err := archive.Open(os.Getenv("SHAIYA_DATA"))
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	a, err := New(Config{}, arc)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Preview("world-stream"); err != nil {
		t.Fatal(err)
	}
	v := a.wv
	a.clock = 10
	v.sendT = 1000
	before := v.pos
	const mobID = 1729
	a.state.Entities[mobID] = &game.Entity{ID: mobID, Kind: game.KindMob, X: before[0] + 1, Y: before[1], Z: before[2]}
	a.state.Target = mobID
	a.keepChasing(mobID)
	attack := proto.NewWriter(proto.CharacterMobAutoAttack).Bytes()
	a.state.LastAttack = proto.AttackResult{Source: a.state.SelfID, Target: mobID, HP: 1}
	a.confirmedAnimation(attack)
	if !v.self.ActionActive() {
		t.Fatal("confirmed basic attack did not start an animation")
	}
	swing, recovery := v.self.ActionID(), v.attackMoveUntil
	back := &platform.Input{Down: "S"}
	for range 12 {
		a.clock += 1.0 / 60
		a.updateWorld(back, 1.0/60)
	}
	if v.attackSent || v.engagedID != 0 || v.approachTarget != 0 || a.state.Target != mobID {
		t.Fatal("backpedal did not stop repeat attacks and keep the selection")
	}
	if !v.self.ActionActive() || v.self.ActionID() != swing || v.self.Time == 0 {
		t.Fatal("backpedal cut the confirmed swing during the server pause")
	}
	if v.pos[0] != before[0] || v.pos[2] != before[2] || v.attackMoveUntil != recovery {
		t.Fatal("backpedal bypassed the confirmed attack recovery")
	}
	a.clock = recovery + .01
	a.updateWorld(back, 1.0/60)
	if v.pos[0] == before[0] && v.pos[2] == before[2] {
		t.Fatal("backpedal did not resume when the attack pause expired")
	}
	if !v.self.ActionActive() || v.self.ActionID() != swing {
		t.Fatal("resuming movement replaced the tail of the swing")
	}
	a.confirmedAnimation(attack)
	if v.attackMoveUntil != recovery || v.self.ActionID() != swing {
		t.Fatal("a late attack reply replaced the swing or paused ongoing backpedal")
	}
	for i := 0; i < 240 && !v.self.PlayingClip("003_bstep"); i++ {
		a.clock += 1.0 / 60
		a.updateWorld(back, 1.0/60)
	}
	if !v.self.PlayingClip("003_bstep") {
		t.Fatal("the swing did not finish into the backpedal clip")
	}

	// A late combat result must never cancel a later jump action.
	v.startJumpAnimation()
	v.beginJump()
	v.moving, v.motion, v.sendT = true, 0x84, 1000
	a.updateWorld(&platform.Input{Down: "W"}, 1.0/60)
	if !v.self.PlayingAction(8) {
		t.Fatal("stale attack recovery canceled a later jump")
	}

	// A reply to a ground attack or spell can arrive after local takeoff.
	action := v.self.ActionID()
	a.state.LastCast = proto.SkillCast{Source: a.state.SelfID, ID: 71, Level: 1}
	a.confirmedAnimation(proto.NewWriter(proto.SkillCastStarted).Bytes())
	a.state.LastAttack = proto.AttackResult{Source: a.state.SelfID}
	a.confirmedAnimation(proto.NewWriter(proto.CharacterMobAutoAttack).Bytes())
	if v.self.PlayCast(85, 1) || v.self.PlayAction(8, false) {
		t.Fatal("another cast or jump could replace the jump clip")
	}
	v.moving, v.motion, v.sendT = false, 0x84, 1000
	a.updateWorld(&platform.Input{}, 1.0/60)
	if v.self.ActionID() != action || !v.self.PlayingAction(8) || !v.jumping || a.attackMovementAllowed() {
		t.Fatal("late combat reply cut the jump or enabled an attack")
	}
	a.clock = v.attackMoveUntil + .01
	v.moving, v.motion, v.sendT = true, 0x84, 1000
	a.updateWorld(&platform.Input{Down: "W"}, 1.0/60)
	if v.self.ActionID() != action || !v.self.PlayingAction(8) {
		t.Fatal("late attack identified the jump as an interruptible attack pose")
	}

}
