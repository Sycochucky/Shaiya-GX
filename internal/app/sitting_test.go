package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"testing"
)

func TestPostureToggleOnePerPress(t *testing.T) {
	input := &platform.Input{Chars: []rune{'c', 'C'}, CharRepeats: []bool{true, true}}
	if postureToggle(input, nil) {
		t.Fatal("holding C toggled posture")
	}
	input.CharRepeats[1] = false
	if !postureToggle(input, nil) {
		t.Fatal("new press of C was ignored")
	}
}

func TestSeatedMovementOnlyStandsForKeyboard(t *testing.T) {
	for _, in := range []platform.Input{
		{Clicked: true},
		{MouseDown: true, RightDown: true},
		{Down: "W", SuppressAutoRun: true},
	} {
		if keyboardPostureMovement(&in, false) {
			t.Fatalf("input %+v should keep the character seated", in)
		}
	}
	for _, key := range "WSADQE" {
		if !keyboardPostureMovement(&platform.Input{Down: string(key)}, false) {
			t.Errorf("%c should stand the character", key)
		}
	}
	if !keyboardPostureMovement(&platform.Input{Keys: []platform.Key{platform.KeySpace}}, false) ||
		!keyboardPostureMovement(&platform.Input{}, true) {
		t.Fatal("jump and keyboard autorun should still stand the character")
	}
}

func TestRemoteSittingKeepsSeatedPoseWithoutMovement(t *testing.T) {
	e := &game.Entity{Sitting: true}
	if clip := remoteLocomotion(nil, e); clip != "012_sit" {
		t.Fatal("seated actor chose", clip)
	}
	e.Sitting = false
	if clip := remoteLocomotion(nil, e); clip != "000_normal" {
		t.Fatal("standing actor chose", clip)
	}
}

func TestAttackQueuesWhilePostureIsPending(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 1
	a.state.Self.Sitting = true
	a.wv.sitPending = true
	a.state.Target = 2
	a.state.Entities[2] = &game.Entity{ID: 2, Kind: game.KindMob}
	if !a.attackTarget() || a.wv.sitAttack != 2 {
		t.Fatal("attack was not deferred until standing")
	}
	a.wv.sitAttack = 0
	a.state.Entities[2].Dead = true
	if !a.attackTarget() || a.wv.sitAttack != 0 {
		t.Fatal("dead target queued an attack")
	}
}
