# Instructions for coding agents

## Repository and scope

This is Shaiya GX: client at the root, integrated **GX Server** in `server/`. One Git repository and root Go module; no nested server Git repository, submodule or `go.mod`. Rapture is the former server name. Do not recreate it as a separate project. Public docs are English. Local ignored Spanish continuity/research notes may carry newer user context; follow the current user's explicit instructions.

Preserve existing uncommitted work. Do not use subagents or delegate unless the user explicitly requests it. The user permanently authorizes a descriptive commit and push for each completed change. Keep commits scoped, validate the affected behavior and publish every completed change to the repository. Releases and repository deletion still require explicit authorization. A migration request that explicitly authorizes publication/deletion applies to that migration, not unrelated future changes.

## Contribution direction

Build a functional foundation faithful to original Shaiya, not a collection of individual custom-server preferences. Welcome bugs, behavior-preserving refactors, missing original features, parity corrections, reverse-engineering evidence, measured optimizations and aligned suggestions. Do not add unrelated features, unusual graphical overhauls or personal gameplay rules.

Keep Go and the current OpenGL 4.5 graphics API. Linux is the primary development platform and must remain so; Windows is a supported build/runtime target whose compatibility must be preserved. See `CONTRIBUTING.md` for contributor requirements.

## Boundaries

Keep game rules and persistence authoritative on the server. Preserve protocol framing, opcodes and binary model/world layouts. Client owns presentation and movement prediction. The intentional GX GM extension and account role are documented in `docs/gm.md`. Repository integration and rebranding do not authorize unrelated rule/schema changes.

The playable client is `Shaiya GX` (`Shaiya GX.exe` on Windows), with loose `data/` beside it. Runtime uses the directory loader, Ogg Vorbis sounds and PNG interface images; SAH/SAF and WAV remain import tools only. Preserve `_dds.png` atlas distinctions, authored PNG mipmaps and legacy filename aliases; see `docs/data-formats.md`. Do not restore old archive packaging. By explicit user request, fonts are external resources under `data/fonts/`, with a coverage manifest and lazy Unicode/colour-emoji fallbacks. Do not embed game fonts or depend on OS font installations. Preserve the licenses and unmodified font metadata in the external resource pack; retain the source font license notices.

**GX Server** runs headlessly. The terminal `server/cmd/gx_console` and standalone desktop `server/cmd/gx_server_control` share local control of `gx_login` and `gx_world`. Keep desktop dependencies out of the headless commands. Keep `RAPTURE_*` environment keys, `rapture.sqlite` and `Rapture_LoginLog` compatible unless explicitly migrating them with tests and documentation.

SQLite is the primary production database. Keep storage behind `server/internal/db` and document SQLite operations, backups and deployment accurately. Do not describe it as a temporary backend or promise a replacement.

## GX Studio

GX Studio is a standalone desktop command in `cmd/gx_studio` with editing and
recovery logic under `internal/studio`. Every editable property must show its
readable name and the exact native field and/or SQLite column it changes.
Preserve unmapped data, and route catalog writes through coordinated backups,
verification and recovery. Keep the Studio GUI dependency out of GX Server
commands. Use disposable file and database copies for verification.

## Local safety and fixtures

Check running clients/services before modifying an installation or moving its mutable database. The user permanently authorizes stopping the GX client and server, including force termination when needed to apply or validate changes, without asking. Match exact GX executable paths, preserve personal data/configuration and restart services when appropriate. Unrelated processes remain outside this authorization. Use disposable DB copies, isolated ports and only processes started by your tests. An existing runtime must never be overwritten by packaging or startup. Windows process control must match executable paths, not stop every process with the same filename.

`server/data/` is explicitly distributed: its test DB intentionally includes the author's current accounts, characters and inventory. Refresh only when requested, using SQLite backup and manifest validation. Never snapshot a different personal DB automatically. Keep mutable runtime, `.local`, logs, packet dumps, private backups and credentials out of Git.

The redistributable external font pack is versioned in `resources/fonts/`, with pinned hashes, LGPL/OFL notices and per-font copyrights. Native client packages include it under `data/fonts/`.

Client game archives/assets and original executables are not distributable source. Do not add them or downloaded assets to the repo. Retain source copyrights and licenses, including server/LICENSE. Other local projects are reference-only unless the user authorizes changes.

## Local iteration and patch notes

Work locally on Linux. Commit and push every completed change with an appropriate descriptive message, as permanently requested by the user. Do not publish releases or build Windows unless explicitly requested; previous release authorization does not carry forward. Maintain cumulative local notes in `docs/patch-notes-local.md`; prepare consolidated public patch notes only when requested. Linux builds and relevant local tests remain part of normal development.

## Working speed and cleanup

Prioritize fast delivery, brief Spanish replies and autonomous routine decisions;
ask only for indispensable information or required authorization. Private current
directives and continuity live in `~/Documents/Shaiya-GX/`; local ignored links
`docs/directivas.md` and `docs/continuar.md` provide access. Do not copy notes to
the Desktop. Historical notes may contain superseded instructions.

Use `tools/check.sh <affected packages>` during iteration; omit packages for all
unit checks. It runs vet once and reuses Go's test cache. `tools/go-env.sh` selects
the local Go 1.27.4 installation for build/check scripts. Build only affected
commands when possible. Keep full integration/N2 for relevant behavior or release
validation; do not repeat passed suites without new evidence. Preserve useful
build/module/test caches; remove abandoned test sessions and obsolete temporary
files selectively. Never discard uncommitted work. GX client/server restarts and force termination are authorized when needed; preserve runtime data and unrelated services.

## Full validation reference

From the GX root:

```sh
go vet ./...
go test ./...
go test -tags integration ./server/... -count=1
python3 -m unittest discover -s server/tools -p 'test_*.py'
python3 tools/check_publication.py --history
```

Use Go 1.27.4. Build Linux with `tools/build.sh`, Windows with `tools/build.ps1`; graphics require CGO and platform prerequisites. The private N2 probe uses local assets and a disposable server DB. Server portables: `python3 server/tools/package.py --output out/portable`. They must include the terminal console and login/world services, validate the seed and pass native relocated startup. Add `--gui` to include GX Server Control. Release assets belong to GX releases, never source-history binary blobs.

Run checks appropriate to changed behavior and report actual evidence. Windows cross-compilation or Wine is not native Windows validation. A CI build without game assets cannot establish full client visual parity. Update user-facing docs and local continuity when behavior/workflow changes. Avoid unnecessary architecture layers and keep rendering optimizations measured and visually comparable.
