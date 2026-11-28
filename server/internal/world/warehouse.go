package world

import (
	"encoding/binary"
	"github.com/Spelunkern/Shaiya-GX/internal/proto"
	"math"

	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
)

// The account warehouse (PS_GameData.dbo.UserStoredItems) is bag 100 in
// the client's move packets.
const (
	WarehouseBag   = 100
	WarehouseSlots = 240 // CUser::IsBank: slots below 0xF0
)

func findStored(items []db.Item, slot uint8) int {
	for i := range items {
		if items[i].Slot == slot {
			return i
		}
	}
	return -1
}

func buildWarehouseItemList(items []db.Item) []byte {
	w := packet.New(packet.WarehouseItemList)
	w.U32(0) // warehouse gold is not used
	w.U8(uint8(len(items)))
	for i := range items {
		it := &items[i]
		w.U8(it.Slot)
		w.U8(it.Type)
		w.U8(it.TypeID)
		w.U16(it.Quality)
		w.Bytes(it.Gems[:])
		w.Count(it.Count)
		w.FixedString(it.CraftName, 21)
	}
	return w.MustFinish()
}

// withMoney appends the character's money to a withdraw 0x0204 and
// rewrites the frame length.
func withMoney(frame []byte, money uint32) []byte {
	frame = binary.LittleEndian.AppendUint32(frame, money)
	binary.LittleEndian.PutUint16(frame, uint16(len(frame)))
	return frame
}

func copyPtr(it *db.Item) *db.Item {
	if it == nil {
		return nil
	}
	c := *it
	return &c
}

func (s *Session) handleWarehouseMove(srcBag, srcSlot, destBag, destSlot uint8) {
	if (srcBag == WarehouseBag && srcSlot >= WarehouseSlots) || (destBag == WarehouseBag && destSlot >= WarehouseSlots) {
		return
	}
	if s.char == nil {
		return
	}
	before := *s.char
	after := before
	plan := warehousePlan{char: &after, items: append([]db.Item(nil), s.items...), stored: append([]db.Item(nil), s.stored...)}
	var frame []byte
	switch {
	case srcBag == WarehouseBag && destBag == WarehouseBag:
		frame = plan.warehouseRearrange(srcSlot, destSlot)
	case srcBag == WarehouseBag:
		if !isValidItemPosition(destBag, destSlot) || destBag == 0 {
			return
		}
		frame = withMoney(plan.warehouseWithdraw(srcSlot, destBag, destSlot), plan.char.Money)
	default:
		if !isValidItemPosition(srcBag, srcSlot) || srcBag == 0 {
			return
		}
		frame = plan.warehouseDeposit(srcBag, srcSlot, destSlot)
	}
	if frame != nil {
		change := db.CharacterChange{Before: before, After: after, ItemsBefore: s.items, ItemsAfter: plan.items}
		if err := db.CommitWarehouse(s.accountID, change, s.stored, plan.stored); err != nil {
			s.logErr("warehouse transaction", err)
			return
		}
		*s.char = after
		s.items, s.stored = plan.items, plan.stored
		s.syncPlayer()
		s.send(frame, "0x0204 warehouse item move")
	}
}

// warehousePlan owns private copies; planning cannot write storage or sockets.
type warehousePlan struct {
	char          *db.Character
	items, stored []db.Item
}

