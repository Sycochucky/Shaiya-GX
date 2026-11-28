# Item data and configuration

## Sources of authority

| Data | Authoritative source | Applying a change |
|---|---|---|
| Enchant cap, weapon/armor rates and stat bonuses | `data/game/ItemEnchant.ini` | Edit the INI and restart GX World |
| Native forge gold table | `lapisiaMoney` in `server/internal/world/crafting.go` | Rebuild and restart world |
| Linking/extraction formulas | `server/internal/world/forge_quote.go` | Rebuild and restart world |
| Item requirements, hammer ReqVg, stack limits | SQL `PS_GameDefs_Items` in the selected server DB | Restart world |
| Item names, descriptions, icons and native flags shown by client | External `data/Item/Item.SData` | Restart client |
| Client stackability matching SQL | `data/Item/Item.SData` | Restart client |
| Mall display / actual products | Client Cash.SData / SQL ProductList | Restart affected component |

The last consecutive level in each SQLite table sets its cap. The shipped
tables cover +1 through +20. See [Blacksmith](blacksmith.md) for the limits,
field units and effects of removing levels.
The old free-cost environment setting remains ignored.

The server quote is read-only. It uses the same functions as execution; the
client checks request serials and inventory snapshots before accepting a reply.
Selecting a gray material hint never grants an item or authorizes an operation.

## Synchronizing the installed item catalog

The installed customized EP5 `Item.SData` is the authoring source for item
definitions. SQL remains authoritative during gameplay. Run
`tools/sync_item_definitions.py` offline to copy its shared native fields to
`PS_GameDefs_Items`, adding missing IDs without removing server-only rows.
It preserves the explicit GX recreation-rune and forge-object exceptions
from the selected SQL baseline, reconciles their client numeric fields, and
keeps client names, descriptions, models and icons. Matching consumable skill
references above 255 retain their full SQL IDs: Item.SData stores only the low
byte of Range. Mount metadata already
present in the installed client is retained.

The current catalog contains **22,101 definitions**, including 56 recreation
runes/aliases and 28 protected forge objects. The synchronization does not
change skills, accounts, characters, inventories, owned item quantities or mall
products. Item synchronization itself does not implement effects. The subsequent lapis
combat implementation supports the 26 Lv6/Lv7 debuff lapis. Their referenced
NPC definitions in `PS_GameDefs_NpcSkills` are aligned with the installed
`Monster/NpcSkill.SData`; player skills use `PS_GameDefs_Skills`. The server uses NPC rank 1 for
monster and lapis effects and keeps the native level 100 in effect packets.

Migration 38 restores 27 known consumable references that older SQLite
runtimes could store as a single byte. It also restores skill 250/2 for the
GM EXP Stone (44/242), granting 50% more monster EXP for 24 hours. The client
expands the matching aliases when it loads Item.SData. Snowball (44/147)
resolves to the targeted attack 193/1 and requires the original targeted-item
cast route, which GX does not yet implement.

```sh
python3 tools/sync_item_definitions.py \
  --item-sdata '/path/to/client/data/Item/Item.SData' \
  --database server/data/rapture.sqlite \
  --database '/path/to/server/data/rapture.sqlite' \
  --database '/path/to/server/runtime/rapture.sqlite' \
  --report work/item-sync/plan.json
```

This is read-only unless `--apply --backup-dir work/item-sync/NEW-BACKUP`
is added. Stop the affected GX installations first. The tool validates all
inputs before applying, creates SQLite backups, checks that every other SQL
table stays identical, and refreshes existing seed manifests.
Client files are replaced atomically to avoid modifying hardlinked portable
copies. Keep backups private; never publish runtime databases or client SData.

## Overloaded native fields

- **ReqIg:** selects the native gold table for lapis and lapisias. For lapis it
  also participates in the linking formula; it does not control lapisia chance.
- **ReqVg on Special 36:** added percentage points for a linking hammer.
- **ReqVg on Special 58:** direct extraction success percentage.
- **ReqVg on lapisias:** no equipment-destruction behavior. Current content sets
  it to zero and describes the safe behavior.
- **Slot on equipment:** total lapis capacity, with six physical storage entries.
- **Slot on a lapis:** capacity consumed by that gem. It is not a socket owned by
  the lapis; lapis/lapisia tooltips do not draw their own socket rows.
- **ReqLuc on equipment:** packed enhancement material ID, type*1000+type ID.
- **ReqWis / Server:** recreation maximum / number of distinct rolled stats.

Native target flags on the lapis/lapisia determine compatible equipment families:
weapon families require nonzero ReqLevel; family16 uses Country; families17–22
use their respective class flag; family23 uses Grow; family24 is accepted;
family25 uses ReqStr. These overloaded fields do not introduce character-level
or character-mode requirements for forge attempts.

## Current hammer values and definition migration

The exact shipped changes, including item keys and matching English descriptions,
are in [`tools/data/forge-item-patches.csv`](../../tools/data/forge-item-patches.csv).
It includes both Type44 and Type100 variants of numbered linking hammers,
common/operator hammers, extraction hammers and lapisias. Lapisias now stack to
255; existing player quantities are not rewritten or automatically merged.

`cmd/itempatch` applies that CSV to a native Item.SData, preserving every unrelated
record byte. It supports encrypted input and writes a plaintext native table
accepted by the client. Use a distinct output file, verify it, then install it
atomically with the client stopped. Original archives and client assets are not
part of the source repository.

`tools/apply_forge_rules.py` applies the specified SQL ReqVg and Count fields plus the names and Special values of the two Cash Mall GM hammers,
creates SQLite backups in an explicit directory, updates a seed manifest when
present and patches the matching stack counts in `Item.SData`. Stop the affected
services first.
It never replaces a personal runtime with the seed. The distributed seed and an
existing personal runtime are separate targets and must be named explicitly.

## Verification

From the repository root with the documented Go toolchain:

```sh
go test ./...
go vet ./...
go test -tags integration ./server/... -count=1
```

The forge regressions cover exact costs, level/Vg independence, additive hammers,
INI cap/rates, safe enchant failures, selective extraction, repeated
consumption, original SysMsg refusals and valid drop rolls surviving pickup and
redrop. An internal probe also exercised the client workflow against isolated processes with a disposable copy of the seed.


Targeted recreation properties are implemented in
`server/internal/world/recreation_runes.go`. ReqVg 101–109 selects individual
random stats; 201–209 selects perfect stats. Type101 IDs1–18 use those properties. IDs19–27 use ReqVg301–309 to remove one recreation stat; ID28 uses ReqVg400 to remove all nine while preserving enchantment and lapis.
The client now retains native Server to determine eligibility for its Max OJ
footer. Keep it and ReqWis aligned with SQL. See [Recreation](recreation-and-loot.md).
