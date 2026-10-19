resource "rixl_billing_address" "office" {
  org_id      = var.rixl_org_id
  name        = "Acme Media Corp"
  line1       = "123 Market Street"
  city        = "San Francisco"
  state       = "CA"
  postal_code = "94105"
  country     = "US"
}
