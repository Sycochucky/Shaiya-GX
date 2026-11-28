package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

type reflectionInput struct {
	target                      *render.WaterTarget
	world                       *scene.World
	revision                    uint64
	view, projection, sky       render.M4
	jitter                      [2]float32
	region                      render.ScreenRegion
	fog                         render.V3
	fogStart, fogEnd, weather   float32
	settings                    ui.ClientSettings
	personalLight, reverseDepth bool
}

type reflectionCapture struct {
	input reflectionInput
	age   float64
}

// The main scene and water normals still animate every frame. Only reuse the
// reflected scene while its screen projection and rendering inputs match.
func (c *reflectionCapture) reuse(input reflectionInput, dt float64) bool {
	const maxAge = 1.0 / 30
	if input.target == nil || c.input != input || !(dt > 0 && dt < maxAge) {
		return false
	}
	c.age += dt
	return c.age < maxAge-1e-9
}
