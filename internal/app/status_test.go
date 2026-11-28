package app

import (
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/data/text"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
)

func TestStatusHotkeyFooterAndPointerOwnership(t *testing.T) {
	a := inputApp()
	a.inventory.open = true
	a.worldControls(&platform.Input{Chars: []rune{'T'}}, 1024, 768)
	if !a.status.open || !a.inventory.open {
		t.Fatal("native Status hotkey closed inventory")
	}
	x, y := a.statusUI.Origin(1024, 768)
	a.worldControls(&platform.Input{MouseX: x + 2, MouseY: y + 157, Pressed: true, MouseDown: true, Down: "W"}, 1024, 768)
	controls := a.worldControls(&platform.Input{MouseX: x + 2, MouseY: y + 157, Clicked: true, Down: "W"}, 1024, 768)
	if !a.statusUI.Expanded || controls.Clicked || controls.Down != "W" {
		t.Fatal("status tab leaked a click or swallowed held movement")
	}
	controls = a.worldControls(&platform.Input{Keys: []platform.Key{platform.KeyEscape}}, 1024, 768)
	if a.status.open || len(controls.Keys) != 0 {
		t.Fatal("status close reached world logout")
	}
	a.worldControls(&platform.Input{MouseX: 682, MouseY: 715, Pressed: true, MouseDown: true}, 1024, 768)
	a.worldControls(&platform.Input{MouseX: 682, MouseY: 715, Clicked: true}, 1024, 768)
	if !a.status.open {
		t.Fatal("status footer button did not open")
	}
	a.toggleSkills(false)
	if !a.status.open || !a.skills.open {
		t.Fatal("opening skills closed status")
	}
}

func TestStatusUsesOnlyConfirmedNativeSpeedMessages(t *testing.T) {
	a := inputApp()
	a.msg = map[int]string{5271: "none", 5276: "normal", 5278: "fast", 5280: "extremely fast", 5281: "can't attack", 5282: "can't move"}
	if s := a.statusState(); s.AttackSpeed != "-" || s.MoveSpeed != "-" {
		t.Fatal("unconfirmed speed inferred", s)
	}
	a.state.Self.Speed = &proto.CharacterSpeed{Attack: 9, Move: 3}
	if s := a.statusState(); s.AttackSpeed != "extremely fast" || s.MoveSpeed != "fast" {
		t.Fatal("native confirmed labels", s)
	}
	a.state.Self.Speed = &proto.CharacterSpeed{Attack: 255, Move: 255}
	if s := a.statusState(); s.AttackSpeed != "can't attack" || s.MoveSpeed != "can't move" {
		t.Fatal("native blocked labels", s)
	}
}

func TestStatusTitleFollowsConfirmedKillUpdates(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 100
	a.msg = map[int]string{5450: "Blue", 5451: "Yellow", 5389: "localized first", 5392: "localized fourth", 5404: "localized final"}
	for _, c := range []struct {
		op    proto.Opcode
		kills uint32
		want  string
	}{{proto.UserKillCountUpdate, 1000, "Blue localized fourth"}, {proto.CharacterPvPKills, 200000, "Yellow localized first"}, {proto.UserKillCountUpdate, 1000000, "localized final"}, {proto.UserKillCountUpdate, 0, ""}} {
		w := proto.NewWriter(c.op)
		if c.op == proto.UserKillCountUpdate {
			w.U8(0)
		} else {
			w.U32(100)
		}
		w.U32(c.kills)
		if _, err := a.state.Apply(w.Bytes()); err != nil {
			t.Fatal(err)
		}
		if got := a.statusState().Title; got != c.want {
			t.Fatalf("kills %d title %q, want %q", c.kills, got, c.want)
		}
	}
}

func TestPvPRewardShowsNextDeathThresholdAndCurrentGuide(t *testing.T) {
	a := inputApp()
	a.msg = text.SysMsg{5348: "Stale legacy guide", 5355: "gold"}
	a.state.PvPRewards = proto.PvPRewards{Known: true, KillStage: 3, DeathStage: 6}
	a.state.Details.Kills, a.state.Details.Deaths = 300, 9999
	state := a.pvpRewardState()
	if !state.CanKill || state.CanDeath || state.Kills != "300 / 300" || state.Deaths != "9999 / 10000" || state.Gold != "10,000 gold" {
		t.Fatal("the death row must show its own claim threshold and payout", state)
	}
	if state.Description != text.Defaults[12375] || state.Description == "" {
		t.Fatal("the panel must use the stock GX reward guide", state.Description)
	}
	a.state.Details.Deaths++
	if state := a.pvpRewardState(); !state.CanDeath || state.Deaths != "10000 / 10000" {
		t.Fatal("the claim must become available at the displayed threshold", state)
	}
}

