#!/usr/bin/env python3
"""Combine verified native releases into a dual-OS folder, optionally with local data.

Original game resources remain local; this command does not upload them.
"""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import shutil
import tarfile
import tempfile
import zipfile

from cash_path import existing_cash_path
from fetch_server_geometry import read_manifests, validate as validate_geometry
from font_pack import remove_readmes, validate
from gx_ini import FILTER_TEMPLATE
from item_sdata import RETIRED_STACK_CSV, reconcile_stack_limits
from release_geometry import verify_sources, verify_tag

ARCHIVES = ('gx-client-linux-amd64.tar.gz', 'gx-client-windows-amd64.zip',
            'gx-server-linux-amd64.tar.gz', 'gx-server-windows-amd64.zip')


def digest(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def text_content(data):
    if b'\0' in data:
        return None
    for encoding in ('utf-8', 'cp1252'):
        try:
            return data.decode(encoding).replace('\r\n', '\n')
        except UnicodeDecodeError:
            pass
    return None


def merge(source, destination):
    for path in sorted(source.rglob('*')):
        if path.name.casefold() == RETIRED_STACK_CSV:
            continue
        if path.is_symlink():
            raise ValueError('unexpected package symlink')
        if not path.is_file():
            continue
        relative = path.relative_to(source)
        if source.name.startswith('gx-client-') and relative.parts[0] == 'Studio':
            continue
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        if target.exists():
            a, b = path.read_bytes(), target.read_bytes()
            if a == b or text_content(a) is not None and text_content(a) == text_content(b):
                continue
            raise ValueError('native packages disagree: ' + str(relative))
        shutil.copy2(path, target)


def verify_cash_copies(client_data, server_data):
    if not existing_cash_path(client_data).is_file():
        raise ValueError('client Cash.SData catalog is missing')
    if not (server_data / 'rapture.sqlite').is_file():
        raise ValueError('server ProductList database is missing')


def prepare(archives, data, output, *, require_pin=False):
    if output.exists():
        raise ValueError('output already exists; choose a new folder')
    if data is not None and (not data.is_dir() or output.is_relative_to(data)):
        raise ValueError('invalid loose data source/output')
    expected = {}
    for line in (archives / 'SHA256SUMS.txt').read_text().splitlines():
        sha, name = line.split(maxsplit=1)
        expected[name.lstrip('*')] = sha
    for name in ARCHIVES:
        if digest(archives / name) != expected[name]:
            raise ValueError('release checksum mismatch: ' + name)
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='.gx-bundle-', dir=output.parent) as temporary:
        temporary = Path(temporary)
        extracted = temporary / 'native'
        extracted.mkdir()
        for name in ARCHIVES:
            path = archives / name
            if name.endswith('.zip'):
                with zipfile.ZipFile(path) as archive:
                    for member in archive.namelist():
                        relative = PurePosixPath(member)
                        if relative.is_absolute() or '..' in relative.parts:
                            raise ValueError('unsafe release archive path')
                    archive.extractall(extracted)
            else:
                with tarfile.open(path) as archive:
                    archive.extractall(extracted, filter='data')
        stage = temporary / 'bundle'
        client, server = stage / 'Client', stage / 'Server'
        merge(extracted / 'gx-client-linux-amd64', client)
        merge(extracted / 'gx-client-windows-amd64', client)
        if data is not None:
            print('Copying local loose game data...', flush=True)
            if any(path.is_file() and path.name.casefold() == 'tip.ini' for path in data.iterdir()):
                (client / 'data/Tip.INI').unlink(missing_ok=True)
            def exclude_local_data(directory, names):
                excluded = {RETIRED_STACK_CSV}
                if Path(directory).parent == data and Path(directory).name.casefold() == 'fonts':
                    excluded.update({'arial.ttf', 'arialbd.ttf'})
                return [name for name in names if name.casefold() in excluded]
            shutil.copytree(data, client / 'data', dirs_exist_ok=True, ignore=exclude_local_data)
        (client / 'data/Filter.INI').write_text(FILTER_TEMPLATE, encoding='utf-8', newline='\n')
        fonts = client / 'data/fonts'
        remove_readmes(fonts)
        validate(fonts)
        for platform in ('linux-amd64', 'windows-amd64'):
            source = extracted / ('gx-server-' + platform)
            manifest = source / 'package.manifest.json'
            record = json.loads(manifest.read_text())
            for name, sha in record['files'].items():
                if digest(source / name) != sha:
                    raise ValueError('native server manifest mismatch: ' + name)
            manifest.unlink()
            merge(source, server)
        server_data = server / 'data'
        if require_pin or data is not None:
            geometry = server_data / 'geometry'
            catalog, manifests = validate_geometry(geometry) if require_pin else read_manifests(geometry)
            client_data = client / 'data' if data is not None else None
            verify_sources(catalog, manifests, server_data / 'maps', client_data)
        if data is not None:
            verify_cash_copies(client / 'data', server_data)
            changes = reconcile_stack_limits(client / 'data', server_data / 'rapture.sqlite')
            if changes:
                print(f'Synchronized {changes} stack limits in the staged Item.SData.', flush=True)
        for path in stage.rglob('*'):
            if path.is_file() and path.name.casefold().split('.')[0] == 'readme':
                path.unlink()
        # Older native archives carried a documentation directory and a Lua host guide.
        if (server / 'docs').is_dir():
            shutil.rmtree(server / 'docs')
        (server / 'ai/HOST.md').unlink(missing_ok=True)
        files = {p.relative_to(server).as_posix(): digest(p) for p in sorted(server.rglob('*')) if p.is_file()}
        (server / 'package.manifest.json').write_text(json.dumps({'platform': 'linux-amd64+windows-amd64', 'files': files}, indent=2) + '\n')
        stage.replace(output)
    print('Distributable folder ready:', output, flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--archives', type=Path, required=True)
    data_source = parser.add_mutually_exclusive_group(required=True)
    data_source.add_argument('--data', type=Path)
    data_source.add_argument('--without-data', action='store_true', help='Keep client game resources separate; include only packaged fonts')
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--tag', help='Require the public release tag and its pinned geometry')
    args = parser.parse_args()
    try:
        if args.tag is not None:
            verify_tag(args.tag)
        data = args.data.resolve() if args.data is not None else None
        prepare(args.archives.resolve(), data, args.output.resolve(), require_pin=args.tag is not None)
    except (ValueError, OSError) as error:
        parser.exit(1, f'{error}\n')


if __name__ == '__main__':
    main()
