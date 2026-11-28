//go:build integration

package render

import (
	"bytes"
	"os"
	"runtime"
	"testing"

	img "github.com/Spelunkern/Shaiya-GX/internal/data/image"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
)

func TestEffectBatchPreservesBlendingGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	window, err := platform.Open(platform.Config{Title: "effect batch", Width: 64, Height: 64, Hidden: true, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	defer window.Close()
	if err := Init(true); err != nil {
		t.Fatal(err)
	}
	s, err := NewScene3D()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Delete()
	s.View, s.Proj = Identity(), Identity()
	mesh := NewMesh([]Vertex{{Pos: V3{-.6, -.6, .5}}, {Pos: V3{.6, -.6, .5}}, {Pos: V3{.6, .6, .5}}, {Pos: V3{-.6, .6, .5}}}, []uint32{0, 1, 2, 0, 2, 3})
	defer mesh.Delete()
	texture := Upload(img.Texture{RGBA: &img.RGBA{W: 1, H: 1, Pix: []byte{255, 255, 255, 255}}})
	defer texture.Delete()
	worlds := []M4{Identity(), Identity(), Identity()}
	worlds[0][12], worlds[1][12], worlds[2][13] = -.2, .2, .2
	colors := [][4]float32{{1, 0, 0, .4}, {0, 1, 0, .6}, {0, 0, 1, .7}}
	draw := func(batched bool, source, destination int32) ([]byte, int) {
		gl.Viewport(0, 0, 64, 64)
		gl.DepthMask(true)
		gl.ClearColor(.1, .1, .1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		s.Begin()
		ResetDrawCalls()
		if batched {
			s.DrawEffectBatch(mesh, texture, worlds, colors, source, destination)
		} else {
			for i := range worlds {
				s.DrawEffect(mesh, texture, worlds[i], colors[i], source, destination)
			}
		}
		// A later ordinary effect must not inherit the per-instance colors.
		following := Identity()
		following[12], following[13] = .4, -.4
		s.DrawEffect(mesh, texture, following, [4]float32{.3, .7, .2, .4}, 4, 5)
		calls := DrawCalls()
		pixels := make([]byte, 64*64*4)
		gl.ReadPixels(0, 0, 64, 64, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))
		return pixels, calls
	}
	for _, blend := range [][2]int32{{4, 5}, {4, 1}, {1, 1}, {2, 0}, {11, 8}, {12, 8}} {
		old, before := draw(false, blend[0], blend[1])
		now, after := draw(true, blend[0], blend[1])
		if !bytes.Equal(old, now) {
			t.Fatalf("batch changed blended pixels for %v", blend)
		}
		if before != 4 || after != 2 {
			t.Fatalf("blend %v: draw calls %d -> %d", blend, before, after)
		}
	}
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatalf("GL error %#x", code)
	}
}

func TestOrderedEffectMultiDrawGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL")
	}
	t.Run("automatic", func(t *testing.T) { testOrderedEffectMultiDrawGL(t, InitOptions{}) })
	t.Run("bound", func(t *testing.T) { testOrderedEffectMultiDrawGL(t, InitOptions{DisableBindless: true}) })
	t.Run("core", func(t *testing.T) {
		testOrderedEffectMultiDrawGL(t, InitOptions{DisableBindless: true, DisableDrawParameters: true})
	})
}

