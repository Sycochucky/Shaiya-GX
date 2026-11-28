# Protocol: Guilds and guild storage

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 0d01

**`0x0D01` — GuildCreateFailed**

Declarations: [packet.GuildCreateFailed](../../server/internal/packet/packet.go#L157), [proto.GuildCreateFailed](../../internal/proto/opcodes.go#L144).

Source notes: Guilds (CUser::PacketGuild and the CGuild senders).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildGuildCreateFailed (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L109)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L397), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildCreateFailed — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L109)

```go
func buildGuildCreateFailed(result uint8, name string) []byte {
	w := packet.New(packet.GuildCreateFailed)
	w.U8(result)
	w.U32(0)
	w.U8(0)
	w.FixedString(name, 25)
	w.Bytes(make([]byte, 98-2-31)) // the original sends a 98-byte frame
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeGuildCreateFailure — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L68)

```go
func DecodeGuildCreateFailure(p []byte) (uint8, error) {
	r := NewReader(p)
	result := r.U8()
	r.Bytes(95)
	return result, r.Err()
}
```

</details>

## Opcode 0d03

**`0x0D03` — GuildDisband**

Declarations: [packet.GuildDisband](../../server/internal/packet/packet.go#L158), [proto.GuildDisband](../../internal/proto/opcodes.go#L145).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L332)

**S→C serializers:** [guildResult (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L138)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L418), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildResult — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L138)

```go
func guildResult(op packet.Opcode, result uint8) []byte {
	w := packet.New(op)
	w.U8(result)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildDisbandPacket — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L136)

```go
func GuildDisbandPacket() []byte { return NewWriter(GuildDisband).Bytes() }
```

</details>

<details><summary>DecodeU8 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L194)

```go
func DecodeU8(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 0d04

**`0x0D04` — GuildDisbandAgree**

Declarations: [packet.GuildDisbandAgree](../../server/internal/packet/packet.go#L159), [proto.GuildDisbandAgree](../../internal/proto/opcodes.go#L146).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L334)

**S→C serializers:** [guildID32 (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L151)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L425), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildID32 — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L151)

```go
func guildID32(op packet.Opcode, id uint32) []byte {
	w := packet.New(op)
	w.U32(id)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildDisbandAnswerPacket — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L137)

```go
func GuildDisbandAnswerPacket(accept bool) []byte {
	w := NewWriter(GuildDisbandAgree)
	if accept {
		w.U8(1)
	} else {
		w.U8(0)
	}
	return w.Bytes()
}
```

</details>

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0d07

**`0x0D07` — GuildApply**

Declarations: [packet.GuildApply](../../server/internal/packet/packet.go#L160), [proto.GuildApply](../../internal/proto/opcodes.go#L147).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L338)

**S→C serializers:** [guildResult (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L138), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L310), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildResult — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L138)

