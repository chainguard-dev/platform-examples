output "secret-command" {
  description = "gcloud secrets command to upload your Jira API token."
  value       = format("echo -n YOUR JIRA TOKEN | gcloud --project %s secrets versions add %s --data-file=-", var.project_id, google_secret_manager_secret.jira-token.secret_id)
}

output "url" {
  value = google_cloud_run_service.jira-iss.status[0].url
}
