resource "rixl_dashboard_widget" "bandwidth_chart" {
  dashboard_id = rixl_dashboard.media_metrics.id
  name         = "Egress Bandwidth"
  type         = "CHART"
  pos_x        = 0
  pos_y        = 0
  width        = 6
  height       = 4
}
