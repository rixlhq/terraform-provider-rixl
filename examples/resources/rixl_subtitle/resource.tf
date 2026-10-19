resource "rixl_subtitle" "english" {
  project_id    = var.rixl_project_id
  video_id      = rixl_video.keynote.id
  language_code = "en"
  label         = "English Subtitles"
  file_path     = "/path/to/subtitles_en.vtt"
}
