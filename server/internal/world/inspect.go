package world

import (
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

func (s *Session) handleInspectEquipment(p []byte) {
	if s.char == nil || s.awaitingMapLoad || len(p) != 6 {
		return
	}
	id := proto.NewReader(p).U32()
	if id == 0 || id == s.char.ID {
		return
	}
	// Native PacketGetInfo resolves the target inside the requester's zone.
	realm.registryMu.Lock()
	target := realm.registry[s.char.MapID][id]
	if target == nil {
		realm.registryMu.Unlock()
		return
	}
	items, ok := target.inspectEquipment(s.char, s.faction, s.isAdmin)
	realm.registryMu.Unlock()
	if ok {
		s.send(buildInspectEquipment(id, items), "0x0307 inspect equipment")
	}
}

func (p *ConnectedPlayer) inspectEquipment(viewer *db.Character, faction uint8, viewerIsAdmin bool) (items [proto.InspectSlots]db.Item, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.IsAdmin && !viewerIsAdmin || p.faction != faction || p.mapID != viewer.MapID {
		return items, false
	}
	dx, dy, dz := viewer.PosX-p.character.PosX, viewer.PosY-p.character.PosY, viewer.PosZ-p.character.PosZ
	if !(dx*dx+dy*dy+dz*dz <= proto.InspectRange*proto.InspectRange) {
		return items, false
	}
	for _, item := range p.items {
		if item.Bag == 0 && int(item.Slot) < len(items) {
			items[item.Slot] = item
		}
	}
	return items, true
}

func buildInspectEquipment(target uint32, items [proto.InspectSlots]db.Item) []byte {
	w := packet.New(packet.InspectEquipment)
	var count uint8
	for _, item := range items {
		if item.Type != 0 && item.TypeID != 0 {
			count++
		}
	}
	w.U8(count)
	for slot, item := range items {
		if item.Type == 0 || item.TypeID == 0 {
			continue
		}
		w.U8(uint8(slot))
		w.U8(item.Type)
		w.U8(item.TypeID)
		w.U16(item.Quality)
		w.Bytes(item.Gems[:])
		w.FixedString(item.CraftName, 21)
	}
	w.U32(target)
	return w.MustFinish()
}
