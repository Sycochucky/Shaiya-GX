package proto

import "fmt"

const InspectEquipment Opcode = 0x0307
const InspectSlots = 13
const InspectRange = 3

// InspectPacket is the native request from Game.exe 0x59b6b0.
func InspectPacket(target uint32) []byte {
	w := NewWriter(InspectEquipment)
	w.U32(target)
	return w.Bytes()
}

// DecodeInspectEquipment accepts the native count and 32-byte records, plus
// the optional GX target ID used to correlate replies with the requested player.
func DecodeInspectEquipment(p []byte) (target uint32, items [InspectSlots]InventoryItem, err error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > len(items) || (r.Remaining() != n*32 && r.Remaining() != n*32+4) {
		return 0, items, fmt.Errorf("invalid inspection equipment count")
	}
	for range n {
		slot := r.U8()
		item := InventoryItem{Slot: slot, Type: r.U8(), TypeID: r.U8(), Quality: r.U16(), Count: 1}
		for i := range item.Gems {
			item.Gems[i] = r.U8()
		}
		item.Craft = r.Fixed(21)
		if int(slot) >= len(items) || item.Type == 0 || item.TypeID == 0 || items[slot].Type != 0 {
			return 0, items, fmt.Errorf("invalid inspection slot %d", slot)
		}
		items[slot] = item
	}
	if r.Remaining() == 4 {
		target = r.U32()
		if target == 0 {
			return 0, items, fmt.Errorf("invalid inspection target")
		}
	}
	return target, items, r.Err()
}