func (s *warehousePlan) warehouseRearrange(src, dest uint8) []byte {
	si, di := findStored(s.stored, src), findStored(s.stored, dest)
	if si < 0 {
		var d *db.Item
		if di >= 0 {
			d = &s.stored[di]
		}
		return buildBagToBag(WarehouseBag, src, nil, WarehouseBag, dest, d)
	}
	if src == dest {
		return buildBagToBag(WarehouseBag, src, &s.stored[si], WarehouseBag, dest, &s.stored[si])
	}
	if di >= 0 {
		a, b := &s.stored[si], &s.stored[di]
		if ms := maxStackOf(a.Type, a.TypeID); itemsMatch(a, b) && ms > 1 && b.Count < ms {
			moved := min(a.Count, ms-b.Count)
			b.Count += moved
			a.Count -= moved

			bCopy := *b
			var aAfter *db.Item
			if a.Count == 0 {
				s.stored = append(s.stored[:si], s.stored[si+1:]...)
			} else {
				aAfter = copyPtr(a)
			}
			return buildBagToBag(WarehouseBag, src, aAfter, WarehouseBag, dest, &bCopy)
		}
		a.Slot, b.Slot = dest, src

		return buildBagToBag(WarehouseBag, src, b, WarehouseBag, dest, a)
	}
	s.stored[si].Slot = dest

	return buildBagToBag(WarehouseBag, src, nil, WarehouseBag, dest, &s.stored[si])
}

func (s *warehousePlan) warehouseDeposit(srcBag, srcSlot, dest uint8) []byte {
	si, di := findItem(s.items, srcBag, srcSlot), findStored(s.stored, dest)
	if si < 0 {
		var d *db.Item
		if di >= 0 {
			d = &s.stored[di]
		}
		return buildBagToBag(srcBag, srcSlot, nil, WarehouseBag, dest, d)
	}
	src := s.items[si]
	if realType(src.Type) == 27 {
		// CUser::ItemBagToBank: quest items stay in the bags.
		var d *db.Item
		if di >= 0 {
			d = &s.stored[di]
		}
		return buildBagToBag(srcBag, srcSlot, &src, WarehouseBag, dest, d)
	}
	if di >= 0 {
		b := &s.stored[di]
		if ms := maxStackOf(src.Type, src.TypeID); itemsMatch(&src, b) && ms > 1 && b.Count < ms {
			moved := min(src.Count, ms-b.Count)
			b.Count += moved

			s.items[si].Count -= moved

			var after *db.Item
			if s.items[si].Count == 0 {
				s.items = append(s.items[:si], s.items[si+1:]...)
			} else {
				after = copyPtr(&s.items[si])
			}
			return buildBagToBag(srcBag, srcSlot, after, WarehouseBag, dest, b)
		}
		return buildBagToBag(srcBag, srcSlot, &src, WarehouseBag, dest, b)
	}
	stored := src
	stored.UID, stored.Bag, stored.Slot = db.NewItemUID(), 0, dest

	s.items = append(s.items[:si], s.items[si+1:]...)
	s.stored = append(s.stored, stored)

	return buildBagToBag(srcBag, srcSlot, nil, WarehouseBag, dest, &stored)
}

func (s *warehousePlan) warehouseWithdraw(src, destBag, destSlot uint8) []byte {
	ch := s.char
	si, di := findStored(s.stored, src), findItem(s.items, destBag, destSlot)
	if si < 0 {
		var d *db.Item
		if di >= 0 {
			d = &s.items[di]
		}
		return buildBagToBag(WarehouseBag, src, nil, destBag, destSlot, d)
	}
	a := s.stored[si]
	// CUser::ItemBankToBag: taking an item out costs 5% of its Sell price
	// times the count (item mall type 94 is free).
	fee := uint32(0)
	if d := db.MustDefs().Item(a.Type, a.TypeID); d != nil && a.Type != 94 {
		fee = uint32(math.RoundToEven(float64(float32(uint32(a.Count)*uint32(max(d.Sell, 0)))) * float64(float32(0.05))))
	}
	if ch.Money < fee {
		var d *db.Item
		if di >= 0 {
			d = &s.items[di]
		}
		return buildBagToBag(WarehouseBag, src, &a, destBag, destSlot, d)
	}
	pay := func() {
		if fee > 0 {
			ch.Money -= fee
		}
	}
	if di >= 0 {
		b := &s.items[di]
		if ms := maxStackOf(a.Type, a.TypeID); itemsCanStack(&a, b) && ms > 1 && b.Count < ms {
			moved := min(a.Count, ms-b.Count)
			pay()
			b.Count += moved

			bCopy := *b
			s.stored[si].Count -= moved
			var after *db.Item
			if s.stored[si].Count == 0 {
				s.stored = append(s.stored[:si], s.stored[si+1:]...)
			} else {
				after = copyPtr(&s.stored[si])
			}
			return buildBagToBag(WarehouseBag, src, after, destBag, destSlot, &bCopy)
		}
		return buildBagToBag(WarehouseBag, src, &a, destBag, destSlot, b)
	}
	it := a
	it.UID, it.Bag, it.Slot = db.NewItemUID(), destBag, destSlot

	s.stored = append(s.stored[:si], s.stored[si+1:]...)
	s.items = append(s.items, it)
	pay()

	return buildBagToBag(WarehouseBag, src, nil, destBag, destSlot, &it)
}

