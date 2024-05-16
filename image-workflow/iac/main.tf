terraform {
  required_providers {
    cosign     = { source = "chainguard-dev/cosign" }
    ko         = { source = "ko-build/ko" }
    google     = { source = "hashicorp/google" }
    chainguard = { source = "chainguard/chainguard" }
  }
}

provider "ko" {
  repo = "gcr.io/${var.project_id}/${var.name}-image-workflow"
}

provider "google" {
  project = var.project_id
}

# Create a service account to run the service, with access to the GitHub PAT secret.
resource "google_service_account" "image-workflow" {
  account_id = "${var.name}-image-workflow"
}

# Create a secret to hold the GitHub personal access token.
resource "google_secret_manager_secret" "gh-pat" {
  project   = var.project_id
  secret_id = "${var.name}-github-pat"
  replication {
    auto {}
  }
}

# Create the initial secret version, which will let the service start up.
# After the infrastructure is provisioned, the secret-command output will
# show you how to populate the secret.
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

# Grant the service account access to read the secret.
resource "google_secret_manager_secret_iam_member" "grant-secret-access" {
  secret_id = google_secret_manager_secret.gh-pat.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.image-workflow.email}"
}

# Verify the base image is signed by the expected identity.
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

# Build the app into a container image.
resource "ko_build" "image" {
  importpath  = "github.com/imjasonh/terraform-playground/image-workflow/cmd/app"
  working_dir = path.module
}

# Deploy the service, running as the Service Account, with the GitHub PAT as a secret env var.
resource "google_cloud_run_service" "image-workflow" {
  name     = "${var.name}-image-workflow"
  location = var.location

  template {
    spec {
      service_account_name = google_service_account.image-workflow.email
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
          name  = "GITHUB_ORG"
          value = var.github_org
        }
        env {
          name  = "GITHUB_REPO"
          value = var.github_repo
        }
        env {
          name  = "GITHUB_WORKFLOW_ID"
          value = var.github_workflow_id
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
}

# Look up the IAM policy for "allUsers" to allow anyone to invoke the service.
data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

# Allow anyone to invoke the service.
resource "google_cloud_run_service_iam_policy" "noauth" {
  location = google_cloud_run_service.image-workflow.location
  service  = google_cloud_run_service.image-workflow.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

# Create a subscription to notify the Cloud Run service on changes under the root group.
resource "chainguard_subscription" "subscription" {
  parent_id = var.group
  sink      = google_cloud_run_service.image-workflow.status[0].url
}

# Create an identity that can be assumed by the GitHub workflow.
resource "chainguard_identity" "puller" {
  parent_id   = var.group
  name        = "${var.name}-image-workflow image puller"
  description = <<EOF
    This is an identity that authorizes the workflow in the
    GitHub repo to pull from the Chainguard Registry, to test it.
  EOF

  claim_match {
    issuer  = "https://token.actions.githubusercontent.com"
    subject = "repo:${var.github_org}/${var.github_repo}:ref:refs/heads/main"
  }
}

# Look up the registry.push role to grant the actions identity below.
data "chainguard_roles" "registry-pull" {
  name = "registry.pull"
}

# Grant the actions identity the "registry.pull" role on the group.
resource "chainguard_rolebinding" "private-pusher-is-public-puller" {
  identity = chainguard_identity.puller.id
  group    = var.group
  role     = data.chainguard_roles.registry-pull.items[0].id
}
