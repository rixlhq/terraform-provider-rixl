resource "rixl_api_key" "ci" {
  org_id     = var.rixl_org_id
  project_id = rixl_project.prod.id
  name       = "ci-uploader-key"
}
