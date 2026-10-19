---
page_title: "IAM, RBAC, and Security Automation - Terraform Provider for Rixl"
subcategory: "Guides"
description: |-
  Security guide for implementing least-privilege access control, role-based access policies (RBAC), machine-to-machine OAuth2 credentials, and enterprise Single Sign-On (SSO) with the Rixl Terraform provider.
---

# Identity, Access Management (IAM), and Security Automation

The **Rixl Terraform Provider** provides native resources for managing enterprise security policies, identity federation, and automated service authentication.

---

## 1. Declarative Access Policies

Define fine-grained IAM policies following the principle of least privilege. Policies assign explicit permission strings to identities:

```terraform
variable "org_id" {
  type        = string
  description = "Rixl Organization ID"
}

# Read-only policy for media consumers
resource "rixl_access_policy" "read_only" {
  org_id      = var.org_id
  name        = "media-read-only"
  description = "Allows viewing and listing videos and content feeds without write access"
  permissions = [
    "video:read",
    "feed:read",
    "image:read",
  ]
}

# Full administrator policy for video ingestion pipelines
resource "rixl_access_policy" "media_publisher" {
  org_id      = var.org_id
  name        = "media-publisher"
  description = "Permits video upload, transcoding trigger, and subtitle management"
  permissions = [
    "video:read",
    "video:write",
    "subtitle:write",
    "audio_track:write",
    "feed:write",
  ]
}
```

---

## 2. Attaching Policies

Bind access policies to users, service principals, or roles using `rixl_policy_attachment`:

```terraform
resource "rixl_policy_attachment" "publisher_binding" {
  org_id        = var.org_id
  policy_id     = rixl_access_policy.media_publisher.id
  identity_id   = var.service_user_id
  identity_type = "POLICY_IDENTITY_TYPE_USER"
}
```

---

## 3. Machine-to-Machine OAuth2 Client Credentials

For automated backend microservices and CI/CD pipelines, generate scoped OAuth2 client credentials instead of user passwords:

```terraform
resource "rixl_client_credential" "ci_runner" {
  org_id = var.org_id
  name   = "github-actions-encoder-bot"
}

output "client_id" {
  value       = rixl_client_credential.ci_runner.client_id
  description = "The OAuth2 Client ID for automation"
}

output "client_secret" {
  value       = rixl_client_credential.ci_runner.client_secret
  sensitive   = true
  description = "The OAuth2 Client Secret (store securely in GitHub Actions Secrets or Vault)"
}
```

---

## 4. Long-Lived API Keys

For systems that do not support dynamic OAuth token exchanges, generate static API keys:

```terraform
resource "rixl_api_key" "webhook_receiver" {
  org_id     = var.org_id
  project_id = var.project_id
  name       = "webhook-dispatcher-key"
}

output "api_key" {
  value       = rixl_api_key.webhook_receiver.secret
  sensitive   = true
  description = "The generated API Key secret token"
}
```

---

## 5. Enterprise Identity Provider Connections

Connect enterprise identity providers (Google, Apple, Microsoft) for seamless authentication:

```terraform
resource "rixl_auth_provider_connection" "google_oauth" {
  id    = "EXTERNAL_ACCOUNT_PROVIDER_GOOGLE"
  token = var.google_oauth_provider_token
}
```

---

## 6. Next Steps & Architecture Guides

- [Getting Started with Rixl](getting-started) &mdash; Initial provider configuration and project provisioning.
- [Automated Video Streaming Pipeline](video-streaming-pipeline) &mdash; Uploading videos, generating HLS streams, and adding subtitles.
- [Custom CDN Streaming Domains and SSL Automation](custom-domains-and-cdn) &mdash; Vanity streaming domains and DNS verification.
- [Observability, Telemetry, and Analytics](monitoring-and-analytics) &mdash; Tracking CDN bandwidth and viewer metrics.
