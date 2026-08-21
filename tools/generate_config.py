#!/usr/bin/env python3
"""Generate a comprehensive generator_config.yml from openapi.yaml.

The script maps every GET operation to a data source and discovers logical
entities with a POST (create) and a matching GET (read) for resources.
It skips user-auth, login, TOTP, password reset, internal, and media
endpoints that are implemented by hand-written files.
"""

from __future__ import annotations

import re
from pathlib import Path
from typing import Any

import yaml

ROOT = Path(__file__).resolve().parent.parent
OPENAPI = ROOT / "openapi.yaml"
CONFIG = ROOT / "generator_config.yml"

# Data source names whose generated files conflict with hand-written wrappers.
MANUAL_DS = {"image", "images", "video", "videos"}

# Resources that have hand-written wrappers and must not be regenerated.
MANUAL_RESOURCES = {"project", "dashboard", "access_policy"}

# Path/operation patterns to skip because they are not infrastructure resources.
SKIP_PATH_PREFIXES = {
    "/internal/",
}

SKIP_PATHS = {
    # user-auth / session / TOTP / password reset
    "/auth/v1/login",
    "/auth/v1/logout",
    "/auth/v1/register",
    "/auth/v1/token",
    "/auth/v1/password/reset",
    "/auth/v1/password/reset/confirm",
    "/auth/v1/verify-passkey",
    "/auth/v1/verify-totp",
    "/auth/v1/passkey/login/begin",
    "/auth/v1/passkey/login/finish",
    "/auth/v1/totp/backup-codes/regenerate",
    "/auth/v1/totp/setup",
    "/auth/v1/totp/verify",
    "/auth/v1/email/verify",
    "/auth/v1/email/verify/resend",
    "/auth/v1/invitations/{token}/accept",
    "/auth/v1/invitations/{token}/decline",
    "/auth/v1/providers/connect",
    "/auth/v1/blog/broadcast",
    "/auth/v1/blog/subscribe",
    "/auth/v1/blog/unsubscribe",
    "/auth/v1/blog/unsubscribe/email",
    "/auth/v1/memberships/{org_id}/domain",
    "/auth/v1/memberships/{org_id}/leave",
    "/auth/v1/memberships/{user.org_id}/domain",
    "/auth/v1/memberships/{user.org_id}/invite",
    "/auth/v1/memberships/{user.org_id}/invite/resend",
    "/auth/v1/users/current/emails",
    "/auth/v1/users/current/emails/status",
    "/auth/v1/users/current/passkeys/register/begin",
    "/auth/v1/users/current/passkeys/register/finish",
    "/auth/v1/users/current/totp/status",
    # operations that are POST-only analytics/not resources
    "/analytics/v1/events",
    "/analytics/v1/funnels",
    "/analytics/v1/retention",
    "/analytics/v1/dashboard/chart-query",
    "/analytics/v1/dashboard/chart-query/batch",
    "/analytics/v1/dashboard/filter-options",
    "/analytics/v1/dashboard/scope-tree",
    "/analytics/v1/dashboards/{dashboard_id}/layout",
    "/analytics/v1/dashboards/{dashboard_id}/widgets",
    "/analytics/v1/dashboards/{id}/default",
    "/billing/v1/bandwidth-usage/refresh",
    "/billing/v1/checkout",
    "/billing/v1/contact-sales",
    "/billing/v1/setup-intent",
    "/billing/v1/storage-usage/refresh",
    "/billing/v1/subscription/cancel",
    "/billing/v1/subscription/reactivate",
    "/billing/v1/subscription/upgrade",
    "/billing/v1/tax/calculate",
    "/billing/v1/tax/calculate-generic",
    "/platform/auth/v1/refresh",
    "/platform/auth/v1/token",
    "/platform/clientauth/v1/token",
    "/organizations/{org_id}/api-keys/v1/{key_id}/rotate",
    "/organizations/{org_id}/projects/v1/{project_id}/move",
    "/media/v1/projects/{project_id}/videos/{video_id}/audio-tracks/upload",
    "/media/v1/projects/{project_id}/videos/{video_id}/subtitles/upload",
    "/media/v1/projects/{project_id}/images/upload",
    "/media/v1/projects/{project_id}/videos/upload",
    "/media/v1/projects/{project_id}/videos/{video_id}/audio-tracks",
    "/media/v1/projects/{project_id}/videos/{video_id}/subtitles",
    "/media/v1/projects/{project_id}/videos/{video_id}/audio-tracks/{track_id}",
    "/media/v1/projects/{project_id}/videos/{video_id}/subtitles/{subtitle_id}",
    "/media/v1/projects/{project_id}/videos/{video_id}/audio-tracks/language/{language_code}",
    "/media/v1/projects/{project_id}/videos/{video_id}/subtitles/language/{language_code}",
    "/media/v1/projects/{project_id}/videos/{video_id}/visibility",
    "/media/v1/projects/{project_id}/images/{image_id}/visibility",
    "/internal/images/mark-failed",
    "/internal/images/mark-processed",
    "/internal/images/take",
    "/internal/videos/mark-failed",
    "/internal/videos/mark-processed",
    # posts use allOf/oneOf which the OpenAPI generator does not support
    "/posts/v1/feeds/{feed_id}",
    "/posts/v1/feeds/{feed_id}/creators/{creator_id}",
    "/posts/v1/feeds/{feed_id}/{post_id}",
    "/posts/v1/projects/{project_id}/feeds/{feed_id}/posts",
    "/posts/v1/projects/{project_id}/feeds/{feed_id}/posts/creators/{creator_id}",
    "/posts/v1/projects/{project_id}/feeds/{feed_id}/posts/{post_id}",
    "/posts/v1/projects/{project_id}/feeds/{feed_id}/posts/upload",
    "/posts/v1/projects/{project_id}/posts/{post_id}",
}

