# Skill and attack cadence: original client versus GX

Static evidence from `Game.exe` EP5.4 plus measured clip lengths from the local
data. No gameplay change is made by this note.

## What the original client does

- Pressing a skill shortcut runs `0x4deda0`. After checking the character state
  (not dead/sitting/mounted, `+0xdc` not 5-8) and the single busy flag
  `DAT_021fd8d0`, it sends the skill (`0x598c60`) **immediately**. There is no
  interval table, cooldown timer or send throttle in this path.
- `DAT_021fd8d0` is set when the action is sent and cleared by `0x40ff40`, which
  `0x418030` (the character's action-finished routine) calls when the action
  animation ends for the own character (`0x4338f0`/`0x431da0` clear it on the
  matching server events). The next press is therefore limited by the action
  animation, not by the server's attack-speed table.
- Only basic attacks change the animation rate with speed (`0x41b300`, 15-45 fps
  for tier minus weapon base plus 5). Every other action, skills included, plays
  at 30 fps.
- Basic attacks are sent once; the chase state resends (250 ms throttle) only
  when the target left reach.

## Clip lengths (Human, 30 fps unless noted)

| Action | Frames | 30 fps | 45 fps (max speed) |
| --- | ---: | ---: | ---: |
| Sword/Spear/… basic attacks | 26-30 | 0.87-1.0 s | 0.58-0.67 s |
| Most skill clips (`skill001`..`011`) | 30 | 1.0 s | - |
| `skill003` | 6 | 0.2 s | - |
| Magic ready / attack | 12 / 28 | 0.4 / 0.93 s | - |
| Buff ready / attack | 14 / 34 | 0.47 / 1.13 s | - |

## What GX does

- Client (`combat_cadence.go`) and server (`combat_cadence.go`) both gate every
  swing and skill on one shared interval from the EP4.5 table: 1000 ms at the
  maximum index. The client measures it from its own send time, the server from
  its execution time.
- The server loop wakes every 100 ms (`loopTick`), so the auto-attack swing and
  a held skill run on the first tick after the deadline: +0..100 ms per action,
  and a request that arrives a few milliseconds early waits for the next tick.
  Real cadence at maximum speed is about 1.05-1.15 s, not 1.0 s.
- Clips shorter than the interval (`skill003`, ready/buff clips, basic attacks at
  45 fps) are held to 1.0 s although the original client would accept the next
  press as soon as the clip ends.

## Options

1. Remove GX's own overhead: send up to ~150 ms before the deadline (the server
   already keeps the latest early request) and run the scheduler at 10-20 ms so
   the interval is exactly the table value.
2. Derive the interval from the clip instead of a fixed table (animation-bound,
   like the original client), keeping the table as the slow-speed ceiling.
   Needs an explicit decision because the EP5.4 server rule cannot be verified
   from the client.

## Implemented (permanent, confirmed in play 2026-11-21)

- `skilltiming.IntervalFor` caps the table by the basic-attack clip (30 frames at
  the weapon-relative animation rate), shared by client and server, so the shared
  action interval at the maximum index with a base-5 weapon is 667 ms instead of
  1000 ms (neutral 1000 ms, slowest 2000 ms). `GX_ANIMATION_CADENCE=off` in both
  processes restores the pure table.
- Skills may be sent up to 150 ms before the deadline (the server keeps the
  latest early request) and the session loop wakes at the exact deadline of a
  held skill, spin combo or auto-attack swing instead of the next 100 ms tick.
- The server's generated table stores one dedicated instant animation duration
  per `SkillID`. The selected skill definition still supplies `ReadyTime` for
  casting. Magic Arrow's instant rank uses the clip floor, while its timed
  ranks finish through the cast path and the shared action cadence.
