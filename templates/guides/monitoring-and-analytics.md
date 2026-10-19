---
page_title: "Observability, Telemetry, and Analytics - Terraform Provider for Rixl"
subcategory: "Guides"
description: |-
  Guide for deploying observability dashboards, tracking video playback analytics, monitoring bandwidth and storage metrics, and querying engagement heatmaps with Rixl and Terraform.
---

# Observability, Telemetry, and Analytics

Rixl continuously captures delivery telemetry across all video streams, image assets, and content feeds. Using the **Rixl Terraform Provider**, you can programmatically provision executive dashboards, monitor storage and bandwidth consumption, and integrate real-time stats into your observability stack.

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

# Add bandwidth consumption widget
resource "rixl_dashboard_widget" "bandwidth_usage" {
  dashboard_id = rixl_dashboard.ops_cockpit.id
  title        = "Total Bandwidth Consumption"
  chart_type   = "timeseries"
  dataset      = "bandwidth"
  metric       = "total_bandwidth_bytes"
  pos_x        = 6
  pos_y        = 0
  width        = 6
  height       = 4
}
```

---

## 2. Tracking Storage & Bandwidth Consumption

Read metered infrastructure metrics for billing alerts and capacity forecasting:

```terraform
variable "org_id" {
  type        = string
  description = "Rixl Organization ID"
}

# Query current storage utilization
data "rixl_storage_usage" "current" {
  org_id = var.org_id
}

# Query historical storage timeseries
data "rixl_storage_usage_history" "history" {
  org_id = var.org_id
}

# Query current bandwidth consumption
data "rixl_bandwidth_usage" "bandwidth" {
  org_id = var.org_id
}

output "total_bandwidth_consumed" {
  value       = data.rixl_bandwidth_usage.bandwidth.total_bandwidth_bytes
  description = "Total bytes delivered"
}
```

---

## 3. Video Engagement Analytics & Retention Heatmaps

Inspect audience drop-off and replayed intervals for individual video assets:

```terraform
variable "video_id" {
  type        = string
  description = "Target Video ID"
}

# Query video viewership statistics
data "rixl_video_stats" "stats" {
  video_id = var.video_id
}

# Query audience retention heatmap
data "rixl_video_heatmap" "retention" {
  video_id = var.video_id
}

# Query peak replayed video segments
data "rixl_video_hot_segments" "hot_segments" {
  video_id = var.video_id
}
```

---

## 4. Real-Time Telemetry

Query live streaming indicators across currently active viewer sessions:

```terraform
data "rixl_realtime_stats" "live_metrics" {
  org_id = var.org_id
}

output "active_viewers" {
  value       = data.rixl_realtime_stats.live_metrics
  description = "Current real-time viewer count"
}
```

---

## Next Steps & Related Guides

- [Getting Started with Rixl](getting-started) &mdash; Initial provider configuration and project provisioning.
- [Automated Video Streaming Pipeline](video-streaming-pipeline) &mdash; Uploading videos, generating HLS streams, and adding subtitles.
- [Custom Domains and Auto-Join Management](custom-domains) &mdash; Organization domain verification, project domain associations, and corporate email auto-join policies.
- [IAM, RBAC, and Security Automation](iam-and-security) &mdash; Managing API keys and access policies.
