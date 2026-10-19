resource "rixl_policy_attachment" "user_editor" {
  org_id    = var.rixl_org_id
  policy_id = rixl_access_policy.editor.id
  user_id   = "usr_01h7abc123"
}
