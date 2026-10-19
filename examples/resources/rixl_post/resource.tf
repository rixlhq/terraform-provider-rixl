resource "rixl_post" "announcement" {
  project_id  = var.rixl_project_id
  feed_id     = rixl_feed.main.id
  video_id    = rixl_video.keynote.id
  description = "Watch the replay of our platform announcement."
}
