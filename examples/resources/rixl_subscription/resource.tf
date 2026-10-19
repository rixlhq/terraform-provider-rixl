resource "rixl_subscription" "enterprise" {
  org_id          = var.rixl_org_id
  stripe_price_id = "price_01h7abc123"
}
