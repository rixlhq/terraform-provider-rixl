data "rixl_policy_attachments" "role_bindings" {
  org_id    = var.rixl_org_id
  policy_id = var.rixl_policy_id
}
