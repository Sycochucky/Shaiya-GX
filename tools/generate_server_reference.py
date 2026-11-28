#!/usr/bin/env python3
"""Generate the common Wiki's server API and protocol reference from Go ASTs.

Run after sourcing tools/go-env.sh. --check fails if committed pages drift.
Only source declarations are inspected: no runtime, assets or database is read.
"""
import argparse
from collections import defaultdict
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[1]
PREFIX = 'github.com/Spelunkern/Shaiya-GX/'
WIKI = ROOT / 'docs/wiki'
FAMILIES = {
    'characters': ('Characters and account selection', {0x01}),
    'actions': ('Character actions and transport', {0x02}),
    'targets': ('Targets, ground objects and world time', {0x03, 0x04}),
    'combat': ('Combat, skills and character resources', {0x05}),
    'mobs': ('Monsters and NPCs', {0x06, 0x0E}),
    'services': ('Shops, storage and blacksmith', {0x07, 0x08, 0xB1}),
    'quests': ('Quest lifecycle', {0x09}),
    'trade': ('Player exchange and personal markets', {0x0A, 0x23}),
    'groups': ('Party, raid and group resources', {0x0B, 0x0C}),
    'guilds': ('Guilds and guild storage', {0x0D}),
    'chat': ('Chat and legacy administration', {0x11, 0xF1, 0xF5, 0xF7}),
    'social': ('Relics, friends and PvP rewards', {0x21, 0x22, 0x24, 0x25}),
    'mall': ('Item Mall and account points', {0x26}),
    'extensions': ('GX protocol extensions', {0x7F}),
    'authentication': ('Login and world handshakes', {0xA1, 0xA2, 0xA3}),
}
PURPOSES = {
 'ai': 'Embedded/default Lua catalog, actor state machines, host API contract, bounded script execution and optional editable catalog loading.',
 'admin': 'Local account creation/password reset and private TLS provisioning; never a public admin endpoint.',
 'config': 'Validated environment policy, optional atomic runtime.env parsing, paths and SQLite DSN.',
 'control': 'Shared desktop and terminal lifecycle, executable-scoped status and shutdown, detached launch, reward settings and log tailing.',
 'controlplane': 'Loopback health/metrics, authenticated readiness and shutdown, optional protected profiling and private PID/executable descriptors.',
 'crypto': 'Legacy login RSA/HMAC, AES-CTR and expanded world XOR compatibility; authenticated remote protection belongs to TLS.',
 'db': 'All authoritative SQLite access, definitions, migrations, guarded economy transactions and crash/session ownership.',
 'filelock': 'OS-released cross-process startup/ownership locks for Linux and Windows.',
 'logx': 'Concurrent structured JSON logs, session tags, bounded file rotation and opt-in debug output.',
 'netx': 'Server TCP/TLS admission, keepalive, bounded frame parsing, rate limiting and ordered bounded output.',
 'packet': 'Server opcode constants and explicit little-endian frame writer; no Go struct layout on the wire.',
 'password': 'Versioned Argon2id hashing/verification, bounded workers, legacy upgrade and normalized failure work.',
 'scheduler': 'Single cancellable serial simulation timeline, deterministic stepping and duration/delay histograms.',
 'session': 'Atomic single-use login-to-world SQLite handoffs, random identities and credential-version validation.',
 'telemetry': 'Bounded registry of atomic counters/gauges and duration histograms in Prometheus text format.',
 'world': 'Authoritative realm, sessions, maps, actors, game rules, protocol adapters and commit-before-publication application actions.',
 'gx_console': 'Terminal start/shutdown controls plus local account and TLS administration commands.',
 'gx_server_control': 'Standalone desktop service controls, live logs and reward-rate settings.',
 'gx_login': 'Login handshake, bounded credential verification, handoff creation, server selection and diagnostics.',
 'gx_world': 'World bootstrap/lease, definition/map loading, owned scheduler, bounded sessions and graceful storage shutdown.',
 'gx_mapbuild': 'Offline collision/navigation generation from locally owned loose assets; not a runtime network service.',
}

def key(record):
    return Path(record['path']).parent.as_posix() + '.' + record['name']

def link(record, label=None):
    return f"[{label or record.get('name') or Path(record['path']).name}](../../{record['path']}#L{record['line']})"

def code(value):
    return '`' + value.replace('|', '&#124;').replace('\n', ' ') + '`'

