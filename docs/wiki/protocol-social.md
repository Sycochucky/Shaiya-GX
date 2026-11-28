# Protocol: Relics, friends and PvP rewards

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode 2101

**`0x2101` — RelicList**

Declarations: [proto.RelicList](../../internal/proto/relic.go#L13).

Source notes: EP5.4 SHA256 8e5ef14a...bb3263 readers 0x594d20/0x594d80/0x594dd0 and 0x594e30: lists contain runtime IDs, not map/INI persistence keys. Owner states 0/1/2 are neutral/Light/Fury; portal states 100..103 are closed/Light/Fury/both. A list replaces the current map snapshot.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [relicListFor (server/internal/world/relics.go)](../../server/internal/world/relics.go#L551)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L384)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>relicListFor — server/internal/world/relics.go</summary>

[Source definition](../../server/internal/world/relics.go#L551)

```go
func relicListFor(mapID uint16) [][]byte {
	relics.Lock()
	defer relics.Unlock()
	type row struct {
		id    uint32
		owner uint8
		x, z  float32
	}
	var rows []row
	for _, r := range relics.byMap[mapID] {
		rows = append(rows, row{r.id, r.owner, r.pos[0], r.pos[2]})
	}
	for _, p := range relics.portals {
		if p.mapID == mapID {
			rows = append(rows, row{p.id, portalMarker(p.country), p.pos[0], p.pos[2]})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].id < rows[j].id })
	var out [][]byte
	for len(rows) > 0 {
		n := min(100, len(rows))
		w := packet.New(relicListOpcode)
		w.U8(uint8(n))
		for _, r := range rows[:n] {
			w.U32(r.id)
			w.U8(r.owner)
			w.F32(r.x)
			w.F32(r.z)
		}
		out = append(out, w.MustFinish())
		rows = rows[n:]
	}
	return out
}
```

</details>

<details><summary>DecodeRelicList — internal/proto/relic.go</summary>

[Source definition](../../internal/proto/relic.go#L25)

```go
func DecodeRelicList(p []byte) ([]Relic, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > 100 || r.Remaining() != n*13 {
		return nil, fmt.Errorf("invalid relic list length")
	}
	out := make([]Relic, 0, n)
	seen := make(map[uint32]bool, n)
	for range n {
		v := Relic{ID: r.U32(), Owner: r.U8(), X: r.F32(), Z: r.F32()}
		if seen[v.ID] || !validRelicOwner(v.Owner) || math.IsNaN(float64(v.X)) || math.IsNaN(float64(v.Z)) || math.IsInf(float64(v.X), 0) || math.IsInf(float64(v.Z), 0) {
			return nil, fmt.Errorf("invalid relic list record")
		}
		seen[v.ID] = true
		out = append(out, v)
	}
	if err := r.Err(); err != nil {
		return nil, err
	}
	if r.Remaining() != 0 {
		return nil, fmt.Errorf("relic list: %d trailing bytes", r.Remaining())
	}
	return out, nil
}
```

```go
type Relic struct {
	ID    uint32
	Owner uint8
	X, Z  float32
}
```

</details>

## Opcode 2102

**`0x2102` — RelicOwner**

Declarations: [proto.RelicOwner](../../internal/proto/relic.go#L14).

Source notes: EP5.4 SHA256 8e5ef14a...bb3263 readers 0x594d20/0x594d80/0x594dd0 and 0x594e30: lists contain runtime IDs, not map/INI persistence keys. Owner states 0/1/2 are neutral/Light/Fury; portal states 100..103 are closed/Light/Fury/both. A list replaces the current map snapshot.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildRelicOwner (server/internal/world/relics.go)](../../server/internal/world/relics.go#L77)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L390)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildRelicOwner — server/internal/world/relics.go</summary>

[Source definition](../../server/internal/world/relics.go#L77)

```go
func buildRelicOwner(id uint32, owner uint8) []byte {
	w := packet.New(relicOwnerOpcode)
	w.U32(id)
	w.U8(owner)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeRelicOwner — internal/proto/relic.go</summary>

[Source definition](../../internal/proto/relic.go#L55)

```go
func DecodeRelicOwner(p []byte) (RelicOwnerChange, error) {
	r := NewReader(p)
	v := RelicOwnerChange{ID: r.U32(), Owner: r.U8()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if r.Remaining() != 0 || !validRelicOwner(v.Owner) {
		return v, fmt.Errorf("relic owner: %d trailing bytes", r.Remaining())
	}
	return v, nil
}
```

```go
type RelicOwnerChange struct {
	ID    uint32
	Owner uint8
}
```

</details>

## Opcode 2103

**`0x2103` — RelicNotice**

Declarations: [proto.RelicNotice](../../internal/proto/relic.go#L15).

Source notes: EP5.4 SHA256 8e5ef14a...bb3263 readers 0x594d20/0x594d80/0x594dd0 and 0x594e30: lists contain runtime IDs, not map/INI persistence keys. Owner states 0/1/2 are neutral/Light/Fury; portal states 100..103 are closed/Light/Fury/both. A list replaces the current map snapshot.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [buildRelicNotice (server/internal/world/relics.go)](../../server/internal/world/relics.go#L83)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L414)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildRelicNotice — server/internal/world/relics.go</summary>

[Source definition](../../server/internal/world/relics.go#L83)

```go
func buildRelicNotice(mob uint16, old, next uint8) []byte {
	w := packet.New(relicNoticeOpcode)
	w.U16(mob)
	w.U8(old)
	w.U8(next)
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeRelicNotice — internal/proto/relic.go</summary>

[Source definition](../../internal/proto/relic.go#L72)

```go
func DecodeRelicNotice(p []byte) (RelicDestroyed, error) {
	r := NewReader(p)
	v := RelicDestroyed{MobID: r.U16(), OldOwner: r.U8(), NewOwner: r.U8()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if r.Remaining() != 0 || v.OldOwner > 2 || v.NewOwner > 2 {
		return v, fmt.Errorf("relic notice: %d trailing bytes", r.Remaining())
	}
	return v, nil
}
```

```go
type RelicDestroyed struct {
	MobID              uint16
	OldOwner, NewOwner uint8
}
```

</details>

## Opcode 2104

**`0x2104` — PortalNotice**

Declarations: [proto.PortalNotice](../../internal/proto/relic.go#L16).

Source notes: EP5.4 SHA256 8e5ef14a...bb3263 readers 0x594d20/0x594d80/0x594dd0 and 0x594e30: lists contain runtime IDs, not map/INI persistence keys. Owner states 0/1/2 are neutral/Light/Fury; portal states 100..103 are closed/Light/Fury/both. A list replaces the current map snapshot.

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [changeRelicPortal (server/internal/world/relic_portals.go)](../../server/internal/world/relic_portals.go#L87)

**Client receive reducers:** [Apply (internal/game/state.go)](../../internal/game/state.go#L401)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>changeRelicPortal — server/internal/world/relic_portals.go</summary>

[Source definition](../../server/internal/world/relic_portals.go#L87)

```go
func changeRelicPortal(p *relicPortal, open [2]bool, out *[]eventPublication, log logPrinter) bool {
	country := int32(3)
	switch {
	case open[0] && open[1]:
		country = 0
	case open[0]:
		country = 1
	case open[1]:
		country = 2
	}
	if country == p.country {
		p.open = open
		return false
	}
	if err := db.SaveKeeps([]db.Keep{{ID: p.key, Country: uint8(country)}}); err != nil {
		if log != nil {
			log.Printf("GX Server World portal keep save failed: %v", err)
		}
		return false
	}
	old := p.country
	p.country, p.open = country, open
	setEventPortal(p.mapID, p.portalID, country)
	*out = append(*out, eventPublication{mapID: p.mapID, frame: buildRelicOwner(p.id, portalMarker(country))})
	w := packet.New(0x2104)
	w.U16(uint16(p.portalID))
	w.U8(portalMarker(old))
	w.U8(portalMarker(country))
	*out = append(*out, eventPublication{global: true, frame: w.MustFinish()})
	return true
}
```

</details>

<details><summary>DecodePortalNotice — internal/proto/relic.go</summary>

[Source definition](../../internal/proto/relic.go#L91)

```go
func DecodePortalNotice(p []byte) (PortalOwnerNotice, error) {
	r := NewReader(p)
	v := PortalOwnerNotice{r.U16(), r.U8(), r.U8()}
	if err := r.Err(); err != nil {
		return v, err
	}
	if r.Remaining() != 0 || v.OldOwner < 100 || v.OldOwner > 103 || v.NewOwner < 100 || v.NewOwner > 103 {
		return v, fmt.Errorf("invalid portal notice")
	}
	return v, nil
}
```

```go
type PortalOwnerNotice struct {
	PortalID           uint16
	OldOwner, NewOwner uint8
}
```

</details>

## Opcode 2201

**`0x2201` — FriendList**

Declarations: [proto.FriendList](../../internal/proto/social.go#L6).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L19), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeSocialList — internal/proto/social.go</summary>

[Source definition](../../internal/proto/social.go#L34)

```go
func DecodeSocialList(p []byte) ([]SocialContact, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > 50 {
		return nil, fmt.Errorf("social list count %d", n)
	}
	rows := make([]SocialContact, n)
	for i := range rows {
		v := &rows[i]
		v.ID = r.U32()
		if Op(p) == FriendList {
			v.Class = r.U8()
			v.Online = r.U8() != 0
		}
		v.Name = r.Fixed(21)
		v.Memo = r.Text()
		if len(v.Memo) > 50 {
			return nil, fmt.Errorf("social memo too long")
		}
	}
	if err := r.Err(); err != nil {
		return nil, err
	}
	if r.Remaining() != 0 {
		return nil, fmt.Errorf("social list trailing data")
	}
	return rows, nil
}
```

```go
type SocialContact struct {
	ID         uint32
	Class      uint8
	Online     bool
	Name, Memo string
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

## Opcode 2202

**`0x2202` — FriendInvite**

Declarations: [proto.FriendInvite](../../internal/proto/social.go#L7).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L38), [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L113)

**S→C serializers:** [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L90)

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L39), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleSocial — server/internal/world/social.go</summary>

[Source definition](../../server/internal/world/social.go#L90)

```go
func (s *Session) handleSocial(op packet.Opcode, p []byte) {
	if s.char == nil {
		return
	}
	result := func(op packet.Opcode, code uint8) {
		w := packet.New(op)
		w.U8(code)
		s.send(w.MustFinish(), "social result")
	}
	id := func() uint32 {
		if len(p) < 6 {
			return 0
		}
		v, _ := packet.U32(p[2:])
		return v
	}
	name := func() string {
		if len(p) < 23 {
			return ""
		}
		return strings.TrimSpace(strings.TrimRight(string(p[2:23]), "\x00"))
	}
	switch op {
	case 0x2202:
		target := FindPlayerByName(name())
		if target == nil || target.CharacterID == s.char.ID || target.Faction() != s.faction || target.MapID() != s.char.MapID || db.SocialBlocked(target.CharacterID, s.char.ID) {
			result(0x2203, 2)
			return
		}
		socialState.Lock()
		previous, busy := socialState.invites[target.CharacterID]
		busy = busy && time.Now().Before(previous.until)
		if !busy {
			socialState.invites[target.CharacterID] = socialInvite{s.char.ID, time.Now().Add(30 * time.Second)}
		}
		socialState.Unlock()
		if busy {
			result(0x2203, 0)
			return
		}
		w := packet.New(0x2202)
		w.FixedString(s.char.Name, 21)
		if err := target.Send(w.MustFinish(), "friend invitation"); err != nil {
			s.log.Printf("friend invite: %v", err)
		}
	case 0x2203:
		if len(p) < 3 {
			return
		}
		socialState.Lock()
		invite, ok := socialState.invites[s.char.ID]
		delete(socialState.invites, s.char.ID)
		socialState.Unlock()
		if !ok || time.Now().After(invite.until) {
			return
		}
		sender := FindPlayerByID(invite.from)
		if sender == nil {
			return
		}
		code := uint8(0)
		if p[2] == 1 && sender.Faction() == s.faction && !db.SocialBlocked(s.char.ID, sender.CharacterID) {
			if err := db.AddFriends(sender.Snapshot(), *s.char); err != nil {
				s.log.Printf("add friends: %v", err)
			} else {
				code = 1
			}
		}
		result(0x2203, code)
		w := packet.New(0x2203)
		w.U8(code)
		if err := sender.Send(w.MustFinish(), "friend invitation result"); err != nil {
			s.log.Printf("friend result: %v", err)
		}
		s.sendSocial()
		if code == 1 {
			if sender.owner != nil {
				sender.owner.socialRefresh.Store(true)
			}
			if friends, err := db.SocialContacts(sender.CharacterID, false); err == nil {
				if err = sender.Send(buildSocialList(friends, false), "friend list"); err != nil {
					s.log.Printf("friend list send: %v", err)
				}
			}
		}
	case 0x2205, 0x220a:
		target := id()
		if target == 0 {
			return
		}
		if err := db.RemoveSocial(s.char.ID, target, op == 0x220a); err != nil {
			s.log.Printf("remove social: %v", err)
			return
		}
		w := packet.New(op)
		w.U32(target)
		s.send(w.MustFinish(), "social removal")
		s.sendSocial()
	case 0x2209:
		target, err := db.SocialCharacter(name())
		if err != nil || target.ID == s.char.ID {
			result(0x220e, 0)
			return
		}
		if err = db.AddSocial(s.char.ID, target, true); err != nil {
			s.log.Printf("block: %v", err)
			result(0x220e, 0)
			return
		}
		w := packet.New(0x2209)
		w.U32(target.ID)
		w.FixedString(target.Name, 21)
		s.send(w.MustFinish(), "block added")
		result(0x220e, 1)
		s.sendSocial()
	case 0x2206, 0x220b:
		if len(p) < 7 {
			return
		}
		n := int(p[6])
		if n < 1 || n > 51 || len(p) != 7+n {
			return
		}
		memo := strings.TrimRight(string(p[7:]), "\x00")
		if err := db.SocialMemo(s.char.ID, id(), memo, op == 0x220b); err != nil {
			s.log.Printf("social memo: %v", err)
			return
		}
		w := packet.New(op)
		w.U32(id())
		w.U8(uint8(len(memo) + 1))
		w.Bytes(append([]byte(memo), 0))
		s.send(w.MustFinish(), "social memo")
		s.sendSocial()
	case 0x220c:
		socialState.Lock()
		socialState.searching[s.char.ID] = true
		socialState.Unlock()
		result(0x220c, 1)
		s.sendPartySearch()
	}
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

## Opcode 2203

**`0x2203` — FriendAnswer**

Declarations: [proto.FriendAnswer](../../internal/proto/social.go#L8).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L38), [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L135)

**S→C serializers:** [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L90)

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L47), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleSocial — server/internal/world/social.go</summary>

[Source definition](../../server/internal/world/social.go#L90)

```go
func (s *Session) handleSocial(op packet.Opcode, p []byte) {
	if s.char == nil {
		return
	}
	result := func(op packet.Opcode, code uint8) {
		w := packet.New(op)
		w.U8(code)
		s.send(w.MustFinish(), "social result")
	}
	id := func() uint32 {
		if len(p) < 6 {
			return 0
		}
		v, _ := packet.U32(p[2:])
		return v
	}
	name := func() string {
		if len(p) < 23 {
			return ""
		}
		return strings.TrimSpace(strings.TrimRight(string(p[2:23]), "\x00"))
	}
	switch op {
	case 0x2202:
		target := FindPlayerByName(name())
		if target == nil || target.CharacterID == s.char.ID || target.Faction() != s.faction || target.MapID() != s.char.MapID || db.SocialBlocked(target.CharacterID, s.char.ID) {
			result(0x2203, 2)
			return
		}
		socialState.Lock()
		previous, busy := socialState.invites[target.CharacterID]
		busy = busy && time.Now().Before(previous.until)
		if !busy {
			socialState.invites[target.CharacterID] = socialInvite{s.char.ID, time.Now().Add(30 * time.Second)}
		}
		socialState.Unlock()
		if busy {
			result(0x2203, 0)
			return
		}
		w := packet.New(0x2202)
		w.FixedString(s.char.Name, 21)
		if err := target.Send(w.MustFinish(), "friend invitation"); err != nil {
			s.log.Printf("friend invite: %v", err)
		}
	case 0x2203:
		if len(p) < 3 {
			return
		}
		socialState.Lock()
		invite, ok := socialState.invites[s.char.ID]
		delete(socialState.invites, s.char.ID)
		socialState.Unlock()
		if !ok || time.Now().After(invite.until) {
			return
		}
		sender := FindPlayerByID(invite.from)
		if sender == nil {
			return
		}
		code := uint8(0)
		if p[2] == 1 && sender.Faction() == s.faction && !db.SocialBlocked(s.char.ID, sender.CharacterID) {
			if err := db.AddFriends(sender.Snapshot(), *s.char); err != nil {
				s.log.Printf("add friends: %v", err)
			} else {
				code = 1
			}
		}
		result(0x2203, code)
		w := packet.New(0x2203)
		w.U8(code)
		if err := sender.Send(w.MustFinish(), "friend invitation result"); err != nil {
			s.log.Printf("friend result: %v", err)
		}
		s.sendSocial()
		if code == 1 {
			if sender.owner != nil {
				sender.owner.socialRefresh.Store(true)
			}
			if friends, err := db.SocialContacts(sender.CharacterID, false); err == nil {
				if err = sender.Send(buildSocialList(friends, false), "friend list"); err != nil {
					s.log.Printf("friend list send: %v", err)
				}
			}
		}
	case 0x2205, 0x220a:
		target := id()
		if target == 0 {
			return
		}
		if err := db.RemoveSocial(s.char.ID, target, op == 0x220a); err != nil {
			s.log.Printf("remove social: %v", err)
			return
		}
		w := packet.New(op)
		w.U32(target)
		s.send(w.MustFinish(), "social removal")
		s.sendSocial()
	case 0x2209:
		target, err := db.SocialCharacter(name())
		if err != nil || target.ID == s.char.ID {
			result(0x220e, 0)
			return
		}
		if err = db.AddSocial(s.char.ID, target, true); err != nil {
			s.log.Printf("block: %v", err)
			result(0x220e, 0)
			return
		}
		w := packet.New(0x2209)
		w.U32(target.ID)
		w.FixedString(target.Name, 21)
		s.send(w.MustFinish(), "block added")
		result(0x220e, 1)
		s.sendSocial()
	case 0x2206, 0x220b:
		if len(p) < 7 {
			return
		}
		n := int(p[6])
		if n < 1 || n > 51 || len(p) != 7+n {
			return
		}
		memo := strings.TrimRight(string(p[7:]), "\x00")
		if err := db.SocialMemo(s.char.ID, id(), memo, op == 0x220b); err != nil {
			s.log.Printf("social memo: %v", err)
			return
		}
		w := packet.New(op)
		w.U32(id())
		w.U8(uint8(len(memo) + 1))
		w.Bytes(append([]byte(memo), 0))
		s.send(w.MustFinish(), "social memo")
		s.sendSocial()
	case 0x220c:
		socialState.Lock()
		socialState.searching[s.char.ID] = true
		socialState.Unlock()
		result(0x220c, 1)
		s.sendPartySearch()
	}
}
```

</details>

<details><summary>FriendAnswerPacket — internal/proto/social.go</summary>

[Source definition](../../internal/proto/social.go#L86)

```go
func FriendAnswerPacket(accept bool) []byte {
	w := NewWriter(FriendAnswer)
	if accept {
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

## Opcode 2204

**`0x2204` — FriendAdded**

Declarations: [proto.FriendAdded](../../internal/proto/social.go#L9).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L66), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

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

## Opcode 2205

**`0x2205` — FriendDelete**

Declarations: [proto.FriendDelete](../../internal/proto/social.go#L10).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L38), [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L175)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L55), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

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

## Opcode 2206

**`0x2206` — FriendMemo**

Declarations: [proto.FriendMemo](../../internal/proto/social.go#L11).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L38), [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L205)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L99), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

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

## Opcode 2207

**`0x2207` — FriendPresence**

Declarations: [proto.FriendPresence](../../internal/proto/social.go#L12).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L88), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

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

## Opcode 2208

**`0x2208` — BlockList**

Declarations: [proto.BlockList](../../internal/proto/social.go#L13).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L19), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeSocialList — internal/proto/social.go</summary>

[Source definition](../../internal/proto/social.go#L34)

```go
func DecodeSocialList(p []byte) ([]SocialContact, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > 50 {
		return nil, fmt.Errorf("social list count %d", n)
	}
	rows := make([]SocialContact, n)
	for i := range rows {
		v := &rows[i]
		v.ID = r.U32()
		if Op(p) == FriendList {
			v.Class = r.U8()
			v.Online = r.U8() != 0
		}
		v.Name = r.Fixed(21)
		v.Memo = r.Text()
		if len(v.Memo) > 50 {
			return nil, fmt.Errorf("social memo too long")
		}
	}
	if err := r.Err(); err != nil {
		return nil, err
	}
	if r.Remaining() != 0 {
		return nil, fmt.Errorf("social list trailing data")
	}
	return rows, nil
}
```

```go
type SocialContact struct {
	ID         uint32
	Class      uint8
	Online     bool
	Name, Memo string
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

## Opcode 2209

**`0x2209` — BlockAdd**

Declarations: [proto.BlockAdd](../../internal/proto/social.go#L14).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L38), [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L188)

**S→C serializers:** [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L90)

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L66), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handleSocial — server/internal/world/social.go</summary>

[Source definition](../../server/internal/world/social.go#L90)

```go
func (s *Session) handleSocial(op packet.Opcode, p []byte) {
	if s.char == nil {
		return
	}
	result := func(op packet.Opcode, code uint8) {
		w := packet.New(op)
		w.U8(code)
		s.send(w.MustFinish(), "social result")
	}
	id := func() uint32 {
		if len(p) < 6 {
			return 0
		}
		v, _ := packet.U32(p[2:])
		return v
	}
	name := func() string {
		if len(p) < 23 {
			return ""
		}
		return strings.TrimSpace(strings.TrimRight(string(p[2:23]), "\x00"))
	}
	switch op {
	case 0x2202:
		target := FindPlayerByName(name())
		if target == nil || target.CharacterID == s.char.ID || target.Faction() != s.faction || target.MapID() != s.char.MapID || db.SocialBlocked(target.CharacterID, s.char.ID) {
			result(0x2203, 2)
			return
		}
		socialState.Lock()
		previous, busy := socialState.invites[target.CharacterID]
		busy = busy && time.Now().Before(previous.until)
		if !busy {
			socialState.invites[target.CharacterID] = socialInvite{s.char.ID, time.Now().Add(30 * time.Second)}
		}
		socialState.Unlock()
		if busy {
			result(0x2203, 0)
			return
		}
		w := packet.New(0x2202)
		w.FixedString(s.char.Name, 21)
		if err := target.Send(w.MustFinish(), "friend invitation"); err != nil {
			s.log.Printf("friend invite: %v", err)
		}
	case 0x2203:
		if len(p) < 3 {
			return
		}
		socialState.Lock()
		invite, ok := socialState.invites[s.char.ID]
		delete(socialState.invites, s.char.ID)
		socialState.Unlock()
		if !ok || time.Now().After(invite.until) {
			return
		}
		sender := FindPlayerByID(invite.from)
		if sender == nil {
			return
		}
		code := uint8(0)
		if p[2] == 1 && sender.Faction() == s.faction && !db.SocialBlocked(s.char.ID, sender.CharacterID) {
			if err := db.AddFriends(sender.Snapshot(), *s.char); err != nil {
				s.log.Printf("add friends: %v", err)
			} else {
				code = 1
			}
		}
		result(0x2203, code)
		w := packet.New(0x2203)
		w.U8(code)
		if err := sender.Send(w.MustFinish(), "friend invitation result"); err != nil {
			s.log.Printf("friend result: %v", err)
		}
		s.sendSocial()
		if code == 1 {
			if sender.owner != nil {
				sender.owner.socialRefresh.Store(true)
			}
			if friends, err := db.SocialContacts(sender.CharacterID, false); err == nil {
				if err = sender.Send(buildSocialList(friends, false), "friend list"); err != nil {
					s.log.Printf("friend list send: %v", err)
				}
			}
		}
	case 0x2205, 0x220a:
		target := id()
		if target == 0 {
			return
		}
		if err := db.RemoveSocial(s.char.ID, target, op == 0x220a); err != nil {
			s.log.Printf("remove social: %v", err)
			return
		}
		w := packet.New(op)
		w.U32(target)
		s.send(w.MustFinish(), "social removal")
		s.sendSocial()
	case 0x2209:
		target, err := db.SocialCharacter(name())
		if err != nil || target.ID == s.char.ID {
			result(0x220e, 0)
			return
		}
		if err = db.AddSocial(s.char.ID, target, true); err != nil {
			s.log.Printf("block: %v", err)
			result(0x220e, 0)
			return
		}
		w := packet.New(0x2209)
		w.U32(target.ID)
		w.FixedString(target.Name, 21)
		s.send(w.MustFinish(), "block added")
		result(0x220e, 1)
		s.sendSocial()
	case 0x2206, 0x220b:
		if len(p) < 7 {
			return
		}
		n := int(p[6])
		if n < 1 || n > 51 || len(p) != 7+n {
			return
		}
		memo := strings.TrimRight(string(p[7:]), "\x00")
		if err := db.SocialMemo(s.char.ID, id(), memo, op == 0x220b); err != nil {
			s.log.Printf("social memo: %v", err)
			return
		}
		w := packet.New(op)
		w.U32(id())
		w.U8(uint8(len(memo) + 1))
		w.Bytes(append([]byte(memo), 0))
		s.send(w.MustFinish(), "social memo")
		s.sendSocial()
	case 0x220c:
		socialState.Lock()
		socialState.searching[s.char.ID] = true
		socialState.Unlock()
		result(0x220c, 1)
		s.sendPartySearch()
	}
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

## Opcode 220a

**`0x220A` — BlockDelete**

Declarations: [proto.BlockDelete](../../internal/proto/social.go#L15).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L38), [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L175)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L55), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

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

## Opcode 220b

**`0x220B` — BlockMemo**

Declarations: [proto.BlockMemo](../../internal/proto/social.go#L16).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L38), [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L205)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L99), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

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

## Opcode 220c

**`0x220C` — PartySearchRequest**

Declarations: [proto.PartySearchRequest](../../internal/proto/social.go#L17).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L38), [handleSocial (server/internal/world/social.go)](../../server/internal/world/social.go#L224)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L47), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

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

## Opcode 220d

**`0x220D` — PartySearchList**

Declarations: [proto.PartySearchList](../../internal/proto/social.go#L18).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [sendPartySearch (server/internal/world/social.go)](../../server/internal/world/social.go#L232)

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L33), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>sendPartySearch — server/internal/world/social.go</summary>

[Source definition](../../server/internal/world/social.go#L232)

```go
func (s *Session) sendPartySearch() {
	rows := []db.Character{}
	for _, p := range AllPlayers() {
		socialState.Lock()
		registered := socialState.searching[p.CharacterID]
		socialState.Unlock()
		if !registered {
			continue
		}
		if p.CharacterID == s.char.ID || p.Faction() != s.faction {
			continue
		}
		partyMu.Lock()
		available := partyOf(p) == nil
		partyMu.Unlock()
		if available {
			rows = append(rows, p.Snapshot())
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	w := packet.New(0x220d)
	w.U8(uint8(min(len(rows), 100)))
	for _, r := range rows[:min(len(rows), 100)] {
		w.U8(uint8(min(r.Level, 255)))
		w.U8(r.Class)
		w.FixedString(r.Name, 21)
	}
	s.send(w.MustFinish(), "party search list")
}
```

</details>

<details><summary>DecodePartySearch — internal/proto/social.go</summary>

[Source definition](../../internal/proto/social.go#L62)

```go
func DecodePartySearch(p []byte) ([]PartySearchEntry, error) {
	r := NewReader(p)
	n := int(r.U8())
	if n > 100 {
		return nil, fmt.Errorf("party search count")
	}
	rows := make([]PartySearchEntry, n)
	for i := range rows {
		rows[i] = PartySearchEntry{r.U8(), r.U8(), r.Fixed(21)}
	}
	if err := r.Err(); err != nil {
		return nil, err
	}
	if r.Remaining() != 0 {
		return nil, fmt.Errorf("party search trailing data")
	}
	return rows, nil
}
```

```go
type PartySearchEntry struct {
	Level, Class uint8
	Name         string
}
```

</details>

## Opcode 220e

**`0x220E` — SocialError**

Declarations: [proto.SocialError](../../internal/proto/social.go#L19).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applySocial (internal/game/social.go)](../../internal/game/social.go#L47), [Apply (internal/game/state.go)](../../internal/game/state.go#L294)

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

## Opcode 2401

**`0x2401` — DuelRequest**

Declarations: [proto.DuelRequest](../../internal/proto/duel.go#L4).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L233), [handleDuel (server/internal/world/duel.go)](../../server/internal/world/duel.go#L256)

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L35), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
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

<details><summary>DecodeDuelChallenge — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L25)

```go
func DecodeDuelChallenge(p []byte) (d DuelChallenge, err error) {
	r := NewReader(p)
	d = DuelChallenge{r.U32(), r.U32()}
	return d, r.Err()
}
```

```go
type DuelChallenge struct{ Requester, Target uint32 }
```

</details>

## Opcode 2402

**`0x2402` — DuelAnswer**

Declarations: [proto.DuelAnswer](../../internal/proto/duel.go#L5).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L233), [handleDuel (server/internal/world/duel.go)](../../server/internal/world/duel.go#L261)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L49), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DuelAnswerPacket — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L42)

```go
func DuelAnswerPacket(accept bool) []byte {
	w := NewWriter(DuelAnswer)
	if accept {
		w.U8(1)
	} else {
		w.U8(0)
	}
	return w.Bytes()
}
```

</details>

<details><summary>DecodeDuelResponse — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L30)

```go
func DecodeDuelResponse(p []byte) (d DuelResponse, err error) {
	r := NewReader(p)
	d = DuelResponse{r.U8(), r.U32()}
	return d, r.Err()
}
```

```go
type DuelResponse struct {
	Result uint8
	Player uint32
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

## Opcode 2403

**`0x2403` — DuelCountdown**

Declarations: [proto.DuelCountdown](../../internal/proto/duel.go#L6).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L69), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeDuelCenter — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L35)

```go
func DecodeDuelCenter(p []byte) (x, z float32, err error) {
	r := NewReader(p)
	x, z = r.F32(), r.F32()
	return x, z, r.Err()
}
```

</details>

## Opcode 2404

**`0x2404` — DuelStart**

Declarations: [proto.DuelStart](../../internal/proto/duel.go#L7).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L75), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeDuelEmpty — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L40)

```go
func DecodeDuelEmpty(p []byte) error     { return NewReader(p).Err() }
```

</details>

## Opcode 2405

**`0x2405` — DuelResult**

Declarations: [proto.DuelResult](../../internal/proto/duel.go#L8).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L80), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeExchangeByte — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L31)

```go
func DecodeExchangeByte(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 2406

**`0x2406` — DuelEnd**

Declarations: [proto.DuelEnd](../../internal/proto/duel.go#L9).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L233), [handleDuel (server/internal/world/duel.go)](../../server/internal/world/duel.go#L265)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L49), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DuelGiveUpPacket — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L51)

```go
func DuelGiveUpPacket() []byte       { return NewWriter(DuelEnd).Bytes() }
```

</details>

<details><summary>DecodeDuelResponse — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L30)

```go
func DecodeDuelResponse(p []byte) (d DuelResponse, err error) {
	r := NewReader(p)
	d = DuelResponse{r.U8(), r.U32()}
	return d, r.Err()
}
```

```go
type DuelResponse struct {
	Result uint8
	Player uint32
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

## Opcode 2407

**`0x2407` — DuelWindow**

Declarations: [proto.DuelWindow](../../internal/proto/duel.go#L10).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L59), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
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

## Opcode 2408

**`0x2408` — DuelWindowEnd**

Declarations: [proto.DuelWindowEnd](../../internal/proto/duel.go#L11).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L87), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeExchangeByte — internal/proto/exchange.go</summary>

[Source definition](../../internal/proto/exchange.go#L31)

```go
func DecodeExchangeByte(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 2409

**`0x2409` — DuelReady**

Declarations: [proto.DuelReady](../../internal/proto/duel.go#L12).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L233), [handleDuel (server/internal/world/duel.go)](../../server/internal/world/duel.go#L267)

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L96), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DuelReadyPacket — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L52)

```go
func DuelReadyPacket(v uint8) []byte { w := NewWriter(DuelReady); w.U8(v); return w.Bytes() }
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

## Opcode 240a

**`0x240A` — DuelAdd**

Declarations: [proto.DuelAdd](../../internal/proto/duel.go#L13).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L233), [handleDuel (server/internal/world/duel.go)](../../server/internal/world/duel.go#L271)

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L123), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DuelAddPacket — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L53)

```go
func DuelAddPacket(bag, slot uint8, count uint16, dest uint8) []byte {
	w := NewWriter(DuelAdd)
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

## Opcode 240b

**`0x240B` — DuelRemove**

Declarations: [proto.DuelRemove](../../internal/proto/duel.go#L14).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L233), [handleDuel (server/internal/world/duel.go)](../../server/internal/world/duel.go#L279)

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L139), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DuelRemovePacket — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L60)

```go
func DuelRemovePacket(slot uint8) []byte  { w := NewWriter(DuelRemove); w.U8(slot); return w.Bytes() }
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

## Opcode 240c

**`0x240C` — DuelMoney**

Declarations: [proto.DuelMoney](../../internal/proto/duel.go#L15).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L233), [handleDuel (server/internal/world/duel.go)](../../server/internal/world/duel.go#L283)

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L110), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DuelMoneyPacket — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L61)

```go
func DuelMoneyPacket(money uint32) []byte { w := NewWriter(DuelMoney); w.U32(money); return w.Bytes() }
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

## Opcode 240d

**`0x240D` — DuelAddOther**

Declarations: [proto.DuelAddOther](../../internal/proto/duel.go#L16).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyDuel (internal/game/duel.go)](../../internal/game/duel.go#L133), [Apply (internal/game/state.go)](../../internal/game/state.go#L643)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
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

## Opcode 2501

**`0x2501` — GuildDuelRequest**

Declarations: [proto.GuildDuelRequest](../../internal/proto/guild_duel.go#L4).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L235), [handleGuildDuel (server/internal/world/guildduel.go)](../../server/internal/world/guildduel.go#L125)

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89), [idPacket (internal/proto/world.go)](../../internal/proto/world.go#L39)

**Client receive reducers:** [applyGuildDuel (internal/game/guild_duel.go)](../../internal/game/guild_duel.go#L25), [Apply (internal/game/state.go)](../../internal/game/state.go#L631)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
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

<details><summary>DecodeDuelChallenge — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L25)

```go
func DecodeDuelChallenge(p []byte) (d DuelChallenge, err error) {
	r := NewReader(p)
	d = DuelChallenge{r.U32(), r.U32()}
	return d, r.Err()
}
```

```go
type DuelChallenge struct{ Requester, Target uint32 }
```

</details>

## Opcode 2502

**`0x2502` — GuildDuelAnswer**

Declarations: [proto.GuildDuelAnswer](../../internal/proto/guild_duel.go#L5).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L235), [handleGuildDuel (server/internal/world/guildduel.go)](../../server/internal/world/guildduel.go#L130)

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyGuildDuel (internal/game/guild_duel.go)](../../internal/game/guild_duel.go#L41), [Apply (internal/game/state.go)](../../internal/game/state.go#L631)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>GuildDuelAnswerPacket — internal/proto/guild_duel.go</summary>

[Source definition](../../internal/proto/guild_duel.go#L23)

```go
func GuildDuelAnswerPacket(accept bool) []byte {
	w := NewWriter(GuildDuelAnswer)
	if accept {
		w.U8(1)
	} else {
		w.U8(0)
	}
	return w.Bytes()
}
```

</details>

<details><summary>DecodeDuelResponse — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L30)

```go
func DecodeDuelResponse(p []byte) (d DuelResponse, err error) {
	r := NewReader(p)
	d = DuelResponse{r.U8(), r.U32()}
	return d, r.Err()
}
```

```go
type DuelResponse struct {
	Result uint8
	Player uint32
}
```

</details>

## Opcode 2503

**`0x2503` — GuildDuelCountdown**

Declarations: [proto.GuildDuelCountdown](../../internal/proto/guild_duel.go#L6).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyGuildDuel (internal/game/guild_duel.go)](../../internal/game/guild_duel.go#L52), [Apply (internal/game/state.go)](../../internal/game/state.go#L631)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeGuildDuelCenter — internal/proto/guild_duel.go</summary>

[Source definition](../../internal/proto/guild_duel.go#L17)

```go
func DecodeGuildDuelCenter(p []byte) (GuildDuelCenter, error) {
	r := NewReader(p)
	v := GuildDuelCenter{X: r.F32(), Z: r.F32(), OpponentGuild: r.U32()}
	return v, r.Err()
}
```

```go
type GuildDuelCenter struct {
	X, Z          float32
	OpponentGuild uint32
}
```

</details>

## Opcode 2504

**`0x2504` — GuildDuelStart**

Declarations: [proto.GuildDuelStart](../../internal/proto/guild_duel.go#L7).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyGuildDuel (internal/game/guild_duel.go)](../../internal/game/guild_duel.go#L59), [Apply (internal/game/state.go)](../../internal/game/state.go#L631)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeDuelEmpty — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L40)

```go
func DecodeDuelEmpty(p []byte) error     { return NewReader(p).Err() }
```

</details>

## Opcode 2505

**`0x2505` — GuildDuelResult**

Declarations: [proto.GuildDuelResult](../../internal/proto/guild_duel.go#L8).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** [applyGuildDuel (internal/game/guild_duel.go)](../../internal/game/guild_duel.go#L64), [Apply (internal/game/state.go)](../../internal/game/state.go#L631)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>DecodeU8 — internal/proto/world.go</summary>

[Source definition](../../internal/proto/world.go#L194)

```go
func DecodeU8(p []byte) (uint8, error) { r := NewReader(p); v := r.U8(); return v, r.Err() }
```

</details>

## Opcode 2506

**`0x2506` — GuildDuelLeave**

Declarations: [proto.GuildDuelLeave](../../internal/proto/guild_duel.go#L9).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [duelPacket (server/internal/world/duel.go)](../../server/internal/world/duel.go#L89)

**Client receive reducers:** [applyGuildDuel (internal/game/guild_duel.go)](../../internal/game/guild_duel.go#L71), [Apply (internal/game/state.go)](../../internal/game/state.go#L631)

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>duelPacket — server/internal/world/duel.go</summary>

[Source definition](../../server/internal/world/duel.go#L89)

```go
func duelPacket(op packet.Opcode, write func(*packet.Writer)) []byte {
	w := packet.New(op)
	if write != nil {
		write(w)
	}
	return w.MustFinish()
}
```

</details>

<details><summary>DecodeDuelResponse — internal/proto/duel.go</summary>

[Source definition](../../internal/proto/duel.go#L30)

```go
func DecodeDuelResponse(p []byte) (d DuelResponse, err error) {
	r := NewReader(p)
	d = DuelResponse{r.U8(), r.U32()}
	return d, r.Err()
}
```

```go
type DuelResponse struct {
	Result uint8
	Player uint32
}
```

</details>
