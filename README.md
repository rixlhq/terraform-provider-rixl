# Terraform Provider for Rixl: Video Streaming, HLS Transcoding & Media Cloud Infrastructure

[![Terraform Registry](https://img.shields.io/badge/Terraform%20Registry-rixlhq%2Frixl-844FBA?logo=terraform&logoColor=white)](https://registry.terraform.io/providers/rixlhq/rixl/latest)
[![Release](https://img.shields.io/github/v/release/rixlhq/terraform-provider-rixl?sort=semver)](https://github.com/rixlhq/terraform-provider-rixl/releases)
[![CI](https://github.com/rixlhq/terraform-provider-rixl/actions/workflows/ci.yml/badge.svg)](https://github.com/rixlhq/terraform-provider-rixl/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rixlhq/terraform-provider-rixl)](https://goreportcard.com/report/github.com/rixlhq/terraform-provider-rixl)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-registry-blue?logo=read-the-docs&logoColor=white)](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs)

The **Rixl Terraform Provider** enables platform engineers, media architects, and DevOps teams to declare, deploy, and manage video streaming pipelines and media infrastructure as code on [Rixl](https://rixl.com).

Automate video ingestion, multi-bitrate HLS (HTTP Live Streaming) transcoding, multi-track audio, localized subtitles (WebVTT/SRT), timecoded chapters, image processing, custom domains, fine-grained IAM policies, and real-time observability telemetry.

---

## Table of Contents

- [Architecture Diagram](#architecture-diagram)
- [Key Capabilities](#key-capabilities)
- [Official Architectural Guides](#official-architectural-guides)
- [Provider Installation & Quickstart](#provider-installation--quickstart)
- [Authentication Options](#authentication-options)
- [Production Terraform Recipes](#production-terraform-recipes)
  - [1. End-to-End Automated Video Streaming Pipeline](#1-end-to-end-automated-video-streaming-pipeline)
  - [2. Custom Domain & Auto-Join Setup](#2-custom-domain--auto-join-setup)
  - [3. Content Feeds, Playlists & Media Posts](#3-content-feeds-playlists--media-posts)
  - [4. Least-Privilege IAM Access Policies & API Keys](#4-least-privilege-iam-access-policies--api-keys)
  - [5. Observability Dashboard & Viewer Telemetry](#5-observability-dashboard--viewer-telemetry)
- [Resource & Data Source Reference](#resource--data-source-reference)
- [Importing Existing Infrastructure](#importing-existing-infrastructure)
- [Frequently Asked Questions (FAQ)](#frequently-asked-questions-faq)
- [Development & Contributing](#development--contributing)

---

## Architecture Diagram

```mermaid
graph TD
    subgraph Infrastructure_as_Code [Infrastructure as Code]
        TF[Terraform CLI / OpenTofu]
        PRV[Terraform Provider for Rixl]
        TF -->|Declares HCL| PRV
    end

    subgraph Rixl_Cloud_Platform [Rixl Video Cloud Platform]
        API[Rixl Control Plane API]
        PRV -->|HTTPS REST API| API
        
        subgraph Media_Pipeline [Video Streaming Engine]
            Ingest[Video Upload / Ingestion]
            Transcoder[Multi-Bitrate HLS Transcoding]
            Subtitles[Localized Subtitles VTT/SRT]
            AudioTracks[Alternate Audio Dubs & Commentary]
            Chapters[Navigation Chapters]
            Ingest --> Transcoder
            Transcoder --> Subtitles
            Transcoder --> AudioTracks
            Transcoder --> Chapters
        end

        subgraph Content_Domains [Content & Custom Domains]
            Feeds[Content Feeds & Posts]
            Domains[Verified Custom Domains]
        end

        subgraph Security_Observability [IAM & Telemetry]
            IAM[Role-Based Access Policies & API Keys]
            Dashboards[Observability Dashboards & Realtime Stats]
        end

        API --> Media_Pipeline
        API --> Content_Domains
        API --> Security_Observability
    end

    subgraph Client_Players [Video Playback]
        HLSPlayer[Web & Mobile Players: Video.js, HLS.js, AVPlayer, ExoPlayer]
        Transcoder -->|HLS Stream| HLSPlayer
    end
```

---

## Key Capabilities

- **Automated Multi-Bitrate HLS Encoding**: Upload raw source MP4/MOV files and automatically receive adaptive bitrate ladders (1080p, 720p, 480p, 360p) with instant poster thumbnail generation.
- **Accessible & Internationalized Streaming**: Add multi-language WebVTT/SRT subtitles and alternate audio dubs or director commentary tracks declaratively.
- **Interactive Player Navigation**: Attach timecoded chapter markers for timeline scrubbing and topic-based seeking.
- **Branded Custom Domains**: Associate verified custom domains with projects and configure corporate email domain auto-join policies.
- **Content Syndication Feeds**: Group videos into structured feeds, curated playlists, and multimedia posts with audience notifications.
- **Enterprise IAM & Security**: Define least-privilege role-based access policies (RBAC), machine-to-machine OAuth2 client credentials, and long-lived API keys.
- **Observability & Analytics**: Deploy custom operational dashboards with timeseries charts, retention heatmaps, hot segments, and bandwidth tracking.

---

## Official Architectural Guides

Read our comprehensive operational walkthroughs in the [Terraform Registry](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs):

- [**Getting Started with Rixl**](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/getting-started): Initial setup, provider configuration, authentication, and provisioning your first project.
- [**Automated Video Streaming Pipeline**](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/video-streaming-pipeline): End-to-end HLS video transcoding, audio tracks, subtitles, chapters, and player integration.
- [**Custom Domains & Auto-Join Management**](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/custom-domains): Organization domain verification, project domain associations, and corporate email auto-join policies.
- [**Content Feeds, Playlists & Media Distribution**](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/content-feeds-and-distribution): Structured video channels, curated playlists, rich media posts, and blog notifications.
- [**IAM, RBAC, and Security Automation**](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/iam-and-security): Principle of least privilege, machine-to-machine OAuth2 credentials, and enterprise SSO connections.
- [**Observability, Telemetry, and Analytics**](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/guides/monitoring-and-analytics): Dashboards, viewer retention heatmaps, hot segments, and bandwidth monitoring.

---

## Provider Installation & Quickstart

```terraform
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    rixl = {
      source  = "rixlhq/rixl"
      version = "~> 1.8.0"
    }
  }
}

provider "rixl" {
  api_key = var.rixl_api_key # Recommended: export RIXL_API_KEY="..."
}
```

---

## Authentication Options

### 1. Static API Key Authentication (CI/CD and Admin)

```bash
export RIXL_API_KEY="rixl_live_01hxxxxxxxxxxxxxxxxxxxxxxx"
```

```terraform
provider "rixl" {}
```

### 2. OAuth2 Bearer Token Authentication (Machine-to-Machine)

```bash
export RIXL_BEARER_TOKEN="eyJhbGciOi..."
```

```terraform
provider "rixl" {
  bearer_token = var.rixl_bearer_token
}
```

---

## Production Terraform Recipes

### 1. End-to-End Automated Video Streaming Pipeline

Upload master video files, attach localized subtitles, add secondary audio dubs, and define interactive chapter cues:

```terraform
variable "org_id" {
  type        = string
  description = "Rixl Organization ID"
}

# Project workspace
resource "rixl_project" "vod_prod" {
  org_id = var.org_id
  name   = "production-vod-pipeline"
}

# Master video upload
resource "rixl_video" "keynote" {
  project_id       = rixl_project.vod_prod.id
  name             = "annual-developer-keynote"
  file_path        = "${path.module}/media/keynote_master.mp4"
  poster_file_path = "${path.module}/media/keynote_poster.jpg"
}

# Localized closed captions (WebVTT)
resource "rixl_subtitle" "keynote_sub_es" {
  project_id    = rixl_project.vod_prod.id
  video_id      = rixl_video.keynote.id
  label         = "Español"
  language_code = "es"
  file_path     = "${path.module}/subtitles/es.vtt"
}

# Multi-track audio commentary
resource "rixl_audio_track" "keynote_audio_commentary" {
  project_id    = rixl_project.vod_prod.id
  video_id      = rixl_video.keynote.id
  label         = "Director Commentary"
  language_code = "en"
  file_path     = "${path.module}/audio/commentary.aac"
}

# Navigation chapters
resource "rixl_video_chapters" "keynote_chapters" {
  project_id = rixl_project.vod_prod.id
  video_id   = rixl_video.keynote.id

  chapters = [
    {
      title          = "Opening Keynote"
      start_time_sec = 0.0
    },
    {
      title          = "Platform Announcements"
      start_time_sec = 240.0
    },
    {
      title          = "Architecture Deep Dive"
      start_time_sec = 780.5
    },
  ]
}
```

### 2. Custom Domain & Auto-Join Setup

Register a custom domain, associate it with a project, and define corporate email domain auto-join:

```terraform
# Register custom domain
resource "rixl_custom_domain" "company_domain" {
  org_id = var.org_id
  domain = "media.example.com"
}

# Map domain to project
resource "rixl_project_custom_domain" "domain_mapping" {
  org_id        = var.org_id
  project_id    = rixl_project.vod_prod.id
  custom_domain = rixl_custom_domain.company_domain.domain
}

# Auto-join rule for verified corporate users
resource "rixl_domain_auto_join" "corp_onboarding" {
  org_id  = var.org_id
  enabled = true
}
```

### 3. Content Feeds, Playlists & Media Posts

Organize media assets into channels and publish articles with subscriber notifications:

```terraform
variable "user_id" {
  type        = string
  description = "Rixl User ID"
}

# Curated video channel
resource "rixl_feed" "tech_talks" {
  project_id   = rixl_project.vod_prod.id
  name         = "tech-talks-series"
  description  = "Weekly technical talks and tutorials"
  allow_videos = true
  allow_images = false
}

# Publish a video post to the channel
resource "rixl_post" "episode_one" {
  project_id  = rixl_project.vod_prod.id
  feed_id     = rixl_feed.tech_talks.id
  video_id    = rixl_video.keynote.id
  description = "Full recording and slides for Episode 1."
}

# Subscriber notification preference
resource "rixl_blog_subscription" "user_sub" {
  user_id    = var.user_id
  subscribed = true
}
```

### 4. Least-Privilege IAM Access Policies & API Keys

Secure programmatic automation with scoped permissions and temporary machine credentials:

```terraform
# Policy restricting actions to video uploads
resource "rixl_access_policy" "video_uploader" {
  org_id      = var.org_id
  name        = "Video Uploader Role"
  description = "Grants write permissions strictly for video ingestion"

  permissions = [
    "videos:create",
    "videos:read",
    "videos:chapters:create",
    "subtitles:create",
  ]
}

# Bind policy to an identity
resource "rixl_policy_attachment" "uploader_binding" {
  org_id        = var.org_id
  policy_id     = rixl_access_policy.video_uploader.id
  identity_id   = var.user_id
  identity_type = "POLICY_IDENTITY_TYPE_USER"
}

# Machine-to-machine OAuth2 client credential
resource "rixl_client_credential" "ci_uploader" {
  name = "ci-cd-uploader-service"
}
```

### 5. Observability Dashboard & Viewer Telemetry

Create an executive monitoring dashboard with real-time viewership metrics:

```terraform
resource "rixl_dashboard" "platform_health" {
  name       = "Streaming Operations Cockpit"
  visibility = "org"
}

resource "rixl_dashboard_widget" "viewer_sessions" {
  dashboard_id = rixl_dashboard.platform_health.id
  title        = "Active Streaming Sessions"
  chart_type   = "timeseries"
  dataset      = "playback"
  metric       = "concurrent_sessions"
  pos_x        = 0
  pos_y        = 0
  width        = 6
  height       = 4
}
```

---

## Resource & Data Source Reference

| Category | Resource Types | Data Source Types |
| :--- | :--- | :--- |
| **Video & Media** | [`rixl_video`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/video), [`rixl_video_chapters`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/video_chapters), [`rixl_subtitle`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/subtitle), [`rixl_audio_track`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/audio_track), [`rixl_image`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/image) | `rixl_video`, `rixl_videos`, `rixl_video_stats`, `rixl_video_heatmap`, `rixl_video_hot_segments`, `rixl_top_videos`, `rixl_image`, `rixl_images`, `rixl_image_stats`, `rixl_top_images`, `rixl_chapters`, `rixl_subtitles`, `rixl_audio_tracks`, `rixl_languages` |
| **Projects & Content** | [`rixl_project`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/project), [`rixl_feed`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/feed), [`rixl_post`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/post), [`rixl_blog_subscription`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/blog_subscription) | `rixl_project`, `rixl_projects`, `rixl_feed`, `rixl_feeds`, `rixl_feed_stats`, `rixl_top_feeds`, `rixl_post`, `rixl_posts`, `rixl_post_stats`, `rixl_top_posts`, `rixl_blog_subscription` |
| **Custom Domains** | [`rixl_custom_domain`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/custom_domain), [`rixl_project_custom_domain`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/project_custom_domain), [`rixl_domain_auto_join`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/domain_auto_join) | `rixl_domain`, `rixl_domain_auto_join` |
| **IAM & Security** | [`rixl_access_policy`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/access_policy), [`rixl_policy_attachment`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/policy_attachment), [`rixl_api_key`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/api_key), [`rixl_client_credential`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/client_credential), [`rixl_auth_provider_connection`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/auth_provider_connection) | `rixl_policy`, `rixl_policies`, `rixl_policy_attachments`, `rixl_user_policies`, `rixl_api_keys`, `rixl_client_credentials`, `rixl_providers`, `rixl_passkeys`, `rixl_permission_registries` |
| **Organizations** | [`rixl_organization`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/organization), [`rixl_organization_invitation`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/organization_invitation), [`rixl_organization_member`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/organization_member) | `rixl_organization_members`, `rixl_memberships`, `rixl_membership_applications`, `rixl_check_membership`, `rixl_internal_membership_info`, `rixl_user`, `rixl_user_info` |
| **Dashboards & Telemetry** | [`rixl_dashboard`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/dashboard), [`rixl_dashboard_widget`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/dashboard_widget) | `rixl_dashboard`, `rixl_dashboards`, `rixl_dashboard_datasets`, `rixl_dashboard_stats`, `rixl_realtime_stats`, `rixl_storage_usage`, `rixl_storage_usage_history`, `rixl_bandwidth_usage`, `rixl_bandwidth_usage_history` |
| **Billing & Subscriptions** | [`rixl_subscription`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/subscription), [`rixl_payment_method`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/payment_method), [`rixl_billing_address`](https://registry.terraform.io/providers/rixlhq/rixl/latest/docs/resources/billing_address) | `rixl_subscription`, `rixl_subscription_history`, `rixl_plan`, `rixl_plans`, `rixl_invoices`, `rixl_payment_methods`, `rixl_payment_method_from_payment_intent`, `rixl_payment_method_from_setup_intent`, `rixl_billing_address` |

---

## Importing Existing Infrastructure

Existing Rixl resources can be imported into Terraform state seamlessly using `terraform import`:

```bash
# Import a project by its unique ID
terraform import rixl_project.vod_prod "proj_01hxxxxxxxxx"

# Import an uploaded video asset
terraform import rixl_video.keynote "vid_01hxxxxxxxxx"

# Import a project custom domain association (org_id/project_id)
terraform import rixl_project_custom_domain.domain_mapping "org_01hxxxxxxxxx/proj_01hxxxxxxxxx"

# Import an API key
terraform import rixl_api_key.deployer "key_01hxxxxxxxxx"

# Import an IAM policy attachment (org_id/policy_id/identity_id)
terraform import rixl_policy_attachment.uploader_binding "org_01hxxxxxxxxx/pol_01hxxxxxxxxx/usr_01hxxxxxxxxx"
```

---

## Frequently Asked Questions (FAQ)

### What video formats and codecs does Rixl support?
Rixl ingests MP4, MOV, MKV, AVI, and WebM container formats encoded with standard codecs (H.264 / AVC, H.265 / HEVC, VP9, AV1). The transcoding pipeline converts uploads into standard multi-bitrate HLS (m3u8 playlists and fMP4 / MPEG-TS segments) compatible with all modern browsers, mobile devices, and OTT players.

### Can I associate a custom domain with my project?
Yes. Using `rixl_custom_domain` and `rixl_project_custom_domain`, you can register and verify an organization domain and associate it with any project in your organization.

### How does the provider handle asynchronous video transcoding?
When creating a `rixl_video` resource, the provider securely uploads the media binary and tracks processing status. Rixl's asynchronous transcoding engine performs encoding in the cloud without blocking Terraform operations, returning valid playback IDs immediately.

### Are sensitive credentials masked in Terraform state and logs?
Yes. All credentials&mdash;including `api_key`, `bearer_token`, and generated client secrets&mdash;are marked with `Sensitive = true` in the Terraform Plugin Framework schema, preventing exposure in CLI outputs, CI/CD logs, and plan diffs.

---

## Development & Contributing

### Prerequisites

- [Go](https://go.dev/doc/install) >= 1.26
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5
- [mise](https://mise.jdx.dev/) tool manager

### Common Commands

```bash
# Format codebase
mise run fmt

# Run unit tests with code coverage
mise run test

# Run linter suite (golangci-lint)
mise run lint

# Generate and validate documentation
mise run docs
mise run docs-check
```

---

## License

This project is licensed under the [MIT License](LICENSE).
