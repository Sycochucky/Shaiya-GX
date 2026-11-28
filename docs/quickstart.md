# First run: Windows and Linux

GX consists of a graphical client and GX Server in the same repository. They can run on one computer or on separate LAN computers. Begin locally; the default addresses are loopback.

## Download a portable server

Open [GX releases](https://github.com/Spelunkern/Shaiya-GX/releases/latest). Download the archive for your 64-bit OS and `SHA256SUMS.txt`:

- Windows: `gx-server-windows-amd64.zip` (Explorer → Extract All).
- Linux: `gx-server-linux-amd64.tar.gz` (`tar -xzf gx-server-linux-amd64.tar.gz`).

Choose a writable folder in your user directory, not Program Files or a protected system directory. Keep `build`, `data`, `docs` and the launchers together. No Go, Python, Docker or database server is needed for ordinary server use. The headless services and terminal console need no desktop graphics libraries. See [runtime requirements](building.md#portable-runtime-requirements).

### Verify a download

Windows PowerShell:

```powershell
Get-FileHash .\gx-server-windows-amd64.zip -Algorithm SHA256
Get-Content .\SHA256SUMS.txt
```

Compare the full displayed hash to the line for that exact filename (case does not matter). Linux:

```sh
sha256sum --ignore-missing -c SHA256SUMS.txt
```

Each extracted package also contains `package.manifest.json` with per-file hashes. It excludes itself and runtime files created after startup.

## Start and stop

If your package includes GX Server Control, open `build/gx_server_control[.exe]`
and choose **Start All**. Wait for both services to show **Ready**.

For terminal use, run `Start GX Server.bat` or `./Start\ GX\ Server.sh`. Enter:

```text
start server
shutdown server
```

Closing the console leaves the detached processes running. Stop explicitly before moving the folder, replacing binaries or resetting the database. Logs are in `logs/gx_login.log` and `logs/gx_world.log` inside the server folder.

The first start copies `data/rapture.sqlite` to `runtime/rapture.sqlite`. Future starts reuse the runtime file. The shipped seed contains the author's requested current **test** accounts, characters and inventory. A fresh extraction starts from that snapshot; an existing runtime is never silently overwritten.

**SQLite is the primary production database. No separate database server is required; preserve the runtime database and use consistent backups when upgrading.**

## Accounts

Use a known development account from the included seed, or create an account
with the terminal console. From the server folder run:

```sh
./build/gx_console account create player
```

On Windows use `build\gx_console.exe account create player`. The console prompts
for the password without terminal echo. Account names accept 1–18 UTF-8 bytes;
passwords accept 1–128 UTF-8 bytes. The account is not overwritten if it exists.
Use `gx_console account password player` to change an existing password; this
revokes its active sessions. Python 3 is optional for the accompanying database
administration tools, not required for ordinary account creation.

## Run the client

Download the [complete gx-v0.1.20 portable](https://drive.google.com/file/d/1ah4NBSVTW53U07MwnIsw-oHYjDE5QNqy/view?usp=sharing) from Drive and extract it. Start `Server/build/gx_console` (Linux) or `Server/build/gx_console.exe` (Windows), enter `start server`, then open `Client/Shaiya GX` or `Client/Shaiya GX.exe`. Keep `Client/data/` beside the client executables. To update an existing installation using the GitHub binaries, follow the [binary update instructions](client-portable.md#updating-existing-client-data).
The game discovers its data automatically; no build tools or data conversion are needed.

See [portable client requirements and updates](client-portable.md). GitHub releases provide executable updates without data. Developers can instead [build from source](building.md) and [import their compatible archives](data-formats.md); keep game data outside Git. Do not use the original `Game.exe` to run GX.

The following examples are for a source checkout with data in a separate installation:

Windows PowerShell, from the GX checkout:

```powershell
$env:SHAIYA_DATA = 'C:\Games\Shaiya GX\data'
& '.\build\Shaiya GX.exe' -login 127.0.0.1:30800
```

Linux:

```sh
SHAIYA_DATA=/path/to/ShaiyaGX/data \
  "./build/Shaiya GX" -login 127.0.0.1:30800
```

Sign in, select the server, choose/create your character and enter the world. GX Server decides outcomes and saves progress. The client folder contains `Shaiya GX` (Linux) or `Shaiya GX.exe` (Windows) alongside `data/`. The executable discovers that adjacent folder automatically. Keep the packaged fonts in `data/fonts/`. Start GX Server separately and run the Shaiya GX executable directly. Public GX server downloads do not include those client assets.

## If startup fails

- No login/world ON: read the end of the server logs. Check that `build/` and `data/` are still together and the folder is writable.
- Address already in use: another instance is using 30800 or 30810. Reuse it or stop it through its own control app; do not kill unrelated game sessions.
- Client cannot create a graphics context: install your GPU vendor's driver with OpenGL 4.5 support. A working server does not imply working client graphics.
- Login succeeds but world fails: check the advertised world address, world port and firewall. See [LAN configuration](server.md#lan-connections).
- Changes appear missing: confirm which `runtime/rapture.sqlite` is being used. Extracting a new folder creates a different independent runtime. Preserve your old runtime when updating.
