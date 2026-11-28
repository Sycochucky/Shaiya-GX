# Protocol: Shops, storage and blacksmith

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 0702

**`0x0702` — ShopBuy**

Declarations: [packet.ShopBuy](../../server/internal/packet/packet.go#L78), [proto.ShopBuy](../../internal/proto/opcodes.go#L62).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L109)

**S→C serializers:** [buildShopBuyResult (server/internal/world/npcquest.go)](../../server/internal/world/npcquest.go#L130)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L756)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildShopBuyResult — server/internal/world/npcquest.go</summary>

[Source definition](../../server/internal/world/npcquest.go#L130)

```go
func buildShopBuyResult(result, bag, slot, itemType, typeID uint8, count uint16, money uint32) []byte {
	w := packet.New(packet.ShopBuy)
	w.Bytes([]byte{result, bag, slot, itemType, typeID})
	w.Count(count)
	w.U32(money)
	return w.MustFinish()
}
```

</details>

<details><summary>ShopBuyPacket — internal/proto/shop.go</summary>

[Source definition](../../internal/proto/shop.go#L4)

```go
func ShopBuyPacket(npc uint32, index uint8, count uint16) []byte {
	w := NewWriter(ShopBuy)
	w.U32(npc)
	w.U8(index)
	w.Count(count)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeShopResult — internal/proto/shop.go</summary>

[Source definition](../../internal/proto/shop.go#L26)

```go
func DecodeShopResult(p []byte) (m ShopResult, err error) {
	r := NewReader(p)
	m.Result, m.Bag, m.Slot, m.Type, m.TypeID, m.Count = r.U8(), r.U8(), r.U8(), r.U8(), r.U8(), r.Count()
	m.Money = r.U32()
	return m, r.Err()
}
```

```go
type ShopResult struct {
	Result, Bag, Slot, Type, TypeID uint8
	Count                           uint16
	Money                           uint32
}
```

</details>

## Opcode 0703

**`0x0703` — ShopSell**

Declarations: [packet.ShopSell](../../server/internal/packet/packet.go#L149), [proto.ShopSell](../../internal/proto/opcodes.go#L133).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L107)

**S→C serializers:** [buildShopSellResult (server/internal/world/warehouse.go)](../../server/internal/world/warehouse.go#L246)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L756)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildShopSellResult — server/internal/world/warehouse.go</summary>

[Source definition](../../server/internal/world/warehouse.go#L246)

```go
func buildShopSellResult(result, bag, slot, itemType, typeID uint8, remaining uint16, money uint32) []byte {
	w := packet.New(packet.ShopSell)
	w.Bytes([]byte{result, bag, slot, itemType, typeID})
	w.Count(remaining)
	w.U32(money)
	return w.MustFinish()
}
```

</details>

<details><summary>ShopSellPacket — internal/proto/shop.go</summary>

[Source definition](../../internal/proto/shop.go#L13)

```go
func ShopSellPacket(bag, slot uint8, count uint16) []byte {
	w := NewWriter(ShopSell)
	w.Raw([]byte{bag, slot})
	w.Count(count)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeShopResult — internal/proto/shop.go</summary>

[Source definition](../../internal/proto/shop.go#L26)

```go
func DecodeShopResult(p []byte) (m ShopResult, err error) {
	r := NewReader(p)
	m.Result, m.Bag, m.Slot, m.Type, m.TypeID, m.Count = r.U8(), r.U8(), r.U8(), r.U8(), r.U8(), r.Count()
	m.Money = r.U32()
	return m, r.Err()
}
```

```go
type ShopResult struct {
	Result, Bag, Slot, Type, TypeID uint8
	Count                           uint16
	Money                           uint32
}
```

</details>

## Opcode 0711

**`0x0711` — WarehouseItemList**

Declarations: [packet.WarehouseItemList](../../server/internal/packet/packet.go#L148), [proto.WarehouseItemList](../../internal/proto/opcodes.go#L132).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildWarehouseItemList (server/internal/world/warehouse.go)](../../server/internal/world/warehouse.go#L27)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1005), [applyInventoryPacket (internal/game/state_inventory_packets.go)](../../internal/game/state_inventory_packets.go#L42)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildWarehouseItemList — server/internal/world/warehouse.go</summary>

[Source definition](../../server/internal/world/warehouse.go#L27)

```go
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
```

</details>

<details><summary>DecodeWarehouse — internal/proto/warehouse.go</summary>

[Source definition](../../internal/proto/warehouse.go#L14)

```go
func DecodeWarehouse(p []byte) (list WarehouseList, err error) {
	r := NewReader(p)
	list.Gold = r.U32()
	n := int(r.U8())
	if n > WarehouseSlots {
		return list, fmt.Errorf("warehouse: too many items %d", n)
	}
	if n > r.Remaining()/33 {
		return list, ErrShort
	}
	list.Items = make([]InventoryItem, n)
	seen := [WarehouseSlots]bool{}
	for i := range list.Items {
		m := &list.Items[i]
		m.Bag = WarehouseBag
		m.Slot, m.Type, m.TypeID = r.U8(), r.U8(), r.U8()
		m.Quality = r.U16()
		for j := range m.Gems {
			m.Gems[j] = r.U8()
		}
		m.Count, m.Craft = r.Count(), r.Fixed(21)
		if int(m.Slot) >= WarehouseSlots || seen[m.Slot] {
			return list, fmt.Errorf("warehouse: invalid/duplicate slot %d", m.Slot)
		}
		seen[m.Slot] = true
	}
	return list, r.Err()
}
```

```go
type WarehouseList struct {
	Gold  uint32
	Items []InventoryItem
}
```

</details>

## Opcode 0801

**`0x0801` — GemAdd**

Declarations: [packet.GemAdd](../../server/internal/packet/packet.go#L222), [proto.GemAdd](../../internal/proto/opcodes.go#L209).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L91)

**S→C serializers:** [buildGemAddResult (server/internal/world/crafting.go)](../../server/internal/world/crafting.go#L188)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L611)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGemAddResult — server/internal/world/crafting.go</summary>

[Source definition](../../server/internal/world/crafting.go#L188)

```go
func buildGemAddResult(result, gBag, gSlot uint8, gCount uint16, iBag, iSlot, pos, gemID uint8, money uint32) []byte {
	w := packet.New(packet.GemAdd)
	w.Bytes([]byte{result, gBag, gSlot})
	w.Count(gCount)
	w.Bytes([]byte{iBag, iSlot, pos, gemID})
	w.U32(0) // four unused bytes (client reader 0x0801: 12 bytes + money)
	w.U32(money)
	return w.MustFinish()
}
```

</details>

<details><summary>GemAddPacket — internal/proto/gem.go</summary>

[Source definition](../../internal/proto/gem.go#L27)

```go
func GemAddPacket(gemBag, gemSlot, itemBag, itemSlot uint8) []byte {
	w := NewWriter(GemAdd)
	w.Raw([]byte{gemBag, gemSlot, itemBag, itemSlot})
	return w.Bytes()
}
```

</details>

<details><summary>GemAddHammerPacket — internal/proto/gem.go</summary>

[Source definition](../../internal/proto/gem.go#L78)

```go
func GemAddHammerPacket(gemBag, gemSlot, itemBag, itemSlot, hammerBag, hammerSlot uint8) []byte {
	w := NewWriter(GemAdd)
	w.Raw([]byte{gemBag, gemSlot, itemBag, itemSlot, hammerBag, hammerSlot})
	return w.Bytes()
}
```

</details>

<details><summary>DecodeGemAddResult — internal/proto/gem.go</summary>

[Source definition](../../internal/proto/gem.go#L13)

```go
func DecodeGemAddResult(p []byte) (GemAddResult, error) {
	r := NewReader(p)
	v := GemAddResult{Result: r.U8(), GemBag: r.U8(), GemSlot: r.U8(), GemCount: r.Count(), ItemBag: r.U8(), ItemSlot: r.U8(), Position: r.U8(), GemID: r.U8()}
	r.U32()
	v.Money = r.U32()
	if err := r.Err(); err != nil {
		return v, err
	}
	if v.Result > 1 || v.GemBag > 5 || v.GemSlot >= 24 || v.ItemBag > 5 || v.ItemSlot >= 24 || (v.Result == 1 && (v.Position >= 6 || v.GemID == 0)) {
		return v, fmt.Errorf("invalid gem add result")
	}
	return v, nil
}
```

```go
type GemAddResult struct {
	Result, GemBag, GemSlot            uint8
	GemCount                           uint16
	ItemBag, ItemSlot, Position, GemID uint8
	Money                              uint32
}
```

</details>

## Opcode 0802

**`0x0802` — GemRemove**

Declarations: [packet.GemRemove](../../server/internal/packet/packet.go#L223), [proto.GemRemove](../../internal/proto/opcodes.go#L210).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L93)

**S→C serializers:** [planGemRemove (server/internal/world/crafting.go)](../../server/internal/world/crafting.go#L366)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L562)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>planGemRemove — server/internal/world/crafting.go</summary>

[Source definition](../../server/internal/world/crafting.go#L366)

```go
func (s *Session) planGemRemove(p []byte) {
	if len(p) < 4 || s.char == nil {
		return
	}
	ch := s.char
	iBag, iSlot := p[2], p[3]
	if !isValidItemPosition(iBag, iSlot) {
		return
	}
	ii := findItem(s.items, iBag, iSlot)
	if ii < 0 {
		return
	}
	item := s.items[ii]
	var removed, ids, bags, slots [maxGemSlots]uint8
	result := func(ok bool) {
		w := packet.New(packet.GemRemove)
		w.Bool(ok)
		w.U8(iBag)
		w.U8(iSlot)
		w.Bytes(removed[:])
		w.Bytes(ids[:])
		w.Bytes(bags[:])
		w.Bytes(slots[:])
		w.U32(ch.Money)
		s.send(w.MustFinish(), "0x0802 gem remove result")
	}
	defs := db.MustDefs()
	positions := []int{}
	toolUID := int64(0)
	rate := -1
	var cost uint32
	if len(p) == 8 {
		hi := findItem(s.items, p[4], p[5])
		pos := int(p[6])
		if hi < 0 || p[4] == 0 || !isValidItemPosition(p[4], p[5]) || pos >= maxGemSlots || p[7] != 0 || s.items[hi].UID == item.UID {
			result(false)
			return
		}
		tool := s.items[hi]
		hd := defs.Item(tool.Type, tool.TypeID)
		if tool.Count == 0 || !validForgeHammer(hd, specialGemExtract) || item.Gems[pos] == 0 {
			result(false)
			return
		}
		rate = extractionChance(hd)
		toolUID = tool.UID
		positions = append(positions, pos)
	} else {
		result(false)
		return
	}
	for _, pos := range positions {
		d := defs.Item(gemItemType, item.Gems[pos])
		if d == nil {
			result(false)
			return
		}
		cost += gemCost(ch, d)
	}
	mixRate, _ := s.houseMix()
	cost = houseRound(float64(cost) * mixRate)
	if len(positions) == 0 || item.Count != 1 || realType(item.Type) == 24 || ch.Money < cost || len(freeInventorySlots(s.items)) < len(positions) {
		result(false)
		return
	}
	before := s.maxPools()
	ch.Money -= cost
	if toolUID != 0 {
		hi := findItemByUID(s.items, toolUID)
		s.adjustCount(toolUID, s.items[hi].Count-1)
	}
	extracted := false
	for _, pos := range positions {
		id := item.Gems[pos]
		chance := rate
		if chance < 0 {
			chance = gemChance(ch, s.faction, defs.Item(gemItemType, id), nil)
		}
		i := findItemByUID(s.items, item.UID)
		if rand.IntN(100) < chance {
			free := freeInventorySlots(s.items)
			g := db.Item{Bag: free[0][0], Slot: free[0][1], Type: gemItemType, TypeID: id, Count: 1}
			if err := s.addInventoryItem(&g); err != nil {
				s.log.Printf("gem extraction insert: %v", err)
				continue
			}
			s.items = append(s.items, g)
			s.send(buildAddItem(&g), "0x0205 extracted lapis")
			removed[pos], ids[pos], bags[pos], slots[pos] = 1, id, g.Bag, g.Slot
			extracted = true
		}
		// A failed roll destroys only this lapis; other sockets stay untouched.
		s.items[i].Gems[pos] = 0
	}
	i := findItemByUID(s.items, item.UID)
	s.saveInventoryItem(&s.items[i])
	s.persist()
	result(extracted)
	if iBag == 0 {
		s.sendChangedPools(before, "gem extraction")
	}
	if s.player != nil {
		s.player.UpdateItems(s.items)
	}
}
```

</details>

<details><summary>GemRemovePacket — internal/proto/gem.go</summary>

[Source definition](../../internal/proto/gem.go#L71)

```go
func GemRemovePacket(itemBag, itemSlot uint8) []byte {
	w := NewWriter(GemRemove)
	w.Raw([]byte{itemBag, itemSlot})
	return w.Bytes()
}
```

</details>

<details><summary>GemRemoveHammerPacket — internal/proto/gem.go</summary>

[Source definition](../../internal/proto/gem.go#L86)

```go
func GemRemoveHammerPacket(item, hammer [2]uint8, socket uint8) []byte {
	w := NewWriter(GemRemove)
	w.Raw([]byte{item[0], item[1], hammer[0], hammer[1], socket, 0})
	return w.Bytes()
}
```

</details>

<details><summary>DecodeGemRemoveResult — internal/proto/gem.go</summary>

[Source definition](../../internal/proto/gem.go#L41)

```go
func DecodeGemRemoveResult(p []byte) (GemRemoveResult, error) {
	r := NewReader(p)
	v := GemRemoveResult{Result: r.U8(), ItemBag: r.U8(), ItemSlot: r.U8()}
	for i := range v.Removed {
		v.Removed[i] = r.U8()
	}
	for i := range v.IDs {
		v.IDs[i] = r.U8()
	}
	for i := range v.Bags {
		v.Bags[i] = r.U8()
	}
	for i := range v.Slots {
		v.Slots[i] = r.U8()
	}
	v.Money = r.U32()
	if err := r.Err(); err != nil {
		return v, err
	}
	if v.Result > 1 || v.ItemBag > 5 || v.ItemSlot >= 24 {
		return v, fmt.Errorf("invalid gem remove result")
	}
	for i, removed := range v.Removed {
		if removed > 1 || (removed != 0 && (v.Result == 0 || v.IDs[i] == 0 || v.Bags[i] < 1 || v.Bags[i] > 5 || v.Slots[i] >= 24)) {
			return v, fmt.Errorf("invalid extracted gem")
		}
	}
	return v, nil
}
```

```go
type GemRemoveResult struct {
	Result, ItemBag, ItemSlot uint8
	Removed, IDs, Bags, Slots [6]uint8
	Money                     uint32
}
```

</details>

## Opcode 0803

**`0x0803` — CloakAdd**

Declarations: [packet.CloakAdd](../../server/internal/packet/packet.go#L224), [proto.CloakAdd](../../internal/proto/opcodes.go#L211).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L95)

**S→C serializers:** [buildEmblemAddResult (server/internal/world/emblem.go)](../../server/internal/world/emblem.go#L15)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L583)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildEmblemAddResult — server/internal/world/emblem.go</summary>

[Source definition](../../server/internal/world/emblem.go#L15)

```go
func buildEmblemAddResult(result, bag, slot uint8, data [6]uint8, gold uint32) []byte {
	w := packet.New(packet.CloakAdd)
	w.Bytes([]byte{result, bag, slot})
	w.Bytes(data[:])
	w.U32(gold)
	return w.MustFinish()
}
```

</details>

<details><summary>EmblemApplyPacket — internal/proto/emblem.go</summary>

[Source definition](../../internal/proto/emblem.go#L46)

```go
func EmblemApplyPacket(bag, slot uint8, pattern [6]uint8) []byte {
	w := NewWriter(CloakAdd)
	w.Raw([]byte{bag, slot})
	w.Raw(pattern[:])
	return w.Bytes()
}
```

</details>

<details><summary>DecodeEmblemResult — internal/proto/emblem.go</summary>

[Source definition](../../internal/proto/emblem.go#L14)

```go
func DecodeEmblemResult(p []byte) (EmblemResult, error) {
	r := NewReader(p)
	v := EmblemResult{Result: r.U8(), Bag: r.U8(), Slot: r.U8()}
	for i := range v.Pattern {
		v.Pattern[i] = r.U8()
	}
	v.Money = r.U32()
	if err := r.Err(); err != nil {
		return v, err
	}
	if v.Result > 1 || v.Bag > 5 || v.Slot >= 24 {
		return v, fmt.Errorf("invalid emblem result")
	}
	return v, nil
}
```

```go
type EmblemResult struct {
	Result, Bag, Slot uint8
	Pattern           [6]uint8
	Money             uint32
}
```

</details>

## Opcode 0804

**`0x0804` — CloakRemove**

Declarations: [packet.CloakRemove](../../server/internal/packet/packet.go#L225), [proto.CloakRemove](../../internal/proto/opcodes.go#L212).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L97)

**S→C serializers:** [planCloakRemove (server/internal/world/emblem.go)](../../server/internal/world/emblem.go#L64)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L600)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>planCloakRemove — server/internal/world/emblem.go</summary>

[Source definition](../../server/internal/world/emblem.go#L64)

```go
func (s *Session) planCloakRemove(p []byte) {
	if len(p) < 4 || s.char == nil {
		return
	}
	bag, slot := p[2], p[3]
	i := s.emblemTarget(bag, slot)
	if i < 0 {
		return
	}
	s.items[i].Gems = [6]uint8{}
	s.saveInventoryItem(&s.items[i])
	w := packet.New(packet.CloakRemove)
	w.Bytes([]byte{bag, slot})
	s.send(w.MustFinish(), "0x0804 cloak emblem removed")
	if bag == 0 && slot == 7 {
		s.equipmentChanged([]uint8{7})
	}
}
```

</details>

<details><summary>EmblemRemovePacket — internal/proto/emblem.go</summary>

[Source definition](../../internal/proto/emblem.go#L54)

```go
func EmblemRemovePacket(bag, slot uint8) []byte {
	w := NewWriter(CloakRemove)
	w.Raw([]byte{bag, slot})
	return w.Bytes()
}
```

</details>

<details><summary>DecodeEmblemRemoved — internal/proto/emblem.go</summary>

[Source definition](../../internal/proto/emblem.go#L33)

```go
func DecodeEmblemRemoved(p []byte) (EmblemRemoved, error) {
	r := NewReader(p)
	v := EmblemRemoved{Bag: r.U8(), Slot: r.U8()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if v.Bag > 5 || v.Slot >= 24 {
		return v, fmt.Errorf("invalid emblem removal")
	}
	return v, nil
}
```

```go
type EmblemRemoved struct{ Bag, Slot uint8 }
```

</details>

## Opcode 0805

**`0x0805` — EnchantAdd**

Declarations: [packet.EnchantAdd](../../server/internal/packet/packet.go#L220), [proto.EnchantAdd](../../internal/proto/opcodes.go#L207).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L89)

**S→C serializers:** [buildEnchantResult (server/internal/world/crafting.go)](../../server/internal/world/crafting.go#L37)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L535)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildEnchantResult — server/internal/world/crafting.go</summary>

[Source definition](../../server/internal/world/crafting.go#L37)

```go
func buildEnchantResult(result, lBag, lSlot uint8, lCount uint16, mBag, mSlot uint8, mCount uint16, iBag, iSlot uint8, money uint32, craft string) []byte {
	w := packet.New(packet.EnchantAdd)
	w.Bytes([]byte{result, lBag, lSlot})
	w.Count(lCount)
	w.Bytes([]byte{mBag, mSlot})
	w.Count(mCount)
	w.Bytes([]byte{iBag, iSlot})
	w.U32(money)
	w.FixedString(craft, 21)
	return w.MustFinish()
}
```

</details>

<details><summary>EnchantPacket — internal/proto/enchant.go</summary>

[Source definition](../../internal/proto/enchant.go#L32)

```go
func EnchantPacket(lapisiaBag, lapisiaSlot, materialBag, materialSlot uint8, materialCount uint16, itemBag, itemSlot uint8) []byte {
	w := NewWriter(EnchantAdd)
	w.Raw([]byte{lapisiaBag, lapisiaSlot, materialBag, materialSlot})
	w.Count(materialCount)
	w.Raw([]byte{itemBag, itemSlot})
	return w.Bytes()
}
```

</details>

<details><summary>DecodeEnchantResult — internal/proto/enchant.go</summary>

[Source definition](../../internal/proto/enchant.go#L15)

```go
func DecodeEnchantResult(p []byte) (EnchantResult, error) {
	r := NewReader(p)
	v := EnchantResult{Result: r.U8(), LapisiaBag: r.U8(), LapisiaSlot: r.U8(), LapisiaCount: r.Count(),
		MaterialBag: r.U8(), MaterialSlot: r.U8(), MaterialCount: r.Count(), ItemBag: r.U8(), ItemSlot: r.U8(), Money: r.U32(), Craft: r.Fixed(21)}
	if err := r.Err(); err != nil {
		return v, err
	}
	if v.Result > 1 || v.LapisiaBag > 5 || v.MaterialBag > 5 || v.ItemBag > 5 || v.LapisiaSlot >= 24 || v.MaterialSlot >= 24 || v.ItemSlot >= 24 {
		return v, fmt.Errorf("invalid enchant result")
	}
	l, m, i := [2]uint8{v.LapisiaBag, v.LapisiaSlot}, [2]uint8{v.MaterialBag, v.MaterialSlot}, [2]uint8{v.ItemBag, v.ItemSlot}
	if l == m || l == i || m == i {
		return v, fmt.Errorf("overlapping enchant positions")
	}
	return v, nil
}
```

```go
type EnchantResult struct {
	Result, LapisiaBag, LapisiaSlot uint8
	LapisiaCount                    uint16
	MaterialBag, MaterialSlot       uint8
	MaterialCount                   uint16
	ItemBag, ItemSlot               uint8
	Money                           uint32
	Craft                           string
}
```

</details>

## Opcode 0806

**`0x0806` — ItemCompose**

Declarations: [packet.ItemCompose](../../server/internal/packet/packet.go#L221), [proto.ItemCompose](../../internal/proto/opcodes.go#L208).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L99)

**S→C serializers:** [buildComposeResult (server/internal/world/crafting.go)](../../server/internal/world/crafting.go#L501)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L519)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildComposeResult — server/internal/world/crafting.go</summary>

[Source definition](../../server/internal/world/crafting.go#L501)

```go
func buildComposeResult(result, bag, slot uint8, craft string) []byte {
	w := packet.New(packet.ItemCompose)
	w.Bytes([]byte{result, bag, slot})
	w.FixedString(craft, 21)
	return w.MustFinish()
}
```

</details>

<details><summary>ComposePacket — internal/proto/compose.go</summary>

[Source definition](../../internal/proto/compose.go#L21)

```go
func ComposePacket(runeBag, runeSlot, itemBag, itemSlot uint8) []byte {
	w := NewWriter(ItemCompose)
	w.Raw([]byte{runeBag, runeSlot, itemBag, itemSlot})
	return w.Bytes()
}
```

</details>

<details><summary>DecodeComposeResult — internal/proto/compose.go</summary>

[Source definition](../../internal/proto/compose.go#L10)

```go
func DecodeComposeResult(p []byte) (ComposeResult, error) {
	r := NewReader(p)
	v := ComposeResult{Result: r.U8(), Bag: r.U8(), Slot: r.U8(), Craft: r.Fixed(21)}
	if err := r.Err(); err != nil {
		return v, err
	}
	if v.Result > 1 || v.Bag > 5 || v.Slot >= 24 {
		return v, fmt.Errorf("invalid recreation result")
	}
	return v, nil
}
```

```go
type ComposeResult struct {
	Result, Bag, Slot uint8
	Craft             string
}
```

</details>

## Opcode 0809

**`0x0809` — GemAddPossibility**

Declarations: [packet.GemAddPossibility](../../server/internal/packet/packet.go#L226), [proto.GemAddPossibility](../../internal/proto/opcodes.go#L213).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

The shape is decoded inline, delegated, or only declared. Use the linked adapter/reducer and the codec appendix below; there is no inferred fixed layout here.

## Opcode 080a

**`0x080A` — GemRemovePossibility**

Declarations: [packet.GemRemovePossibility](../../server/internal/packet/packet.go#L227), [proto.GemRemovePossibility](../../internal/proto/opcodes.go#L214).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

The shape is decoded inline, delegated, or only declared. Use the linked adapter/reducer and the codec appendix below; there is no inferred fixed layout here.

## Opcode b101

**`0xB101` — BankItemList**

Declarations: [packet.BankItemList](../../server/internal/packet/packet.go#L209), [proto.BankItemList](../../internal/proto/opcodes.go#L196).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildZeroCount (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L74)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildZeroCount — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L74)

```go
func buildZeroCount(op packet.Opcode) []byte {
	w := packet.New(op)
	w.U8(0)
	return w.MustFinish()
}
```

</details>
