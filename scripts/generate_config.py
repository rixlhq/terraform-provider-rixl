import json, re, os, sys

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SPEC_PATH = sys.argv[1] if len(sys.argv) > 1 else os.path.join(BASE_DIR, 'openapi.json')
OUT_CONFIG = os.path.join(BASE_DIR, 'generator_config.yml')
OUT_SPECS = os.path.join(BASE_DIR, 'internal/provider/data_source_specs.go')

with open(SPEC_PATH) as f:
    spec = json.load(f)

paths = spec['paths']

NAME_OVERRIDES = {
    '/feeds/v1/projects/{project_id}/feeds': 'feeds',
    '/feeds/v1/projects/{project_id}/feeds/{feed_id}': 'feed',
    '/media/v1/images/{image_id}': 'image',
    '/media/v1/projects/{project_id}/images': 'images',
    '/media/v1/videos/{video_id}': 'video',
    '/media/v1/projects/{project_id}/videos': 'videos',
    '/organizations/{org_id}/projects/v1': 'projects',
    '/organizations/{org_id}/projects/v1/{project_id}': 'project',
    '/organizations/{org_id}/api-keys/v1': 'api_keys',
    '/media/v1/languages': 'languages',
}

PLURAL_SINGULAR = {
    'api-keys': 'api_key',
    'api_keys': 'api_key',
    'keys': 'key',
    'feeds': 'feed',
    'images': 'image',
    'videos': 'video',
    'projects': 'project',
    'posts': 'post',
    'policies': 'policy',
    'members': 'member',
    'memberships': 'membership',
    'applications': 'application',
    'credentials': 'credential',
    'invoices': 'invoice',
    'plans': 'plan',
    'methods': 'method',
    'permissions': 'permission',
    'providers': 'provider',
    'passkeys': 'passkey',
    'tracks': 'track',
    'subtitles': 'subtitle',
    'chapters': 'chapter',
    'languages': 'language',
    'events': 'event',
    'attachments': 'attachment',
    'usages': 'usage',
    'payments': 'payment',
    'subscriptions': 'subscription',
    'addresses': 'address',
    'domains': 'domain',
    'emails': 'email',
    'stats': 'stat',
    'segments': 'segment',
    'dashboards': 'dashboard',
}

def singular(word):
    if word in PLURAL_SINGULAR:
        return PLURAL_SINGULAR[word]
    if word.endswith('s') and not word.endswith('ss') and not word.endswith('us'):
        return word[:-1]
    return word

def snake(s):
    s = s.replace('-', '_').replace('.', '_').replace('{', '').replace('}', '')
    s = re.sub(r'([a-z0-9])([A-Z])', r'\1_\2', s)
    return s.lower()

def norm_attr(name):
    return name.replace('.', '').lower()

def resource_name_from_path(path):
    if path in NAME_OVERRIDES:
        return NAME_OVERRIDES[path]
    parts = [p for p in path.split('/') if p and not re.match(r'^v\d', p)]
    out = []
    for i, p in enumerate(parts):
        if p.startswith('{') and p.endswith('}'):
            param = p[1:-1]
            if i > 0:
                prev_raw = parts[i-1].replace('-', '_')
                prev = singular(prev_raw)
                if param == f"{prev}_id" or param.endswith(f"_{prev}_id"):
                    if out:
                        out[-1] = prev
                    continue
            out.append(snake(param))
        else:
            out.append(snake(p))
    return '_'.join(out)

def resolve_ref(s):
    while '$ref' in s:
        name = s['$ref'].split('/')[-1]
        s = spec['components']['schemas'][name]
    return s

def get_response_schema(op):
    resp = op.get('responses', {}).get('200', {}).get('content', {}).get('application/json', {}).get('schema')
    if not resp:
        return None
    return resolve_ref(resp)

def is_list_response(resp):
    if not resp:
        return False
    if resp.get('type') == 'array':
        return True
    for k, v in resp.get('properties', {}).items():
        v = resolve_ref(v)
        if v.get('type') == 'array':
            return True
    return False

def top_level_attr_names(resp):
    names = set()
    if not resp:
        return names
    if resp.get('type') == 'array':
        return names
    for k in resp.get('properties', {}).keys():
        names.add(norm_attr(k))
    return names

def detect_id_alias(path, resp):
    parts = [p for p in path.split('/') if p]
    if not parts:
        return None
    last = parts[-1]
    if not (last.startswith('{') and last.endswith('}')):
        return None
    param = last[1:-1]
    if not param.endswith('_id'):
        return None
    if is_list_response(resp):
        return None
    if len(parts) >= 2:
        prev = parts[-2].replace('-', '_')
        s = singular(prev)
        if param == f"{s}_id" or param.endswith(f"_{s}_id"):
            return param, 'id'
    return None

def path_params(path):
    return re.findall(r'\{([^}]+)\}', path)

def quote(s):
    if re.match(r'^[A-Za-z0-9_.-]+$', str(s)):
        return str(s)
    return '"' + str(s).replace('\\', '\\\\').replace('"', '\\"') + '"'

