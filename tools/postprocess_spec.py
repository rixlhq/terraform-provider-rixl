#!/usr/bin/env python3
"""Post-process provider_code_spec.json after tfplugingen-openapi.

* Removes sensitive 'secret' attributes.
* Adds and fixes resource schemas for generic managed resources.
* Ensures path parameters are required and have RequiresReplace.
"""

from __future__ import annotations

import copy
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SPEC = ROOT / "provider_code_spec.json"


def load_spec() -> dict:
    return json.loads(SPEC.read_text())


def save_spec(spec: dict) -> None:
    SPEC.write_text(json.dumps(spec, indent="\t") + "\n")


def _child_attributes(attr: dict) -> list | None:
    """Return the child attribute list for a nested attribute, if any."""
    for kind in ("single_nested", "list_nested", "set_nested", "map_nested", "object"):
        if kind not in attr:
            continue
        container = attr[kind]
        if "attributes" in container:
            return container["attributes"]
        if "nested_object" in container and "attributes" in container["nested_object"]:
            return container["nested_object"]["attributes"]
    return None


def walk_attributes(attributes: list, fn) -> None:
    for attr in attributes:
        fn(attr)
        children = _child_attributes(attr)
        if children is not None:
            walk_attributes(children, fn)


def rename_attributes(attributes: list, old: str, new: str) -> None:
    def fn(attr):
        if attr.get("name") == old:
            attr["name"] = new

    walk_attributes(attributes, fn)


def remove_secret(attributes: list) -> list:
    def fn(attr):
        if attr.get("name") == "secret":
            attr["_delete"] = True

    def filter_attrs(attrs: list) -> list:
        for a in attrs:
            if a.get("_delete"):
                continue
            children = _child_attributes(a)
            if children is not None:
                if "single_nested" in a:
                    a["single_nested"]["attributes"] = filter_attrs(children)
                elif "object" in a:
                    a["object"]["attributes"] = filter_attrs(children)
                elif "list_nested" in a:
                    a["list_nested"]["nested_object"]["attributes"] = filter_attrs(children)
                elif "set_nested" in a:
                    a["set_nested"]["nested_object"]["attributes"] = filter_attrs(children)
                elif "map_nested" in a:
                    a["map_nested"]["nested_object"]["attributes"] = filter_attrs(children)
        return [a for a in attrs if not a.get("_delete")]

    walk_attributes(attributes, fn)
    return filter_attrs(attributes)


TIMESTAMP_REPLACEMENT = 'RFC 3339 timestamp, e.g. "2024-12-25T12:00:00Z".'


def sanitize_timestamp_descriptions(attributes: list) -> None:
    """Replace verbose protobuf Timestamp descriptions with a concise form.

    The OpenAPI spec embeds the full google.protobuf.Timestamp description
    (including '# Examples' headings) which breaks tfplugindocs rendering.
    """

    def fn(attr: dict) -> None:
        for kind in (
            "string",
            "int64",
            "bool",
            "float64",
            "number",
            "list",
            "set",
            "map",
            "single_nested",
            "list_nested",
            "set_nested",
            "map_nested",
            "object",
            "dynamic",
        ):
            container = attr.get(kind)
            if not isinstance(container, dict):
                continue
            desc = container.get("description")
            if isinstance(desc, str) and desc.startswith("A Timestamp represents"):
                container["description"] = TIMESTAMP_REPLACEMENT

    walk_attributes(attributes, fn)


def get_attr_type(attr: dict) -> str | None:
    for key in (
        "string",
        "int64",
        "bool",
        "float64",
        "number",
        "list",
        "set",
        "map",
        "single_nested",
        "list_nested",
        "set_nested",
        "map_nested",
        "object",
        "dynamic",
    ):
        if key in attr:
            return key
    return None


def set_cor(attr: dict, value: str) -> None:
    kind = get_attr_type(attr)
    if kind and kind in attr:
        attr[kind]["computed_optional_required"] = value


def get_cor(attr: dict) -> str | None:
    kind = get_attr_type(attr)
    if kind and kind in attr:
        return attr[kind].get("computed_optional_required")
    return None


