# Protocol: Character actions and transport

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 0201

**`0x0201` — CharacterEnteredMap**

Declarations: [packet.CharacterEnteredMap](../../server/internal/packet/packet.go#L38), [proto.CharacterEnteredMap](../../internal/proto/opcodes.go#L22).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L223)

**S→C serializers:** [buildCharacterEnteredMapGuild (server/internal/world/character.go)](../../server/internal/world/character.go#L489)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1293), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L10)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharacterEnteredMapGuild — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L489)

```go
func buildCharacterEnteredMapGuild(ch *db.Character, isAdmin, firstLoad bool, guildID uint32) []byte {
	marker := uint8(0)
	if isAdmin {
		marker = 2
	} else if firstLoad {
		marker = 1
	}
	x, y, z := charPos(ch)
	w := packet.New(packet.CharacterEnteredMap)
	w.U32(charID(ch))
	w.U8(marker)
	w.U16(0)
	w.F32(x)
	w.F32(y)
	w.F32(z)
	w.U32(guildID)
	w.U32(0) // vehicle id
	return w.MustFinish()
}
```

</details>

<details><summary>MapLoadedPacket — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L60)

```go
func MapLoadedPacket() []byte { return NewWriter(CharacterEnteredMap).Bytes() }
```

</details>

<details><summary>DecodeUserEnter — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L232)

```go
func DecodeUserEnter(p []byte) (m UserEnter, err error) {
	r := NewReader(p)
	m.ID, m.Marker, m.Unk = r.U32(), r.U8(), r.U16()
	m.X, m.Y, m.Z = r.F32(), r.F32(), r.F32()
	m.GuildID, m.Vehicle = r.U32(), r.U32()
	return m, r.Err()
}
```

```go
type UserEnter struct {
	ID      uint32
	Marker  uint8 // 1 first load, 2 admin
	Unk     uint16
	X, Y, Z float32
	GuildID uint32
	Vehicle uint32
}
```

</details>

## Opcode 0202

**`0x0202` — UserLeave**

