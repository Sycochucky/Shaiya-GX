#!/usr/bin/env python3
"""Apply explicit forge item definitions to stopped, explicitly selected DBs.

Only declared PS_GameDefs_Items fields and native client stack counts are changed.
Backups are required and created using SQLite backup, never by replacing a
personal runtime with the distributed seed.
"""
import argparse
from contextlib import closing
import csv
import hashlib
import json
from pathlib import Path
import sqlite3
from typing import NamedTuple

from item_sdata import item_sdata_path, patch_stack_limits


class ForgePatch(NamedTuple):
    typ: int
    identity: int
    vg: int
    stack: int
    special: int | None = None
    name: str | None = None


def read_patches(path):
    with path.open(newline='', encoding='utf-8') as f:
        return [
            ForgePatch(int(row['type']), int(row['type_id']), int(row['req_vg']), int(row['stack']),
                       int(row['special']) if row.get('special') else None, row.get('name') or None)
            for row in csv.DictReader(f)
        ]


def apply(path, patches):
    if not path.is_file(): raise ValueError('missing database '+str(path))
    with closing(sqlite3.connect(path)) as db:
        with db:
            for patch in patches:
                typ, identity, vg, stack, special, name = ForgePatch(*patch)
                if not 0 <= vg <= 100 or stack not in (0, 255):
                    raise ValueError('invalid forge definition')
                row=db.execute('SELECT Special FROM PS_GameDefs_Items WHERE Type=? AND TypeID=?',(typ,identity)).fetchone()
                if row is None: raise ValueError('missing forge item')
                if special is None:
                    if row[0] not in (36, 58) and typ != 95:
                        raise ValueError('refusing unrelated item')
                elif (typ, identity, special) not in ((100, 191, 58), (100, 192, 36)) or row[0] not in (0, special):
                    raise ValueError('refusing unrelated item conversion')
                db.execute('''UPDATE PS_GameDefs_Items
                    SET ReqVg=?, Count=CASE WHEN ?=0 THEN Count ELSE ? END,
                        Special=COALESCE(?,Special), ItemName=COALESCE(?,ItemName)
                    WHERE Type=? AND TypeID=?''', (vg, stack, stack, special, name, typ, identity))
            if db.execute('PRAGMA integrity_check').fetchone()[0]!='ok': raise ValueError('integrity failure')
        limits=list(db.execute('SELECT Type*1000+TypeID,CASE WHEN Count>1 THEN 255 ELSE 1 END FROM PS_GameDefs_Items ORDER BY Type,TypeID'))
    manifest=path.with_suffix('.manifest.json')
    if manifest.exists():
        doc=json.loads(manifest.read_text());doc['sha256']=hashlib.sha256(path.read_bytes()).hexdigest();manifest.write_text(json.dumps(doc,indent=2)+'\n')
    return limits


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--database',action='append',type=Path,required=True)
    ap.add_argument('--backup-dir',type=Path,required=True)
    ap.add_argument('--patches',type=Path,default=Path(__file__).parent/'data/forge-item-patches.csv')
    ap.add_argument('--client-data',type=Path,required=True)
    args=ap.parse_args();args.backup_dir.mkdir(parents=True,exist_ok=True)
    patches=read_patches(args.patches)
    item = item_sdata_path(args.client_data)
    raw = item.read_bytes()
    patched = patch_stack_limits(raw, {p.typ * 1000 + p.identity: p.stack for p in patches if p.stack})
    (args.backup_dir / 'Item.SData').write_bytes(raw)
    for n,path in enumerate(args.database):
        if not path.is_file():raise ValueError('missing database')
        backup=args.backup_dir/f'database-{n}-before.sqlite'
        if backup.exists():raise ValueError('backup already exists; choose a new backup directory')
        with closing(sqlite3.connect(path)) as source, closing(sqlite3.connect(backup)) as dest:source.backup(dest)
    limits=None
    for path in args.database:
        current=apply(path,patches)
        if limits is not None and current!=limits:raise ValueError('target item definitions differ')
        limits=current
    if patched != raw:
        staged = item.with_suffix('.SData.gx-new')
        staged.write_bytes(patched)
        staged.replace(item)
    print(f'Updated {len(patches)} forge definitions in {len(args.database)} databases; no character rows edited')

if __name__=='__main__':main()
