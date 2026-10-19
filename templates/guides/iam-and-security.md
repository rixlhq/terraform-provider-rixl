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

Define fine-grained IAM policies following the principle of least privilege. Policies support explicit `allow` or `deny` effects on specific actions and resources:

```terraform
# Read-only policy for media consumers
resource "rixl_access_policy" "read_only" {
  org_id      = var.org_id
  name        = "media-read-only"
  description = "Allows viewing and listing videos and content feeds without write access"
  statement   = jsonencode({
    Effect   = "Allow"
    Action   = [
      "videos:GetVideo",
      "videos:ListVideos",
      "feeds:GetFeed",
      "feeds:ListFeeds"
    ]
    Resource = "arn:rixl:project:*"
  })
}

# Full administrator policy for video ingestion pipelines
resource "rixl_access_policy" "media_publisher" {
  org_id      = var.org_id
  name        = "media-publisher"
  description = "Permits video upload, transcoding trigger, and subtitle management"
  statement   = jsonencode({
    Effect   = "Allow"
    Action   = [
      "videos:*",
      "subtitles:*",
      "audio_tracks:*"
    ]
    Resource = "arn:rixl:project:${rixl_project.vod.id}"
  })
}
```

---

## 2. Attaching Policies

Bind access policies to users, service principals, or roles:

```terraform
resource "rixl_policy_attachment" "publisher_binding" {
  policy_id = rixl_access_policy.media_publisher.id
  user_id   = var.service_user_id
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
  org_id = var.org_id
  name   = "webhook-dispatcher-key"
}

output "api_key" {
  value       = rixl_api_key.webhook_receiver.key
  sensitive   = true
  description = "The generated API Key secret"
}
```

---

## 5. Enterprise Identity Provider (IdP) SSO

Federate authentication with enterprise identity providers (Okta, Entra ID, Google Workspace, GitHub):

```terraform
resource "rixl_auth_provider_connection" "okta_sso" {
  org_id    = var.org_id
  provider_type  = "oidc"
  client_id      = var.okta_client_id
  client_secret  = var.okta_client_secret
}
```
