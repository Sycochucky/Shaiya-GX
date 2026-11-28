# Server security, resource limits and observability

Security policy is enforced by the server before application mutation and again
at the storage boundary. Shared development accounts/seed are intentional content;
modernization does not delete them. [Architecture](server-architecture.md),
[package contracts](server-packages.md) and [wire reference](server-protocol.md)
describe the implementation rather than a blanket security certification.

## Credentials and authorization

Passwords use salted, versioned Argon2id: 19 MiB, two iterations, one lane and at
most two simultaneous hashing workers/process. Encoded hash parameters are
strictly validated. Unknown/failed credentials perform dummy work to reduce
enumeration timing. Legacy credentials upgrade on successful login; new account
and password commands write hashes directly. No seed rewrite is needed.

From the server installation root:

```sh
./build/gx_console account create example
./build/gx_console account create example_gm --admin
./build/gx_console account password example
```

Use `.exe` on Windows. Password input is hidden at a terminal; automation can
provide stdin or `RAPTURE_ACCOUNT_PASSWORD`. Avoid shell history and CLI password
arguments. Usernames allow 1–18 UTF-8 bytes, passwords 1–128; no embedded NUL or
invalid UTF-8. Longer passwords use `0x7F12`, keeping legacy `0xA102` unchanged.

Account role/status comes from SQLite, never a client flag. Unknown status values
fail closed. Handoffs are single-use transactional tickets carrying credential
version, expire after ten minutes and cannot be burned by an invalid token.
The world refuses duplicate active account ownership. Password reset increments
version/removes tickets; live account status/role/version checks run on a
30-second interval, so revocation is bounded rather than instant.

## TLS and portable client trust

Remote listeners require TLS; loopback remains a permitted development mode.
Provision a private development CA and server certificate into a new directory:

```sh
./build/gx_console tls-init -out /private/gx-tls -dns gx.example -ip 127.0.0.1,192.168.1.20
```

Set `RAPTURE_TLS_CERT` and `RAPTURE_TLS_KEY` to the generated server.env paths for
both cores. Copy **only** generated `client/network.json` and `client/gx-ca.pem`
to the playable client's `data/` folder. Never distribute server/CA private keys.
The leaf lasts 90 days and the CA five years; provision/replace certificates before
expiry. The command refuses an existing destination instead of destroying keys.

Client `data/network.json` next to the executable supports:

```json
{
  "tls": true,
  "ca_file": "gx-ca.pem",
  "server_name": "gx.example"
}
```

Relative CA paths resolve beside that configuration file. `GX_NETWORK_CONFIG`
selects an explicit file. `GX_TLS`, `GX_TLS_CA` and `GX_TLS_SERVER_NAME` override
individual fields. With public/system trust use `GX_TLS=1` and the appropriate
certificate identity. `GX_TLS=0` selects development plaintext only when no CA
or server name is configured; configured trust always requires TLS. Invalid
certificates or names never trigger an insecure fallback. TLS minimum is 1.3.

For an intentionally insecure development LAN, both server
`RAPTURE_ALLOW_INSECURE=1` and client plaintext policy must be explicit. This is an
opt-in compatibility path, not the normal remote setup. Existing local loopback
configuration is preserved. Legacy RSA/AES/XOR is not a replacement for TLS.

## Network resource policy

Policies are validated before services accept clients. Invalid zero, malformed
or excessive overrides fail startup rather than silently removing protection.
Token buckets refill continuously; "per minute/second" below means refill/burst
policy, not an unbounded array of historical attempts.

