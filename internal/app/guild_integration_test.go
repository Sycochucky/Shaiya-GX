//go:build integration

package app

import (
	"os"
	"testing"
	"time"

	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/render"
	"github.com/Spelunkern/Shaiya-GX/internal/testworld"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func TestGuildDirectoryAgainstRapture(t *testing.T) {
	st, _, _, wait := testworld.Enter(t)
	wait(func() bool { return st.Guild.DirectoryReady })
	for id, guild := range st.Guild.Directory {
		if id == 0 || guild.ID != id || guild.Name == "" {
			t.Fatalf("invalid guild directory entry: %+v", guild)
		}
	}
	t.Logf("received complete Rapture guild directory (%d entries)", len(st.Guild.Directory))
}

func TestGuildApplicationAgainstRapture(t *testing.T) {
	name := os.Getenv("SHAIYA_TEST_GUILD")
	if name == "" || os.Getenv("SHAIYA_TEST_PARTNER") == "" {
		t.Skip("requires disposable guild fixture")
	}
	a, wa, _, waitA := testworld.Enter(t)
	b, wb, _, waitB := testworld.EnterAs(t, os.Getenv("SHAIYA_TEST_PARTNER"), os.Getenv("SHAIYA_TEST_PARTNER_PASSWORD"))
	appA, appB := inputApp(), inputApp()
	appA.state, appA.world = a, wa
	appB.state, appB.world = b, wb
	appA.guildUI, _ = ui.NewGuildPanel(inputTextures{})
	appB.guildDirectoryUI, _ = ui.NewGuildDirectory(inputTextures{})
	waitA(func() bool { return a.Guild.Rank == 1 && a.Guild.DirectoryReady })
	waitB(func() bool { return b.Guild.DirectoryReady })
	guildID := a.Guild.ID
	if b.Guild.ID != 0 || b.Guild.Directory[guildID].Name != name {
		t.Fatal("application fixture unavailable")
	}
	poll := func(app *App, cond func() bool) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for !cond() {
			app.pumpWorld()
			if app.state == nil || time.Now().After(deadline) {
				t.Fatal("guild application timed out")
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	appA.worldControls(&platform.Input{Chars: []rune{'g'}}, 1024, 768)
	x, y := appA.guildUI.Origin(1024, 768)
	click := func(cx, cy int) {
		appA.worldControls(&platform.Input{MouseX: x + cx, MouseY: y + cy, Pressed: true, MouseDown: true}, 1024, 768)
		appA.worldControls(&platform.Input{MouseX: x + cx, MouseY: y + cy, Clicked: true}, 1024, 768)
	}
	click(240, 50)
	if !appA.guild.waiting {
		t.Fatal("waiting list not opened")
	}
	var npcID uint32
	waitB(func() bool {
		for id, e := range b.Entities {
			if e.Kind == game.KindNpc && e.NpcType == 11 && e.TypeID == 1 {
				npcID = id
				return true
			}
		}
		return false
	})
	manager := b.Entities[npcID]
	appB.wv.pos = render.V3{manager.X, manager.Y, manager.Z}
	appB.openNPC(npcID)
	if appB.shop.npcID != npcID {
		t.Fatal("did not open the actual guild manager")
	}
	dx, dy := appB.guildDirectoryUI.Origin(1024, 768)
	appB.worldControls(&platform.Input{}, 1024, 768)
	directoryRow := -1
	for i, entry := range appB.guild.directory {
		if entry.ID == guildID {
			directoryRow = i
			break
		}
	}
	if directoryRow < 0 || directoryRow >= ui.GuildDirectoryRows {
		t.Fatal("fixture outside visible guild directory")
	}
	clickDirectory := func(cx, cy int) {
		appB.worldControls(&platform.Input{MouseX: dx + cx, MouseY: dy + cy, Pressed: true, MouseDown: true}, 1024, 768)
		appB.worldControls(&platform.Input{MouseX: dx + cx, MouseY: dy + cy, Clicked: true}, 1024, 768)
	}
	clickDirectory(275, 95+directoryRow*18)
	if appB.guild.directorySelected != guildID {
		t.Fatal("guild directory selection failed")
	}
	clickDirectory(130, 290)
	poll(appB, func() bool { return b.Guild.ApplySerial == 1 })
	poll(appA, func() bool { _, ok := a.Guild.Applicants[b.SelfID]; return ok })
	if b.Guild.ApplyResult != 1 || b.Guild.ID != 0 {
		t.Fatal("application changed membership early")
	}
	click(105, 116)
	if appA.guild.selected != b.SelfID {
		t.Fatal("applicant not selectable")
	}
	click(115, 470) // Reject.
	poll(appA, func() bool { _, ok := a.Guild.Applicants[b.SelfID]; return !ok })
	poll(appB, func() bool { return b.Guild.JoinSerial == 1 })
	if b.Guild.JoinResult != 0 || b.Guild.ID != 0 {
		t.Fatal("rejected application changed membership")
	}
	appB.guildCommand("/guildapply " + name)
	poll(appB, func() bool { return b.Guild.ApplySerial == 2 })
	poll(appA, func() bool { _, ok := a.Guild.Applicants[b.SelfID]; return ok })
	appA.clock = 1
	click(105, 116)
	click(45, 470) // Accept.
	if b.Guild.ID != 0 {
		t.Fatal("application accepted before server confirmation")
	}
	poll(appA, func() bool { _, ok := a.Guild.Members[b.SelfID]; return ok })
	poll(appB, func() bool { return b.Guild.ID == guildID && b.Guild.Rank == 9 })
	if _, ok := a.Guild.Applicants[b.SelfID]; ok {
		t.Fatal("accepted applicant still waiting")
	}
	t.Log("application, waiting list, rejection and acceptance confirmed")
}

func TestGuildDisbandSmallAgainstRapture(t *testing.T) {
	if os.Getenv("SHAIYA_TEST_GUILD") == "" || os.Getenv("SHAIYA_TEST_PARTNER") == "" {
		t.Skip("requires disposable guild")
	}
	a, wa, _, waitA := testworld.Enter(t)
	b, wb, _, waitB := testworld.EnterAs(t, os.Getenv("SHAIYA_TEST_PARTNER"), os.Getenv("SHAIYA_TEST_PARTNER_PASSWORD"))
	appA, appB := inputApp(), inputApp()
	appA.state, appA.world = a, wa
	appB.state, appB.world = b, wb
	appA.guildDirectoryUI, _ = ui.NewGuildDirectory(inputTextures{})
	appA.guildConfirm, _ = ui.NewMessage(inputTextures{})
	appA.guildVoteMessage, _ = ui.NewMessage(inputTextures{})
	waitA(func() bool {
		return a.Guild.ID != 0 && a.Guild.Rank == 1 && a.Guild.DirectoryReady && len(a.Guild.Members) == 2
	})
	waitB(func() bool { return b.Guild.ID == a.Guild.ID && b.Guild.DirectoryReady })
	id := a.Guild.ID
	var manager *game.Entity
	waitA(func() bool {
		for _, e := range a.Entities {
			if e.Kind == game.KindNpc && e.NpcType == 11 && e.TypeID == 1 {
				manager = e
				return true
			}
		}
		return false
	})
	appA.wv.pos = render.V3{manager.X, manager.Y, manager.Z}
	appA.openNPC(manager.ID)
	if appA.shop.npcID != manager.ID {
		t.Fatal("guild manager not opened")
	}
	x, y := appA.guildDirectoryUI.Origin(1024, 768)
	appA.worldControls(&platform.Input{MouseX: x + 130, MouseY: y + 325, Pressed: true, MouseDown: true}, 1024, 768)
	appA.worldControls(&platform.Input{MouseX: x + 130, MouseY: y + 325, Clicked: true}, 1024, 768)
	if !appA.guildConfirm.Visible {
		t.Fatal("disband confirmation missing")
	}
	appA.updateGuildConfirm(&platform.Input{Keys: []platform.Key{platform.KeyEnter}}, 1024, 768)
	if a.Guild.ID != id {
		t.Fatal("disband applied before confirmation")
	}
	deadline := time.Now().Add(10 * time.Second)
	for a.Guild.ID != 0 || b.Guild.ID != 0 {
		appA.pumpWorld()
		appB.pumpWorld()
		if appA.state == nil || appB.state == nil || time.Now().After(deadline) {
			t.Fatal("small guild disband timed out")
		}
		time.Sleep(10 * time.Millisecond)
	}
	appA.updateGuildVote(&platform.Input{}, 1024, 768)
	if appA.guild.lifecyclePending || a.Guild.DisbandVoteSerial != 0 || b.Guild.DisbandVoteSerial != 0 {
		t.Fatal("small guild disband should not require a vote")
	}
	t.Log("two-member guild disband confirmed without vote")
}
