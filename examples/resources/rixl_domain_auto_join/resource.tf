resource "rixl_domain_auto_join" "corp" {
  org_id  = var.rixl_org_id
  enabled = true
}
