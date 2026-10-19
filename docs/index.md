---
page_title: "Provider: Rixl - Cloud Video Streaming, HLS Transcoding & Media Infrastructure"
description: |-
  Official Terraform provider for Rixl: declare and automate video streaming pipelines, multi-bitrate HLS transcoding, media feeds, custom domains, IAM access policies, API keys, OAuth client credentials, analytics dashboards, and billing subscriptions via the Rixl REST API.
---

# Rixl Provider

The **Rixl Terraform Provider** enables platform engineers, media architects, and DevOps teams to declare, deploy, and manage video streaming pipelines and media infrastructure as code on [Rixl](https://rixl.com).

Rixl delivers high-performance, developer-first video streaming with automated multi-bitrate HLS (HTTP Live Streaming) transcoding, multi-track audio, localized subtitles (WebVTT/SRT), timecoded chapters, image processing, content syndication feeds, granular IAM policies, and real-time observability telemetry.

Built on the modern **Terraform Plugin Framework** (protocol version 6), the provider supports strongly typed models, comprehensive schema validation, sensitive credential protection, and automatic retry handling for asynchronous media processing.

## Guides & Architectural Walkthroughs

Explore step-by-step guides covering core video streaming and platform automation workflows:

- [Getting Started with Rixl](guides/getting-started) &mdash; Installation, authentication options (API Key vs OAuth2 Bearer), and provisioning your first project.
- [Automated Video Streaming Pipeline](guides/video-streaming-pipeline) &mdash; End-to-end HLS video transcoding, multi-track audio, localized subtitles, and navigation chapters.
- [Custom Domains and Auto-Join Management](guides/custom-domains) &mdash; Organization domain verification, project domain associations, and corporate email auto-join policies.
- [Content Feeds, Playlists, and Media Distribution](guides/content-feeds-and-distribution) &mdash; Structured video channels, curated playlists, rich media posts, image publishing, and automated blog notifications.
- [IAM, RBAC, and Security Automation](guides/iam-and-security) &mdash; Least-privilege access policies, machine-to-machine OAuth2 credentials, API key rotation, and enterprise SSO connections.
- [Observability, Telemetry, and Analytics](guides/monitoring-and-analytics) &mdash; Custom telemetry dashboards, audience retention heatmaps, hot segments, bandwidth consumption, and storage metrics.

## Overview of Resources & Data Sources

| Category | Description | Resources | Data Sources |
| :--- | :--- | :--- | :--- |
| **Video & Media** | Video asset uploading, multi-bitrate HLS transcoding, chapters, multi-language subtitles, alternative audio tracks, and optimized images | `rixl_video`, `rixl_video_chapters`, `rixl_subtitle`, `rixl_audio_track`, `rixl_image` | `rixl_video`, `rixl_videos`, `rixl_video_stats`, `rixl_video_heatmap`, `rixl_video_hot_segments`, `rixl_top_videos`, `rixl_image`, `rixl_images`, `rixl_image_stats`, `rixl_top_images`, `rixl_chapters`, `rixl_subtitles`, `rixl_audio_tracks`, `rixl_languages` |
| **Projects & Content** | Isolated environments, content feeds (channels, playlists), posts, and blog subscriptions | `rixl_project`, `rixl_feed`, `rixl_post`, `rixl_blog_subscription` | `rixl_project`, `rixl_projects`, `rixl_feed`, `rixl_feeds`, `rixl_feed_stats`, `rixl_top_feeds`, `rixl_post`, `rixl_posts`, `rixl_post_stats`, `rixl_top_posts`, `rixl_blog_subscription` |
| **Custom Domains** | Organization custom domains, project domain associations, and email auto-join policies | `rixl_custom_domain`, `rixl_project_custom_domain`, `rixl_domain_auto_join` | `rixl_domain`, `rixl_domain_auto_join` |
| **Identity & Access Management (IAM)** | Granular RBAC access policies, policy attachments, API keys, OAuth client credentials, and SSO | `rixl_access_policy`, `rixl_policy_attachment`, `rixl_api_key`, `rixl_client_credential`, `rixl_auth_provider_connection` | `rixl_policy`, `rixl_policies`, `rixl_policy_attachments`, `rixl_user_policies`, `rixl_api_keys`, `rixl_client_credentials`, `rixl_providers`, `rixl_passkeys`, `rixl_permission_registries` |
| **Organizations & Collaboration** | Multi-tenant organization structure, invitations, and member access | `rixl_organization`, `rixl_organization_invitation`, `rixl_organization_member` | `rixl_organization_members`, `rixl_memberships`, `rixl_membership_applications`, `rixl_check_membership`, `rixl_internal_membership_info`, `rixl_user`, `rixl_user_info` |
| **Dashboards & Analytics** | Custom observability dashboards, widgets, bandwidth usage, and storage tracking | `rixl_dashboard`, `rixl_dashboard_widget` | `rixl_dashboard`, `rixl_dashboards`, `rixl_dashboard_datasets`, `rixl_dashboard_stats`, `rixl_realtime_stats`, `rixl_storage_usage`, `rixl_storage_usage_history`, `rixl_bandwidth_usage`, `rixl_bandwidth_usage_history` |
| **Billing & Subscriptions** | Subscription plan management, payment methods, and billing addresses | `rixl_subscription`, `rixl_payment_method`, `rixl_billing_address` | `rixl_subscription`, `rixl_subscription_history`, `rixl_plan`, `rixl_plans`, `rixl_invoices`, `rixl_payment_methods`, `rixl_payment_method_from_payment_intent`, `rixl_payment_method_from_setup_intent`, `rixl_billing_address` |

## Example Usage

### API Key Authentication (Recommended for CI/CD and Automation)

```terraform
terraform {
  required_providers {
    rixl = {
      source  = "rixlhq/rixl"
      version = "~> 1.8"
    }
  }
}

provider "rixl" {
  api_key = var.rixl_api_key # or export RIXL_API_KEY
}

# 1. Create a project
resource "rixl_project" "prod" {
  org_id = var.rixl_org_id
  name   = "production-streaming"
}

# 2. Configure an organization custom domain
resource "rixl_custom_domain" "company_domain" {
  org_id = var.rixl_org_id
  domain = "media.example.com"
}

resource "rixl_project_custom_domain" "domain_mapping" {
  org_id        = var.rixl_org_id
  project_id    = rixl_project.prod.id
  custom_domain = rixl_custom_domain.company_domain.domain
}

# 3. Create a content feed
resource "rixl_feed" "main" {
  project_id   = rixl_project.prod.id
  name         = "main-stream"
  description  = "Featured video feed"
  allow_videos = true
  allow_images = true
}

# 4. Upload and transcode a video
resource "rixl_video" "intro" {
  project_id       = rixl_project.prod.id
  name             = "Product Intro"
  file_path        = "${path.module}/media/intro.mp4"
  poster_file_path = "${path.module}/media/poster.jpg"
}

# 5. Attach navigation chapters to the video
resource "rixl_video_chapters" "intro_chapters" {
  project_id = rixl_project.prod.id
  video_id   = rixl_video.intro.id

  chapters = [
    {
      title          = "Introduction"
      start_time_sec = 0.0
    },
    {
      title          = "Overview"
      start_time_sec = 30.5
    },
    {
      title          = "Next Steps"
      start_time_sec = 90.0
    },
  ]
}
```

### Dynamic OAuth2 Bearer Token Authentication

```terraform
provider "rixl" {
  bearer_token = var.rixl_bearer_token # or export RIXL_BEARER_TOKEN
}
```

## Schema

<!-- schema generated by tfplugindocs -->
<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) Rixl API key for authenticating requests. May also be set via the `RIXL_API_KEY` environment variable.
- `base_url` (String) Rixl API base URL. Defaults to `https://api.rixl.com`. May also be set via the `RIXL_BASE_URL` environment variable.
- `bearer_token` (String, Sensitive) Rixl OAuth / Bearer token for authenticating requests. May also be set via the `RIXL_BEARER_TOKEN` environment variable.
