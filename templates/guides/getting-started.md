---
page_title: "Getting Started with Rixl - Terraform Provider for Rixl"
subcategory: "Guides"
description: |-
  Step-by-step guide for setting up the Rixl Terraform provider, configuring authentication with API keys or OAuth2 bearer tokens, and provisioning your first video streaming project.
---

# Getting Started with the Rixl Terraform Provider

The **Rixl Terraform Provider** enables infrastructure teams, streaming platforms, and content engineers to declare and manage video streaming infrastructure as code.

This guide walks through:
1. Installing and configuring the provider.
2. Selecting an authentication method (Static API Key vs OAuth2 Bearer Token).
3. Provisioning an organization project, custom domain, and media pipeline.

---

## 1. Provider Requirements & Configuration

Add the `rixlhq/rixl` provider to your Terraform configuration:

```terraform
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    rixl = {
      source  = "rixlhq/rixl"
      version = "~> 1.7"
    }
  }
}

provider "rixl" {
  # Authentication is typically provided via environment variables:
  # RIXL_API_KEY or RIXL_BEARER_TOKEN
}
```

---

## 2. Authentication Methods

The provider supports two production-ready authentication methods:

### Option A: Static API Key (Recommended for CI/CD runners)

Provide a personal or organization API key generated from the Rixl Console:

```bash
export RIXL_API_KEY="rixl_key_live_xxxxxxxxxxxxxxxxxxxxxxxx"
```

Or configure directly in HCL:

```terraform
provider "rixl" {
  api_key = var.rixl_api_key
}
```

### Option B: OAuth2 Bearer Token

For zero-trust environments, workload identity federation, or temporary session tokens:

```bash
export RIXL_BEARER_TOKEN="eyJhbGciOi..."
```

Or configure directly in HCL:

```terraform
provider "rixl" {
  bearer_token = var.rixl_bearer_token
}
```

---

## 3. Provisioning Your First Media Pipeline

Here is a complete setup provisioning a production media project, custom streaming domain, and an initial video asset with thumbnail and chapters:

```terraform
variable "org_id" {
  type        = string
  description = "Rixl Organization ID"
}

# 1. Create a dedicated project for video workloads
resource "rixl_project" "streaming_prod" {
  org_id = var.org_id
  name   = "production-streaming-v1"
}

# 2. Attach a verified custom domain for CDN distribution
resource "rixl_custom_domain" "cdn" {
  org_id = var.org_id
  domain = "stream.example.com"
}

# 3. Associate the custom domain with the project
resource "rixl_project_custom_domain" "cdn_link" {
  org_id        = var.org_id
  project_id    = rixl_project.streaming_prod.id
  custom_domain = rixl_custom_domain.cdn.domain
}

# 4. Upload and process a video asset
resource "rixl_video" "welcome" {
  project_id       = rixl_project.streaming_prod.id
  name             = "welcome-onboarding"
  file_path        = "${path.module}/media/welcome.mp4"
  poster_file_path = "${path.module}/media/poster.jpg"
}

# 5. Add interactive player chapters
resource "rixl_video_chapters" "welcome_chapters" {
  project_id = rixl_project.streaming_prod.id
  video_id   = rixl_video.welcome.id

  chapters = [
    {
      title          = "Introduction"
      start_time_sec = 0.0
    },
    {
      title          = "Platform Walkthrough"
      start_time_sec = 45.2
    },
    {
      title          = "Next Steps"
      start_time_sec = 120.0
    },
  ]
}
```

---

## 4. Next Steps & Architecture Guides

Explore in-depth architectural guides for production deployment:
- [Automated Video Streaming Pipeline](video-streaming-pipeline) &mdash; End-to-end multi-bitrate HLS encoding, multi-track audio, localized subtitles, and chapters.
- [Custom CDN Streaming Domains and SSL Automation](custom-domains-and-cdn) &mdash; Vanity streaming domains, DNS verification, and project routing.
- [Content Feeds, Playlists, and Media Distribution](content-feeds-and-distribution) &mdash; Curated video channels, rich media posts, and subscriber notifications.
- [IAM, RBAC, and Security Automation](iam-and-security) &mdash; Least-privilege access policies, machine-to-machine OAuth2 credentials, and enterprise SSO.
- [Observability, Telemetry, and Analytics](monitoring-and-analytics) &mdash; Dashboards, playback sessions, bandwidth/storage metrics, and heatmaps.
