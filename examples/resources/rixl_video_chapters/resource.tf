resource "rixl_video_chapters" "chapters" {
  project_id = var.rixl_project_id
  video_id   = rixl_video.keynote.id

  chapters = [
    {
      title          = "Opening Remarks"
      start_time_sec = 0.0
    },
    {
      title          = "Architecture Overview"
      start_time_sec = 180.5
    },
    {
      title          = "Demo & Closing"
      start_time_sec = 620.0
    },
  ]
}
