# Map rules, weather and progression

GX World loads map rules from `Map.ini` and experience thresholds from SQLite at startup. Edit map rules and birth positions in `data/game/Map.ini`. See [server operations](server-operations.md) for backups and restarts.

## Map access and respawn

`Map.ini` uses `[SET_ZONE_N]` sections for map type, level limits,
faction respawns, instance schedules, weather and map curses. Shared
defaults belong in `[DEFAULT_SET_ZONE]`. The four racial birth positions
are stored in `[BASE]`. GX World discovers maps from their sections.
The optional default section supplies keys omitted by individual maps.

| Map | Inclusive level range |
| --- | --- |
| 18 | 1–15 |
| 30 | 20–30 |

Zero leaves that side of the level range open. The server checks destination
access during travel and rechecks saved locations on login. A GM can bypass
the range. A player who exceeds the range by leveling is transferred to their
racial birth position.

Respawn uses the faction's configured destination or the nearest eligible
SVMAP respawn group. If neither provides an eligible destination, the character
returns to their racial birth position. Instance schedules remain part of the
map row, while [obelisk rules](obelisks.md) retain their own configuration.

## Portal volumes

GX World includes 248 original trigger volumes matched to the SVMAP routes.
While a character moves, World detects those portals using the original
boxes. Other routes keep the SVMAP point trigger. In Cloron's Lair, C2 → C1
uses the original trigger volume and C2 → C3 lands in the upper entry corridor.
Custom installations can replace the built-in catalog with
`data/game/PortalBounds.json`.

## Safe zones

GX World enforces `data/game/SafeZones.INI` at startup. Each named section
sets `Map`, `ProtectedFaction` and four `Vertices` in X, Z perimeter order.
Either winding works. The strict interior applies at every height and in
instances of the configured map. Restart GX World after editing the catalog.

- `Light` and `Fury` protect members of that faction from player attacks.
  A protected player can attack the opposing faction, subject to the target's
  own protection.
- `Both` protects both factions and blocks offensive actions from inside.
  Players can still heal, buff and use other non-offensive skills.
- Caster-centered offensive skills such as Frozen Ice, as well as traps and
  hostile fields, cannot be cast inside any sanctuary. The server checks at
  cast start and release.
- Area attacks, traps, reflections and damage from player debuffs respect
  each target's protection. Normal monster aggression and environmental
  damage still apply.

The default catalog preserves ten combat perimeters on maps 45, 46 and 47.
The original EP5.4 reader (Game.exe `0x439270`) and position check
(`0x438650`) distinguish these from the other tactical records. The shipped
combat records share one rule code and name Light or Fury in their authored
descriptions. GX uses those names to assign faction protection. Refused
basic attacks and skills use native SysMsg IDs 316 and 216 respectively.

## Character progression and revival

Every character follows one progression. Base HP, MP and SP use a 1.5
multiplier before race and attribute bonuses. Each level grants nine
allocatable stat points and five skill points in addition to the automatic
class stat. Existing characters receive the difference in per-level awards
once. Previously allocated attributes, learned skills and claimed PvP stages
are preserved.

Any character can revive at their bind point or choose to revive at an eligible
party leader. Revival at the party leader takes 10 seconds. The automatic bind
timer still applies. Death never deletes a character or ends the session. Quests
grant their authored EXP and use their level, class, faction and other active
eligibility rules. Skills use their level, class, faction and skill-point
requirements.

## Guild founding and roles

Create a guild at the guild manager. A solo character can found one
immediately. A party or raid must have at least two founders, with everyone
in the same map instance and faction. Every founder in the group must
accept. Founders can be any level. They must be free to
join and choose an available guild name.

The founder pays the fee once the guild is created. Set
`CreateGuildCost` under `[House]` in `GuildHouse.ini`. The default
is 1,000,000 gold. Leaving a guild carries no waiting period before
applying again, including to the same guild. Normal approval still applies.

