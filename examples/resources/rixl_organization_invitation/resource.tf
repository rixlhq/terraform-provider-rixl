resource "rixl_organization_invitation" "engineer" {
  org_id   = var.rixl_org_id
  username = "engineer"
  role     = "ORGANIZATION_ROLE_MEMBER"
}
