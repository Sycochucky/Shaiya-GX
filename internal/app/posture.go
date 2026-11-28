package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

func postureToggle(in *platform.Input, actor *scene.Character) bool {
	if actor.PlayingAction(10) || actor.PlayingAction(11) {
		return false
	}
	for i, c := range in.Chars {
		if (c == 'c' || c == 'C') && (i >= len(in.CharRepeats) || !in.CharRepeats[i]) {
			return true
		}
	}
	return false
}

// Only keyboard travel (including keyboard autorun and jump) stands a
// seated character. Terrain travel and two-button mouse-run do not.
func keyboardPostureMovement(in *platform.Input, autoRun bool) bool {
	return !in.SuppressAutoRun && (autoRun || in.Held('W') || in.Held('S') ||
		in.Held('A') || in.Held('D') || in.Held('Q') || in.Held('E') || in.HasKey(platform.KeySpace))
}

// Rapture retains auto-attack when a motion arrives. Stop it before sitting;
// the posture itself remains pending until the echoed 0x0506 arrives.
func (a *App) requestPosture(motion uint8) bool {
	// Movement and queued attacks can stand up once the sit clip ends.
	if motion == proto.MotionStandUp && a.wv.self.PlayingAction(10) {
		return true
	}
	if motion == proto.MotionSitDown {
		a.wv.sitAttack = 0
		if err := a.world.Send(proto.StopAttackPacket()); err != nil {
			a.backToLoginID(12170)
			return false
		}
	}
	if err := a.world.Send(proto.MotionPacket(motion)); err != nil {
		a.backToLoginID(12171)
		return false
	}
	a.wv.sitPending = true
	a.wv.sitDeadline = a.clock + 30
	return true
}
