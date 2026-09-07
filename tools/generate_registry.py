#!/usr/bin/env python3
"""Generate internal/provider/registry.go and update provider.go."""

import json
import re
import subprocess
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
PROVIDER_DIR = ROOT / "internal" / "provider"
SPEC = ROOT / "provider_code_spec.json"
OPENAPI = ROOT / "openapi.yaml"
CONFIG = ROOT / "generator_config.yml"

# Data sources with hand-written wrappers (not generic managed data sources).
MANUAL_DS = {"domain", "image", "images", "post", "posts", "video", "videos", "feed", "feeds"}

# Hand-written resources that must always be registered.
MANUAL_RESOURCES = [
    "NewAccessPolicyResource",
    "NewAudioTrackResource",
    "NewBillingAddressResource",
    "NewBlogSubscriptionResource",
    "NewCustomDomainResource",
    "NewDashboardResource",
    "NewDashboardWidgetResource",
    "NewDomainAutoJoinResource",
    "NewFeedResource",
    "NewImageResource",
    "NewOrganizationInvitationResource",
    "NewOrganizationMemberResource",
    "NewOrganizationResource",
    "NewPostResource",
    "NewProjectResource",
    "NewSubtitleResource",
    "NewVideoResource",
]

# Resource-specific metadata for generic managed resources. Method names are
# SDK SimpleClient method names; path params are tfsdk attribute names.
RESOURCE_META: dict[str, dict] = {
    "api_key": {
        "create_method": "CreateApiKey",
        "read_method": "ListApiKeys",
        "delete_method": "DeleteApiKey",
        "path_params": ["org_id"],
        "delete_path_params": ["org_id", "id"],
        "create_keep_path_keys": [],
        "update_keep_path_keys": [],
        "body_renames": {},
        "computed_body_keys": ["id", "created_at", "last_used", "project_name", "secret"],
        "create_response_field": "api_key",
        "read_list_field": "api_keys",
        "read_after_create": False,
    },
    "client_credential": {
        "create_method": "CreateClientCredential",
        "read_method": "ListClientCredentials",
        "delete_method": "RevokeClientCredential",
        "path_params": [],
        "delete_path_params": ["id"],
        "create_keep_path_keys": [],
        "update_keep_path_keys": [],
        "body_renames": {},
        "computed_body_keys": ["id", "created_at", "client_id", "kid", "last_used_at", "status", "client_secret"],
        "create_response_field": "",
        "create_flatten_field": "credential",
        "read_list_field": "credentials",
        "read_after_create": False,
    },
    "subscription": {
        "create_method": "CreateSubscription",
        "read_method": "GetSubscription",
        "update_method": "UpgradeSubscription",
        "update_client_field": "Payments",
        "delete_method": "CancelSubscription",
        "path_params": [],
        "create_keep_path_keys": [],
        "update_keep_path_keys": [],
        "body_renames": {},
        "computed_body_keys": [
            "id", "plan_id", "plan_name", "plan_type", "status", "current_period_end",
            "cancel_at_period_end", "stripe_customer_id", "stripe_subscription_id", "currency",
            "expiring_soon", "price", "trials_ending_soon",
        ],
        "update_computed_body_keys": [
            "id", "payment_method_id", "billing_address", "plan_id", "plan_name", "plan_type", "status", "current_period_end",
            "cancel_at_period_end", "stripe_customer_id", "stripe_subscription_id", "currency",
            "expiring_soon", "price", "trials_ending_soon",
        ],
        "create_response_field": "",
        "read_list_field": "",
        "read_after_create": True,
        "read_after_update": True,
    },
    "project_custom_domain": {
        "create_method": "SetCustomDomain",
        "read_method": "GetProject",
        "update_method": "SetCustomDomain",
        "delete_method": "RemoveCustomDomain",
        "path_params": ["org_id", "project_id"],
        "create_keep_path_keys": [],
        "update_keep_path_keys": [],
        "body_renames": {},
        "computed_body_keys": ["id", "created_at", "updated_at"],
        "create_response_field": "",
        "read_list_field": "",
        "read_after_create": False,
    },
    "payment_method": {
        "create_method": "UpsertPaymentMethod",
        "read_method": "ListPaymentMethods",
        "update_method": "UpsertPaymentMethod",
        "delete_method": "DeletePaymentMethod",
        "path_params": [],
        "delete_path_params": ["id"],
        "create_keep_path_keys": [],
        "update_keep_path_keys": [],
        "body_renames": {},
        "computed_body_keys": ["id", "type", "provider", "details", "is_default", "brand", "last4", "exp_month", "exp_year", "created_at"],
        "create_response_field": "",
        "read_list_field": "payment_methods",
        "read_list_id_field": "id",
        "read_after_create": True,
        "read_after_update": True,
    },
    "policy_attachment": {
        "create_method": "AttachPolicy",
        "read_method": "ListPolicyAttachments",
        "delete_method": "DetachPolicy",
        "path_params": ["org_id", "policy_id"],
        "delete_path_params": ["org_id", "id"],
        "create_keep_path_keys": [],
        "update_keep_path_keys": [],
        "body_renames": {},
        "computed_body_keys": ["id", "created_at"],
        "create_response_field": "",
        "read_list_field": "attachments",
        "read_list_id_field": "id",
        "read_after_create": True,
        "read_after_update": False,
    },
}


