# Protocol: Player exchange and personal markets

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 0a01

**`0x0A01` — ExchangeRequest**

Declarations: [packet.ExchangeRequest](../../server/internal/packet/packet.go#L228), [proto.ExchangeRequest](../../internal/proto/opcodes.go#L215).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L181)

**S→C serializers:** [handleTradeRequest (server/internal/world/trade.go)](../../server/internal/world/trade.go#L157)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L37), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleTradeRequest — server/internal/world/trade.go</summary>

[Source definition](../../server/internal/world/trade.go#L157)

```go
func (s *Session) handleTradeRequest(target uint32) {
	me := s.player
	if me == nil || target == me.CharacterID || tradeFor(me.CharacterID) != nil {
		return
	}
	// CUser::PacketExchange 0x0A01: a user of the same zone and country
	// that is not trading.
	other := FindPlayerByID(target)
	if other == nil || !me.canSee(other) || other.MapID() != me.MapID() || other.Faction() != me.Faction() || tradeFor(target) != nil {
		return
	}
	tradeMu.Lock()
	reciprocal := tradeRequests[target] == me.CharacterID
	if reciprocal {
		delete(tradeRequests, target)
		delete(tradeRequests, me.CharacterID)
	} else {
		tradeRequests[me.CharacterID] = target
	}
	tradeMu.Unlock()
	if reciprocal {
		startTrade(me, other)
		return
	}
	w := packet.New(packet.ExchangeRequest)
	w.U32(me.CharacterID)
	other.Send(w.MustFinish(), "0x0A01 exchange request")
}
```

</details>

<details><summary>ExchangeRequestPacket — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L64)

```go
func ExchangeRequestPacket(id uint32) []byte {
	w := NewWriter(ExchangeRequest)
	w.U32(id)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeExchangeID — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L24)

```go
func DecodeExchangeID(p []byte) (uint32, error) { r := NewReader(p); id := r.U32(); return id, r.Err() }
```

</details>

## Opcode 0a02

**`0x0A02` — ExchangeAnswer**

Declarations: [packet.ExchangeAnswer](../../server/internal/packet/packet.go#L229), [proto.ExchangeAnswer](../../internal/proto/opcodes.go#L216).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L186)

**S→C serializers:** [tradePacket (server/internal/world/trade.go)](../../server/internal/world/trade.go#L76)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L44), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

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

<details><summary>ExchangeAnswerPacket — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L69)

```go
func ExchangeAnswerPacket(accept bool) []byte {
	w := NewWriter(ExchangeAnswer)
	if accept {
		w.U8(0)
	} else {
		w.U8(1)
	}
	return w.Bytes()
}
```

</details>

<details><summary>DecodeExchangeByte — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L31)

```go
func DecodeExchangeByte(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 0a03

**`0x0A03` — ExchangeStart**

Declarations: [packet.ExchangeStart](../../server/internal/packet/packet.go#L230), [proto.ExchangeStart](../../internal/proto/opcodes.go#L217).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildTradeStart (server/internal/world/trade.go)](../../server/internal/world/trade.go#L136)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L51), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildTradeStart — server/internal/world/trade.go</summary>

[Source definition](../../server/internal/world/trade.go#L136)

```go
func buildTradeStart(partner uint32) []byte {
	w := packet.New(packet.ExchangeStart)
	w.U32(partner)
	w.U8(0)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeExchangeStart — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L25)

```go
func DecodeExchangeStart(p []byte) (id uint32, result uint8, err error) {
	r := NewReader(p)
	id = r.U32()
	result = r.U8()
	return id, result, r.Err()
}
```

</details>

## Opcode 0a04

**`0x0A04` — ExchangeEnd**

Declarations: [packet.ExchangeEnd](../../server/internal/packet/packet.go#L231), [proto.ExchangeEnd](../../internal/proto/opcodes.go#L218).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [tradePacket (server/internal/world/trade.go)](../../server/internal/world/trade.go#L76)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L61), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

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

<details><summary>DecodeExchangeByte — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L31)

```go
func DecodeExchangeByte(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 0a05

**`0x0A05` — ExchangeReady**

Declarations: [packet.ExchangeReady](../../server/internal/packet/packet.go#L232), [proto.ExchangeReady](../../internal/proto/opcodes.go#L219).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L190)

**S→C serializers:** [tradePacket (server/internal/world/trade.go)](../../server/internal/world/trade.go#L76)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L74), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

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

<details><summary>ExchangeReadyPacket — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L78)

```go
func ExchangeReadyPacket(value uint8) []byte {
	w := NewWriter(ExchangeReady)
	w.U8(value)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeExchangeUpdate — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L32)

```go
func DecodeExchangeUpdate(p []byte) (u ExchangeUpdate, err error) {
	r := NewReader(p)
	u = ExchangeUpdate{r.U8(), r.U8()}
	return u, r.Err()
}
```

```go
type ExchangeUpdate struct{ Kind, Value uint8 }
```

</details>

<details><summary>Op — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L161)

```go
func Op(payload []byte) Opcode {
	if len(payload) < 2 {
		return 0
	}
	return Opcode(itemstack.Base(binary.LittleEndian.Uint16(payload)))
}
```

</details>

## Opcode 0a06

**`0x0A06` — ExchangeAdd**

Declarations: [packet.ExchangeAdd](../../server/internal/packet/packet.go#L233), [proto.ExchangeAdd](../../internal/proto/opcodes.go#L220).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L194)

**S→C serializers:** [handleTradeAdd (server/internal/world/trade.go)](../../server/internal/world/trade.go#L261)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L110), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleTradeAdd — server/internal/world/trade.go</summary>

[Source definition](../../server/internal/world/trade.go#L261)

```go
func (s *Session) handleTradeAdd(bag, slot uint8, count uint16, dest uint8) {
	if dest >= tradeSlots || bag == 0 || !isValidItemPosition(bag, slot) {
		return
	}
	i := findItem(s.items, bag, slot)
	if i < 0 {
		return
	}
	// 0x0A06: count 1..stack, not a quest item (real type 27), not offered
	// twice.
	it := s.items[i]
	if count == 0 || count > it.Count || realType(it.Type) == 27 {
		return
	}
	if t := tradeFor(s.char.ID); t != nil {
		if off := t.offer(s.char.ID); off != nil {
			for _, x := range off.items {
				if x != nil && x.Item.UID == it.UID {
					return
				}
			}
		}
	}
	n := count
	o := &offeredItem{Item: it, Count: n}
	_, partner, ok := s.mutateOffer(func(off *tradeOffer) { off.items[dest] = o })
	if !ok {
		return
	}
	w := packet.New(packet.ExchangeAdd)
	w.U8(bag)
	w.U8(slot)
	w.Count(n)
	w.U8(dest)
	s.send(w.MustFinish(), "0x0A06 exchange add (self)")
	if partner != nil {
		partner.Send(buildTradeAddDest(dest, o), "0x0A09 exchange add dest")
	}
}
```

</details>

<details><summary>ExchangeAddPacket — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L83)

```go
func ExchangeAddPacket(bag, slot uint8, count uint16, dest uint8) []byte {
	w := NewWriter(ExchangeAdd)
	w.Raw([]byte{bag, slot})
	w.Count(count)
	w.U8(dest)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeExchangeOwnItem — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L42)

```go
func DecodeExchangeOwnItem(p []byte) (item ExchangeOwnItem, err error) {
	r := NewReader(p)
	item = ExchangeOwnItem{Bag: r.U8(), Slot: r.U8(), Count: r.Count(), Destination: r.U8()}
	if err = r.Err(); err == nil && item.Destination >= ExchangeSlots {
		err = fmt.Errorf("exchange: invalid slot %d", item.Destination)
	}
	return
}
```

```go
type ExchangeOwnItem struct {
	Bag, Slot, Destination uint8
	Count                  uint16
}
```

</details>

## Opcode 0a07

**`0x0A07` — ExchangeRemove**

Declarations: [packet.ExchangeRemove](../../server/internal/packet/packet.go#L234), [proto.ExchangeRemove](../../internal/proto/opcodes.go#L221).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L202)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>ExchangeRemovePacket — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L90)

```go
func ExchangeRemovePacket(slot uint8) []byte {
	w := NewWriter(ExchangeRemove)
	w.U8(slot)
	return w.Bytes()
}
```

</details>

## Opcode 0a08

**`0x0A08` — ExchangeAddMoney**

Declarations: [packet.ExchangeAddMoney](../../server/internal/packet/packet.go#L235), [proto.ExchangeAddMoney](../../internal/proto/opcodes.go#L222).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L206)

**S→C serializers:** [buildTradeMoney (server/internal/world/trade.go)](../../server/internal/world/trade.go#L315)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L97), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildTradeMoney — server/internal/world/trade.go</summary>

[Source definition](../../server/internal/world/trade.go#L315)

```go
func buildTradeMoney(kind uint8, money uint32) []byte {
	w := packet.New(packet.ExchangeAddMoney)
	w.U8(kind)
	w.U32(money)
	return w.MustFinish()
}
```

</details>

<details><summary>ExchangeMoneyPacket — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L95)

```go
func ExchangeMoneyPacket(money uint32) []byte {
	w := NewWriter(ExchangeAddMoney)
	w.U32(money)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeExchangeMoney — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L37)

```go
func DecodeExchangeMoney(p []byte) (m ExchangeMoneyUpdate, err error) {
	r := NewReader(p)
	m = ExchangeMoneyUpdate{r.U8(), r.U32()}
	return m, r.Err()
}
```

```go
type ExchangeMoneyUpdate struct {
	Kind  uint8
	Money uint32
}
```

</details>

## Opcode 0a09

**`0x0A09` — ExchangeAddDest**

Declarations: [packet.ExchangeAddDest](../../server/internal/packet/packet.go#L236), [proto.ExchangeAddDest](../../internal/proto/opcodes.go#L223).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildTradeAddDest (server/internal/world/trade.go)](../../server/internal/world/trade.go#L232)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L120), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildTradeAddDest — server/internal/world/trade.go</summary>

[Source definition](../../server/internal/world/trade.go#L232)

```go
func buildTradeAddDest(slot uint8, o *offeredItem) []byte {
	w := packet.New(packet.ExchangeAddDest)
	w.U8(slot)
	writeOffered(w, o)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeExchangePartnerItem — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L50)

```go
func DecodeExchangePartnerItem(p []byte) (slot uint8, item ExchangeItem, err error) {
	r := NewReader(p)
	slot = r.U8()
	item.Type, item.TypeID, item.Count = r.U8(), r.U8(), r.Count()
	item.Quality = r.U16()
	for i := range item.Gems {
		item.Gems[i] = r.U8()
	}
	item.Craft = r.Fixed(21)
	if err = r.Err(); err == nil && slot >= ExchangeSlots {
		err = fmt.Errorf("exchange: invalid slot %d", slot)
	}
	return
}
```

```go
type ExchangeItem struct {
	Type, TypeID uint8
	Count        uint16
	Quality      uint16
	Gems         [6]uint8
	Craft        string
}
```

</details>

## Opcode 0a0a

**`0x0A0A` — ExchangeConfirm**

Declarations: [packet.ExchangeConfirm](../../server/internal/packet/packet.go#L237), [proto.ExchangeConfirm](../../internal/proto/opcodes.go#L224).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L211)

**S→C serializers:** [tradePacket (server/internal/world/trade.go)](../../server/internal/world/trade.go#L76)

**Client receive reducers:** [applyExchange (internal/game/exchange.go)](../../internal/game/exchange.go#L74), [Apply (internal/game/state.go)](../../internal/game/state.go#L647)

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

<details><summary>ExchangeConfirmPacket — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L100)

```go
func ExchangeConfirmPacket(confirm bool) []byte {
	w := NewWriter(ExchangeConfirm)
	if confirm {
		w.U8(1)
	} else {
		w.U8(0)
	}
	return w.Bytes()
}
```

</details>

<details><summary>DecodeExchangeUpdate — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L32)

```go
func DecodeExchangeUpdate(p []byte) (u ExchangeUpdate, err error) {
	r := NewReader(p)
	u = ExchangeUpdate{r.U8(), r.U8()}
	return u, r.Err()
}
```

```go
type ExchangeUpdate struct{ Kind, Value uint8 }
```

</details>

<details><summary>Op — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L161)

```go
func Op(payload []byte) Opcode {
	if len(payload) < 2 {
		return 0
	}
	return Opcode(itemstack.Base(binary.LittleEndian.Uint16(payload)))
}
```

</details>

## Opcode 2301

**`0x2301` — MyShopBegin**

Declarations: [proto.MyShopBegin](../../internal/proto/myshop.go#L12).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523. setup; the reply carries a result

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L203)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L39), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>NewReader — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L28)

```go
func NewReader(payload []byte) *Reader {
	r := &Reader{b: payload, off: 2}
	if len(payload) >= 2 {
		r.wide = itemstack.Wide(binary.LittleEndian.Uint16(payload))
	}
	if len(payload) < 2 {
		r.err = ErrShort
	}
	return r
}
```

```go
type Reader struct {
	wide bool
	b    []byte
	off  int
	err  error
}
```

</details>

## Opcode 2302

**`0x2302` — MyShopCancel**

Declarations: [proto.MyShopCancel](../../internal/proto/myshop.go#L13).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523. cancel; also the notice that the market ended

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L219)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L49), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

## Opcode 2303

**`0x2303` — MyShopOpen**

Declarations: [proto.MyShopOpen](../../internal/proto/myshop.go#L14).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523. sell, with the market message

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L221)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L51), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>MyShopOpenPacket — internal/proto/myshop.go</summary>

[Source definition](../../internal/proto/myshop.go#L63)

```go
func MyShopOpenPacket(title string) []byte {
	w := NewWriter(MyShopOpen)
	text := []byte(textwire.Prefix(title, MyShopTitleBytes))
	w.U8(uint8(len(text)))
	w.Raw(text)
	return w.Bytes()
}
```

</details>

## Opcode 2304

**`0x2304` — MyShopStop**

Declarations: [proto.MyShopStop](../../internal/proto/myshop.go#L15).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L223)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L55), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

## Opcode 2305

**`0x2305` — MyShopAdd**

Declarations: [proto.MyShopAdd](../../internal/proto/myshop.go#L16).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L229)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L59), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>MyShopAddPacket — internal/proto/myshop.go</summary>

[Source definition](../../internal/proto/myshop.go#L71)

```go
func MyShopAddPacket(index, bag, slot uint8, price uint32) []byte {
	w := NewWriter(MyShopAdd)
	w.U8(index)
	w.U8(bag)
	w.U8(slot)
	w.U32(price)
	return w.Bytes()
}
```

</details>

<details><summary>NewReader — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L28)

```go
func NewReader(payload []byte) *Reader {
	r := &Reader{b: payload, off: 2}
	if len(payload) >= 2 {
		r.wide = itemstack.Wide(binary.LittleEndian.Uint16(payload))
	}
	if len(payload) < 2 {
		r.err = ErrShort
	}
	return r
}
```

```go
type Reader struct {
	wide bool
	b    []byte
	off  int
	err  error
}
```

</details>

## Opcode 2306

**`0x2306` — MyShopRemove**

Declarations: [proto.MyShopRemove](../../internal/proto/myshop.go#L17).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L249)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L68), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>MyShopRemovePacket — internal/proto/myshop.go</summary>

[Source definition](../../internal/proto/myshop.go#L80)

```go
func MyShopRemovePacket(index uint8) []byte {
	w := NewWriter(MyShopRemove)
	w.U8(index)
	return w.Bytes()
}
```

</details>

<details><summary>NewReader — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L28)

```go
func NewReader(payload []byte) *Reader {
	r := &Reader{b: payload, off: 2}
	if len(payload) >= 2 {
		r.wide = itemstack.Wide(binary.LittleEndian.Uint16(payload))
	}
	if len(payload) < 2 {
		r.err = ErrShort
	}
	return r
}
```

```go
type Reader struct {
	wide bool
	b    []byte
	off  int
	err  error
}
```

</details>

## Opcode 2307

**`0x2307` — MyShopPrice**

Declarations: [proto.MyShopPrice](../../internal/proto/myshop.go#L18).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L255)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L77), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>MyShopPricePacket — internal/proto/myshop.go</summary>

[Source definition](../../internal/proto/myshop.go#L86)

```go
func MyShopPricePacket(index uint8, price uint32) []byte {
	w := NewWriter(MyShopPrice)
	w.U8(index)
	w.U32(price)
	return w.Bytes()
}
```

</details>

<details><summary>NewReader — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L28)

```go
func NewReader(payload []byte) *Reader {
	r := &Reader{b: payload, off: 2}
	if len(payload) >= 2 {
		r.wide = itemstack.Wide(binary.LittleEndian.Uint16(payload))
	}
	if len(payload) < 2 {
		r.err = ErrShort
	}
	return r
}
```

```go
type Reader struct {
	wide bool
	b    []byte
	off  int
	err  error
}
```

</details>

## Opcode 2308

**`0x2308` — MyShopSold**

Declarations: [proto.MyShopSold](../../internal/proto/myshop.go#L19).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523. to the seller: index, count, gold

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L86), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>NewReader — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L28)

```go
func NewReader(payload []byte) *Reader {
	r := &Reader{b: payload, off: 2}
	if len(payload) >= 2 {
		r.wide = itemstack.Wide(binary.LittleEndian.Uint16(payload))
	}
	if len(payload) < 2 {
		r.err = ErrShort
	}
	return r
}
```

```go
type Reader struct {
	wide bool
	b    []byte
	off  int
	err  error
}
```

</details>

## Opcode 2309

**`0x2309` — MyShopEnter**

Declarations: [proto.MyShopEnter](../../internal/proto/myshop.go#L20).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L265)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L97), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>MyShopEnterPacket — internal/proto/myshop.go</summary>

[Source definition](../../internal/proto/myshop.go#L93)

```go
func MyShopEnterPacket(seller uint32) []byte {
	w := NewWriter(MyShopEnter)
	w.U32(seller)
	return w.Bytes()
}
```

</details>

<details><summary>NewReader — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L28)

```go
func NewReader(payload []byte) *Reader {
	r := &Reader{b: payload, off: 2}
	if len(payload) >= 2 {
		r.wide = itemstack.Wide(binary.LittleEndian.Uint16(payload))
	}
	if len(payload) < 2 {
		r.err = ErrShort
	}
	return r
}
```

```go
type Reader struct {
	wide bool
	b    []byte
	off  int
	err  error
}
```

</details>

## Opcode 230a

**`0x230A` — MyShopLeave**

Declarations: [proto.MyShopLeave](../../internal/proto/myshop.go#L21).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L270)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L139), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

## Opcode 230b

**`0x230B` — MyShopList**

Declarations: [proto.MyShopList](../../internal/proto/myshop.go#L22).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [myShopList (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L420)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L107), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>myShopList — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L420)

```go
func (s *Session) myShopList() []byte {
	count := 0
	for _, sl := range s.shop.slots {
		if sl.used {
			count++
		}
	}
	w := packet.New(opMyShopList)
	w.U8(uint8(count))
	for idx, sl := range s.shop.slots {
		if !sl.used {
			continue
		}
		it := sl.item
		w.U8(uint8(idx))
		w.U32(sl.price)
		w.U8(it.Type)
		w.U8(it.TypeID)
		w.U8(uint8(min(it.Count, 255)))
		w.U16(it.Quality)
		w.Bytes(it.Gems[:])
		w.FixedString(it.CraftName, 21)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeMyShopList — internal/proto/myshop.go</summary>

[Source definition](../../internal/proto/myshop.go#L113)

```go
func DecodeMyShopList(p []byte) ([]MyShopItem, error) {
	r := NewReader(p)
	n := int(r.U8())
	if err := r.Err(); err != nil {
		return nil, err
	}
	const recordSize = 37
	if Op(p) != MyShopList || n > MyShopSlots || n*recordSize != r.Remaining() {
		return nil, fmt.Errorf("market list: invalid count %d", n)
	}
	out := make([]MyShopItem, n)
	var seen [MyShopSlots]bool
	for i := range out {
		e := &out[i]
		e.Index, e.Price = r.U8(), r.U32()
		e.Item.Type, e.Item.TypeID, e.Item.Count = r.U8(), r.U8(), uint16(r.U8())
		e.Item.Quality = r.U16()
		for j := range e.Item.Gems {
			e.Item.Gems[j] = r.U8()
		}
		e.Item.Craft = r.Fixed(21)
		if int(e.Index) >= MyShopSlots || seen[e.Index] || e.Price == 0 || e.Price > MyShopPriceMax || e.Item.Type == 0 || e.Item.Count == 0 {
			return nil, fmt.Errorf("market list: invalid item at index %d", e.Index)
		}
		seen[e.Index] = true
	}
	return out, r.Err()
}
```

```go
type MyShopItem struct {
	Index byte
	Price uint32
	Item  InventoryItem
}
```

</details>

## Opcode 230c

**`0x230C` — MyShopClosed**

Declarations: [proto.MyShopClosed](../../internal/proto/myshop.go#L23).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523. to buyers

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L139), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

## Opcode 230d

**`0x230D` — MyShopBuy**

Declarations: [proto.MyShopBuy](../../internal/proto/myshop.go#L24).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L177), [handleMyShop (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L273)

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L132), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>MyShopBuyPacket — internal/proto/myshop.go</summary>

[Source definition](../../internal/proto/myshop.go#L99)

```go
func MyShopBuyPacket(index, count uint8) []byte {
	w := NewWriter(MyShopBuy)
	w.U8(index)
	w.U8(count)
	return w.Bytes()
}
```

</details>

<details><summary>NewReader — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L28)

```go
func NewReader(payload []byte) *Reader {
	r := &Reader{b: payload, off: 2}
	if len(payload) >= 2 {
		r.wide = itemstack.Wide(binary.LittleEndian.Uint16(payload))
	}
	if len(payload) < 2 {
		r.err = ErrShort
	}
	return r
}
```

```go
type Reader struct {
	wide bool
	b    []byte
	off  int
	err  error
}
```

</details>

## Opcode 230e

**`0x230E` — MyShopStock**

Declarations: [proto.MyShopStock](../../internal/proto/myshop.go#L25).

Source notes: Personal markets (setup market): the request and reply opcodes of CUser::PacketMyShop (0x2301-0x230e) and the bubble broadcast 0x0523. to buyers: index, new count

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [shopAck (server/internal/world/myshop.go)](../../server/internal/world/myshop.go#L95)

**Client receive reducers:** [applyMyShop (internal/game/myshop.go)](../../internal/game/myshop.go#L119), [Apply (internal/game/state.go)](../../internal/game/state.go#L639)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>shopAck — server/internal/world/myshop.go</summary>

[Source definition](../../server/internal/world/myshop.go#L95)

```go
func shopAck(op packet.Opcode, body func(w *packet.Writer)) []byte {
	w := packet.New(op)
	if body != nil {
		body(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>NewReader — internal/proto/rw.go</summary>

[Source definition](../../internal/proto/rw.go#L28)

```go
func NewReader(payload []byte) *Reader {
	r := &Reader{b: payload, off: 2}
	if len(payload) >= 2 {
		r.wide = itemstack.Wide(binary.LittleEndian.Uint16(payload))
	}
	if len(payload) < 2 {
		r.err = ErrShort
	}
	return r
}
```

```go
type Reader struct {
	wide bool
	b    []byte
	off  int
	err  error
}
```

</details>
