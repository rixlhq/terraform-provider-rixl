---
page_title: "Custom CDN Streaming Domains and SSL Automation - Terraform Provider for Rixl"
subcategory: "Guides"
description: |-
  Architecture guide for provisioning branded streaming domains, automated SSL/TLS certificates, project domain routing, and corporate email auto-join policies with Rixl and Terraform.
---

# Custom CDN Streaming Domains and SSL Automation

By default, Rixl delivers video streams, HLS manifests, and image assets via high-speed global edge points. For enterprise platforms, streaming under branded vanity URLs (such as `stream.example.com` or `video.acme.corp`) builds audience trust and maintains consistent domain authority for search engines and CDN caching.

Using the **Rixl Terraform Provider**, you can automate the provisioning of custom domains, bind them to specific project environments, verify DNS ownership, and configure corporate auto-join policies.

---

## Domain Routing Architecture

```mermaid
graph TD
    User([Viewer / Client Player]) -->|HTTPS Request| DNS[Custom CNAME: stream.example.com]
    DNS --> CDN[Rixl Edge CDN / Cloudflare Anycast]
    CDN --> ProjectRouter{Project Domain Mapping}
    ProjectRouter -->|stream.example.com| Prod[rixl_project: Production]
    ProjectRouter -->|staging-stream.example.com| Staging[rixl_project: Staging]
    Prod --> HLS[Multi-Bitrate HLS Manifests & Video Segments]
```

---

## 1. Registering Custom Domains

Declare a custom streaming domain for your Rixl organization:

```terraform
variable "org_id" {
  type        = string
  description = "Rixl Organization ID"
}

resource \"rixl_custom_domain\" \"streaming\" {
  org_id = var.org_id
  domain = \"stream.example.com\"
}

output \"domain_id\" {
  value       = rixl_custom_domain.streaming.id
  description = \"Identifier of the custom streaming domain\"
}
```

### DNS Verification & Verification Lookup

After provisioning the resource, retrieve DNS verification tokens using the `rixl_domain` data source to populate your DNS provider (Route 53, Cloudflare, Google Cloud DNS, Azure DNS, etc.):

```terraform
data \"rixl_domain\" \"streaming_status\" {
  domain = rixl_custom_domain.streaming.domain
}

output \"domain_verified\" {
  value       = data.rixl_domain.streaming_status.status
  description = \"Current verification status of the custom domain\"
}
```

---

## 2. Binding Custom Domains to Streaming Projects

To prevent cross-tenant traffic leakage and ensure strict environment isolation, bind custom domains to specific Rixl projects using `rixl_project_custom_domain`:

```terraform
# Create isolated project environment
resource \"rixl_project\" \"media_prod\" {
  org_id = var.org_id
  name   = \"production-vod\"
}

# Bind custom domain to the production project
resource \"rixl_project_custom_domain\" \"prod_mapping\" {
  org_id        = var.org_id
  project_id    = rixl_project.media_prod.id
  custom_domain = rixl_custom_domain.streaming.domain
}
```

Once mapped, all HLS manifests (`.m3u8`), video segments (`.ts` / `.m4s`), poster images, and subtitle tracks uploaded in `rixl_project.media_prod` will be delivered via `https://stream.example.com/...`.

---

## 3. Configuring Corporate Email Domain Auto-Join Policies

When onboarding team members or content creators, you can allow any user with a verified corporate email address to automatically join your Rixl organization:

```terraform
# Configure automated corporate onboarding
resource \"rixl_domain_auto_join\" \"corp_onboarding\" {
  org_id  = var.org_id
  enabled = true
}
```

Inspect the auto-join configuration with the `rixl_domain_auto_join` data source:

```terraform
data \"rixl_domain_auto_join\" \"corp_status\" {
  org_id = var.org_id
}
```

---

## Complete Multi-Environment Example

Below is a complete Terraform configuration provisioning custom domains across staging and production environments:

```terraform
terraform {
  required_version = \">= 1.5.0\"
  required_providers {
    rixl = {
      source  = \"rixlhq/rixl\"
      version = \"~> 1.7\"
    }
  }
}

provider \"rixl\" {
  api_key = var.rixl_api_key
}

variable \"rixl_api_key\" {
  type        = string
  sensitive   = true
  description = \"Rixl API Key\"
}

variable \"org_id\" {
  type        = string
  description = \"Rixl Organization ID\"
}

# Production Domain & Project
resource \"rixl_custom_domain\" \"prod_cdn\" {
  org_id = var.org_id
  domain = \"video.example.com\"
}

resource \"rixl_project\" \"prod_project\" {
  org_id = var.org_id
  name   = \"prod-media-delivery\"
}

resource \"rixl_project_custom_domain\" \"prod_binding\" {
  org_id        = var.org_id
  project_id    = rixl_project.prod_project.id
  custom_domain = rixl_custom_domain.prod_cdn.domain
}

# Staging Domain & Project
resource \"rixl_custom_domain\" \"staging_cdn\" {
  org_id = var.org_id
  domain = \"staging-video.example.com\"
}

resource \"rixl_project\" \"staging_project\" {
  org_id = var.org_id
  name   = \"staging-media-delivery\"
}

resource \"rixl_project_custom_domain\" \"staging_binding\" {
  org_id        = var.org_id
  project_id    = rixl_project.staging_project.id
  custom_domain = rixl_custom_domain.staging_cdn.domain
}
```

---

## Best Practices for Custom Video Domains

1. **CNAME Before Cutover**: Point your DNS CNAME record to Rixl's edge CDN endpoint prior to mapping domains to production feeds to ensure zero playback downtime.
2. **Automated TLS Renewals**: Rixl manages certificate renewals automatically once DNS ownership is verified.
3. **Environment Separation**: Maintain distinct vanity subdomains (e.g., `video.example.com` vs `staging-video.example.com`) mapped to separate Rixl projects.
4. **Combined with IAM**: Pair custom domains with `rixl_access_policy` to control which team members and service accounts have rights to alter domain routing.

---

## Next Steps & Related Guides

- [Getting Started with Rixl](getting-started) &mdash; Initial provider configuration and project provisioning.
- [Automated Video Streaming Pipeline](video-streaming-pipeline) &mdash; Uploading videos, generating HLS streams, and adding subtitles.
- [Content Feeds, Playlists, and Media Distribution](content-feeds-and-distribution) &mdash; Syndicating videos and posts across channels.
- [Observability, Telemetry, and Analytics](monitoring-and-analytics) &mdash; Tracking CDN bandwidth and viewer metrics.
