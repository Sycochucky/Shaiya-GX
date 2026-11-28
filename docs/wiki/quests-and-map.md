# Quests and map markers

Available quest givers and ready turn-in locations are shown across the entire
current map. They do not depend on which NPCs are inside the player's 3D view.
The small minimap still clips to its displayed area; the expanded map can show
locations across the map, subject to its current zoom and pan.

## Lightweight map catalog

On map entry, GX Server sends quest-bearing NPC locations as metadata. It does
not spawn distant NPC entities, load their models, extend visibility, or allow
remote quest interaction. Multiple placements of the same NPC definition remain
separate locations. Patrolling NPCs use their authored first waypoint as a stable
map anchor, rather than streaming distant patrol positions continuously.

The server shares one immutable catalog per map. Normal movement and animation
never rebuild or resend it. Adding quest-bearing NPCs invalidates the catalog;
adding a non-quest NPC does not resend an identical result. A session checks only
the map's revision during its normal loop. No periodic client polling is used.

## When availability changes

The client caches the filtered quest markers and reuses them while drawing.
Availability is rebuilt after changes to:

- The character, map, level, sex, or received map catalog.
- Active quests, counters, completion, abandonment or prerequisites.
- Party and raid membership or composition.
- Inventory readiness or relevant quest-item quantities in carried bags.
- A known quest minimum-duration deadline, once that deadline is reached.

Picking up an unrelated item does not rebuild quest availability. Moving or
splitting a quest-item stack between carried bags also does not rebuild it if
the aggregate quantities are unchanged. Equipment changes only matter if they
change the carried quantity of an item required by a quest. Cached rendering
reuses marker storage and does not scan all NPC quests every frame.

A deadline is scheduled at the exact next relevant minimum-time transition.
There is no repeated eligibility scan while waiting, and an expired deadline
does not keep triggering rebuilds. Missing timing snapshots retain the existing
presentation hint; the server still validates actual turn-in conditions.

Quest eligibility remains a client presentation hint based on quest definitions
and confirmed character/progress/inventory state. Starting and finishing quests
still use the native server requests and the real NPC interaction requirements.
The map marker is not an authorization to accept a quest remotely.

Party positions remain live overlays. Creature objectives mark spawn sites
throughout the current map until the objective is filled, the quest ends or
the player changes maps. The compact minimap shows only the part in view.
The creature marker uses `radar_part.png` with `radar_party.png` as the
fallback for existing data packs.
Quest markers remain individually present with stable paint order, including
nearby overlapping markers.

## GX protocol

`0x7F04` is a map quest-NPC catalog chunk:

| Field | Type |
|---|---|
| Map ID | `u16`, rendered/base map ID |
| Flags | `u8`: begin = 1, end = 2 |
| Record count | `u16` |
| Each record | NPC type `u8`, definition ID `u16`, X `f32`, Z `f32` |

`0x7F2C` uses the same header and flags for monster spawn sites. A record
contains the monster definition ID (`u16`), X (`f32`) and Z (`f32`).
The server sends a new snapshot on map entry and when event spawn sites change.

All numeric fields are little-endian. Chunks stay within native frame limits.
The client assembles a replacement snapshot and publishes it only when the end
flag arrives, so a large catalog does not flicker into view in partial batches.
A delayed catalog from a previous map is ignored. Nonfinite positions, malformed
lengths, invalid flags and unexpected fragments are rejected. The metadata does
not contain interactive entity IDs and never creates 3D entities.

Matching GX client/world builds provide the map-wide catalog. During initial
loading, or with an older world without the extension, the client retains its
previous nearby-only presentation until catalog metadata is available.

## Verification

Regression tests cover distant quest givers without visible NPC entities,
atomic multi-chunk catalogs, map changes, progress and level invalidation,
relevant versus irrelevant inventory changes, and a minimum-time transition
without sleeping. A steady cached marker query performs no allocations in the
benchmark fixture. The encrypted login/world integration checks a real map's
catalog against its much smaller visible NPC set and verifies that movement
neither resends the catalog nor rebuilds the eligibility cache.

## NPC and quest definitions

GX World loads NPCs, shops, gate destinations, quest links and quest details
from `data/game/NpcQuest.SData` at startup. The client uses its own
`Npc/NpcQuest.SData` for presentation. GX Studio saves shared rules to both
files, keeps their texts and encodings separate, and backs up both before
writing. See [GX Studio](gx-studio.md) and [map access rules](map-rules.md).

### Quest properties in EP5 and GX

The audit compared the original EP5 server and client with 1,720 quest rows
from the distributed server catalog and the installed server and client catalogs.
`QuestRow` is the one-based lookup key. All audited `EmbeddedID` values match
their rows. Keep these two values aligned when editing. NPC and item link
tables preserve the direction and order of offers and turn-ins.