def add_plan_modifier(attr: dict, pkg: str, expr: str) -> None:
    kind = get_attr_type(attr)
    if not kind or kind not in attr:
        return
    if "plan_modifiers" not in attr[kind]:
        attr[kind]["plan_modifiers"] = []
    attr[kind]["plan_modifiers"].append(
        {
            "custom": {
                "imports": [
                    {"path": "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"},
                    {"path": f"github.com/hashicorp/terraform-plugin-framework/resource/schema/{pkg}"},
                ],
                "schema_definition": expr,
            }
        }
    )


def string_modifier(attr: dict, expr: str) -> None:
    add_plan_modifier(attr, "stringplanmodifier", expr)


def object_modifier(attr: dict, expr: str) -> None:
    add_plan_modifier(attr, "objectplanmodifier", expr)


def find_resource(spec: dict, name: str) -> dict | None:
    for r in spec.get("resources", []):
        if r["name"] == name:
            return r
    return None


def add_resource(spec: dict, name: str, attributes: list) -> dict:
    existing = find_resource(spec, name)
    if existing:
        existing["schema"]["attributes"] = attributes
    else:
        existing = {"name": name, "schema": {"attributes": attributes}}
        spec.setdefault("resources", []).append(existing)
    return existing


def attr_by_name(attributes: list, name: str) -> dict | None:
    for a in attributes:
        if a["name"] == name:
            return a
    return None


def nested_object_attrs(parent_attr: dict) -> list | None:
    for kind in ("single_nested", "list_nested", "set_nested", "map_nested"):
        if kind in parent_attr and "nested_object" in parent_attr[kind]:
            return parent_attr[kind]["nested_object"]["attributes"]
    return None


def extract_nested_item_attributes(ds_name: str, list_attr_name: str, spec: dict) -> list:
    for ds in spec.get("datasources", []):
        if ds["name"] == ds_name:
            for attr in ds["schema"]["attributes"]:
                if attr["name"] == list_attr_name:
                    return copy.deepcopy(nested_object_attrs(attr) or [])
    return []


def find_nested_attr(parent_attr: dict, name: str) -> dict | None:
    for a in nested_object_attrs(parent_attr) or []:
        if a["name"] == name:
            return a
    return None


def fix_subscription(spec: dict) -> None:
    r = find_resource(spec, "subscription")
    if not r:
        return

    attrs = r["schema"]["attributes"]

    # Stripe price is required on create and does not come back in Get.
    sp = attr_by_name(attrs, "stripe_price_id")
    if sp:
        set_cor(sp, "required")
        string_modifier(sp, "stringplanmodifier.UseStateForUnknown()")
        string_modifier(sp, "stringplanmodifier.RequiresReplace()")

    # Payment method is optional on create and does not come back in Get.
    pm = attr_by_name(attrs, "payment_method_id")
    if pm:
        set_cor(pm, "computed_optional")
        string_modifier(pm, "stringplanmodifier.UseStateForUnknown()")
        string_modifier(pm, "stringplanmodifier.RequiresReplace()")

    # Billing address is optional on create and does not come back in Get.
    ba = attr_by_name(attrs, "billing_address")
    if ba:
        set_cor(ba, "computed_optional")
        object_modifier(ba, "objectplanmodifier.UseStateForUnknown()")
        object_modifier(ba, "objectplanmodifier.RequiresReplace()")

    # org_id is a query/body param; changing it forces a new resource.
    org = attr_by_name(attrs, "org_id")
    if org:
        set_cor(org, "computed_optional")
        string_modifier(org, "stringplanmodifier.RequiresReplace()")

    # All read-only response fields are computed.
    computed = {
        "id",
        "plan_id",
        "plan_name",
        "plan_type",
        "status",
        "current_period_end",
        "cancel_at_period_end",
        "stripe_customer_id",
        "stripe_subscription_id",
        "currency",
        "expiring_soon",
        "price",
        "trials_ending_soon",
    }
    for a in attrs:
        if a["name"] in computed:
            set_cor(a, "computed")


def fix_project_custom_domain(spec: dict) -> None:
    r = find_resource(spec, "project_custom_domain")
    if not r:
        return

    attrs = r["schema"]["attributes"]

    # Remove any project attributes that are not relevant to this subresource.
    keep = {"custom_domain", "org_id", "project_id", "id", "created_at", "updated_at"}
    attrs[:] = [a for a in attrs if a["name"] in keep]

    cd = attr_by_name(attrs, "custom_domain")
    if cd:
        set_cor(cd, "required")

    for name in ("org_id", "project_id"):
        a = attr_by_name(attrs, name)
        if a:
            set_cor(a, "required")
            string_modifier(a, "stringplanmodifier.RequiresReplace()")

    for name in ("id", "created_at", "updated_at"):
        a = attr_by_name(attrs, name)
        if a:
            set_cor(a, "computed")