// --- NPC shop ---

func buildShopSellResult(result, bag, slot, itemType, typeID uint8, remaining uint16, money uint32) []byte {
	w := packet.New(packet.ShopSell)
	w.Bytes([]byte{result, bag, slot, itemType, typeID})
	w.Count(remaining)
	w.U32(money)
	return w.MustFinish()
}

// handleShopSell (0x0703): bag, slot, count. Items.Sell is per unit.
func (s *Session) handleShopSell(p []byte) {
	if len(p) < 5 || s.char == nil || s.autoAttack || s.auctionContact {
		return
	}
	// npc_shop sell (CUser::PacketShop 0x0703): not from the equipment (bag
	// 0), count 1..stack, else ignored; real type 27 pays nothing and a
	// partial sale of it keeps the stack.
	before := *s.char
	after := before
	ch := &after
	items := append([]db.Item(nil), s.items...)
	r := packet.NewReader(p)
	bag, slot, count := r.U8(), r.U8(), r.Count()
	if r.Err() != nil {
		return
	}
	if bag == 0 || !isValidItemPosition(bag, slot) {
		return
	}
	i := findItem(items, bag, slot)
	if i < 0 || count == 0 || count > items[i].Count {
		return
	}
	it := items[i]
	def := db.MustDefs().Item(it.Type, it.TypeID)
	if def == nil {
		return
	}
	whole := it.Count == count
	quest := realType(it.Type) == 27
	unitPrice := proto.ScaleVendorUnitPrice(uint32(max(def.Sell, 0)), activeEconomy().VendorSellPercent)
	if !s.economyAmountMatches(unitPrice) {
		return
	}
	proceeds := uint64(unitPrice) * uint64(count)
	if !quest && (ch.Money >= db.MoneyCap || uint64(ch.Money)+proceeds > db.MoneyCap) {
		s.send(buildShopSellResult(1, bag, slot, it.Type, it.TypeID, it.Count, ch.Money), "0x0703 shop sell (gold limit)")
		return
	}
	remaining := uint16(0)
	if whole {

		items = append(items[:i], items[i+1:]...)
	} else if !quest {
		items[i].Count -= count

		remaining = items[i].Count
	} else {
		remaining = it.Count
	}
	if !quest {
		ch.Money = uint32(uint64(ch.Money) + proceeds)
	}
	if err := db.CommitGameplay([]db.CharacterChange{{Before: before, After: after, ItemsBefore: s.items, ItemsAfter: items}}, nil); err != nil {
		s.logErr("shop sale transaction", err)
		return
	}
	*s.char = after
	s.items = items
	s.syncPlayer()
	s.goldActivity(before.Money, "NPC shop sale")
	s.log.Printf("GX Server World sold item %d/%d count=%d proceeds=%d", it.Type, it.TypeID, count, proceeds)
	s.send(buildShopSellResult(0, bag, slot, it.Type, it.TypeID, remaining, ch.Money), "0x0703 shop sell")
}
