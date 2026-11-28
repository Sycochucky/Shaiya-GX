//go:build integration

package app

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/archive"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func TestInspectPanelGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL and the GX inspect artwork")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	const w, h = 1024, 768
	win, err := platform.Open(platform.Config{Width: w, Height: h, Hidden: true, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err := render.Init(true); err != nil {
		t.Fatal(err)
	}
	arc, err := archive.Open(os.Getenv("SHAIYA_DATA"))
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	a, err := New(Config{}, arc)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if a.inspectUI == nil {
		t.Skip("data does not have inspect.png")
	}
	a.state = game.New()
	a.wv = &worldView{}
	a.state.SelfID = 1
	a.state.Slots[0] = proto.CharacterSlot{ID: 1}
	shape := proto.UserShape{ID: 2, Race: 0, Class: 5, Hair: 1, Face: 1, Height: 2, Name: "Ally"}
	a.state.Entities[2] = &game.Entity{ID: 2, Kind: game.KindUser, X: 2, Shape: &shape}
	a.inspect = inspectView{target: 2, deadline: 5}
	reply := proto.NewWriter(proto.InspectEquipment)
	reply.U8(2)
	for _, item := range [][3]uint8{{0, 16, 216}, {5, 1, 2}} {
		reply.Raw(item[:])
		reply.U16(120)
		reply.Raw([]byte{1, 0, 0, 0, 0, 0})
		reply.Fixed("01020304050600000007", 21, false)
	}
	reply.U32(2)
	a.applyWorldPacket(reply.Bytes())
	x, y := a.inspectUI.Origin(w, h)
	mx, my, _ := ui.InspectCell(5)
	a.lastMouseX, a.lastMouseY = x+mx+10, y+my+10
	a.sprites.Begin(w, h)
	a.drawInspect(w, h)
	a.drawSharedTooltip(w, h)
	a.sprites.End()
	if !a.inspect.open || a.inspect.preview == nil || !a.tooltipItem || len(a.tooltip) == 0 {
		t.Fatal("missing rendered inspection preview or inventory tooltip")
	}
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatalf("OpenGL error %#x", code)
	}
	if out := os.Getenv("SHAIYA_TEST_ARTIFACTS"); out != "" {
		if err := render.SavePNG(filepath.Join(out, "inspect.png"), w, h); err != nil {
			t.Fatal(err)
		}
	}
}
