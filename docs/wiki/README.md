# Shaiya GX Wiki

## [Download Shaiya GX — Client + Server + Data](https://drive.google.com/file/d/1ah4NBSVTW53U07MwnIsw-oHYjDE5QNqy/view?usp=sharing)

**The complete gx-v0.1.20 portable for Windows and Linux includes the client, GX Server,
game data, fonts and the development database.** Extract the package, start
`Server/build/gx_console` (or `.exe`) in a terminal, enter `start server`, then open the executable in `Client/`.
[Setup, requirements and updates](client-portable.md).
Current binary release: [GitHub Releases](https://github.com/Spelunkern/Shaiya-GX/releases/latest).

The reference for current GX gameplay features, item data, administration and
operation. GX is a Go client with an integrated authoritative world/login
server, using original EP5 interface resources and protocol foundations.

These pages follow the current development branch. See the [binary update instructions](client-portable.md#updating-existing-client-data) to update and preserve your runtime database.

## Gameplay and items

- [Original client typography](native-typography.md) covers the native fonts, rendering paths and GX metrics.
- [Chat controls and item links](chat-parity.md), [User Options](user-options.md) and [Revolver](revolver.md).
- [Auction House](auction-house.md): listings, bids, buyout, favorites,
  item-family searches, escrow and claims.
- [Teleport Runes](teleport-runes.md): faction destinations, channeling, guild
  house eligibility and SQLite configuration.
- [Group loot](group-loot.md) explains how party and raid items are shared.
- [Monster sizes](monster-sizes.md) covers the original scale codes, target widths and examples.
- [Combat stats and damage](combat-stats.md) covers the live damage, critical,
  accuracy, evasion and higher-level monster rules.
- [Combat, world and interface changes in v0.1.17](../releases/gx-v0.1.17.md).


- [Map rules, weather and progression](map-rules.md): map access and respawn,
  synchronized storms, permanent map curses and the shared level cap.

- [Mounts](mounts.md): single/two-seat riding, native effects/audio, jumping,
  invitations, server boarding range and shared racial bones.ini.
- [Fall damage](fall-damage.md): enabled by default, reduced curve, safe jumps,
  server-owned landings and loss-free environmental death.


- [Goddess blessing](goddess-blessing.md): faction counters, original tier effects,
  PvP/altar contributions, database restoration, footer UI and tooltips.

- [Obelisks and linked bosses](obelisks.md): capture, guards, persistent ownership,
  goddess blessing, map icons, linked portals and boss phases.

- [Attack speed](attack-speed.md): native weapon indices, server swing intervals,
  Flash equipment, mastery/buffs and confirmed Stats/tooltips.

- [Movement speed](movement-speed.md): original on-foot tiers, authoritative Sonic,
  haste and consumables, speed limits and the original running trail.

- [Quests and map markers](quests-and-map.md): map-wide available missions,
  lightweight NPC metadata and event-driven eligibility caching.

- [Stealth and GM protection](stealth-and-gm-protection.md): authoritative
  concealment, monster disengagement, original ghost rendering and immunity.

- [Elements](elements.md): complete EP5 matrix, item/lapis affinity, PvE/PvP,
  monster and spell rules, buff lifecycle, detection and native evidence.

- [Blacksmith](blacksmith.md): native buttons, continuous attempts, exact gold
  table, linking formulas, hammer values, SQLite enchant rates and caps,
  selective extraction, failure protection and original system messages.
- [Item effects](item-effects.md): passive stats, triggered gem skills, recovery,
  healing, cooldown groups, utility items, elements, native field overloading and
  compatible-target tooltips.
- [Recreation and world loot](recreation-and-loot.md): rune requirements, valid
  random rolls, attribute pools, Craftname encoding, creation versus pickup and
  persistence without free rerolls.
- [Item Mall](item-mall.md): native Cash.SData, SQL ProductList, product bundles,
  255 stacks, catalog synchronization and migration tooling.
- [Weapon visual effects](weapon-effects.md): enchant glow tiers, native weapon
  families, particles and server-confirmed local/remote equipment updates.
- [Interface reference](interface.md): implemented interaction, inventory,
  controls and recovered original presentation details.
- [Examples and troubleshooting](troubleshooting.md): worked forge examples,
  rejected items, stale data, retired INI configuration and version differences.

## Administration and content

- [GX Studio](gx-studio.md): standalone client and server catalog editor with backups,
  saved installation paths and exact field labels.

- [GM commands](gm-commands.md): role source, complete command syntax, examples,
  limits, transactional inventory changes, temporary spawns and protocol details.
- [Item data and compiled rules](item-data.md): where each setting lives, what
  requires rebuilding, native target flags, safe content migrations and tests.
- [Server operations](server-operations.md): Terminal operation, LAN setup, seed versus
  mutable database, persistence, updates, backup and recovery.
- [Data formats](data-formats.md): loose resources, native tables, PNG/Ogg,
  external fonts and importing legally obtained original assets.

## Server architecture and complete package reference

The server reference is generated from the current Go source.

- [Server architecture and lifecycle](server-architecture.md): processes, ownership,
  concurrency, simulation timeline, transaction flow and failure semantics.
- [All server packages](server-packages.md): every command, runtime package, Lua AI,
  shared dependency and gameplay subsystem, with implementation contracts.
- [Infrastructure API](server-reference-infrastructure.md),
  [SQLite API](server-reference-storage.md) and [world API](server-reference-world.md):
  complete production file/symbol inventories, private helpers, dependencies and tests.
- [SQLite schema, transactions and recovery](server-storage.md): migration ledger,
  table ownership, atomic operation matrix, WAL, backups and failure injection.
- [Security, TLS, limits and observability](server-security.md): credentials,
  certificate/trust setup, configuration, queue bounds, diagnostics and metrics.
- [Wire protocol and session phases](server-protocol.md): framing, primitive widths,
  handshake sequence, authority and packet examples.
- [Complete opcode catalog](server-opcodes.md): 275 current opcodes organized into
  fifteen families, linked to serializers, decoders, handlers and client reducers.
- [Shared codec appendix](server-codecs.md): exact row types and helper codecs,
  including layouts that cannot be associated precisely by static dispatch.

## Development and distribution

- [Architecture](architecture.md): ownership boundaries, rendering, network
  state, root module and integrated server storage abstraction.
- [Protocol](protocol.md): explicit packet encoding, authoritative decisions,
  inventory acknowledgment and disposable integration testing.
- [Building and testing](building-and-testing.md): Linux/Windows prerequisites,
  native builds, test commands and portable server packaging.
- [Client portable distribution](client-portable.md): package structure,
  external assets and fonts, launch behavior and validation.

The client requests actions and presents confirmed state. The world validates
live inventory, computes chance and cost, consumes resources and persists the
result. A tooltip, quoted probability or gray material preview cannot grant an
item or authorize a mutation.

## Documentation maintenance

Editable Markdown sources live under `docs/wiki/` in the development checkout.
The GitHub Wiki is a separate Git repository, `Shaiya-GX.wiki.git`; changing a
local Markdown file alone does not publish it. `tools/generate_server_reference.py` regenerates the server API/opcode/codec
inventories; `--check` detects stale generated pages. `tools/publish_wiki.py` validates
page links and exports native Wiki page filenames before an explicit push.
Keep rules and examples aligned with executable handlers, not only legacy item
descriptions. Do not publish private runtime files, credentials or original
proprietary assets as documentation attachments.
