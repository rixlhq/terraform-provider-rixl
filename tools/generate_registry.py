#!/usr/bin/env python3
"""Generate internal/provider/registry.go and update provider.go."""

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
PROVIDER = ROOT / "internal" / "provider"
SPEC = ROOT / "provider_code_spec.json"

# Data sources with hand-written wrappers (not generic).
MANUAL_DS = {"image", "images", "video", "videos", "feed", "feeds"}

# Mapping from data_source name to SDK client field, read method and path params.
DS_META = {
    "api_keys":                  ("APIKeys",           "ListApiKeys",                 ["org_id"]),
    "audio_tracks":              ("AudioTracks",       "ListAudioTracks",             ["video_id"]),
    "bandwidth_usage":           ("Usage",             "GetBandwidthUsage",           []),
    "bandwidth_usage_history":   ("Usage",             "GetBandwidthUsageHistory",    []),
    "billing_address":           ("Payments",          "GetBillingAddress",           []),
    "blog_subscription":         ("Blog",              "GetBlogSubscription",         []),
    "chapters":                  ("Chapters",          "GetVideoChapters",            ["project_id", "video_id"]),
    "client_credentials":        ("ClientCredentials", "ListClientCredentials",       []),
    "dashboard":                 ("Dashboards",        "GetDashboard",                ["id"]),
    "dashboards":                ("Dashboards",        "ListDashboards",              []),
    "invoices":                  ("Invoices",          "ListInvoices",                []),
    "languages":                 ("Languages",         "ListLanguages",               []),
    "membership_applications":   ("Memberships",       "ListMembershipApplications",  []),
    "memberships":               ("Memberships",       "ListMemberships",             []),
    "passkeys":                  ("Passkeys",          "ListPasskeys",                []),
    "payment_methods":           ("Payments",          "ListPaymentMethods",          []),
    "plan":                      ("Plans",             "GetPlan",                     ["id"]),
    "plans":                     ("Plans",             "ListPlans",                   []),
    "project":                   ("Projects",          "GetProject",                  ["org_id", "id"]),
    "projects":                  ("Projects",          "ListProjects",                ["org_id"]),
    "providers":                 ("SocialProviders",   "ListProviders",               []),
    "storage_usage":             ("Usage",             "GetStorageUsage",             []),
    "storage_usage_history":     ("Usage",             "GetStorageUsageHistory",      []),
    "subscription":              ("Subscriptions",     "GetSubscription",             []),
    "subscription_history":      ("Subscriptions",     "GetSubscriptionHistory",      []),
    "subtitles":                 ("Subtitles",         "ListSubtitles",               ["video_id"]),
    "user":                      ("Users",             "GetUser",                     []),
    "user_info":                 ("Users",             "GetUserInfo",                 []),
}


def to_pascal(snake: str) -> str:
    return "".join(p[:1].upper() + p[1:] for p in snake.split("_"))


def generate_registry(datasources: list[str]) -> str:
    lines = [
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
        client, method, params = DS_META[name]
        pascal = to_pascal(name)
        lines.append(f"func New{pascal}DataSource() datasource.DataSource {{")
        lines.append(f"\treturn newManagedDataSource({name}DataSourceDescriptor)")
        lines.append("}")
        lines.append("")
        descriptors.append((name, pascal, client, method, params))

    for name, pascal, client, method, params in descriptors:
        lines.append(f"var {name}DataSourceDescriptor = DataSourceDescriptor{{")
        lines.append(f'\tTypeName:    "{name}",')
        lines.append(f"\tSchemaFn:    {pascal}DataSourceSchema,")
        lines.append(f"\tModel:       (*{pascal}DataSourceModel)(nil),")
        lines.append(f'\tClientField: "{client}",')
        lines.append(f'\tReadMethod:  "{method}",')
        params_str = ", ".join(f'"{p}"' for p in params)
        lines.append(f"\tPathParams:  []string{{{params_str}}},")
        lines.append("}")
        lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def update_provider(datasources: list[str]) -> None:
    path = PROVIDER / "provider.go"
    src = path.read_text()

    # Data source names in provider.go order: manual first, then generic sorted.
    ds_names = sorted(n for n in datasources if n not in MANUAL_DS)
    all_ds = ["feed", "feeds", "image", "images", "video", "videos", "project", "projects", "api_keys"] + [n for n in ds_names if n not in {"project", "projects", "api_keys"}]
    ds_funcs = sorted(set(f"New{to_pascal(n)}DataSource" for n in all_ds), key=lambda s: s.lower())

    ds_block = "\n\t\t".join(f"{fn}," for fn in ds_funcs)
    src = re.sub(
        r"func \(p \*rixlProvider\) DataSources\(_ context.Context\) \[\]func\(\) datasource\.DataSource \{\n\treturn \[\]func\(\) datasource\.DataSource\{(.*?)\n\t\}\n\}",
        f"func (p *rixlProvider) DataSources(_ context.Context) []func() datasource.DataSource {{\n\treturn []func() datasource.DataSource{{\n\t\t{ds_block}\n\t}}\n}}",
        src,
        flags=re.DOTALL,
    )

    resources = ["NewFeedResource", "NewProjectResource"]
    res_block = "\n\t\t".join(f"{r}," for r in resources)
    src = re.sub(
        r"func \(p \*rixlProvider\) Resources\(_ context.Context\) \[\]func\(\) resource\.Resource \{\n\treturn \[\]func\(\) resource\.Resource\{(.*?)\n\t\}\n\}",
        f"func (p *rixlProvider) Resources(_ context.Context) []func() resource.Resource {{\n\treturn []func() resource.Resource{{\n\t\t{res_block}\n\t}}\n}}",
        src,
        flags=re.DOTALL,
    )

    path.write_text(src)


def main() -> None:
    spec = json.loads(SPEC.read_text())
    datasources = [ds["name"] for ds in spec.get("datasources", [])]

    registry = generate_registry(datasources)
    (PROVIDER / "registry.go").write_text(registry)
    update_provider(datasources)


if __name__ == "__main__":
    main()
