# Protocol: Characters and account selection

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 0101

**`0x0101` — CharacterList**

Declarations: [packet.CharacterList](../../server/internal/packet/packet.go#L21), [proto.CharacterList](../../internal/proto/opcodes.go#L9).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildEmptyCharacterListSlot (server/internal/world/character.go)](../../server/internal/world/character.go#L647), [buildCharacterListSlot (server/internal/world/character.go)](../../server/internal/world/character.go#L654)

**Client receive reducers:** [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [Apply (internal/game/state.go)](../../internal/game/state.go#L1122)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildEmptyCharacterListSlot — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L647)

```go
func buildEmptyCharacterListSlot(slot uint8) []byte {
	w := packet.New(packet.CharacterList)
	w.U8(slot)
	w.U32(0)
	return w.MustFinish()
}
```

</details>

<details><summary>buildCharacterListSlot — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L654)

```go
func buildCharacterListSlot(ch *db.Character, items []db.Item) []byte {
	w := packet.New(packet.CharacterList)
	w.U8(ch.Slot)
	w.U32(ch.ID)
	w.U32(0)
	w.U16(ch.Level)
	w.Bytes([]byte{ch.Race, 0, ch.Hair, ch.Face, ch.Height, ch.Class, ch.Gender})
	w.U16(ch.MapID)
	for _, v := range []uint16{ch.Str, ch.Rec, ch.Int, ch.Wis, ch.Dex, ch.Luc} {
		w.U16(v)
	}
	w.U16(uint16(ch.MaxHP))
	w.U16(uint16(ch.MaxMP))
	w.U16(uint16(ch.MaxSP))
	for slot := uint8(0); slot < 8; slot++ {
		if it := equippedItem(items, slot); it != nil {
			w.U8(it.Type)
		} else {
			w.U8(0)
		}
	}
	for slot := uint8(0); slot < 8; slot++ {
		if it := equippedItem(items, slot); it != nil {
			w.U8(it.TypeID)
		} else {
			w.U8(0)
		}
	}
	w.FixedString(ch.Name, 19)
	w.Bool(ch.PendingDelete)
	w.Bool(ch.RenamePending)
	w.Bytes(make([]byte, 6))
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeCharacterSlot — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L159)

```go
func DecodeCharacterSlot(p []byte) (m CharacterSlot, err error) {
	r := NewReader(p)
	m.Slot, m.ID = r.U8(), r.U32()
	if m.ID == 0 {
		return m, r.Err()
	}
	m.Unk, m.Level = r.U32(), r.U16()
	m.Race, m.Mode, m.Hair, m.Face, m.Height, m.Class, m.Gender = r.U8(), r.U8(), r.U8(), r.U8(), r.U8(), r.U8(), r.U8()
	m.MapID = r.U16()
	m.Str, m.Rec, m.Int, m.Wis, m.Dex, m.Luc = r.U16(), r.U16(), r.U16(), r.U16(), r.U16(), r.U16()
	m.MaxHP, m.MaxMP, m.MaxSP = r.U16(), r.U16(), r.U16()
	for i := range m.EquipType {
		m.EquipType[i] = r.U8()
	}
	for i := range m.EquipTypeID {
		m.EquipTypeID[i] = r.U8()
	}
	m.Name = r.Fixed(19)
	m.PendingDelete, m.RenamePending = r.Bool(), r.Bool()
	return m, r.Err()
}
```

```go
type CharacterSlot struct {
	Slot                                          uint8
	ID                                            uint32
	Unk                                           uint32
	Level                                         uint16
	Race, Mode, Hair, Face, Height, Class, Gender uint8
	MapID                                         uint16
	Str, Rec, Int, Wis, Dex, Luc                  uint16
	MaxHP, MaxMP, MaxSP                           uint16
	EquipType, EquipTypeID                        [8]uint8
	Cloak                                         [6]uint8 // gem bytes of the equipped cloak: its emblem
	Name                                          string
	PendingDelete, RenamePending                  bool
}
```

</details>

## Opcode 0102

**`0x0102` — CreateCharacter**

Declarations: [packet.CreateCharacter](../../server/internal/packet/packet.go#L22), [proto.CreateCharacter](../../internal/proto/opcodes.go#L10).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L53)

**S→C serializers:** [buildCreateCharacterResult (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L41)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1133)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCreateCharacterResult — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L41)

```go
func buildCreateCharacterResult(ok bool) []byte {
	w := packet.New(packet.CreateCharacter)
	w.U8(resultByte(ok))
	return w.MustFinish()
}
```

</details>

<details><summary>CreateCharacterPacket — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L26)

```go
func CreateCharacterPacket(c NewCharacter) []byte {
	w := NewWriter(CreateCharacter)
	for _, v := range []uint8{c.Slot, c.Race, c.Mode, c.Hair, c.Face, c.Height, c.Class, c.Gender} {
		w.U8(v)
	}
	name := EncodeText(c.Name)
	if len(name) > 18 {
		name = []byte(textwire.Prefix(string(name), 18))
	}
	w.Raw(append(name, 0))
	return w.Bytes()
}
```

```go
type NewCharacter struct {
	Slot, Race, Mode, Hair, Face, Height, Class, Gender uint8
	Name                                                string
}
```

</details>

<details><summary>DecodeU8 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L194)

```go
func DecodeU8(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 0103

**`0x0103` — DeleteCharacter**

Declarations: [packet.DeleteCharacter](../../server/internal/packet/packet.go#L25), [proto.DeleteCharacter](../../internal/proto/opcodes.go#L11).

Source notes: Request charId:u32; response result:u8 (0=ok), charId:u32. Soft delete, restorable for 2 hours.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L55)

**S→C serializers:** [buildCharResult (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L47), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1139)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharResult — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L47)

```go
func buildCharResult(op packet.Opcode, ok bool, characterID uint32) []byte {
	w := packet.New(op)
	w.U8(resultByte(ok))
	w.U32(characterID)
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

<details><summary>DecodeIDResult — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L187)

```go
func DecodeIDResult(p []byte) (m IDResult, err error) {
	r := NewReader(p)
	m.Result, m.ID = r.U8(), r.U32()
	return m, r.Err()
}
```

```go
type IDResult struct {
	Result uint8
	ID     uint32
}
```

</details>

## Opcode 0104

**`0x0104` — SelectCharacter**

Declarations: [packet.SelectCharacter](../../server/internal/packet/packet.go#L26), [proto.SelectCharacter](../../internal/proto/opcodes.go#L12).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L71)

**S→C serializers:** [buildCharResult (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L47), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1153)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharResult — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L47)

```go
func buildCharResult(op packet.Opcode, ok bool, characterID uint32) []byte {
	w := packet.New(op)
	w.U8(resultByte(ok))
	w.U32(characterID)
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

<details><summary>DecodeIDResult — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L187)

```go
func DecodeIDResult(p []byte) (m IDResult, err error) {
	r := NewReader(p)
	m.Result, m.ID = r.U8(), r.U32()
	return m, r.Err()
}
```

```go
type IDResult struct {
	Result uint8
	ID     uint32
}
```

</details>

## Opcode 0105

**`0x0105` — CharacterDetails**

Declarations: [packet.CharacterDetails](../../server/internal/packet/packet.go#L31), [proto.CharacterDetails](../../internal/proto/opcodes.go#L15).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildCharacterDetails (server/internal/world/character.go)](../../server/internal/world/character.go#L400)

**Client receive reducers:** [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [Apply (internal/game/state.go)](../../internal/game/state.go#L1201)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharacterDetails — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L400)

```go
func buildCharacterDetails(ch *db.Character, s *CharacterStats) []byte {
	guildName := ""
	if ch != nil {
		guildName = guildNameFor(ch.ID)
	}
	w := packet.New(packet.CharacterDetails)
	var statPoint, skillPoint uint16
	var exp, money, k1, k2, k3, k4 uint32
	if ch != nil {
		statPoint, skillPoint, exp, money = ch.StatPoint, ch.SkillPoint, ch.Exp, ch.Money
		k1, k2, k3, k4 = ch.Kills, ch.Deaths, ch.Victories, ch.Defeats
	}
	x, y, z := charPos(ch)
	var dir uint16
	var prev, next uint32
	if ch != nil {
		dir = uint16(ch.Dir)
		prev, next = expBounds(ch)
	}
	w.U16(statPoint)
	w.U16(skillPoint)
	w.U32(s.MaxHP)
	w.U32(s.MaxMP)
	w.U32(s.MaxSP)
	w.U16(dir)
	w.U32(prev)
	w.U32(next)
	w.U32(exp)
	w.U32(money)
	w.F32(x)
	w.F32(y)
	w.F32(z)
	w.U32(k1)
	w.U32(k2)
	w.U32(k3)
	w.U32(k4)
	w.FixedString(guildName, 25)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeDetails — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L210)

```go
func DecodeDetails(p []byte) (m Details, err error) {
	r := NewReader(p)
	m.StatPoint, m.SkillPoint = r.U16(), r.U16()
	m.MaxHP, m.MaxMP, m.MaxSP = r.U32(), r.U32(), r.U32()
	m.Dir = r.U16()
	m.PrevExp, m.NextExp, m.Exp, m.Money = r.U32(), r.U32(), r.U32(), r.U32()
	m.X, m.Y, m.Z = r.F32(), r.F32(), r.F32()
	m.Kills, m.Deaths, m.Wins, m.Losses = r.U32(), r.U32(), r.U32(), r.U32()
	m.Guild = r.Fixed(25)
	return m, r.Err()
}
```

```go
type Details struct {
	StatPoint, SkillPoint        uint16
	MaxHP, MaxMP, MaxSP          uint32
	Dir                          uint16
	PrevExp, NextExp, Exp, Money uint32
	X, Y, Z                      float32
	Kills, Deaths, Wins, Losses  uint32
	Guild                        string
}
```

</details>

## Opcode 0106

**`0x0106` — CharacterItems**

Declarations: [packet.CharacterItems](../../server/internal/packet/packet.go#L32), [proto.CharacterItems](../../internal/proto/opcodes.go#L16).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildInventoryPackets (server/internal/world/inventory.go)](../../server/internal/world/inventory.go#L17), [buildZeroCount (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L74)

**Client receive reducers:** [questMapInventoryChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L48), [Apply (internal/game/state.go)](../../internal/game/state.go#L1005), [applyInventoryPacket (internal/game/state_inventory_packets.go)](../../internal/game/state_inventory_packets.go#L22)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildInventoryPackets — server/internal/world/inventory.go</summary>

[Source definition](../../server/internal/world/inventory.go#L17)

```go
func buildInventoryPackets(items []db.Item) [][]byte {
	if len(items) == 0 {
		return [][]byte{buildZeroCount(packet.CharacterItems)}
	}
	var out [][]byte
	for start := 0; start < len(items); start += inventoryChunk {
		chunk := items[start:min(start+inventoryChunk, len(items))]
		w := packet.New(packet.CharacterItems)
		w.U8(uint8(len(chunk)))
		for i := range chunk {
			it := &chunk[i]
			w.U8(it.Bag)
			w.U8(it.Slot)
			w.U8(it.Type)
			w.U8(it.TypeID)
			w.U16(it.Quality)
			w.Bytes(it.Gems[:])
			w.Count(it.Count)
			w.FixedString(it.CraftName, 21)
		}
		out = append(out, w.MustFinish())
	}
	return out
}
```

</details>

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

<details><summary>DecodeInventory — internal/proto/inventory.go</summary>

[Source definition](../../internal/proto/inventory.go#L14)

```go
func DecodeInventory(p []byte) ([]InventoryItem, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/34 {
		return nil, ErrShort
	}
	items := make([]InventoryItem, n)
	for i := range items {
		m := &items[i]
		m.Bag, m.Slot, m.Type, m.TypeID = r.U8(), r.U8(), r.U8(), r.U8()
		m.Quality = r.U16()
		for j := range m.Gems {
			m.Gems[j] = r.U8()
		}
		m.Count = r.Count()
		m.Craft = r.Fixed(21)
		if m.Bag > 5 || m.Slot >= 24 {
			return nil, fmt.Errorf("inventory: invalid position %d/%d", m.Bag, m.Slot)
		}
	}
	return items, r.Err()
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

## Opcode 0107

**`0x0107` — Logout**

Declarations: [packet.Logout](../../server/internal/packet/packet.go#L33), [proto.Logout](../../internal/proto/opcodes.go#L17).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L302)

**S→C serializers:** [buildLeaveWorldAck (server/internal/world/character.go)](../../server/internal/world/character.go#L543)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1307)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildLeaveWorldAck — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L543)

```go
func buildLeaveWorldAck(characterID uint32) []byte {
	w := packet.New(packet.Logout)
	w.U32(characterID)
	return w.MustFinish()
}
```

</details>

<details><summary>LogoutPacket — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L59)

```go
func LogoutPacket() []byte    { return NewWriter(Logout).Bytes() }
```

</details>

## Opcode 0108

**`0x0108` — CharacterSkills**

Declarations: [packet.CharacterSkills](../../server/internal/packet/packet.go#L34), [proto.CharacterSkills](../../internal/proto/opcodes.go#L18).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildCharacterSkills (server/internal/world/skills.go)](../../server/internal/world/skills.go#L25)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L768)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharacterSkills — server/internal/world/skills.go</summary>

[Source definition](../../server/internal/world/skills.go#L25)

```go
func buildCharacterSkills(skills []db.CharacterSkill) []byte {
	w := packet.New(packet.CharacterSkills)
	skills = skills[:min(len(skills), 255)]
	w.U8(uint8(len(skills)))
	for _, s := range skills {
		w.U16(s.SkillID)
		w.U8(s.Level)
		w.U8(s.Number)
		w.I32(int32(s.Delay))
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeSkills — internal/proto/skills.go</summary>

[Source definition](../../internal/proto/skills.go#L18)

```go
func DecodeSkills(p []byte) ([]LearnedSkill, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/8 {
		return nil, ErrShort
	}
	out := make([]LearnedSkill, n)
	for i := range out {
		out[i] = LearnedSkill{r.U16(), r.U8(), r.U8(), r.I32()}
	}
	return out, r.Err()
}
```

```go
type LearnedSkill struct {
	ID            uint16
	Level, Number uint8
	Delay         int32
}
```

</details>

## Opcode 0109

**`0x0109` — AccountFaction**

Declarations: [packet.AccountFaction](../../server/internal/packet/packet.go#L35), [proto.AccountFaction](../../internal/proto/opcodes.go#L19).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L40)

**S→C serializers:** [buildAccountFaction (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L24)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1110)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildAccountFaction — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L24)

```go
func buildAccountFaction(faction uint8) []byte {
	w := packet.New(packet.AccountFaction)
	w.U8(faction)
	w.U8(0)
	return w.MustFinish()
}
```

</details>

<details><summary>FactionPacket — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L13)

```go
func FactionPacket(faction uint8) []byte {
	w := NewWriter(AccountFaction)
	w.U8(faction)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeFaction — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L135)

```go
func DecodeFaction(p []byte) (m Faction, err error) {
	r := NewReader(p)
	m.Faction = r.U8()
	r.U8() // reserved byte
	return m, r.Err()
}
```

```go
type Faction struct{ Faction uint8 }
```

</details>

## Opcode 010a

**`0x010A` — CharacterActiveBuffs**

Declarations: [packet.CharacterActiveBuffs](../../server/internal/packet/packet.go#L36), [proto.CharacterActiveBuffs](../../internal/proto/opcodes.go#L20).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildCharacterActiveBuffs (server/internal/world/skillcast.go)](../../server/internal/world/skillcast.go#L104)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L661)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharacterActiveBuffs — server/internal/world/skillcast.go</summary>

[Source definition](../../server/internal/world/skillcast.go#L104)

```go
func buildCharacterActiveBuffs(buffs []ActiveBuff) []byte {
	w := packet.New(packet.CharacterActiveBuffs)
	buffs = buffs[:min(len(buffs), 255)]
	w.U8(uint8(len(buffs)))
	for _, b := range buffs {
		w.U32(b.ID)
		w.U16(b.SkillID)
		w.U8(b.SkillLevel)
		if b.MapCurse {
			w.U32(0)
		} else {
			w.U32(uint32(max(time.Until(b.EndsAt), 0) / time.Second))
		}
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeActiveBuffs — internal/proto/buffs.go</summary>

[Source definition](../../internal/proto/buffs.go#L14)

```go
func DecodeActiveBuffs(p []byte) (out []ActiveBuff, err error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/11 {
		return nil, ErrShort
	}
	out = make([]ActiveBuff, n)
	for i := range out {
		out[i] = ActiveBuff{Instance: r.U32(), Skill: r.U16(), Level: r.U8(), Remaining: r.U32(), HasRemaining: true}
	}
	return out, r.Err()
}
```

```go
type ActiveBuff struct {
	Instance     uint32
	Skill        uint16
	Level        uint8
	Remaining    uint32
	HasRemaining bool
}
```

</details>

## Opcode 010b

**`0x010B` — CharacterSkillBar**

Declarations: [packet.CharacterSkillBar](../../server/internal/packet/packet.go#L37), [proto.CharacterSkillBar](../../internal/proto/opcodes.go#L21).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L229)

**S→C serializers:** [buildCharacterSkillBar (server/internal/world/skills.go)](../../server/internal/world/skills.go#L65)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L777)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharacterSkillBar — server/internal/world/skills.go</summary>

[Source definition](../../server/internal/world/skills.go#L65)

```go
func buildCharacterSkillBar(slots []db.QuickSlot) []byte {
	var valid []db.QuickSlot
	for _, q := range slots {
		if isValidQuickSlot(q) && len(valid) < maxQuickSlots {
			valid = append(valid, q)
		}
	}
	w := packet.New(packet.CharacterSkillBar)
	w.U8(uint8(len(valid)))
	for _, q := range valid {
		w.U8(q.QuickBar)
		w.U8(q.QuickSlot)
		w.U8(q.Bag)
		w.U16(q.Number)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>SkillBarPacket — internal/proto/skills.go</summary>

[Source definition](../../internal/proto/skills.go#L50)

```go
func SkillBarPacket(slots []QuickSlot) ([]byte, error) {
	if len(slots) > 128 {
		return nil, fmt.Errorf("skill bar: too many slots")
	}
	w := NewWriter(CharacterSkillBar)
	w.U8(uint8(len(slots)))
	for _, q := range slots {
		if q.Bar >= 5 || q.Slot >= 10 || !((q.Bag <= 5 && q.Number < 24) || (q.Bag == 100 && q.Number > 0)) {
			return nil, fmt.Errorf("skill bar: invalid slot %+v", q)
		}
		w.U8(q.Bar)
		w.U8(q.Slot)
		w.U8(q.Bag)
		w.U16(q.Number)
	}
	return w.Bytes(), nil
}
```

```go
type QuickSlot struct {
	Bar, Slot, Bag uint8
	Number         uint16
}
```

</details>

<details><summary>DecodeQuickSlots — internal/proto/skills.go</summary>

[Source definition](../../internal/proto/skills.go#L36)

```go
func DecodeQuickSlots(p []byte) ([]QuickSlot, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/5 {
		return nil, ErrShort
	}
	out := make([]QuickSlot, n)
	for i := range out {
		out[i] = QuickSlot{r.U8(), r.U8(), r.U8(), r.U16()}
	}
	return out, r.Err()
}
```

</details>

## Opcode 010e

**`0x010E` — RenameCharacter**

Declarations: [packet.RenameCharacter](../../server/internal/packet/packet.go#L30), [proto.RenameCharacter](../../internal/proto/opcodes.go#L14).

Source notes: Request charId:u32, newName; response result:u8 with 1 = ok, 0 = fail.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L69)

**S→C serializers:** [buildRenameResult (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L55)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildRenameResult — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L55)

```go
func buildRenameResult(ok bool) []byte {
	w := packet.New(packet.RenameCharacter)
	w.Bool(ok)
	return w.MustFinish()
}
```

</details>

## Opcode 010f

**`0x010F` — RestoreCharacter**

Declarations: [packet.RestoreCharacter](../../server/internal/packet/packet.go#L28), [proto.RestoreCharacter](../../internal/proto/opcodes.go#L13).

Source notes: Same shape as DeleteCharacter.

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L62)

**S→C serializers:** [buildCharResult (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L47), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1139)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharResult — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L47)

```go
func buildCharResult(op packet.Opcode, ok bool, characterID uint32) []byte {
	w := packet.New(op)
	w.U8(resultByte(ok))
	w.U32(characterID)
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

<details><summary>DecodeIDResult — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L187)

```go
func DecodeIDResult(p []byte) (m IDResult, err error) {
	r := NewReader(p)
	m.Result, m.ID = r.U8(), r.U32()
	return m, r.Err()
}
```

```go
type IDResult struct {
	Result uint8
	ID     uint32
}
```

</details>
