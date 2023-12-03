terraform {
  required_providers {
    chainguard = { source = "chainguard-dev/chainguard" }
    cosign     = { source = "chainguard-dev/cosign" }
    google     = { source = "hashicorp/google" }
    ko         = { source = "ko-build/ko" }
  }
}

locals {
  importpath = "github.com/chainguard-dev/enforce-events/github-issue-opener/cmd/app"
}

resource "google_service_account" "gh-iss-opener" {
  project    = var.project_id
  account_id = "${var.name}-issue-opener"
}

resource "google_secret_manager_secret" "gh-pat" {
  project   = var.project_id
  secret_id = "${var.name}-github-pat"
  replication {
    auto {}
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
  repo = "gcr.io/${var.project_id}/${local.importpath}"
}

resource "cosign_sign" "image" {
  image = ko_build.image.image_ref
}

resource "google_cloud_run_service" "gh-iss" {
  project  = var.project_id
  name     = "${var.name}-issue-opener"
  location = var.location

  template {
    spec {
      service_account_name = google_service_account.gh-iss-opener.email
      containers {
        image = cosign_sign.image.signed_ref
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
          name  = "LABELS"
          value = var.labels
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

// Subscribe to events under the root group.
resource "chainguard_subscription" "subscription" {
  parent_id = var.group
  sink      = google_cloud_run_service.gh-iss.status[0].url
}
