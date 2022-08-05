output "secret-command" {
  description = "gcloud secrets command to upload the Slack webhook URL."
  value       = format("echo -n YOUR SLACK WEBHOOK | gcloud --project %s secrets versions add %s --data-file=-", var.project_id, google_secret_manager_secret.slack-url.secret_id)
}

output "url" {
  value = google_cloud_run_service.slack-notifier.status[0].url
}

