resource "rixl_payment_method" "card" {
  org_id            = var.rixl_org_id
  payment_method_id = "pm_01h7abc123"
}
