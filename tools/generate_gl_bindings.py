#!/usr/bin/env python3
"""Generate GX's OpenGL 4.5 bindings with the upstream glow tool."""

import json
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
from xml.etree import ElementTree


ROOT = Path(__file__).resolve().parents[1]
PACKAGE = ROOT / 'internal/render/gl'
IMPORT = 'github.com/Spelunkern/Shaiya-GX/internal/render/gl'
GLOW = 'github.com/go-gl/glow@v0.0.0-20260331232744-394d535962b3'
HELPERS = {'GoStr', 'Init', 'InitWithProcAddrFunc', 'Ptr', 'Str', 'Strs'}


def used_symbols():
    symbols = set()
    binding_import = re.compile(r'^[ \t]*(?:import[ \t]+)?(?:(\w+)[ \t]+)?"' + re.escape(IMPORT) + r'"', re.M)
    for directory in (ROOT / 'cmd', ROOT / 'internal'):
        for path in directory.rglob('*.go'):
            if path.is_relative_to(PACKAGE):
                continue
            text = path.read_text()
            match = binding_import.search(text)
            if match:
                alias = match[1] or 'gl'
                symbols.update(re.findall(r'\b' + alias + r'\.([A-Z][a-zA-Z0-9_]*)', text))
    return symbols - HELPERS


def main():
    symbols = used_symbols()
    if not symbols:
        raise SystemExit('No OpenGL symbols found')
    symbols.add('TRUE')  # glow uses GL_TRUE for GLboolean return values.
    if 'MultiDrawElementsIndirectWithOffset' in symbols:
        symbols.remove('MultiDrawElementsIndirectWithOffset')
        symbols.add('MultiDrawElementsIndirect')
    restriction = {
        'Enums': sorted('GL_' + symbol for symbol in symbols if symbol.isupper()),
        'Functions': sorted('gl' + symbol for symbol in symbols if not symbol.isupper()),
    }
    glow = Path(json.loads(subprocess.check_output(['go', 'mod', 'download', '-json', GLOW], cwd=ROOT))['Dir'])
    with tempfile.TemporaryDirectory(prefix='gx-gl-bindings-') as tmp:
        tmp = Path(tmp)
        xml = tmp / 'xml'
        shutil.copytree(glow / 'xml', xml, copy_function=shutil.copyfile)
        overloads = xml / 'overload/gl.xml'
        tree = ElementTree.parse(overloads)
        overload = ElementTree.SubElement(tree.getroot(), 'overload', {
            'name': 'glMultiDrawElementsIndirect',
            'overloadName': 'MultiDrawElementsIndirectWithOffset',
        })
        params = ElementTree.SubElement(overload, 'parameterChanges')
        offset = ElementTree.SubElement(params, 'change', {'index': '2'})
        ElementTree.SubElement(offset, 'type', {'signature': 'uintptr_t'})
        tree.write(overloads)
        subset = tmp / 'symbols.json'
        subset.write_text(json.dumps(restriction))
        generated = tmp / 'gl'
        subprocess.run([
            'go', 'run', GLOW, 'generate', '-api', 'gl', '-version', '4.5', '-profile', 'core',
            '-xml', str(xml), '-tmpl', str(glow / 'tmpl'), '-out', str(generated),
            '-addext', '^GL_ARB_bindless_texture$', '-restrict', str(subset),
        ], cwd=ROOT, check=True)
        # An OpenGL buffer offset stays an integer until the generated C
        # overload. GX does not use the legacy Go integer-to-pointer helpers.
        conversions = generated / 'conversions.go'
        content = conversions.read_text().replace('\tcase reflect.Uintptr:\n\t\taddr = unsafe.Pointer(data.(uintptr))\n', '')
        start, end = content.index('// PtrOffset '), content.index('// Str takes ')
        # Keep the Go string alive while C reads its NUL-terminated bytes.
        content = content.replace(
            'header := (*reflect.StringHeader)(unsafe.Pointer(&str))\n\treturn (*uint8)(unsafe.Pointer(header.Data))',
            'return (*uint8)(unsafe.Pointer(unsafe.StringData(str)))',
        )
        conversions.write_text(content[:start] + content[end:])
        # glow already installs its C trampoline. Passing the Go function's
        # address adds a Go pointer to the C call without serving a purpose.
        package = generated / 'package.go'
        package.write_text(package.read_text().replace('(C.GLDEBUGPROC)(unsafe.Pointer(&callback))', 'nil'))
        shutil.copytree(generated, PACKAGE, dirs_exist_ok=True)
        shutil.copyfile(glow / 'LICENSE', PACKAGE / 'LICENSE')


if __name__ == '__main__':
    main()
