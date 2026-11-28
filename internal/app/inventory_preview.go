package app

import (
	"log"
	"math"

	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

// The full-body opening starts below the title at y=33 and ends at the bags.
// Fit height, not the wider bind-pose arm span, to preserve the reference size.
// A separate actor owns its idle time, sharing only cached geometry/assets.
func (a *App) drawInventoryPreview(w, h int) {
	slot, ok := a.selfSlot()
	if !ok {
		return
	}
	v := &a.inventory
	ap := a.appearance(&slot)
	if !v.previewLoaded || ap != v.previewAppearance {
		v.previewLoaded, v.previewAppearance = true, ap
		var err error
		v.preview, err = scene.BuildCharacter(a.assets, ap, "000_normal")
		if err != nil {
			log.Printf("inventory preview: %v", err)
		}
	}
	if v.preview == nil {
		return
	}
	x, y := a.inventoryUI.Origin(w, h)
	a.drawEquipmentActor(v.preview, ap, x+54, y+33, 136, 217, w, h)
}

func (a *App) drawEquipmentActor(actor *scene.Character, ap scene.Appearance, x, y, width, height, w, h int) {
	a.sprites.Flush()
	render.InViewport(x, y, width, height, h, func() {
		s := a.actorPreviewScene()
		centerY, distance := actor.InventoryFrame(34*math.Pi/180, ap.HeightScale())
		s.SetCameraView(render.V3{0, centerY, distance}, render.V3{0, centerY, 0}, render.V3{0, 1, 0})
		s.Proj = render.PerspectiveFovLH(34*math.Pi/180, float32(width)/float32(height), .1, 100)
		s.FogStart, s.FogEnd = 10000, 20000
		// Match the world character's frontal fill instead of lighting its back.
		s.LightDir = render.V3{-1, -1, -1}
		s.CharacterAmbient = 1
		s.Begin()
		scale := ap.ScaleMatrix()
		actor.World = scale.Mul(render.Basis(render.V3{}, render.V3{0, 0, -1}, render.V3{0, 1, 0}))
		actor.Update(float32(a.frameDT))
		actor.Draw(a.skinner, &s)
		s.End()
	})
	a.sprites.Begin(w, h)
}

// UI actors share material/pose rendering but own camera state. World clipping,
// shadow silhouettes and water state must never leak into a cached portrait.
func (a *App) actorPreviewScene() render.Scene3D {
	s := *a.scene3d
	s.Shadow = false
	s.Occlusion = nil
	s.GeometryReduction = 0
	s.ClipPlane = [4]float32{}
	s.ReverseDepth = false
	s.WaterReflection = nil
	s.WaterNormals = [2]*render.Texture{}
	s.WaterBelow = false
	s.FogStart, s.FogEnd = 0, 0
	return s
}