config_lines = [
    'provider:',
    '  name: rixl',
    'resources:',
    '  feed:',
    '    create:',
    '      path: /feeds/v1/projects/{project_id}/feeds',
    '      method: POST',
    '    read:',
    '      path: /feeds/v1/projects/{project_id}/feeds/{feed_id}',
    '      method: GET',
    '    update:',
    '      path: /feeds/v1/projects/{project_id}/feeds/{feed_id}',
    '      method: PUT',
    '    delete:',
    '      path: /feeds/v1/projects/{project_id}/feeds/{feed_id}',
    '      method: DELETE',
    '    schema:',
    '      attributes:',
    '        aliases:',
    '          feed_id: id',
    '        overrides:',
    '          project_id:',
    '            computed_optional_required: required',
    '          id:',
    '            computed_optional_required: required',
    'data_sources:',
]

specs = []

for path, ops in sorted(paths.items()):
    if 'get' not in ops:
        continue
    if path.startswith('/internal') or '/posts/' in path or path.startswith('/posts'):
        continue
    op = ops['get']
    resp = get_response_schema(op)
    name = resource_name_from_path(path)
    params = path_params(path)
    query = [p for p in op.get('parameters', []) if p.get('in') == 'query']

    aliases = {}
    overrides = {}
    query_params = {}

    for param in params:
        attr = param
        if param == 'user.org_id':
            attr = 'org_id'
            aliases['user.org_id'] = 'org_id'
        overrides[attr] = 'required'
    id_alias = detect_id_alias(path, resp)
    if id_alias:
        param, attr = id_alias
        aliases[param] = attr
        overrides['id'] = 'required'
        if param in overrides:
            del overrides[param]

    top_names = top_level_attr_names(resp)
    for attr in overrides:
        top_names.add(norm_attr(attr))

    for q in query:
        qname = q['name']
        attr = norm_attr(qname)
        if '.' in qname:
            nice = qname.replace('.', '_')
            if nice != attr:
                aliases[qname] = nice
                attr = nice
        if attr in top_names:
            alias = 'query_' + attr
            aliases[qname] = alias
            attr = alias
        overrides[attr] = 'optional'
        query_params[attr] = qname

    config_lines.append(f'  {name}:')
    config_lines.append('    read:')
    config_lines.append(f'      path: {quote(path)}')
    config_lines.append('      method: GET')
    config_lines.append('    schema:')
    config_lines.append('      attributes:')
    if aliases:
        config_lines.append('        aliases:')
        for k, v in sorted(aliases.items()):
            config_lines.append(f'          {quote(k)}: {v}')
    config_lines.append('        overrides:')
    for k, v in sorted(overrides.items()):
        config_lines.append(f'          {k}:')
        config_lines.append(f'            computed_optional_required: {v}')

    def pascal(s):
        return ''.join(x.capitalize() or '_' for x in s.split('_'))
    schema_fn = f'{pascal(name)}DataSourceSchema'
    param_aliases_go = {}
    for k, v in aliases.items():
        if k in params:
            param_aliases_go[k] = v
    if id_alias:
        param, attr = id_alias
        param_aliases_go[param] = attr
    for param in params:
        if param == 'user.org_id':
            param_aliases_go['user.org_id'] = 'org_id'
    specs.append({
        'name': name,
        'type_name': name,
        'read_path': path,
        'schema_fn': schema_fn,
        'param_aliases': param_aliases_go,
        'query_params': query_params,
    })

with open(OUT_CONFIG, 'w') as f:
    f.write('\n'.join(config_lines) + '\n')

lines = [
    '// Code generated by gen-provider-config.py from openapi.json. DO NOT EDIT.',
    '',
    'package provider',
    '',
    'import "context"',
    'import "github.com/hashicorp/terraform-plugin-framework/datasource"',
    '',
    'func dataSourceSpecs() []rixlDataSourceSpec {',
    '    return []rixlDataSourceSpec{',
]
for s in specs:
    if s['param_aliases']:
        pa_items = [f'"{k}": "{v}"' for k, v in sorted(s['param_aliases'].items())]
        pa = 'map[string]string{' + ', '.join(pa_items) + '}'
    else:
        pa = 'nil'
    if s['query_params']:
        qp_items = [f'"{k}": "{v}"' for k, v in sorted(s['query_params'].items())]
        qp = 'map[string]string{' + ', '.join(qp_items) + '}'
    else:
        qp = 'nil'
    lines.append('        {')
    lines.append(f'            typeName:     "{s["type_name"]}",')
    lines.append(f'            readPath:     "{s["read_path"]}",')
    lines.append(f'            paramAliases: {pa},')
    lines.append(f'            queryParams:  {qp},')
    lines.append(f'            schemaFn:     {s["schema_fn"]},')
    lines.append('        },')
lines.extend([
    '    }',
    '}',
    '',
    'func (p *RixlProvider) DataSources(_ context.Context) []func() datasource.DataSource {',
    '    specs := dataSourceSpecs()',
    '    out := make([]func() datasource.DataSource, len(specs))',
    '    for i := range specs {',
    '        spec := specs[i]',
    '        out[i] = func(s rixlDataSourceSpec) func() datasource.DataSource {',
    '            return func() datasource.DataSource { return newRixlDataSource(s) }',
    '        }(spec)',
    '    }',
    '    return out',
    '}',
])

with open(OUT_SPECS, 'w') as f:
    f.write('\n'.join(lines) + '\n')

print(f'Wrote {OUT_CONFIG} with {len(specs)} data sources')
print(f'Wrote {OUT_SPECS}')