Declarations: [proto.UserLeave](../../internal/proto/opcodes.go#L231).

Source notes: Opcodes not named in Rapture's catalog. client handler 0x0059cf70 case 0x0202

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildUserLeave (server/internal/world/view.go)](../../server/internal/world/view.go#L234)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1293), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L38)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildUserLeave — server/internal/world/view.go</summary>

[Source definition](../../server/internal/world/view.go#L234)

```go
func buildUserLeave(id uint32) []byte {
	w := packet.New(0x0202)
	w.U32(id)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0204

**`0x0204` — ItemBagToBag**

Declarations: [packet.ItemBagToBag](../../server/internal/packet/packet.go#L85), [proto.ItemBagToBag](../../internal/proto/opcodes.go#L69).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L105)

**S→C serializers:** [buildBagToBag (server/internal/world/inventory.go)](../../server/internal/world/inventory.go#L79)

**Client receive reducers:** [questMapInventoryChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L48), [Apply (internal/game/state.go)](../../internal/game/state.go#L997), [applyInventoryPacket (internal/game/state_inventory_packets.go)](../../internal/game/state_inventory_packets.go#L31)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildBagToBag — server/internal/world/inventory.go</summary>

[Source definition](../../server/internal/world/inventory.go#L79)

```go
func buildBagToBag(srcBag, srcSlot uint8, src *db.Item, destBag, destSlot uint8, dest *db.Item) []byte {
	w := packet.New(packet.ItemBagToBag)
	writeSlot(w, srcBag, srcSlot, src)
	writeSlot(w, destBag, destSlot, dest)
	return w.MustFinish()
}
```

</details>

<details><summary>ItemMovePacket — internal/proto/inventory.go</summary>

[Source definition](../../internal/proto/inventory.go#L60)

```go
func ItemMovePacket(srcBag, srcSlot, dstBag, dstSlot uint8) []byte {
	w := NewWriter(ItemBagToBag)
	w.Raw([]byte{srcBag, srcSlot, dstBag, dstSlot})
	return w.Bytes()
}
```

</details>

<details><summary>DecodeWarehouseMove — internal/proto/warehouse.go</summary>

[Source definition](../../internal/proto/warehouse.go#L45)

```go
func DecodeWarehouseMove(p []byte) (items [2]InventoryItem, money uint32, hasMoney bool, err error) {
	r := NewReader(p)
	items = [2]InventoryItem{readItemSlot(r), readItemSlot(r)}
	if err := r.Err(); err != nil {
		return items, 0, false, err
	}
	for _, item := range items {
		if item.Bag == GuildBankBag && int(item.Slot) >= GuildBankSlots {
			return items, 0, false, fmt.Errorf("guild bank: invalid slot %d", item.Slot)
		}
	}
	if (items[0].Bag == WarehouseBag && items[1].Bag != WarehouseBag) || (items[0].Bag == GuildBankBag && items[1].Bag != GuildBankBag) {
		money, hasMoney = r.U32(), true
	}
	return items, money, hasMoney, r.Err()
}
```

</details>

## Opcode 0205

**`0x0205` — AddItem**

Declarations: [packet.AddItem](../../server/internal/packet/packet.go#L86), [proto.AddItem](../../internal/proto/opcodes.go#L70).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L103)

**S→C serializers:** [buildAddItem (server/internal/world/inventory.go)](../../server/internal/world/inventory.go#L86), [buildAddItemError (server/internal/world/inventory.go)](../../server/internal/world/inventory.go#L98), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [questMapInventoryChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L48), [Apply (internal/game/state.go)](../../internal/game/state.go#L997), [applyInventoryPacket (internal/game/state_inventory_packets.go)](../../internal/game/state_inventory_packets.go#L52)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildAddItem — server/internal/world/inventory.go</summary>

[Source definition](../../server/internal/world/inventory.go#L86)

```go
func buildAddItem(it *db.Item) []byte {
	w := packet.New(packet.AddItem)
	writeSlot(w, it.Bag, it.Slot, it)
	return w.MustFinish()
}
```

</details>

<details><summary>buildAddItemError — server/internal/world/inventory.go</summary>

[Source definition](../../server/internal/world/inventory.go#L98)

```go
func buildAddItemError(code uint8) []byte {
	w := packet.New(packet.AddItem)
	w.U8(0)
	w.U8(code)
	return w.MustFinish()
}
```

</details>

<details><summary>idPacket — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L39)

```go
func idPacket(op Opcode, id uint32) []byte {
	w := NewWriter(op)
	w.U32(id)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeItemAdd — internal/proto/inventory.go</summary>

[Source definition](../../internal/proto/inventory.go#L52)

```go
func DecodeItemAdd(p []byte) (InventoryItem, error) {
	r := NewReader(p)
	m := readItemSlot(r)
	return m, r.Err()
}
```

```go
type InventoryItem struct {
	Bag, Slot, Type, TypeID uint8
	Quality                 uint16
	Gems                    [6]uint8
	Count                   uint16
	Craft                   string
}
```

</details>

## Opcode 0206

**`0x0206` — RemoveItem**

Declarations: [packet.RemoveItem](../../server/internal/packet/packet.go#L87), [proto.RemoveItem](../../internal/proto/opcodes.go#L71).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L101)

**S→C serializers:** [buildRemoveItem (server/internal/world/inventory.go)](../../server/internal/world/inventory.go#L118)

**Client receive reducers:** [questMapInventoryChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L48), [Apply (internal/game/state.go)](../../internal/game/state.go#L997), [applyInventoryPacket (internal/game/state_inventory_packets.go)](../../internal/game/state_inventory_packets.go#L66)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildRemoveItem — server/internal/world/inventory.go</summary>

[Source definition](../../server/internal/world/inventory.go#L118)

```go
func buildRemoveItem(it *db.Item, full bool) []byte {
	w := packet.New(packet.RemoveItem)
	w.U8(it.Bag)
	w.U8(it.Slot)
	if full {
		w.Bytes(make([]byte, 2))
		w.Count(0)
	} else {
		w.Bytes([]byte{it.Type, it.TypeID})
		w.Count(it.Count)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>ItemDropPacket — internal/proto/shop.go</summary>

[Source definition](../../internal/proto/shop.go#L35)

```go
func ItemDropPacket(bag, slot uint8, count uint16) []byte {
	w := NewWriter(RemoveItem)
	w.Raw([]byte{bag, slot})
	w.Count(count)
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

## Opcode 0207

**`0x0207` — ExpGain**

Declarations: [packet.ExpGain](../../server/internal/packet/packet.go#L41), [proto.ExpGain](../../internal/proto/opcodes.go#L25).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildExpGain (server/internal/world/combat.go)](../../server/internal/world/combat.go#L649)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L958)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildExpGain — server/internal/world/combat.go</summary>

[Source definition](../../server/internal/world/combat.go#L649)

```go
func buildExpGain(credited experienceCredit) []byte {
	w := packet.New(packet.ExpGain)
	w.U32(credited.base)
	w.U32(credited.perfect)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeExperienceGain — internal/proto/experience.go</summary>

[Source definition](../../internal/proto/experience.go#L7)

```go
func DecodeExperienceGain(p []byte) (ExperienceGain, error) {
	r := NewReader(p)
	gain := ExperienceGain{Base: r.U32(), Perfect: r.U32()}
	return gain, r.Err()
}
```

```go
type ExperienceGain struct{ Base, Perfect uint32 }
```

</details>

## Opcode 0208

**`0x0208` — SetStatPoint**

Declarations: [packet.SetStatPoint](../../server/internal/packet/packet.go#L88), [proto.SetStatPoint](../../internal/proto/opcodes.go#L72).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L227)

**S→C serializers:** [handleSetStatPoint (server/internal/world/skills.go)](../../server/internal/world/skills.go#L265)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L680)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleSetStatPoint — server/internal/world/skills.go</summary>

[Source definition](../../server/internal/world/skills.go#L265)

```go
func (s *Session) handleSetStatPoint(p []byte) {
	if len(p) < 14 || s.char == nil {
		return
	}
	ch := s.char
	var d [6]uint16
	total := uint32(0)
	for i := range d {
		d[i], _ = packet.U16(p[2+2*i:])
		total += uint32(d[i])
	}
	if total == 0 || total > uint32(ch.StatPoint) {
		s.log.Printf("GX Server World rejected stat points total=%d available=%d", total, ch.StatPoint)
		return
	}
	// A stat set to the DB cap by a GM must remain persistible after a
	// normal allocation too. Refuse the whole request before spending points.
	for i, base := range []uint16{ch.Str, ch.Dex, ch.Rec, ch.Int, ch.Wis, ch.Luc} {
		if d[i] > 0 && uint32(base)+uint32(d[i]) > db.BaseStatCap {
			s.log.Printf("GX Server World rejected stat allocation beyond base-stat limit")
			return
		}
	}
	add := func(v *uint16, n uint16) { *v += n }
	add(&ch.Str, d[0])
	add(&ch.Dex, d[1])
	add(&ch.Rec, d[2])
	add(&ch.Int, d[3])
	add(&ch.Wis, d[4])
	add(&ch.Luc, d[5])
	ch.StatPoint -= uint16(total)
	s.persist()
	stats := s.stats()
	w := packet.New(packet.SetStatPoint)
	for _, v := range d {
		w.U16(v)
	}
	s.send(w.MustFinish(), "0x0208 stat point applied")
	s.send(buildAdditionalStats(&stats), "0x0526 additional stats after stat allocation")
	if d[2] > 0 {
		s.send(buildMaxEnergy(ch, 0, stats.MaxHP), "0x050B max hp")
	}
	if d[1] > 0 {
		s.send(buildMaxEnergy(ch, 1, stats.MaxSP), "0x050B max sp")
	}
	if d[4] > 0 {
		s.send(buildMaxEnergy(ch, 2, stats.MaxMP), "0x050B max mp")
	}
	s.log.Printf("GX Server World applied stat points character_id=%d deltas=%v remaining=%d", ch.ID, d, ch.StatPoint)
}
```

</details>

<details><summary>StatPointsPacket — internal/proto/stats.go</summary>

[Source definition](../../internal/proto/stats.go#L15)

```go
func StatPointsPacket(d StatPoints) []byte {
	w := NewWriter(SetStatPoint)
	for _, n := range d {
		w.U16(n)
	}
	return w.Bytes()
}
```

```go
type StatPoints [6]uint16
```

</details>

<details><summary>DecodeStatPoints — internal/proto/stats.go</summary>

[Source definition](../../internal/proto/stats.go#L23)

```go
func DecodeStatPoints(p []byte) (d StatPoints, err error) {
	r := NewReader(p)
	for i := range d {
		d[i] = r.U16()
	}
	return d, r.Err()
}
```

</details>

## Opcode 0209

**`0x0209` — LearnNewSkill**

Declarations: [packet.LearnNewSkill](../../server/internal/packet/packet.go#L90), [proto.LearnNewSkill](../../internal/proto/opcodes.go#L74).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L225)

**S→C serializers:** [buildLearnSkillError (server/internal/world/skills.go)](../../server/internal/world/skills.go#L130), [buildLearnSkillResult (server/internal/world/skills.go)](../../server/internal/world/skills.go#L159)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L777)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildLearnSkillError — server/internal/world/skills.go</summary>

[Source definition](../../server/internal/world/skills.go#L130)

```go
func buildLearnSkillError(reason uint8, skillID uint16, level uint8) []byte {
	w := packet.New(packet.LearnNewSkill)
	w.U8(reason)
	w.U8(0)
	w.U16(skillID)
	w.U8(level)
	return w.MustFinish()
}
```

</details>

<details><summary>buildLearnSkillResult — server/internal/world/skills.go</summary>

[Source definition](../../server/internal/world/skills.go#L159)

```go
func buildLearnSkillResult(ok bool, number uint8, skillID uint16, level uint8) []byte {
	w := packet.New(packet.LearnNewSkill)
	w.U8(resultByte(ok))
	w.U8(number)
	w.U16(skillID)
	w.U8(level)
	return w.MustFinish()
}
```

</details>

<details><summary>LearnSkillPacket — internal/proto/skills.go</summary>

[Source definition](../../internal/proto/skills.go#L81)

```go
func LearnSkillPacket(id uint16, level uint8) []byte {
	w := NewWriter(LearnNewSkill)
	w.U16(id)
	w.U8(level)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeLearnResult — internal/proto/skills.go</summary>

[Source definition](../../internal/proto/skills.go#L74)

```go
func DecodeLearnResult(p []byte) (m LearnResult, err error) {
	r := NewReader(p)
	m.Result, m.Number, m.ID, m.Level = r.U8(), r.U8(), r.U16(), r.U8()
	return m, r.Err()
}
```

```go
type LearnResult struct {
	Result, Number uint8
	ID             uint16
	Level          uint8
}
```

</details>

## Opcode 020a

**`0x020A` — CharacterEnteredPortal**

Declarations: [packet.CharacterEnteredPortal](../../server/internal/packet/packet.go#L91), [proto.CharacterEnteredPortal](../../internal/proto/opcodes.go#L75).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L219)

**S→C serializers:** [buildPortalResult (server/internal/world/partyzone.go)](../../server/internal/world/partyzone.go#L84)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildPortalResult — server/internal/world/partyzone.go</summary>

[Source definition](../../server/internal/world/partyzone.go#L84)

```go
func buildPortalResult(index, result uint8) []byte {
	w := packet.New(packet.CharacterEnteredPortal)
	w.U8(index)
	w.U8(result)
	return w.MustFinish()
}
```

</details>

<details><summary>PortalRequestPacket — internal/proto/portal.go</summary>

[Source definition](../../internal/proto/portal.go#L12)

```go
func PortalRequestPacket(index uint8) []byte {
	w := NewWriter(CharacterEnteredPortal)
	w.U8(index)
	return w.Bytes()
}
```

</details>

## Opcode 020b

**`0x020B` — CharacterMapTeleport**

Declarations: [packet.CharacterMapTeleport](../../server/internal/packet/packet.go#L92), [proto.CharacterMapTeleport](../../internal/proto/opcodes.go#L76).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildCharacterMapPosition (server/internal/world/character.go)](../../server/internal/world/character.go#L509)

**Client receive reducers:** [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [Apply (internal/game/state.go)](../../internal/game/state.go#L1202)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharacterMapPosition — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L509)

```go
func buildCharacterMapPosition(ch *db.Character) []byte {
	mapID := uint16(0)
	if ch != nil {
		mapID = ch.MapID
	}
	x, y, z := charPos(ch)
	w := packet.New(packet.CharacterMapTeleport)
	w.U32(charID(ch))
	w.U16(realMap(mapID))
	w.F32(x)
	w.F32(y)
	w.F32(z)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeMapPosition — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L401)

```go
func DecodeMapPosition(p []byte) (m MapPosition, err error) {
	r := NewReader(p)
	m.ID, m.MapID, m.X, m.Y, m.Z = r.U32(), r.U16(), r.F32(), r.F32(), r.F32()
	return m, r.Err()
}
```

```go
type MapPosition struct {
	ID      uint32
	MapID   uint16
	X, Y, Z float32
}
```

</details>

## Opcode 020c

**`0x020C` — Gatekeeper**

Declarations: [packet.Gatekeeper](../../server/internal/packet/packet.go#L77), [proto.Gatekeeper](../../internal/proto/opcodes.go#L61).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L179)

**S→C serializers:** [buildGatekeeperResult (server/internal/world/npcquest.go)](../../server/internal/world/npcquest.go#L224)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L503)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGatekeeperResult — server/internal/world/npcquest.go</summary>

[Source definition](../../server/internal/world/npcquest.go#L224)

```go
func buildGatekeeperResult(ok bool, money uint32) []byte {
	w := packet.New(packet.Gatekeeper)
	w.Bool(ok)
	w.U32(money)
	return w.MustFinish()
}
```

</details>

<details><summary>GatekeeperPacket — internal/proto/gatekeeper.go</summary>

[Source definition](../../internal/proto/gatekeeper.go#L21)

```go
func GatekeeperPacket(id uint32, destination uint8) []byte {
	w := NewWriter(Gatekeeper)
	w.U32(id)
	w.U8(destination)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeGatekeeperResult — internal/proto/gatekeeper.go</summary>

[Source definition](../../internal/proto/gatekeeper.go#L10)

```go
func DecodeGatekeeperResult(p []byte) (GatekeeperResult, error) {
	r := NewReader(p)
	v := GatekeeperResult{Result: r.U8(), Money: r.U32()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if v.Result > 1 {
		return v, fmt.Errorf("invalid gatekeeper result")
	}
	return v, nil
}
```

```go
type GatekeeperResult struct {
	Result uint8
	Money  uint32
}
```

</details>

## Opcode 020e

**`0x020E` — UserKillCountUpdate**

Declarations: [packet.UserKillCountUpdate](../../server/internal/packet/packet.go#L93), [proto.UserKillCountUpdate](../../internal/proto/opcodes.go#L77).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildKillCountUpdate (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L67)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L354)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildKillCountUpdate — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L67)

```go
func buildKillCountUpdate(index uint8, value uint32) []byte {
	w := packet.New(packet.UserKillCountUpdate)
	w.U8(index)
	w.U32(value)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodePvPCountUpdate — internal/proto/pvp_count.go</summary>

[Source definition](../../internal/proto/pvp_count.go#L12)

```go
func DecodePvPCountUpdate(p []byte) (PvPCountUpdate, error) {
	r := NewReader(p)
	v := PvPCountUpdate{Index: r.U8(), Value: r.U32()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if r.Remaining() != 0 {
		return v, fmt.Errorf("pvp count: %d trailing bytes", r.Remaining())
	}
	return v, nil
}
```

```go
type PvPCountUpdate struct {
	Index uint8
	Value uint32
}
```

</details>

## Opcode 020f

**`0x020F` — CountryKillCount**

Declarations: [proto.CountryKillCount](../../internal/proto/opcodes.go#L78).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [kcBroadcast (server/internal/world/killcount.go)](../../server/internal/world/killcount.go#L241)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L323)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>kcBroadcast — server/internal/world/killcount.go</summary>

[Source definition](../../server/internal/world/killcount.go#L241)

```go
func kcBroadcast(c int, snapshot bool) {
	kcSendMu.Lock()
	defer kcSendMu.Unlock()
	kcMu.Lock()
	count := min(kcWorld[c].count, kcMaxKill())
	var details []byte
	if snapshot {
		details = blessingSnapshotLocked(uint8(c), false)
	}
	kcMu.Unlock()
	w := packet.New(0x020F)
	w.U8(uint8(c))
	w.U32(count)
	frame := w.MustFinish()
	for _, p := range AllPlayers() {
		if int(p.Faction()) == c {
			p.Send(frame, "0x020F country kill count")
			if details != nil {
				p.Send(details, "GX blessing snapshot")
			}
		}
	}
}
```

</details>

<details><summary>DecodeBlessing — internal/proto/blessing.go</summary>

[Source definition](../../internal/proto/blessing.go#L12)

```go
func DecodeBlessing(p []byte) (BlessingUpdate, error) {
	r := NewReader(p)
	b := BlessingUpdate{Country: r.U8(), Count: r.U32()}
	if Op(p) == BlessInit {
		b.RemainingMS = r.U32()
	}
	if err := r.Err(); err != nil {
		return b, err
	}
	if b.Country > 1 || r.Remaining() != 0 {
		return b, fmt.Errorf("invalid blessing country or trailing bytes")
	}
	return b, nil
}
```

```go
type BlessingUpdate struct {
	Country            uint8
	Count, RemainingMS uint32
}
```

</details>

## Opcode 0211

**`0x0211` — BlessInit**

Declarations: [packet.BlessInit](../../server/internal/packet/packet.go#L132), [proto.BlessInit](../../internal/proto/opcodes.go#L117).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildKCStatus (server/internal/world/killcount.go)](../../server/internal/world/killcount.go#L275), [buildBlessInit (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L80)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L323)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildKCStatus — server/internal/world/killcount.go</summary>

[Source definition](../../server/internal/world/killcount.go#L275)

```go
func buildKCStatus(c uint8) []byte {
	kcMu.Lock()
	var count uint32
	var left int64
	if c < 2 {
		count = kcWorld[c].count
		if r := kcWorld[c].resetAt; !r.IsZero() {
			left = max(time.Until(r).Milliseconds(), 0)
		}
	}
	kcMu.Unlock()
	w := packet.New(0x0211)
	w.U8(c)
	w.U32(count)
	w.U32(uint32(left))
	return w.MustFinish()
}
```

</details>

<details><summary>buildBlessInit — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L80)

```go
func buildBlessInit(faction uint8) []byte {
	w := packet.New(packet.BlessInit)
	w.U8(faction)
	w.I32(0)
	w.U32(0)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeBlessing — internal/proto/blessing.go</summary>

[Source definition](../../internal/proto/blessing.go#L12)

```go
func DecodeBlessing(p []byte) (BlessingUpdate, error) {
	r := NewReader(p)
	b := BlessingUpdate{Country: r.U8(), Count: r.U32()}
	if Op(p) == BlessInit {
		b.RemainingMS = r.U32()
	}
	if err := r.Err(); err != nil {
		return b, err
	}
	if b.Country > 1 || r.Remaining() != 0 {
		return b, fmt.Errorf("invalid blessing country or trailing bytes")
	}
	return b, nil
}
```

```go
type BlessingUpdate struct {
	Country            uint8
	Count, RemainingMS uint32
}
```

</details>

## Opcode 0212

**`0x0212` — AutoAttackStop**

Declarations: [packet.AutoAttackStop](../../server/internal/packet/packet.go#L39), [proto.AutoAttackStop](../../internal/proto/opcodes.go#L23).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L231)

**S→C serializers:** [beginSkill (server/internal/world/skillcast.go)](../../server/internal/world/skillcast.go#L198), [syncVisibility (server/internal/world/visibility.go)](../../server/internal/world/visibility.go#L40)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>beginSkill — server/internal/world/skillcast.go</summary>

[Source definition](../../server/internal/world/skillcast.go#L198)

```go
func (s *Session) beginSkill(op packet.Opcode, k db.CharacterSkill, def *db.SkillDef, target, evolution uint32, p []byte) {
	if !s.attackMovementAllowed() {
		return
	}
	ch := s.char
	number := k.Number
	if def.TargetType == proto.SkillTargetRaid && !proto.SupportsRaidSkillTarget(def.TypeEffect) {
		s.sendSkillMessage(k.SkillID, 317)
		return
	}
	if !def.MobSkill && !s.skillWeaponAllowed(def) {
		s.sendSkillMessage(k.SkillID, msgUnavailableEquipment)
		return
	}
	s.autoAttack = false
	if !s.skillActionReady(def, time.Now()) {
		if p == nil || !s.queueSpinCombo(def, op, p) {
			if evolution == 0 && s.holdSkillUse(def, op, p, time.Now()) {
				return
			}
			s.sendSkillMessage(k.SkillID, 215)
		}
		return
	}
	if kind := formOf(def); kind == proto.FormMonsterCopy || kind == proto.FormDisguise {
		if _, ok := s.copyFormTarget(def, target, false); !ok || !s.canAddEffect(def, ch.ID) {
			s.rejectCopyForm(op, k)
			return
		}
	} else {
		target = s.friendlySkillRecipient(def, target)
	}
	castTime := time.Duration(def.ReadyTime) * 250 * time.Millisecond
	if castTime == 0 {
		s.resolveSkillDefinition(target, k, def, op, nil)
		return
	}
	if s.safeZoneBlocksSkill(def, op, target) {
		s.rejectSafeZoneSkill(k, false)
		return
	}
	if !s.definitionReady(def) {
		s.sendSkillMessage(k.SkillID, 215)
		return
	}
	if !def.MobSkill && !s.previousSkillCooling(def) {
		s.sendSkillMessage(k.SkillID, msgPreviousSkillRequired)
		return
	}
	stats := s.stats()
	if s.cannotAct() || stats.Disabled[min(int(def.Kind), 2)] > 0 {
		s.log.Printf("GX Server World rejected skill use: category disabled skill_id=%d", k.SkillID)
		return
	}
	if sp, mp := s.skillCosts(def, &stats); stats.CurrentMP < mp || stats.CurrentSP < sp {
		s.log.Printf("GX Server World rejected skill use: not enough MP/SP skill_id=%d", k.SkillID)
		return
	}
	castTarget := target
	if def.TypeEffect == typeEffectHeal || def.TypeEffect == typeEffectBuff {
		castTarget = ch.ID
		if def.TargetType == targetTypeOther || def.TargetType == targetTypeOtherAlt {
			targets, _, ok := s.friendlyTargets(def, target)
			if !ok {
				s.sendFriendlySkillFailure(k, def, target)
				return
			}
			castTarget = targets[0].CharacterID
		}
	}
	var pos *[2]float32
	var corpse corpseTarget
	if def.TypeEffect == typeEffectOnce {
		p, ok := s.userSkillTarget(def, target)
		if def.TypeDetail == detailKillSoul {
			corpse = s.soulBurnCorpse(p)
			if corpse.player == nil {
				s.rejectSoulBurn(k)
				return
			}
		}
		if !ok {
			return
		}
		if !s.inUserSkillRange(p, def) {
			s.sendUserSkillRangeFailure(k, p)
			return
		}
		castTarget = p.CharacterID
	} else if op == opUserSkillPvP && (def.TypeEffect == typeEffectAttack || def.TypeEffect == typeEffectDebuff) {
		// A hostile skill on a user: ChkTargetUser before the cast.
		if def.TargetType != targetTypeEnemiesNearCaster {
			p := s.hostileUser(target)
			if p == nil {
				s.send(buildSkillUseResult(packet.SkillUseNoTarget, skillResultNotTarget, ch.ID, 0, k.SkillID, k.Level, 0, 0, 0),
					"0x0511 skill target refused")
				return
			}
			if !s.inUserSkillRange(p, def) {
				s.sendUserSkillRangeFailure(k, p)
				return
			}
			x, z := p.PositionXZ()
			castTarget, pos = p.CharacterID, &[2]float32{x, z}
		}
	} else if (def.TypeEffect == typeEffectAttack || def.TypeEffect == typeEffectDebuff) && castTarget != ch.ID && def.TargetType != targetTypeEnemiesNearCaster {
		if def.TypeEffect == typeEffectDebuff && def.TypeDetail == detailDegeneration {
			s.sendSkillMessage(k.SkillID, 317) // it cannot be cast upon a monster
			return
		}
		mob, ok := MobByID(ch.MapID, target)
		if !ok {
			return
		}
		if mobFriendly(db.MustDefs().Mobs[mob.MobID], s.faction) {
			s.sendSkillMessage(k.SkillID, 317)
			return
		}
		if !s.inSkillRange(&mob, def, &stats) {
			s.sendMobSkillRangeFailure(k, &mob)
			return
		}
		pos = &[2]float32{mob.PosX, mob.PosZ}
	}
	if def.TypeEffect == typeEffectAttack || def.TypeEffect == typeEffectDebuff {
		s.breakCombatForm()
		s.breakStealth()
	}
	s.send(packet.New(packet.AutoAttackStop).MustFinish(), "0x0212 auto attack stop (casting)")
	w := packet.New(packet.SkillCastStarted)
	w.U32(ch.ID)
	w.U32(castTarget)
	w.U16(k.SkillID)
	w.U8(k.Level)
	s.broadcastNear(w.MustFinish(), "0x0510 skill cast started", true)
	if pos == nil {
		pos = &[2]float32{}
	}
	now := time.Now()
	s.lastActionAt = now
	s.cast = &pendingCast{targetMobID: target, number: number, op: op, readyAt: now.Add(castTime), targetPos: pos, corpse: corpse, evolution: evolution}
	s.autoAttack = false
	s.syncPlayerSpeed(false)
	s.log.Printf("GX Server World skill cast started skill_id=%d level=%d cast_time=%s", k.SkillID, k.Level, castTime)
}
```

</details>

<details><summary>syncVisibility — server/internal/world/visibility.go</summary>

[Source definition](../../server/internal/world/visibility.go#L40)

```go
func (s *Session) syncVisibility() {
	p := s.player
	if p == nil {
		return
	}
	hidden, noTarget := s.isAdmin && s.gmConcealed, false
	for _, b := range s.buffs {
		if b.Effect != typeEffectBuff {
			continue
		}
		if b.Detail == 88 {
			hidden = true
		}
		if b.Detail == 89 {
			noTarget = true
		}
	}
	p.mu.Lock()
	gmChanged := p.gmVisible != (!s.isAdmin || !s.gmConcealed)
	p.gmVisible = !s.isAdmin || !s.gmConcealed
	p.mu.Unlock()
	changed := p.hidden.Swap(hidden) != hidden || gmChanged
	p.noMobTarget.Store(hidden || noTarget)
	p.immune.Store(s.invulnerable())
	if changed {
		if hidden {
			s.autoAttack, s.cast = false, nil
			s.send(packet.New(packet.AutoAttackStop).MustFinish(), "stealth attack stop")
		}
		s.refreshUserView()
		frame := p.visibilityPacket()
		s.send(frame, "GX visibility")
		s.sendToViewers(frame, "GX visibility")
	}
}
```

</details>

<details><summary>StopAttackPacket — internal/proto/combat.go</summary>

[Source definition](../../internal/proto/combat.go#L80)

```go
func StopAttackPacket() []byte { return NewWriter(AutoAttackStop).Bytes() }
```

</details>

## Opcode 0213

**`0x0213` — SetMoney**

Declarations: [packet.SetMoney](../../server/internal/packet/packet.go#L40), [proto.SetMoney](../../internal/proto/opcodes.go#L24).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildSetMoney (server/internal/world/inventory.go)](../../server/internal/world/inventory.go#L92)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L997), [applyInventoryPacket (internal/game/state_inventory_packets.go)](../../internal/game/state_inventory_packets.go#L76)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildSetMoney — server/internal/world/inventory.go</summary>

[Source definition](../../server/internal/world/inventory.go#L92)

```go
func buildSetMoney(money uint32) []byte {
	w := packet.New(packet.SetMoney)
	w.U32(money)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0214

**`0x0214` — StatsReset**

Declarations: [packet.StatsReset](../../server/internal/packet/packet.go#L89), [proto.StatsReset](../../internal/proto/opcodes.go#L73).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildStatsReset (server/internal/world/character.go)](../../server/internal/world/character.go#L449)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L686)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildStatsReset — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L449)

```go
func buildStatsReset(statPoint, str, rec, intel, wis, dex, luc uint16) []byte {
	w := packet.New(packet.StatsReset)
	w.U8(1)
	for _, v := range []uint16{statPoint, str, rec, intel, wis, dex, luc} {
		w.U16(v)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeStatReset — internal/proto/stats.go</summary>

[Source definition](../../internal/proto/stats.go#L54)

```go
func DecodeStatReset(p []byte) (s StatReset, err error) {
	r := NewReader(p)
	s.Result = r.U8()
	if s.Result != 1 {
		return s, r.Err()
	}
	s.Points, s.Str, s.Rec, s.Int, s.Wis, s.Dex, s.Luc = r.U16(), r.U16(), r.U16(), r.U16(), r.U16(), r.U16(), r.U16()
	return s, r.Err()
}
```

```go
type StatReset struct {
	Result                       uint8
	Points                       uint16
	Str, Rec, Int, Wis, Dex, Luc uint16
}
```

</details>

## Opcode 0216

**`0x0216` — UseVehicle**

Declarations: [packet.UseVehicle](../../server/internal/packet/packet.go#L84), [proto.UseVehicle](../../internal/proto/opcodes.go#L68).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L23), [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L243)

**S→C serializers:** [vehicleResult (server/internal/world/vehicle.go)](../../server/internal/world/vehicle.go#L108)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L467)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>vehicleResult — server/internal/world/vehicle.go</summary>

[Source definition](../../server/internal/world/vehicle.go#L108)

```go
func (s *Session) vehicleResult(ok, on bool) {
	w := packet.New(packet.UseVehicle)
	w.Bool(ok)
	w.Bool(on)
	s.send(w.MustFinish(), "0x0216 vehicle result")
}
```

</details>

<details><summary>UseVehiclePacket — internal/proto/vehicle.go</summary>

[Source definition](../../internal/proto/vehicle.go#L29)

```go
func UseVehiclePacket() []byte { return NewWriter(UseVehicle).Bytes() }
```

</details>

<details><summary>DecodeVehicleResult — internal/proto/vehicle.go</summary>

[Source definition](../../internal/proto/vehicle.go#L7)

```go
func DecodeVehicleResult(p []byte) (VehicleResult, error) {
	r := NewReader(p)
	v := VehicleResult{Result: r.U8(), On: r.U8()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if v.Result > 1 || v.On > 1 {
		return v, fmt.Errorf("invalid vehicle result")
	}
	return v, nil
}
```

```go
type VehicleResult struct{ Result, On uint8 }
```

</details>

## Opcode 0217

**`0x0217` — VehicleCast**

Declarations: [proto.VehicleCast](../../internal/proto/mount_state.go#L8).

Source notes: Native requests and invitation messages. 0x021C is also the server's passenger attachment notification; direction determines its layout.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [handleUseVehicle (server/internal/world/vehicle.go)](../../server/internal/world/vehicle.go#L134)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L463)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleUseVehicle — server/internal/world/vehicle.go</summary>

[Source definition](../../server/internal/world/vehicle.go#L134)

```go
func (s *Session) handleUseVehicle() {
	if s.char == nil || s.player == nil {
		return
	}
	v := s.player.vehicleState()
	if v.State != proto.VehicleOff {
		s.dismount()
		s.vehicleResult(true, false)
		return
	}
	it, d := s.mountDefinition()
	vehicleMu.Lock()
	cooldown := vehicleCooldowns[s.char.ID]
	vehicleMu.Unlock()
	if !validMount(d) || mountForbiddenMap(s.char.MapID) || s.dead || s.cast != nil || s.player.IsSitting() || s.movement.ladder.Active || s.vehicleBlocked() || time.Now().Before(cooldown) || !canEquip(s.char, d, mountSlot, s.items) {
		s.vehicleResult(false, false)
		return
	}
	s.cancelMove()
	s.autoAttack = false
	s.mountItemUID = it.UID
	s.movement.motion = 0
	s.movement.credit = 0
	vehicleMu.Lock()
	s.player.ride.state = proto.VehicleState{ID: s.char.ID, State: proto.VehicleCasting, Race: s.char.Race, Model: uint8(d.Range), Speed: d.Speed, Seats: d.AttackTime}
	s.player.ride.castEnd = time.Now().Add(vehicleCastDuration)
	vehicleMu.Unlock()
	w := packet.New(packet.Opcode(proto.VehicleCast))
	w.U32(s.char.ID)
	s.send(w.MustFinish(), "0x0217 mount cast")
	s.sendToViewers(w.MustFinish(), "0x0217 mount cast")
	publishVehicle(s.player)
}
```

</details>

## Opcode 0218

**`0x0218` — PvPRewardStatus**

Declarations: [proto.PvPRewardStatus](../../internal/proto/pvp_reward.go#L10).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L124)

**S→C serializers:** [buildPvPRewardStatus (server/internal/world/pvpreward.go)](../../server/internal/world/pvpreward.go#L46)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L329)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildPvPRewardStatus — server/internal/world/pvpreward.go</summary>

[Source definition](../../server/internal/world/pvpreward.go#L46)

```go
func buildPvPRewardStatus(ch *db.Character) []byte {
	w := packet.New(packet.Opcode(proto.PvPRewardStatus))
	w.U8(min(ch.KillLevel, pvprank.RewardCount) + 1)
	w.U8(min(ch.DeadLevel, pvprank.RewardCount) + 1)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodePvPRewards — internal/proto/pvp_reward.go</summary>

[Source definition](../../internal/proto/pvp_reward.go#L26)

```go
func DecodePvPRewards(p []byte) (PvPRewards, error) {
	r := NewReader(p)
	v := PvPRewards{KillStage: r.U8(), DeathStage: r.U8(), Known: true}
	if err := r.Err(); err != nil {
		return v, err
	}
	if r.Remaining() != 0 || v.KillStage < 1 || v.KillStage > PvPRewardFinishedStage || v.DeathStage < 1 || v.DeathStage > PvPRewardFinishedStage {
		return v, fmt.Errorf("invalid PvP reward stages")
	}
	return v, nil
}
```

```go
type PvPRewards struct {
	KillStage, DeathStage uint8
	Known                 bool
}
```

</details>

## Opcode 0219

**`0x0219` — PvPRewardKill**

Declarations: [proto.PvPRewardKill](../../internal/proto/pvp_reward.go#L11).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L124)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L335)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodePvPRewardResult — internal/proto/pvp_reward.go</summary>

[Source definition](../../internal/proto/pvp_reward.go#L39)

```go
func DecodePvPRewardResult(p []byte, death bool) (PvPRewardResult, error) {
	r := NewReader(p)
	result := r.U8()
	v := PvPRewardResult{OK: result == 1}
	if death {
		v.Value = r.U32()
	} else {
		v.Value = uint32(r.U8())
	}
	if err := r.Err(); err != nil {
		return v, err
	}
	if result > 1 || r.Remaining() != 0 {
		return v, fmt.Errorf("invalid PvP reward result")
	}
	return v, nil
}
```

```go
type PvPRewardResult struct {
	OK    bool
	Value uint32
}
```

</details>

## Opcode 021a

**`0x021A` — PvPRewardDeath**

Declarations: [proto.PvPRewardDeath](../../internal/proto/pvp_reward.go#L12).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L124)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L335)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodePvPRewardResult — internal/proto/pvp_reward.go</summary>

[Source definition](../../internal/proto/pvp_reward.go#L39)

```go
func DecodePvPRewardResult(p []byte, death bool) (PvPRewardResult, error) {
	r := NewReader(p)
	result := r.U8()
	v := PvPRewardResult{OK: result == 1}
	if death {
		v.Value = r.U32()
	} else {
		v.Value = uint32(r.U8())
	}
	if err := r.Err(); err != nil {
		return v, err
	}
	if result > 1 || r.Remaining() != 0 {
		return v, fmt.Errorf("invalid PvP reward result")
	}
	return v, nil
}
```

```go
type PvPRewardResult struct {
	OK    bool
	Value uint32
}
```

</details>

## Opcode 021c

**`0x021C` — VehicleLeave**

Declarations: [proto.VehicleLeave](../../internal/proto/mount_state.go#L9).

Source notes: Native requests and invitation messages. 0x021C is also the server's passenger attachment notification; direction determines its layout.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L249)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L465)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>VehicleLeavePacket — internal/proto/mount_state.go</summary>

[Source definition](../../internal/proto/mount_state.go#L72)

```go
func VehicleLeavePacket() []byte { return NewWriter(VehicleLeave).Bytes() }
```

</details>

## Opcode 021d

**`0x021D` — VehicleInvite**

Declarations: [proto.VehicleInvite](../../internal/proto/mount_state.go#L10).

Source notes: Native requests and invitation messages. 0x021C is also the server's passenger attachment notification; direction determines its layout.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L245)

**S→C serializers:** [handleVehicleInvite (server/internal/world/vehicle.go)](../../server/internal/world/vehicle.go#L279)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L449)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleVehicleInvite — server/internal/world/vehicle.go</summary>

[Source definition](../../server/internal/world/vehicle.go#L279)

```go
func (s *Session) handleVehicleInvite(packetData []byte) {
	if s.player == nil || len(packetData) != 6 {
		return
	}
	id, _ := packet.U32(packetData[2:])
	other := FindPlayerByID(id)
	if s.cast != nil || s.dead || s.vehicleBlocked() || !vehicleNear(s.player, other) || !s.player.canSee(other) {
		vehicleReply(s.player, 2)
		return
	}
	vehicleMu.Lock()
	a, b := s.player, other
	driver := a
	if a.ride.state.State == proto.VehicleOff && b.ride.state.State == proto.VehicleDriver {
		driver = b
	}
	passenger := b
	if driver == b {
		passenger = a
	}
	now := time.Now()
	valid := driver.ride.state.State == proto.VehicleDriver && driver.ride.state.Seats == 2 && driver.ride.state.Passenger == 0 && passenger.ride.state.State == proto.VehicleOff && passenger.formLook.Load() == nil && !(a.ride.inviteTo != 0 && now.Before(a.ride.inviteEnd)) && !(b.ride.inviteFrom != 0 && now.Before(b.ride.inviteEnd))
	if valid {
		a.ride.inviteTo = b.CharacterID
		a.ride.inviteEnd = now.Add(30 * time.Second)
		b.ride.inviteFrom = a.CharacterID
		b.ride.inviteDriver = driver.CharacterID
		b.ride.inviteEnd = a.ride.inviteEnd
	}
	vehicleMu.Unlock()
	if !valid {
		vehicleReply(a, 2)
		return
	}
	w := packet.New(packet.Opcode(proto.VehicleInvite))
	w.U32(a.CharacterID)
	b.Send(w.MustFinish(), "0x021D go together invitation")
}
```

</details>

<details><summary>VehicleInvitePacket — internal/proto/mount_state.go</summary>

[Source definition](../../internal/proto/mount_state.go#L56)

```go
func VehicleInvitePacket(id uint32) []byte {
	w := NewWriter(VehicleInvite)
	w.U32(id)
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

## Opcode 021e

**`0x021E` — VehicleAnswer**

Declarations: [proto.VehicleAnswer](../../internal/proto/mount_state.go#L11).

Source notes: Native requests and invitation messages. 0x021C is also the server's passenger attachment notification; direction determines its layout.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L247)

**S→C serializers:** [vehicleReply (server/internal/world/vehicle.go)](../../server/internal/world/vehicle.go#L272)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L456)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>vehicleReply — server/internal/world/vehicle.go</summary>

[Source definition](../../server/internal/world/vehicle.go#L272)

```go
func vehicleReply(p *ConnectedPlayer, code uint8) {
	if p != nil {
		w := packet.New(packet.Opcode(proto.VehicleAnswer))
		w.U8(code)
		p.Send(w.MustFinish(), "0x021E vehicle answer")
	}
}
```

</details>

<details><summary>VehicleAnswerPacket — internal/proto/mount_state.go</summary>

[Source definition](../../internal/proto/mount_state.go#L63)

```go
func VehicleAnswerPacket(accept bool) []byte {
	w := NewWriter(VehicleAnswer)
	if accept {
		w.U8(0)
	} else {
		w.U8(1)
	}
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

## Opcode 0225

**`0x0225` — TeleportSavePositionList**

Declarations: [packet.TeleportSavePositionList](../../server/internal/packet/packet.go#L131), [proto.TeleportSavePositionList](../../internal/proto/opcodes.go#L116).

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

## Opcode 0226

**`0x0226` — ChangeAppearance**

Declarations: [packet.ChangeAppearance](../../server/internal/packet/packet.go#L96), [proto.ChangeAppearance](../../internal/proto/opcodes.go#L80).

Source notes: Request bag,slot,hair,face,height,sex:u8; reply charId:u32,hair,face,height,sex:u8.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L87)

**S→C serializers:** [buildChangeAppearance (server/internal/world/character.go)](../../server/internal/world/character.go#L587)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1293), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L52)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildChangeAppearance — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L587)

```go
func buildChangeAppearance(characterID uint32, hair, face, height, sex uint8) []byte {
	w := packet.New(packet.ChangeAppearance)
	w.U32(characterID)
	w.Bytes([]byte{hair, face, height, sex})
	return w.MustFinish()
}
```

</details>

<details><summary>ChangeAppearancePacket — internal/proto/entities.go</summary>

[Source definition](../../internal/proto/entities.go#L12)

```go
func ChangeAppearancePacket(bag, slot, hair, face, height, gender uint8) []byte {
	w := NewWriter(ChangeAppearance)
	w.U8(bag)
	w.U8(slot)
	w.U8(hair)
	w.U8(face)
	w.U8(height)
	w.U8(gender)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeAppearanceChange — internal/proto/entities.go</summary>

[Source definition](../../internal/proto/entities.go#L23)

```go
func DecodeAppearanceChange(p []byte) (m AppearanceChange, err error) {
	r := NewReader(p)
	m.ID = r.U32()
	m.Hair, m.Face, m.Height, m.Gender = r.U8(), r.U8(), r.U8(), r.U8()
	return m, r.Err()
}
```

```go
type AppearanceChange struct {
	ID                         uint32
	Hair, Face, Height, Gender uint8
}
```

</details>

## Opcode 0228

**`0x0228` — ChangeSexDenied**

Declarations: [packet.ChangeSexDenied](../../server/internal/packet/packet.go#L97), [proto.ChangeSexDenied](../../internal/proto/opcodes.go#L81).

Source notes: empty; SysMsg 862, Game.exe 0x56b820

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [handleChangeAppearance (server/internal/world/inventory.go)](../../server/internal/world/inventory.go#L460)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleChangeAppearance — server/internal/world/inventory.go</summary>

[Source definition](../../server/internal/world/inventory.go#L460)

```go
func (s *Session) handleChangeAppearance(p []byte) {
	if len(p) < 8 || s.char == nil {
		return
	}
	ch := s.char
	bag, slot, hair, face, height, sex := p[2], p[3], p[4], p[5], p[6], p[7]
	if !isValidItemPosition(bag, slot) {
		return
	}
	i := findItem(s.items, bag, slot)
	if i < 0 || s.items[i].Count == 0 || !isConsumableType(s.items[i].Type) {
		return
	}
	def := db.MustDefs().Item(s.items[i].Type, s.items[i].TypeID)
	if def == nil || (def.Special != specialChangeAppearance && def.Special != specialChangeSex) {
		return
	}
	newSex := ch.Gender
	if def.Special == specialChangeSex {
		if sex > 1 {
			return
		}
		newSex = sex
		if s.genderRestrictedEquipment() {
			// Client 0x56b820 (opcode 0x0228) shows SysMsg 862.
			s.send(packet.New(packet.ChangeSexDenied).MustFinish(), "0x0228 change sex denied")
			return
		}
	}
	ch.Hair, ch.Face, ch.Height, ch.Gender = hair, face, height, newSex
	if err := db.SaveAppearance(ch.ID, hair, face, height, newSex); err != nil {
		s.log.Printf("GX Server World failed to save appearance: %v", err)
		return
	}
	res, err := s.consumeOne(i)
	if err != nil {
		s.log.Printf("GX Server World failed to consume item: %v", err)
	}
	s.syncPlayer()
	s.log.Printf("GX Server World change appearance character_id=%d hair=%d face=%d height=%d sex=%d", ch.ID, hair, face, height, newSex)
	s.send(buildUseItemResult(ch.ID, &res), "0x050A use item (change appearance)")
	frame := buildChangeAppearance(ch.ID, hair, face, height, newSex)
	s.broadcastNear(frame, "0x0226 change appearance", true)
}
```

</details>