def sdk_dir() -> Path:
    """Return the SDK module directory in the Go module cache."""
    modcache = subprocess.run(
        ["go", "env", "GOMODCACHE"],
        capture_output=True,
        text=True,
        check=True,
    ).stdout.strip()
    # go mod cache encodes module paths with URL escaping for capitals.
    return Path(modcache) / "github.com/rixlhq" / "rixl-go@v0.7.0" / "sdk"


def parse_resources_gen(sdk: Path) -> dict[str, str]:
    """Map SDK package base (e.g. 'feeds') to Client field name (e.g. 'Feeds')."""
    src = (sdk / "resources.gen.go").read_text()
    imports: dict[str, str] = {}
    for m in re.finditer(r'"github.com/rixlhq/rixl-go/sdk/([^"]+)"', src):
        pkg = m.group(1)
        field_pat = r"\n\s+([A-Za-z0-9_]+)\s+\*" + re.escape(pkg) + r"\.SimpleClient"
        mm = re.search(field_pat, src)
        if mm:
            imports[pkg] = mm.group(1)
    return imports


def build_method_to_field(sdk: Path, pkg_to_field: dict[str, str]) -> dict[str, str]:
    """Map SDK method name (e.g. 'ListApiKeys') to Client field (e.g. 'APIKeys')."""
    method_to_field: dict[str, str] = {}
    for pkg, field in pkg_to_field.items():
        pkg_dir = sdk / pkg
        if not pkg_dir.is_dir():
            continue
        for p in pkg_dir.glob("*.gen.go"):
            text = p.read_text()
            for m in re.finditer(r"func \(c \*SimpleClient\) ([A-Za-z0-9_]+)\(", text):
                method = m.group(1)
                # Methods are unique across the SDK; the first (and only) wins.
                if method not in method_to_field:
                    method_to_field[method] = field
    return method_to_field


def to_pascal(snake: str) -> str:
    return "".join(p[:1].upper() + p[1:] for p in snake.split("_"))


def load_openapi() -> dict | None:
    if not OPENAPI.exists():
        return None
    return yaml.safe_load(OPENAPI.read_text())


def load_config() -> dict:
    return yaml.safe_load(CONFIG.read_text())


def load_spec() -> dict:
    return json.loads(SPEC.read_text())


def operation_at_path(spec: dict, path: str, method: str) -> dict | None:
    return spec.get("paths", {}).get(path, {}).get(method)


def path_params(operation: dict) -> list[str]:
    return [p["name"] for p in operation.get("parameters", []) if p.get("in") == "path"]


def alias_for(name: str, aliases: dict[str, str]) -> str:
    if name in aliases:
        return aliases[name]
    # Fallback for dotted user.org_id if not explicitly aliased.
    if name == "user.org_id":
        return "org_id"
    return name


def read_config_for_ds(name: str, cfg: dict) -> dict:
    ds = cfg["data_sources"].get(name, {})
    attrs = ds.get("schema", {}).get("attributes", {})
    aliases = attrs.get("aliases", {})
    return {"path": ds.get("read", {}).get("path"), "aliases": aliases}


def build_ds_meta(name: str, cfg: dict, openapi: dict, method_to_field: dict) -> dict:
    info = read_config_for_ds(name, cfg)
    path = info["path"]
    aliases = info["aliases"]

    if not path:
        raise ValueError(f"data source {name}: no read path in generator_config")

    op = operation_at_path(openapi, path, "get")
    if not op:
        raise ValueError(f"data source {name}: no GET operation at {path}")

    opid = op.get("operationId", "")
    if not opid:
        raise ValueError(f"data source {name}: no operationId at {path}")

    read_method = opid.split(".")[-1]
    if read_method not in method_to_field:
        raise ValueError(f"data source {name}: SDK method {read_method} not found")

    client_field = method_to_field[read_method]
    params = [alias_for(p, aliases) for p in path_params(op)]

    return {
        "client_field": client_field,
        "read_method": read_method,
        "path_params": params,
    }


