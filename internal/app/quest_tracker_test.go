package app

import (
	"fmt"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func trackerTestApp() *App {
	a := inputApp()
	a.questUI, _ = ui.NewQuestLog(inputTextures{})
	a.questConfirm, _ = ui.NewMessage(inputTextures{})
	a.questTrackerUI, _ = ui.NewQuestTracker(inputTextures{}, nil)
	a.questDataLoaded = true
	a.font = &ui.Font{}
	return a
}

func TestQuestTrackerFollowsOriginalFlow(t *testing.T) {
	a := trackerTestApp()
	a.questDefs = []sdata.Quest{
		{Name: "Defense", EndType: 3, PVPKills: 3, ResultType: 1},
		{Name: "Duel", EndType: 3, PVPKills: 2},
		{Name: "Report", MaxMinutes: 5},
	}
	for id := uint16(1); id <= 3; id++ {
		a.applyWorldPacket(proto.QuestStartPacket(0, id))
	}
	click := func(x, y int) {
		t.Helper()
		a.worldControls(&platform.Input{MouseX: x, MouseY: y, Pressed: true, MouseDown: true}, 1024, 768)
		if r := a.worldControls(&platform.Input{MouseX: x, MouseY: y, Clicked: true}, 1024, 768); r.Clicked {
			t.Fatal("panel click reached the world")
		}
	}
	a.worldControls(&platform.Input{Chars: []rune{'u'}}, 1024, 768)
	jx, jy := a.questUI.Origin(1024, 768)
	click(jx+55, jy+70)
	if !a.questTracker.open || !a.questTracker.expanded[1] {
		t.Fatal("journal row did not open the tracker with that quest toggled")
	}
	click(jx+55, jy+100)
	if a.questTracker.expanded[2] {
		t.Fatal("journal row changed an open tracker")
	}
	a.worldControls(&platform.Input{Chars: []rune{'u'}}, 1024, 768)
	if !a.questTracker.open {
		t.Fatal("closing the journal closed the tracker")
	}
	e := a.questTracker.entries
	if want := fmt.Sprintf("%s : 0/3", ui.Tr(438)); len(e[0].Lines) != 1 || e[0].Lines[0] != want {
		t.Fatalf("PvP line = %q, want %q", e[0].Lines, want)
	}
	if e[2].Time != fmt.Sprintf("%5d%s", 5, ui.Tr(440)) {
		t.Fatalf("time column = %q", e[2].Time)
	}
	x, y := a.questTrackerUI.Origin(1024, 768)
	click(x+20, y+43+2*20) // quest 2 title, under quest 1 and its line
	if want := fmt.Sprintf("%s : 0", ui.Tr(438)); a.questTracker.entries[1].Lines[0] != want {
		t.Fatalf("goal-less PvP line = %q, want %q", a.questTracker.entries[1].Lines, want)
	}
	click(x+218+5, y+1+5)
	if a.questTracker.open || !a.questTracker.expanded[1] {
		t.Fatal("close button must close the window and keep per-quest expansion")
	}
}

func TestQuestTrackerKeepsUnknownMonsterRow(t *testing.T) {
	a := trackerTestApp()
	def := sdata.Quest{EndType: 2, MobIDs: [2]uint16{9, 10}, MobCounts: [2]uint8{1, 2}}
	if lines := a.questTrackerLines(def, proto.ActiveQuest{}); len(lines) != 2 || lines[0] != "" || lines[1] != "" {
		t.Fatalf("unknown monsters must keep their rows: %q", lines)
	}
}
