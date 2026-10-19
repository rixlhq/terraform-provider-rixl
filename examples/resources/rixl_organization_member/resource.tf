resource "rixl_organization_member" "admin" {
  org_id  = var.rixl_org_id
  user_id = "usr_01h7abc456"
  role    = "ORGANIZATION_ROLE_ADMIN"
}