def generate_registry(datasources: list[str], meta: dict[str, dict]) -> str:
    lines = [
        "//nolint:revive // registry file is long by design",
        "package provider",
        "",
        'import "github.com/hashicorp/terraform-plugin-framework/datasource"',
        "",
        "// Data sources.",
        "",
    ]

    descriptors = []
    for name in sorted(datasources):
        if name in MANUAL_DS:
            continue
        m = meta[name]
        pascal = to_pascal(name)
        lines.append(f"func New{pascal}DataSource() datasource.DataSource {{")
        lines.append(f"\treturn newManagedDataSource({name}DataSourceDescriptor)")
        lines.append("}")
        lines.append("")
        descriptors.append((name, pascal, m))

    for name, pascal, m in descriptors:
        lines.append(f"var {name}DataSourceDescriptor = DataSourceDescriptor{{")
        lines.append(f'\tTypeName:    "{name}",')
        lines.append(f"\tSchemaFn:    {pascal}DataSourceSchema,")
        lines.append(f"\tModel:       (*{pascal}DataSourceModel)(nil),")
        lines.append(f'\tClientField: "{m["client_field"]}",')
        lines.append(f'\tReadMethod:  "{m["read_method"]}",')
        params = ", ".join(f'"{p}"' for p in m["path_params"])
        lines.append(f"\tPathParams:  []string{{{params}}},")
        lines.append("}")
        lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def go_string_list(items: list[str]) -> str:
    return ", ".join(f'"{item}"' for item in items)


def go_string_map(m: dict[str, str]) -> str:
    if not m:
        return "map[string]string{}"
    entries = ", ".join(f'"{k}": "{v}"' for k, v in sorted(m.items()))
    return f"map[string]string{{{entries}}}"


