# Protocol: Login and world handshakes

> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.

[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)

Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.

## Opcode a101

**`0xA101` — LoginHandshake**

Declarations: [packet.LoginHandshake](../../server/internal/packet/packet.go#L211), [proto.LoginHandshake](../../internal/proto/opcodes.go#L199).

**C→S receive adapters:** [frame (server/cmd/gx_login/main.go)](../../server/cmd/gx_login/main.go#L200)

**S→C serializers:** [handshake (server/cmd/gx_login/main.go)](../../server/cmd/gx_login/main.go#L103)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>handshake — server/cmd/gx_login/main.go</summary>

[Source definition](../../server/cmd/gx_login/main.go#L103)

```go
func handshake() []byte {
	exponent, modulus := crypto.LoginPublicKey()
	w := packet.New(packet.LoginHandshake)
	w.U8(0)
	w.U8(uint8(len(exponent)))
	w.U8(uint8(len(modulus)))
	w.Bytes(exponent[:])
	w.Bytes(modulus[:])
	return w.MustFinish()
}
```

</details>

<details><summary>KeyReply — internal/proto/login.go</summary>

[Source definition](../../internal/proto/login.go#L27)

```go
func KeyReply(block []byte) ([]byte, error) {
	if len(block) != 128 {
		return nil, fmt.Errorf("proto: RSA block is %d bytes, the client sends 128", len(block))
	}
	w := NewWriter(LoginHandshake)
	w.U8(128)
	w.Raw(block)
	return w.Bytes(), nil
}
```

</details>

## Opcode a102

**`0xA102` — LoginRequest**

Declarations: [packet.LoginRequest](../../server/internal/packet/packet.go#L212), [proto.LoginRequest](../../internal/proto/opcodes.go#L200).

**C→S receive adapters:** [frame (server/cmd/gx_login/main.go)](../../server/cmd/gx_login/main.go#L226)

**S→C serializers:** [loginReply (server/cmd/gx_login/main.go)](../../server/cmd/gx_login/main.go#L114)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>loginReply — server/cmd/gx_login/main.go</summary>

[Source definition](../../server/cmd/gx_login/main.go#L114)

```go
func loginReply(result uint8, userUID uint32, permission uint8, identity [16]byte) []byte {
	w := packet.New(packet.LoginRequest)
	w.U8(result)
	if result == 0 {
		w.U32(userUID)
		w.U8(permission)
		w.Bytes(identity[:])
	}
	return w.MustFinish()
}
```

</details>

<details><summary>LoginRequestPacket — internal/proto/login.go</summary>

[Source definition](../../internal/proto/login.go#L39)

```go
func LoginRequestPacket(user, password string) []byte {
	w := NewWriter(LoginRequest)
	w.Fixed(user, 32, true)
	w.Fixed(password, 16, false)
	return w.Bytes()
}
```

</details>

## Opcode a106

**`0xA106` — LoginRequestEx**

Declarations: [proto.LoginRequestEx](../../internal/proto/opcodes.go#L232).

Source notes: Opcodes not named in Rapture's catalog. 0x5992f0 with its flag set: 33-byte password

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** no direct mapping found; inspect source before assuming support or direction.

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

The shape is decoded inline, delegated, or only declared. Use the linked adapter/reducer and the codec appendix below; there is no inferred fixed layout here.

## Opcode a201

**`0xA201` — ServerList**

Declarations: [packet.ServerList](../../server/internal/packet/packet.go#L213), [proto.ServerList](../../internal/proto/opcodes.go#L201).

**C→S receive adapters:** no direct mapping found; inspect source before assuming support or direction.

**S→C serializers:** [serverList (server/cmd/gx_login/main.go)](../../server/cmd/gx_login/main.go#L125)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>serverList — server/cmd/gx_login/main.go</summary>

[Source definition](../../server/cmd/gx_login/main.go#L125)

```go
func serverList() []byte {
	w := packet.New(packet.ServerList)
	w.U8(1)    // server count
	w.U8(1)    // server id
	w.U8(0)    // status
	w.U16(0)   // users
	w.U16(100) // capacity
	w.FixedString("GX Server", 32)
	return w.MustFinish()
}
```

</details>

## Opcode a202

**`0xA202` — SelectServer**

Declarations: [packet.SelectServer](../../server/internal/packet/packet.go#L214), [proto.SelectServer](../../internal/proto/opcodes.go#L202).

**C→S receive adapters:** [frame (server/cmd/gx_login/main.go)](../../server/cmd/gx_login/main.go#L275)

**S→C serializers:** [selectServerReply (server/cmd/gx_login/main.go)](../../server/cmd/gx_login/main.go#L136)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>selectServerReply — server/cmd/gx_login/main.go</summary>

[Source definition](../../server/cmd/gx_login/main.go#L136)

```go
func selectServerReply(status uint8) []byte {
	w := packet.New(packet.SelectServer)
	w.U8(status)
	ip := net.ParseIP(config.PublicIP()).To4()
	if ip == nil {
		ip = net.IPv4(127, 0, 0, 1).To4()
	}
	w.Bytes(ip)
	return w.MustFinish()
}
```

</details>

<details><summary>SelectServerPacket — internal/proto/login.go</summary>

[Source definition](../../internal/proto/login.go#L88)

```go
func SelectServerPacket(serverID uint8, version uint32) []byte {
	w := NewWriter(SelectServer)
	w.U8(serverID)
	w.U32(version)
	return w.Bytes()
}
```

</details>

## Opcode a301

**`0xA301` — GameHandshake**

Declarations: [packet.GameHandshake](../../server/internal/packet/packet.go#L215), [proto.GameHandshake](../../internal/proto/opcodes.go#L203).

**C→S receive adapters:** [frame (server/internal/world/session.go)](../../server/internal/world/session.go#L607)

**S→C serializers:** [buildWorldHandshake (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L16)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildWorldHandshake — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L16)

```go
func buildWorldHandshake() []byte {
	w := packet.New(packet.GameHandshake)
	w.U8(0)
	w.U8(2)
	w.Bytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	return w.MustFinish()
}
```

</details>

<details><summary>WorldHelloPacket — internal/proto/login.go</summary>

[Source definition](../../internal/proto/login.go#L110)

```go
func WorldHelloPacket(userUID uint32, token [16]byte) []byte {
	w := NewWriter(GameHandshake)
	w.U32(userUID)
	w.Raw(token[:])
	return w.Bytes()
}
```

</details>

## Opcode a303

**`0xA303` — Ping**

Declarations: [packet.Ping](../../server/internal/packet/packet.go#L216), [proto.Ping](../../internal/proto/opcodes.go#L204).

**C→S receive adapters:** [dispatch (server/internal/world/dispatch.go)](../../server/internal/world/dispatch.go#L298)

**S→C serializers:** [buildPing (server/internal/world/protocol.go)](../../server/internal/world/protocol.go#L101)

**Client receive reducers:** no direct mapping found; inspect source before assuming support or direction.

### Wire layouts and validation

These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.

<details><summary>buildPing — server/internal/world/protocol.go</summary>

[Source definition](../../server/internal/world/protocol.go#L101)

```go
func buildPing() []byte {
	w := packet.New(packet.Ping)
	w.U16(0)
	return w.MustFinish()
}
```

</details>
