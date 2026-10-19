---
page_title: "Observability, Telemetry, and Analytics - Terraform Provider for Rixl"
subcategory: "Guides"
description: |-
  Guide for deploying observability dashboards, tracking video playback analytics, monitoring bandwidth and storage metrics, and querying engagement heatmaps with Rixl and Terraform.
---

# Observability, Telemetry, and Analytics

Rixl continuously captures delivery telemetry across all video streams, image assets, and custom CDN domains. Using the **Rixl Terraform Provider**, you can programmatically provision executive dashboards, monitor storage and bandwidth consumption, and integrate real-time stats into your observability stack.

---

## 1. Creating Custom Telemetry Dashboards

Provision centralized operational dashboards to visualize video platform health:

```terraform
# Create a central streaming dashboard
resource "rixl_dashboard" "ops_cockpit" {
  name       = "Executive Streaming Telemetry"
  visibility = "org"
}

# Add playback metrics widget
resource "rixl_dashboard_widget" "viewer_sessions" {
  dashboard_id = rixl_dashboard.ops_cockpit.id
  title        = "Concurrent Streaming Sessions"
  chart_type   = "timeseries"
  dataset      = "playback"
  metric       = "concurrent_sessions"
  pos_x        = 0
  pos_y        = 0
  width        = 6
  height       = 4
}

# Add CDN bandwidth widget
resource "rixl_dashboard_widget" "cdn_egress" {
  dashboard_id = rixl_dashboard.ops_cockpit.id
  title        = "CDN Egress Bandwidth"
  chart_type   = "timeseries"
  dataset      = "bandwidth"
  metric       = "egress_bytes"
  pos_x        = 6
  pos_y        = 0
  width        = 6
  height       = 4
}
```

---

## 2. Querying Real-Time Playback and Bandwidth

Use Terraform data sources to feed infrastructure metrics into downstream alerting configurations or automated scaling rules:

```terraform
variable "org_id" {
  type        = string
  description = "Rixl Organization ID"
}

variable "project_id" {
  type        = string
  description = "Rixl Project ID"
}

# Fetch real-time streaming metrics
data "rixl_realtime_stats" "current" {
  project_id = var.project_id
}

# Fetch historical bandwidth utilization for capacity forecasting
data "rixl_bandwidth_usage_history" "monthly" {
  org_id = var.org_id
}

# Fetch current storage quota consumption
data "rixl_storage_usage" "current" {
  org_id = var.org_id
}
```

---

## 3. Video Engagement Heatmaps and Hot Segments

Analyze which sections of your video library generate the highest viewer interest and replay frequency:

```terraform
variable "video_id" {
  type        = string
  description = "Rixl Video Asset ID"
}

# Retrieve audience retention across video duration
data "rixl_video_heatmap" "retention" {
  video_id = var.video_id
}

# Identify peak replayed timestamps
data "rixl_video_hot_segments" "highlights" {
  video_id = var.video_id
}

# Discover top-performing videos across the platform
data "rixl_top_videos" "trending" {
  project_id = var.project_id
}
```

---

## 4. Next Steps & Architecture Guides

- [Getting Started with Rixl](getting-started) &mdash; Initial provider configuration and project provisioning.
- [Automated Video Streaming Pipeline](video-streaming-pipeline) &mdash; Uploading videos, generating HLS streams, and adding subtitles.
- [Custom CDN Streaming Domains and SSL Automation](custom-domains-and-cdn) &mdash; Vanity streaming domains and DNS verification.
- [Content Feeds, Playlists, and Media Distribution](content-feeds-and-distribution) &mdash; Syndicating videos and posts across channels.
- [IAM, RBAC, and Security Automation](iam-and-security) &mdash; Restrict video upload permissions with least-privilege IAM policies.
