resource "rixl_project_custom_domain" "cdn" {
  org_id        = var.rixl_org_id
  project_id    = rixl_project.prod.id
  custom_domain = "video.example.com"
}