# Hardcoded overrides for data source names that cannot be derived cleanly.
DS_NAME_OVERRIDES: dict[str, str] = {
    "/analytics/v1/dashboard": "dashboard_stats",
    "/analytics/v1/dashboard/datasets": "dashboard_datasets",
    "/analytics/v1/dashboards": "dashboards",
    "/analytics/v1/dashboards/{id}": "dashboard",
    "/analytics/v1/feeds/{feed_id}/stats": "feed_stats",
    "/analytics/v1/images/{image_id}/stats": "image_stats",
    "/analytics/v1/posts/{post_id}/stats": "post_stats",
    "/analytics/v1/realtime": "realtime_stats",
    "/analytics/v1/top/feeds": "top_feeds",
    "/analytics/v1/top/images": "top_images",
    "/analytics/v1/top/posts": "top_posts",
    "/analytics/v1/top/videos": "top_videos",
    "/analytics/v1/videos/{video_id}/heatmap": "video_heatmap",
    "/analytics/v1/videos/{video_id}/hot-segments": "video_hot_segments",
    "/analytics/v1/videos/{video_id}/stats": "video_stats",
    "/auth/v1/membership-applications": "membership_applications",
    "/auth/v1/memberships": "memberships",
    "/auth/v1/memberships/{org_id}/check": "check_membership",
    "/auth/v1/memberships/{org_id}/domain": "domain_status",
    "/auth/v1/memberships/{org_id}/domain/auto-join": "domain_auto_join",
    "/auth/v1/memberships/{org_id}/info": "internal_membership_info",
    "/auth/v1/memberships/{org_id}/policies": "policies",
    "/auth/v1/memberships/{user.org_id}/members": "organization_members",
    "/auth/v1/memberships/{user.org_id}/members/{member_id}/policies": "user_policies",
    "/auth/v1/memberships/{user.org_id}/policies/permissions": "permission_registries",
    "/auth/v1/memberships/{user.org_id}/policies/{policy_id}": "policy",
    "/auth/v1/memberships/{user.org_id}/policies/{policy_id}/attachments": "policy_attachments",
    "/auth/v1/providers": "providers",
    "/auth/v1/userinfo": "user_info",
    "/auth/v1/users/current": "user",
    "/auth/v1/users/current/passkeys": "passkeys",
    "/auth/v1/users/current/totp/status": "totp_status",
    "/billing/v1/address": "billing_address",
    "/billing/v1/bandwidth-usage": "bandwidth_usage",
    "/billing/v1/bandwidth-usage/history": "bandwidth_usage_history",
    "/billing/v1/invoices": "invoices",
    "/billing/v1/payment-methods": "payment_methods",
    "/billing/v1/payment-methods/from-payment-intent": "payment_method_from_payment_intent",
    "/billing/v1/payment-methods/from-setup-intent": "payment_method_from_setup_intent",
    "/billing/v1/plans": "plans",
    "/billing/v1/plans/{plan_id}": "plan",
    "/billing/v1/storage-usage": "storage_usage",
    "/billing/v1/storage-usage/history": "storage_usage_history",
    "/billing/v1/subscription": "subscription",
    "/billing/v1/subscription/history": "subscription_history",
    "/feeds/v1/projects/{project_id}/feeds": "feeds",
    "/feeds/v1/projects/{project_id}/feeds/{feed_id}": "feed",
    "/media/v1/images/{image_id}": "image",
    "/media/v1/languages": "languages",
    "/media/v1/projects/{project_id}/images": "images",
    "/media/v1/projects/{project_id}/videos": "videos",
    "/media/v1/projects/{project_id}/videos/{video_id}/chapters": "chapters",
    "/media/v1/videos/{video_id}": "video",
    "/media/v1/videos/{video_id}/audio-tracks": "audio_tracks",
    "/media/v1/videos/{video_id}/subtitles": "subtitles",
    "/organizations/{org_id}/api-keys/v1": "api_keys",
    "/organizations/{org_id}/projects/v1": "projects",
    "/organizations/{org_id}/projects/v1/{project_id}": "project",
    "/platform/clientauth/v1/credentials": "client_credentials",
    "/auth/v1/blog/subscription": "blog_subscription",
}

