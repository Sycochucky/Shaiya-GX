//go:build integration

package app

import (
	"image/png"
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
)

func TestPvPRankNameplateAndStatusGL(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_RENDER") != "1" {
		t.Skip("requires GL and original loose assets")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	const w, h = 1280, 800
	win, err := platform.Open(platform.Config{Width: w, Height: h, Hidden: true, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	defer win.Close()
	if err = render.Init(true); err != nil {
		t.Fatal(err)
	}
	arc, err := archive.OpenDirectory(os.Getenv("SHAIYA_DATA"))
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	a, err := New(Config{}, arc)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.Preview("world-stream"); err != nil {
		t.Fatal(err)
	}
	for i := 0; !a.SceneReady(); i++ {
		if i > 500 {
			t.Fatal("world did not load")
		}
		a.Draw(w, h)
	}
	f, err := os.Open(filepath.Join(os.Getenv("SHAIYA_DATA"), "interface", "grade2.png"))
	if err != nil {
		t.Fatal(err)
	}
	atlas, err := png.Decode(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if atlas.Bounds().Dx() != 64 || atlas.Bounds().Dy() != 512 {
		t.Fatal("native grade2 atlas must contain 2x16 cells of 32 pixels")
	}
	clear := func() {
		gl.Viewport(0, 0, w, h)
		gl.ClearColor(0, 0, 0, 1)
		gl.DepthFunc(gl.LESS)
		gl.ClearDepth(1)
		gl.DepthMask(true)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	}
	pixels := func() []byte {
		p := make([]byte, w*h*4)
		gl.ReadPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(p))
		return p
	}
	slot, _ := a.selfSlot()
	anchor, _ := a.playerNameplateAnchor(a.state.SelfID, a.wv.self, slot.Race, false, a.state.Self.Vehicle)
	x, y, _, ok := projectLabel(anchor, a.scene3d.View.Mul(a.scene3d.Proj), w, h)
	if !ok {
		t.Fatal("self name anchor off screen")
	}
	slot.Name = "RankedPlayer"
	a.state.Slots[0] = slot
	a.state.Self.IsGM = false
	a.scene3d.Occlusion = nil // Isolate the annotation pass from the earlier terrain frame.
	t.Logf("anchor=%d,%d own=%v name=%q grade=%v", x, y, a.clientSettings().OwnName, slot.Name, a.assets.Texture("interface/grade2.png") != nil)
	// The expected pixels are read independently from the source atlas, so a
	// swapped column, wrong row or incorrect geometry cannot pass this check.
	for _, remote := range []bool{false, true} {
		a.gameMenu.Settings.OwnName = !remote
		for _, c := range []struct {
			kills uint32
			cell  int
		}{{1, 0}, {200000, 15}, {250000, 16}, {1000000, 30}} {
			a.state.Details.Kills = c.kills
			name := slot.Name
			if remote {
				name = "OtherPlayer"
				a.state.Entities[99] = &game.Entity{Kind: game.KindUser, ID: 99, Shape: &proto.UserShape{ID: 99, Race: slot.Race, Name: name}}
				a.wv.actors[99] = &remoteActor{character: a.wv.self, shape: *a.state.Entities[99].Shape}
				update := proto.NewWriter(proto.CharacterPvPKills)
				update.U32(99)
				update.U32(c.kills)
				if _, err := a.state.Apply(update.Bytes()); err != nil {
					t.Fatal(err)
				}
			}
			clear()
			a.sprites.Begin(w, h)
			a.drawPlayerNames(w, h)
			a.sprites.End()
			p := pixels()
			nameFont := a.chatFont()
			left, top := x-nameFont.Width(name)/2+nameFont.Width(name)+1, y-3+(nameFont.Height-a.font.Height)/2-(nameFont.Height-a.font.Height)
			colored := 0
			for dy := 0; dy < 32; dy++ {
				for dx := 0; dx < 32; dx++ {
					r, g, b, alpha := atlas.At(c.cell/16*32+dx, c.cell%16*32+dy).RGBA()
					if alpha != 65535 {
						continue
					}
					at := ((h-1-(top+dy))*w + left + dx) * 4
					for channel, want := range []uint32{r, g, b} {
						if d := int(p[at+channel]) - int(want>>8); d < -3 || d > 3 {
							t.Fatalf("rank cell %d pixel %d,%d channel %d = %d, want %d", c.cell, dx, dy, channel, p[at+channel], want>>8)
						}
					}
					colored++
				}
			}
			if colored < 20 {
				t.Fatal("rank atlas cell is empty")
			}
		}
		delete(a.state.Entities, 99)
		delete(a.wv.actors, 99)
	}
	a.gameMenu.Settings.OwnName = true

	a.state.Details.Kills = 1000
	if got := a.statusState().Title; got != a.msg.Get(5450)+" "+a.msg.Get(5392) {
		t.Fatalf("status title did not use real SysMsg: %q", got)
	}
	for _, expanded := range []bool{false, true} {
		a.statusUI.Expanded = expanded
		draw := func(title string) []byte {
			clear()
			state := a.statusState()
			state.Title = title
			a.sprites.Begin(w, h)
			a.statusUI.Draw(a.sprites, a.font, w, h, state)
			a.sprites.End()
			return pixels()
		}
		before, after := draw(""), draw(a.statusState().Title)
		sx, sy := a.statusUI.Origin(w, h)
		if expanded {
			sx += 132
		}
		changed := 0
		for py := sy + 101; py < sy+101+a.font.Height; py++ {
			for px := sx + 73; px < sx+255; px++ {
				i := ((h-1-py)*w + px) * 4
				if before[i] != after[i] || before[i+1] != after[i+1] || before[i+2] != after[i+2] {
					changed++
				}
			}
		}
		if changed < 20 {
			t.Fatalf("Title not rendered in native field (expanded=%v)", expanded)
		}
	}
	if dir := os.Getenv("SHAIYA_CAPTURE_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := render.SavePNG(filepath.Join(dir, "status-pvp-title.png"), w, h); err != nil {
			t.Fatal(err)
		}
	}
	a.loadNPCs()
	for key := range a.npcDefs {
		if key.Type == 4 {
			a.shop.npcID = 42
			a.state.Entities[42] = &game.Entity{Kind: game.KindNpc, ID: 42, NpcType: 4, TypeID: key.ID}
			break
		}
	}
	if a.shop.npcID == 0 {
		t.Fatal("original PvP manager definition missing")
	}
	slot.Mode, slot.Level = 1, 60
	a.state.Slots[0] = slot
	a.state.Details.Kills, a.state.Details.Deaths = 50, 300
	a.state.PvPRewards = proto.PvPRewards{Known: true, KillStage: 2, DeathStage: 3}
	clear()
	a.sprites.Begin(w, h)
	a.drawPvPRewards(w, h)
	a.sprites.End()
	state := a.pvpRewardState()
	if !state.CanKill || !state.CanDeath || state.Reward == "" || state.Deaths != "300 / 300" {
		t.Fatal("PvP manager did not expose earned native stages", state)
	}
	rx, ry := a.pvpRewardUI.Origin(w, h)
	if state.DeathProgress < .09677418 || state.DeathProgress > .09677421 {
		t.Fatal("native death bar must show claim level 3/31", state.DeathProgress)
	}
	frame := pixels()
	barPixel := func(px, py int) (byte, byte, byte) {
		i := ((h-1-py)*w + px) * 4
		return frame[i], frame[i+1], frame[i+2]
	}
	r, g, _ := barPixel(rx+340, ry+122)
	if r < 100 || int(r) < 2*int(g) {
		t.Fatal("native kill progress must use the red atlas strip", r, g)
	}
	r, _, b := barPixel(rx+296, ry+297)
	if b < 100 || int(b) < int(r)+30 {
		t.Fatal("native death progress must start at x288 and use the violet strip", r, b)
	}
	before := a.state.Details.StatPoint
	a.updatePvPRewards(&platform.Input{MouseX: rx + 340, MouseY: ry + 156, Pressed: true, MouseDown: true}, w, h)
	a.updatePvPRewards(&platform.Input{MouseX: rx + 340, MouseY: ry + 156, Clicked: true}, w, h)
	if !a.pvpRewardMessage.Visible || a.pvpReward.confirm != 1 || a.state.Details.StatPoint != before {
		t.Fatal("PvP reward confirmation missing or predicted points")
	}
	a.updatePvPRewards(&platform.Input{Keys: []platform.Key{platform.KeyEscape}}, w, h)
	if a.pvpRewardMessage.Visible || a.state.Details.StatPoint != before {
		t.Fatal("cancelled PvP reward changed state")
	}
	if dir := os.Getenv("SHAIYA_CAPTURE_DIR"); dir != "" {
		if err := render.SavePNG(filepath.Join(dir, "pvp-rewards.png"), w, h); err != nil {
			t.Fatal(err)
		}
	}
	a.state.Details.Kills, a.state.Details.Deaths = 1000000, 1000000
	a.state.PvPRewards.KillStage = proto.PvPRewardFinishedStage
	a.state.PvPRewards.DeathStage = proto.PvPRewardFinishedStage
	state = a.pvpRewardState()
	if a.font.Width(state.Deaths) > 153 || a.font.Width(state.Gold) > 153 ||
		state.KillStatus == "" || a.font.Width(state.KillStatus) > 200 {
		t.Fatal("terminal reward text does not fit the stock panel", state)
	}
	if dir := os.Getenv("SHAIYA_CAPTURE_DIR"); dir != "" {
		clear()
		a.sprites.Begin(w, h)
		a.drawPvPRewards(w, h)
		a.sprites.End()
		if err := render.SavePNG(filepath.Join(dir, "pvp-rewards-complete.png"), w, h); err != nil {
			t.Fatal(err)
		}
	}
	if code := gl.GetError(); code != gl.NO_ERROR {
		t.Fatalf("rank/status GL error %#x", code)
	}
}