| Setting | Default | Environment key |
|---|---:|---|
| Login simultaneous connections | 200 | `RAPTURE_LOGIN_MAX_CONNECTIONS` |
| World simultaneous connections | 1,200 | `RAPTURE_WORLD_MAX_CONNECTIONS` |
| Simultaneous connections per source IP | 32 | `RAPTURE_CONNECTIONS_PER_IP` |
| New connections per source IP/minute | 120 | `RAPTURE_CONNECTIONS_PER_IP_PER_MINUTE` |
| Login attempts per source IP/minute | 30 | `RAPTURE_LOGIN_ATTEMPTS_PER_IP` |
| Login attempts per account/minute | 10 | `RAPTURE_LOGIN_ATTEMPTS_PER_ACCOUNT` |
| World input frames/second | 512 | `RAPTURE_INPUT_FRAMES_PER_SECOND` |
| World input bytes/second | 1,048,576 | `RAPTURE_INPUT_BYTES_PER_SECOND` |
| Queued output packets/connection | 256 | `RAPTURE_OUTPUT_PACKETS` |
| Queued output bytes/connection, including active write | 1,048,576 | `RAPTURE_OUTPUT_BYTES` |
| Economy actions/second | 20 | `RAPTURE_ECONOMY_ACTIONS_PER_SECOND` |

See [policy.go](../../server/internal/config/policy.go) for accepted upper bounds.
These are safety budgets, not a supported maximum-player benchmark. Shared-NAT
connections can need measured tuning. Account/IP rate maps are memory-bounded and
reclaim idle entries; filling the source budget fails closed.

TLS handshakes have a ten-second context; login protocol lifetime is 30 seconds
with five-second writes. World output has ordered bounded asynchronous writes and
a five-second final drain. Overflow aborts that connection instead of silently
dropping authoritative messages. Partial input frames are also budgeted before
application ticks/dispatch.

GX Shield is a local, time-limited admission mode in GX Server Control. Standard
limits remain active. Under Attack reduces each source to 30 new connections per
minute and 16 concurrent connections (or the lower configured ordinary limit).
It affects new connections only. Login also charges a separate per-source budget
before RSA decryption, using the configured login-attempt count. At most 32 RSA
exchanges run simultaneously. The dashboard
uses the authenticated loopback control channel to set the mode on both cores.
Its automatic trigger requires 64 global-capacity rejections from at least four
source IPs within ten seconds;
the automatic mode lasts ten minutes. Manual mode lasts 30 minutes and can be
reversed. Restarting a core resets its mode to standard. The admission limits
do not protect against link saturation or replace upstream TCP DDoS mitigation.
GX uses the socket's remote IP; it does not parse PROXY protocol. A TCP proxy
that hides client addresses places all players in the same per-IP budget.

## Gameplay queue policy

| Queue/state | Bound / behavior |
|---|---|
| Pending combat/event mailboxes | 4,096 entries; disconnect unresponsive recipient on overflow |
| Pending item grants | 256; caller uses original ground-loot fallback on refusal |
| Experience/gold rewards | Fixed accumulators with saturating numeric bounds and commit acknowledgement |
| Quest-kill counters | Definition-keyed bounded counters for relevant active quests; acknowledge durable batches |
| Inventory-operation publication | 256 deferred packets before commit; overflow rejects plan |
| AI command queue | 1,024 commands and weighted cost 2,048 |
| Summoned AI mobs | 4,096 active/pending budget, enforced at execution too |

See [players](../../server/internal/world/players.go),
[inventory operation](../../server/internal/world/inventory_operation.go) and
[AI commands](../../server/internal/world/mob_ai.go). The overload behavior is
part of the domain contract; deleting earned economic messages to reduce memory
would change game state incorrectly.

## Structured logging

JSON records contain `time` (UTC), `level`, `component`, `message`, optional
`session` and explicit event attributes. Session tags can identify a connection/
account/character, so protect personal logs. Passwords, ticket identities and
AES material are not diagnostic fields. New log files are private; no runtime
logs or packet captures belong in source history.

Files are `logs/gx_login.log` and `logs/gx_world.log`. Default rotation is 16 MiB
with three numbered backups. `RAPTURE_LOG_MAX_MB` accepts 1–1,024.
`RAPTURE_DEBUG=1` enables verbose opcode diagnostics. Existing historical log
content can predate JSON format; parsers must account for an upgraded installation.

## Loopback control and health

