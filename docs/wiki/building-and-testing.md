# Building, testing and packaging GX

Run commands from the repository root unless noted. Client and server share the root `go.mod` / `go.sum`; `server/` has no separate Go module or Git repository. Go 1.27.4 is the tested toolchain. Use an amd64 toolchain and `go version` to confirm it.

## Linux

Install Go, GCC, `pkg-config`, OpenGL development headers, X11, Wayland/xkbcommon headers and ALSA development libraries. For Debian/Ubuntu:

```sh
sudo apt-get install build-essential pkg-config libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev libasound2-dev
./tools/build.sh
```

The script runs vet and unit tests. It builds `build/Shaiya GX`,
`build/shaiyacli`, `build/assets` and the three
GX Server programs in `server/build/`. On other distributions install the
equivalent packages. The client needs an OpenGL 4.5-capable GPU and driver.
To build the terminal-only server programs with Go and CGO disabled, use
`./server/tools/build.sh`.

## Portable runtime requirements

GX Server's login, world and terminal console are built with CGO=0. Prebuilt
server programs require no desktop, OpenGL, C compiler or separate database
library. Run `build/gx_console[.exe]`, then enter `start server` or
`shutdown server`. Closing the console leaves the services running.

The client needs system graphics/audio libraries and a working OpenGL 4.5 driver.
GX selects rendering paths from driver capabilities using the same policy on
Linux and Windows.
Windows users should install their GPU vendor's driver. No Go toolchain or build
headers are needed to run prebuilt programs. Older released packages may still
contain the removed graphical controller; use their included terminal console.

## Windows (native)

