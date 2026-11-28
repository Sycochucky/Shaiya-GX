#!/usr/bin/env python3
"""Package an already built GX client with notices, with external fonts, without game artwork."""
import argparse
import hashlib
from pathlib import Path
import shutil
import tarfile
import zipfile
from tempfile import TemporaryDirectory
from font_pack import FONT_PACK, remove_readmes, validate
from gx_ini import FILTER_TEMPLATE

ROOT = Path(__file__).resolve().parents[1]


def verify_archive(archive, stage):
    expected = {}
    for path in stage.rglob('*'):
        if path.is_file():
            with path.open('rb') as f:
                expected[path.relative_to(stage.parent).as_posix()] = hashlib.file_digest(f, 'sha256').digest()
    actual = {}
    if archive.suffix == '.zip':
        with zipfile.ZipFile(archive) as package:
            for name in package.namelist():
                with package.open(name) as f:
                    actual[name] = hashlib.file_digest(f, 'sha256').digest()
    else:
        with tarfile.open(archive, 'r:gz') as package:
            for entry in package:
                if entry.isfile():
                    with package.extractfile(entry) as f:
                        actual[entry.name] = hashlib.file_digest(f, 'sha256').digest()
    if actual != expected:
        raise ValueError('generated client archive failed verification')


def package_client(platform, output):
    validate()
    output.mkdir(parents=True, exist_ok=True)
    name = 'gx-client-' + platform
    with TemporaryDirectory(prefix='.client-package-', dir=output) as temporary:
        stage = Path(temporary) / name
        stage.mkdir()
        exe = 'Shaiya GX.exe' if platform.startswith('windows') else 'Shaiya GX'
        shutil.copy2(ROOT/'build'/exe, stage/exe)
        fonts = stage/'data/fonts'
        shutil.copytree(FONT_PACK, fonts)
        (stage/'data/Filter.INI').write_text(FILTER_TEMPLATE, encoding='utf-8', newline='\n')
        shutil.copy2(ROOT/'resources/client/Tip.INI', stage/'data')
        remove_readmes(fonts)
        validate(fonts)
        shutil.copytree(ROOT/'resources/vehicle', stage/'data/Vehicle')
        for file in ['LICENSE', 'NOTICE']:
            shutil.copy2(ROOT/file, stage/file)
        licenses = stage/'licenses'
        shutil.copytree(ROOT/'licenses', licenses)
        if platform.startswith('windows'):
            archive = stage.parent/(name+'.zip')
            with zipfile.ZipFile(archive, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as z:
                for file in sorted(stage.rglob('*')):
                    if file.is_file():
                        z.write(file, file.relative_to(stage.parent))
        else:
            archive = stage.parent/(name+'.tar.gz')
            with tarfile.open(archive, 'w:gz') as package:
                package.add(stage, arcname=name)
        verify_archive(archive, stage)
        target = output / archive.name
        archive.replace(target)
        return target


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--platform', choices=['linux-amd64', 'windows-amd64'], required=True)
    parser.add_argument('--output', type=Path, default=ROOT/'out/client')
    args = parser.parse_args()
    print(package_client(args.platform, args.output.resolve()))


if __name__ == '__main__':
    main()
