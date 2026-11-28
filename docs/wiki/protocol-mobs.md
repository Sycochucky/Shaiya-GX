# Protocol: Monsters and NPCs

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 0601

**`0x0601` — MobEnter**

Declarations: [packet.MobEnter](../../server/internal/packet/packet.go#L110), [proto.MobEnter](../../internal/proto/opcodes.go#L93).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L124)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeMobEnter — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L272)

```go
func DecodeMobEnter(p []byte) (m MobSpawn, err error) {
	r := NewReader(p)
	m.ID, m.New, m.MobID, m.X = r.U32(), r.Bool(), r.U16(), r.F32()
	m.HeightKnown = len(p) >= 2 && Opcode(binary.LittleEndian.Uint16(p[:2])) == GXMobEnter
	if m.HeightKnown {
		m.Y = r.F32()
	}
	m.Z = r.F32()
	if m.HeightKnown && r.Remaining() > 0 {
		if r.Remaining() != 2 {
			return m, fmt.Errorf("invalid GX mob metadata length")
		}
		m.Country = r.U8()
		height := r.U8()
		if m.Country > 2 || height > 1 {
			return m, fmt.Errorf("invalid GX mob metadata")
		}
		m.CountryKnown = true
		m.HeightKnown = height == 1
	}
	return m, r.Err()
}
```

```go
type MobSpawn struct {
	Country      uint8
	CountryKnown bool
	ID           uint32
	New          bool
	MobID        uint16
	X, Y, Z      float32
	HeightKnown  bool
}
```

</details>

## Opcode 0602

**`0x0602` — MobLeave**

Declarations: [packet.MobLeave](../../server/internal/packet/packet.go#L111), [proto.MobLeave](../../internal/proto/opcodes.go#L94).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobLeave (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1405)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L38)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobLeave — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1405)

```go
func buildMobLeave(id uint32) []byte {
	w := packet.New(packet.MobLeave)
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

## Opcode 0603

**`0x0603` — MobMove**

Declarations: [packet.MobMove](../../server/internal/packet/packet.go#L112), [proto.MobMove](../../internal/proto/opcodes.go#L95).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L130)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeMobMove — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L304)

```go
func DecodeMobMove(p []byte) (m MobMotion, err error) {
	r := NewReader(p)
	m.ID, m.Motion, m.X = r.U32(), r.U8(), r.F32()
	m.HeightKnown = len(p) >= 2 && Opcode(binary.LittleEndian.Uint16(p[:2])) == GXMobMove
	if m.HeightKnown {
		m.Y = r.F32()
	}
	m.Z = r.F32()
	if m.HeightKnown {
		m.DurationMS = r.U32()
	}
	return m, r.Err()
}
```

```go
type MobMotion struct {
	ID          uint32
	Motion      uint8
	X, Y, Z     float32
	HeightKnown bool
	DurationMS  uint32
}
```

</details>

## Opcode 0605

**`0x0605` — MobAttackCharacter**

Declarations: [packet.MobAttackCharacter](../../server/internal/packet/packet.go#L113), [proto.MobAttackCharacter](../../internal/proto/opcodes.go#L96).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobAttackCharacter (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1441)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1079)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobAttackCharacter — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1441)

```go
func buildMobAttackCharacter(result uint8, mobID, characterID uint32, damage uint16) []byte {
	w := packet.New(packet.MobAttackCharacter)
	w.U8(result)
	w.U32(mobID)
	w.U32(characterID)
	w.U16(damage)
	w.U16(0)
	w.U16(0)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeAttackResult — internal/proto/combat.go</summary>

[Source definition](../../internal/proto/combat.go#L35)

```go
func DecodeAttackResult(p []byte) (m AttackResult, err error) {
	r := NewReader(p)
	m.Result = r.U8()
	m.Source, m.Target = r.U32(), r.U32()
	m.HP, m.SP, m.MP = r.U16(), r.U16(), r.U16()
	return m, r.Err()
}
```

```go
type AttackResult struct {
	Result         uint8
	Source, Target uint32
	HP, SP, MP     uint16
}
```

</details>

## Opcode 0606

**`0x0606` — MobDeath**

Declarations: [packet.MobDeath](../../server/internal/packet/packet.go#L114), [proto.MobDeath](../../internal/proto/opcodes.go#L97).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobDeathBy (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1507)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1085)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobDeathBy — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1507)

```go
func buildMobDeathBy(mobID, killerID uint32, deathType uint8) []byte {
	w := packet.New(packet.MobDeath)
	w.U32(mobID)
	w.U8(deathType)
	w.U32(killerID)
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

## Opcode 0607

**`0x0607` — MobPeriodicDamage**

Declarations: [packet.MobPeriodicDamage](../../server/internal/packet/packet.go#L115), [proto.MobPeriodicDamage](../../internal/proto/opcodes.go#L98).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobPeriodicDamage (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1515)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L928)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobPeriodicDamage — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1515)

```go
func buildMobPeriodicDamage(mobID uint32, skillID uint16, level uint8, hp, sp, mp uint16) []byte {
	w := packet.New(packet.MobPeriodicDamage)
	w.U32(mobID)
	w.U16(skillID)
	w.U8(level)
	w.U16(hp)
	w.U16(sp)
	w.U16(mp)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodePeriodicEffect — internal/proto/periodic.go</summary>

[Source definition](../../internal/proto/periodic.go#L13)

```go
func DecodePeriodicEffect(p []byte) (v PeriodicEffect, err error) {
	r := NewReader(p)
	v.Target, v.Skill, v.Level = r.U32(), r.U16(), r.U8()
	v.HP, v.Resource1, v.Resource2 = r.U16(), r.U16(), r.U16()
	return v, r.Err()
}
```

```go
type PeriodicEffect struct {
	Target                   uint32
	Skill                    uint16
	Level                    uint8
	HP, Resource1, Resource2 uint16
}
```

</details>

## Opcode 0609

**`0x0609` — MobSpeed**

Declarations: [proto.MobSpeed](../../internal/proto/opcodes.go#L240).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobSpeed (server/internal/world/mobeffects.go)](../../server/internal/world/mobeffects.go#L149)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1007)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobSpeed — server/internal/world/mobeffects.go</summary>

[Source definition](../../server/internal/world/mobeffects.go#L149)

```go
func buildMobSpeed(m *MapMob) []byte {
	o := m.mods()
	w := packet.New(0x0609)
	w.U32(m.ID)
	w.U8(o.attackSpeedIndex())
	w.U8(o.moveSpeedIndex())
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

## Opcode 060b

**`0x060B` — MobSkillUse**

Declarations: [proto.MobSkillUse](../../internal/proto/opcodes.go#L235).

Source notes: Monster skill results: native handlers 0x57add0/0x579b80/0x57a2c0.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobSkillPacket (server/internal/world/combat.go)](../../server/internal/world/combat.go#L406)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L940)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobSkillPacket — server/internal/world/combat.go</summary>

[Source definition](../../server/internal/world/combat.go#L406)

```go
func buildMobSkillPacket(op packet.Opcode, result uint8, mobID, target uint32, sk *db.SkillDef, damage [3]uint16, healed *[3]uint16, attackSlot ...uint8) []byte {
	w := packet.New(op)
	w.U8(result)
	w.U32(mobID)
	w.U32(target)
	if op != 0x060D {
		slot := uint8(0)
		if len(attackSlot) > 0 {
			slot = min(attackSlot[0], 2)
		}
		w.U8(slot)
	}
	w.U16(sk.SkillID)
	w.U8(sk.SkillLevel)
	for _, v := range damage {
		w.U16(v)
	}
	if healed != nil {
		for _, v := range healed {
			w.U16(v)
		}
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeSkillResult — internal/proto/skills.go</summary>

[Source definition](../../internal/proto/skills.go#L123)

```go
func DecodeSkillResult(p []byte) (m SkillResult, err error) {
	r := NewReader(p)
	m.Result = r.U8()
	m.Source, m.Target = r.U32(), r.U32()
	if Op(p) == MobSkillUse || Op(p) == MobSkillDrain {
		m.AttackSlot = r.U8()
	}
	m.ID, m.Level = r.U16(), r.U8()
	m.HP, m.SP, m.MP = r.U16(), r.U16(), r.U16()
	if Op(p) == SkillUserDrain || Op(p) == SkillMobDrain || Op(p) == MobSkillDrain {
		m.HealedHP, m.HealedSP, m.HealedMP = r.U16(), r.U16(), r.U16()
	}
	if err := r.Err(); err != nil {
		return m, err
	}
	if r.Remaining() != 0 {
		return m, fmt.Errorf("skill result: %d trailing bytes", r.Remaining())
	}
	return m, nil
}
```

```go
type SkillResult struct {
	Result                       uint8
	AttackSlot                   uint8
	Source, Target               uint32
	ID                           uint16
	Level                        uint8
	HP, SP, MP                   uint16
	HealedHP, HealedSP, HealedMP uint16
}
```

</details>

## Opcode 060c

**`0x060C` — MobSkillDrain**

Declarations: [proto.MobSkillDrain](../../internal/proto/opcodes.go#L236).

Source notes: Monster skill results: native handlers 0x57add0/0x579b80/0x57a2c0.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobSkillPacket (server/internal/world/combat.go)](../../server/internal/world/combat.go#L406)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L940)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobSkillPacket — server/internal/world/combat.go</summary>

[Source definition](../../server/internal/world/combat.go#L406)

```go
func buildMobSkillPacket(op packet.Opcode, result uint8, mobID, target uint32, sk *db.SkillDef, damage [3]uint16, healed *[3]uint16, attackSlot ...uint8) []byte {
	w := packet.New(op)
	w.U8(result)
	w.U32(mobID)
	w.U32(target)
	if op != 0x060D {
		slot := uint8(0)
		if len(attackSlot) > 0 {
			slot = min(attackSlot[0], 2)
		}
		w.U8(slot)
	}
	w.U16(sk.SkillID)
	w.U8(sk.SkillLevel)
	for _, v := range damage {
		w.U16(v)
	}
	if healed != nil {
		for _, v := range healed {
			w.U16(v)
		}
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeSkillResult — internal/proto/skills.go</summary>

[Source definition](../../internal/proto/skills.go#L123)

```go
func DecodeSkillResult(p []byte) (m SkillResult, err error) {
	r := NewReader(p)
	m.Result = r.U8()
	m.Source, m.Target = r.U32(), r.U32()
	if Op(p) == MobSkillUse || Op(p) == MobSkillDrain {
		m.AttackSlot = r.U8()
	}
	m.ID, m.Level = r.U16(), r.U8()
	m.HP, m.SP, m.MP = r.U16(), r.U16(), r.U16()
	if Op(p) == SkillUserDrain || Op(p) == SkillMobDrain || Op(p) == MobSkillDrain {
		m.HealedHP, m.HealedSP, m.HealedMP = r.U16(), r.U16(), r.U16()
	}
	if err := r.Err(); err != nil {
		return m, err
	}
	if r.Remaining() != 0 {
		return m, fmt.Errorf("skill result: %d trailing bytes", r.Remaining())
	}
	return m, nil
}
```

```go
type SkillResult struct {
	Result                       uint8
	AttackSlot                   uint8
	Source, Target               uint32
	ID                           uint16
	Level                        uint8
	HP, SP, MP                   uint16
	HealedHP, HealedSP, HealedMP uint16
}
```

</details>

## Opcode 060d

**`0x060D` — MobSkillOther**

Declarations: [proto.MobSkillOther](../../internal/proto/opcodes.go#L237).

Source notes: Monster skill results: native handlers 0x57add0/0x579b80/0x57a2c0.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobSkillPacket (server/internal/world/combat.go)](../../server/internal/world/combat.go#L406)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L940)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobSkillPacket — server/internal/world/combat.go</summary>

[Source definition](../../server/internal/world/combat.go#L406)

```go
func buildMobSkillPacket(op packet.Opcode, result uint8, mobID, target uint32, sk *db.SkillDef, damage [3]uint16, healed *[3]uint16, attackSlot ...uint8) []byte {
	w := packet.New(op)
	w.U8(result)
	w.U32(mobID)
	w.U32(target)
	if op != 0x060D {
		slot := uint8(0)
		if len(attackSlot) > 0 {
			slot = min(attackSlot[0], 2)
		}
		w.U8(slot)
	}
	w.U16(sk.SkillID)
	w.U8(sk.SkillLevel)
	for _, v := range damage {
		w.U16(v)
	}
	if healed != nil {
		for _, v := range healed {
			w.U16(v)
		}
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeSkillResult — internal/proto/skills.go</summary>

[Source definition](../../internal/proto/skills.go#L123)

```go
func DecodeSkillResult(p []byte) (m SkillResult, err error) {
	r := NewReader(p)
	m.Result = r.U8()
	m.Source, m.Target = r.U32(), r.U32()
	if Op(p) == MobSkillUse || Op(p) == MobSkillDrain {
		m.AttackSlot = r.U8()
	}
	m.ID, m.Level = r.U16(), r.U8()
	m.HP, m.SP, m.MP = r.U16(), r.U16(), r.U16()
	if Op(p) == SkillUserDrain || Op(p) == SkillMobDrain || Op(p) == MobSkillDrain {
		m.HealedHP, m.HealedSP, m.HealedMP = r.U16(), r.U16(), r.U16()
	}
	if err := r.Err(); err != nil {
		return m, err
	}
	if r.Remaining() != 0 {
		return m, fmt.Errorf("skill result: %d trailing bytes", r.Remaining())
	}
	return m, nil
}
```

```go
type SkillResult struct {
	Result                       uint8
	AttackSlot                   uint8
	Source, Target               uint32
	ID                           uint16
	Level                        uint8
	HP, SP, MP                   uint16
	HealedHP, HealedSP, HealedMP uint16
}
```

</details>

## Opcode 060e

**`0x060E` — Literal opcode**

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [applyMobSkill (server/internal/world/combat.go)](../../server/internal/world/combat.go#L279)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>applyMobSkill — server/internal/world/combat.go</summary>

[Source definition](../../server/internal/world/combat.go#L279)

```go
func (s *Session) applyMobSkill(d pendingDamage, md *db.MobDef, mc *mobCombat, sk *db.SkillDef, stats *CharacterStats) bool {
	ch := s.char
	mob, _ := MobByID(ch.MapID, d.MobGlobalID)
	hits := 1
	if sk.TypeDetail == detailMulti {
		hits = int(sk.MultiAttack)
	}
	killed := false
	for i := 0; i < hits && !killed; i++ {
		first := d.First && sk.TypeDetail != detailMulti
		drain := first && sk.TypeDetail == detailDrain
		var damage [3]int32
		result := uint8(attackSuccessNormal)
		if s.blocksAttack(int(sk.Kind)) || s.immuneToLapisSkill(sk) || s.blocksStatusSkill(sk) || !mobSkillHits(mc, sk, stats) {
			result = attackResultMiss
		} else {
			duelHitBy(s.player, 0)
			var crit bool
			damage, crit = mobSkillDamage(mc, md, sk, stats, mob.CurrentHP)
			if crit {
				result = attackResultCritical
			}
		}
		if damage[0] > 0 || damage[1] > 0 || damage[2] > 0 {
			s.breakCombatForm()
			s.breakStealth()
		}
		hp := uint32(max(damage[0], 0))
		if hp > 0 && stats.Absorption > 0 {
			hp = max(satSub(hp, stats.Absorption), 1)
		}
		// CMob::AttackUser: a landed skill goes through CUser::ChkMirror for its
		// category; a mirrored hit spares the player and hurts the mob instead.
		mirrored := result != attackResultMiss && s.reflects(min(int(sk.Kind), 2), md)
		var back [3]uint16
		if mirrored {
			back = [3]uint16{uint16(min(hp, 65535)), uint16(min(max(damage[1], 0), 65535)), uint16(min(max(damage[2], 0), 65535))}
			hp, damage[1], damage[2] = 0, 0, 0
		}
		if damage[1] > 0 || damage[2] > 0 {
			s.spendDelta(uint32(max(damage[2], 0)), uint32(max(damage[1], 0)))
			if hp == 0 {
				s.persist()
			}
		}
		if !mirrored {
			killed = s.takeMobDamage(d, hp, stats)
		}
		if hp > 0 {
			hp = s.assisted[0]
			damage[1], damage[2] = damage[1]+int32(s.assisted[1]), damage[2]+int32(s.assisted[2])
		}
		words := [3]uint16{uint16(min(hp, 65535)), uint16(min(max(damage[1], 0), 65535)), uint16(min(max(damage[2], 0), 65535))}
		var frame []byte
		switch {
		case drain:
			var healed [3]uint16
			if mirrored {
				// the mob is not healed by a mirrored drain
			} else if r, gain, ok := HealMob(ch.MapID, d.MobGlobalID, hp); ok {
				healed[0] = uint16(min(gain, 65535))
				sendNear(ch.MapID, r.PosX, r.PosZ, CombatRadius, buildMobCurrentHP(r.ID, r.CurrentHP), "0x0610 mob hp after drain")
			}
			frame = buildMobSkillPacket(0x060C, result, d.MobGlobalID, ch.ID, sk, words, &healed, d.Slot)
		case i == 0 && (d.First || d.StartsCast):
			frame = buildMobSkillPacket(0x060B, result, d.MobGlobalID, ch.ID, sk, words, nil, d.Slot)
		default:
			frame = buildMobSkillPacket(0x060D, result, d.MobGlobalID, ch.ID, sk, words, nil)
		}
		sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius, frame, "mob skill on character")
		if mirrored {
			s.reflectToMob(d.MobGlobalID, back)
		}
		s.log.Printf("GX Server World mob skill mob_global_id=%d mob_id=%d skill_id=%d result=%d damage=%d killed=%t",
			d.MobGlobalID, d.MobID, sk.SkillID, result, hp, killed)
		if !killed && !drain && result != attackResultMiss && (sk.AddDamageHP != 0 || sk.AddDamageSP != 0 || sk.AddDamageMP != 0) {
			add := extraSkillDamage(sk, 100+mobLevelPenalty(stats.Level, mc.Level))
			extra := [3]uint16{uint16(add[0]), uint16(add[1]), uint16(add[2])}
			addMirrored := mirrored || s.reflects(min(int(sk.Kind), 2), md)
			shownAdd := extra
			if addMirrored {
				shownAdd = [3]uint16{}
			} else {
				if add[1] != 0 || add[2] != 0 {
					s.spendDelta(add[2], add[1])
					if add[0] == 0 {
						s.persist()
					}
				}
				killed = s.takeMobDamage(d, add[0], stats)
			}
			w := packet.New(0x060E)
			w.U32(d.MobGlobalID)
			w.U32(ch.ID)
			w.U16(sk.SkillID)
			w.U8(sk.SkillLevel)
			w.U16(shownAdd[0])
			w.U16(shownAdd[1])
			w.U16(shownAdd[2])
			sendNear(ch.MapID, ch.PosX, ch.PosZ, CombatRadius, w.MustFinish(), "0x060E mob skill extra damage")
			if addMirrored {
				s.reflectToMob(d.MobGlobalID, extra)
			}
		}
		// CMob::AttackUser(skill): a debuff stays on the player.
		if !killed && result != attackResultMiss && sk.TypeEffect == typeEffectDebuff {
			s.applyMobDebuff(sk, d.MobGlobalID)
		}
		if !killed && result != attackResultMiss {
			s.gemCastsOnMob(d.MobGlobalID, armourGemSkills(s.items)) // BeAttackedItemOption(CMob)
		}
		refreshed := s.stats()
		stats = &refreshed
	}
	// A sacrifice skill (detail 23) pays the caster's whole HP after the blow
	// resolves. For a mob ("Self Destruction") that means dying with no credit,
	// drop or experience, exactly like the player's afterSkillStrikes.
	if sk.TypeDetail == detailSacrifice {
		s.eraseMob(d.MobGlobalID)
	}
	s.afterMobDamage(d, killed)
	return killed
}
```

</details>

## Opcode 0610

**`0x0610` — MobCurrentHitpoints**

Declarations: [packet.MobCurrentHitpoints](../../server/internal/packet/packet.go#L116), [proto.MobCurrentHitpoints](../../internal/proto/opcodes.go#L99).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildMobCurrentHP (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1526)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1069)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildMobCurrentHP — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1526)

```go
func buildMobCurrentHP(mobID, hp uint32) []byte {
	w := packet.New(packet.MobCurrentHitpoints)
	w.U32(mobID)
	w.U32(hp)
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

## Opcode 0e01

**`0x0E01` — MapNpcEnter**

Declarations: [packet.MapNpcEnter](../../server/internal/packet/packet.go#L117), [proto.MapNpcEnter](../../internal/proto/opcodes.go#L100).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildNpcEnter (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1328)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L143)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildNpcEnter — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1328)

```go
func buildNpcEnter(n *MapNpc) []byte {
	w := packet.New(packet.MapNpcEnter)
	w.U32(n.ID)
	w.U8(n.Type)
	w.U16(n.TypeID)
	w.F32(n.PosX)
	w.F32(n.PosY)
	w.F32(n.PosZ)
	w.U16(n.Angle)
	if n.Summoned {
		w.Bool(true)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeNpcEnter — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L328)

```go
func DecodeNpcEnter(p []byte) (m NpcSpawn, err error) {
	r := NewReader(p)
	m.ID, m.Type, m.TypeID = r.U32(), r.U8(), r.U16()
	m.X, m.Y, m.Z, m.Angle = r.F32(), r.F32(), r.F32(), r.U16()
	if r.Remaining() > 0 {
		m.Summoned = r.Bool()
	}
	return m, r.Err()
}
```

```go
type NpcSpawn struct {
	Summoned bool
	ID       uint32
	Type     uint8
	TypeID   uint16
	X, Y, Z  float32
	Angle    uint16
}
```

</details>

## Opcode 0e02

**`0x0E02` — MapNpcLeave**

Declarations: [packet.MapNpcLeave](../../server/internal/packet/packet.go#L118), [proto.MapNpcLeave](../../internal/proto/opcodes.go#L101).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildNpcLeave (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1343)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L38)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildNpcLeave — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1343)

```go
func buildNpcLeave(id uint32) []byte {
	w := packet.New(packet.MapNpcLeave)
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

## Opcode 0e03

**`0x0E03` — MapNpcMove**

Declarations: [packet.MapNpcMove](../../server/internal/packet/packet.go#L119), [proto.MapNpcMove](../../internal/proto/opcodes.go#L102).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildNpcMove (server/internal/world/entities.go)](../../server/internal/world/entities.go#L1349)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L149)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildNpcMove — server/internal/world/entities.go</summary>

[Source definition](../../server/internal/world/entities.go#L1349)

```go
func buildNpcMove(n *MapNpc, motion uint8) []byte {
	w := packet.New(packet.MapNpcMove)
	w.U32(n.ID)
	w.U8(motion)
	w.F32(n.PosX)
	w.F32(n.PosY)
	w.F32(n.PosZ)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeNpcMove — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L345)

```go
func DecodeNpcMove(p []byte) (n NpcMotion, err error) {
	r := NewReader(p)
	n.ID, n.Motion, n.X, n.Y, n.Z = r.U32(), r.U8(), r.F32(), r.F32(), r.F32()
	return n, r.Err()
}
```

```go
type NpcMotion struct {
	ID      uint32
	Motion  uint8
	X, Y, Z float32
}
```

</details>

## Opcode 0e04

**`0x0E04` — NpcTalk**

Declarations: [packet.NpcTalk](../../server/internal/packet/packet.go#L141), [proto.NpcTalk](../../internal/proto/opcodes.go#L125).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L119)

**S→C serializers:** [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0e05

**`0x0E05` — MapNpcAttackPlayer**

Declarations: [packet.MapNpcAttackPlayer](../../server/internal/packet/packet.go#L120), [proto.MapNpcAttackPlayer](../../internal/proto/opcodes.go#L103).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L158)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeNpcAttack — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L357)

```go
func DecodeNpcAttack(p []byte) (a NpcAttack, err error) {
	r := NewReader(p)
	a.Result, a.Source, a.Target, a.HP = r.U8(), r.U32(), r.U32(), r.U16()
	return a, r.Err()
}
```

```go
type NpcAttack struct {
	Result         uint8
	Source, Target uint32
	HP             uint16
}
```

</details>

## Opcode 0e06

**`0x0E06` — MapNpcAttackMob**

Declarations: [packet.MapNpcAttackMob](../../server/internal/packet/packet.go#L121), [proto.MapNpcAttackMob](../../internal/proto/opcodes.go#L104).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L1303), [applyEntitiesPacket (internal/game/state_entities_packets.go)](../../internal/game/state_entities_packets.go#L158)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeNpcAttack — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L357)

```go
func DecodeNpcAttack(p []byte) (a NpcAttack, err error) {
	r := NewReader(p)
	a.Result, a.Source, a.Target, a.HP = r.U8(), r.U32(), r.U32(), r.U16()
	return a, r.Err()
}
```

```go
type NpcAttack struct {
	Result         uint8
	Source, Target uint32
	HP             uint16
}
```

</details>
