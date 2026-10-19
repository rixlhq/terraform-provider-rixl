resource "rixl_client_credential" "backend_service" {
  org_id = var.rixl_org_id
  name   = "backend-transcoder-service"
}
