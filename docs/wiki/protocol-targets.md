# Protocol: Targets, ground objects and world time

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 0303

**`0x0303` — CharacterShape**

Declarations: [packet.CharacterShape](../../server/internal/packet/packet.go#L95), [proto.CharacterShape](../../internal/proto/opcodes.go#L79).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildCharacterShapeGuild (server/internal/world/character.go)](../../server/internal/world/character.go#L599)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L94)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharacterShapeGuild — server/internal/world/character.go</summary>

[Source definition](../../server/internal/world/character.go#L599)

```go
func buildCharacterShapeGuild(ch *db.Character, faction uint8, items []db.Item, motion uint8, guildName string) []byte {
	w := packet.New(packet.CharacterShape)
	w.U32(charID(ch))
	w.U8(0) // alive
	if motion == 1 {
		w.U8(1)
	} else {
		w.U8(0)
	}
	w.U8(faction)
	if ch != nil {
		w.Bytes([]byte{ch.Race, ch.Hair, ch.Face, ch.Height, ch.Class, ch.Gender, 0, 0})
	} else {
		w.Bytes(make([]byte, 8))
	}
	if ch != nil {
		w.U32(ch.Kills)
	} else {
		w.U32(0)
	}
	for slot := uint8(0); slot < 8; slot++ {
		if it := equippedItem(items, slot); it != nil {
			w.Bytes([]byte{it.Type, it.TypeID, enchantLevel(it)})
		} else {
			w.Bytes([]byte{0, 0, 0})
		}
	}
	w.FixedString(charName(ch), 21)
	if cloak := equippedItem(items, 7); cloak != nil {
		w.Bytes(cloak.Gems[:])
	} else {
		w.Bytes(make([]byte, 6))
	}
	w.FixedString(guildName, 25)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeUserShape — internal/proto/entities.go</summary>

[Source definition](../../internal/proto/entities.go#L43)

```go
func DecodeUserShape(p []byte) (m UserShape, err error) {
	r := NewReader(p)
	m.ID = r.U32()
	m.Dead, m.Motion, m.Faction = r.U8(), r.U8(), r.U8()
	m.Race, m.Hair, m.Face, m.Height, m.Class, m.Gender, m.Unk, m.Mode = r.U8(), r.U8(), r.U8(), r.U8(), r.U8(), r.U8(), r.U8(), r.U8()
	m.Kills = r.U32()
	for i := range m.Equipment {
		for j := range m.Equipment[i] {
			m.Equipment[i][j] = r.U8()
		}
	}
	m.Name = r.Fixed(21)
	// The original reads cloak gems only when equipment slot 7 is set.
	// Rapture also sends six zero bytes without a cloak (discrepancy #2).
	if m.Equipment[7][0] != 0 || r.Remaining() == 31 {
		for i := range m.Cloak {
			m.Cloak[i] = r.U8()
		}
	}
	m.Guild = r.Fixed(25)
	if err := r.Err(); err != nil {
		return m, err
	}
	if r.Remaining() != 0 {
		return m, fmt.Errorf("user shape: %d trailing bytes", r.Remaining())
	}
	return m, nil
}
```

```go
type UserShape struct {
	ID                                                 uint32
	Dead, Motion, Faction                              uint8
	Race, Hair, Face, Height, Class, Gender, Unk, Mode uint8
	Kills                                              uint32
	Equipment                                          [8][3]uint8
	Name                                               string
	Cloak                                              [6]uint8
	Guild                                              string
}
```

</details>

## Opcode 0304

**`0x0304` — TargetMobGetState**

Declarations: [packet.TargetMobGetState](../../server/internal/packet/packet.go#L81), [proto.TargetMobGetState](../../internal/proto/opcodes.go#L65).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L276)

**S→C serializers:** [buildMobState (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1420), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1017)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobState — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1420)

```go
func buildMobState(m *MapMob) []byte {
	w := packet.New(packet.TargetMobGetState)
	w.U32(m.ID)
	w.I32(int32(min(m.CurrentHP, math.MaxInt32)))
	mods := m.mods()
	w.U8(mods.attackSpeedIndex())
	w.U8(mods.moveSpeedIndex())
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

## Opcode 0305

**`0x0305` — TargetMobHpUpdate**

Declarations: [packet.TargetMobHpUpdate](../../server/internal/packet/packet.go#L82), [proto.TargetMobHpUpdate](../../internal/proto/opcodes.go#L66).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L276)

**S→C serializers:** [buildMobTargetHP (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1412)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1069)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobTargetHP — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1412)

```go
func buildMobTargetHP(m *MapMob) []byte {
	w := packet.New(packet.TargetMobHpUpdate)
	w.U32(m.ID)
	w.I32(int32(min(m.CurrentHP, math.MaxInt32)))
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeMobHealth — internal/proto/combat.go</summary>

[Source definition](../../internal/proto/combat.go#L21)

```go
func DecodeMobHealth(p []byte) (m MobHealth, err error) {
	r := NewReader(p)
	m.ID, m.HP = r.U32(), r.I32()
	return m, r.Err()
}
```

```go
type MobHealth struct {
	ID uint32
	HP int32
}
```

</details>

## Opcode 0307

**`0x0307` — InspectEquipment**

Declarations: [packet.InspectEquipment](../../server/internal/packet/packet.go#L83), [proto.InspectEquipment](../../internal/proto/inspect.go#L5).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L274)

**S→C serializers:** [buildInspectEquipment (server/internal/world/inspect.go)](../../server/internal/world/inspect.go#L49)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildInspectEquipment — server/internal/world/inspect.go</summary>

[Source definition](../../server/internal/world/inspect.go#L49)

```go
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
```

</details>

<details><summary>InspectPacket — internal/proto/inspect.go</summary>

[Source definition](../../internal/proto/inspect.go#L10)

```go
func InspectPacket(target uint32) []byte {
	w := NewWriter(InspectEquipment)
	w.U32(target)
	return w.Bytes()
}
```

</details>

## Opcode 030a

**`0x030A` — TargetClear**

Declarations: [packet.TargetClear](../../server/internal/packet/packet.go#L84), [proto.TargetClear](../../internal/proto/opcodes.go#L67).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L231)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

The shape is decoded inline, delegated, or only declared. Use the linked adapter/reducer and the codec appendix below; there is no inferred fixed layout here.

## Opcode 030b

**`0x030B` — TargetEffects**

Declarations: [proto.TargetEffects](../../internal/proto/target_effects.go#L7).

Source notes: Native target-effect readers: 0x5667a0, rendered by 0x50b6a0.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [tickTargetEffects (server/internal/world/target_effects.go)](../../server/internal/world/target_effects.go#L34)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L800)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>tickTargetEffects — server/internal/world/target_effects.go</summary>

[Source definition](../../server/internal/world/target_effects.go#L34)

```go
func (s *Session) tickTargetEffects(now time.Time) {
	if s.effectWatchID == 0 || s.player == nil || s.char == nil || now.Before(s.effectWatchAt) {
		return
	}
	s.effectWatchAt = now.Add(250 * time.Millisecond)
	var effects []proto.TargetEffect
	valid := false
	left := func(end time.Time) uint32 { return uint32(max(0, math.Ceil(end.Sub(now).Seconds()))) }
	switch s.effectWatchKind {
	case proto.TargetEffectMob:
		if s.player.knows(viewMob, s.effectWatchID) {
			if m, ok := MobByID(s.char.MapID, s.effectWatchID); ok && m.CurrentHP > 0 {
				valid = true
				for _, b := range m.Effects {
					if n := left(b.endsAt); n > 0 {
						effects = append(effects, proto.TargetEffect{Skill: b.Def.SkillID, Level: b.Def.SkillLevel, Remaining: n})
					}
				}
			}
		}
	case proto.TargetEffectUser:
		p := FindPlayerByID(s.effectWatchID)
		if p != nil && p.MapID() == s.char.MapID && p.IsAlive() && s.player.canSee(p) && (p == s.player || s.player.knows(viewUser, p.CharacterID)) {
			valid = true
			s.sendTargetHealth(p)
			p.mu.Lock()
			for _, b := range p.targetEffects {
				if n := left(b.EndsAt); n > 0 {
					effects = append(effects, proto.TargetEffect{Skill: b.SkillID, Level: b.SkillLevel, Remaining: n})
				}
			}
			p.mu.Unlock()
		}
	}
	effects = effects[:min(len(effects), 255)]
	w := packet.New(packet.Opcode(proto.TargetEffects))
	w.U8(s.effectWatchKind)
	w.U32(s.effectWatchID)
	w.U8(uint8(len(effects)))
	for _, e := range effects {
		w.U16(e.Skill)
		w.U8(e.Level)
		w.U32(e.Remaining)
	}
	frame := w.MustFinish()
	hash := sha256.Sum256(frame)
	if hash != s.effectWatchHash {
		s.effectWatchHash = hash
		s.send(frame, "native selected-target effects")
	}
	if !valid {
		s.effectWatchID = 0
	}
}
```

</details>

<details><summary>DecodeTargetEffects — internal/proto/target_effects.go</summary>

[Source definition](../../internal/proto/target_effects.go#L36)

```go
func DecodeTargetEffects(p []byte) (v TargetEffectSnapshot, err error) {
	r := NewReader(p)
	v.Kind, v.ID = r.U8(), r.U32()
	n := int(r.U8())
	if v.Kind != 1 && v.Kind != 2 || r.Remaining() != n*7 {
		return v, fmt.Errorf("invalid target effects snapshot")
	}
	v.Effects = make([]TargetEffect, n)
	for i := range v.Effects {
		v.Effects[i] = TargetEffect{Skill: r.U16(), Level: r.U8(), Remaining: r.U32(), HasRemaining: true}
	}
	return v, r.Err()
}
```

```go
type TargetEffect struct {
	Skill        uint16
	Level        uint8
	Remaining    uint32
	HasRemaining bool
}
```

```go
type TargetEffectSnapshot struct {
	Kind    uint8
	ID      uint32
	Effects []TargetEffect
}
```

</details>

## Opcode 030c

**`0x030C` — TargetEffectAdd**

Declarations: [proto.TargetEffectAdd](../../internal/proto/target_effects.go#L8).

Source notes: Native target-effect readers: 0x5667a0, rendered by 0x50b6a0.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L807)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeTargetEffectChange — internal/proto/target_effects.go</summary>

[Source definition](../../internal/proto/target_effects.go#L49)

```go
func DecodeTargetEffectChange(p []byte) (v TargetEffectSnapshot, e TargetEffect, err error) {
	if len(p) != 10 {
		return v, e, fmt.Errorf("invalid target effect change")
	}
	r := NewReader(p)
	v.Kind, v.ID = r.U8(), r.U32()
	e.Skill, e.Level = r.U16(), r.U8()
	if v.Kind != 1 && v.Kind != 2 {
		return v, e, fmt.Errorf("invalid target effect kind")
	}
	return v, e, r.Err()
}
```

```go
type TargetEffect struct {
	Skill        uint16
	Level        uint8
	Remaining    uint32
	HasRemaining bool
}
```

```go
type TargetEffectSnapshot struct {
	Kind    uint8
	ID      uint32
	Effects []TargetEffect
}
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

## Opcode 030d

**`0x030D` — TargetEffectRemove**

Declarations: [proto.TargetEffectRemove](../../internal/proto/target_effects.go#L9).

Source notes: Native target-effect readers: 0x5667a0, rendered by 0x50b6a0.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L807)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeTargetEffectChange — internal/proto/target_effects.go</summary>

[Source definition](../../internal/proto/target_effects.go#L49)

```go
func DecodeTargetEffectChange(p []byte) (v TargetEffectSnapshot, e TargetEffect, err error) {
	if len(p) != 10 {
		return v, e, fmt.Errorf("invalid target effect change")
	}
	r := NewReader(p)
	v.Kind, v.ID = r.U8(), r.U32()
	e.Skill, e.Level = r.U16(), r.U8()
	if v.Kind != 1 && v.Kind != 2 {
		return v, e, fmt.Errorf("invalid target effect kind")
	}
	return v, e, r.Err()
}
```

```go
type TargetEffect struct {
	Skill        uint16
	Level        uint8
	Remaining    uint32
	HasRemaining bool
}
```

```go
type TargetEffectSnapshot struct {
	Kind    uint8
	ID      uint32
	Effects []TargetEffect
}
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

## Opcode 0401

**`0x0401` — MapAddItem**

Declarations: [packet.MapAddItem](../../server/internal/packet/packet.go#L107), [proto.MapAddItem](../../internal/proto/opcodes.go#L90).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMapAddItem (server/internal/world/mapitems.go)](../../server/internal/world/mapitems.go#L129)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1005), [applyInventoryPacket (internal/game/state_inventory_packets.go)](../../internal/game/state_inventory_packets.go#L10)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMapAddItem — server/internal/world/mapitems.go</summary>

[Source definition](../../server/internal/world/mapitems.go#L129)

```go
func buildMapAddItem(m *MapItem) []byte {
	w := packet.New(packet.MapAddItem)
	w.U32(m.ID)
	w.U8(1)
	if m.Money > 0 {
		w.Bytes([]byte{moneyItemType, 0})
		w.Count(1)
	} else {
		w.Bytes([]byte{m.Item.Type, m.Item.TypeID})
		w.Count(m.Item.Count)
	}
	w.F32(m.PosX)
	w.F32(m.PosY)
	w.F32(m.PosZ) // the client reads no owner (0x0058f320)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeGroundItem — internal/proto/inventory.go</summary>

[Source definition](../../internal/proto/inventory.go#L84)

```go
func DecodeGroundItem(p []byte) (m GroundItem, err error) {
	r := NewReader(p)
	m.ID = r.U32()
	m.Unk, m.Type, m.TypeID, m.Count = r.U8(), r.U8(), r.U8(), r.Count()
	m.X, m.Y, m.Z = r.F32(), r.F32(), r.F32()
	return m, r.Err()
}
```

```go
type GroundItem struct {
	ID                uint32
	Unk, Type, TypeID uint8
	Count             uint16
	X, Y, Z           float32
}
```

</details>

## Opcode 0402

**`0x0402` — MapRemoveItem**

Declarations: [packet.MapRemoveItem](../../server/internal/packet/packet.go#L108), [proto.MapRemoveItem](../../internal/proto/opcodes.go#L91).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMapRemoveItem (server/internal/world/mapitems.go)](../../server/internal/world/mapitems.go#L146)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1005), [applyInventoryPacket (internal/game/state_inventory_packets.go)](../../internal/game/state_inventory_packets.go#L16)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMapRemoveItem — server/internal/world/mapitems.go</summary>

[Source definition](../../server/internal/world/mapitems.go#L146)

```go
func buildMapRemoveItem(id uint32) []byte {
	w := packet.New(packet.MapRemoveItem)
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

## Opcode 0404

**`0x0404` — WorldDay**

Declarations: [packet.WorldDay](../../server/internal/packet/packet.go#L99), [proto.WorldDay](../../internal/proto/opcodes.go#L82).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildWorldDay (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L61)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildWorldDay — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L61)

```go
func buildWorldDay() []byte {
	w := packet.New(packet.WorldDay)
	w.U32(worldDay()) // CWorld::GetWorldDay
	return w.MustFinish()
}
```

</details>

## Opcode 0406

**`0x0406` — CharacterReviveInit**

Declarations: [packet.CharacterReviveInit](../../server/internal/packet/packet.go#L109), [proto.CharacterReviveInit](../../internal/proto/opcodes.go#L92).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildCharacterReviveInit (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1468)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildCharacterReviveInit — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1468)

```go
func buildCharacterReviveInit(characterID uint32) []byte {
	w := packet.New(packet.CharacterReviveInit)
	w.U32(characterID)
	return w.MustFinish()
}
```

</details>
