#!/usr/bin/env python3
"""Build self-contained server archives with the versioned database and game data."""
import argparse
from contextlib import closing
import hashlib
import json
import os
import platform as host_platform
import re
import sys
from pathlib import Path
import shutil
import socket
import sqlite3
import subprocess
import tarfile
import tempfile
import time
import zipfile

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT.parent))
from tools.fetch_server_geometry import validate as validate_geometry
from tools.release_geometry import verify_sources

PROGRAMS = ("gx_login", "gx_world", "gx_console")
GUI_PROGRAM = "gx_server_control"
GAME_DATA = (
    "GuildHouse.ini", "Obelisk.ini", "ItemEnchant.ini", "MobAIMessages.ini",
    "KillStatus.ini", "WorldDrops.ini", "ItemBoxes.ini", "Map.ini",
    "NpcQuest.SData", "TeleportRunes.ini", "Notices.ini", "SafeZones.INI",
)
GEOMETRY_EXTENSIONS = ("gxnav", "gxcol", "gxphysics", "manifest.json")


def checksum(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def check_navigation(root, *, require_pin=False):
    catalog, manifests = validate_geometry(root / "data/geometry", require_pin=require_pin)
    if require_pin:
        verify_sources(catalog, manifests, root / "data/maps")


def check_seed(root, *, require_pin=False):
    check_navigation(root, require_pin=require_pin)
    seed = root / "data/rapture.sqlite"
    manifest = json.loads(seed.with_suffix(".manifest.json").read_text())
    if checksum(seed) != manifest["sha256"]:
        raise ValueError("SQLite seed and manifest disagree; run tools/snapshot_sqlite.py")
    with closing(sqlite3.connect(seed.as_uri() + "?mode=ro", uri=True)) as database:
        if database.execute("PRAGMA integrity_check").fetchall() != [("ok",)]:
            raise ValueError("SQLite seed integrity check failed")
        for table, expected in manifest["tables"].items():
            name = '"' + table.replace('"', '""') + '"'
            if database.execute("SELECT COUNT(*) FROM " + name).fetchone()[0] != expected:
                raise ValueError("SQLite seed row-count mismatch: " + table)
    if not list((root / "data/maps").glob("*.svmap")):
        raise ValueError("server maps are missing")
    for name in GAME_DATA:
        if not (root / "data/game" / name).is_file():
            raise ValueError("required server data is missing: " + name)


def verify_native(package):
    # The binaries discover this relocated copy from an unrelated working directory.
    with tempfile.TemporaryDirectory(prefix="gx-server-package-check-") as temporary:
        temporary = Path(temporary)
        copied = temporary / "GX Server portable copy"
        shutil.copytree(package, copied)
        environment = {key: value for key, value in os.environ.items() if not key.startswith("RAPTURE_")}
        reservations = []
        for _ in range(2):
            sock = socket.socket()
            sock.bind(("127.0.0.1", 0))
            reservations.append(sock)
        ports = [sock.getsockname()[1] for sock in reservations]
        for sock in reservations:
            sock.close()
        environment.update(RAPTURE_LOGIN_LISTEN=f"127.0.0.1:{ports[0]}", RAPTURE_WORLD_LISTEN=f"127.0.0.1:{ports[1]}")
        before = checksum(copied / "data/rapture.sqlite")
        processes = []
        try:
            suffix = ".exe" if sys.platform == "win32" else ""
            for name in PROGRAMS[:2]:
                log = (temporary / (name + ".log")).open("w")
                process = subprocess.Popen([str(copied / "build" / (name + suffix))], cwd=temporary,
                                           env=environment, stdout=log, stderr=subprocess.STDOUT)
                log.close()
                processes.append(process)
            for port in ports:
                deadline = time.monotonic() + 30
                while True:
                    if any(process.poll() is not None for process in processes):
                        details = "\n".join(path.read_text()[-2000:] for path in temporary.glob("*.log"))
                        raise RuntimeError("portable server exited:\n" + details)
                    with socket.socket() as sock:
                        sock.settimeout(.2)
                        if sock.connect_ex(("127.0.0.1", port)) == 0:
                            break
                    if time.monotonic() > deadline:
                        raise TimeoutError("portable server startup timed out")
                    time.sleep(.05)
            if not (copied / "runtime/rapture.sqlite").is_file():
                raise ValueError("runtime copy was not created inside the package")
            if checksum(copied / "data/rapture.sqlite") != before:
                raise ValueError("server modified its distributed seed")
            with closing(sqlite3.connect(copied / "runtime/rapture.sqlite")) as database:
                if database.execute("PRAGMA integrity_check").fetchone()[0] != "ok":
                    raise ValueError("runtime integrity check failed")
            print(f"{sys.platform} verified: relocated folder, unrelated cwd, isolated ports, unchanged seed", flush=True)
        finally:
            for process in reversed(processes):
                if process.poll() is None:
                    process.terminate()
                    try:
                        process.wait(timeout=10)
                    except subprocess.TimeoutExpired:
                        process.kill()
                        process.wait()
        verify_controller(copied, temporary, environment, ports)


def verify_controller(package, temporary, environment, ports):
    """Exercise the real console and ensure a second installation survives."""
    suffix = ".exe" if sys.platform == "win32" else ""
    foreign = temporary / "Another GX Server copy"
    shutil.copytree(package, foreign)
    with socket.socket() as reservation:
        reservation.bind(("127.0.0.1", 0))
        other_port = reservation.getsockname()[1]
    other_env = dict(environment, RAPTURE_LOGIN_LISTEN=f"127.0.0.1:{other_port}")
    with (temporary / "foreign.log").open("w") as log:
        other = subprocess.Popen([str(foreign / "build" / ("gx_login" + suffix))],
                                 cwd=temporary, env=other_env, stdout=log, stderr=subprocess.STDOUT)
    console = None

    def listening(port):
        with socket.socket() as sock:
            sock.settimeout(.2)
            return sock.connect_ex(("127.0.0.1", port)) == 0

    try:
        deadline = time.monotonic() + 30
        while not listening(other_port):
            if other.poll() is not None or time.monotonic() > deadline:
                raise RuntimeError("second installation did not start")
            time.sleep(.1)
        with (temporary / "console.log").open("w") as log:
            console = subprocess.Popen([str(package / "build" / ("gx_console" + suffix))],
                                       cwd=temporary, env=environment, stdin=subprocess.PIPE,
                                       stdout=log, stderr=subprocess.STDOUT, text=True)
        console.stdin.write("start server\n")
        console.stdin.flush()
        deadline = time.monotonic() + 60
        while not all(listening(port) for port in ports):
            if console.poll() is not None or time.monotonic() > deadline:
                raise RuntimeError("console failed to start login/world")
            time.sleep(.1)
        console.stdin.write("shutdown server\n")
        console.stdin.close()
        console.wait(timeout=60)
        if console.returncode != 0 or any(listening(port) for port in ports):
            raise RuntimeError("console failed to stop its own installation")
        if other.poll() is not None or not listening(other_port):
            raise RuntimeError("controller stopped another portable installation")
        print("Console verified: start/stop and second-installation isolation", flush=True)
    finally:
        if console is not None and console.poll() is None:
            # Cleanup is scoped to this copied package, never by image name.
            if console.stdin and not console.stdin.closed:
                console.stdin.write("shutdown server\n")
                console.stdin.close()
            try:
                console.wait(timeout=60)
            except subprocess.TimeoutExpired:
                console.kill()
                console.wait()
        if other.poll() is None:
            other.terminate()
            other.wait(timeout=10)


def copy_runtime_data(source, target):
    """Stage server inputs explicitly, keeping workstation exports out of archives."""
    target.mkdir()
    for name in ("rapture.sqlite", "rapture.manifest.json"):
        shutil.copy2(source / name, target / name)
    (target / "game").mkdir()
    for name in GAME_DATA:
        shutil.copy2(source / "game" / name, target / "game" / name)
    portal_overrides = source / "game/PortalBounds.json"
    if portal_overrides.is_file():
        shutil.copy2(portal_overrides, target / "game")
    (target / "maps").mkdir()
    for path in (source / "maps").iterdir():
        if path.is_file() and re.fullmatch(r"[0-9]+\.svmap", path.name, re.IGNORECASE):
            shutil.copy2(path, target / "maps" / path.name)
    geometry = source / "geometry"
    (target / "geometry").mkdir()
    catalog = json.loads((geometry / "catalog.json").read_text())
    shutil.copy2(geometry / "catalog.json", target / "geometry" / "catalog.json")
    for map_id in catalog["maps"]:
        if type(map_id) is not int or not 0 <= map_id <= 65535:
            raise ValueError("invalid map ID in server navigation catalog")
        for extension in GEOMETRY_EXTENSIONS:
            name = f"{map_id}.{extension}"
            shutil.copy2(geometry / name, target / "geometry" / name)


def build_package(platform, output, go, *, require_pin, gui=False):
    system, architecture = platform.split("-", 1)
    suffix = ".exe" if system == "windows" else ""
    name = "gx-server-" + platform
    with tempfile.TemporaryDirectory(prefix=".package-", dir=output) as temporary:
        root = Path(temporary) / name
        (root / "build").mkdir(parents=True)
        environment = dict(os.environ, CGO_ENABLED="0", GOOS=system, GOARCH=architecture, GOTOOLCHAIN="local")
        for program in PROGRAMS:
            subprocess.run([go, "build", "-trimpath", "-ldflags=-s -w", "-o", str(root / "build" / (program + suffix)),
                            "./cmd/" + program], cwd=ROOT, env=environment, check=True)
        if gui:
            flags = "-s -w -H windowsgui" if system == "windows" else "-s -w"
            subprocess.run([go, "build", "-trimpath", "-ldflags=" + flags, "-o",
                            str(root / "build" / (GUI_PROGRAM + suffix)), "./cmd/" + GUI_PROGRAM],
                           cwd=ROOT, env=dict(environment, CGO_ENABLED="1"), check=True)
        copy_runtime_data(ROOT / "data", root / "data")
        # Editable policies accompany the same embedded default catalog.
        ai = root / "ai"
        ai.mkdir()
        shutil.copy2(ROOT / "ai/catalog.lua", ai / "catalog.lua")
        for asset_name in ("core", "definitions"):
            shutil.copytree(ROOT / "ai" / asset_name, ai / asset_name,
                            ignore=lambda directory, names: [n for n in names if not n.endswith(".lua")])
        shutil.copy2(ROOT / "LICENSE", root / "LICENSE")
        shutil.copytree(ROOT / "licenses", root / "licenses")
        (root / "tools").mkdir()
        for tool in ("sqlite.py", "runtime_lock.py", "snapshot_sqlite.py", "enable_test_gms.sql"):
            shutil.copy2(ROOT / "tools" / tool, root / "tools" / tool)
        for tool in ("update_mount_content.py", "update_recreation_content.py"):
            shutil.copy2(ROOT.parent / "tools" / tool, root / "tools" / tool)
        shutil.copytree(ROOT.parent / "resources/vehicle", root / "resources/vehicle")
        if system == "windows":
            (root / "Start GX Server.bat").write_bytes(b'@echo off\r\ncd /d "%~dp0"\r\n"%~dp0build\\gx_console.exe"\r\n')
        else:
            launcher = root / "Start GX Server.sh"
            launcher.write_text('#!/bin/sh\nset -eu\ncd "$(dirname "$0")"\nexec ./build/gx_console\n')
            launcher.chmod(0o755)
        check_seed(root, require_pin=require_pin)
        files = {str(path.relative_to(root)).replace(os.sep, "/"): checksum(path)
                 for path in sorted(root.rglob("*")) if path.is_file()}
        (root / "package.manifest.json").write_text(json.dumps({"platform": platform, "files": files}, indent=2) + "\n")
        if (system == "linux" and sys.platform.startswith("linux") or system == "windows" and sys.platform == "win32") and host_platform.machine().lower() in ("x86_64", "amd64"):
            verify_native(root)
        if system == "windows":
            staged = Path(temporary) / (name + ".zip")
            with zipfile.ZipFile(staged, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
                for path in sorted(root.rglob("*")):
                    if path.is_file():
                        archive.write(path, name + "/" + path.relative_to(root).as_posix())
        else:
            staged = Path(temporary) / (name + ".tar.gz")
            with tarfile.open(staged, "w:gz", compresslevel=9) as archive:
                archive.add(root, arcname=name)
        destination = output / staged.name
        os.replace(staged, destination)
        print("Ready:", destination, flush=True)
        return destination


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--platform", action="append", choices=("linux-amd64", "windows-amd64"))
    parser.add_argument("--output", type=Path, default=ROOT / "out/portable")
    parser.add_argument("--custom-geometry", action="store_true",
                        help="Package a custom generation for a private server")
    parser.add_argument("--gui", action="store_true",
                        help="Include GX Server Control (requires the target's graphics build prerequisites)")
    args = parser.parse_args()
    go = shutil.which("go")
    if go is None:
        candidates = sorted((Path.home() / ".local/opt").glob("go-*/bin/go"))
        go = str(candidates[-1]) if candidates else None
    if go is None:
        parser.error("Go 1.27 or later is needed to build the packages")
    require_pin = not args.custom_geometry
    check_seed(ROOT, require_pin=require_pin)
    output = args.output.expanduser().resolve()
    output.mkdir(parents=True, exist_ok=True)
    archives = [build_package(platform, output, go, require_pin=require_pin, gui=args.gui)
                for platform in (args.platform or ("linux-amd64", "windows-amd64"))]
    staged = output / ".SHA256SUMS.next"
    staged.write_text("".join(checksum(path) + "  " + path.name + "\n" for path in archives))
    os.replace(staged, output / "SHA256SUMS.txt")


if __name__ == "__main__":
    main()