| Native fields | Behaviour in GX |
|---|---|
| `MinLevel`, `MaxLevel`, `Faction`, six class flags | The client filters offers and map markers. World validates these fields at acceptance. `Faction` selects eligible character families. |
| `Male`, `Female` | The original client checks these fields for NPC and item offers. GX applies them to offers, markers and World admission. Seven audited quests restrict one sex. |
| `PrevQuest`, `HG`, `VG` | World admits the character after any of the configured quest IDs has been completed successfully. The client uses the same rule to present offers. |
| `CG`, `OG`, `IG` | Optional minimum values of the corresponding character attributes at acceptance. Zero leaves the attribute unrestricted. |
| `RequireParty`, six party class counts | With `RequireParty=1`, the character needs a party or raid of at least two members. Each class field sets its minimum count in that roster, including the character accepting the quest. `RequireParty=0` disables the group requirements. |
| `MinMinutes`, `MaxMinutes` | Minimum elapsed time gates completion, and the maximum triggers failure. World persists timing and counters. The catalogs contain 162 maximum durations and zero minimum durations. |
| `TickReceive` | Any nonzero value makes the quest repeatable. World deletes the active record when it ends so it can be offered again. The 69 authored repeatable quests all use `1`. |
| `TickStart`, `TickKeep` | `TickStart` is the minimum interval in seconds between accepted starts of the same quest. `TickKeep` is a time limit in seconds since acceptance. The earliest of `TickKeep` and `MaxMinutes` expires the quest. Zero disables the corresponding limit. |
| `StartType`, `StartNPCType`, `StartNPCID` | Types 0–2 use a nearby NPC. Types 1 and 2 require the configured NPC. Type 0 can name a specific NPC or use the NPC in the quest chain. |
| `StartItemType`, `StartItemID` | Type 2 consumes one matching carried item on acceptance. Type 3 offers the quest after acquisition and requires possession at acceptance. The item stays in the bag. Type 4 offers the quest when using the item from inventory. |
| Three `StartItem` type, ID and count groups | World plans the initial grants and commits them with acceptance and the type 2 cost. A full inventory rejects the start before publishing quest or item changes. |
| `EndType`, `EndNPCType`, `EndNPCID` | Types 0–3 cover dialogue, inventory, monster and PvP turn-ins. A named NPC is required and must be within 30 units, including when selecting the reward. Eligible item and kill objectives can complete automatically when the NPC type is zero. An end type of 0 with no NPC has no recovered completion route. |
| Three `FarmItem` type, ID and count groups | World counts carried items and consumes the required quantities as part of the reward transaction. Native quest-item stacks are removed in full when handed in. |
| `Mob1ID`, `Mob1Count`, `Mob2ID`, `Mob2Count`, `PVPKills` | World tracks the two monster counters and the PvP kill counter, with native party credit for supported monster kills. |
| `ResultType`, `UserSelect`, each variant's `NeedMobID`, `NeedMobCount`, `NeedItemID`, `NeedItemCount` | World uses the fixed reward, offers a reward choice, or selects a variant from the qualifying objective counts. The variant IDs in the `NeedMobID` and `NeedItemID` columns select objective slots. |
| Each variant's `NeedTime`, `NeedHG`, `NeedVG`, `NeedOG` | Optional cumulative requirements on the reward variant. `NeedTime` sets the minimum seconds since acceptance. The other fields set minimum character HG, signed VG and OG. Zero disables that requirement. World checks them for fixed, automatic and chosen rewards. |
| Each variant's `XP`, `Gold`, `ItemType`, `ItemTypeID` | World commits the reward with the quest and any item costs. The XP value is awarded as authored. |
| Each variant's `NextQuest` | The original client reads this field from the confirmed reward variant. GX offers the next quest when the reward dialog is acknowledged and its NPC context and eligibility still apply. Acceptance is a separate request. The audited catalogs have 757 variant links, all with a turn-in NPC. |
| `QuestName`, `MsgSummary`, `MsgTake`, `MsgComplete`, `MsgResponse1`, `MsgResponse2`, `MsgIncomplete` | GX uses these for quest titles, journal objectives, initial offers, completion messages by variant, and incomplete turn-ins. |
| `MsgWelcome`, `MsgResponse4` | GX displays the welcome text before the offer. The alternate response appears when a quest fails or is abandoned. It also supplies the incomplete turn-in text when `MsgIncomplete` is empty. |

For type 4, the original client sends `0x050A` before `0x0902` when the player
accepts the item offer. World records the last used quest item for five minutes.
Items of types 27 and 28 remain in the inventory. Type 29 consumes one unit on
use, so the following start request can succeed even if that stack is gone.
The owner receives the inventory change and nearby viewers receive `0x050A`.
GX matches both item type and ID. The type 3 fallback `/quest start ID` uses
the same possession check as the inventory offer.

The additional conditions for previously inactive fields are GX rules. The
audited EP5 catalogs use zero for the added timers, reward attributes and
character attributes, and leave `RequireParty` off. For a type 29 item linked
to a quest with an added admission condition, GX consumes the selected item
in the same SQLite transaction as acceptance. This keeps the item if the
start is refused. A normal type 29 item retains the EP5 item-use sequence.

World persists start intervals per character and quest. An interval begins
when acceptance commits and survives success, abandonment and reconnects.
`TickKeep` and `NeedTime` use elapsed wall-clock time, including time offline.
Existing active quests with these fields can recover their start time from
the legacy minute snapshot once. Legacy quest timing uses its original
minute-based snapshot. The native quest packets still report whole minutes.
World checks the current inventory and requirements again at reward
selection. It consumes the chosen variant's item requirement and the base
objective quantity whenever the base requires more.

Original server references are `0x4CA5D0` (definition load), `0x47F080`
(admission), `0x4655E0` (item use), `0x47F500` (type 4 start) and `0x47F8E0`
(inventory objectives). Original client references are `0x461946` and
`0x462220` (NPC offer eligibility), `0x462540` (item links and offers),
`0x43ED13` (confirmation packet order), and `0x52DAD7` / `0x43EC20`
(reward follow-up offer).
The text loader at `0x42BB96` reads all seven dialogue strings after the
numeric fields. The offer and reward dialog at `0x43EB84` reads `MsgTake`,
the selected response or the summary.
