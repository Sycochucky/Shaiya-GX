# Goddess blessing

Blessing belongs to a country, independently of an individual character's PvP
rank. Both player kills and captured altars feed the same authoritative world
counter. GX World sends the progress, timer and tooltip bonuses to the client.

## Data and persistence

GX World loads shared `[Tier_N]` sections from `data/game/KillStatus.ini`.
Both factions use the same 30 stages and keep independent kill counters.
GX World requires consecutive tier numbers and increasing thresholds.
Restart GX World after saving changes to `data/game/KillStatus.ini`.

The thresholds for both countries are:

```
150 300 450 600 750 900 1050 1200 1350 1500
1650 1800 1950 2100 2250 2400 2550 2700 2850 3000
3750 4500 5400 6300 7200 8400 10200 11400 12000 12288
```

Every later stage retains the maximum earlier value for each effect type.
For example, at 450 kills both factions have 6% maximum pools and 5% HP /
6% SP and MP resting recovery. Below 150 the benefit set is empty.

`PS_GameData_WorldInfo.GodBless_Light` and `GodBless_Dark` store the country
counts through the existing `ReadGodBless`/`SaveGodBless` storage API
(`usp_Read_GodBless_R`/`usp_Save_GodBless_E` contracts). World startup restores
them; missing WorldInfo is created through the existing calendar path. The
periodic save/announcement interval is ten minutes; maximum activation and
expiry save immediately, and relic debits also save. Individual kill increments
are not separate synchronous DB transactions. Unsaved progress can be lost on
an abrupt process exit.

The active deadline is in memory, not a new DB column. Restoring a saved maximum
starts the reference server's fresh ten-minute hold. Relic ownership and portal
state live separately in Keeps; see [Obelisks](obelisks.md).

## PvP, altars and activation

Each eligible `KillPCTo` adds one country count, with the original party credit
rules. The original gates that addition at the personal kill cap; GX preserves
the gate at its intentional 1,000,000 personal-kill limit. This does not change
the blessing threshold of 12,288 or the character's 31 rank stages. GM
`/set kill` does not create country credit.

An altar capture adds `final threshold / eligible altar count + 1` for the new
owner and debits the previous owner, floored at zero. With the supplied nine
altars this is 1,366 per capture. There is no 50% altar cap. Server debit tier
selection retains the original previous-tier hysteresis; the original client
selects its descriptive benefit list from the received count.

Reaching maximum resets the other country's count immediately, holds the
winner for ten minutes and blocks both countries' scoring during that hold.
The winner resets when the hold expires. Existing players refresh their
server-calculated stats when the active tier changes. Maximum activation also
resets altars to their configured bases and delays their return ten minutes.

## Effects

| Types | Reference server behavior |
| --- | --- |
| 1–3 | Percent of maximum HP, MP, SP |
| 4–9 | Flat STR, REC, INT, WIS, DEX, LUC; normal stat formulas apply |
| 10–12 | Percent of maximum HP, SP, MP added to sitting/standing recovery |
| 13–15 | Percent of maximum HP, SP, MP added to combat recovery |
| 16 | Critical chance in percentage points |
| 17 | Melee accuracy in percentage points |
| 18–19 | Percent of shared physical defense / additional ranged defense percentage |
| 20, 23 | Magic evasion contributions in percentage points |
| 21–22 | Melee/ranged evasion in percentage points |
| 24 | Lapis linking/extraction-all chance modifier in existing forge rules |
| 25 | Death EXP-loss reduction |
| 26–27 | Present in the table and client descriptions; no reader in the reference EP4.5 server |

The EP5 tooltip labels types 20/26/27 as magic defense, repair discount and
consumable cast-time reduction respectively. GX retains those original SysMsg
labels and the verified EP4.5 server behavior. A label alone does not establish
a gameplay effect; no custom repair/cast-time bonus is invented to reconcile
the version difference. The existing selective extraction rules remain as
documented in [Blacksmith](blacksmith.md).

See [combat stats and damage](combat-stats.md) for how these values affect hits.

## Original client presentation

The footer uses `main_bottom_bar8.png` below 1024 pixels and
`main_bottom_bar.png` at 1024 and above, unscaled and centered. Blessing fills
from the right using the original atlas crop, seven pixels high: 336 pixels at
(446,53) in the 800 layout and 428 at (572,49) in the 1024 layout, relative to
the footer's top. The original percentage text belongs to EXP; no extra
blessing percentage label is added.

Normally the bar represents count / the country's final threshold sent by GX World.
During maximum it represents remaining milliseconds / 600,000 and pulses with
the native additive atlas pass. Frame elapsed time advances presentation only;
confirmed packets still own resets and counts.

Hovering shows a translucent ARGB CD1D1F24 panel, using the native 400-pixel wide variant (expanded for longer translations
and clamped to the viewport), with 15-pixel
rows and the original current cumulative type labels, SysMsg 5127–5153 in type
order. It adds no amounts, future benefits or invented title. Below the first
stage it is the original empty panel. Windows covering the footer and modal
controls suppress its tooltip.

A live maximum GX snapshot starts the ten-minute presentation, plays the local
country sequence of `Effect/bless.EFT` on the local character only, plays `ui_etainbless01.ogg` and displays
`interface/worldname/event.png` for five seconds with the original fade and
placement. Assets remain external. An entry GX snapshot restores the count and
remaining time without replaying activation effects. The native count packets remain available.
Each winning-faction client plays the burst on its own character, never remote
rigs or the opposing faction. Repeated maximum notifications do not replay the
burst or extend the countdown. Both sequences have 13 original emitters and run
their full authored lifetime, including the late eruption; the ordinary
five-second skill-impact cutoff and a busy 32-impact budget do not suppress it.

| Packet | Payload |
| --- | --- |
| `020F` | Country U8, count U32, sent on tier transition/debit and periodically |
| `0211` | Country U8, count U32, remaining milliseconds U32 at character selection |
| `7F1B` | Country U8, entry flag U8, count U32, maximum U32, remaining milliseconds U32, 27 cumulative U16 bonuses in original type order. Sent on map entry and blessing changes. |

## Evidence and checks

Client reference SHA256 is
`8e5ef14a2a3bfef15bae61b03d185e551b6c3fd44d09a2861d57428400bb3263`:
loader 00479df0, tier lookup 00479d70, maximum 00479dc0, footer 00484ef0,
tooltip 00484be0, countdown 00484980, broadcast 0055fc90, entry 0055fd30 and
banner 004bf670. Server reference SHA256 is
`0fb7c9f13b5bcf186118f8bfe73ab214226b93e8aa502b8e6fded4a8312bc547`:
loader 004bf4a0, AddKillCount 00413600, SubKillCount 00413750 and KillPCTo
0045aed0. Research was read-only; proprietary executables/assets are not added
to the source repository.

Tests cover migrated tiers and inheritance, malformed input, authoritative
zero snapshots, elapsed-time countdown, localized tooltip ordering, activation
versus reconnect, and a PvP/altar/DB/maximum/expiry lifecycle. Graphics checks
compare the footer against independent native draw constants at partial/full
counts and during maximum, and render both original blessing effect sequences, checking visible output,
late lifetime and eventual expiration without omitted emitters.
Encrypted disposable-world checks cover saved-state entry, altar contribution,
ownership restart and maximum-time restoration without activation replay.
