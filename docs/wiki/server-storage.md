# SQLite schema, transactions and recovery

SQLite is the primary production database for GX Server. All durable game writes
belong behind `server/internal/db`; neither client packets nor UI state are SQL
authority. See [complete storage API](server-reference-storage.md),
[package contracts](server-packages.md) and [operator procedures](server-operations.md).

## Files and process model

| Resource | Meaning |
|---|---|
| `data/rapture.sqlite` | Approved distributed development snapshot; intentionally contains shared accounts/characters/items |
| `data/rapture.manifest.json` | Snapshot hash/table manifest |
| `runtime/rapture.sqlite` | Mutable state, initialized only if absent |
| `runtime/rapture.sqlite-wal`, `-shm` | Live SQLite WAL/shared-memory companions; not standalone backups |
| `<database>.world.lock` | OS-backed one-world ownership lock; a leftover filename is not active ownership |
| `GX_SchemaMigrations` | Checksummed runtime migration ledger |
| `GX_SessionHandoff`, `GX_ActiveSessions` | Transient durable authentication/ownership records in the same runtime |

All former logical databases use prefixed tables in one physical file. Calls to
`Pool(UserData)`, `Pool(GameData)` and `Pool(GameDefs)` share the process writer,
capped at one open connection. Login and World coordinate their writer
connections through SQLite locking. Queries outside write transactions can
use a separate pool of up to four read-only connections per process. This permits account/item/
guild/audit mutations to share an atomic transaction without a distributed
transaction protocol.

Connection settings are `journal_mode=WAL`, `synchronous=FULL`,
`foreign_keys=ON`, `busy_timeout=10000` and IMMEDIATE transaction acquisition.
`PRAGMA user_version=1` is the legacy seed-format marker. `PRAGMA optimize` runs
at graceful close; observations report WAL file size without truncating live WAL.
These are deliberate settings; opening many writers or weakening synchronization
is not the default optimization strategy.

Keep a runtime on a filesystem supporting SQLite WAL/locking semantics. Use the
backup API for snapshots; a copied main file can omit committed WAL changes.
Do not share one runtime between multiple authoritative world processes or bypass
the ownership lease by deleting an open lock file.

## Schema families

| Table family / model | Responsibility and ownership |
|---|---|
| `PS_UserData_*`, including `PS_UserData_Users_Master` | Accounts, credential field, role/status, points and login/logout state |
| `PS_GameData_*`, including `PS_GameData_Chars` | Characters, items, skills, quest state, personal/shared storage and guild game data |
| `PS_GameDefs_*` | Content definitions cached for authoritative item/skill/mob/product/rule lookups |
| Existing guild/social/keep/audit tables | Membership, mirrors, friends/blocks, keep ownership and economic/login audit history; see corresponding API/source SQL |
| `GX_AccountSecurity` | UserUID primary key plus monotonically changed credential Version; missing entry represents version 1 |
| `GX_SessionHandoff` | UserUID primary key, Identity blob, AESMaterial blob, SessionID, ExpiresAt and CredentialVersion |
| `GX_ActiveSessions` | UserUID primary key and exact SessionID of the active world owner |
| `GX_SchemaMigrations` | Version primary key, Name, Checksum, AppliedAt |

Column/table names of the inherited schema are retained where possible. Password
hashes replace the existing Pw value; no parallel plaintext credential field is
introduced. The seeded `Rapture_LoginLog` and `rapture.sqlite` names remain for
compatibility, not a dependency on a separate Rapture service.

The [storage API inventory](server-reference-storage.md) links every file and
function, including all SQL call sites. It is the source map for complete legacy
column shapes; this page groups the schema by ownership instead of inventing a
new relational model for old tables.

## Transactional migrations

[migrations.go](../../server/internal/db/migrations.go) applies all pending steps
and ledger updates in one transaction. It rejects a newer runtime, checksum/name
changes, non-contiguous version definitions and ledger gaps. A failed migration
rolls back both DDL/data and ledger records. Never edit the SQL of an already
published migration; append a new version.

