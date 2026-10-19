resource "rixl_audio_track" "spanish_dub" {
  project_id    = var.rixl_project_id
  video_id      = rixl_video.keynote.id
  language_code = "es"
  label         = "Spanish Audio Track"
  file_path     = "/path/to/audio_es.aac"
}