# Resource definitions discovered from the OpenAPI operations.
RESOURCE_OVERRIDES: dict[str, dict[str, Any]] = {
    "feed": {
        "create": {"path": "/feeds/v1/projects/{project_id}/feeds", "method": "POST"},
        "read": {"path": "/feeds/v1/projects/{project_id}/feeds/{feed_id}", "method": "GET"},
        "update": {"path": "/feeds/v1/projects/{project_id}/feeds/{feed_id}", "method": "PUT"},
        "delete": {"path": "/feeds/v1/projects/{project_id}/feeds/{feed_id}", "method": "DELETE"},
        "schema": {
            "attributes": {
                "aliases": {"feed_id": "id"},
                "overrides": {
                    "project_id": {"computed_optional_required": "required"},
                    "id": {"computed_optional_required": "required"},
                },
            }
        },
    },
}


def to_snake(name: str) -> str:
    """Convert a PascalCase string to snake_case."""
    name = re.sub(r"([A-Z]+)([A-Z][a-z])", r"\1_\2", name)
    name = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", name)
    return re.sub(r"[^a-z0-9_]+", "_", name.lower()).strip("_")


def derive_ds_name(path: str, opid: str, tags: list[str]) -> str:
    """Derive a snake_case data source name from the operation or path."""
    if path in DS_NAME_OVERRIDES:
        return DS_NAME_OVERRIDES[path]

    # Use the last segment of the operationId.
    last = opid.split(".")[-1]
    # Drop leading Get/List and trailing numeric disambiguation.
    if last.startswith("Get"):
        last = last[3:]
    elif last.startswith("List"):
        last = last[4:]
    last = re.sub(r"\d+$", "", last)
    base = to_snake(last)

    # If the name is too generic, prefix with the (first) tag to avoid collisions.
    if base in ("stats", "datasets", "status", "info", "feed", "image", "video"):
        tag = to_snake(tags[0]) if tags else ""
        if tag:
            base = f"{tag}_{base}"
    return base