func testOrderedEffectMultiDrawGL(t *testing.T, options InitOptions) {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	window, err := platform.Open(platform.Config{Title: "ordered effects", Width: 64, Height: 64, Hidden: true})
	if err != nil {
		t.Fatal(err)
	}
	defer window.Close()
	if err := InitWithOptions(options); err != nil {
		t.Fatal(err)
	}
	s, err := NewScene3D()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Delete()
	s.SetCameraView(V3{63783, 0, -3}, V3{63783, 0, 0}, V3{0, 1, 0})
	s.Proj = PerspectiveFovLH(1.1, 1, .1, 20)
	verts := []Vertex{
		{Pos: V3{-.7, -.7, 0}, UV: [2]float32{-.3, 0}},
		{Pos: V3{.7, -.7, 0}, UV: [2]float32{1.3, 0}},
		{Pos: V3{.7, .7, 0}, UV: [2]float32{1.3, 1}},
		{Pos: V3{-.7, .7, 0}, UV: [2]float32{-.3, 1}},
	}
	indices := []uint32{0, 1, 2, 0, 2, 3}
	quad := NewMesh(verts, indices)
	defer quad.Delete()
	var frames EffectGeometryCache
	defer frames.Delete()
	animated := func(id uint64, shift float32) *EffectGeometry {
		return frames.Geometry(id, verts, indices, 8, func() []EffectKeyVertex {
			keys := make([]EffectKeyVertex, 0, 8)
			for i := range 2 {
				for _, v := range verts {
					keys = append(keys, EffectKeyVertex{Position: [4]float32{v.Pos[0] + shift*float32(i), v.Pos[1], v.Pos[2], 1}, UV: [4]float32{v.UV[0], v.UV[1] + .35*float32(i)}})
				}
			}
			return keys
		})
	}
	left, right := animated(1, -.3), animated(2, .3)
	a := Upload(img.Texture{RGBA: &img.RGBA{W: 2, H: 1, Pix: []byte{255, 80, 30, 190, 50, 200, 150, 40}}})
	b := Upload(img.Texture{RGBA: &img.RGBA{W: 2, H: 1, Pix: []byte{30, 80, 255, 170, 20, 230, 80, 60}}})
	defer a.Delete()
	defer b.Delete()
	worlds, colors, poses := make([]M4, 3), make([][4]float32, 3), make([]EffectPose, 3)
	for i := range worlds {
		worlds[i] = Identity()
		worlds[i][12], worlds[i][13] = 63783+float32(i-1)*.18, float32(i-1)*.12
		colors[i] = [4]float32{.6, .7, .9, .52}
		poses[i] = EffectPose{0, 4, float32(i+1) / 4}
	}
	pipe := &s.effectPipeline
	pipe.init()
	if pipe.program == 0 || (sceneStorageSupported && pipe.indirect.program == 0) {
		t.Fatal("supported ordered effect shader did not link")
	}
	if options.DisableBindless && pipe.indirect.program != 0 && !pipe.indirect.bound {
		t.Fatal("effect batching selected bindless despite the override")
	}
	if options.DisableDrawParameters && sceneStorageSupported && !pipe.indirect.materialAttribute {
		t.Fatal("effect batching did not select the core material attribute")
	}
	compact, indirect := pipe.program, pipe.indirect.program
	defer func() { pipe.program, pipe.indirect.program = compact, indirect }()
	gl.Viewport(0, 0, 64, 64)
	draw := func(program, multi uint32) ([]byte, int) {
		pipe.program, pipe.indirect.program = program, multi
		s.Begin()
		gl.ClearColor(.06, .07, .1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		occluder := worlds[1]
		occluder[0], occluder[5], occluder[14] = .35, .35, -.1
		s.drawModel(quad, a, occluder, -1, false, false)
		s.End()
		ResetDrawCalls()
		s.BeginEffects()
		for _, blend := range [][2]int32{{4, 5}, {4, 1}, {4, 5}} {
			s.DrawEffectBatch(quad, a, worlds, colors, blend[0], blend[1], true)
			s.DrawEffectGeometry(left, b, worlds, colors, poses, blend[0], blend[1])
			s.DrawEffectBatch(quad, a, worlds, colors, blend[0], blend[1])
			s.DrawEffectGeometry(right, a, worlds, colors, poses, blend[0], blend[1])
		}
		s.EndEffects()
		pixels := make([]byte, 64*64*4)
		gl.ReadPixels(0, 0, 64, 64, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))
		return pixels, DrawCalls()
	}
	for _, clip := range [][4]float32{{}, {0, 1, 0, 0}} {
		s.ClipPlane = clip
		for _, reverse := range []bool{false, true} {
			s.ReverseDepth = reverse
			s.TemporalJitter = [2]float32{.002, -.001}
			reference, before := draw(0, 0)
			now, compactCalls := draw(compact, 0)
			if before != 12 || compactCalls != before || !bytes.Equal(reference, now) || pipe.instances.bytes != 36*96 {
				t.Fatalf("compact queue differs with clip=%v reverse=%v: calls=%d/%d, bytes=%d", clip, reverse, before, compactCalls, pipe.instances.bytes)
			}
			if indirect != 0 {
				now, calls := draw(compact, indirect)
				if calls != 3 || !bytes.Equal(reference, now) {
					t.Fatalf("multi-draw differs with clip=%v reverse=%v: %d calls", clip, reverse, calls)
				}
			}
		}
	}
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatalf("GL error %#x", code)
	}
}
