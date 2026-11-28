package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
)

// Default USE_MOUSE=true (0x408bf0 / 0x43bfa0). Picking runs after entities,
// loot, NPCs and panels, so a consumed click never becomes a terrain command.
func (a *App) moveToGround(in *platform.Input, width, height int) {
	v := a.wv
	if a.state.Self.Sitting || v.sitPending || !a.clientSettings().UseMouse || a.scene3d == nil || v.world == nil {
		return
	}
	origin, direction, ok := clickRay(in.MouseX, in.MouseY, width, height, a.scene3d.View.Mul(a.scene3d.Proj))
	if !ok {
		return
	}
	if p, hit := v.world.RayFloor(origin, direction, clickMoveDistance); hit {
		if !a.abandonAttack() {
			return
		}
		v.destination, v.travel, v.autoRun = p, true, false
		a.startTravelEffect()
	}
}

// clickMoveDistance is the farthest point a click can select.
const clickMoveDistance = 400