The founder becomes rank 1 (master). Other founders and new members start at
rank 9. Ranks 1 and 2 can invite, process applications, manage lower ranks
and upgrade house NPCs, subject to the house requirements. A guild can
have up to six rank 2 members. The master can create or dismiss rank 2
members, buy the house, edit the notice and initiate disbanding. The
founding form also accepts the initial notice. In an unlocked guild bank,
ranks 1–8 can deposit in empty slots and ranks 1–2 can withdraw, rearrange
or swap items. Ranks 3–9 are ordinary members.

The guild's ranking does not gate house purchases or active house services.
The house has its own merchants, gatekeeper, blacksmith and guild warehouse.
Their rates and upkeep are configured in `GuildHouse.ini`. Running out of
etin at settlement suspends the house until upkeep can be paid. A GM-founded
guild receives an active house immediately and is exempt from upkeep.

## Guild Ranking Battle

While a Guild Ranking Battle is open, each guild uses its own instance of the
battle map. Guild members can leave and reenter through the battle entrance
until closing time, including after disconnecting. An empty battle instance
keeps its map state. Monster respawns and drop expiration still follow their
normal rules.

The guild's score is the sum of points awarded by monster kills in its own
instance. Eligible monsters have `Money1 = 0` and `Money2 > 0`, with `Money2`
setting the point award, up to 65,535 per kill. Entries and duels award zero points. When the battle
closes, the server sends remaining players to their bind point and removes the
empty instances. The standings are published 30 minutes after closing.
A GM who leads a guild can use its instance outside the schedule. The guild
receives eligible points when that GM leaves.

## Weather and map curses

| Column | Meaning |
| --- | --- |
| `WeatherState` | 0 clear, 1 rain, 2 snow |
| `WeatherFrequency` | Minutes between event starts, counted from GX World startup. Zero selects 60 minutes. |
| `MapCurse1`–`MapCurse3` | Zero disables the slot. A skill ID applies its supported level-one effect on entry, persists through death and dispels, and is removed when leaving. |
| `MapModifier1`–`MapModifier3` | Reserved for future server rules. The shipped values are zero. |

The initial catalog schedules weather on ten maps. Each event lasts ten
minutes. GX World sends a map-scoped snapshot on entry and when weather
changes. Compatible clients render PS107 precipitation resources, a gray rain
fog or snow fog, the storm sky and scrolling clouds.

A GM can preview rain with `/weather 1` and snow with `/weather 2` for ten
minutes in their own session. `/weather 0` returns that session to scheduled
weather immediately. See the [GM command list](gm-commands.md).

## Experience curve

`PS_GameDefs_ExpDefs` has one cumulative `Exp` threshold per `Level`, shared by
all characters. The rows start at level 1 and are consecutive. Each threshold
is the cumulative experience needed to reach the following level. The highest
row sets the server cap. The initial cap is **60**, reached at **295,000,000
EXP**. Characters at the cap receive zero experience. A reward that reaches the
cap credits only the amount needed to get there.

Migration 19 takes the highest threshold from the former mode rows at each
level and rounds in bands starting at 100,000 EXP. Existing levels, progress
within a level and unspent points are retained. Restart GX World after editing
the curve.

### Party and raid experience

Monster EXP is calculated for each recipient using their level. A solo
kill grants 120% of that amount. In a normal party the total roster size sets
the factor. Two members use 100%, then three through seven use 73%, 64.5%,
56%, 47.5% and 39%. The result is truncated before applying the level share.
Relative to the highest level in the roster, gaps of 0–7, 8–15, 16–31, 32–47
and 48 or more use 100%, 80%, 60%, 40% and 20% respectively.

Each recipient must be alive, in the same instance and within 100 horizontal
units of the player credited with the kill. A raid of any size uses the
lowest native factor, 1%, before its level share. Raids receive no Perfect
Party bonus. PvP kills award kill credit and grant zero EXP.

A seven-member normal party qualifies for **Perfect Party** regardless of
class composition. The server grants an extra amount equal to the recipient's
eligible EXP, including quest rewards, when the base award alone falls short
of the next level. The original `CUser::AddExpFromUser` (`0x458240`) selects
this branch. GX processes any level reached by the bonus immediately and
clips both amounts to the shared cap. The original `0x0207` packet carries
normal and bonus EXP separately. The client uses SysMsg IDs 50 and 66 and adds
both amounts to the EXP bar.
