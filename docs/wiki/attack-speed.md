# Attack speed

The shared-clock rules below describe GX. The
[original timing investigation](../spell-fluidity-research.md) distinguishes
native execution deadlines, per-skill reuse and animation replacement, and
records remaining parity gaps. It does not establish faster combo damage.

GX Server derives attack speed from the weapon's native `Items.AttackTime`,
equipped item bonuses, active lapis, learned weapon masteries and applied ability
**16**. The final index is clamped to 0–9 after adding all signed modifiers.
The weapon field is the index itself: subtracting one selects the wrong delay.

| Index | Native status description | Server interval (ms) |
| --- | --- | --- |
| 0 | None / unarmed | 4000 (table fallback; a valid weapon is required to attack) |
| 1 | Extremely slow | 3000 |
| 2 | Very slow | 2750 |
| 3 | Slow | 2500 |
| 4 | A bit slow | 2250 |
| 5 | Normal | 2000 |
| 6 | A bit fast | 1750 |
| 7 | Fast | 1500 |
| 8 | Very fast | 1250 |
| 9 | Extremely fast | 1000 |
| 255 | Cannot attack | Blocked |

PvE basics, PvP basics, player skills and combos use this table and **one shared
server-side action timestamp**. Instant skills reserve it on release; prepared
casts reserve it on start and complete without charging the interval twice.
Extremely Fast has a **1000ms** ordinary recovery. There is no
independent 300ms spell rate. The client buffers only the latest spell intent
during recovery and stops basic auto-repeat so the queued spell gets its turn.
Restart requests, target switches, stop/start floods and client speed packets do
not reset that clock. Removing or expiring a bonus changes the next allowed
swing according to the new derived interval. Stun, sleep, mounting and active
cast preparation prohibit normal attacks. Beginning preparation stops an existing
autoattack; attack speed is published as blocked until preparation ends or is
cancelled. Spell cooldowns and preparation times retain their own skill rules.

Wind Spin → a distinct spin or instant hostile caster-centred area has an
explicit exception: after the native first contact (4/30 s), the new skill can
replace the two-second looping spin and execute early. Own cooldowns and costs
remain intact; same-spin repetition, buffs, basics and prepared spells retain
the ordinary rules. The server keeps at most the latest early eligible intent.
Stinger movement is server-authored and replicated by 0x7F11; ordinary movement
cannot override it. See [skill timing and effects](../skill-targeting-and-effects.md).

## Native item and skill fields

- `Items.AttackTime` is a weapon's base attack index. On an equipped accessory or
  an active lapis it is an additive attack-speed bonus. It is counted once.
- Oblivion Flash Ring (`22/56`) adds one index; Flash Ring (`22/213`) adds six.
  Flash Lapis (`30/64`, `30/65`, `30/66`) add one, two and three respectively.
- `Items.Speed` is the movement-tier bonus, including Oblivion Sonic Amulet
  (`23/57`) and Sonic Lapis. See [Movement speed](movement-speed.md).
- Learned mastery detail **8** adds its `Weaponvalue` only for matching weapon
  families. General passive abilities and active buffs use ability **16** for
  attack and **17** for movement. Debuffs subtract their configured value.
- Equipment kept in carried bags grants no equipment bonus. Native compatible
  targets, equipment requirements and active-lapis selection still apply.
- Delicious Candy (`25/175`) applies native skill `211/1`, adding one attack
  index with normal consumption, cooldown and expiry.
- These fields are overloaded for consumables: `Range` and `AttackTime` can
  identify an item-cast skill and its level. A consumable's skill level is never
  displayed or applied as an equipped attack-speed bonus.

The owner and eligible viewers receive the final indices through native `0x051C`.
View-entry snapshots include them. Equipping, linking, extracting, learning a
mastery, applying/cancelling/expiring effects and reconnecting refresh the derived
state without allowing the client to supply its own rate.

## Stats and tooltips

The original Stats fields at panel coordinates `(92,345)` and `(92,365)` show
attack and movement descriptions from `sysmsg-uni.txt`. They display confirmed
server indices, with an unknown placeholder before confirmation. Attack uses
messages 5271–5281; movement uses the original selected-message switch and 5282
for blocked movement.