```go
func guildResult(op packet.Opcode, result uint8) []byte {
	w := packet.New(op)
	w.U8(result)
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

<details><summary>DecodeU8 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L194)

```go
func DecodeU8(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 0d08

**`0x0D08` — GuildJoinResult**

Declarations: [packet.GuildJoinResult](../../server/internal/packet/packet.go#L161), [proto.GuildJoinResult](../../internal/proto/opcodes.go#L148).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L343)

**S→C serializers:** [buildGuildJoined (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L129)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L359), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildJoined — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L129)

```go
func buildGuildJoined(result uint8, g *guild, rank uint8) []byte {
	w := packet.New(packet.GuildJoinResult)
	w.U8(result)
	w.U32(g.ID)
	w.U8(rank)
	w.FixedString(g.Name, 25)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildApplicantAnswerPacket — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L163)

```go
func GuildApplicantAnswerPacket(id uint32, accept bool) []byte {
	w := NewWriter(GuildJoinResult)
	if accept {
		w.U8(1)
	} else {
		w.U8(0)
	}
	w.U32(id)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeGuildJoin — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L122)

```go
func DecodeGuildJoin(p []byte) (GuildJoin, error) {
	r := NewReader(p)
	v := GuildJoin{Result: r.U8(), ID: r.U32(), Rank: r.U8(), Name: r.Fixed(25)}
	return v, r.Err()
}
```

```go
type GuildJoin struct {
	Result uint8
	ID     uint32
	Rank   uint8
	Name   string
}
```

</details>

## Opcode 0d09

**`0x0D09` — GuildLeave**

Declarations: [packet.GuildLeave](../../server/internal/packet/packet.go#L162), [proto.GuildLeave](../../internal/proto/opcodes.go#L149).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L348)

**S→C serializers:** [guildResult (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L138)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L404), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildResult — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L138)

```go
func guildResult(op packet.Opcode, result uint8) []byte {
	w := packet.New(op)
	w.U8(result)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildLeavePacket — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L135)

```go
func GuildLeavePacket() []byte   { return NewWriter(GuildLeave).Bytes() }
```

</details>

<details><summary>DecodeU8 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L194)

```go
func DecodeU8(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 0d0a

**`0x0D0A` — GuildExpel**

Declarations: [packet.GuildExpel](../../server/internal/packet/packet.go#L163), [proto.GuildExpel](../../internal/proto/opcodes.go#L150).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L350)

**S→C serializers:** [guildResultID (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L144), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L432), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildResultID — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L144)

```go
func guildResultID(op packet.Opcode, result uint8, id uint32) []byte {
	w := packet.New(op)
	w.U8(result)
	w.U32(id)
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

<details><summary>DecodePartyReply — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L68)

```go
func DecodePartyReply(p []byte) (PartyReply, error) {
	r := NewReader(p)
	m := PartyReply{r.U8(), r.U32()}
	return m, r.Err()
}
```

```go
type PartyReply struct {
	Result uint8
	ID     uint32
}
```

</details>

## Opcode 0d0c

**`0x0D0C` — GuildStatus**

Declarations: [packet.GuildStatus](../../server/internal/packet/packet.go#L164), [proto.GuildStatus](../../internal/proto/opcodes.go#L151).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L355)

**S→C serializers:** [guildStatus (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L7)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L317), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildStatus — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L7)

```go
func guildStatus(mode uint8, charID uint32) []byte {
	w := packet.New(packet.GuildStatus)
	w.U8(mode)
	w.U32(charID)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildRankPacket — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L173)

```go
func GuildRankPacket(id uint32, demote bool) []byte {
	w := NewWriter(GuildStatus)
	if demote {
		w.U8(1)
	} else {
		w.U8(0)
	}
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

## Opcode 0d0d

**`0x0D0D` — GuildListBegin**

Declarations: [packet.GuildListBegin](../../server/internal/packet/packet.go#L165), [proto.GuildListBegin](../../internal/proto/opcodes.go#L152).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [guildListPackets (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L171)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L218), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildListPackets — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L171)

```go
func guildListPackets(country uint8) [][]byte {
	out := [][]byte{packet.New(packet.GuildListBegin).MustFinish()}
	var batch []*guild
	flush := func() {
		if len(batch) == 0 {
			return
		}
		w := packet.New(packet.GuildList)
		w.U8(uint8(len(batch)))
		for _, g := range batch {
			writeGuildInfo(w, g)
		}
		out = append(out, w.MustFinish())
		batch = nil
	}
	for _, g := range guildsByID {
		if g.Country == country {
			batch = append(batch, g)
			if len(batch) == guildListBatch {
				flush()
			}
		}
	}
	flush()
	return append(out, packet.New(packet.GuildListEnd).MustFinish())
}
```

</details>

## Opcode 0d0e

**`0x0D0E` — GuildListEnd**

Declarations: [packet.GuildListEnd](../../server/internal/packet/packet.go#L166), [proto.GuildListEnd](../../internal/proto/opcodes.go#L153).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [guildListPackets (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L171)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L221), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildListPackets — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L171)

```go
func guildListPackets(country uint8) [][]byte {
	out := [][]byte{packet.New(packet.GuildListBegin).MustFinish()}
	var batch []*guild
	flush := func() {
		if len(batch) == 0 {
			return
		}
		w := packet.New(packet.GuildList)
		w.U8(uint8(len(batch)))
		for _, g := range batch {
			writeGuildInfo(w, g)
		}
		out = append(out, w.MustFinish())
		batch = nil
	}
	for _, g := range guildsByID {
		if g.Country == country {
			batch = append(batch, g)
			if len(batch) == guildListBatch {
				flush()
			}
		}
	}
	flush()
	return append(out, packet.New(packet.GuildListEnd).MustFinish())
}
```

</details>

## Opcode 0d11

**`0x0D11` — GuildDisbanded**

Declarations: [packet.GuildDisbanded](../../server/internal/packet/packet.go#L167), [proto.GuildDisbanded](../../internal/proto/opcodes.go#L154).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [guildID32 (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L151)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L249), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildID32 — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L151)

```go
func guildID32(op packet.Opcode, id uint32) []byte {
	w := packet.New(op)
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

## Opcode 0d12

**`0x0D12` — GuildOnlineList**

Declarations: [packet.GuildOnlineList](../../server/internal/packet/packet.go#L168), [proto.GuildOnlineList](../../internal/proto/opcodes.go#L155).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L256), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeGuildMembers — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L102)

```go
func DecodeGuildMembers(p []byte, online bool) ([]GuildMember, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/29 {
		return nil, ErrShort
	}
	list := make([]GuildMember, n)
	for i := range list {
		list[i] = readGuildMember(r, online)
	}
	return list, r.Err()
}
```

```go
type GuildMember struct {
	ID     uint32
	Rank   uint8
	Level  uint16
	Class  uint8
	Name   string
	Online bool
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

## Opcode 0d13

**`0x0D13` — GuildOfflineList**

Declarations: [packet.GuildOfflineList](../../server/internal/packet/packet.go#L169), [proto.GuildOfflineList](../../internal/proto/opcodes.go#L156).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L256), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeGuildMembers — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L102)

```go
func DecodeGuildMembers(p []byte, online bool) ([]GuildMember, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/29 {
		return nil, ErrShort
	}
	list := make([]GuildMember, n)
	for i := range list {
		list[i] = readGuildMember(r, online)
	}
	return list, r.Err()
}
```

```go
type GuildMember struct {
	ID     uint32
	Rank   uint8
	Level  uint16
	Class  uint8
	Name   string
	Online bool
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

## Opcode 0d14

**`0x0D14` — GuildMemberAdded**

Declarations: [packet.GuildMemberAdded](../../server/internal/packet/packet.go#L170), [proto.GuildMemberAdded](../../internal/proto/opcodes.go#L157).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildGuildMemberAdded (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L159)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L271), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildMemberAdded — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L159)

```go
func buildGuildMemberAdded(m *guildMember) []byte {
	w := packet.New(packet.GuildMemberAdded)
	if m.player != nil {
		w.U8(1)
	} else {
		w.U8(0)
	}
	writeGuildMember(w, m)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeGuildMemberAdded — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L115)

```go
func DecodeGuildMemberAdded(p []byte) (GuildMember, error) {
	r := NewReader(p)
	online := r.U8() != 0
	m := readGuildMember(r, online)
	return m, r.Err()
}
```

```go
type GuildMember struct {
	ID     uint32
	Rank   uint8
	Level  uint16
	Class  uint8
	Name   string
	Online bool
}
```

</details>

## Opcode 0d16

**`0x0D16` — GuildApplicantList**

Declarations: [packet.GuildApplicantList](../../server/internal/packet/packet.go#L171), [proto.GuildApplicantList](../../internal/proto/opcodes.go#L158).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [applicantListPackets (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L65)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L284), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>applicantListPackets — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L65)

```go
func (g *guild) applicantListPackets() [][]byte {
	var out [][]byte
	var batch []*guildMember
	flush := func() {
		if len(batch) == 0 {
			return
		}
		w := packet.New(packet.GuildApplicantList)
		w.U8(uint8(len(batch)))
		for _, m := range batch {
			writeGuildApplicant(w, m)
		}
		out = append(out, w.MustFinish())
		batch = nil
	}
	for _, m := range g.applicants {
		batch = append(batch, m)
		if len(batch) == guildMemberBatch {
			flush()
		}
	}
	flush()
	return out
}
```

</details>

<details><summary>DecodeGuildApplicants — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L31)

```go
func DecodeGuildApplicants(p []byte) ([]GuildCandidate, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/28 {
		return nil, ErrShort
	}
	list := make([]GuildCandidate, n)
	for i := range list {
		list[i] = readGuildApplicant(r)
	}
	return list, r.Err()
}
```

```go
type GuildCandidate struct {
	ID    uint32
	Level uint16
	Class uint8
	Name  string
}
```

</details>

## Opcode 0d17

**`0x0D17` — GuildApplicant**

Declarations: [packet.GuildApplicant](../../server/internal/packet/packet.go#L172), [proto.GuildApplicant](../../internal/proto/opcodes.go#L159).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [guildApply (server/internal/world/guild_membership.go)](../../server/internal/world/guild_membership.go#L8)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L295), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildApply — server/internal/world/guild_membership.go</summary>

[Source definition](../../server/internal/world/guild_membership.go#L8)

```go
func (s *Session) guildApply(guildID uint32) {
	me, ch := s.player, s.char
	guildMu.Lock()
	defer guildMu.Unlock()
	g := guildsByID[guildID]
	if me.guild.guild != nil || me.guild.create != nil || g == nil ||
		g.Country != s.faction || g.applicants[ch.ID] != nil {
		s.send(guildResult(packet.GuildApply, 0), "0x0D07 guild apply refused")
		return
	}
	if old := me.guild.applied; old != nil {
		old.cancelApplication(me)
	}
	m := &guildMember{CharID: ch.ID, Name: ch.Name, Job: ch.Class, Level: ch.Level}
	if err := db.AddGuildChar(g.ID, ch.ID, 0); err != nil {
		s.log.Errorf("GX Server World guild application insert failed: %v", err)
		return
	}
	g.applicants[ch.ID] = m
	guildOfChar[ch.ID] = g
	me.guild.applied = g
	w := packet.New(packet.GuildApplicant)
	writeGuildApplicant(w, m)
	g.sendAdmin(w.MustFinish(), "0x0D17 guild applicant")
	s.send(guildResult(packet.GuildApply, 1), "0x0D07 guild applied")
}
```

</details>

<details><summary>DecodeGuildApplicant — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L44)

```go
func DecodeGuildApplicant(p []byte) (GuildCandidate, error) {
	r := NewReader(p)
	v := readGuildApplicant(r)
	return v, r.Err()
}
```

```go
type GuildCandidate struct {
	ID    uint32
	Level uint16
	Class uint8
	Name  string
}
```

</details>

## Opcode 0d18

**`0x0D18` — GuildApplicantRemoved**

Declarations: [packet.GuildApplicantRemoved](../../server/internal/packet/packet.go#L173), [proto.GuildApplicantRemoved](../../internal/proto/opcodes.go#L160).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [guildID32 (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L151)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L304), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildID32 — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L151)

```go
func guildID32(op packet.Opcode, id uint32) []byte {
	w := packet.New(op)
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

## Opcode 0d20

**`0x0D20` — GuildRemark**

Declarations: [packet.GuildRemark](../../server/internal/packet/packet.go#L174), [proto.GuildRemark](../../internal/proto/opcodes.go#L161).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L360)

**S→C serializers:** [guildRemark (server/internal/world/guild.go)](../../server/internal/world/guild.go#L513)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L147), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildRemark — server/internal/world/guild.go</summary>

[Source definition](../../server/internal/world/guild.go#L513)

```go
func (s *Session) guildRemark(p []byte) {
	remark := fixedText(p, 2, 64)
	reply := func(result uint8) {
		w := packet.New(packet.GuildRemark)
		w.U8(result)
		w.FixedString(remark, 65)
		s.send(w.MustFinish(), "0x0D20 guild remark")
	}
	guildMu.Lock()
	defer guildMu.Unlock()
	g := s.player.guild.guild
	if g == nil {
		return
	}
	if m := g.members[s.char.ID]; m == nil || m.Rank != guildRankMaster || strings.ContainsRune(remark, '\'') {
		reply(0)
		return
	}
	if err := db.SaveGuildRemark(g.ID, remark); err != nil {
		s.logErr("guild remark", err)
		reply(0)
		return
	}
	g.Remark = remark
	w := packet.New(packet.GuildRemark)
	w.U8(1)
	w.FixedString(remark, 65)
	g.send(w.MustFinish(), "0x0D20 guild remark", nil)

}
```

</details>

<details><summary>GuildRemarkPacket — internal/proto/guild_notice.go</summary>

[Source definition](../../internal/proto/guild_notice.go#L10)

```go
func GuildRemarkPacket(remark string) []byte {
	w := NewWriter(GuildRemark)
	w.Fixed(remark, 65, true)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeGuildRemark — internal/proto/guild_notice.go</summary>

[Source definition](../../internal/proto/guild_notice.go#L15)

```go
func DecodeGuildRemark(p []byte) (GuildRemarkResult, error) {
	r := NewReader(p)
	v := GuildRemarkResult{Result: r.U8(), Remark: r.Fixed(65)}
	if err := r.Err(); err != nil {
		return v, err
	}
	if r.Remaining() != 0 {
		return v, fmt.Errorf("guild remark: trailing bytes")
	}
	return v, nil
}
```

```go
type GuildRemarkResult struct {
	Result uint8
	Remark string
}
```

</details>

## Opcode 0d21

**`0x0D21` — GuildCreate**

Declarations: [packet.GuildCreate](../../server/internal/packet/packet.go#L182), [proto.GuildCreate](../../internal/proto/opcodes.go#L169).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L326)

**S→C serializers:** [buildGuildCreateResult (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L119)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L376), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildCreateResult — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L119)

```go
func buildGuildCreateResult(result uint8, id uint32, rank uint8, name, remark string) []byte {
	w := packet.New(packet.GuildCreate)
	w.U8(result)
	w.U32(id)
	w.U8(rank)
	w.FixedString(name, 25)
	w.FixedString(remark, 65)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildCreatePacket — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L148)

```go
func GuildCreatePacket(name, remark string) []byte {
	w := NewWriter(GuildCreate)
	w.Fixed(name, 25, true)
	w.Fixed(remark, 65, true)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeGuildCreate — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L128)

```go
func DecodeGuildCreate(p []byte) (GuildJoin, string, error) {
	r := NewReader(p)
	v := GuildJoin{Result: r.U8(), ID: r.U32(), Rank: r.U8(), Name: r.Fixed(25)}
	remark := r.Fixed(65)
	return v, remark, r.Err()
}
```

```go
type GuildJoin struct {
	Result uint8
	ID     uint32
	Rank   uint8
	Name   string
}
```

</details>

## Opcode 0d22

**`0x0D22` — GuildCreateAgree**

Declarations: [packet.GuildCreateAgree](../../server/internal/packet/packet.go#L183), [proto.GuildCreateAgree](../../internal/proto/opcodes.go#L170).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L328)

**S→C serializers:** [guildCreateRequest (server/internal/world/guild_creation.go)](../../server/internal/world/guild_creation.go#L11)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L390), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildCreateRequest — server/internal/world/guild_creation.go</summary>

[Source definition](../../server/internal/world/guild_creation.go#L11)

```go
func (s *Session) guildCreateRequest(p []byte) {
	ch, me := s.char, s.player
	name, remark := strings.TrimSpace(fixedText(p, 2, 24)), fixedText(p, 0x1b, 64)
	fail := func(result uint8) {
		s.send(buildGuildCreateResult(result, 0, 0, name, remark), "0x0D21 guild create refused")
	}
	if name == "" {
		fail(6)
		return
	}
	partyMu.Lock()
	group := partyOf(me)
	players := []*ConnectedPlayer{me}
	if group != nil {
		players = append([]*ConnectedPlayer(nil), group.members...)
	}
	partyMu.Unlock()
	if group != nil && len(players) < 2 {
		fail(3)
		return
	}
	guildMu.Lock()
	defer guildMu.Unlock()
	if me.guild.guild != nil || me.guild.create != nil {
		fail(8)
		return
	}
	for _, m := range players {
		switch {
		case m.guild.guild != nil:
			fail(8)
			return
		case m.guild.create != nil || m.MapID() != ch.MapID || m.Faction() != me.Faction():
			fail(3)
			return
		}
	}
	for _, g := range guildsByID {
		if strings.EqualFold(g.Name, name) {
			fail(6)
			return
		}
	}
	c := &guildCreate{name: name, remark: remark, masterID: ch.ID, master: me, pending: map[uint32]bool{}}
	for _, m := range players {
		if g := m.guild.applied; g != nil {
			g.cancelApplication(m)
			if m.guild.applied != nil {
				fail(0)
				return
			}
		}
	}
	c.players = players
	if group == nil {
		c.complete(s.logger, s)
		return
	}
	for _, m := range players {
		c.pending[m.CharacterID] = true
		m.guild.create = c
	}
	// SendGuildCreateAgree (0x0D22): master id, name, remark to all.
	w := packet.New(packet.GuildCreateAgree)
	w.U32(ch.ID)
	w.FixedString(name, 25)
	w.FixedString(remark, 65)
	c.sendAll(w.MustFinish(), "0x0D22 guild create agree")
}
```

</details>

<details><summary>GuildCreateAnswerPacket — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L154)

```go
func GuildCreateAnswerPacket(accept bool) []byte {
	w := NewWriter(GuildCreateAgree)
	if accept {
		w.U8(1)
	} else {
		w.U8(0)
	}
	return w.Bytes()
}
```

</details>

<details><summary>DecodeGuildCreateVote — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L62)

```go
func DecodeGuildCreateVote(p []byte) (GuildCreateVote, error) {
	r := NewReader(p)
	v := GuildCreateVote{Master: r.U32(), Name: r.Fixed(25), Remark: r.Fixed(65)}
	return v, r.Err()
}
```

```go
type GuildCreateVote struct {
	Master       uint32
	Name, Remark string
}
```

</details>

## Opcode 0d23

**`0x0D23` — GuildEtin**

Declarations: [packet.GuildEtin](../../server/internal/packet/packet.go#L184), [proto.GuildEtin](../../internal/proto/opcodes.go#L171).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L362)

**S→C serializers:** [guildID32 (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L151)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L85), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildID32 — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L151)

```go
func guildID32(op packet.Opcode, id uint32) []byte {
	w := packet.New(op)
	w.U32(id)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildEtinPacket — internal/proto/guild_notice.go</summary>

[Source definition](../../internal/proto/guild_notice.go#L26)

```go
func GuildEtinPacket() []byte { return NewWriter(GuildEtin).Bytes() }
```

</details>

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0d24

**`0x0D24` — GuildNpcList**

Declarations: [packet.GuildNpcList](../../server/internal/packet/packet.go#L185), [proto.GuildNpcList](../../internal/proto/opcodes.go#L172).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [npcListPacket (server/internal/world/guildhouse.go)](../../server/internal/world/guildhouse.go#L285)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L91), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>npcListPacket — server/internal/world/guildhouse.go</summary>

[Source definition](../../server/internal/world/guildhouse.go#L285)

```go
func (g *guild) npcListPacket() []byte {
	var entries []guildNpcSlot
	var slots []uint8
	for i, sl := range g.npcs {
		if sl.Type != 0 {
			entries, slots = append(entries, sl), append(slots, uint8(i))
		}
	}
	if len(entries) == 0 {
		return nil
	}
	w := packet.New(packet.GuildNpcList)
	w.U8(uint8(len(entries)))
	for i, sl := range entries {
		w.U16(sl.Type)
		w.U8(uint8(sl.Level))
		w.U8(slots[i])
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeGuildNpcList — internal/proto/guild_house.go</summary>

[Source definition](../../internal/proto/guild_house.go#L26)

```go
func DecodeGuildNpcList(p []byte) ([]GuildNpc, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/4 {
		return nil, ErrShort
	}
	out := make([]GuildNpc, n)
	for i := range out {
		out[i] = GuildNpc{Type: r.U16(), Level: r.U8(), Slot: r.U8()}
	}
	return out, r.Err()
}
```

```go
type GuildNpc struct {
	Type  uint16
	Level uint8
	Slot  uint8
}
```

</details>

## Opcode 0d25

**`0x0D25` — GuildNpcUpgrade**

Declarations: [packet.GuildNpcUpgrade](../../server/internal/packet/packet.go#L186), [proto.GuildNpcUpgrade](../../internal/proto/opcodes.go#L173).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L373)

**S→C serializers:** [buildGuildNpcResult (server/internal/world/guildhouse.go)](../../server/internal/world/guildhouse.go#L306)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L104), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildNpcResult — server/internal/world/guildhouse.go</summary>

[Source definition](../../server/internal/world/guildhouse.go#L306)

```go
func buildGuildNpcResult(result uint8, npcType uint16, level, slot uint8) []byte {
	w := packet.New(packet.GuildNpcUpgrade)
	w.U8(result)
	w.U16(npcType)
	w.U8(level)
	w.U8(slot)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildNpcUpgradePacket — internal/proto/guild_house.go</summary>

[Source definition](../../internal/proto/guild_house.go#L66)

```go
func GuildNpcUpgradePacket(npcType uint16, level uint8) []byte {
	w := NewWriter(GuildNpcUpgrade)
	w.U16(npcType)
	w.U8(level)
	return w.Bytes()
}
```

</details>

<details><summary>DecodeGuildNpcUpgrade — internal/proto/guild_house.go</summary>

[Source definition](../../internal/proto/guild_house.go#L39)

```go
func DecodeGuildNpcUpgrade(p []byte) (GuildNpcUpgradeResult, error) {
	r := NewReader(p)
	v := GuildNpcUpgradeResult{Result: r.U8()}
	v.Npc = GuildNpc{Type: r.U16(), Level: r.U8(), Slot: r.U8()}
	return v, r.Err()
}
```

```go
type GuildNpc struct {
	Type  uint16
	Level uint8
	Slot  uint8
}
```

```go
type GuildNpcUpgradeResult struct {
	Result uint8
	Npc    GuildNpc
}
```

</details>

## Opcode 0d26

**`0x0D26` — GuildRankingNotice**

Declarations: [packet.GuildRankingNotice](../../server/internal/packet/packet.go#L175), [proto.GuildRankingNotice](../../internal/proto/opcodes.go#L162).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [TickGuildRanking (server/internal/world/grb.go)](../../server/internal/world/grb.go#L270)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L161), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>TickGuildRanking — server/internal/world/grb.go</summary>

[Source definition](../../server/internal/world/grb.go#L270)

```go
func TickGuildRanking(now time.Time, log *logx.Logger) {
	minute := now.Unix() / 60
	grbState.Lock()
	newMinute := minute != grbState.lastMinute
	grbState.lastMinute = minute
	grbState.Unlock()
	for _, mapID := range grbMaps() {
		cfg := mapConfig.Zones[mapID]
		open := grbOpen(grbSlots(cfg), now)
		if newMinute {
			if t, ok := grbNotice(grbSlots(cfg), now); ok {
				w := packet.New(packet.GuildRankingNotice)
				w.U16(mapID)
				w.U8(t)
				broadcast(AllPlayers(), w.MustFinish(), "0x0D26 guild ranking notice", log)
			}
		}
		grbState.Lock()
		wasOpen := grbState.open[mapID]
		grbState.open[mapID] = open
		if wasOpen && !open {
			grbState.rankAt = now.Add(grbRankDelay)
		}
		grbState.Unlock()
		if !open {
			grbCloseInstances(mapID, log)
		}
	}
	grbState.Lock()
	due := !grbState.rankAt.IsZero() && !now.Before(grbState.rankAt)
	if due {
		grbState.rankAt = time.Time{}
	}
	grbState.Unlock()
	if due {
		updateGuildRanking(log)
	}
}
```

</details>

<details><summary>DecodeGuildRankingNotice — internal/proto/guild_ranking.go</summary>

[Source definition](../../internal/proto/guild_ranking.go#L17)

```go
func DecodeGuildRankingNotice(p []byte) (GuildRankingNoticeInfo, error) {
	r := NewReader(p)
	v := GuildRankingNoticeInfo{Map: r.U16(), Kind: r.U8()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if r.Remaining() != 0 {
		return v, fmt.Errorf("guild ranking notice: trailing bytes")
	}
	return v, nil
}
```

```go
type GuildRankingNoticeInfo struct {
	Map  uint16
	Kind uint8
}
```

</details>

## Opcode 0d27

**`0x0D27` — GuildRankingPoints**

Declarations: [packet.GuildRankingPoints](../../server/internal/packet/packet.go#L176), [proto.GuildRankingPoints](../../internal/proto/opcodes.go#L163).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [grbKill (server/internal/world/grb.go)](../../server/internal/world/grb.go#L228)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L167), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>grbKill — server/internal/world/grb.go</summary>

[Source definition](../../server/internal/world/grb.go#L228)

```go
func grbKill(zone uint16, mob *MapMob, killerID uint32) bool {
	if mapCreateType(zone) != grbCreateType || mob.MoneyMin != 0 || mob.MoneyMax == 0 {
		return false
	}
	p := FindPlayerByID(killerID)
	if p == nil {
		return true
	}
	guildMu.Lock()
	defer guildMu.Unlock()
	g := p.guild.guild
	if g == nil {
		return true
	}
	mapID := realMap(zone)
	grbState.Lock()
	if owned, found := findInstance(mapID, g.ID); !found || owned != zone || !g.gmGRB && !grbOpen(grbSlots(mapConfig.Zones[mapID]), grbNow()) {
		grbState.Unlock()
		return true
	}
	g.grbPoints = satAdd(g.grbPoints, uint32(min(mob.MoneyMax, 65535)))
	c := int(g.Country) & 1
	if g.grbPoints > grbState.maxPoints[c] {
		grbState.maxPoints[c], grbState.maxPointsID[c] = g.grbPoints, g.ID
	}
	maxPoints, maxID := grbState.maxPoints[c], grbState.maxPointsID[c]
	grbState.Unlock()
	w := packet.New(packet.GuildRankingPoints)
	w.U32(g.grbPoints)
	w.U32(maxPoints)
	w.U32(maxID)
	frame := w.MustFinish()
	for _, m := range g.members {
		if m.player != nil && mapCreateType(m.player.MapID()) == grbCreateType {
			m.player.Send(frame, "0x0D27 guild ranking points")
		}
	}
	return true
}
```

</details>

<details><summary>DecodeGuildRankingPoints — internal/proto/guild_ranking.go</summary>

[Source definition](../../internal/proto/guild_ranking.go#L28)

```go
func DecodeGuildRankingPoints(p []byte) (GuildRankingPointsInfo, error) {
	r := NewReader(p)
	v := GuildRankingPointsInfo{Points: r.U32(), Maximum: r.U32(), Leader: r.U32()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if r.Remaining() != 0 {
		return v, fmt.Errorf("guild ranking points: trailing bytes")
	}
	return v, nil
}
```

```go
type GuildRankingPointsInfo struct{ Points, Maximum, Leader uint32 }
```

</details>

## Opcode 0d28

**`0x0D28` — GuildBankResult**

Declarations: [packet.GuildBankResult](../../server/internal/packet/packet.go#L178), [proto.GuildBankResult](../../internal/proto/opcodes.go#L165).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildGuildBankResult (server/internal/world/guildbank.go)](../../server/internal/world/guildbank.go#L94)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L211), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildBankResult — server/internal/world/guildbank.go</summary>

[Source definition](../../server/internal/world/guildbank.go#L94)

```go
func buildGuildBankResult(result, arg uint8) []byte {
	w := packet.New(packet.GuildBankResult)
	w.U8(result)
	w.U8(arg)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeGuildBankResult — internal/proto/guild_bank.go</summary>

[Source definition](../../internal/proto/guild_bank.go#L52)

```go
func DecodeGuildBankResult(p []byte) (GuildBankResultInfo, error) {
	r := NewReader(p)
	result := GuildBankResultInfo{Result: r.U8(), Arg: r.U8()}
	return result, r.Err()
}
```

```go
type GuildBankResultInfo struct{ Result, Arg uint8 }
```

</details>

## Opcode 0d29

**`0x0D29` — GuildBankList**

Declarations: [packet.GuildBankList](../../server/internal/packet/packet.go#L179), [proto.GuildBankList](../../internal/proto/opcodes.go#L166).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [bankListPackets (server/internal/world/guildbank.go)](../../server/internal/world/guildbank.go#L41)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L187), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>bankListPackets — server/internal/world/guildbank.go</summary>

[Source definition](../../server/internal/world/guildbank.go#L41)

```go
func (g *guild) bankListPackets() [][]byte {
	var out [][]byte
	var batch []*db.Item
	flush := func() {
		if len(batch) == 0 {
			return
		}
		w := packet.New(packet.GuildBankList)
		w.U8(uint8(len(batch)))
		for _, it := range batch {
			w.U8(it.Slot)
			w.U8(it.Type)
			w.U8(it.TypeID)
			w.U16(it.Quality)
			w.Bytes(it.Gems[:])
			w.Count(it.Count)
			w.FixedString(it.CraftName, 21)
		}
		out = append(out, w.MustFinish())
		batch = nil
	}
	for _, it := range g.bank {
		if it != nil {
			batch = append(batch, it)
			if len(batch) == guildBankBatch {
				flush()
			}
		}
	}
	flush()
	return out
}
```

</details>

<details><summary>DecodeGuildBankList — internal/proto/guild_bank.go</summary>

[Source definition](../../internal/proto/guild_bank.go#L22)

```go
func DecodeGuildBankList(p []byte) ([]InventoryItem, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/33 {
		return nil, ErrShort
	}
	items := make([]InventoryItem, n)
	var seen [GuildBankSlots]bool
	for i := range items {
		items[i] = readGuildBankItem(r)
		if int(items[i].Slot) >= GuildBankSlots || seen[items[i].Slot] {
			return nil, fmt.Errorf("guild bank: invalid or duplicate slot %d", items[i].Slot)
		}
		seen[items[i].Slot] = true
	}
	return items, r.Err()
}
```

</details>

## Opcode 0d2f

**`0x0D2F` — GuildList**

Declarations: [packet.GuildList](../../server/internal/packet/packet.go#L190), [proto.GuildList](../../internal/proto/opcodes.go#L177).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [guildListPackets (server/internal/world/guild_packets.go)](../../server/internal/world/guild_packets.go#L171)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L223), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildListPackets — server/internal/world/guild_packets.go</summary>

[Source definition](../../server/internal/world/guild_packets.go#L171)

```go
func guildListPackets(country uint8) [][]byte {
	out := [][]byte{packet.New(packet.GuildListBegin).MustFinish()}
	var batch []*guild
	flush := func() {
		if len(batch) == 0 {
			return
		}
		w := packet.New(packet.GuildList)
		w.U8(uint8(len(batch)))
		for _, g := range batch {
			writeGuildInfo(w, g)
		}
		out = append(out, w.MustFinish())
		batch = nil
	}
	for _, g := range guildsByID {
		if g.Country == country {
			batch = append(batch, g)
			if len(batch) == guildListBatch {
				flush()
			}
		}
	}
	flush()
	return append(out, packet.New(packet.GuildListEnd).MustFinish())
}
```

</details>

<details><summary>DecodeGuildList — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L79)

```go
func DecodeGuildList(p []byte) ([]GuildInfo, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/120 {
		return nil, ErrShort
	}
	list := make([]GuildInfo, n)
	for i := range list {
		list[i] = readGuildInfo(r)
	}
	return list, r.Err()
}
```

```go
type GuildInfo struct {
	ID                   uint32
	Name, Master, Remark string
	Rank                 uint8
	Points               uint32
}
```

</details>

## Opcode 0d30

**`0x0D30` — GuildNew**

Declarations: [packet.GuildNew](../../server/internal/packet/packet.go#L191), [proto.GuildNew](../../internal/proto/opcodes.go#L178).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [complete (server/internal/world/guild_creation.go)](../../server/internal/world/guild_creation.go#L130)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L237), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>complete — server/internal/world/guild_creation.go</summary>

[Source definition](../../server/internal/world/guild_creation.go#L130)

```go
func (c *guildCreate) complete(log *logx.Logger, caller *Session) bool {
	for _, g := range guildsByID {
		if strings.EqualFold(g.Name, c.name) {
			c.sendAll(buildGuildCreateFailed(6, c.name), "0x0D01 guild create failed")
			return true
		}
	}
	owner := c.master.owner
	if owner == nil {
		c.sendAll(buildGuildCreateFailed(6, c.name), "guild master unavailable")
		return true
	}
	if owner != caller {
		if !owner.stateMu.TryLock() {
			return false
		}
		defer owner.stateMu.Unlock()
	}
	if owner.char == nil || owner.player != c.master {
		return true
	}
	master := *owner.char
	fee := guildHouse.CreateMoney
	if master.Money < fee {
		c.sendAll(buildGuildCreateFailed(2, c.name), "0x0D01 guild create failed")
		return true
	}
	g := &guild{ID: nextGuildID, Name: c.name, Master: master.Name, Remark: c.remark, Country: c.master.Faction(),
		Rank: 31, gmHouse: owner.isAdmin, members: map[uint32]*guildMember{}, applicants: map[uint32]*guildMember{}}
	g.forceGMHouse()
	founders := make([]db.GuildCharRow, 0, len(c.players))
	for _, p := range c.players {
		rank := uint8(guildRankMember)
		if p == c.master {
			rank = guildRankMaster
		}
		founders = append(founders, db.GuildCharRow{CharID: p.CharacterID, Rank: rank})
	}
	if err := db.CreateGuildGroup(g.ID, g.Name, c.master.guild.userID, g.Remark, g.Country, master, fee, founders); err != nil {
		log.Printf("GX Server World guild create failed: %v", err)
		c.sendAll(buildGuildCreateFailed(6, c.name), "0x0D01 guild create failed")
		return true
	}
	nextGuildID++
	guildsByID[g.ID] = g
	owner.char.Money -= fee
	owner.syncPlayer()
	owner.send(buildSetMoney(owner.char.Money), "0x0213 guild fee")
	for _, p := range c.players {
		snap := p.Snapshot()
		rank := uint8(guildRankMember)
		if p == c.master {
			rank = guildRankMaster
		}
		g.members[p.CharacterID] = &guildMember{CharID: p.CharacterID, Name: snap.Name, Job: snap.Class, Level: snap.Level,
			Rank: rank, player: p}
		guildOfChar[p.CharacterID] = g
		p.guild.guild, p.guild.applied = g, nil
	}
	for _, p := range c.players {
		p.Send(buildGuildCreateResult(1, g.ID, g.members[p.CharacterID].Rank, g.Name, g.Remark), "0x0D21 guild created")
		for _, f := range g.memberListPackets() {
			p.Send(f, "0x0D12 guild members")
		}
	}
	w := packet.New(packet.GuildNew)
	writeGuildInfo(w, g)
	guildSendCountry(g.Country, w.MustFinish(), "0x0D30 new guild")
	log.Printf("GX Server World guild created id=%d name=%q master=%q", g.ID, g.Name, g.Master)
	return true
}
```

</details>

<details><summary>DecodeGuildInfo — internal/proto/guild.go</summary>

[Source definition](../../internal/proto/guild.go#L92)

```go
func DecodeGuildInfo(p []byte) (GuildInfo, error) {
	r := NewReader(p)
	v := readGuildInfo(r)
	return v, r.Err()
}
```

```go
type GuildInfo struct {
	ID                   uint32
	Name, Master, Remark string
	Rank                 uint8
	Points               uint32
}
```

</details>

## Opcode 0d31

**`0x0D31` — GuildBankPlaced**

Declarations: [packet.GuildBankPlaced](../../server/internal/packet/packet.go#L180), [proto.GuildBankPlaced](../../internal/proto/opcodes.go#L167).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildGuildBankSlot (server/internal/world/guildbank.go)](../../server/internal/world/guildbank.go#L75)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L198), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildBankSlot — server/internal/world/guildbank.go</summary>

[Source definition](../../server/internal/world/guildbank.go#L75)

```go
func buildGuildBankSlot(op packet.Opcode, slot uint8, it *db.Item, charID uint32) []byte {
	w := packet.New(op)
	w.U8(slot)
	if it != nil {
		w.U8(it.Type)
		w.U8(it.TypeID)
		w.U16(it.Quality)
		w.Bytes(it.Gems[:])
		w.Count(it.Count)
		w.FixedString(it.CraftName, 21)
	} else {
		w.Bytes(make([]byte, 10))
		w.Count(0)
		w.FixedString("", 21)
	}
	w.U32(charID)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeGuildBankEvent — internal/proto/guild_bank.go</summary>

[Source definition](../../internal/proto/guild_bank.go#L40)

```go
func DecodeGuildBankEvent(p []byte) (GuildBankEvent, error) {
	r := NewReader(p)
	event := GuildBankEvent{Item: readGuildBankItem(r), Actor: r.U32()}
	if err := r.Err(); err != nil {
		return event, err
	}
	if int(event.Item.Slot) >= GuildBankSlots {
		return event, fmt.Errorf("guild bank: invalid slot %d", event.Item.Slot)
	}
	return event, nil
}
```

```go
type GuildBankEvent struct {
	Item  InventoryItem
	Actor uint32
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

## Opcode 0d32

**`0x0D32` — GuildBankTaken**

Declarations: [packet.GuildBankTaken](../../server/internal/packet/packet.go#L181), [proto.GuildBankTaken](../../internal/proto/opcodes.go#L168).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildGuildBankSlot (server/internal/world/guildbank.go)](../../server/internal/world/guildbank.go#L75)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L198), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildBankSlot — server/internal/world/guildbank.go</summary>

[Source definition](../../server/internal/world/guildbank.go#L75)

```go
func buildGuildBankSlot(op packet.Opcode, slot uint8, it *db.Item, charID uint32) []byte {
	w := packet.New(op)
	w.U8(slot)
	if it != nil {
		w.U8(it.Type)
		w.U8(it.TypeID)
		w.U16(it.Quality)
		w.Bytes(it.Gems[:])
		w.Count(it.Count)
		w.FixedString(it.CraftName, 21)
	} else {
		w.Bytes(make([]byte, 10))
		w.Count(0)
		w.FixedString("", 21)
	}
	w.U32(charID)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeGuildBankEvent — internal/proto/guild_bank.go</summary>

[Source definition](../../internal/proto/guild_bank.go#L40)

```go
func DecodeGuildBankEvent(p []byte) (GuildBankEvent, error) {
	r := NewReader(p)
	event := GuildBankEvent{Item: readGuildBankItem(r), Actor: r.U32()}
	if err := r.Err(); err != nil {
		return event, err
	}
	if int(event.Item.Slot) >= GuildBankSlots {
		return event, fmt.Errorf("guild bank: invalid slot %d", event.Item.Slot)
	}
	return event, nil
}
```

```go
type GuildBankEvent struct {
	Item  InventoryItem
	Actor uint32
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

## Opcode 0d33

**`0x0D33` — GuildHouseKeepEtin**

Declarations: [packet.GuildHouseKeepEtin](../../server/internal/packet/packet.go#L187), [proto.GuildHouseKeepEtin](../../internal/proto/opcodes.go#L174).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [houseStatusPacket (server/internal/world/guildhouse.go)](../../server/internal/world/guildhouse.go#L603)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L117), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>houseStatusPacket — server/internal/world/guildhouse.go</summary>

[Source definition](../../server/internal/world/guildhouse.go#L603)

```go
func (g *guild) houseStatusPacket() []byte {
	if g.BuyHouse == 0 {
		return nil
	}
	v := max(g.KeepEtin, 1) & 0x7fffffff
	if g.gmHouse {
		v |= 1 << 31
	}
	off := func(factor float64) uint8 { return uint8(min(100, max(0, houseRound((1-factor)*100)))) }
	w := packet.New(packet.GuildHouseKeepEtin)
	w.U32(v)
	for shop := uint8(1); shop <= 4; shop++ {
		level := min(g.npcLevel(uint16(shop)<<8|1), 3)
		w.U8(off(guildHouse.ShopFactor[shop][level]))
	}
	w.U8(off(guildHouse.GateRate[min(g.npcLevel(guildNpcGate), 3)]))
	smith := min(g.npcLevel(guildNpcSmith), 5)
	w.U8(uint8(min(100, max(0, guildHouse.MixPercent[smith]))))
	w.U8(off(guildHouse.MixRate[smith]))
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

## Opcode 0d34

**`0x0D34` — GuildHouseBuy**

Declarations: [packet.GuildHouseBuy](../../server/internal/packet/packet.go#L188), [proto.GuildHouseBuy](../../internal/proto/opcodes.go#L175).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L382)

**S→C serializers:** [buildGuildHouseBuy (server/internal/world/guildhouse.go)](../../server/internal/world/guildhouse.go#L400)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L128), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildGuildHouseBuy — server/internal/world/guildhouse.go</summary>

[Source definition](../../server/internal/world/guildhouse.go#L400)

```go
func buildGuildHouseBuy(result uint8, money uint32) []byte {
	w := packet.New(packet.GuildHouseBuy)
	w.U8(result)
	w.U32(money)
	return w.MustFinish()
}
```

</details>

<details><summary>GuildHouseBuyPacket — internal/proto/guild_house.go</summary>

[Source definition](../../internal/proto/guild_house.go#L74)

```go
func GuildHouseBuyPacket() []byte { w := NewWriter(GuildHouseBuy); w.U32(0); return w.Bytes() }
```

</details>

<details><summary>DecodeGuildHouseBuy — internal/proto/guild_house.go</summary>

[Source definition](../../internal/proto/guild_house.go#L46)

```go
func DecodeGuildHouseBuy(p []byte) (GuildHouseBuyResult, error) {
	r := NewReader(p)
	v := GuildHouseBuyResult{Result: r.U8(), Money: r.U32()}
	return v, r.Err()
}
```

```go
type GuildHouseBuyResult struct {
	Result uint8
	Money  uint32
}
```

</details>

## Opcode 0d35

**`0x0D35` — GuildDonate**

Declarations: [packet.GuildDonate](../../server/internal/packet/packet.go#L189), [proto.GuildDonate](../../internal/proto/opcodes.go#L176).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L172), [handleGuild (server/internal/world/guild.go)](../../server/internal/world/guild.go#L393)

**S→C serializers:** [guildDonate (server/internal/world/guildhouse.go)](../../server/internal/world/guildhouse.go#L464)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L137), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>guildDonate — server/internal/world/guildhouse.go</summary>

[Source definition](../../server/internal/world/guildhouse.go#L464)

```go
func (s *Session) guildDonate(g *guild) {
	next := *g
	var total uint32
	committed := s.inventoryOperationWithCommit(func() {
		var taken [][2]uint8
		for bag := uint8(1); bag <= 5; bag++ {
			for slot := uint8(0); slot < 24; slot++ {
				i := findItem(s.items, bag, slot)
				if i < 0 {
					continue
				}
				value, ok := guildEtinValue(&s.items[i])
				if !ok {
					continue
				}
				total += value
				s.items = append(s.items[:i], s.items[i+1:]...)
				taken = append(taken, [2]uint8{bag, slot})
			}
		}
		w := packet.New(packet.GuildDonate)
		w.U8(uint8(len(taken)))
		for _, t := range taken {
			w.U8(t[0])
			w.U8(t[1])
		}
		s.send(w.MustFinish(), "0x0D35 guild etin donated")
	}, func(change db.CharacterChange) error {
		if total > 0 {
			next.Etin += total
			next.EtinReturn++
		}
		return db.CommitGuildHouse(next.row(), nil, &change)
	})
	if committed && total > 0 {
		g.Etin, g.EtinReturn = next.Etin, next.EtinReturn
		g.send(guildID32(packet.GuildEtin, g.Etin), "0x0D23 guild etin", nil)
	}
}
```

</details>

<details><summary>GuildDonatePacket — internal/proto/guild_house.go</summary>

[Source definition](../../internal/proto/guild_house.go#L75)

```go
func GuildDonatePacket() []byte   { w := NewWriter(GuildDonate); w.U32(0); return w.Bytes() }
```

</details>

<details><summary>DecodeGuildDonation — internal/proto/guild_house.go</summary>

[Source definition](../../internal/proto/guild_house.go#L53)

```go
func DecodeGuildDonation(p []byte) ([][2]uint8, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > r.Remaining()/2 {
		return nil, ErrShort
	}
	out := make([][2]uint8, n)
	for i := range out {
		out[i] = [2]uint8{r.U8(), r.U8()}
	}
	return out, r.Err()
}
```

</details>

## Opcode 0d37

**`0x0D37` — GuildRankingList**

Declarations: [packet.GuildRankingList](../../server/internal/packet/packet.go#L177), [proto.GuildRankingList](../../internal/proto/opcodes.go#L164).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [updateGuildRanking (server/internal/world/grb.go)](../../server/internal/world/grb.go#L348)

**Client receive reducers:** [applyGuild (internal/game/guild.go)](../../internal/game/guild.go#L173), [Apply (internal/game/state.go)](../../internal/game/state.go#L635)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>updateGuildRanking — server/internal/world/grb.go</summary>

[Source definition](../../server/internal/world/grb.go#L348)

```go
func updateGuildRanking(log *logx.Logger) {
	guildMu.Lock()
	type entry struct {
		g     *guild
		value uint32
	}
	var ranked [2][]entry
	ids := make([]uint32, 0, len(guildsByID))
	for id := range guildsByID {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var rows []db.GuildRow
	seen := [2]map[uint32]bool{{}, {}}
	for _, id := range ids {
		g := guildsByID[id]
		value := g.grbPoints
		g.grbPoints = 0
		g.Points = value
		if g.gmHouse {
			g.forceGMHouse() // test guilds stay out of the standings
			continue
		}
		c := int(g.Country) & 1
		if value == 0 {
			g.Rank = grbUnranked
		} else if !seen[c][value] {
			seen[c][value] = true
			ranked[c] = append(ranked[c], entry{g, value})
		}
	}
	lists := [2][]byte{}
	for c := range ranked {
		sort.SliceStable(ranked[c], func(i, j int) bool { return ranked[c][i].value > ranked[c][j].value })
		w := packet.New(packet.GuildRankingList)
		count := 0
		for i, e := range ranked[c] {
			if i < grbRankedGuilds {
				e.g.Rank = uint8(i + 1)
				count++
			} else {
				e.g.Rank = grbUnranked
			}
		}
		w.U8(uint8(count))
		for _, e := range ranked[c][:count] {
			w.U32(e.g.ID)
			w.U32(e.value)
			w.U8(e.g.Rank)
		}
		lists[c] = w.MustFinish()
	}
	for _, id := range ids {
		g := guildsByID[id]
		g.settleGuildHouse() // every guild with a house pays its upkeep, ranked or not
		rows = append(rows, db.GuildRow{ID: g.ID, Points: g.Points, Rank: g.Rank, Etin: g.Etin, Remark: g.Remark,
			BuyHouse: g.BuyHouse, UseHouse: g.UseHouse, EtinReturn: g.EtinReturn, KeepEtin: g.KeepEtin})
	}
	grbState.Lock()
	grbState.maxPoints, grbState.maxPointsID = [2]uint32{}, [2]uint32{}
	grbState.Unlock()
	guildMu.Unlock()
	for _, r := range rows {
		if err := db.SaveGuildRanking(r); err != nil {
			log.Printf("GX Server World guild ranking save failed guild=%d: %v", r.ID, err)
		}
	}
	for c, frame := range lists {
		guildSendCountry(uint8(c), frame, "0x0D37 guild ranking")
	}
	log.Printf("GX Server World guild ranking updated guilds=%d", len(rows))
}
```

</details>

<details><summary>DecodeGuildRankingList — internal/proto/guild_ranking.go</summary>

[Source definition](../../internal/proto/guild_ranking.go#L39)

```go
func DecodeGuildRankingList(p []byte) ([]GuildRankingEntry, error) {
	r := NewReader(p)
	count := int(r.U8())
	if r.Err() != nil {
		return nil, r.Err()
	}
	// Game.exe's handler has space for 40 rows; Rapture sends at most 30.
	if count > 40 {
		return nil, fmt.Errorf("guild ranking: too many rows %d", count)
	}
	if r.Remaining() != count*9 {
		return nil, fmt.Errorf("guild ranking: incorrect row length")
	}
	out := make([]GuildRankingEntry, 0, count)
	seen := make(map[uint32]bool, count)
	for i := 0; i < count; i++ {
		v := GuildRankingEntry{Guild: r.U32(), Points: r.U32(), Rank: r.U8()}
		if v.Guild == 0 || seen[v.Guild] {
			return nil, fmt.Errorf("guild ranking: duplicate or zero guild")
		}
		seen[v.Guild] = true
		out = append(out, v)
	}
	return out, r.Err()
}
```

```go
type GuildRankingEntry struct {
	Guild, Points uint32
	Rank          uint8
}
```

</details>
