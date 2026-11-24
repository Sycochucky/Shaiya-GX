# Operating GX Server

GX Server is part of GX under `server/`. In a source checkout, server paths in this guide are relative to `server/`; in a portable archive they are relative to the extracted top-level folder. The former standalone Rapture repository is retired after migration.

## GX Server Control

Open `build/gx_server_control[.exe]` inside the server folder. Login and World
show their readiness, listening endpoints and available actions. The World
status provides the in-game player count. Start All starts missing services.
Stop and Restart request authenticated shutdown and wait for the process to
exit. Keep matching versions of all server executables in the same folder.

Live Logs follows the existing rotating Login and World JSON logs. Filter by
service or severity, pause auto-scroll, and select an entry for its full
message. Closing the GUI leaves both services running.

Open **Settings** to edit reward rates. Save writes the three settings to
`.local/runtime.env`, preserving other entries. Restart World to apply the
changes. Inherited environment overrides are displayed and cannot be edited
from the GUI.

| Setting | Runtime key | Default and valid values |
| --- | --- | --- |
| Monster EXP | `RAPTURE_MOB_EXP_PERCENT` | 100%. Whole percentages from 0 to 10,000, applied to solo, party, raid and offline monster kill EXP. |
| Quest EXP | `RAPTURE_QUEST_EXP_PERCENT` | 100%. Whole percentages from 0 to 10,000, applied to quest completion EXP. |
| PvP rank kills | `RAPTURE_PVP_KILL_MULTIPLIER` | 1×. Whole multipliers from 1 to 100 for eligible rank credit, subject to the existing rank cap. PvP quest and faction blessing credit count actual qualifying kills. |

## Terminal console

Run `build/gx_console[.exe]` in a terminal, or use `Start GX Server.sh` / `.bat`. Enter `status`, `start server`, `shutdown server` or `restart server`. Closing the console leaves login/world detached. The GUI and console serialize operations per installation.

Alternatively run `build/gx_login[.exe]` and `build/gx_world[.exe]` directly and manage those process lifetimes yourself. Each binary finds `data/` relative to its executable when no data/database overrides are provided, so launching from another working directory works. Do not mix binaries from different versions.

The controller checks executable paths, including on Windows, so stopping one installation does not stop a separate copy merely because its process names match. Status checks may take several seconds, especially Windows process enumeration.

## Database and persistence

- Distributed seed: `data/rapture.sqlite`; accompanying hash/table counts: `data/rapture.manifest.json`.
- Mutable state: `runtime/rapture.sqlite`. Created from the seed only when absent.
- Login/world handoff keys: the SQLite `GX_SessionHandoff` table.
- Local overrides: `.local/runtime.env` (KEY=VALUE, not committed).
- Logs: `logs/gx_login.log`, `logs/gx_world.log`.

The distributed database intentionally retains the current development accounts, characters and inventory requested by the author. Treat it as a shared test fixture. **SQLite is the primary production database, chosen for its embedded operation and straightforward deployment.** The gameplay-facing database API remains isolated in `server/internal/db`. See the [persistence audit](../server/docs/persistence-audit.md) for transaction-boundary fixes, regression coverage and the scope of concurrency/recovery checks.

Existing `RAPTURE_*` configuration names and SQLite filenames are retained for compatibility. They do not imply an external Rapture installation is required.

## LAN connections

Defaults listen on `127.0.0.1:30800` (login) and `127.0.0.1:30810` (world). To accept another computer, set the following before starting the controller.

Windows PowerShell (replace the example IP with your server's reachable IPv4 address):

```powershell
$env:RAPTURE_LOGIN_LISTEN = '0.0.0.0:30800'
$env:RAPTURE_WORLD_LISTEN = '0.0.0.0:30810'
$env:RAPTURE_PUBLIC_IP = '192.168.1.20'
.\build\gx_console.exe
```

Linux:

```sh
RAPTURE_LOGIN_LISTEN=0.0.0.0:30800 RAPTURE_WORLD_LISTEN=0.0.0.0:30810 \
RAPTURE_PUBLIC_IP=192.168.1.20 ./build/gx_console
```

Allow inbound TCP 30800/30810 on the server firewall. Connect the client with `-login 192.168.1.20:30800`. If using another world port, also pass `-world-port PORT` to the GX client. `RAPTURE_PUBLIC_IP` is the address advertised to the client for world, so it must be reachable from the client's computer; `0.0.0.0` is a listen wildcard, not an advertised address.

For persistent overrides create `.local/runtime.env` in the server directory. Explicit environment values take priority. `RAPTURE_DATA_DIR`, `RAPTURE_DB_PATH` and `RAPTURE_DB_SEED` select data/runtime/seed locations independently. Relative paths in overrides are resolved against the server folder when using the controllers. Prefer absolute paths for direct headless launches with explicit overrides.

## Mob AI

All mob behavior selects the [modern Lua catalog](../server/ai/README.md) by
`Mobs.AI`. Portable installations include an editable `ai/` beside `data/`;
`RAPTURE_AI_DIR` selects another complete catalog. The same Lua definitions are
embedded for installations without that directory. Restart world after edits.
Invalid selected scripts stop startup; there is no legacy Go AI fallback.

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

Starting a newly extracted package without copying a runtime creates a new world from the included test snapshot. Replacing the seed alone never resets an existing runtime. Future database-engine migrations will need explicit migration instructions; no automatic cross-engine conversion exists today.

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
fresh quote before opening a link/extract confirmation. Enchant attempts execute
immediately; their success percentage is displayed above the action button.

Forge costs use the native ReqIg table without Vg/level discounts or the old
free-cost environment override. GX World loads `ItemEnchant.ini` from `data/game`. The shipped
cap is +20. `MaxLevel` configures both equipment types and
`WeaponMaxLevel` or `DefenseMaxLevel` can override it up to +49.
Restart GX World after editing the file in GX Studio.
The client receives the server bonus tables for its tooltips. See the [feature wiki](wiki/README.md) for exact rules,
item content fields, extraction, recreation and drop persistence.

Operational architecture, resource limits, secure remote transport and local
account management are described in [server modernization](../server/docs/modernization.md).

## Experience curve

GX World loads `PS_GameDefs_ExpDefs` from SQLite at startup. Each consecutive
`Level` has one cumulative `Exp` threshold for reaching the following level.
The highest configured level is the cap. Experience stops at the threshold
used to reach that level. Restart GX World after editing the table.

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

## File-backed game catalogs

`data/game/SafeZones.INI` defines the server's PvP sanctuaries for Light, Fury
or both factions. Keep the file with the server data when upgrading. See the
[map rules guide](wiki/map-rules.md#safe-zones) for the vertex syntax and combat rules.

GX World reads the nine gameplay catalogs from `data/game` at startup.
These files take precedence over the retired SQLite definition tables.
Existing migration histories and rows are retained for upgrades and export.
Use GX Studio for structured edits and verified file backups.

To carry a customized SQLite catalog into a file-based installation, export
from a verified SQLite backup into a new directory. Supply the matching
server or client `NpcQuest.SData` to preserve its text and binary envelope.
The optional historical `Map.ini` preserves its additional keys.

```sh
go run ./server/cmd/gx_export_catalogs \
  -db /backups/rapture.sqlite -out /backups/exported-game \
  -npc-template /current/data/Npc/NpcQuest.SData \
  -mob-translations server/data/game/MobAIMessages.ini
```

Review the output and copy it into the target server's `data/game` folder
before restarting GX World. Keep a copy of any existing server files.