def is_skipped_path(path: str) -> bool:
    if any(path.startswith(p) for p in SKIP_PATH_PREFIXES):
        return True
    if path in SKIP_PATHS:
        return True
    return False


def path_params(operation: dict[str, Any]) -> list[str]:
    return [p["name"] for p in operation.get("parameters", []) if p.get("in") == "path"]


def param_attribute_name(name: str) -> str:
    """Map dotted or path parameter names to Terraform attribute names."""
    if name == "user.org_id":
        return "org_id"
    return name


def build_data_source_schema(path: str, opid: str, operation: dict[str, Any]) -> dict[str, Any]:
    """Build schema aliases/overrides for a data source based on path params."""
    params = path_params(operation)
    aliases: dict[str, str] = {}
    overrides: dict[str, Any] = {}

    # Only the last path component being a {param} indicates a singular resource
    # read (e.g. .../{feed_id}). For collection or action reads (e.g.
    # .../{project_id}/feeds or .../{post_id}/stats) the final path component
    # is a literal, and the path params are parent ids.
    is_instance = path.rstrip("/").rsplit("/", 1)[-1].startswith("{")
    if is_instance and params:
        final = params[-1]
        attr = param_attribute_name(final)
        if attr != final:
            aliases[final] = attr
        if attr == "id":
            overrides["id"] = {"computed_optional_required": "required"}
        else:
            # Entity id parameters like feed_id are exposed as the canonical id.
            if attr.endswith("_id"):
                aliases[final] = "id"
            overrides["id"] = {"computed_optional_required": "required"}

    # Parent path parameters are required inputs.
    if is_instance and params:
        parents = params[:-1]
    else:
        parents = params
    for p in parents:
        attr = param_attribute_name(p)
        if attr != p:
            aliases[p] = attr
        overrides[attr] = {"computed_optional_required": "required"}

    schema: dict[str, Any] = {"attributes": {}}
    if aliases:
        schema["attributes"]["aliases"] = aliases
    if overrides:
        schema["attributes"]["overrides"] = overrides
    return schema


def main() -> None:
    spec = yaml.safe_load(OPENAPI.read_text())

    data_sources: dict[str, dict[str, Any]] = {}
    for path, ops in spec["paths"].items():
        get_op = ops.get("get")
        if not get_op:
            continue
        if is_skipped_path(path):
            continue

        opid = get_op.get("operationId", "")
        tags = get_op.get("tags", [])
        name = derive_ds_name(path, opid, tags)

        if name in MANUAL_DS:
            continue
        if name in data_sources:
            # Disambiguate by prefixing the first tag/service name.
            tag = to_snake(tags[0]) if tags else "service"
            name = f"{tag}_{name}"

        # Final duplicate guard.
        base = name
        counter = 2
        while name in data_sources:
            name = f"{base}_{counter}"
            counter += 1

        data_sources[name] = {
            "read": {"path": path, "method": "GET"},
            "schema": build_data_source_schema(path, opid, get_op),
        }

    resources = dict(RESOURCE_OVERRIDES)

    config = {
        "provider": {"name": "rixl"},
        "resources": resources,
        "data_sources": data_sources,
    }

    with CONFIG.open("w") as f:
        yaml.safe_dump(config, f, sort_keys=False, allow_unicode=True)

    print(f"Wrote {CONFIG} with {len(data_sources)} data sources and {len(resources)} resources.")


if __name__ == "__main__":
    main()
