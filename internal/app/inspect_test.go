package app

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/game"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

func TestInspectKeepsRequestedAllyAndClosesOutside3DRange(t *testing.T) {
	a := inputApp()
	a.inspectUI, _ = ui.NewInspect(inputTextures{}, nil)
	a.state.SelfID = 1
	a.state.Slots[0] = proto.CharacterSlot{ID: 1, Race: 0}
	a.state.Entities[42] = &game.Entity{ID: 42, Kind: game.KindUser, Y: 3, Shape: &proto.UserShape{Race: 1}}
	a.state.Entities[43] = &game.Entity{ID: 43, Kind: game.KindUser, Shape: &proto.UserShape{Race: 2}}
	hasInspect := func(id uint32) bool {
		for _, action := range a.contextActions(id) {
			if action.label == "Inspect" {
				return true
			}
		}
		return false
	}
	if !hasInspect(42) || hasInspect(43) || hasInspect(1) {
		t.Fatal("context did not limit inspection to other allies")
	}
	a.state.Entities[42].IsGM = true
	if hasInspect(42) {
		t.Fatal("regular player can inspect a GM")
	}
	a.state.Self.IsGM = true
	if !hasInspect(42) || hasInspect(43) {
		t.Fatal("GM should inspect allied GMs only")
	}
	a.inspect = inspectView{target: 42, deadline: a.clock + 5}
	reply := func(id uint32) []byte {
		w := proto.NewWriter(proto.InspectEquipment)
		w.U8(0)
		w.U32(id)
		return w.Bytes()
	}
	a.applyWorldPacket(reply(43))
	if a.inspect.open {
		t.Fatal("an unrelated response opened the panel")
	}
	a.applyWorldPacket(reply(42))
	a.state.Target = 43
	a.validateInspect()
	if !a.inspect.open || a.inspect.target != 42 {
		t.Fatal("changing target closed or replaced the inspected ally")
	}
	a.state.Entities[42].Y = 3.01
	a.validateInspect()
	if a.inspect.target != 0 || a.inspect.open {
		t.Fatal("moving outside the 3D range did not close the panel")
	}
}
