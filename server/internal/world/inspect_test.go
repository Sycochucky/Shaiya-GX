package world

import (
	"testing"

	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

func TestInspectOnlyNearbyAlliedEquipment(t *testing.T) {
	const mapID = 60086
	viewer := db.Character{ID: 920001, MapID: mapID, PosX: 20, PosY: 5, PosZ: 20}
	subject := db.Character{ID: 920002, MapID: mapID, PosX: 20, PosY: 8, PosZ: 20}
	item := db.Item{Slot: 5, Type: 1, TypeID: 2, Quality: 123, Gems: [6]uint8{1, 2, 3, 4, 5, 6}, CraftName: "01020304050600000007"}
	target := RegisterPlayer(subject, []db.Item{item, {Bag: 1, Slot: 1, Type: 1, TypeID: 3}, {Slot: 13, Type: 42, TypeID: 1}}, false, FactionLight, nil, nil)
	t.Cleanup(target.Unregister)
	s, packets := gmWireSession(t, viewer, false)
	s.faction = FactionLight
	request := proto.InspectPacket(subject.ID)
	s.dispatch(packet.InspectEquipment, request)
	p := waitWorldPacket(t, packets, proto.InspectEquipment)
	id, got, err := proto.DecodeInspectEquipment(p)
	if want := (proto.InventoryItem{Slot: 5, Type: item.Type, TypeID: item.TypeID, Quality: item.Quality, Gems: item.Gems, Count: 1, Craft: item.CraftName}); err != nil || id != subject.ID || got != ([proto.InspectSlots]proto.InventoryItem{5: want}) {
		t.Fatal("inspection wire lost item data or exposed another bag", id, got, err)
	}
	if id, native, err := proto.DecodeInspectEquipment(p[:len(p)-4]); err != nil || id != 0 || native != got {
		t.Fatal("core response must remain native", native, err)
	}

	gm := RegisterPlayer(db.Character{ID: 920003, MapID: mapID, PosX: 20, PosY: 5, PosZ: 20}, []db.Item{item}, true, FactionLight, nil, nil)
	t.Cleanup(gm.Unregister)
	if _, ok := gm.inspectEquipment(s.char, s.faction, false); ok {
		t.Fatal("a regular character could inspect a GM")
	}
	s.isAdmin = true
	s.dispatch(packet.InspectEquipment, proto.InspectPacket(gm.CharacterID))
	gmReply := waitWorldPacket(t, packets, proto.InspectEquipment)
	if id, _, err := proto.DecodeInspectEquipment(gmReply); err != nil || id != gm.CharacterID {
		t.Fatal("GM inspection did not return the nearby allied GM", id, err)
	}

	for _, tc := range []struct {
		name    string
		mapID   uint16
		y       float32
		faction uint8
	}{
		{"far in height", mapID, 8.01, FactionLight},
		{"other faction", mapID, 5, FactionDark},
		{"other instance", mapID + 1, 5, FactionLight},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target.SetMap(tc.mapID)
			target.mu.Lock()
			target.character.PosY, target.faction = tc.y, tc.faction
			target.mu.Unlock()
			if _, ok := target.inspectEquipment(s.char, s.faction, s.isAdmin); ok {
				t.Fatal("inspection should be refused")
			}
		})
	}
}
