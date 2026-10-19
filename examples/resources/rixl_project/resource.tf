resource "rixl_project" "prod" {
  org_id        = var.rixl_org_id
  name          = "production-streaming"
  video_quality = "1080p"
}
