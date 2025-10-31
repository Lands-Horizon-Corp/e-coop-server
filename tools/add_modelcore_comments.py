#!/usr/bin/env python3
import re
import sys
from pathlib import Path

def process_file(path: Path) -> bool:
    text = path.read_text()
    lines = text.splitlines()
    out = []
    changed = False

    inside_type_block = False
    i = 0
    # helper to get previous non-empty, non-comment line index
    def prev_non_comment(j):
        k = j-1
        while k >= 0:
            s = lines[k].strip()
            if s == "":
                k -= 1
                continue
            if s.startswith("//"):
                return k
            return k
        return -1

    while i < len(lines):
        line = lines[i]
        stripped = line.lstrip()
        # detect start/end of type block
        if re.match(r'^\s*type\s*\($', line):
            inside_type_block = True
            out.append(line)
            i += 1
            continue
        if inside_type_block and re.match(r'^\s*\)\s*$', line):
            inside_type_block = False
            out.append(line)
            i += 1
            continue

        # match type line inside block (no leading 'type')
        m = None
        if inside_type_block:
            m = re.match(r'^(?P<indent>\s*)(?P<name>[A-Z][A-Za-z0-9_]*)\s+struct\s*{', line)
            if m:
                name = m.group('name')
                # check previous line for comment starting with name
                prev_idx = i-1
                has_comment = False
                if prev_idx >= 0:
                    prev = lines[prev_idx].lstrip()
                    if prev.startswith('//') and prev[2:].lstrip().startswith(name):
                        has_comment = True
                if not has_comment:
                    # build comment
                    if name.endswith('Response'):
                        comment = f"// {name} represents the response structure for {name[:-8]}."
                    elif name.endswith('Request'):
                        comment = f"// {name} represents the request structure for {name[:-7]}."
                    else:
                        comment = f"// {name} represents the {name} model."
                    out.append(m.group('indent') + comment)
                    changed = True
                out.append(line)
                i += 1
                continue
        else:
            # match standalone type declaration
            m = re.match(r'^(?P<indent>\s*)type\s+(?P<name>[A-Z][A-Za-z0-9_]*)\s+struct\s*{', line)
            if m:
                name = m.group('name')
                prev_idx = i-1
                has_comment = False
                if prev_idx >= 0:
                    prev = lines[prev_idx].lstrip()
                    if prev.startswith('//') and prev[2:].lstrip().startswith(name):
                        has_comment = True
                if not has_comment:
                    if name.endswith('Response'):
                        comment = f"// {name} represents the response structure for {name[:-8]}."
                    elif name.endswith('Request'):
                        comment = f"// {name} represents the request structure for {name[:-7]}."
                    else:
                        comment = f"// {name} represents the {name} model."
                    out.append(m.group('indent') + comment)
                    changed = True
                out.append(line)
                i += 1
                continue

        # match exported ModelCore method
        mm = re.match(r'^(?P<indent>\s*)func\s*\(m\s+\*ModelCore\)\s+(?P<name>[A-Z][A-Za-z0-9_]*)\s*\(', line)
        if mm:
            name = mm.group('name')
            prev_idx = i-1
            has_comment = False
            if prev_idx >= 0:
                prev = lines[prev_idx].lstrip()
                if prev.startswith('//') and prev[2:].lstrip().startswith(name):
                    has_comment = True
            if not has_comment:
                comment = f"// {name} returns {name} for the current branch or organization where applicable."
                out.append(mm.group('indent') + comment)
                changed = True
            out.append(line)
            i += 1
            continue

        out.append(line)
        i += 1

    if changed:
        path.write_text('\n'.join(out) + '\n')
    return changed


def main():
    base = Path('server/model/modelcore')
    if not base.exists():
        print('modelcore path not found', file=sys.stderr)
        sys.exit(1)
    files = sorted(base.glob('*.go'))
    any_changed = False
    for f in files:
        try:
            changed = process_file(f)
            if changed:
                print(f'updated: {f}')
                any_changed = True
        except Exception as e:
            print(f'error processing {f}: {e}', file=sys.stderr)
    if not any_changed:
        print('no changes')

if __name__ == '__main__':
    main()