def generate_resource_registry(resources: list[str], method_to_field: dict[str, str]) -> str:
    lines = [
        "package provider",
        "",
        'import "github.com/hashicorp/terraform-plugin-framework/resource"',
        "",
        "// genericResourceConstructors returns the constructors for all generic",
        "// managed resources. Manual resources are still returned directly from",
        "// rixlProvider.Resources.",
        "func genericResourceConstructors() []func() resource.Resource {",
        "\treturn []func() resource.Resource{",
    ]
    for name in sorted(resources):
        pascal = to_pascal(name)
        lines.append(f"\t\tNew{pascal}Resource,")
    lines.extend([
        "\t}",
        "}",
        "",
    ])

    for name in sorted(resources):
        m = RESOURCE_META.get(name, {})
        if not m:
            print(f"warning: no RESOURCE_META for {name}; skipping", file=sys.stderr)
            continue

        pascal = to_pascal(name)
        create = m["create_method"]
        client_field = method_to_field.get(create)
        if not client_field:
            print(f"error: no SDK client field for {create}", file=sys.stderr)
            return ""

        lines.append(f"func New{pascal}Resource() resource.Resource {{")
        lines.append("\treturn newManagedResource(ResourceDescriptor{")
        lines.append(f'\t\tTypeName:            "{name}",')
        lines.append(f"\t\tSchemaFn:            {pascal}ResourceSchema,")
        lines.append(f"\t\tModel:               &{pascal}Model{{}},")
        lines.append(f'\t\tClientField:         "{client_field}",')
        if m.get("create_client_field"):
            lines.append(f'\t\tCreateClientField:   "{m["create_client_field"]}",')
        if m.get("read_client_field"):
            lines.append(f'\t\tReadClientField:     "{m["read_client_field"]}",')
        if m.get("update_client_field"):
            lines.append(f'\t\tUpdateClientField:   "{m["update_client_field"]}",')
        if m.get("delete_client_field"):
            lines.append(f'\t\tDeleteClientField:   "{m["delete_client_field"]}",')
        lines.append(f'\t\tCreateMethod:        "{create}",')
        if m.get("read_method"):
            lines.append(f'\t\tReadMethod:          "{m["read_method"]}",')
        if m.get("update_method"):
            lines.append(f'\t\tUpdateMethod:        "{m["update_method"]}",')
        lines.append(f'\t\tDeleteMethod:        "{m["delete_method"]}",')
        lines.append(f"\t\tPathParams:          []string{{{go_string_list(m.get('path_params', []))}}},")
        if m.get("create_path_params") is not None:
            lines.append(f"\t\tCreatePathParams:    []string{{{go_string_list(m['create_path_params'])}}},")
        if m.get("read_path_params") is not None:
            lines.append(f"\t\tReadPathParams:      []string{{{go_string_list(m['read_path_params'])}}},")
        if m.get("update_path_params") is not None:
            lines.append(f"\t\tUpdatePathParams:    []string{{{go_string_list(m['update_path_params'])}}},")
        if m.get("delete_path_params") is not None:
            lines.append(f"\t\tDeletePathParams:    []string{{{go_string_list(m['delete_path_params'])}}},")
        lines.append(f"\t\tCreateKeepPathKeys:  []string{{{go_string_list(m.get('create_keep_path_keys', []))}}},")
        lines.append(f"\t\tUpdateKeepPathKeys:  []string{{{go_string_list(m.get('update_keep_path_keys', []))}}},")
        lines.append(f"\t\tBodyRenames:         {go_string_map(m.get('body_renames', {}))},")
        lines.append(f"\t\tComputedBodyKeys:    []string{{{go_string_list(m.get('computed_body_keys', []))}}},")
        if m.get("update_computed_body_keys"):
            lines.append(f"\t\tUpdateComputedBodyKeys: []string{{{go_string_list(m['update_computed_body_keys'])}}},")
        if m.get("preserve_missing"):
            lines.append(f"\t\tPreserveMissing:     []string{{{go_string_list(m['preserve_missing'])}}},")
        lines.append(f'\t\tCreateResponseField: "{m.get("create_response_field", "")}",')
        if m.get("create_flatten_field"):
            lines.append(f'\t\tCreateFlattenField:  "{m["create_flatten_field"]}",')
        lines.append(f'\t\tReadListField:       "{m.get("read_list_field", "")}",')
        lines.append(f'\t\tReadListIDField:     "{m.get("read_list_id_field", "id")}",')
        lines.append(f"\t\tReadAfterCreate:     {str(m.get('read_after_create', False)).lower()},")
        lines.append(f"\t\tReadAfterUpdate:     {str(m.get('read_after_update', False)).lower()},")
        lines.append("\t})")
        lines.append("}")
        lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def update_provider(datasources: list[str]) -> None:
    path = PROVIDER_DIR / "provider.go"
    src = path.read_text()

    all_ds = sorted(set(datasources) | MANUAL_DS)
    ds_funcs = [f"New{to_pascal(n)}DataSource" for n in all_ds]
    ds_block = "\n\t\t".join(f"{fn}," for fn in ds_funcs)

    src = re.sub(
        r"func \(p \*rixlProvider\) DataSources\(_ context.Context\) \[\]func\(\) datasource\.DataSource \{\n\treturn \[\]func\(\) datasource\.DataSource\{(.*?)\n\t\}\n\}",
        f"func (p *rixlProvider) DataSources(_ context.Context) []func() datasource.DataSource {{\n\treturn []func() datasource.DataSource{{\n\t\t{ds_block}\n\t}}\n}}",
        src,
        flags=re.DOTALL,
    )

    res_funcs = sorted(MANUAL_RESOURCES)
    res_block = "\n\t\t".join(f"{r}," for r in res_funcs)
    src = re.sub(
        r"func \(p \*rixlProvider\) Resources\(_ context.Context\) \[\]func\(\) resource\.Resource \{.*?\}(?=\n*func |$)",
        f"func (p *rixlProvider) Resources(_ context.Context) []func() resource.Resource {{\n\treturn append([]func() resource.Resource{{\n\t\t{res_block}\n\t}}, genericResourceConstructors()...)\n}}\n",
        src,
        flags=re.DOTALL,
    )

    path.write_text(src)


def main() -> int:
    cfg = load_config()
    openapi = load_openapi()
    spec = load_spec()

    sdk = sdk_dir()
    pkg_to_field = parse_resources_gen(sdk)
    method_to_field = build_method_to_field(sdk, pkg_to_field)

    datasources = [ds["name"] for ds in spec.get("datasources", [])]

    if openapi is None:
        print("warning: openapi.yaml not found; skipping data source registry generation", file=sys.stderr)
    else:
        meta: dict[str, dict] = {}
        for name in datasources:
            if name in MANUAL_DS:
                continue
            try:
                meta[name] = build_ds_meta(name, cfg, openapi, method_to_field)
            except ValueError as e:
                print(f"error: {e}", file=sys.stderr)
                return 1

        registry = generate_registry(datasources, meta)
        (PROVIDER_DIR / "registry.go").write_text(registry)

    resources = sorted({r["name"] for r in spec.get("resources", [])} | set(RESOURCE_META.keys()))
    resource_registry = generate_resource_registry(resources, method_to_field)
    if not resource_registry:
        return 1
    (PROVIDER_DIR / "resource_registry.go").write_text(resource_registry)

    update_provider(datasources)

    for go_file in ["registry.go", "resource_registry.go", "provider.go"]:
        path = PROVIDER_DIR / go_file
        subprocess.run(["gofmt", "-w", str(path)], check=False)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
