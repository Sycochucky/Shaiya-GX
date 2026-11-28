# GX documentation

Start with [first run](quickstart.md), then [building and testing](building.md).

| Guide | Contents |
|---|---|
| [CPU/render measurements](client-measurement.md) | Reproducible scenes, crowd/packet stress, percentiles, GPU waits and CPU profiles |
| [Performance metrics](performance.md) | On-screen FPS, frame/CPU/GPU timings, RAM, estimated VRAM and draw calls |
| [Streaming and LOD](streaming-lod.md) | Sector visibility, CPU loading workers, conservative object LOD and measurements |
| [Progress](progress.md) | Delivered workflows, evidence and remaining work |
| [Changes](../CHANGELOG.md) | Releases, media migration and separately distributed data updates |
| [Portable client](client-portable.md) | Complete Drive download for Linux/Windows, runtime requirements and executable updates |
| [Data formats](data-formats.md) | Loose data layout, PNG and Ogg conversion |
| [First run](quickstart.md) | Windows/Linux downloads, verification, accounts and local play |
| [Build and test](building.md) | Toolchains, native/cross builds, CI, integration and packaging |
| [Server operations](server.md) | Terminal console, LAN, logs, backup and upgrades |
| [GX Studio](gx-studio.md) | Independent desktop editor for Shaiya GX |
| [Map rules and progression](wiki/map-rules.md) | Access, respawn, weather, curses and the experience cap |
| [Server Wiki reference](wiki/server-architecture.md) | Complete server structure, packages, storage, security and protocol catalog |
| [Architecture](architecture.md) | Single Go module, client/server boundaries and state ownership |
| [Interface](interface.md) | Implemented behavior, parameters and remaining work |
| [Auction House](auction-house.md) | Realm-wide NPC market, auctions, claims and SQLite operations |
| [Auction House research](auction-house-research.md) | NPC routing, original market UI and rules, SQLite escrow and validation design |
| [Protocol](protocol.md) | Wire compatibility and server authority |
| [Original spell timing](spell-fluidity-research.md) | Reuse, execution scheduling, the two Wind Spins and Stinger dash evidence |
| [Animation transitions](animation-transitions.md) | Original blend duration, pose sampling and GX differences |
| [Mounts](mounts.md) | Mount casting, movement, two riders, MON bones/effects and content migration |
| [Weapon enchantments](weapon-enchantments.md) | Local/remote equipment presentation and validation |
| [Server storage](../server/docs/sqlite.md) | Primary production SQLite, WAL, backups and recovery |
| [Portable distribution](../server/docs/distribution.md) | Archive contents, verification and publishing |
| [Server parity](../server/docs/original-parity.md) | Recovered rules and known differences |
| [Social service](../server/docs/social.md) | Friends, blocks and party search |
| [Contributing](../CONTRIBUTING.md) | Scope, review, evidence and data policy |
| [Agents](../AGENTS.md) | Automation and repository instructions |

Server data and the explicitly requested test database are distributed fixtures. Client archives and private workstation research/captures remain outside the source tree. Rapture is incorporated as GX Server, not an external project dependency.
