resource "rixl_video" "keynote" {
  project_id       = var.rixl_project_id
  name             = "annual-keynote-2026"
  file_path        = "/path/to/keynote.mp4"
  poster_file_path = "/path/to/poster.jpg"
}