| GX version | Migration | Result |
|---|---|---|
| 1 | `atomic_session_handoff` | Creates the transactional handoff table, accepting an already-created compatible table from the first modernization |
| 2 | `indexed_account_lookup` | Adds `GX_Users_UserID` on Users_Master(UserID) |
| 3 | `credential_revocation` | Adds account credential-version table and handoff CredentialVersion |
| 4 | `warehouse_lookup` | Adds `GX_Warehouse_User` on UserStoredItems(ServerID,UserUID,ItemUID) |
| 5 | `active_account_ownership` | Adds exact-session active account ownership |
| 6 | `character_teleport_points` | Stores four recorded teleport points per character in `GX_TeleportPoints` |
| 7–9 | Skill catalog updates | Removes retired skill levels and updates effect and prerequisite data |
| 10–11 | Enchant rules and target levels | Stores independent enchant curves indexed by the level reached |
| 12 | `mob_ai_messages` | Stores indexed monster AI text |
| 13–14 | Goddess blessing | Imports and merges the faction bonus tables into a shared catalog |
| 15 | `world_drops` | Adds independent chance rules by mob level and item grade |
| 16 | `item_boxes` | Stores box definitions and reward weights or chances |
| 17 | `npc_quest` | Adds six tables for NPCs, shops, destinations, quest details and links |
| 18 | `map_properties` | Adds map rules and racial birth positions |
| 19 | `unified_experience` | Replaces mode-specific EXP rows with one threshold per level and migrates character progress |
| 35 | `remove_growth_modes` | Credits the difference in level points once and neutralizes the legacy mode fields without changing faction or reward claims |
| 20 | `separate_npc_skills` | Moves former level-100 NPC definitions to an independent `SkillID, Rank` catalog |
| 21 | `server_notices` | Adds the former SQLite notice catalog |
| 22 | `teleport_runes` | Adds the former SQLite rune destination catalog and updates item references |
| 23 | `auction_house` | Adds auction, claim and watch tables |
| 24 | `runtime_write_indexes` | Indexes handoff expiry, guild storage, character events, reverse social lookups and live names |
| 25 | `indexed_legacy_row_identity` | Indexes identity allocation for actively written legacy tables |
| 26 | `remove_inactive_sqlite_tables` | Backs up the runtime and drops the retired catalog, Billing, market, backup and audit tables |

The migration ledger is separate from legacy user_version. The shipped seed contains the active schema. Existing installations upgrade their
runtime in place with a verified backup. Back up before
upgrading and keep matching binaries. Downgrading to code that understands only
plaintext passwords or a different migration ledger is unsupported.

## Editing game definitions

[GX Studio](gx-studio.md) edits supported client catalogs, SQLite definitions and NpcQuest rules. GX World reads the game files in `data/game` at startup. The historical
SQLite definition rows can be exported from a pre-26 backup. Migration 26
creates and verifies that backup beside the runtime before dropping the tables. SQLite continues to store durable game state and the
active item, skill, experience and mall catalogs.


## Write concurrency and durability

World and Login each use one WAL writer with full synchronization. Gameplay
reads use up to four read-only connections per process. Independent writes can
share a durable commit with per-operation savepoints. Character exit commits
character state, buffs, skill delays and quest counters in one transaction.

World handoff tickets store a token digest and expire from the runtime.
The local `gx_console account upgrade-legacy` command hashes legacy password
rows in place and reports conversion and reset counts. See the
[SQLite production guide](../../server/docs/sqlite.md) for disk-backed load
validation, administration and backups.

## Transaction contracts

| Application operation | Durable unit | Publication / failure rule |
|---|---|---|
| Account create/password reset | Hash, account row/version, ticket revocation as applicable | Never store new plaintext; changed version invalidates old authorization |
| Login/password upgrade | Recheck captured credential and write login state/upgrade atomically | KDF happens before write lock; concurrent password change cannot authenticate an obsolete snapshot |
| Handoff consume | Validate identity/expiry/version and delete ticket | Exactly one successful consumer; invalid identity does not consume |
| Active account acquire/release | Insert-on-conflict ownership / delete exact SessionID | Duplicate world session refused; old cleanup cannot remove new owner |
| Character creation | Slot/name checks, character and starter inventory | Starter failure leaves no partial character; restore cannot occupy another slot |
| Inventory/crafting/shop | Guarded before-character/items, staged after-character/items | Rollback retains live state; deferred packets/effects publish only after commit |
| Personal warehouse | Character/items, stored inventory and fees | Both sides move together; stale snapshot/replay rejected |
| Guild bank | Character/items, shared bank and fees under guild ownership | Durable mutation before shared cache/acknowledgement |
| Player trade | Both participants' guarded balances and item sets | No partial receiver/source write; both final states durable before success |
| Quest completion | Character resources/items plus finished/active quest transition | Failed reward cannot mark completion; repeated completion cannot duplicate reward |
| Skill learning | Previous skill level plus skill-point cost and new skill level | Cost and grant are one guarded mutation |
| Mall purchase/gift/claim | Point debit, product items or stored gift, audit entries and exact claim state | Insufficient funds/replayed claim cannot partially grant |
| Guild creation/mutations | Master fee, membership/permissions, guild and mirror rows | Cache/public announcements follow commit; final-member deletion/disband atomic |
| Buff/bar/quest snapshots | Whole snapshot/batch | Failure does not leave a half-written replacement |

