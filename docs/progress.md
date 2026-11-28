# Progress and delivered features

Development snapshot updated November 19, 2026. **Delivered** means the named workflow is implemented and verified within its stated scope.

## Server parity checklist

The [original server parity table](../server/docs/original-parity.md) distinguishes
implemented rules, remaining partial workflows and deliberate GX policies.
Its entries have different sizes and complexity; they are not an overall
project completion percentage.

## Delivery milestones

| Milestone | Status |
|---|---|
| Client and GX Server in one Go repository | Delivered |
| Native Linux and Windows release builds | Delivered |
| Loose portable data beside the executables | Delivered |
| OGG audio, PNG interface and cursors | Delivered |
| External licensed fonts and EP5 skill catalog | Delivered |
| Complete visual/gameplay parity | Ongoing |
| Navigation, collision and cloak details | Ongoing |
| Remaining service/interface variants | Ongoing |

## Delivered functional workflows

| Workflow | Implemented scope | Evidence |
|---|---|---|
| Account entry | Login, server selection, faction confirmation, character creation/selection and entry into the world | [Account integration tests](../internal/app/integration_test.go) |
| Movement and combat input | Movement, target turning/approach, server-confirmed attacks and skill cooldowns | [Native control tests](../internal/app/native_controls_integration_test.go) |
| Skill presentation | Skill book and action bars, OGG cues, caster/target effects and area release synchronization | [Skill visuals](../internal/app/skill_visuals_integration_test.go), [area effects](../internal/app/area_effects_integration_test.go) |
| Inventory and services | Inventory/equipment, merchant purchases, warehouse transfers and blacksmith interactions | [Service status](interface.md), [warehouse tests](../internal/app/warehouse_integration_test.go) |
| Gatekeeper travel | NPC approach, destination selection, confirmation, fees, same-map/cross-map travel and saved position | [Gatekeeper tests](../internal/app/gatekeeper_integration_test.go) |
| Guild workflows | Creation, roster/ranks, notices, Etin and guild storage | [Guild creation](../internal/app/guild_create_integration_test.go), [bank tests](../internal/app/guild_bank_integration_test.go) |
| Portable client | Linux and Windows executables share adjacent data; external fonts and built-in Vorbis decoder | [Portable guide](client-portable.md), [build workflow](../.github/workflows/ci.yml) |
| Portable GX Server | Terminal console controls login/world; development seed creates a separate mutable runtime | [Distribution guide](../server/docs/distribution.md), [native package verification](../server/tools/package.py) |

Chat, party/raid, quests, trade/duels, social windows and the mall are also implemented; their remaining states and presentation gaps are detailed in [interface status](interface.md). A delivered workflow does not mark every feature variant complete.

## Current development

- [GX Studio](gx-studio.md) is a separate editor for the Item Mall, items, skills, monsters, NPCs, quests and client model catalogs.
- [Map rules and progression](wiki/map-rules.md) add level-gated maps, racial respawn fallback, synchronized weather, map curses and a shared experience cap.
- [World drops and item boxes](wiki/recreation-and-loot.md) use SQLite definitions. [Goddess blessing](wiki/goddess-blessing.md) loads a shared bonus catalog for both factions.
- Recorded travel points, party and raid summoning, and account-targeted GM travel extend [travel and administration](wiki/gm-commands.md).

## gx-v0.1.13 release

- [Auction House](auction-house.md) supports a shared market, escrow, bids, buyouts, favorites and claims.
- [Release notes](releases/gx-v0.1.13.md) and [upgrade guide](releases/gx-v0.1.13-upgrade.md).

## gx-v0.1.12 hotfix

- Correct the map-load handshake after cross-map placements; repeated teleports stay responsive.
- Preserve existing v0.1.11 game data and database schema.
- [Release notes](releases/gx-v0.1.12.md) and [upgrade guide](releases/gx-v0.1.12-upgrade.md).

## gx-v0.1.11 additions

- Corrected original combat/effect timing, monster scale/drop rules and group respawns.
- Default modern Lua monster AI, documented policies and real host primitives.
- Atomic economy/trade/quest persistence and modernized ownership, scheduling and bounded networking.
- Argon2id/revocation, remote TLS, versioned SQLite migrations, diagnostics and authenticated shutdown.
- Terminal-only GX Server without GUI dependencies; measured animation/bounds/effect optimizations.
- Native Linux/Windows CI and the expanded common server/protocol Wiki.

See the [release notes](releases/gx-v0.1.11.md) and [upgrade guide](releases/gx-v0.1.11-upgrade.md).

## gx-v0.1.10 additions

- [Single/two-seat mounts](mounts.md), native presentation/audio, authoritative jumps and boarding range.
- Functional native ladders, dungeon click movement and minimaps, normalized creature headers, bounded collision work and cursor/mob optimizations.
- Persistent obelisk/guard/boss/portal lifecycle and goddess blessing; all 31 PvP ranks, native titles and rewards.
- Physical ranged projectiles, skill combos and shared recovery; four innate ITM glows and original feedback.
- Empty Status Mode, live armor element, native Z/X counts and [reduced fall damage](fall-damage.md).

## Remaining work

| Area | Remaining scope |
|---|---|
| Original parity | Full visual comparisons and rare gameplay/service states |
| Navigation | Complete collision/navigation behavior and remaining actions |
| Cloaks | Body/ground contact and map wind |
| Interface | Remaining quest markers, speech bubbles, remaining target indicators and service windows |
| Options | Shortcut editing and remaining controls; resolution selection is implemented |
| Server rules | Remaining partial workflows in the [parity table](../server/docs/original-parity.md) |
| Persistence | SQLite is the primary production database; maintain consistent backups and validated restores |

## Validation boundaries

Native Linux and Windows CI builds and ordinary tests run without private game assets. Server integration, relocated startup and controller isolation are tested natively. Local graphics and N2 tests exercise compatible real data; Windows client graphics have additionally been checked under Wine. CI success is not full gameplay validation on every GPU or native Windows desktop.

Linux remains the primary development platform. Go and OpenGL 4.5 remain fixed project choices. See [contribution direction](../CONTRIBUTING.md), [supported formats](data-formats.md) and [changes](../CHANGELOG.md).