func TestPvPFinalRewardStageRemainsVisibleAndCannotReplay(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 100
	a.state.Slots[0] = proto.CharacterSlot{ID: 100, Mode: 2, Level: 40}
	a.msg = map[int]string{5404: "final title", 5353: "<v>:<v2>", 5354: "level", 5355: "gold", 12376: "completed"}
	a.state.Details.Kills, a.state.Details.Deaths = 1000000, 1000000
	a.state.PvPRewards = proto.PvPRewards{Known: true, KillStage: 31, DeathStage: 31}
	if state := a.pvpRewardState(); !state.CanKill || !state.CanDeath || state.Reward != "31:30" || state.KillStatus != "" || state.DeathStatus != "" {
		t.Fatal("rank31 must offer unified points", state)
	}
	for _, op := range []proto.Opcode{proto.PvPRewardKill, proto.PvPRewardDeath} {
		w := proto.NewWriter(op)
		w.U8(1)
		if op == proto.PvPRewardKill {
			w.U8(26)
		} else {
			w.U32(4000000)
		}
		if _, err := a.state.Apply(w.Bytes()); err != nil {
			t.Fatal(err)
		}
		if op == proto.PvPRewardKill {
			state := a.pvpRewardState()
			if state.CanKill || !state.CanDeath || state.KillStatus != "completed" || state.DeathStatus != "" {
				t.Fatal("finishing kills must leave the death track claimable", state)
			}
		}
	}
	if a.state.PvPRewards.KillStage != 32 || a.state.PvPRewards.DeathStage != 32 {
		t.Fatal("final paid stage lost completion marker", a.state.PvPRewards)
	}
	s := a.pvpRewardState()
	if s.CanKill || s.CanDeath || s.Reward != "31:30" || s.Kills != "1000000 / 1000000" || s.Deaths != "1000000 / 1000000" || s.Gold != "4,000,000 gold" || s.DeathLevel != "31 level / 31 level" || s.KillStatus != "completed" || s.DeathStatus != "completed" || s.KillProgress != 1 || s.DeathProgress != 1 {
		t.Fatal("paid final reward must remain visible without a stage32 preview or replay", s)
	}
	if a.statusState().Title != "final title" {
		t.Fatal("one million kills lost final SysMsg title")
	}
}

func TestMountedStatusShowsAbsoluteSpeed(t *testing.T) {
	a := inputApp()
	a.msg = map[int]string{5279: "very fast", 5280: "extremely fast", 5276: "normal"}
	a.state.Self.Speed = &proto.CharacterSpeed{Move: 2}
	a.state.Self.Vehicle = proto.VehicleState{State: proto.VehicleDriver, Speed: 5}
	if got := a.statusState().MoveSpeed; got != "extremely fast" {
		t.Fatal(got)
	}
	a.state.Self.Vehicle.Speed = 4
	if got := a.statusState().MoveSpeed; got != "very fast" {
		t.Fatal(got)
	}
	a.state.Self.Vehicle.State = proto.VehicleOff
	if got := a.statusState().MoveSpeed; got != "normal" {
		t.Fatal(got)
	}
}

func TestStatusHidesModeAndUsesUpperArmorElement(t *testing.T) {
	a := inputApp()
	a.state.SelfID = 100
	a.state.Slots[0] = proto.CharacterSlot{ID: 100, Mode: 3}
	a.msg = map[int]string{473: "ULTIMATE", 5270: "Sin elemento", 5204: "Fuego", 5205: "Agua"}
	if value := a.statusState(); value.Element != "Sin elemento" {
		t.Fatal(value)
	}
	a.items = map[sdata.ItemKey]*sdata.Item{
		{Type: 17, TypeID: 1}: {Type: 17, TypeID: 1, Attribute: 0},
		{Type: 16, TypeID: 1}: {Type: 16, TypeID: 1, Attribute: 1},
		{Type: 30, TypeID: 1}: {Type: 30, TypeID: 1, Attribute: 6},
	}
	a.state.Inventory[[2]uint8{0, 0}] = proto.InventoryItem{Type: 16, TypeID: 1}
	if a.statusState().Element != "Sin elemento" {
		t.Fatal("helmet supplied armor element")
	}
	a.state.Inventory[[2]uint8{0, 1}] = proto.InventoryItem{Type: 17, TypeID: 1, Gems: [6]uint8{1}}
	if a.statusState().Element != "Agua II" {
		t.Fatal(a.statusState().Element)
	}
	a.items[sdata.ItemKey{Type: 17, TypeID: 1}].Attribute = 1
	if a.statusState().Element != "Fuego I" {
		t.Fatal("intrinsic priority lost")
	}
	delete(a.state.Inventory, [2]uint8{0, 1})
	if a.statusState().Element != "Sin elemento" {
		t.Fatal("removed armor left stale element")
	}
}
