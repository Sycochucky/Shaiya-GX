//go:build integration

package render

import (
	"bytes"
	"os"
	"runtime"
	"testing"
	"unsafe"

	img "github.com/Spelunkern/Shaiya-GX/internal/data/image"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
	"github.com/go-gl/glfw/v3.4/glfw"
)

func TestGXOpenGLCommandsAndDrawPathsGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	window, err := platform.Open(platform.Config{Width: 64, Height: 64, Hidden: true, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	defer window.Close()
	resolve := func(name string) unsafe.Pointer {
		switch name {
		case "glGetnTexImage", "glGetnCompressedTexImage", "glGetnUniformdv":
			t.Errorf("looked up unused command %s", name)
			return nil
		default:
			return glfw.GetProcAddress(name)
		}
	}
	if err := gl.InitWithProcAddrFunc(resolve); err != nil {
		t.Fatal(err)
	}
	if err := gl.InitWithProcAddrFunc(func(name string) unsafe.Pointer {
		if name == "glDrawElements" {
			return nil
		}
		return resolve(name)
	}); err == nil || err.Error() != "glDrawElements" {
		t.Fatalf("missing draw command returned %v", err)
	}
	// A driver can advertise the optional extension while failing to
	// resolve one of its commands for the current context.
	missingNonResident := false
	if err := initDevice(InitOptions{}, func(name string) unsafe.Pointer {
		if name == "glMakeTextureHandleNonResidentARB" {
			missingNonResident = true
			return nil
		}
		return resolve(name)
	}); err != nil {
		t.Fatal(err)
	}
	if !missingNonResident || skinIndirectSupported {
		t.Fatal("missing bindless command did not select the ordinary drawing path")
	}
	if err := initDevice(InitOptions{DisableBindless: true}, func(name string) unsafe.Pointer {
		if name == "glGetTextureHandleARB" {
			t.Error("bindless command was resolved despite explicit opt-out")
		}
		return resolve(name)
	}); err != nil {
		t.Fatal(err)
	}
	if skinIndirectSupported {
		t.Fatal("bindless opt-out enabled the indirect path")
	}
	if err := Init(true); err != nil {
		t.Fatal(err)
	}
	oldIndirect, oldDrawParameters := skinIndirectSupported, drawParametersSupported
	oldSceneStorage, oldSkinCrowd := sceneStorageSupported, skinCrowdSupported
	oldMapIndirect := mapIndirectSupported
	oldSkinMultidraw := skinMultidrawSupported
	defer func() {
		skinIndirectSupported, drawParametersSupported = oldIndirect, oldDrawParameters
		sceneStorageSupported, skinCrowdSupported = oldSceneStorage, oldSkinCrowd
		mapIndirectSupported = oldMapIndirect
		skinMultidrawSupported = oldSkinMultidraw
	}()

	configureDrawPaths(true, true, 8, 8)
	if !skinIndirectSupported || !drawParametersSupported || !sceneStorageSupported || !skinCrowdSupported {
		t.Fatal("full driver capabilities did not enable accelerated paths")
	}
	configureDrawPaths(false, true, 8, 8)
	if skinIndirectSupported || !drawParametersSupported || !mapIndirectSupported || !skinMultidrawSupported {
		t.Fatal("missing bindless support changed independent draw-parameter support")
	}

	if err := InitWithOptions(InitOptions{DisableBindless: true, DisableDrawParameters: true}); err != nil {
		t.Fatal(err)
	}
	scene, err := NewScene3D()
	if err != nil {
		t.Fatal(err)
	}
	defer scene.Delete()
	skin, err := NewSkinner()
	if err != nil {
		t.Fatal(err)
	}
	defer skin.Delete()
	scene.effectPipeline.init()
	if !skin.CrowdInstancing() || skin.indirect != nil || drawParametersSupported {
		t.Fatal("extension override did not preserve ordinary crowd instancing")
	}
	if p := scene.mapPrograms(); p.draw != 0 || p.bound == 0 || !p.materialAttribute {
		t.Fatal("extension override did not select core map batching")
	}
	if p := &scene.effectPipeline.indirect; p.program == 0 || !p.bound || !p.materialAttribute {
		t.Fatal("extension override did not select core effect batching")
	}

	configureDrawPaths(false, false, 0, 0)
	serial, err := NewScene3D()
	if err != nil {
		t.Fatal(err)
	}
	defer serial.Delete()
	uniform, err := NewSkinner()
	if err != nil {
		t.Fatal(err)
	}
	defer uniform.Delete()
	if uniform.CrowdInstancing() || serial.uClothBase != -1 || serial.uEffectPoseOffset != -1 || serial.uInstanceColorOffset != -1 {
		t.Fatal("ordinary shaders still use vertex storage")
	}

	vertices := []Vertex{{Pos: V3{-.2, -.2, .5}}, {Pos: V3{.2, -.2, .5}}, {Pos: V3{0, .2, .5}}}
	mesh := NewMesh(vertices, []uint32{0, 1, 2})
	defer mesh.Delete()
	tex := Upload(img.Texture{RGBA: &img.RGBA{W: 1, H: 1, Pix: []byte{255, 255, 255, 255}}})
	defer tex.Delete()
	worlds := []M4{Identity(), Identity()}
	worlds[0][12], worlds[1][12] = -.4, .4
	colors := [][4]float32{{1, .2, 0, 1}, {0, .3, 1, 1}}
	frame := func(s *Scene3D, storage bool) []byte {
		sceneStorageSupported = storage
		gl.Viewport(0, 0, 64, 64)
		gl.ClearColor(0, 0, 0, 1)
		gl.DepthMask(true)
		gl.ClearDepth(1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		s.View, s.Proj = Identity(), Identity()
		s.Begin()
		s.DrawGroundItems(mesh, tex, worlds, colors)
		worlds[0][13], worlds[1][13] = -.55, -.55
		s.BeginEffects()
		s.DrawEffectBatch(mesh, tex, worlds, colors, 4, 5)
		s.EndEffects()
		worlds[0][13], worlds[1][13] = 0, 0
		pixels := make([]byte, 64*64*4)
		gl.ReadPixels(0, 0, 64, 64, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
		return pixels
	}
	batched, direct := frame(scene, true), frame(serial, false)
	if !bytes.Equal(batched, direct) || !bytes.Contains(direct, []byte{255, 51, 0, 255}) {
		t.Fatal("ordinary draws changed item or effect pixels")
	}
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatalf("GL error %#x", code)
	}
}
