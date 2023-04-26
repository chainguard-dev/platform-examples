terraform {
  required_providers {
    cosign = {
      source = "chainguard-dev/cosign"
    }
    ko = {
      source = "ko-build/ko"
    }
    google = {
      source = "hashicorp/google"
    }
  }
}

locals {
  importpath = "github.com/chainguard-dev/enforce-events/slack-webhook/cmd/app"
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

data "cosign_verify" "base-image" {
  image = "cgr.dev/chainguard/static:latest-glibc"

  policy = jsonencode({
    apiVersion = "policy.sigstore.dev/v1beta1"
    kind       = "ClusterImagePolicy"
    metadata = {
      name = "chainguard-images-are-signed"
    }
    spec = {
      images = [{
        glob = "cgr.dev/**"
      }]
      authorities = [{
        keyless = {
          url = "https://fulcio.sigstore.dev"
          identities = [{
            issuer  = "https://token.actions.githubusercontent.com"
            subject = "https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main"
          }]
        }
        ctlog = {
          url = "https://rekor.sigstore.dev"
        }
      }]
    }
  })
}

resource "ko_build" "image" {
  base_image  = data.cosign_verify.base-image.verified_ref
  importpath  = local.importpath
  working_dir = path.module
  # repo overrides KO_DOCKER_REPO environment variable
  repo        = "gcr.io/${var.project_id}/${local.importpath}"
}

resource "cosign_sign" "image" {
  image = ko_build.image.image_ref
}

resource "google_cloud_run_service" "slack-notifier" {
  project  = var.project_id
  name     = "${var.name}-slack-notifier"
  location = var.location

  template {
    spec {
      service_account_name = google_service_account.slack-notifier.email
      containers {
        image = cosign_sign.image.signed_ref
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
        env {
          name  = "NOTIFY_LEVEL"
          value = var.notify_level
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
