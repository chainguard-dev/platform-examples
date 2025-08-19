terraform {
  required_providers {
    chainguard = { source = "chainguard-dev/chainguard" }
    cosign     = { source = "chainguard-dev/cosign" }
    google     = { source = "hashicorp/google" }
    ko         = { source = "ko-build/ko" }
  }
}

locals {
  importpath = "github.com/chainguard-dev/platform-examples/image-copy-gcp"
}

provider "google" {
  project = var.project_id
}

provider "cosign" {}
provider "ko" {}

resource "google_service_account" "image-copy" {
  account_id = "${var.name}-image-copy"
}

resource "ko_build" "image" {
  base_image  = data.cosign_verify.base-image.verified_ref
  importpath  = local.importpath
  working_dir = path.module
  repo        = "${google_artifact_registry_repository.dst-repo.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.dst-repo.repository_id}/image-copy"
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

resource "google_cloud_run_service" "image-copy" {
  name     = "${var.name}-image-copy"
  location = var.location

  template {
    spec {
      service_account_name = google_service_account.image-copy.email
      containers {
        image = ko_build.image.image_ref
        env {
          name  = "ISSUER_URL"
          value = "https://issuer.${var.env}"
        }
        env {
          name  = "API_ENDPOINT"
          value = "https://console-api.${var.env}"
        }
        env {
          name  = "GROUP_NAME"
          value = var.group_name
        }
        env {
          name  = "GROUP"
          value = data.chainguard_group.group.id
        }
        env {
          name  = "IDENTITY"
          value = chainguard_identity.puller-identity.id
        }
        env {
          name  = "DST_REPO"
          value = "${var.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.dst-repo.name}"
        }
        env {
          name  = "IGNORE_REFERRERS"
          value = var.ignore_referrers
        }
        env {
          name  = "VERIFY_SIGNATURES"
          value = var.verify_signatures
        }
      }
    }
  }
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
  location = google_cloud_run_service.image-copy.location
  service  = google_cloud_run_service.image-copy.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_artifact_registry_repository" "dst-repo" {
  location      = var.location
  repository_id = "${var.name}-${var.dst_repo}"
  description   = "image-copy repository"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository_iam_member" "pusher" {
  location   = google_artifact_registry_repository.dst-repo.location
  repository = google_artifact_registry_repository.dst-repo.name
  role       = "roles/artifactregistry.createOnPushWriter"
  member     = "serviceAccount:${google_service_account.image-copy.email}"
}

data "chainguard_group" "group" {
  name = var.group_name
}

# Create the identity for our Cloud Run service to assume.
resource "chainguard_identity" "puller-identity" {
  parent_id = data.chainguard_group.group.id
  name      = "image-copy cgr puller"

  claim_match {
    issuer  = "https://accounts.google.com"
    subject = google_service_account.image-copy.unique_id
  }
}

# Look up the registry.pull role to grant the identity.
data "chainguard_role" "puller" {
  name = "registry.pull"
}

# Grant the identity the "registry.pull" role on the root group.
resource "chainguard_rolebinding" "puller" {
  identity = chainguard_identity.puller-identity.id
  group    = data.chainguard_group.group.id
  role     = data.chainguard_role.puller.items[0].id
}

# Create a role that can find the catalog_syncer and apko_builder identities
# for signature verification
resource "chainguard_role" "verifier" {
  count = var.verify_signatures ? 1 : 0

  parent_id   = data.chainguard_group.group.id
  name        = "${var.name}-image-copy-verifier"
  description = "Custom role for ${var.name}-image-copy image verification"

  capabilities = [
    "identity.list",
  ]
}

# Grant the identity the custom role on the root group.
resource "chainguard_rolebinding" "verifier" {
  count = var.verify_signatures ? 1 : 0

  identity = chainguard_identity.puller-identity.id
  group    = data.chainguard_group.group.id
  role     = chainguard_role.verifier[0].id
}

# Create a subscription to notify the Cloud Run service on changes under the root group.
resource "chainguard_subscription" "subscription" {
  parent_id = data.chainguard_group.group.id
  sink      = google_cloud_run_service.image-copy.status[0].url
}
