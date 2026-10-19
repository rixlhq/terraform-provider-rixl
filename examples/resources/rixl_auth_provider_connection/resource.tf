resource "rixl_auth_provider_connection" "google" {
  id    = "EXTERNAL_ACCOUNT_PROVIDER_GOOGLE"
  token = var.oauth_provider_token
}