def add_api_key(spec: dict) -> None:
    # The data source already has an api_keys list with the object schema.
    item_attrs = extract_nested_item_attributes("api_keys", "api_keys", spec)
    if not item_attrs:
        return

    attrs = copy.deepcopy(item_attrs)

    # Make org_id and name required, project_id and expiring_at optional.
    for a in attrs:
        name = a["name"]
        if name in ("org_id", "name"):
            set_cor(a, "required")
        elif name in ("project_id", "expiring_at"):
            set_cor(a, "computed_optional")
        elif name == "id":
            set_cor(a, "computed")
        else:
            set_cor(a, "computed")

    # secret is returned only on create; capture it as a sensitive computed value.
    secret = attr_by_name(attrs, "secret")
    if secret:
        set_cor(secret, "computed")
        secret["string"]["sensitive"] = True
        string_modifier(secret, "stringplanmodifier.UseStateForUnknown()")

    # Path/org changes force a new key; there is no update.
    for name in ("org_id", "name", "project_id", "expiring_at"):
        a = attr_by_name(attrs, name)
        if a:
            string_modifier(a, "stringplanmodifier.RequiresReplace()")
            if name in ("project_id", "expiring_at"):
                string_modifier(a, "stringplanmodifier.UseStateForUnknown()")

    add_resource(spec, "api_key", attrs)


def add_client_credential(spec: dict) -> None:
    item_attrs = extract_nested_item_attributes("client_credentials", "credentials", spec)
    if not item_attrs:
        return

    attrs = copy.deepcopy(item_attrs)

    # Add org_id which is used for create/list/revoke but not in list items.
    if not attr_by_name(attrs, "org_id"):
        attrs.insert(
            0,
            {
                "name": "org_id",
                "string": {"computed_optional_required": "computed_optional"},
            },
        )

    for a in attrs:
        name = a["name"]
        if name == "name":
            set_cor(a, "required")
        elif name in ("org_id", "alg"):
            set_cor(a, "computed_optional")
        elif name == "id":
            set_cor(a, "computed")
        else:
            set_cor(a, "computed")

    # client_secret is returned only at create time; capture it as sensitive.
    client_secret = {
        "name": "client_secret",
        "string": {
            "computed_optional_required": "computed",
            "sensitive": True,
        },
    }
    string_modifier(client_secret, "stringplanmodifier.UseStateForUnknown()")
    attrs.append(client_secret)

    for name in ("name", "org_id", "alg"):
        a = attr_by_name(attrs, name)
        if a:
            string_modifier(a, "stringplanmodifier.RequiresReplace()")
            if name in ("org_id", "alg"):
                string_modifier(a, "stringplanmodifier.UseStateForUnknown()")

    add_resource(spec, "client_credential", attrs)


def main() -> int:
    spec = load_spec()

    # Resources generated by tfplugingen-openapi.
    fix_subscription(spec)
    fix_project_custom_domain(spec)

    # Resources that must be built by hand because they are list-based.
    # Add these before stripping secrets from data sources so the resource
    # schemas can keep one-time secrets as sensitive computed attributes.
    add_api_key(spec)
    add_client_credential(spec)

    # Remove sensitive 'secret' attributes from data sources only.
    for ds in spec.get("datasources", []):
        ds["schema"]["attributes"] = remove_secret(ds["schema"]["attributes"])

    # Fix mangled attribute names produced by the OpenAPI generator.
    for r in spec.get("resources", []):
        rename_attributes(r["schema"]["attributes"], "useruser_id", "user_id")
        sanitize_timestamp_descriptions(r["schema"]["attributes"])
    for ds in spec.get("datasources", []):
        rename_attributes(ds["schema"]["attributes"], "useruser_id", "user_id")
        sanitize_timestamp_descriptions(ds["schema"]["attributes"])

    # Deduplicate resources.
    seen = set()
    deduped = []
    for r in spec.get("resources", []):
        if r["name"] not in seen:
            seen.add(r["name"])
            deduped.append(r)
    spec["resources"] = deduped

    save_spec(spec)
    print(f"Post-processed {SPEC}: {len(spec.get('resources', []))} resources.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