def unique(records):
    return list({(r['path'], r['line']): r for r in records}.values())

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    records = json.loads(subprocess.check_output(['go', 'run', 'tools/server_reference_ast.go'], cwd=ROOT))
    prod = [r for r in records if not r['path'].endswith('_test.go')]
    funcs = defaultdict(list)
    for r in prod:
        if r['kind'] == 'func':
            funcs[key(r)].append(r)
    constants = {}
    declarations = defaultdict(list)
    for r in prod:
        if r['kind'] != 'const':
            continue
        is_opcode = r.get('signature') == 'Opcode' or (r['path'] == 'internal/authwire/credentials.go' and r['name'] == 'Opcode')
        try:
            value = int(r.get('value', ''), 0)
        except ValueError:
            if is_opcode:
                raise ValueError(f"Cannot resolve opcode constant {key(r)}: {r.get('value', '')}")
            continue
        constants[key(r)] = value
        if is_opcode:
            declarations[value].append(r)
    def evaluate(expr):
        expr = expr.removeprefix(PREFIX)
        if expr in constants:
            return constants[expr]
        try:
            return int(expr, 0)
        except ValueError:
            # Opcode casts contain an already-qualified constant, not an offset.
            match = re.fullmatch(r'(?:packet|proto)\.Opcode\((?:proto|packet|authwire)\.(\w+)\)', expr)
            if match:
                for package in ('internal/proto', 'server/internal/packet', 'internal/authwire'):
                    if package+'.'+match[1] in constants:
                        return constants[package+'.'+match[1]]
            return None
    writers, receivers, client_receivers, codecs = (defaultdict(list) for _ in range(4))
    known = set(declarations)
    constructors = {}
    for r in prod:
        if r['kind'] == 'func' and r.get('ops'):
            for expr in r['ops']:
                op = evaluate(expr)
                if op is not None and 0x100 <= op <= 0xFFFF:
                    known.add(op)
                    (writers if r['path'].startswith('server/') else codecs)[op].append(r)
                else:
                    # Generic builder: resolve op from its call-site arguments.
                    name = expr.rsplit('.', 1)[-1]
                    parameters = re.search(r'\b'+re.escape(r['name'])+r'\((.*?)\)', r['signature'])
                    if parameters:
                        names = re.findall(r'\b(\w+)\s+(?:packet\.)?Opcode\b', parameters[1])
                        if name in names:
                            # The generic packet builders in this source take op first.
                            if re.match(r'\s*'+re.escape(name)+r'\b', parameters[1]):
                                constructors[key(r)] = r
    for r in prod:
        if r['kind'] == 'call' and r['value'].removeprefix(PREFIX) in constructors and r.get('ops'):
            op = evaluate(r['ops'][0])
            if op is not None and 0x100 <= op <= 0xFFFF:
                known.add(op)
                writers[op].append(constructors[r['value'].removeprefix(PREFIX)])
        if r['kind'] != 'case':
            continue
        name, path, tag = r['name'], r['path'], r.get('signature', '').lower()
        if 'op' not in tag:
            continue
        dest = None
        if path == 'server/internal/world/dispatch.go' and name == 'dispatch':
            dest = receivers
        elif path.startswith('server/internal/world/') and name.startswith('handle'):
            dest = receivers
        elif path == 'server/cmd/gx_login/main.go' and name == 'frame':
            dest = receivers
        elif path.startswith('internal/game/') or path.startswith('internal/session/'):
            dest = client_receivers
        if dest is None:
            continue
        for expr in r.get('ops', []):
            op = evaluate(expr)
            if op is None or not 0x100 <= op <= 0xFFFF:
                continue
            known.add(op)
            dest[op].append(r)
            for call in r.get('calls', []):
                for fn in funcs.get(call.removeprefix(PREFIX), []):
                    if fn['path'].startswith('internal/proto/'):
                        codecs[op].append(fn)
    # The pre-encryption handshake is an if-guard, not a switch case.
    for op, path, name in [(0xA101, 'server/cmd/gx_login/main.go', 'frame'), (0xA301, 'server/internal/world/session.go', 'frame')]:
        for r in prod:
            if r['kind'] == 'func' and r['path'] == path and r['name'] == name:
                receivers[op].append(r)
    for r in prod:
        if r['kind'] == 'func' and r['path'] == 'internal/authwire/credentials.go':
            codecs[0x7F12].append(r)
    family_of = {}
    for op in sorted(known):
        for family, (_, groups) in FAMILIES.items():
            if op >> 8 in groups:
                family_of[op] = family
                break
        if op not in family_of:
            raise ValueError(f'Unclassified opcode 0x{op:04X}; extend FAMILIES')
    outputs = {}
    header = '> Generated from Go source by `tools/generate_server_reference.py`; do not edit by hand.\n'
    index = ['# Complete opcode catalog', '', header,
        'This catalog covers every typed opcode declaration in the shared/client and server codecs, plus literal opcodes found in live opcode switches and serializers. It documents the current GX implementation, not every opcode in every historical Shaiya client. Unknown fields keep their source names. A declaration alone is not evidence of a working endpoint.', '',
        'Read [wire conventions and session phases](server-protocol.md) first. The family pages include exact serializer/decoder declarations, including conditional and repeated fields, with links to the request handlers and client reducers. Server gameplay validation is described in [server architecture](server-architecture.md).', '',
        '**Evidence legend:** C→S means a direct server receive case/handshake was found; S→C means a server packet builder was found; client receive means a reducer case was found. A dash means no direct mapping in this inventory, not proof of absence. Generic forwarding and helper-based decoding are linked through source, and may not yield a direct mapping. Numeric opcodes are normalized before stack-width handling.', '',
        '| Opcode | Names / family reference | C→S | S→C | Client receive |', '|---|---|---|---|---|']
    for op in sorted(known):
        names = sorted({r['name'] for r in declarations[op]}) or ['Literal opcode']
        family = family_of[op]
        index.append(f"| `0x{op:04X}` | [{', '.join(names)}](protocol-{family}.md#opcode-{op:04x}) | {'yes' if receivers[op] else '—'} | {'yes' if writers[op] else '—'} | {'yes' if client_receivers[op] else '—'} |")
    index += ['', f'Inventory: **{len(known)} numeric opcodes**, **{sum(len(v) for v in declarations.values())} typed declarations**, **{len(FAMILIES)} families**.', '', 'Regenerate with `python3 tools/generate_server_reference.py`; verify freshness with `--check`. The generator parses source for both operating systems and tests are excluded from implementation evidence. Source links track main; use Git history when comparing older binaries.', '']
    outputs['server-opcodes.md'] = '\n'.join(index)
    proto_types = [r for r in prod if r['kind'] == 'type' and r['path'].startswith('internal/proto/')]
    for family, (title, _) in FAMILIES.items():
        lines = [f'# Protocol: {title}', '', header, '[All opcodes](server-opcodes.md) · [Wire conventions](server-protocol.md) · [Server architecture](server-architecture.md)', '',
            'Client codecs operate on opcode + body; server packet.Writer emits length + opcode + body. Excerpts preserve the respective construction API. Go wire methods specify widths, not host struct alignment. `Count` is special: see the stack-width compatibility rules. Loops, conditions, helper calls and fixed string sizes are significant; do not flatten them into a single fixed-size struct.', '']
        for op in sorted(k for k in known if family_of[k] == family):
            names = sorted({r['name'] for r in declarations[op]}) or ['Literal opcode']
            lines += [f'## Opcode {op:04x}', '', f'**`0x{op:04X}` — {", ".join(names)}**', '']
            if declarations[op]:
                lines += ['Declarations: '+', '.join(link(r, f"{r['package']}.{r['name']}") for r in declarations[op])+'.', '']
                docs = list(dict.fromkeys(r.get('doc','').strip() for r in declarations[op] if r.get('doc','').strip() and not r['doc'].startswith('Opcodes.')))
                if docs:
                    lines += ['Source notes: '+ ' '.join(docs).replace('\n',' '), '']
            for label, rows in [('C→S receive adapters', receivers[op]), ('S→C serializers', writers[op]), ('Client receive reducers', client_receivers[op])]:
                rows = unique(rows)
                lines += [f'**{label}:** '+(', '.join(link(r, f"{r['name']} ({r['path']})") for r in rows) if rows else 'no direct mapping found; inspect source before assuming support or direction.') , '']
            layouts = unique(writers[op]+codecs[op])
            if layouts:
                lines += ['### Wire layouts and validation', '', 'These are the exact source serializers/decoders. Parameter names and assignments identify fields; conditions identify optional variants. Referenced helpers and domain validation remain authoritative at the linked source.', '']
                types_done=set()
                for fn in layouts:
                    lines += [f'<details><summary>{fn["name"]} — {fn["path"]}</summary>', '', link(fn, 'Source definition'), '', '```go', fn['source'], '```', '']
                    for typ in proto_types:
                        if typ['path'] == fn['path'] and typ['name'] not in types_done and re.search(r'\b'+re.escape(typ['name'])+r'\b', fn['source']):
                            lines += ['```go', 'type '+typ['source'], '```', ''];types_done.add(typ['name'])
                    lines += ['</details>', '']
            else:
                lines += ['The shape is decoded inline, delegated, or only declared. Use the linked adapter/reducer and the codec appendix below; there is no inferred fixed layout here.', '']
        outputs['protocol-'+family+'.md']='\n'.join(lines)
    # Every shared codec file is included once in the appendix, even if a dynamic
    # caller prevents precise opcode association. This avoids silent omissions.
    codec_files=sorted({r['path'] for r in prod if r['kind']=='file' and r['path'].startswith('internal/proto/')})
    lines=['# Shared codec appendix','',header,
        'Complete shared codec source inventory, including helper readers, nested row definitions and codecs without a statically mapped opcode. Each file lists types and function definitions. This closes gaps left by dynamic dispatch; it is not a claim that every declared packet is accepted by the server.', '',
        '[Opcode catalog](server-opcodes.md) · [Wire conventions](server-protocol.md)', '']
    for path in codec_files:
        if path.endswith('opcodes.go'):
            continue
        rows=[r for r in prod if r['path']==path and r['kind'] in ('type','func')]
        lines += [f'## {Path(path).stem.replace("_"," ")}', '', link(next(r for r in prod if r['path']==path and r['kind']=='file'),path), '', '<details><summary>Types and codecs</summary>', '']
        for r in rows:
            lines += ['```go', ('type ' if r['kind']=='type' else '')+r['source'], '```', '']
        lines += ['</details>', '']
    outputs['server-codecs.md']='\n'.join(lines)
    # Complete file/symbol inventories for all server packages, including platform
    # alternatives; no build-dependent go list omissions.
    for section,title,select in [
        ('infrastructure','Infrastructure and command package API',lambda p:p.startswith('server/ai/') or p.startswith('server/cmd/') or p.startswith('server/internal/') and '/db/' not in p and '/world/' not in p),
        ('storage','SQLite package API',lambda p:p.startswith('server/internal/db/')),
        ('world','World package API',lambda p:p.startswith('server/internal/world/'))]:
        lines=[f'# {title}', '', header, '[Package responsibilities](server-packages.md) · [Server architecture](server-architecture.md)', '',
            'Every production Go file is listed, including OS-specific alternatives. Signatures include private helpers to make ownership and call boundaries inspectable. Declaration comments are source documentation, not an independent guarantee of behavior; the architecture guide explains transaction, concurrency and failure contracts.', '']
        files=sorted({r['path'] for r in prod if r['kind']=='file' and select(r['path'])})
        previous=''
        for path in files:
            pkg=Path(path).parent.as_posix()
            if pkg!=previous:
                lines += ['## '+pkg, '', PURPOSES.get(Path(pkg).name,''), ''];previous=pkg
                tests=sorted({r['path'] for r in records if r['kind']=='file' and Path(r['path']).parent.as_posix()==pkg and r['path'].endswith('_test.go')})
                if tests:
                    lines += ['Tests: '+', '.join(f'[{Path(t).name}](../../{t})' for t in tests)+'.', '']
            file=next(r for r in prod if r['path']==path and r['kind']=='file')
            lines += ['### '+Path(path).name, '', link(file,path), '']
            if file.get('doc'):
                lines += [file['doc'].strip(), '']
            deps=sorted(x.removeprefix(PREFIX) for x in file.get('imports',[]) if x.startswith(PREFIX))
            if deps:
                lines += ['Direct GX dependencies: '+', '.join(code(d) for d in deps)+'.', '']
            rows=[r for r in prod if r['path']==path and r['kind'] in ('type','func','const','var')]
            lines += ['| Declaration | Responsibility / source comment |', '|---|---|']
            for r in rows:
                declaration = r.get('signature') if r['kind']=='func' else r['kind']+' '+r['name']
                if r['kind']=='const' and r.get('value'):
                    declaration += ' = '+r['value']
                if r['kind']=='var' and r.get('signature'):
                    declaration += ' '+r['signature']
                comment=r.get('doc','').strip().replace('\n',' ').replace('|','&#124;') or 'See the source definition and the subsystem ownership contract.'
                lines += [f'| {link(r,code(r["name"]))}<br>{code(declaration)} | {comment} |']
            lines += ['']
            types = [r for r in rows if r['kind'] == 'type']
            if types:
                lines += ['<details><summary>Type definitions and owned fields</summary>', '']
                for typ in types:
                    lines += ['```go', 'type '+typ['source'], '```', '']
                lines += ['</details>', '']
        outputs['server-reference-'+section+'.md']='\n'.join(lines)
    stale=[]
    for name,content in outputs.items():
        content=content.rstrip()+'\n'
        target=WIKI/name
        if args.check:
            if not target.exists() or target.read_text(encoding='utf-8')!=content:stale.append(name)
        else:target.write_text(content, encoding='utf-8', newline='\n')
    if stale:
        raise SystemExit('Outdated generated Wiki pages: '+', '.join(stale))
    print(f'{"Checked" if args.check else "Generated"} {len(outputs)} Wiki pages; {len(known)} opcodes; {len(codec_files)} codec files; {sum(r["kind"]=="file" and r["path"].startswith("server/") for r in prod)} server production files.')

if __name__=='__main__':
    main()
