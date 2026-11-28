# Protocol: Party, raid and group resources

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 0b01

**`0x0B01` — PartyRoster**

Declarations: [packet.PartyRoster](../../server/internal/packet/packet.go#L42), [proto.PartyRoster](../../internal/proto/opcodes.go#L26).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [enterParty (server/internal/world/party.go)](../../server/internal/world/party.go#L234)

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L62), [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>enterParty — server/internal/world/party.go</summary>

[Source definition](../../server/internal/world/party.go#L234)

```go
func enterParty(pt *party, p *ConnectedPlayer) bool {
	if pt.isRaid() {
		return enterRaid(pt, p)
	}
	if len(pt.members) >= partyMax {
		return false
	}
	if len(pt.members) > 0 {
		w := packet.New(packet.PartyJoined)
		writePartyMember(w, p)
		pt.send(w.MustFinish(), "0x0B04 party member joined")
	}
	w := packet.New(packet.PartyRoster)
	w.U8(uint8(pt.leader))
	w.U8(uint8(len(pt.members)))
	for _, m := range pt.members {
		writePartyMember(w, m)
	}
	p.Send(w.MustFinish(), "0x0B01 party roster")
	pt.members = append(pt.members, p)
	pt.lootJoin(p.CharacterID, time.Now())
	p.party, p.inviterID = pt, 0
	pt.refreshLevel()
	partyBadge(p)
	return true
}
```

</details>

<details><summary>DecodePartyList — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L46)

```go
func DecodePartyList(p []byte) (list PartyList, err error) {
	r := NewReader(p)
	list.LeaderIndex = r.U8()
	n := int(r.U8())
	// Native scratch contains eight records; the normal party has at most seven.
	if n > 8 {
		return list, fmt.Errorf("party: invalid member count %d", n)
	}
	if n > r.Remaining()/67 {
		return list, ErrShort
	}
	list.Members = make([]PartyMember, n)
	for i := range list.Members {
		list.Members[i] = readPartyMember(r)
	}
	return list, r.Err()
}
```

```go
type PartyMember struct {
	ID                              uint32
	Name                            string
	Level                           uint16
	Class                           uint8
	HP, MaxHP, MP, MaxMP, SP, MaxSP uint32
	Instance                        uint16
	X, Y, Z                         float32
	Effects                         []PartyEffect
}
```

```go
type PartyList struct {
	LeaderIndex uint8
	Members     []PartyMember
}
```

</details>

## Opcode 0b02

**`0x0B02` — PartyInvite**

Declarations: [packet.PartyInvite](../../server/internal/packet/packet.go#L43), [proto.PartyInvite](../../internal/proto/opcodes.go#L27).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L126)

**S→C serializers:** [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L101), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

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

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0b03

**`0x0B03` — PartyAnswer**

Declarations: [packet.PartyAnswer](../../server/internal/packet/packet.go#L44), [proto.PartyAnswer](../../internal/proto/opcodes.go#L28).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L130)

**S→C serializers:** [partyError (server/internal/world/party.go)](../../server/internal/world/party.go#L207)

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L126), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>partyError — server/internal/world/party.go</summary>

[Source definition](../../server/internal/world/party.go#L207)

```go
func partyError(to *ConnectedPlayer, code uint8, id uint32) {
	w := packet.New(packet.PartyAnswer)
	w.U8(code)
	w.U32(id)
	to.Send(w.MustFinish(), "0x0B03 party error")
}
```

</details>

<details><summary>PartyAnswerPacket — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L85)

```go
func PartyAnswerPacket(accept bool, characterID uint32) []byte {
	w := NewWriter(PartyAnswer)
	if accept {
		w.U8(0)
	} else {
		w.U8(1)
	}
	w.U32(characterID)
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

## Opcode 0b04

**`0x0B04` — PartyJoined**

Declarations: [packet.PartyJoined](../../server/internal/packet/packet.go#L45), [proto.PartyJoined](../../internal/proto/opcodes.go#L29).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [enterParty (server/internal/world/party.go)](../../server/internal/world/party.go#L234)

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L88), [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>enterParty — server/internal/world/party.go</summary>

[Source definition](../../server/internal/world/party.go#L234)

```go
func enterParty(pt *party, p *ConnectedPlayer) bool {
	if pt.isRaid() {
		return enterRaid(pt, p)
	}
	if len(pt.members) >= partyMax {
		return false
	}
	if len(pt.members) > 0 {
		w := packet.New(packet.PartyJoined)
		writePartyMember(w, p)
		pt.send(w.MustFinish(), "0x0B04 party member joined")
	}
	w := packet.New(packet.PartyRoster)
	w.U8(uint8(pt.leader))
	w.U8(uint8(len(pt.members)))
	for _, m := range pt.members {
		writePartyMember(w, m)
	}
	p.Send(w.MustFinish(), "0x0B01 party roster")
	pt.members = append(pt.members, p)
	pt.lootJoin(p.CharacterID, time.Now())
	p.party, p.inviterID = pt, 0
	pt.refreshLevel()
	partyBadge(p)
	return true
}
```

</details>

<details><summary>DecodePartyJoined — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L63)

```go
func DecodePartyJoined(p []byte) (PartyMember, error) {
	r := NewReader(p)
	m := readPartyMember(r)
	return m, r.Err()
}
```

```go
type PartyMember struct {
	ID                              uint32
	Name                            string
	Level                           uint16
	Class                           uint8
	HP, MaxHP, MP, MaxMP, SP, MaxSP uint32
	Instance                        uint16
	X, Y, Z                         float32
	Effects                         []PartyEffect
}
```

</details>

## Opcode 0b05

**`0x0B05` — PartyLeft**

Declarations: [packet.PartyLeft](../../server/internal/packet/packet.go#L46), [proto.PartyLeft](../../internal/proto/opcodes.go#L30).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L132)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L133), [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>PartyLeavePacket — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L95)

```go
func PartyLeavePacket() []byte           { return NewWriter(PartyLeft).Bytes() }
```

</details>

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0b06

**`0x0B06` — PartyExpelled**

Declarations: [packet.PartyExpelled](../../server/internal/packet/packet.go#L47), [proto.PartyExpelled](../../internal/proto/opcodes.go#L31).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L134)

**S→C serializers:** [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L133), [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

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

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0b07

**`0x0B07` — PartyLeader**

Declarations: [packet.PartyLeader](../../server/internal/packet/packet.go#L48), [proto.PartyLeader](../../internal/proto/opcodes.go#L32).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L136)

**S→C serializers:** [partyIDPacket (server/internal/world/party.go)](../../server/internal/world/party.go#L201), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L148), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>partyIDPacket — server/internal/world/party.go</summary>

[Source definition](../../server/internal/world/party.go#L201)

```go
func partyIDPacket(op packet.Opcode, id uint32) []byte {
	w := packet.New(op)
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

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0b09

**`0x0B09` — PartyInviteName**

Declarations: [packet.PartyInviteName](../../server/internal/packet/packet.go#L49), [proto.PartyInviteName](../../internal/proto/opcodes.go#L33).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L126)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L112), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>PartyInviteNamePacket — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L98)

```go
func PartyInviteNamePacket(name string) []byte {
	w := NewWriter(PartyInviteName)
	w.Fixed(name, 21, true)
	return w.Bytes()
}
```

</details>

<details><summary>DecodePartyInviteName — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L73)

```go
func DecodePartyInviteName(p []byte) (string, error) {
	r := NewReader(p)
	s := r.Fixed(21)
	return s, r.Err()
}
```

</details>

## Opcode 0b0b

**`0x0B0B` — RaidRoster**

Declarations: [packet.RaidRoster](../../server/internal/packet/packet.go#L52), [proto.RaidRoster](../../internal/proto/opcodes.go#L36).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [sendRaidRoster (server/internal/world/raid.go)](../../server/internal/world/raid.go#L94)

**Client receive reducers:** [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L21), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>sendRaidRoster — server/internal/world/raid.go</summary>

[Source definition](../../server/internal/world/raid.go#L94)

```go
func (pt *party) sendRaidRoster(to *ConnectedPlayer) {
	for start := 0; start < len(pt.members); start += raidRosterChunk {
		end := min(start+raidRosterChunk, len(pt.members))
		w := packet.New(packet.RaidRoster)
		w.Bool(end < len(pt.members))
		w.U8(uint8(pt.slots[pt.leader]))
		w.U8(uint8(pt.slots[pt.subLeader]))
		w.U16(uint16(pt.itemDivision))
		w.U8(pt.kind)
		w.Bool(pt.autoJoin)
		w.U8(uint8(end - start))
		for i := start; i < end; i++ {
			w.U16(pt.slots[i])
			writePartyMember(w, pt.members[i])
		}
		to.Send(w.MustFinish(), "0x0B0B raid roster")
	}
}
```

</details>

<details><summary>DecodeRaidRoster — internal/proto/raid.go</summary>

[Source definition](../../internal/proto/raid.go#L27)

```go
func DecodeRaidRoster(p []byte) (RaidRosterChunk, error) {
	r := NewReader(p)
	var chunk RaidRosterChunk
	chunk.More = r.U8() != 0
	chunk.LeaderSlot, chunk.SubLeaderSlot = r.U8(), r.U8()
	chunk.Division, chunk.Kind, chunk.AutoJoin = r.U16(), r.U8(), r.U8() != 0
	n := int(r.U8())
	if err := r.Err(); err != nil {
		return chunk, err
	}
	if n > 6 {
		return chunk, fmt.Errorf("raid: invalid chunk size %d", n)
	}
	if n > r.Remaining()/69 {
		return chunk, ErrShort
	}
	chunk.Members = make([]RaidMember, n)
	for i := range chunk.Members {
		chunk.Members[i] = RaidMember{r.U16(), readPartyMember(r)}
	}
	return chunk, r.Err()
}
```

```go
type RaidMember struct {
	Slot   uint16
	Member PartyMember
}
```

```go
type RaidRosterChunk struct {
	More                      bool
	LeaderSlot, SubLeaderSlot uint8
	Division                  uint16
	Kind                      uint8
	AutoJoin                  bool
	Members                   []RaidMember
}
```

</details>

## Opcode 0b0c

**`0x0B0C` — RaidJoined**

Declarations: [packet.RaidJoined](../../server/internal/packet/packet.go#L53), [proto.RaidJoined](../../internal/proto/opcodes.go#L37).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [enterRaid (server/internal/world/raid.go)](../../server/internal/world/raid.go#L65)

**Client receive reducers:** [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L66), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>enterRaid — server/internal/world/raid.go</summary>

[Source definition](../../server/internal/world/raid.go#L65)

```go
func enterRaid(pt *party, p *ConnectedPlayer) bool {
	if len(pt.members) >= raidMax {
		return false
	}
	slot := pt.emptyRaidSlot()
	if slot < 0 {
		return false
	}
	pt.members = append(pt.members, p)
	pt.lootJoin(p.CharacterID, time.Now())
	pt.slots = append(pt.slots, uint16(slot))
	if len(pt.members) == 2 && pt.subLeader == pt.leader {
		pt.subLeader = 1 // a raid started by an invitation: the first to join assists
	}
	p.party, p.inviterID = pt, 0
	if len(pt.members) > 1 {
		w := packet.New(packet.RaidJoined)
		w.U16(uint16(slot))
		writePartyMember(w, p)
		pt.send(w.MustFinish(), "0x0B0C raid member joined")
	}
	pt.sendRaidRoster(p)
	pt.refreshLevel()
	partyBadge(p)
	return true
}
```

</details>

<details><summary>DecodeRaidJoined — internal/proto/raid.go</summary>

[Source definition](../../internal/proto/raid.go#L50)

```go
func DecodeRaidJoined(p []byte) (RaidMember, error) {
	r := NewReader(p)
	m := RaidMember{r.U16(), readPartyMember(r)}
	return m, r.Err()
}
```

```go
type RaidMember struct {
	Slot   uint16
	Member PartyMember
}
```

</details>

## Opcode 0b0d

**`0x0B0D` — PartyLeaveAlt**

Declarations: [packet.PartyLeaveAlt](../../server/internal/packet/packet.go#L50), [proto.PartyLeaveAlt](../../internal/proto/opcodes.go#L34).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L132)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L241), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0b0e

**`0x0B0E` — RaidConvert**

Declarations: [packet.RaidConvert](../../server/internal/packet/packet.go#L54), [proto.RaidConvert](../../internal/proto/opcodes.go#L38).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L138), [handleRaidSettings (server/internal/world/raid.go)](../../server/internal/world/raid.go#L184)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L92), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>RaidConvertPacket — internal/proto/raid.go</summary>

[Source definition](../../internal/proto/raid.go#L56)

```go
func RaidConvertPacket(kind uint8, autoJoin bool, division uint32) []byte {
	w := NewWriter(RaidConvert)
	w.U8(kind)
	if autoJoin {
		w.U8(1)
	} else {
		w.U8(0)
	}
	w.U32(division)
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

## Opcode 0b0f

**`0x0B0F` — RaidDivision**

Declarations: [packet.RaidDivision](../../server/internal/packet/packet.go#L55), [proto.RaidDivision](../../internal/proto/opcodes.go#L39).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L138), [handleRaidSettings (server/internal/world/raid.go)](../../server/internal/world/raid.go#L216)

**S→C serializers:** [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L109), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

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

## Opcode 0b10

**`0x0B10` — RaidAutoJoin**

Declarations: [packet.RaidAutoJoin](../../server/internal/packet/packet.go#L56), [proto.RaidAutoJoin](../../internal/proto/opcodes.go#L40).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L138), [handleRaidSettings (server/internal/world/raid.go)](../../server/internal/world/raid.go#L224)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L109), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>RaidAutoJoinPacket — internal/proto/raid.go</summary>

[Source definition](../../internal/proto/raid.go#L68)

```go
func RaidAutoJoinPacket(enabled bool) []byte {
	w := NewWriter(RaidAutoJoin)
	if enabled {
		w.U8(1)
	} else {
		w.U8(0)
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

## Opcode 0b11

**`0x0B11` — RaidAutoJoinName**

Declarations: [packet.RaidAutoJoinName](../../server/internal/packet/packet.go#L57), [proto.RaidAutoJoinName](../../internal/proto/opcodes.go#L41).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L144)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

The shape is decoded inline, delegated, or only declared. Use the linked adapter/reducer and the codec appendix below; there is no inferred fixed layout here.

## Opcode 0b12

**`0x0B12` — RaidSwap**

Declarations: [packet.RaidSwap](../../server/internal/packet/packet.go#L58), [proto.RaidSwap](../../internal/proto/opcodes.go#L42).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L146), [handleRaidRoles (server/internal/world/raid.go)](../../server/internal/world/raid.go#L414)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L138), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>RaidSwapPacket — internal/proto/raid.go</summary>

[Source definition](../../internal/proto/raid.go#L77)

```go
func RaidSwapPacket(first, second uint32) []byte {
	w := NewWriter(RaidSwap)
	w.U32(first)
	w.U32(second)
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

## Opcode 0b13

**`0x0B13` — RaidSubLeader**

Declarations: [packet.RaidSubLeader](../../server/internal/packet/packet.go#L59), [proto.RaidSubLeader](../../internal/proto/opcodes.go#L43).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L146), [handleRaidRoles (server/internal/world/raid.go)](../../server/internal/world/raid.go#L386)

**S→C serializers:** [partyIDPacket (server/internal/world/party.go)](../../server/internal/world/party.go#L201), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L170), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>partyIDPacket — server/internal/world/party.go</summary>

[Source definition](../../server/internal/world/party.go#L201)

```go
func partyIDPacket(op packet.Opcode, id uint32) []byte {
	w := packet.New(op)
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

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0b14

**`0x0B14` — RaidInvite**

Declarations: [packet.RaidInvite](../../server/internal/packet/packet.go#L60), [proto.RaidInvite](../../internal/proto/opcodes.go#L44).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L140)

**S→C serializers:** [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L194), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

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

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

<details><summary>DecodePartyInviteName — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L73)

```go
func DecodePartyInviteName(p []byte) (string, error) {
	r := NewReader(p)
	s := r.Fixed(21)
	return s, r.Err()
}
```

</details>

## Opcode 0b15

**`0x0B15` — RaidAnswer**

Declarations: [packet.RaidAnswer](../../server/internal/packet/packet.go#L61), [proto.RaidAnswer](../../internal/proto/opcodes.go#L45).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L142)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L220), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>RaidAnswerPacket — internal/proto/raid.go</summary>

[Source definition](../../internal/proto/raid.go#L88)

```go
func RaidAnswerPacket(accept bool, id uint32) []byte {
	w := NewWriter(RaidAnswer)
	if accept {
		w.U8(0)
	} else {
		w.U8(1)
	}
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

## Opcode 0b16

**`0x0B16` — RaidDisband**

Declarations: [packet.RaidDisband](../../server/internal/packet/packet.go#L62), [proto.RaidDisband](../../internal/proto/opcodes.go#L46).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L138), [handleRaidSettings (server/internal/world/raid.go)](../../server/internal/world/raid.go#L231)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L237), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>RaidDisbandPacket — internal/proto/raid.go</summary>

[Source definition](../../internal/proto/raid.go#L87)

```go
func RaidDisbandPacket() []byte            { return NewWriter(RaidDisband).Bytes() }
```

</details>

## Opcode 0b17

**`0x0B17` — RaidLeader**

Declarations: [packet.RaidLeader](../../server/internal/packet/packet.go#L63), [proto.RaidLeader](../../internal/proto/opcodes.go#L47).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L146), [handleRaidRoles (server/internal/world/raid.go)](../../server/internal/world/raid.go#L393)

**S→C serializers:** [partyIDPacket (server/internal/world/party.go)](../../server/internal/world/party.go#L201), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L170), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>partyIDPacket — server/internal/world/party.go</summary>

[Source definition](../../server/internal/world/party.go#L201)

```go
func partyIDPacket(op packet.Opcode, id uint32) []byte {
	w := packet.New(op)
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

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0b18

**`0x0B18` — RaidInviteName**

Declarations: [packet.RaidInviteName](../../server/internal/packet/packet.go#L64), [proto.RaidInviteName](../../internal/proto/opcodes.go#L48).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L140)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L194), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

<details><summary>DecodePartyInviteName — internal/proto/party.go</summary>

[Source definition](../../internal/proto/party.go#L73)

```go
func DecodePartyInviteName(p []byte) (string, error) {
	r := NewReader(p)
	s := r.Fixed(21)
	return s, r.Err()
}
```

</details>

## Opcode 0b19

**`0x0B19` — RaidError**

Declarations: [packet.RaidError](../../server/internal/packet/packet.go#L65), [proto.RaidError](../../internal/proto/opcodes.go#L49).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [raidError (server/internal/world/raid.go)](../../server/internal/world/raid.go#L57)

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L228), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>raidError — server/internal/world/raid.go</summary>

[Source definition](../../server/internal/world/raid.go#L57)

```go
func raidError(to *ConnectedPlayer, reason, id uint32) {
	w := packet.New(packet.RaidError)
	w.U32(reason)
	w.U32(id)
	to.Send(w.MustFinish(), "0x0B19 raid error")
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

## Opcode 0b1a

**`0x0B1A` — RaidExpel**

Declarations: [packet.RaidExpel](../../server/internal/packet/packet.go#L66), [proto.RaidExpel](../../internal/proto/opcodes.go#L50).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L148)

**S→C serializers:** [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [questMapStateChanged (internal/game/quest_map.go)](../../internal/game/quest_map.go#L13), [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L241), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

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

<details><summary>DecodeU32 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L197)

```go
func DecodeU32(p []byte) (uint32, error) { r := NewReader(p); v := r.U32(); return v, r.Err() }
```

</details>

## Opcode 0c03

**`0x0C03` — PartyPools**

Declarations: [packet.PartyPools](../../server/internal/packet/packet.go#L68), [proto.PartyPools](../../internal/proto/opcodes.go#L52).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L173), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c04

**`0x0C04` — PartyEffectAdd**

Declarations: [packet.PartyEffectAdd](../../server/internal/packet/packet.go#L72), [proto.PartyEffectAdd](../../internal/proto/opcodes.go#L56).

Source notes: handlers 0x5929a0 / 0x5747b0

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L182), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c05

**`0x0C05` — PartyEffectRemove**

Declarations: [packet.PartyEffectRemove](../../server/internal/packet/packet.go#L73), [proto.PartyEffectRemove](../../internal/proto/opcodes.go#L57).

Source notes: handlers 0x592a00 / 0x573b50

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L182), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c07

**`0x0C07` — PartyMove**

Declarations: [packet.PartyMove](../../server/internal/packet/packet.go#L70), [proto.PartyMove](../../internal/proto/opcodes.go#L54).

Source notes: 0x592a90 -> 0x573670: id, map, x, y, z

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L195), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c09

**`0x0C09` — PartyLevel**

Declarations: [packet.PartyLevel](../../server/internal/packet/packet.go#L69), [proto.PartyLevel](../../internal/proto/opcodes.go#L53).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L204), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c0c

**`0x0C0C` — RaidPools**

Declarations: [packet.RaidPools](../../server/internal/packet/packet.go#L67), [proto.RaidPools](../../internal/proto/opcodes.go#L51).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyRaid (internal/game/raid.go)](../../internal/game/raid.go#L268), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c0e

**`0x0C0E` — RaidEffectAdd**

Declarations: [packet.RaidEffectAdd](../../server/internal/packet/packet.go#L74), [proto.RaidEffectAdd](../../internal/proto/opcodes.go#L58).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L182), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c10

**`0x0C10` — RaidEffectRemove**

Declarations: [packet.RaidEffectRemove](../../server/internal/packet/packet.go#L75), [proto.RaidEffectRemove](../../internal/proto/opcodes.go#L59).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L182), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c12

**`0x0C12` — RaidMove**

Declarations: [packet.RaidMove](../../server/internal/packet/packet.go#L71), [proto.RaidMove](../../internal/proto/opcodes.go#L55).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L195), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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

## Opcode 0c13

**`0x0C13` — RaidLevel**

Declarations: [packet.RaidLevel](../../server/internal/packet/packet.go#L347), [proto.RaidLevel](../../internal/proto/raid.go#L105).

Source notes: RaidLevel is the original member level notice, parallel to PartyLevel. RaidLevel updates a raid member's level (0x592b80).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyParty (internal/game/party.go)](../../internal/game/party.go#L204), [Apply (internal/game/state.go)](../../internal/game/state.go#L653)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

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
