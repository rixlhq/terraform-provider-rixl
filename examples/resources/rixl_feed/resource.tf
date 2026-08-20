resource "rixl_feed" "main" {
  project_id  = var.rixl_project_id
  name        = "main"
  description = "Primary content feed"

  allow_images = true
  allow_videos = true
  has_likes    = true
  has_shares   = true
  has_comments = true
}
