# Protocol: Item Mall and account points

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 2601

**`0x2601` — AccountPoints**

Declarations: [packet.AccountPoints](../../server/internal/packet/packet.go#L203), [proto.AccountPoints](../../internal/proto/opcodes.go#L190).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildAccountPoints (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L88)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildAccountPoints — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L88)

```go
func buildAccountPoints(points uint32) []byte {
	w := packet.New(packet.AccountPoints)
	w.U32(points)
	w.U8(0)
	return w.MustFinish()
}
```

</details>

## Opcode 2602

**`0x2602` — PointBuy**

Declarations: [packet.PointBuy](../../server/internal/packet/packet.go#L204), [proto.PointBuy](../../internal/proto/opcodes.go#L191).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L288)

**S→C serializers:** [buildPointBuyResult (server/internal/world/itemmall.go)](../../server/internal/world/itemmall.go#L28)

**Client receive reducers:** [applyMall (internal/game/mall.go)](../../internal/game/mall.go#L31), [Apply (internal/game/state.go)](../../internal/game/state.go#L657)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildPointBuyResult — server/internal/world/itemmall.go</summary>

[Source definition](../../server/internal/world/itemmall.go#L28)

```go
func buildPointBuyResult(result uint8, points uint32, code string, cost uint32, items []grantedItem) []byte {
	w := packet.New(packet.PointBuy)
	w.U8(result)
	w.U32(points)
	w.FixedString(code, 21)
	w.U32(time32())
	w.U32(cost)
	w.U8(uint8(len(items)))
	for _, g := range items {
		w.Bytes([]byte{g.Bag, g.Slot, g.Type, g.TypeID})
		w.Count(g.Count)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>MallBuyPacket — internal/proto/mall.go</summary>

[Source definition](../../internal/proto/mall.go#L107)

```go
func MallBuyPacket(code string) []byte {
	w := NewWriter(PointBuy)
	w.Fixed(code, 21, true)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeMallPurchase — internal/proto/mall.go</summary>

[Source definition](../../internal/proto/mall.go#L33)

```go
func DecodeMallPurchase(p []byte) (MallPurchase, error) {
	r := NewReader(p)
	result := MallPurchase{Result: r.U8()}
	if err := r.Err(); err != nil {
		return result, err
	}
	if result.Result != 0 && r.Remaining() == 0 {
		return result, nil
	}
	result.Points, result.Code, result.Timestamp, result.Cost = r.U32(), r.Fixed(21), r.U32(), r.U32()
	n := int(r.U8())
	if err := r.Err(); err != nil {
		return result, err
	}
	if n > r.Remaining()/5 {
		return result, ErrShort
	}
	result.Items = make([]MallItem, n)
	for i := range result.Items {
		result.Items[i] = MallItem{r.U8(), r.U8(), r.U8(), r.U8(), r.Count()}
	}
	result.HasDetails = true
	return result, r.Err()
}
```

```go
type MallItem struct {
	Bag, Slot, Type, TypeID uint8
	Count                   uint16
}
```

```go
type MallPurchase struct {
	Result          uint8
	HasDetails      bool
	Points          uint32
	Code            string
	Timestamp, Cost uint32
	Items           []MallItem
}
```

</details>

## Opcode 2603

**`0x2603` — PointGiftSend**

Declarations: [packet.PointGiftSend](../../server/internal/packet/packet.go#L205), [proto.PointGiftSend](../../internal/proto/opcodes.go#L192).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L290)

**S→C serializers:** [buildGiftSendResult (server/internal/world/itemmall.go)](../../server/internal/world/itemmall.go#L43)

**Client receive reducers:** [applyMall (internal/game/mall.go)](../../internal/game/mall.go#L41), [Apply (internal/game/state.go)](../../internal/game/state.go#L657)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGiftSendResult — server/internal/world/itemmall.go</summary>

[Source definition](../../server/internal/world/itemmall.go#L43)

```go
func buildGiftSendResult(result uint8, target string, points uint32, code string, cost uint32) []byte {
	w := packet.New(packet.PointGiftSend)
	w.U8(result)
	w.FixedString(target, 21)
	w.U32(points)
	w.FixedString(code, 21)
	w.U32(time32())
	w.U32(cost)
	return w.MustFinish()
}
```

</details>

<details><summary>MallGiftPacket — internal/proto/mall.go</summary>

[Source definition](../../internal/proto/mall.go#L112)

```go
func MallGiftPacket(code, target string) []byte {
	w := NewWriter(PointGiftSend)
	w.Fixed(code, 21, true)
	w.Fixed(target, 21, true)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeMallGiftResult — internal/proto/mall.go</summary>

[Source definition](../../internal/proto/mall.go#L57)

```go
func DecodeMallGiftResult(p []byte) (MallGiftResult, error) {
	r := NewReader(p)
	result := MallGiftResult{Result: r.U8()}
	if err := r.Err(); err != nil {
		return result, err
	}
	if result.Result != 0 && r.Remaining() == 0 {
		return result, nil
	}
	result.Target, result.Points, result.Code, result.Timestamp, result.Cost = r.Fixed(21), r.U32(), r.Fixed(21), r.U32(), r.U32()
	if err := r.Err(); err != nil {
		return result, err
	}
	result.HasDetails = true
	return result, nil
}
```

```go
type MallGiftResult struct {
	Result          uint8
	HasDetails      bool
	Target          string
	Points          uint32
	Code            string
	Timestamp, Cost uint32
}
```

</details>

## Opcode 2605

**`0x2605` — CashPoint**

Declarations: [packet.CashPoint](../../server/internal/packet/packet.go#L208), [proto.CashPoint](../../internal/proto/opcodes.go#L195).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L298)

**S→C serializers:** [buildCashPoints (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L95)

**Client receive reducers:** [applyMall (internal/game/mall.go)](../../internal/game/mall.go#L25), [Apply (internal/game/state.go)](../../internal/game/state.go#L657)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCashPoints — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L95)

```go
func buildCashPoints(points uint32) []byte {
	w := packet.New(packet.CashPoint)
	w.U32(points)
	return w.MustFinish()
}
```

</details>

<details><summary>MallPointsPacket — internal/proto/mall.go</summary>

[Source definition](../../internal/proto/mall.go#L106)

```go
func MallPointsPacket() []byte { return NewWriter(CashPoint).Bytes() }
```

</details>

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 2606

**`0x2606` — PointStoredItemList**

Declarations: [packet.PointStoredItemList](../../server/internal/packet/packet.go#L206), [proto.PointStoredItemList](../../internal/proto/opcodes.go#L193).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildGiftBoxList (server/internal/world/itemmall.go)](../../server/internal/world/itemmall.go#L55)

**Client receive reducers:** [applyMall (internal/game/mall.go)](../../internal/game/mall.go#L51), [Apply (internal/game/state.go)](../../internal/game/state.go#L657)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGiftBoxList — server/internal/world/itemmall.go</summary>

[Source definition](../../server/internal/world/itemmall.go#L55)

```go
func buildGiftBoxList(userUID uint32) ([]byte, error) {
	gifts, err := db.GiftBox(userUID)
	if err != nil {
		return nil, err
	}
	w := packet.New(packet.PointStoredItemList)
	w.U8(uint8(min(len(gifts), db.GiftBoxSlots)))
	for _, g := range gifts {
		w.Bytes([]byte{g.Slot, g.Type, g.TypeID})
		w.Count(g.Count)
	}
	return w.MustFinish(), nil
}
```

</details>

<details><summary>DecodeMallStored — internal/proto/mall.go</summary>

[Source definition](../../internal/proto/mall.go#L73)

```go
func DecodeMallStored(p []byte) ([]MallStoredItem, error) {
	r := NewReader(p)
	n := int(r.U8())
	if err := r.Err(); err != nil {
		return nil, err
	}
	if n > 240 || n > r.Remaining()/4 {
		return nil, ErrShort
	}
	items := make([]MallStoredItem, n)
	var seen [240]bool
	for i := range items {
		it := MallStoredItem{r.U8(), r.U8(), r.U8(), r.Count()}
		if it.Slot >= 240 || it.Type == 0 || it.Count == 0 || seen[it.Slot] {
			return nil, fmt.Errorf("mall: invalid stored item %d", it.Slot)
		}
		seen[it.Slot] = true
		items[i] = it
	}
	return items, r.Err()
}
```

```go
type MallStoredItem struct {
	Slot, Type, TypeID uint8
	Count              uint16
}
```

</details>

## Opcode 2607

**`0x2607` — PointStoredItemGet**

Declarations: [packet.PointStoredItemGet](../../server/internal/packet/packet.go#L207), [proto.PointStoredItemGet](../../internal/proto/opcodes.go#L194).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L292)

**S→C serializers:** [tradePacket (server/internal/world/trade.go)](../../server/internal/world/trade.go#L76)

**Client receive reducers:** [applyMall (internal/game/mall.go)](../../internal/game/mall.go#L61), [Apply (internal/game/state.go)](../../internal/game/state.go#L657)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>tradePacket — server/internal/world/trade.go</summary>

[Source definition](../../server/internal/world/trade.go#L76)

```go
func tradePacket(op packet.Opcode, b ...byte) []byte {
	w := packet.New(op)
	w.Bytes(b)
	return w.MustFinish()
}
```

</details>

<details><summary>MallClaimPacket — internal/proto/mall.go</summary>

[Source definition](../../internal/proto/mall.go#L118)

```go
func MallClaimPacket(slot uint8) []byte {
	w := NewWriter(PointStoredItemGet)
	w.U8(slot)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeMallClaim — internal/proto/mall.go</summary>

[Source definition](../../internal/proto/mall.go#L94)

```go
func DecodeMallClaim(p []byte) (MallClaim, error) {
	r := NewReader(p)
	m := MallClaim{r.U8(), r.U8(), r.U8()}
	if err := r.Err(); err != nil {
		return m, err
	}
	if m.Slot >= 240 || m.Bag < 1 || m.Bag > 5 || m.InventorySlot >= 24 {
		return m, fmt.Errorf("mall: invalid claim location")
	}
	return m, nil
}
```

```go
type MallClaim struct{ Slot, Bag, InventorySlot uint8 }
```

</details>
