# SQLite production database

**SQLite is the primary production database. Its embedded operation simplifies setup, deployment and portable distribution without a separate database service.** Keep gameplay/storage boundaries explicit behind `server/internal/db`.

Current validation and transaction-boundary regression coverage are recorded in the
[persistence and concurrency audit](persistence-audit.md). Passing engine recovery
checks does not imply every gameplay operation is atomic.

## Files and first launch

- `data/rapture.sqlite`: the explicitly requested shared test snapshot, including current development accounts, characters, inventory, guild/social state and game definitions.
- `data/rapture.manifest.json`: table counts and seed SHA-256.
- `runtime/rapture.sqlite`: mutable working copy, ignored by Git. Startup copies the seed only when absent. An existing runtime is never reset by an update.
- `RAPTURE_DB_PATH`, `RAPTURE_DB_SEED`, `RAPTURE_DATA_DIR`: compatible overrides for runtime, seed and supplied server maps/configuration.

The names remain compatible with the former Rapture server. They no longer identify a separate project. Login and world share one SQLite file using `modernc.org/sqlite`; no C SQLite library, credentials, Docker or database daemon is required. Each process has one write connection and up to four read-only connections.
WAL, a bounded busy timeout, full synchronization and immediate transactions
coordinate durable writes. Verified readers observe committed state while a
writer is active. Runtime opening rejects a path that aliases the seed.

The snapshot includes test account credentials and gameplay state by explicit choice. Personal future runtime files, backups, local overrides and logs remain outside Git and releases.

## Schema and behavior

The active `PS_UserData`, `PS_GameData` and `PS_GameDefs` tables share one
SQLite file. The schema retains active source columns, primary keys, indexes
and IDs. Migration 26 removes the unused `PS_Billing` namespace, retired
market tables, historical logs and catalog copies now loaded from INI/SData. Former stored procedures are parameterized queries/transactions in `server/internal/db`. Character creation/deletion/rename, guild changes, item claims, login/logout and point deductions preserve the existing API. The original `Rapture_LoginLog` name remains intact.

Dates use UTC text; integer/blob data is not exported through floating-point CSV. SQLite NOCASE is not identical to every SQL Server non-ASCII collation. SQLite initialization is serialized and atomic. The distributed seed is closed/checkpointed with no required WAL/SHM; a running runtime can have both companions.

## Administration

From `server/` in source, or the extracted portable folder:

```sh
python3 tools/sqlite.py init
python3 tools/sqlite.py account player
python3 tools/sqlite.py query --file query.sql
```

On Windows use `py -3`. Account creation prompts for a password. Automation may use `RAPTURE_ACCOUNT_PASSWORD` in the environment. The account tool never overwrites an existing account. SQL files use SQLite syntax and prefixed table names.

For a runtime imported from an older server, `build/gx_console account
upgrade-legacy` hashes the existing legacy passwords in place and reports
only counts. The command preserves valid passwords and leaves credentials
requiring a reset for the operator to handle. Backups made before conversion
may still contain legacy values. Rotate credentials that have been shared or
exposed. Keep the runtime, its WAL/SHM companions and backups private.

For a consistent backup:

```sh
python3 tools/snapshot_sqlite.py --source runtime/rapture.sqlite --output /private/backups/gx.sqlite
```

To deliberately update the distributed test seed, omit `--output` or specify `data/rapture.sqlite`. This includes accounts and gameplay state; review the intended source before publishing. Keep seed and manifest together. It does not touch the source runtime and includes committed WAL. Rebuilding a portable alone does not take a fresh snapshot.

Stop both services before restoring/replacing a runtime. Use the backup API for a live copy; never copy only an open main file and discard its WAL. Detailed upgrade/LAN/restore instructions are in the source [operations guide](../../docs/server.md).

## Legacy schema cleanup

Stop both services before upgrading an existing installation. Migration 26
creates a private, verified SQLite backup in a `.gx-before-schema-26-*`
directory beside the runtime before dropping any tables. If the backup
cannot be created and verified, startup stops before deleting data.
Fresh installs use the compacted seed with only the active schema.
Retired catalog exports work against the pre-26 backup using
`gx_export_catalogs --db <backup>` and the original templates. Keep these backups
outside public packages. Compact the upgraded runtime with SQLite
`VACUUM` while both services are stopped if you need to return freed
pages to the filesystem.

## Legacy migration and tests

`tools/export_sqlite.py` and `tools/sqlserver.py` remain migration/reference utilities. Ordinary play needs no SQL Server. Existing external SQL Server volumes are not altered by GX integration. Production backups and restores must preserve a consistent SQLite database, including committed WAL state.

Storage tests and server integration tests use disposable runtime copies. Snapshot tests verify WAL inclusion, 64-bit IDs, Unicode/blob preservation and preservation of previous output on validation failure. Portables validate both hash and row counts. The gameplay/server rules do not change as part of naming and repository integration.

Account creation and password changes now write Argon2id directly through the
built gx_console command. See [modernization](modernization.md) for commands,
credential revocation and TLS provisioning. Existing legacy credentials upgrade
on successful authentication; do not downgrade to plaintext-only binaries.

## Write path and capacity test

Independent inventory, purchase, character and quick-bar writes can share a
durable commit. A failed operation rolls back to its own savepoint. Success is
reported only after the outer commit. A character exit commits the character
and its buff, skill and quest snapshots together. Changes that use a
gameplay transaction still validate their expected inventory and balances
on the writer.

GX migrations 24 and 25 index active write lookups and the legacy
`MAX(RowID)` computation. The latter keeps original identity semantics while
avoiding table scans as players accumulate items and quick slots. Historical
tables and authored game definitions remain compatible with exports.
World handoff tokens are stored as SHA-256 digests. The login encryption
material is short-lived and expired handoffs are cleared in bounded batches.

Reproduce the mixed write probe from the source checkout on the target
filesystem. Set `TMPDIR` to a writable directory on that filesystem so the
probe measures its durability cost. The test always creates a separate
runtime from the approved seed.

```sh
source tools/go-env.sh
GX_STORAGE_LOAD_CLIENTS=1000 GX_STORAGE_LOAD_ROUNDS=32 \
  go test -tags integration ./server/cmd/gx_world -run '^TestPersistentWorldLoad$' -count=1 -v
```

The probe uses encrypted World sockets, scattered placements and independent
accounts. It checks balances, sale logs, item stacks, positions in inventory,
quick bars, logouts, sampled relogs and SQLite integrity after the load.
It provisions tickets directly to isolate the World write workload from
Argon2 and does not characterize simultaneous credential logins or a single
PvP crowd. The live World default is 1,200 connections. The probe raises
loopback per-IP limits only in its own process. Confirm that the operating
system's open-file limit exceeds the connection cap plus logs and SQLite
files. Keep the per-IP limits enabled on a public listener. Monitor writer
wait, transaction latency, WAL growth and storage errors on the deployment
disk during rollout.
