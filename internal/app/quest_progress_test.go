package app

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/archive"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/data/text"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/scene"
)

func TestQuestProgressMessagesUseConfirmedCounters(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "Monster"), 0755); err != nil {
		t.Fatal(err)
	}
	data := binary.LittleEndian.AppendUint32(nil, 3)
	for _, name := range []string{"", "Prowler", "Scout"} {
		data = binary.LittleEndian.AppendUint32(data, uint32(len(name)))
		data = append(data, name...)
		data = append(data, make([]byte, 31)...)
	}
	if err := os.WriteFile(filepath.Join(root, "Monster", "Monster.SData"), data, 0644); err != nil {
		t.Fatal(err)
	}
	arc, err := archive.OpenDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()

	mobQuest := sdata.Quest{Name: "Patrol", EndType: 2, MobIDs: [2]uint16{1, 2}, MobCounts: [2]uint8{5, 3}}
	pvpQuest := sdata.Quest{Name: "Defense", EndType: 3, PVPKills: 4}
	for _, tc := range []struct {
		name     string
		def      sdata.Quest
		slot     uint8
		value    uint8
		inactive bool
		wantID   int
		want     []string
	}{
		{name: "both mob slots", def: mobQuest, slot: 1, value: 2, wantID: 351, want: []string{"Patrol Prowler 1", "Patrol Scout 2"}},
		{
			name:  "mob display cap",
			def:   sdata.Quest{Name: "Patrol", EndType: 2, MobIDs: [2]uint16{1}, MobCounts: [2]uint8{5}, ResultType: 1},
			value: 8, wantID: 352, want: []string{"Patrol Prowler 5/5"},
		},
		{name: "PvP", def: pvpQuest, value: 2, wantID: 349, want: []string{"Defense faction 2"}},
		{
			name:  "PvP display cap with user selection",
			def:   sdata.Quest{Name: "Defense", EndType: 3, PVPKills: 4, UserSelect: 1},
			value: 6, wantID: 350, want: []string{"Defense faction 4/4"},
		},
		{name: "inactive quest", def: mobQuest, inactive: true},
		{name: "invalid wire slot", def: mobQuest, slot: 3},
		{name: "unused PvP slot", def: pvpQuest, slot: 1},
		{
			name: "missing monster", def: sdata.Quest{Name: "Patrol", EndType: 2, MobIDs: [2]uint16{99}, MobCounts: [2]uint8{5}},
			value: 1, want: []string{"Patrol: objective 1 = 1"},
		},
		{
			name: "item quest", def: sdata.Quest{Name: "Supplies", EndType: 1},
			slot: 2, value: 3, want: []string{"Supplies: objective 3 = 3"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := inputApp()
			a.assets = scene.NewAssets(arc)
			a.questDataLoaded = true
			a.questDefs = []sdata.Quest{tc.def}
			a.msg = text.SysMsg{
				349: "<p> faction <v>", 350: "<p> faction <v>/<v2>",
				351: "<p> <t> <v>", 352: "<p> <t> <v>/<v2>",
			}
			if !tc.inactive {
				a.state.Quests = map[uint16]proto.ActiveQuest{1: {ID: 1, Counters: [3]uint8{1}}}
			}
			packet := proto.NewWriter(proto.QuestCounter)
			packet.U16(1)
			packet.U8(tc.slot)
			packet.U8(tc.value)
			a.applyWorldPacket(packet.Bytes())

			if got := a.state.Chat; len(got) != len(tc.want) {
				t.Fatalf("chat = %+v, want %v", got, tc.want)
			}
			for i, want := range tc.want {
				line := a.state.Chat[i]
				var style uint8
				if tc.wantID != 0 {
					style = 6
				}
				if line.Text != want || line.MessageID != tc.wantID || line.Style != style {
					t.Errorf("chat[%d] = %+v, want %q in style %d with ID %d", i, line, want, style, tc.wantID)
				}
			}
			if !tc.inactive && tc.slot < 3 && a.state.Quests[1].Counters[tc.slot] != tc.value {
				t.Fatal("display formatting changed the confirmed quest counter")
			}
		})
	}
}
