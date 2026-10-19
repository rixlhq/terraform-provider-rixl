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
variable "org_id" {
  type        = string
  description = "Rixl Organization ID"
}

# Create a central streaming dashboard
resource "rixl_dashboard" "ops_cockpit" {
  org_id      = var.org_id
  name        = "Executive Streaming Telemetry"
  description = "Real-time viewer engagement, CDN bandwidth, and storage growth tracking"
}

# Add playback metrics widget
resource "rixl_dashboard_widget" "viewer_sessions" {
  dashboard_id = rixl_dashboard.ops_cockpit.id
  name         = "Concurrent Streaming Sessions"
  widget_type  = "timeseries"
  config       = jsonencode({
    metric      = "playback.sessions.concurrent"
    aggregation = "sum"
    interval    = "1m"
  })
}

# Add CDN bandwidth widget
resource "rixl_dashboard_widget" "cdn_egress" {
  dashboard_id = rixl_dashboard.ops_cockpit.id
  name         = "CDN Egress Bandwidth"
  widget_type  = "area_chart"
  config       = jsonencode({
    metric      = "cdn.egress.bytes"
    aggregation = "rate"
    interval    = "5m"
  })
}
```

---

## 2. Querying Real-Time Playback and Bandwidth

Use Terraform data sources to feed infrastructure metrics into downstream alerting configurations or automated scaling rules:

```terraform
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
