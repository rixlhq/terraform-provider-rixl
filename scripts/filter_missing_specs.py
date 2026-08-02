import json, re, sys

def main(spec_path, go_path):
    with open(spec_path) as f:
        names = {d['name'] for d in json.load(f).get('datasources', [])}
    with open(go_path) as f:
        text = f.read()
    pattern = r'        \{\n            typeName:     "([a-z0-9_]+)",.*?        \},\n'
    def repl(m):
        return m.group(0) if m.group(1) in names else ''
    text = re.sub(pattern, repl, text, flags=re.DOTALL)
    with open(go_path, 'w') as f:
        f.write(text)
    missing = 55 - len(names)  # rough; not used
    print(f'Filtered {go_path} to {len(names)} data sources')

if __name__ == '__main__':
    main(sys.argv[1], sys.argv[2])
