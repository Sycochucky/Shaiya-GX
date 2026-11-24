# GX Server

The integrated server component of [Shaiya GX](../README.md), formerly Rapture. This directory is part of the parent GX repository and Go module; it is not a separate project or submodule.

**GX Server** runs headlessly. Open `build/gx_server_control[.exe]` for the
standalone dashboard, live logs and reward-rate settings. In a terminal, use
`build/gx_console[.exe]` and enter `start server` or `shutdown server`. Closing
either controller leaves Login and World running.

From the GX root:

```sh
./server/tools/build.sh
python3 server/tools/package.py --output out/portable
```

Add `--gui` to either Linux command to include GX Server Control. On Windows
use `server/tools/build.ps1 -Gui` for a desktop build. See
[build prerequisites](../docs/building.md). Portable Linux/Windows archives are
published in [GX releases](https://github.com/Spelunkern/Shaiya-GX/releases/latest).

**SQLite is the primary production database, with embedded operation for straightforward deployment and distribution.** The distributed seed retains the author's current test accounts, characters, items and definitions. First startup copies it into an independent mutable runtime; updates preserve existing runtime state.

Read [first run](../docs/quickstart.md), [build instructions](../docs/building.md), [operations](../docs/server.md), [storage](docs/sqlite.md), [distribution](docs/distribution.md) and [parity notes](docs/original-parity.md). Keep existing `RAPTURE_*` environment variables and `rapture.sqlite` filenames for compatibility. Source copyright/license is preserved in [LICENSE](LICENSE).

GM role and the six server-authorized commands are documented in [GM controls](../docs/gm.md).

Ground navigation uses generated `data/geometry/` alongside the original
`data/maps/` files. See [mob navigation](../docs/mob-navigation.md) for the
offline generation step, compatibility and map-content diagnostics.

The [common GX Wiki server reference](https://github.com/Spelunkern/Shaiya-GX/wiki/Server-Architecture)
documents the complete server structure, every package, SQLite transactions,
security/operations and the opcode/codec catalog. Editable sources live in the
parent repository under docs/wiki/ and publish to that same common Wiki.
