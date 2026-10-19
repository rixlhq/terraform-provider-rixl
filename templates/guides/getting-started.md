---
page_title: "Getting Started with the Rixl Terraform Provider - Terraform Provider for Rixl"
subcategory: "Guides"
description: |-
  Step-by-step onboarding guide to provisioning and managing your video streaming pipeline and media cloud infrastructure with the Rixl Terraform Provider.
---

# Getting Started with the Rixl Provider

The **Rixl Terraform Provider** enables platform engineering, SRE, and media infrastructure teams to declare, deploy, and manage video streaming pipelines, media assets, custom domains, IAM access policies, and real-time observability telemetry on [Rixl](https://rixl.com) as code.

---

## 1. Provider Requirements

Ensure your environment satisfies the following minimum requirements:

- [Terraform](https://www.terraform.io/downloads.html) **>= 1.5.0** or OpenTofu **>= 1.6.0**
- A valid [Rixl Account](https://rixl.com) with Organization Administrator or Project Contributor permissions
- A Rixl **API Key** or machine-to-machine **OAuth2 Bearer Token**

---

## 2. Authentication Configuration

The provider supports two authentication methods: static **API Keys** and dynamic **OAuth2 Bearer Tokens**.

### Option A: API Key (Recommended for CI/CD and Production)

Set your API key via the `RIXL_API_KEY` environment variable:

```bash
export RIXL_API_KEY="rixl_live_01hxxxxxxxxxxxxxxxxxxxxxxx"
```

Then configure the minimal provider block in your Terraform root module:

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
  # Automatically reads RIXL_API_KEY environment variable if omitted here
}
```

### Option B: OAuth2 Bearer Token (Dynamic Machine Credentials)

For temporary, automated, or identity-federated pipelines, authenticate using a bearer token minted from client credentials:

```bash
export RIXL_BEARER_TOKEN="eyJhbGciOi..."
```

```terraform
provider "rixl" {
  bearer_token = var.rixl_bearer_token # Or set RIXL_BEARER_TOKEN in your environment
}
```

---

## 3. Provisioning Your First Media Pipeline

Create a dedicated Terraform configuration (`main.tf`) to provision a project, register a custom domain, and upload a video:

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

# 2. Register an organization custom domain
resource "rixl_custom_domain" "company_domain" {
  org_id = var.org_id
  domain = "media.example.com"
}

# 3. Associate the custom domain with the project
resource "rixl_project_custom_domain" "domain_link" {
  org_id        = var.org_id
  project_id    = rixl_project.streaming_prod.id
  custom_domain = rixl_custom_domain.company_domain.domain
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
- [Custom Domains and Auto-Join Management](custom-domains) &mdash; Organization domain verification, project domain associations, and corporate email auto-join policies.
- [Content Feeds, Playlists, and Media Distribution](content-feeds-and-distribution) &mdash; Curated video channels, rich media posts, and subscriber notifications.
- [IAM, RBAC, and Security Automation](iam-and-security) &mdash; Least-privilege access policies, machine-to-machine OAuth2 credentials, and enterprise SSO.
- [Observability, Telemetry, and Analytics](monitoring-and-analytics) &mdash; Dashboards, playback sessions, bandwidth/storage metrics, and heatmaps.
