terraform {
  required_providers {
    chainguard = {
      source = "chainguard/chainguard"
    }
    ko = {
      source  = "chainguard-dev/ko"
      version = "0.0.2"
    }
    google = {
      source  = "hashicorp/google"
      version = "4.26.0" // Or whatever release
    }
  }
}

resource "google_service_account" "slack-notifier" {
  project    = var.project_id
  account_id = "${var.name}-slack-notifier"
}

resource "google_secret_manager_secret" "slack-url" {
  project   = var.project_id
  secret_id = "${var.name}-slack-url"
  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "initial-secret-version" {
  secret = google_secret_manager_secret.slack-url.id

  secret_data = "you need to populate the secret."

  lifecycle {
    ignore_changes = [
      # This is populated after everything is up.
      secret_data
    ]
  }
}

resource "google_secret_manager_secret_iam_member" "grant-secret-access" {
  secret_id = google_secret_manager_secret.slack-url.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.slack-notifier.email}"
}

resource "ko_image" "app" {
  // TODO: switch to ghcr.io/distroless/static when Cloud Run supports OCI.
  base_image  = "gcr.io/distroless/static:nonroot"
  importpath  = "chainguard.dev/demos/slack-webhook/cmd/app"
  working_dir = path.module
}

resource "google_cloud_run_service" "slack-notifier" {
  project  = var.project_id
  name     = "${var.name}-slack-notifier"
  location = var.location

  template {
    spec {
      service_account_name = google_service_account.slack-notifier.email
      containers {
        image = ko_image.app.image_ref
        env {
          name  = "CONSOLE_URL"
          value = "https://console.${var.env}"
        }
        env {
          name  = "ISSUER_URL"
          value = "https://issuer.${var.env}"
        }
        env {
          name  = "GROUP"
          value = var.group
        }
        env {
          name = "SLACK_WEBHOOK"
          value_from {
            secret_key_ref {
              name = google_secret_manager_secret.slack-url.secret_id
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
  location = google_cloud_run_service.slack-notifier.location
  service  = google_cloud_run_service.slack-notifier.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "chainguard_subscription" "slack-notifier-subscription" {
  parent_id = var.group

  sink = google_cloud_run_service.slack-notifier.status[0].url
}
