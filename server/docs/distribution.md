# GX Server portable distribution

Download portable Linux/Windows builds from [GX releases](https://github.com/Spelunkern/Shaiya-GX/releases/latest). Both include login, world, the **GX Server** terminal console, server maps/configuration and the requested development database with current test accounts, characters and inventory.

SQLite is the primary production database. Packages include the embedded database engine and require no separate database service.

## Build from the GX checkout

From the GX root, with Go 1.27.4 and Python 3 installed:

```sh
python3 server/tools/package.py --output out/portable
```

From `server/`, the equivalent is `python3 tools/package.py`. Use `--platform linux-amd64` or `--platform windows-amd64` for one target. All server programs use CGO=0; no C compiler or graphics libraries are required. See the source [build guide](../../docs/building.md) for prerequisites.

Output:

- `gx-server-linux-amd64.tar.gz`
- `gx-server-windows-amd64.zip`
- `SHA256SUMS.txt`

Binaries are attached to GX releases, rather than committed into source history. Tagged release automation reproduces the packages on native runners. Archive manifests list every included file's SHA-256. Packaging validates seed integrity, manifest and table counts; it does not silently refresh the seed from personal runtime state.

The package stages the SQLite seed and its manifest, the server game inputs,
numbered maps in `data/maps/`, and geometry listed in its catalog. The packager
checks the pinned release manifests and each corresponding walk mask. Use
`--custom-geometry` for a private build with locally generated geometry. Server
item definitions and stack limits come from SQLite. When preparing a complete
portable, the packager synchronizes the staged client `Item.SData` stack fields
with the included database.

## Run a package

Keep the directory structure intact in a writable folder. If supplied, open
`build/gx_server_control[.exe]` for the desktop dashboard. For terminal control,
open `build/gx_console[.exe]` or use the `Start GX Server` launcher and enter
`start server` or `shutdown server`. Closing either controller leaves services
running. First start copies the seed to `runtime/rapture.sqlite`; existing runtimes
are preserved.

The headless server and console need no Go, Python, Docker, SQL Server or separate SQLite library. Python 3 is optional for administration tools in `tools/`.

Default endpoints are 127.0.0.1:30800 and :30810. Use `RAPTURE_LOGIN_LISTEN`, `RAPTURE_WORLD_LISTEN` and `RAPTURE_PUBLIC_IP` for other addresses. See the [server operation guide](../../docs/server.md).

## Seed updates and validation

From the GX root:

```sh
python3 server/tools/snapshot_sqlite.py --source /path/to/test/runtime.sqlite
python3 server/tools/package.py --output out/portable
```

This explicitly publishes the chosen development database, including its accounts and inventory. The snapshot uses SQLite backup, includes committed WAL, validates integrity and updates `data/rapture.manifest.json`. Only snapshot data intentionally approved for distribution. The runtime source is not modified. Never include `.local/`, logs, backups or `runtime/` in an archive.

The native packager starts a private relocated copy in a path with spaces and an unrelated working directory, checks both ports, verifies runtime integrity and unchanged seed, then stops only its own processes. CI runs this on native Windows and Linux. Windows cross-builds and Wine checks are additional evidence, not substitutes for native Windows CI or client GPU testing.
