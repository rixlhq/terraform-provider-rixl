# Terraform Provider for Rixl

[![Terraform Registry](https://img.shields.io/badge/Terraform%20Registry-rixlhq%2Frixl-844FBA?logo=terraform&logoColor=white)](https://registry.terraform.io/providers/rixlhq/rixl/latest)
[![Release](https://img.shields.io/github/v/release/rixlhq/terraform-provider-rixl?sort=semver)](https://github.com/rixlhq/terraform-provider-rixl/releases)
[![CI](https://github.com/rixlhq/terraform-provider-rixl/actions/workflows/ci.yml/badge.svg)](https://github.com/rixlhq/terraform-provider-rixl/actions/workflows/ci.yml)
[![License: MPL-2.0](https://img.shields.io/github/license/rixlhq/terraform-provider-rixl)](LICENSE)

A [Terraform](https://www.terraform.io) provider for managing cloud media infrastructure and developer services at [Rixl](https://rixl.com) via official Rixl REST APIs:

- **Video & Audio Streaming** — upload video and audio assets, configure automated multi-bitrate HLS transcoding, manage timecoded video chapters, multi-language subtitles (VTT/SRT), and audio tracks.
- **Images & Content Feeds** — upload image media with automatic responsive thumbnail generation, build structured content feeds, and publish rich media posts.
- **Projects & Custom Domains** — segregate environments with projects and configure custom CDN streaming domains with automated SSL/TLS certificate verification.
- **Identity & Access Management (IAM)** — define granular RBAC access policies, attach policies to members or service accounts, issue long-lived API keys, and manage OAuth 2.0 client credentials.
- **Organizations & Collaboration** — manage organization membership, invitations, and enterprise SSO auth provider connections.
- **Observability & Analytics** — deploy real-time analytics dashboards and widgets tracking video heatmaps, hot segments, bandwidth consumption, and storage metrics.
- **Billing & Subscriptions** — automate subscription tier management, attach payment methods, and maintain billing addresses.

Built with the modern **Terraform Plugin Framework** (protocol version 6), featuring strongly-typed schemas, robust validation, sensitive attribute protection, and automatic retry handling for asynchronous media processing.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.26
- [mise](https://mise.jdx.dev/) (recommended for task execution and toolchain management)

## Overview of Resources & Data Sources

| Category | Resources | Data Sources |
| :--- | :--- | :--- |
| **Media & Video Streaming** | `rixl_video`, `rixl_video_chapters`, `rixl_subtitle`, `rixl_audio_track`, `rixl_image` | `rixl_video`, `rixl_videos`, `rixl_video_stats`, `rixl_video_heatmap`, `rixl_video_hot_segments`, `rixl_top_videos`, `rixl_image`, `rixl_images`, `rixl_image_stats`, `rixl_top_images`, `rixl_chapters`, `rixl_subtitles`, `rixl_audio_tracks` |
| **Feeds & Content Posts** | `rixl_feed`, `rixl_post`, `rixl_blog_subscription` | `rixl_feed`, `rixl_feeds`, `rixl_feed_stats`, `rixl_top_feeds`, `rixl_post`, `rixl_posts`, `rixl_post_stats`, `rixl_top_posts`, `rixl_blog_subscription` |
| **Projects & Custom Domains** | `rixl_project`, `rixl_custom_domain`, `rixl_project_custom_domain`, `rixl_domain_auto_join` | `rixl_project`, `rixl_projects`, `rixl_domain`, `rixl_domain_auto_join` |
| **Identity & Access Management (IAM)** | `rixl_access_policy`, `rixl_policy_attachment`, `rixl_api_key`, `rixl_client_credential`, `rixl_auth_provider_connection` | `rixl_policy`, `rixl_policies`, `rixl_policy_attachments`, `rixl_user_policies`, `rixl_api_keys`, `rixl_client_credentials`, `rixl_providers`, `rixl_passkeys`, `rixl_permission_registries` |
| **Organizations & Members** | `rixl_organization`, `rixl_organization_invitation`, `rixl_organization_member` | `rixl_organization_members`, `rixl_memberships`, `rixl_membership_applications`, `rixl_check_membership`, `rixl_internal_membership_info`, `rixl_user`, `rixl_user_info` |
| **Analytics & Dashboards** | `rixl_dashboard`, `rixl_dashboard_widget` | `rixl_dashboard`, `rixl_dashboards`, `rixl_dashboard_datasets`, `rixl_dashboard_stats`, `rixl_realtime_stats`, `rixl_storage_usage`, `rixl_storage_usage_history`, `rixl_bandwidth_usage`, `rixl_bandwidth_usage_history` |
| **Billing & Payments** | `rixl_subscription`, `rixl_payment_method`, `rixl_billing_address` | `rixl_subscription`, `rixl_subscription_history`, `rixl_plan`, `rixl_plans`, `rixl_invoices`, `rixl_payment_methods`, `rixl_payment_method_from_payment_intent`, `rixl_payment_method_from_setup_intent`, `rixl_billing_address` |

## Provider Configuration

Configure the provider with an API key or a Bearer token:

```terraform
terraform {
  required_providers {
    rixl = {
      source  = "rixlhq/rixl"
      version = "~> 1.0"
    }
  }
}

# API Key Authentication (Recommended)
provider "rixl" {
  api_key = var.rixl_api_key # or set RIXL_API_KEY environment variable
}
```

Or using an OAuth Bearer token:

```terraform
provider "rixl" {
  bearer_token = var.rixl_bearer_token # or set RIXL_BEARER_TOKEN environment variable
}
```

### Environment Variables

You can provide credentials and configuration via environment variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `RIXL_API_KEY` | Rixl API key (`X-API-Key` header) | — |
| `RIXL_BEARER_TOKEN` | Rixl OAuth 2.0 / Bearer token (`Authorization: Bearer` header) | — |
| `RIXL_BASE_URL` | Rixl API endpoint | `https://api.rixl.com` |

## Example Usage

### End-to-End Media Delivery Pipeline

The following example configures a streaming project with a custom CDN domain, a content feed, an uploaded video with adaptive transcoding, and navigation chapters:

```terraform
# 1. Create a project to isolate assets
resource "rixl_project" "prod" {
  org_id = var.rixl_org_id
  name   = "production-streaming"
}

# 2. Configure a verified custom CDN streaming domain
resource "rixl_project_custom_domain" "cdn" {
  org_id        = var.rixl_org_id
  project_id    = rixl_project.prod.id
  custom_domain = "video.example.com"
}

# 3. Create a content feed
resource "rixl_feed" "main" {
  project_id   = rixl_project.prod.id
  name         = "main-stream"
  description  = "Featured video feed"
  allow_videos = true
  allow_images = true
}

# 4. Upload a video asset and initiate transcoding
resource "rixl_video" "keynote" {
  project_id = rixl_project.prod.id
  name       = "Annual Keynote"
  file_path  = "/media/keynote.mp4"
}

# 5. Attach navigation chapters to the transcoded video
resource "rixl_video_chapters" "keynote_chapters" {
  video_id = rixl_video.keynote.id

  chapters = [
    {
      title      = "Opening Remarks"
      start_time = 0.0
    },
    {
      title      = "Platform Architecture"
      start_time = 320.5
    },
    {
      title      = "Q&A"
      start_time = 1800.0
    },
  ]
}
```

### IAM & Service Automation

Provision an IAM policy and issue an automation API key scoped to a specific project:

```terraform
resource "rixl_access_policy" "editor_policy" {
  org_id      = var.rixl_org_id
  name        = "media-editor"
  description = "Grants write permissions to videos and feeds"
}

resource "rixl_api_key" "ci_uploader" {
  org_id     = var.rixl_org_id
  project_id = rixl_project.prod.id
  name       = "github-actions-uploader"
}

output "api_key_secret" {
  value     = rixl_api_key.ci_uploader.secret
  sensitive = true
}
```

## Importing Existing Infrastructure

Existing Rixl resources can be imported into Terraform state using standard Terraform import syntax:

```shell
# Import a video by its ID
terraform import rixl_video.keynote video_01h7abc123def

# Import a project by its ID
terraform import rixl_project.prod proj_01h7abc456ghi

# Import an API key with org_id/key_id
terraform import rixl_api_key.ci_uploader org_123/key_456

# Import a project custom domain with org_id/project_id
terraform import rixl_project_custom_domain.cdn org_123/proj_456
```

## Development

This repository uses [mise](https://mise.jdx.dev/) for toolchain version pinning and task running. Git hooks are managed by [lefthook](https://github.com/evilmartians/lefthook).

```shell
# Install tools and git hooks
mise install

# Run full local checks (formatting, compilation, unit tests, and linting)
mise run

# Run individual workflows
mise run fmt       # format Go source code
mise run build     # compile provider binary
mise run test      # execute unit tests with coverage
mise run lint      # run golangci-lint
mise run docs      # generate Terraform Registry documentation
```

## Publishing to the Terraform Registry

The repository is configured for automated release publishing to the [Terraform Registry](https://registry.terraform.io/providers/rixlhq/rixl):

- The provider address is `registry.terraform.io/rixlhq/rixl`.
- `terraform-registry-manifest.json` specifies protocol version `6.0`.
- Continuous Integration (`.github/workflows/ci.yml`) runs tests, formatting, and lint checks against Go 1.26.
- Release workflows (`.github/workflows/release.yml`) automate semantic versioning via `release-please` and cross-platform binary signing with GoReleaser.
