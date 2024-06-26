variable "project" {
  type        = string
  description = "The GCP project ID"
}

variable "repo" {
  type        = string
  description = "The GCP Artifact Registry repository name"
  default     = "remote"
}

variable "regions" {
  type        = list(string)
  description = "The GCP regions to deploy to"
  default     = ["us-east4"]
}

variable "username" {
  type        = string
  description = "Username for the cgr.dev pull token"
}

variable "password" {
  sensitive   = true
  type        = string
  description = "Password for the cgr.dev pull token"
}

provider "google" {
  project = var.project
}

resource "google_secret_manager_secret" "pull-token" {
  secret_id = "pull-token"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "pull-token" {
  secret      = google_secret_manager_secret.pull-token.name
  secret_data = var.password
}

resource "google_artifact_registry_repository" "repo" {
  for_each = toset(var.regions)

  location      = each.key
  repository_id = var.repo
  description   = "Remote Chainguard Images repository"
  format        = "DOCKER"
  mode          = "REMOTE_REPOSITORY"
  remote_repository_config {
    description = "chainguard"

    // The password won't be populated at first, so we need to disable
    // validating the token at repo-creation time.
    disable_upstream_validation = true

    docker_repository {
      custom_repository {
        uri = "https://cgr.dev"
      }
    }

    upstream_credentials {
      username_password_credentials {
        username                = var.username
        password_secret_version = google_secret_manager_secret_version.pull-token.name
      }
    }
  }
}

data "google_project" "project" {}

locals { project-number = data.google_project.project.number }

// Give the AR service account the ability to read the secret.
resource "google_secret_manager_secret_iam_member" "reader" {
  secret_id = google_secret_manager_secret.pull-token.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:service-${local.project-number}@gcp-sa-artifactregistry.iam.gserviceaccount.com"
}

output "repos" {
  value = [for r in var.regions : "${google_artifact_registry_repository.repo[r].location}-docker.pkg.dev/${var.project}/${google_artifact_registry_repository.repo[r].name}"]
}
