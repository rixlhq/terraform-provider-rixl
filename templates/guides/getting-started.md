---
page_title: "Getting Started with Rixl - Terraform Provider for Rixl"
subcategory: "Guides"
description: |-
  Step-by-step guide for setting up the Rixl Terraform provider, configuring authentication with API tokens or OAuth2 client credentials, and provisioning your first video streaming project.
---

# Getting Started with the Rixl Terraform Provider

The **Rixl Terraform Provider** enables infrastructure teams, streaming platforms, and content engineers to declare and manage video streaming infrastructure as code.

This guide walks through:
1. Installing and configuring the provider.
2. Selecting an authentication method (API Token vs OAuth2 Client Credentials).
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
      version = "~> 1.7.0"
    }
  }
}

provider "rixl" {
  # Authentication is typically provided via environment variables:
  # RIXL_API_KEY or RIXL_CLIENT_ID + RIXL_CLIENT_SECRET
}
```

---

## 2. Authentication Methods

The provider supports two production-ready authentication methods:

### Option A: Static API Token (Recommended for CI/CD runners)

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

### Option B: Machine-to-Machine OAuth2 Client Credentials

For zero-trust environments and enterprise automation, generate machine credentials:

```bash
export RIXL_CLIENT_ID="cc_01hxxxxxxxxxxxxxxxxxxxx"
export RIXL_CLIENT_SECRET="secret_xxxxxxxxxxxxxxxxxxxxxxxx"
```

Or configure directly in HCL:

```terraform
provider "rixl" {
  client_id     = var.rixl_client_id
  client_secret = var.rixl_client_secret
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
  project_id = rixl_project.streaming_prod.id
  domain_id  = rixl_custom_domain.cdn.id
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
      title      = "Introduction"
      start_time = 0.0
    },
    {
      title      = "Platform Walkthrough"
      start_time = 45.2
    },
    {
      title      = "Next Steps"
      start_time = 120.0
    }
  ]
}
```

---

## 4. Next Steps

- Explore the [Automated Video Streaming Pipeline Guide](video-streaming-pipeline) to learn about multi-language subtitles and audio tracks.
- Learn how to implement least-privilege RBAC using the [IAM and Security Guide](iam-and-security).
