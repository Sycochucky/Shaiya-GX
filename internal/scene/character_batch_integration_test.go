//go:build integration

package scene

import (
	"bytes"
	"os"
	"runtime"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/archive"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
)

func TestCharacterBatchIndependentPalettesGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires original assets and GL")
	}
	t.Run("automatic", func(t *testing.T) { testCharacterBatchPalettesGL(t, false) })
	t.Run("bound", func(t *testing.T) { testCharacterBatchPalettesGL(t, true) })
}

func testCharacterBatchPalettesGL(t *testing.T, disableBindless bool) {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	const size = 400
	win, err := platform.Open(platform.Config{Title: "Crowd batches", Width: size, Height: size, Hidden: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err := render.InitWithOptions(render.InitOptions{Debug: true, DisableBindless: disableBindless}); err != nil {
		t.Fatal(err)
	}
	arc, err := archive.Open(os.Getenv("SHAIYA_DATA"))
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	assets := NewAssets(arc)
	defer assets.CloseCharacterResources()
	s, err := render.NewScene3D()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Delete()
	k, err := render.NewSkinner()
	if err != nil {
		t.Fatal(err)
	}
	defer k.Delete()
	s.Proj = render.PerspectiveFovLH(.8, 1, .1, 120)
	s.LightDir, s.HideWeaponEffects = render.V3{-1, -1, -1}, true
	actors := make([]*Character, MinCrowdActors)
	for i := range actors {
		look := Appearance{Race: 0, Job: 0, Helmet: -1, WeaponType: 1, CloakType: 24}
		actor, err := BuildCharacter(assets, look, "002_run")
		if err != nil || !actor.HasCloak() || len(actor.attachments) == 0 {
			t.Fatalf("incomplete actor fixture: %v", err)
		}
		actor.World = render.Basis(render.V3{float32(i%6)*2.6 - 6.5, 0, float32(i/6) * 3}, render.V3{0, 0, -1}, render.V3{0, 1, 0})
		actor.Time = float32(i) * .027
		actors[i] = actor
	}
	actors[3].CorpseFade = .42
	var batch CharacterBatch
	defer batch.Delete()
	capture := func(instanced bool) ([]byte, int) {
		gl.Viewport(0, 0, size, size)
		gl.DepthMask(true)
		gl.ClearDepth(1)
		gl.ClearColor(.12, .15, .17, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		s.Begin()
		render.ResetDrawCalls()
		if instanced {
			batch.Draw(k, s, actors)
		} else {
			// Reference depths are complete before translucent weapon effects.
			showEffects := !s.HideWeaponEffects
			s.HideWeaponEffects = true
			for _, c := range actors {
				// Compare the compact crowd palette with the complete serial one.
				c.preparePose()
				clear(c.poseBones)
				c.poseReady = false
				c.Draw(k, s)
				c.findUsedPoseBones()
				c.poseReady = false
			}
			s.HideWeaponEffects = !showEffects
			if showEffects {
				s.BeginEffects()
				for _, c := range actors {
					for i := range c.attachments {
						a := &c.attachments[i]
						world := a.extra.Mul(c.lastFinals[a.bone]).Mul(c.World)
						c.drawWeaponEnchant(s, a, world)
						if a.glow != nil {
							c.updateWeaponGlow(a, world)
							a.glow.drawInPass(s)
						}
					}
				}
				s.EndEffects()
			}
		}
		pixels := make([]byte, size*size*4)
		gl.ReadPixels(0, 0, size, size, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))
		return pixels, render.DrawCalls()
	}
	for _, eye := range []render.V3{{0, 14, -19}, {1, -14, -19}} {
		s.SetCameraView(eye, render.V3{0, 0, 4}, render.V3{0, 1, 0})
		if eye[1] < 0 {
			s.ClipPlane = [4]float32{0, 1, 0, -.18}
		}
		before, n := capture(false)
		after, m := capture(true)
		if !bytes.Equal(before, after) {
			different := 0
			for i := range before {
				if before[i] != after[i] {
					different++
				}
			}
			t.Fatalf("instancing changed %d color bytes for eye %v", different, eye)
		}
		if m >= n/2 {
			t.Fatalf("expected substantial draw reduction, got %d -> %d", n, m)
		}
		t.Logf("independent palettes and equipment, eye %v, %d -> %d draws", eye, n, m)
	}
	for _, actor := range actors {
		actor.cloak.prepared = true
		actor.World[12] += .1
	}
	capture(true)
	if batch.Stats.ClothUploads != 0 || batch.Stats.ClothBytes != 0 {
		t.Fatalf("moving unchanged cloth uploaded positions again: %+v", batch.Stats)
	}
	for _, actor := range actors {
		actor.cloak.prepared = true
	}
	changed := actors[2].cloak
	changed.vertices[0].Pos[0] += .2
	changed.version++
	capture(true)
	if want := (len(changed.vertices) + 1) * 16; batch.Stats.ClothUploads != 1 || batch.Stats.ClothBytes != want {
		t.Fatalf("one updated cloth submitted %+v, want %d bytes", batch.Stats, want)
	}
	for i, actor := range actors {
		actor.enchantTime = 1.0 + float64(i)*.01
		actor.attachments[0].enchant = actor.newWeaponEnchant(1, 10)
		actor.attachments[0].glow = assets.itemGlow(1, int32(i%4))
	}
	s.HideWeaponEffects = false
	for _, eye := range []render.V3{{0, 14, -19}, {1, -14, -19}} {
		s.SetCameraView(eye, render.V3{0, 0, 4}, render.V3{0, 1, 0})
		s.ClipPlane = [4]float32{}
		if eye[1] < 0 {
			s.ClipPlane = [4]float32{0, 1, 0, -.18}
		}
		before, n := capture(false)
		after, m := capture(true)
		if batch.Stats.WeaponFX == 0 || m >= n || !bytes.Equal(before, after) {
			t.Fatalf("weapon effects differ in camera %v, calls %d → %d, batch %+v", eye, n, m, batch.Stats)
		}
		t.Logf("weapon glows/enchantments and opaque depth, eye %v, %d -> %d draws", eye, n, m)
	}
	for i := range actors {
		mob, err := assets.BuildMonster([]uint16{1, 2, 3, 2009}[i%4])
		if err != nil {
			t.Fatal(err)
		}
		mob.SetTransform(render.Basis(render.V3{float32(i%6)*3 - 7.5, 0, float32(i/6) * 3}, render.V3{0, 0, -1}, render.V3{0, 1, 0}))
		mob.Update(float32(i) * .027)
		actors[i] = mob.Character
	}
	actors[3].CorpseFade = .42
	s.ClipPlane = [4]float32{}
	s.SetCameraView(render.V3{0, 14, -22}, render.V3{0, 0, 4}, render.V3{0, 1, 0})
	before, n := capture(false)
	after, m := capture(true)
	if !bytes.Equal(before, after) || m >= n/2 {
		t.Fatalf("monster batches changed the image or missed the draw reduction: %d -> %d draws", n, m)
	}
	t.Logf("monster poses, materials and corpse fade, %d -> %d draws", n, m)
	batch.BeginFrame()
	first, _ := capture(true)
	written := batch.Stats.PaletteBytes
	mirror := render.Identity()
	mirror[5] = -1
	s.View = mirror.Mul(s.View)
	s.ClipPlane = [4]float32{0, 1, 0, -.18}
	secondReference, _ := capture(false)
	second, _ := capture(true)
	secondWritten := batch.Stats.PaletteBytes
	batch.EndFrame()
	if written == 0 || secondWritten != 0 || !bytes.Equal(first, before) || !bytes.Equal(second, secondReference) {
		t.Fatalf("camera passes did not share the palettes while preserving pixels: %d -> %d bytes", written, secondWritten)
	}
	t.Logf("mob palettes across two cameras: %d bytes then %d", written, secondWritten)
	actors[0].Update(.12)
	freshReference, _ := capture(false)
	batch.BeginFrame()
	fresh, _ := capture(true)
	batch.EndFrame()
	if batch.Stats.PaletteBytes == 0 || !bytes.Equal(fresh, freshReference) {
		t.Fatal("the next frame did not upload the updated animation")
	}
	if err := gl.GetError(); err != gl.NO_ERROR {
		t.Fatalf("OpenGL error %#x", err)
	}
}
