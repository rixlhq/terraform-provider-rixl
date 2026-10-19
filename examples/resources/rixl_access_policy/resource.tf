resource "rixl_access_policy" "editor" {
  org_id      = var.rixl_org_id
  name        = "media-editor"
  description = "Grants write permissions to videos and feeds"

  permissions = [
    "video:read",
    "video:write",
    "feed:read",
    "feed:write",
  ]
}
