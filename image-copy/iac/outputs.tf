output "url" {
  value = google_cloud_run_service.image-copy.status[0].url
}
