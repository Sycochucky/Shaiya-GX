package app

import "testing"

func TestAbandonAttackKeepsSwingPause(t *testing.T) {
	a := inputApp()
	a.clock = 10
	a.wv.attackSent, a.wv.engagedID, a.wv.approachTarget = true, 7, 7
	a.wv.attackMoveUntil = 10.5
	if !a.abandonAttack() {
		t.Fatal("abandon failed")
	}
	v := a.wv
	if v.attackSent || v.engagedID != 0 || v.approachTarget != 0 {
		t.Fatal("attack order survived explicit movement")
	}
	if v.attackMoveUntil != 10.5 {
		t.Fatal("abandon changed the existing movement recovery")
	}
}
