terraform {
  required_providers {
    ko = {
      source  = "chainguard-dev/ko"
    }
    google = {
      source  = "hashicorp/google"
    }
  }
}

resource "google_service_account" "gh-iss-opener" {
  project    = var.project_id
  account_id = "${var.name}-issue-opener"
}

resource "google_secret_manager_secret" "gh-pat" {
  project   = var.project_id
  secret_id = "${var.name}-github-pat"
  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "initial-secret-version" {
  secret = google_secret_manager_secret.gh-pat.id

  secret_data = "you need to populate the secret."

  lifecycle {
    ignore_changes = [
      # This is populated after everything is up.
      secret_data
    ]
  }
}

resource "google_secret_manager_secret_iam_member" "grant-secret-access" {
  secret_id = google_secret_manager_secret.gh-pat.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.gh-iss-opener.email}"
}

resource "ko_image" "image" {
  base_image  = "ghcr.io/distroless/static"
  importpath  = "github.com/chainguard-dev/enforce-events/github-issue-opener/cmd/app"
  working_dir = path.module
}

resource "google_cloud_run_service" "gh-iss" {
  project  = var.project_id
  name     = "${var.name}-issue-opener"
  location = var.location

  template {
    spec {
      service_account_name = google_service_account.gh-iss-opener.email
      containers {
        image = ko_image.image.image_ref
        env {
          name  = "ISSUER_URL"
          value = "https://issuer.${var.env}"
        }
        env {
          name  = "GROUP"
          value = var.group
        }
        env {
          name  = "GITHUB_ORG"
          value = var.github_org
        }
        env {
          name  = "GITHUB_REPO"
          value = var.github_repo
        }
        env {
          name = "GITHUB_TOKEN"
          value_from {
            secret_key_ref {
              name = google_secret_manager_secret.gh-pat.secret_id
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
  location = google_cloud_run_service.gh-iss.location
  service  = google_cloud_run_service.gh-iss.name

  policy_data = data.google_iam_policy.noauth.policy_data
}