Install the Go 1.27.4 amd64 toolchain and a 64-bit MinGW-w64 GCC (for example MSYS2 UCRT64's `mingw-w64-ucrt-x86_64-gcc`). Add Go and that GCC's `bin` directory to your PowerShell PATH, then confirm:

```powershell
go version
gcc --version
$env:CGO_ENABLED = '1'
powershell -NoProfile -ExecutionPolicy Bypass -File tools/build.ps1
```

Use 64-bit MinGW-w64 GCC for CGO. The script builds `build/Shaiya GX.exe`,
`build/shaiyacli.exe`, `build/assets.exe` and
`server/build/gx_{login,world,console}.exe`. Server programs build with Go and
CGO disabled. The client uses Windows audio. Run the client from a desktop
session with your GPU vendor's OpenGL 4.5 driver.

## Windows cross build from Linux

With `x86_64-w64-mingw32-gcc` installed:

```sh
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -o "build/Shaiya GX.exe" ./cmd/shaiya
python3 server/tools/package.py --platform windows-amd64 --output out/portable
```

All server programs use CGO=0 and can be cross-built without MinGW. The client still requires its CGO toolchain. Cross-compilation proves build compatibility; native Windows execution is checked separately by CI.

## Tests and continuous integration

```sh
go vet ./...
go test ./...
go test -tags integration ./server/... -count=1
python3 -m unittest discover -s server/tools -p 'test_*.py'
python3 tools/check_publication.py --history
```

The publication check inspects staged source and, with `--history`, reachable history. Stage intended files first. It permits the explicitly distributed `server/data/` fixtures, the synthetic audio test tone, while rejecting client assets, local runtime copies, credentials and generated binaries.

`.github/workflows/ci.yml` builds and tests on native Linux and Windows runners, including server integration, portable startup from a relocated directory containing spaces, and console Start/Stop isolation from a second running installation. Portable packages are uploaded as workflow artifacts. Tagged GX releases publish the portable archives and checksums. A green run does not establish full visual parity: graphical tests need licensed local assets and a real OpenGL context.

## Local assets and graphics checks

Linux CI checks OpenGL startup and runs asset-free renderer integration tests
under Xvfb with Mesa software rendering.
Windows CI verifies process startup with `-help`. Run `-gl-check` on a desktop
with an OpenGL 4.5 driver to verify context creation, startup shaders and sprite
output without game data. See [Troubleshooting](troubleshooting.md) for
comparing bindless and ordinary texture bindings.

When reporting a client problem, include `logs/client.log` from beside the
executable. The header records the build, OS, CPU and physical RAM. After
graphics initialization begins, the log records the OpenGL renderer and driver.
Use `-gl-no-bindless` and `-gl-no-draw-parameters` to compare the optional
graphics paths on the same driver and scene. The flags also work with
`-gl-check` for startup diagnostics.
The default log keeps failures and limits recurring warnings. Run with
`-verbose` for detailed graphics initialization, shader names and all recurring
warnings. Add `-gldebug` to capture OpenGL driver messages when the client can
create a context. If the regular log cannot be written, the startup error shows
the fallback path.

Place the converted loose `data/` directory beside the Shaiya GX executable, or set `SHAIYA_DATA` / `-data` to that directory. The game no longer reads SAH/SAF. See [data formats and conversion](data-formats.md). The client loads external fonts from `data/fonts/`. Release packaging installs the versioned `resources/fonts/` pack, including its coverage manifest and licenses; no OS font installation is used.

```sh
SHAIYA_TEST_RENDER=1 SHAIYA_DATA=/path/to/ShaiyaGX/data \
SHAIYA_TEST_ARTIFACTS=/path/to/private/captures \
  go test -tags integration ./internal/app \
  -run 'TestHUDPolishOriginalAssetsGL|TestNativeChatAndLoginLogoGL' -count=1 -v
```

The internal N2 probe exercises end-to-end client/server behavior on Linux with a disposable SQLite database and isolated ports. Public integration tests live in `internal/app` and `server/`.

## Package a portable server

```sh
python3 server/tools/package.py --output out/portable
```

This requires Go and Python 3; the server needs no C compiler or GUI libraries. It emits `gx-server-linux-amd64.tar.gz`, `gx-server-windows-amd64.zip` and `SHA256SUMS.txt`. Every package includes the terminal console, login/world, seed, server data, license notices and optional account/backup tools. The packager validates seed hashes, table counts and integrity before shipping and starts the native target in a disposable relocated directory.

Rebuilding does not take a new snapshot from a personal runtime. Refresh deliberately using [server storage instructions](../server/docs/sqlite.md), then build. Release binaries are attached to the GX repository's releases rather than committed into source history.

## Prepare a private client folder (Linux)

`python3 tools/package.py --source /path/to/converted/data --output /path/to/play-folder` writes only `Shaiya GX` and its adjacent `data/` directory. It installs the versioned external fonts into `data/fonts/`; no launcher or development tools are needed. `--update` replaces the Linux executable atomically and preserves the installed game data and refreshes the external font resources. Start GX Server separately, then run the client directly. Windows uses the same converted data directory and its own `Shaiya GX.exe`; the Linux binary does not run on Windows. Keep source documentation/licenses in the repository and game archives out of Git.

## External typography and UTF-8

The bundled regular face is unmodified Wine Tahoma in `data/fonts/ui.ttf`, with 217 lazy
Noto Unicode/colour-emoji fallback faces indexed by `fonts.json`. The complete
redistributable pack is versioned under `resources/fonts/` and included in both
native client release archives. Preserve its LGPL/OFL license notices and per-font
copyrights. No game fonts are embedded and no OS font installation is required.
Place your local Arial regular and bold files at `data/fonts/arial.ttf` and
`data/fonts/arialbd.ttf`. `SHAIYA_FONT` and `SHAIYA_FONT_BOLD` override these paths. `python3 tools/font_pack.py`
verifies the pack inventory and hashes; `tools/prepare_fonts.py` is an optional
maintainer import step requiring fontTools, not a runtime/build dependency.

The gx-v0.1.4 client/server changes use UTF-8, retaining legacy ANSI decoding
for older resources. Update both GX endpoints together. Original packet sizes
remain fixed: character creation names have 18 bytes, UI account names 18 bytes,
the legacy short-password field 16 bytes and chat payloads at most 127 bytes.
The coordinated GX authentication extension supports passwords up to 128 UTF-8
bytes; account administration and both endpoints must be updated together.
These are byte budgets,
not character counts. Truncation and backspace preserve whole grapheme clusters.
SQLite stores the resulting text without an encoding migration.

The current Drive bundle includes gx-v0.1.11 client/server binaries and complete
client data. For later binary-only updates, update both endpoints from the
matching GitHub release. Prepare each public tag with the
[release geometry rebuild](mob-navigation.md#reproducible-release-resources),
using the client data intended for that build. Tagged releases use the reviewed
notes in `docs/releases/<tag>.md`. Local work does not trigger publication automatically.

## Server physics resources for a source checkout

Before running or packaging the server from a clean checkout, install the pinned
resources with `python3 tools/fetch_server_geometry.py`. The script verifies the
archive and every map manifest, and never replaces an existing generation.
Native Linux/Windows CI runs the same step before server integration tests and
portable packaging. Published server packages include the resources already.

The resource release contains navigation and collision geometry for maps 0–67;
it does not contain original client artwork. For future maps or layout changes,
regenerate the server resources with the map builder in
[server navigation](mob-navigation.md), then ship the new generation with the server.

## Server documentation checks

After changing server packages or packet codecs, run
`python3 tools/generate_server_reference.py` and verify with `--check`.
`tools/publish_wiki.py` exports these pages into the same common repository Wiki;
local Markdown changes alone do not publish it. See [server packages](server-packages.md).
