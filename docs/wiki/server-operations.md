# Operating GX Server

GX Server is part of GX under `server/`. In a source checkout, server paths in this guide are relative to `server/`; in a portable archive they are relative to the extracted top-level folder. The former standalone Rapture repository is retired after migration.

## Desktop and terminal controls

Run `build/gx_server_control[.exe]` for the standalone desktop controller with
service readiness, live logs and [reward-rate settings](../server.md#gx-server-control).
On headless installs, run `build/gx_console[.exe]` or the terminal launcher and
enter `status`, `start server`, `shutdown server` or `restart server`.
Closing either controller leaves Login and World running.

Alternatively run `build/gx_login[.exe]` and `build/gx_world[.exe]` directly and manage those process lifetimes yourself. Each binary finds `data/` relative to its executable when no data/database overrides are provided, so launching from another working directory works. Do not mix binaries from different versions.

The controller checks executable paths, including on Windows, so stopping one installation does not stop a separate copy merely because its process names match. Status checks may take several seconds, especially Windows process enumeration.

## Database and persistence

- Distributed seed: `data/rapture.sqlite`; accompanying hash/table counts: `data/rapture.manifest.json`.
- Mutable state: `runtime/rapture.sqlite`. Created from the seed only when absent.
- Login/world handoffs and active ownership: `GX_SessionHandoff` / `GX_ActiveSessions` in the runtime SQLite database. Old session files are not trusted.
- Local overrides: `.local/runtime.env` (KEY=VALUE, not committed).
- Logs: `logs/gx_login.log`, `logs/gx_world.log`.

The distributed database intentionally retains the current development accounts, characters and inventory requested by the author. Treat it as a shared test fixture. **SQLite is the primary production database, chosen for its embedded operation and straightforward deployment.** The gameplay-facing database API remains isolated in `server/internal/db`. See the [persistence audit](../server/docs/persistence-audit.md) for transaction-boundary fixes, regression coverage and the scope of concurrency/recovery checks.

Existing `RAPTURE_*` configuration names and SQLite filenames are retained for compatibility. They do not imply an external Rapture installation is required.

## LAN connections

Defaults listen on `127.0.0.1:30800` (login) and `127.0.0.1:30810` (world). To accept another computer, set the following before starting the controller.

First [provision TLS certificates and client trust](server-security.md#tls-and-portable-client-trust); remote listeners refuse unconfigured plaintext by default. Keep private keys outside the client bundle. The examples assume the generated certificate covers the configured server identity.

Windows PowerShell (replace the example IP with your server's reachable IPv4 address):

```powershell
$env:RAPTURE_LOGIN_LISTEN = '0.0.0.0:30800'
$env:RAPTURE_WORLD_LISTEN = '0.0.0.0:30810'
$env:RAPTURE_PUBLIC_IP = '192.168.1.20'
$env:RAPTURE_TLS_CERT = 'C:\GXPrivate\tls\server-cert.pem'
$env:RAPTURE_TLS_KEY = 'C:\GXPrivate\tls\server-key.pem'
.\build\gx_console.exe
```

Linux:

```sh
RAPTURE_LOGIN_LISTEN=0.0.0.0:30800 RAPTURE_WORLD_LISTEN=0.0.0.0:30810 \
RAPTURE_PUBLIC_IP=192.168.1.20 RAPTURE_TLS_CERT=/private/gx-tls/server-cert.pem \
RAPTURE_TLS_KEY=/private/gx-tls/server-key.pem ./build/gx_console
```

Allow inbound TCP 30800/30810 on the server firewall. Connect the client with `-login 192.168.1.20:30800`. If using another world port, also pass `-world-port PORT` to the GX client. `RAPTURE_PUBLIC_IP` is the address advertised to the client for world, so it must be reachable from the client's computer; `0.0.0.0` is a listen wildcard, not an advertised address.

For persistent overrides create `.local/runtime.env` in the server directory. Explicit environment values take priority. `RAPTURE_DATA_DIR`, `RAPTURE_DB_PATH` and `RAPTURE_DB_SEED` select data/runtime/seed locations independently. Relative paths in overrides are resolved against the server folder when using the controllers. Prefer absolute paths for direct headless launches with explicit overrides.

[Combat DEX thresholds](combat-stats.md#server-configuration) are configurable
in the same environment file and take effect when World starts.

## Content and runtime upgrades

Run matching GX client and server builds. Pending SQLite migrations are
applied on startup to the runtime, with versions recorded in
`GX_SchemaMigrations`. The current source applies migrations through version
23. File-backed catalogs load from `data/game`. The former definition
tables remain available for exporting a customized database. See the
[file-backed catalog guide](../server.md#file-backed-game-catalogs).

Use [GX Studio](gx-studio.md) for supported catalogs and verify the chosen
database path before editing. Restart GX World to reload changed definitions.
See [SQLite storage](server-storage.md#transactional-migrations) for the ledger
and [map rules](map-rules.md) for weather and the level cap.

## Back up and restore

For a consistent live backup, use the SQLite backup API through the optional Python tool:

```sh
python3 tools/snapshot_sqlite.py --source runtime/rapture.sqlite --output /private/backups/gx.sqlite
```

On Windows replace `python3` with `py -3` and use your destination path. This preserves committed WAL data, validates integrity and emits a manifest. Backups can contain account credentials and player state; keep personal backups outside Git.

To restore, stop both services, verify they are stopped, replace the runtime database with the validated backup and remove any stale `-wal` / `-shm` companions from the stopped runtime before restarting. Never overwrite an open SQLite database or restore only its main file while its old WAL is still present. Keep the previous stopped runtime until restoration is verified.

## Update an installation

1. Stop login/world using the controller.
2. Back up `runtime/rapture.sqlite` consistently and retain `.local/runtime.env`.
3. Extract the new release to a new folder and copy your stopped runtime into its `runtime/` directory. Preserve your override settings. Do not replace the new seed with your runtime unless deliberately preparing a shared test fixture.
4. Start the new version and verify login, character selection and expected progress. Keep the previous folder until verified.

Starting a newly extracted package without copying a runtime creates a new world from the included test snapshot. Replacing the seed alone never resets an existing runtime. Runtime schema updates use additive checksummed migrations; keep a validated backup and matching binaries. SQLite remains the primary production engine.

## Renaming and migration

Rapture's source is now `server/` in the GX module. Commands and binaries use `gx_*`, and the advertised server name is GX Server. The legacy `Rapture_LoginLog` table and `rapture.sqlite` filenames intentionally remain, preserving compatibility. Existing clients still use the same login/world endpoints.


## Equipment condition

GX equipment has no durability: attacks never wear items, legacy stored quality
never reduces equipment statistics, and repair requests are ignored without
charging gold. Enhancement failure does not create a broken/zero-quality item.
Legacy database and wire quality fields remain compatible; authored quality
metadata still has unrelated uses such as quest-item drop probabilities. No
personal database migration is required. Source changes take effect after the
world server is rebuilt and restarted.


The static Item Mall catalog loads from SQL `PS_GameDefs_ProductList`. The
client presents native `Cash.SData`; keep both catalogs aligned. See
[cash catalog](cash-catalog.md). Balances, transactions and gifts use SQL.


Blacksmith previews use the read-only GX forge quote (`0x7F03`), with request
serial, live inventory slot references and material count. Replies supply the
success rate per million, gold cost and compatible enchantment hints; no gold or
items change during preview. Actual attempts still validate inventory and use
the same rate/cost functions. The client discards stale quotes and requests a
fresh quote before opening a link/extract/enchant confirmation.

Forge costs use the native ReqIg table without Vg/level discounts or the old
free-cost environment override. GX World loads `ItemEnchant.ini` from `data/game`. The shipped
cap is +20. `MaxLevel` configures both equipment types and
`WeaponMaxLevel` or `DefenseMaxLevel` can override it up to +49.
Restart GX World after editing the file in GX Studio.
The client receives the server bonus tables for its tooltips. See the [feature wiki](wiki/README.md) for exact rules,
item content fields, extraction, recreation and drop persistence.

## Monster health

Monsters do not regenerate health passively, regardless of the legacy
`RecoverValue` definition. Returning home preserves their remaining HP. A fresh
spawn/respawn still starts with its defined maximum health. Explicit healing
attacks or skill effects retain their own rules; they are not passive recovery.

GX fall damage is **enabled by default**, with half the previous damage curve.
Set `RAPTURE_FALL_DAMAGE=false` in the environment or `.local/runtime.env`, then
restart the world server to opt out. See [the complete fall rules](../fall-damage.md).

## Modernized runtime and administration

Both controllers use authenticated loopback shutdown and verify process exit.
The world owns a cancellable scheduler and a database ownership lease. JSON logs
rotate; loopback metrics expose storage/queue/tick pressure. Configure local
accounts and passwords using `build/gx_console account create NAME [--admin]`
and `build/gx_console account password NAME`; these write Argon2id directly.

Read [server architecture](server-architecture.md),
[SQLite internals/recovery](server-storage.md) and
[security, limits and diagnostics](server-security.md) before tuning a deployment.
Do not downgrade to plaintext-only password handling or mix login/world versions.

## Global server notices

`data/game/Notices.ini` uses `[Notice_N]` sections with `Message`,
`Frequency` and `FixedHour1` through `FixedHour5`. GX World reads them at
startup. All connected players receive each announcement in the system chat
and the on-screen news panel, with the same sound as an altar notice.

- `Frequency` is an interval in minutes measured from world startup. The
  first send occurs after one interval. Zero leaves the row available for
  manual use.
- Fixed times use `day/hour/minute` in the server's local time zone, with
  Sunday = 0 through Saturday = 6. For example, `1/20/30` is Monday at 20:30.
  Leave unused slots empty. One valid slot switches the row to fixed
  times. Duplicate slots send once, and invalid slots are ignored.
- `/notice message` sends text immediately. `/notice index1` sends the
  configured message with index 1. Manual sends leave the schedule intact.

The starter row sends `GX Automatic Notice - Test` every 60 minutes. Edit
the INI and restart GX World to apply catalog changes.
