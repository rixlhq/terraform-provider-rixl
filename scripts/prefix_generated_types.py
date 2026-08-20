import os, re, glob, sys

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DIR = os.path.join(BASE_DIR, 'internal/provider')

def get_prefix(text):
    # main schema function: func XxxDataSourceSchema or func XxxResourceSchema
    m = re.search(r'^func\s+([A-Za-z0-9_]+?)(?:DataSourceSchema|ResourceSchema)\s*\(', text, re.MULTILINE)
    if not m:
        return None
    return m.group(1)

def main():
    files = glob.glob(os.path.join(DIR, '*_data_source_gen.go')) + glob.glob(os.path.join(DIR, '*_resource_gen.go'))
    for path in files:
        with open(path) as f:
            text = f.read()
        prefix = get_prefix(text)
        if not prefix:
            continue
        # find main model type to exclude
        main_model = None
        for m in re.finditer(r'^type\s+([A-Za-z0-9_]+?)\s+struct\s*\{', text, re.MULTILINE):
            name = m.group(1)
            if name == prefix + 'Model' or name == prefix + 'DataSourceModel' or name == prefix + 'ResourceModel':
                main_model = name
                break
        if not main_model:
            # fallback: look for model ending with Model in file
            pass
        # find top-level type names and top-level function names
        ids = set()
        # type declarations
        for m in re.finditer(r'^type\s+([A-Z][A-Za-z0-9_]*)\s+(?:struct|interface)\s*\{', text, re.MULTILINE):
            name = m.group(1)
            if name == main_model:
                continue
            ids.add(name)
        # top-level func declarations: ^func Name(  (not methods)
        for m in re.finditer(r'^func\s+([A-Z][A-Za-z0-9_]*)\s*\(', text, re.MULTILINE):
            name = m.group(1)
            # exclude main schema function
            if name == prefix + 'DataSourceSchema' or name == prefix + 'ResourceSchema':
                continue
            ids.add(name)
        # replace longest first
        for orig in sorted(ids, key=len, reverse=True):
            new = prefix + orig
            text = re.sub(r'\b' + re.escape(orig) + r'\b', new, text)
        with open(path, 'w') as f:
            f.write(text)
        print(f'{os.path.basename(path)}: prefixed {len(ids)} identifiers with {prefix}')

if __name__ == '__main__':
    main()
