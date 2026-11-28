//go:build integration

package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/archive"
	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/render/gl"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestStatusAndPotionNativeFieldsGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL and original data")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	win, err := platform.Open(platform.Config{Width: 800, Height: 600, Hidden: true, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err = render.Init(true); err != nil {
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
	a.state = game.New()
	a.state.SelfID = 1
	a.state.Slots[0] = proto.CharacterSlot{ID: 1, Mode: 3, Str: 20, Dex: 21, Rec: 22, Int: 23, Wis: 24, Luc: 25}
	var slots [2]ui.ActionSlot
	for index := range slots {
		def := a.itemsItem(25, uint8(index+4))
		if def == nil {
			t.Fatal("missing potion definition")
		}
		file, uv, ok := ui.ItemIcon(25, def.Icon)
		if !ok {
			t.Fatal("missing potion icon")
		}
		slots[index] = ui.ActionSlot{Icon: a.assets.Texture("interface/" + file), UV: uv, Count: 255}
	}
	out := os.Getenv("SHAIYA_TEST_ARTIFACTS")
	if out == "" {
		out = t.TempDir()
	}
	for index, value := range []uint8{0, 6, 6, 6} {
		view := a.statusState()
		view.Element = elementName(value, a.msg)
		if index == 2 {
			a.statusUI.Expanded = true
			a.statusUI.Draft = proto.StatPoints{12345, 2, 3, 4, 5, 6}
			view.Title, view.Class = "Liberator of the Alliance", "Battle Priest"
			view.AttackSpeed, view.MoveSpeed = "Extremely Fast", "Very Fast"
			view.Points = 65535
			view.Base = proto.StatPoints{65535, 9, 100, 4, 19, 235}
			view.Details = proto.Details{Exp: 2800000000, NextExp: 3100000000}
			view.Details.MaxHP, view.Details.MaxMP, view.Details.MaxSP = 4294967295, 25000, 15000
			view.HP, view.MP, view.SP = 2147483648, 18000, 9500
			view.Details.Kills, view.Details.Deaths, view.Details.Wins, view.Details.Losses = 600000, 5000, 250, 100
			view.Stats = proto.AdditionalStats{
				Str: 70000, Rec: 100, Int: 4, Wis: 19, Dex: 9, Luc: 235,
				MinAttack: 12345, MaxAttack: 987654, MinMagic: 500, MaxMagic: 1000, Defense: 15500, Resistance: 29000,
			}
			view.Known = true
		}
		if index == 3 {
			a.statusUI.Clear()
			a.statusUI.Expanded = true
			a.state.Details.StatPoint = 2
			packet := proto.NewWriter(proto.CharacterAdditionalStats)
			for _, value := range [12]uint32{28, 22, 23, 24, 21, 25, 100, 110, 60, 65, 40, 30} {
				packet.U32(value)
			}
			if _, err := a.state.Apply(packet.Bytes()); err != nil {
				t.Fatal(err)
			}
			view = a.statusState()
			x, y := a.statusUI.Origin(800, 600)
			a.statusUI.Update(&platform.Input{MouseX: x + 86, MouseY: y + 129, Pressed: true, MouseDown: true}, 800, 600, view, 0)
			a.statusUI.Update(&platform.Input{MouseX: x + 86, MouseY: y + 129, Clicked: true}, 800, 600, view, 0)
			if !view.Known || view.Stats.Str != 28 || view.Base[0] != 20 || a.statusUI.Draft != (proto.StatPoints{1}) {
				t.Fatalf("confirmed stats and base draft did not stay independent: %+v, draft %v", view, a.statusUI.Draft)
			}
			t.Log("rendered confirmed STR 28 alongside base STR 20 with one draft point")
		}
		gl.Viewport(0, 0, 800, 600)
		gl.ClearColor(.15, .15, .15, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		a.sprites.Begin(800, 600)
		a.statusUI.Draw(a.sprites, a.font, 800, 600, view)
		ui.DrawQuickPotions(a.sprites, a.font, a.assets.Texture("interface/QuickPotion.png"), a.white, 80, 350, slots)
		a.sprites.End()
		if code := gl.GetError(); code != gl.NO_ERROR {
			t.Fatal(code)
		}
		if err := render.SavePNG(filepath.Join(out, []string{"status-neutral-potions.png", "status-element-potions.png", "status-expanded-values.png", "status-confirmed-total-and-draft.png"}[index]), 800, 600); err != nil {
			t.Fatal(err)
		}
	}
	a.status.open = true
	x, y := a.statusUI.Origin(800, 600)
	a.lastMouseX, a.lastMouseY = x+60, y+133
	a.sprites.Begin(800, 600)
	a.drawStatus(800, 600)
	want := a.msg.Get(5283)
	if want == "" || len(a.tooltip) != 1 || a.tooltip[0].text != want || a.tooltipItem {
		t.Fatalf("hover did not queue the original STR text in the shared tooltip: %v", a.tooltip)
	}
	a.drawSharedTooltip(800, 600)
	a.sprites.End()
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatal(code)
	}
}
