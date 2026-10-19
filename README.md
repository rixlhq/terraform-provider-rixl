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

## Guides & Architecture Walkthroughs

Explore step-by-step guides on the [Terraform Registry](https://registry.terraform.io/providers/rixlhq/rixl/latest):

- [Getting Started with Rixl](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/getting-started) &mdash; Installation, authentication options, and provisioning your first project.
- [Automated Video Streaming Pipeline](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/video-streaming-pipeline) &mdash; End-to-end HLS video transcoding, custom CDN domain, multi-language subtitles, and audio tracks.
- [IAM, RBAC, and Security Automation](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/iam-and-security) &mdash; Access policies, machine-to-machine OAuth2 credentials, and enterprise SSO connections.
- [Observability, Telemetry, and Analytics](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/monitoring-and-analytics) &mdash; Dashboards, playback sessions, bandwidth/storage metrics, and heatmaps.

## Overview of Resources & Data Sources

| Category | Resources | Data Sources |
| :--- | :--- | :--- |
| **Video & Media** | `rixl_video`, `rixl_video_chapters`, `rixl_subtitle`, `rixl_audio_track`, `rixl_image` | `rixl_video`, `rixl_videos`, `rixl_video_stats`, `rixl_video_heatmap`, `rixl_video_hot_segments`, `rixl_top_videos`, `rixl_image`, `rixl_images`, `rixl_image_stats`, `rixl_top_images`, `rixl_chapters`, `rixl_subtitles`, `rixl_audio_tracks`, `rixl_languages` |
| **Projects & Content** | `rixl_project`, `rixl_feed`, `rixl_post`, `rixl_blog_subscription` | `rixl_project`, `rixl_projects`, `rixl_feed`, `rixl_feeds`, `rixl_feed_stats`, `rixl_top_feeds`, `rixl_post`, `rixl_posts`, `rixl_post_stats`, `rixl_top_posts`, `rixl_blog_subscription` |
| **Custom Domains** | `rixl_custom_domain`, `rixl_project_custom_domain`, `rixl_domain_auto_join` | `rixl_domain`, `rixl_domain_auto_join` |
| **Identity & Access Management (IAM)** | `rixl_access_policy`, `rixl_policy_attachment`, `rixl_api_key`, `rixl_client_credential`, `rixl_auth_provider_connection` | `rixl_policy`, `rixl_policies`, `rixl_policy_attachments`, `rixl_user_policies`, `rixl_api_keys`, `rixl_client_credentials`, `rixl_providers`, `rixl_passkeys`, `rixl_permission_registries` |
| **Organizations & Collaboration** | `rixl_organization`, `rixl_organization_invitation`, `rixl_organization_member` | `rixl_organization_members`, `rixl_memberships`, `rixl_membership_applications`, `rixl_check_membership`, `rixl_internal_membership_info`, `rixl_user`, `rixl_user_info` |
| **Dashboards & Analytics** | `rixl_dashboard`, `rixl_dashboard_widget` | `rixl_dashboard`, `rixl_dashboards`, `rixl_dashboard_datasets`, `rixl_dashboard_stats`, `rixl_realtime_stats`, `rixl_storage_usage`, `rixl_storage_usage_history`, `rixl_bandwidth_usage`, `rixl_bandwidth_usage_history` |
| **Billing & Subscriptions** | `rixl_subscription`, `rixl_payment_method`, `rixl_billing_address` | `rixl_subscription`, `rixl_subscription_history`, `rixl_plan`, `rixl_plans`, `rixl_invoices`, `rixl_payment_methods`, `rixl_payment_method_from_payment_intent`, `rixl_payment_method_from_setup_intent`, `rixl_billing_address` |

## Provider Configuration

Configure the provider with an API key or a Bearer token:

```terraform
terraform {
  required_providers {
    rixl = {
      source  = "rixlhq/rixl"
      version = "~> 1.7"
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

# 2. Register custom streaming domain and map to project
resource "rixl_custom_domain" "cdn" {
  org_id = var.rixl_org_id
  domain = "video.example.com"
}

resource "rixl_project_custom_domain" "cdn_mapping" {
  project_id = rixl_project.prod.id
  domain_id  = rixl_custom_domain.cdn.id
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
  project_id       = rixl_project.prod.id
  name             = "Annual Keynote"
  file_path        = "/media/keynote.mp4"
  poster_file_path = "/media/poster.jpg"
}

# 5. Attach navigation chapters to the transcoded video
resource "rixl_video_chapters" "keynote_chapters" {
  project_id = rixl_project.prod.id
  video_id   = rixl_video.keynote.id

  chapters = [
    {
      title      = "Opening Remarks"
      start_time = 0.0
    },
    {
      title      = "Product Demo"
      start_time = 62.5
    },
    {
      title      = "Roadmap & Closing"
      start_time = 184.0
    }
  ]
}
```

## Importing Existing Resources

Existing Rixl infrastructure can be imported into Terraform state using `terraform import`:

```bash
# Import a project by its unique ID
terraform import rixl_project.example "proj_01hxxxxxxxxx"

# Import a video by its unique ID
terraform import rixl_video.example "vid_01hxxxxxxxxx"

# Import a custom domain
terraform import rixl_custom_domain.example "dom_01hxxxxxxxxx"

# Import an access policy
terraform import rixl_access_policy.example "pol_01hxxxxxxxxx"
```

## Contributing & Development

```bash
# Run formatting check
mise run fmt-check

# Run unit tests
mise run test

# Run linter
mise run lint

# Generate and validate documentation
mise run docs
mise run docs-check
```

## License

Mozilla Public License v2.0 &mdash; see [LICENSE](LICENSE) for details.
