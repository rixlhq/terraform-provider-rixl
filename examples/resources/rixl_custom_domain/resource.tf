resource "rixl_custom_domain" "stream" {
  org_id = var.rixl_org_id
  domain = "stream.example.com"
}