`CommitGameplay` composes guarded character/item mutations and optional quest
completion. Its internal transaction helper is reusable by warehouse/guild-bank
operations. Guards compare expected balances/resources and inventory state;
`CommitItems` handles inventory-only changes without unrelated character writes.
Do not replace these with independent updates and compensating packets.

`withTx` rolls back on returned error and panic; the commit is the durability
boundary. New code must handle transaction errors explicitly and avoid updating
live pointers/caches or emitting success before that boundary.

## Authentication and crash ownership

Ticket expiry is ten minutes. Identity comparison and credential-version checks
occur in the same transaction as consumption. A new password increments version
and removes old handoffs; connected sessions recheck account status/role/version
on a 30-second authorization interval. This is bounded revocation, not an
instantaneous broadcast guarantee.

A world owner obtains an OS lock before clearing stale `GX_ActiveSessions` rows.
No second world can clear an active owner's rows. Per-account release uses both
UserUID and SessionID; failures enter a bounded retry map serviced by the world
scheduler. Logout accounting is exact-session and idempotent, so an old
disconnect cannot overwrite a newer login or double-credit online time.

## Failure categories

Ordinary domain failures, optimistic conflicts, injected constraint failures and
busy responses do not imply a corrupted database. They roll back or return the
operation's refusal/error path. Persisted state remains authoritative.

`StorageUnavailable` recognizes unsafe SQLite base error classes such as full
disk, IO failure, read-only/unwritable state, corruption and invalid database.
The world registers a storage-failure callback that cancels the service; session
save failures suppress further success publication. Resolve the storage fault
before restart. Do not catch these errors and keep simulating unsaved progress.

Pending in-memory rewards/loot are not a durable event journal. A completed
transaction survives SQLite recovery; an uncommitted operation may abort. A
client that lost the acknowledgement must reconcile by reloading authoritative
state, not assume its last request failed and grant itself compensation.

## Backups, restore and query inspection

Use the operator's [backup/restore procedure](server-operations.md#back-up-and-restore).
The Python tool uses SQLite backup, integrity validation and a manifest:

```sh
python3 tools/snapshot_sqlite.py --source runtime/rapture.sqlite --output /private/backups/gx.sqlite
```

Run from the server installation root. Protect personal backups as account/player
data. Restore only while both services are stopped, retain the previous stopped
runtime, and remove incompatible stale WAL/SHM companions before opening the
restored database. Never replace an open runtime or silently refresh its seed.

Investigate waits/latency and query plans before tuning. For example:

```sql
EXPLAIN QUERY PLAN SELECT UserUID FROM PS_UserData_Users_Master WHERE UserID='example';
PRAGMA integrity_check;
PRAGMA journal_mode;
PRAGMA synchronous;
SELECT Version,Name,AppliedAt FROM GX_SchemaMigrations ORDER BY Version;
```

The account query should use `GX_Users_UserID`; warehouse access has its dedicated
composite index. Small per-character row sets do not justify blanket indexes or
additional connections. `integrity_check` is an explicit diagnostic, not something
to run on every simulation tick.

## Validation scope

Unit/integration tests use temporary databases and explicit fault injection:
source/recipient write failure, stale replay, partial grants, disk-full via
max_page_count, invalid migration/checksum, concurrent ownership, password change,
trade/reward durability and snapshot rollback. A separate process drill verifies
four processes and 800 committed transfers, forced termination before/after commit,
WAL recovery and backup/restore integrity. Real-wire tests inspect persisted state
at acknowledgement and after reconnect.

Tests establish these invariants for their fixtures, not an unlimited throughput
claim or full crash persistence of all live entities. See
[persistence audit](../../server/docs/persistence-audit.md) and the linked tests.