Each core opens an ephemeral numeric loopback address unless
`RAPTURE_LOGIN_DIAGNOSTICS` or `RAPTURE_WORLD_DIAGNOSTICS` specifies a loopback
address/port. Non-loopback binds are refused. The private descriptor is
`runtime/control/gx_login.json` or `gx_world.json`, containing PID, executable,
address and a random bearer token. Do not publish descriptors or proxy them to
a public listener.

| Endpoint | Method | Access / behavior |
|---|---|---|
| `/healthz` | GET | Loopback; ready response while process context is active, stopping returns 503 |
| `/metrics` | GET | Loopback; bounded Prometheus text, observes pool/WAL state |
| `/shutdown` | POST | Exact bearer token; accepts cancellation for graceful stop |
| `/shield?minutes=N` | POST | Exact bearer token; 0 returns to standard, 1–60 enables temporary Under Attack admission |
| `/debug/pprof/…` | GET | Available only with `RAPTURE_PPROF=1`; same bearer token, duration capped at 30 s |

Health here is service-lifecycle health, not a gameplay round-trip or backup
verification. The controller consumes descriptors internally, checks PID and
canonical executable, and bypasses HTTP proxies. Normal operator shutdown should
use the console rather than expose tokens in shell history. HTTP header/read/
idle/write timeouts and header-size limits apply. No public account-admin web
service is introduced.

## Metrics and interpretation

| Metric | Interpretation |
|---|---|
| `gx_connections_active` | Current admitted sockets in that core process |
| `gx_connections_accepted_total`, `gx_connections_rejected_total` | Admission activity/refusal; correlate with legitimate shared-NAT load |
| `gx_shield_under_attack`, `gx_shield_rejected_total`, `gx_shield_automatic_activations_total` | Temporary admission mode, its additional rejects and automatic activations |
| `gx_network_output_bytes_pending` | Aggregate queued/in-flight output pressure |
| `gx_network_output_overflows_total` | Slow-client/backpressure disconnects |
| `gx_player_mailbox_overflows_total` | Gameplay mailbox pressure |
| `gx_db_connection_wait_total`, `gx_db_connection_wait_nanoseconds_total` | Go pool wait count/duration, not every SQLite lock wait |
| `gx_db_connections_open` | Open pool connections; deliberate per-process cap is one |
| `gx_db_wal_bytes` | WAL file length; growth alone is not proof of corruption |
| `gx_db_transaction_seconds` | Instrumented guarded transaction histogram |
| `gx_db_transaction_failures_total` | Instrumented transaction failures; inspect logs/domain refusal context |
| `gx_tick_<job>_seconds` | Periodic callback duration histogram |
| `gx_tick_<job>_delay_seconds` | Delay from scheduled execution time |

Histograms expose cumulative buckets/count/sum. Duration boundaries are 1, 5, 10,
50, 100, 500 ms, 1 s and 5 s, plus infinity. Metric names/cardinality are bounded;
no per-account/IP labels. Values belong to each process; scrape login/world
separately. A missing sample or a zero instrumented counter is not proof that all
unrelated code paths succeeded.

If output pressure rises, inspect receiver speed and packet volume. If storage
waits rise, inspect query plans and transaction duration before adding connections.
If tick delay rises, inspect the corresponding callback and Lua/world workload.
Do not hide overload by disabling bounds, weakening durability or dropping rewards.

## Validation and maintenance boundaries

Go 1.27.4 fixes the HTTP-library issue found while adding diagnostics; toolchain
and CI are aligned. The scoped server/transport govulncheck scan reports no
vulnerabilities at modernization validation. This result is time/version-specific;
keep scanning dependencies and runtime patches.

Unit/integration/race/fuzz tests cover authentication, TLS trust/name rejection,
revocation, packet bounds, queue pressure, atomic economy, migration failures,
disk-full rollback and native Linux lifecycle. A local 32-client fixture completed
160 durable purchases with p50 4.5 ms and p95 16.3 ms plus disconnect/relogin checks.
Those figures are not a capacity commitment. Native Windows execution and a
maximum-player benchmark were not performed locally. Certificate rotation,
backups and resource monitoring remain operating responsibilities.