Weapon tooltips use native attack descriptions 5192–5201, including active linked
lapis; socket-enhanced attack text is yellow. Additive item/lapis bonuses use the
localized attack-speed prefix and `+N`. Movement bonuses use message 5209 and
`+N`. Tooltips combine native item and active socket fields while preserving the
cached definition and the server's independent calculation. The descriptions
remain available for carried items, shops and the Item Mall.

## Native evidence

Reference `ps_game.exe` SHA-256:
`0fb7c9f13b5bcf186118f8bfe73ab214226b93e8aa502b8e6fded4a8312bc547`.
`InitEquipment` clears attack speed at `0x4545e2`; weapon `ItemEquipmentAdd`
adds `stItemInfo +0x44` directly at `0x454fac`. Accessories add the same field at
`0x454d0a`; gem addition is at `0x456122`. `GetAttackSpeed`/`GetAttackSpeedNow`
(`0x48a460`/`0x48a4b0`) clamp to 0–9, and `g_ConstUserAttackSpeed` at `0x5424e0`
contains the ten millisecond intervals above. Isolated execution verifies 108
weapon/accessory/Flash/Sonic accumulation combinations.

Reference `Game.exe` SHA-256:
`8e5ef14a2a3bfef15bae61b03d185e551b6c3fd44d09a2861d57428400bb3263`.
Direct execution of the status dispatch at `0x506b57` and `0x506da1` confirms
11 attack and 6 movement lookups, including blocked indices. Status field
positions are the native constants at `0x70adb8..0x70adc4`.

Regression coverage includes real seed item/passive/buff definitions, raw index
and cap behavior, cast blocking, all ten cooldowns under restart/target floods,
confirmed-only UI, socket aggregation and original-asset OpenGL captures. Isolated
N2 also measures encrypted PvP swing intervals of 1500/1250/1000 ms while changing
Flash Ring and a native attack buff and flooding stop/start requests.

## Movement and jumping during attacks

A successful basic attack or any skill (including heals, buffs and self casts) commits a **800 ms horizontal
movement pause**, predicted by the client and enforced by the world server.
The weapon's attack interval remains independent and cannot be reset by
stop/start requests. The pause clears horizontal movement credit rather than
banking speed for a later burst; vertical gravity continues normally.

Basic attacks and spells execute after landing. During a jump in place, the
client holds the latest combat order until the landing animation ends. A jump
that travels horizontally discards the order. The client rechecks selection,
reach and skill requirements before sending the action. A new jump interrupts
any attack sequence or spell preparation in progress. World checks the
attacker's physical support. The victim's movement never grants immunity.

Movement input cancels basic auto-repeat. The confirmed swing continues through
its animation and contact cues during the movement pause. Held movement or a
queued terrain destination resumes when the pause expires. The swing may finish
blending into locomotion after movement resumes. Healing, buffs and self casts
use the same movement pause.

## Spell reach and release

The server validates targeted spells when preparation starts. Launch reach remains
`AttackRange + target mob size + 2.3` units in GX, with player targets adding no
mob size. An accepted prepared cast does **not repeat the distance check at
release**: the original server's UpdateAttack state 3 resolves the existing
valid target directly. Moving beyond launch range, or beyond the retired GX
2.5-unit grace, does not itself cancel that cast. Instant spells still use launch
reach. Dead, unavailable, protected or no-longer-hostile targets remain invalid.

Walls and floors can block a targeted spell at launch or release. Attack areas
use exact `ApplyRange` in 3D and visibility from the current release center.
Friendly party skills with `TargetType=4` and whole-raid skills with
`TargetType=8` use horizontal `ApplyRange` from the caster at cast completion.
The native single-target launch helper uses horizontal distance and an
additional movement-dependent allowance.

Movement interrupts preparation only after the server accepts a locomotion
packet with its native non-interrupting 0x80 flag clear. Motion-zero idle/height
refreshes and malformed/rejected updates leave the cast intact. Idle packets
cannot translate the caster horizontally during preparation. A real interruption restores the published attack speed,
so the client does not remain stuck with the casting speed sentinel 255.

See [prepared spells and area effects](../prepared-spells-and-area-effects.md)
for native evidence and visual routing.

An accepted hit is committed at release; subsequently leaving range does not
revoke queued damage or a projectile's visual flight. This does not introduce a
new server projectile travel simulation. The 800 ms offensive movement pause
and existing resource/cooldown rules remain in effect.
