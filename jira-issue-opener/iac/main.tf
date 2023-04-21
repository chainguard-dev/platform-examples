terraform {
  required_providers {
    ko = {
      source = "ko-build/ko"
    }
    google = {
      source = "hashicorp/google"
    }
  }
}

locals {
  importpath = "github.com/chainguard-dev/enforce-events/jira-issue-opener/cmd/app"
}

resource "google_service_account" "jira-iss-opener" {
  project    = var.project_id
  account_id = "${var.name}-jira-iss-opener"
}

resource "google_secret_manager_secret" "jira-token" {
  project   = var.project_id
  secret_id = "${var.name}-jira-token"
  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "initial-secret-version" {
  secret = google_secret_manager_secret.jira-token.id

  secret_data = "you need to populate the secret."

  lifecycle {
    ignore_changes = [
      # This is populated after everything is up.
      secret_data
    ]
  }
}

resource "google_secret_manager_secret_iam_member" "grant-secret-access" {
  secret_id = google_secret_manager_secret.jira-token.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.jira-iss-opener.email}"
}

resource "ko_build" "image" {
  base_image  = "cgr.dev/chainguard/static"
  importpath  = local.importpath
  working_dir = path.module
  repo        = "gcr.io/${var.project_id}/${local.importpath}"
}

resource "google_cloud_run_service" "jira-iss" {
  project  = var.project_id
  name     = "${var.name}-jira-iss-opener"
  location = var.location

  template {
    spec {
      service_account_name = google_service_account.jira-iss-opener.email
      containers {
        image = ko_build.image.image_ref
        env {
          name  = "ISSUER_URL"
          value = "https://issuer.${var.env}"
        }
        env {
          name  = "GROUP"
          value = var.group
        }
        env {
          name  = "JIRA_USER"
          value = var.jira_user
        }
        env {
          name  = "JIRA_PROJECT"
          value = var.jira_project
        }
        env {
          name  = "JIRA_URL"
          value = var.jira_url
        }
        env {
          name  = "ISSUE_TYPE"
          value = var.issue_type
        }
        env {
          name = "JIRA_TOKEN"
          value_from {
            secret_key_ref {
              name = google_secret_manager_secret.jira-token.secret_id
              key  = "latest"
            }
          }
        }
      }
    }
  }

  # Wait for the initial secret version to be created.
  depends_on = [
    google_secret_manager_secret_version.initial-secret-version
  ]
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  project  = var.project_id
  location = google_cloud_run_service.jira-iss.location
  service  = google_cloud_run_service.jira-iss.name

  policy_data = data.google_iam_policy.noauth.policy_data
}
