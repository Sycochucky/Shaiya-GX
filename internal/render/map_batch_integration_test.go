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

func TestBoundMapBatchPreservesMaterialsAndOrderGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL")
	}
	t.Run("automatic", func(t *testing.T) { testBoundMapBatchPreservesMaterialsAndOrderGL(t, false) })
	t.Run("core", func(t *testing.T) { testBoundMapBatchPreservesMaterialsAndOrderGL(t, true) })
}

func testBoundMapBatchPreservesMaterialsAndOrderGL(t *testing.T, core bool) {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	const width, height = 128, 32
	win, err := platform.Open(platform.Config{Width: width, Height: height, Hidden: true, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err := InitWithOptions(InitOptions{DisableBindless: true, DisableDrawParameters: core, Debug: true}); err != nil {
		t.Fatal(err)
	}
	s, err := NewScene3D()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Delete()
	if !mapIndirectSupported {
		t.Skip("requires shader storage blocks")
	}
	if s.mapPrograms().bound == 0 {
		t.Fatal("bound map program failed to link")
	}
	if core && !s.mapPrograms().materialAttribute {
		t.Fatal("bound map program did not select the core material attribute")
	}
	s.View, s.Proj, s.Ambient = Identity(), Identity(), 1
	batch := MapBatch{Transforms: []M4{Identity(), Identity()}}
	batch.Transforms[1][13] = .8
	allPlacements := batch.Placements([]uint32{0, 1})
	firstPlacement := batch.Placements([]uint32{0})
	defer batch.Delete()
	light := Upload(img.Texture{RGBA: &img.RGBA{W: 1, H: 1, Pix: []byte{128, 255, 255, 255}}})
	defer light.Delete()
	for i := range drawTextureLimit + 3 {
		x := -.94 + float32(i%drawTextureLimit)*.12
		vertices := []Vertex{
			{Pos: V3{x - .05, -.75, .5}, Normal: V3{0, 1, 0}},
			{Pos: V3{x + .05, -.75, .5}, Normal: V3{0, 1, 0}},
			{Pos: V3{x, .75, .5}, Normal: V3{0, 1, 0}},
		}
		mesh := NewMesh(vertices, []uint32{0, 1, 2})
		defer mesh.Delete()
		pixels := []byte{byte(10 + 11*i), 96, 160, 255}
		if i == drawTextureLimit+1 {
			pixels[3] = 0
		}
		texture := Upload(img.Texture{RGBA: &img.RGBA{W: 1, H: 1, Pix: pixels}})
		defer texture.Delete()
		var lightmap *Texture
		if i%5 == 0 {
			lightmap = light
		}
		if lightmap != nil {
			batch.Add(mesh, texture, lightmap, NativeWorldAlphaRef(true), firstPlacement, 1)
		} else {
			batch.Add(mesh, texture, nil, NativeWorldAlphaRef(true), allPlacements, 2)
		}
	}
	begin := func() {
		gl.Viewport(0, 0, width, height)
		gl.DepthMask(true)
		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		s.Begin()
	}
	read := func() []byte {
		pixels := make([]byte, width*height*4)
		gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))
		return pixels
	}
	begin()
	serial := append([]mapDraw(nil), batch.draws...)
	for _, d := range serial {
		if d.lightmap != nil {
			s.drawLightmapped(d.mesh, d.tex, d.lightmap, true)
		} else {
			for _, id := range batch.visible[d.first : d.first+d.count] {
				s.DrawVisibleWorldModel(d.mesh, d.tex, batch.Transforms[id], d.alpha)
			}
		}
	}
	want := read()
	begin()
	ResetDrawCalls()
	batch.Draw(s)
	if calls := DrawCalls(); calls != 2 {
		t.Fatalf("expected two ordered submissions after filling the texture units, got %d", calls)
	}
	if got := read(); !bytes.Equal(got, want) {
		t.Fatal("bound batches differ from direct draws")
	}
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatalf("GL error %#x", code)
	}
}

func TestMapBatchDrawsAfterPackedBufferGrowthGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	const side = 64
	win, err := platform.Open(platform.Config{Width: side, Height: side, Hidden: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err := Init(false); err != nil {
		t.Fatal(err)
	}
	s, err := NewScene3D()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Delete()
	if p := s.mapPrograms(); p.draw == 0 && p.bound == 0 {
		t.Skip("requires indirect map drawing")
	}
	s.View, s.Proj, s.Ambient = Identity(), Identity(), 1
	triangle := []Vertex{
		{Pos: V3{-.6, -.6, .5}, Normal: V3{0, 1, 0}},
		{Pos: V3{.6, -.6, .5}, Normal: V3{0, 1, 0}},
		{Pos: V3{0, .6, .5}, Normal: V3{0, 1, 0}},
	}
	indices := []uint32{0, 1, 2}
	small := NewMesh(triangle, indices)
	defer small.Delete()
	// Both additions cross a capacity threshold in a single draw. A driver
	// can recycle the name of the first vertex buffer on the second resize.
	padded := make([]Vertex, 7000)
	copy(padded, triangle)
	first := NewMesh(padded, indices)
	defer first.Delete()
	second := NewMesh(padded, indices)
	defer second.Delete()

	b := MapBatch{Transforms: []M4{Identity()}}
	defer b.Delete()
	begin := func() {
		gl.Viewport(0, 0, side, side)
		gl.DepthMask(true)
		gl.ClearColor(0, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		s.Begin()
	}
	draw := func(meshes ...*Mesh) {
		at := b.Placements([]uint32{0})
		for _, m := range meshes {
			b.Add(m, nil, nil, -1, at, 1)
		}
		b.Draw(s)
	}
	read := func() []byte {
		pixels := make([]byte, side*side*4)
		gl.ReadPixels(0, 0, side, side, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))
		return pixels
	}

	begin()
	draw(small, small, small, small)
	if b.vao == 0 {
		t.Fatal("warm-up did not use the packed draw")
	}
	originalVertex := b.vertices.id
	gl.BindVertexArray(0)
	begin()
	draw(first, second, first, second)
	got := read()
	if b.vertices.capacity < 1<<20 {
		t.Fatal("packed vertex buffer did not grow twice")
	}
	t.Logf("packed vertex buffer name %d -> %d", originalVertex, b.vertices.id)

	begin()
	s.DrawVisibleWorldModel(second, nil, Identity(), -1)
	if want := read(); want[(side*side/2+side/2)*4] == 0 || !bytes.Equal(got, want) {
		t.Fatal("indirect pixels differ from the direct draw after resizing the vertex buffer")
	}
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatalf("GL error %#x", code)
	}
}
