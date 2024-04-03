terraform {
  required_providers {
    google     = { source = "hashicorp/google" }
    chainguard = { source = "chainguard-dev/chainguard" }
    ko = {
      source  = "ko-build/ko"
      version = "0.0.11"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "ko" {
  repo = "gcr.io/${var.project_id}/pull-event-recorder"
}

variable "project_id" { type = string } // GCP project
variable "region" { type = string }     // GCP region
variable "group" { type = string }      // Chainguard group ID

// Create a network with several regional subnets
module "networking" {
  source = "chainguard-dev/common/infra//modules/networking"

  name       = "event-recorder"
  project_id = var.project_id
  regions    = [var.region]
}

// Create the Broker abstraction.
module "cloudevent-broker" {
  source = "chainguard-dev/common/infra//modules/cloudevent-broker"

  name       = "pull-event-broker"
  project_id = var.project_id
  regions    = module.networking.regional-networks

  notification_channels = []
}

// Authorize the "foo" service account to publish events.
module "trampoline-emits-events" {
  for_each = module.networking.regional-networks

  source = "chainguard-dev/common/infra//modules/authorize-private-service"

  project_id = var.project_id
  region     = each.key
  name       = module.cloudevent-broker.ingress.name

  service-account = google_service_account.trampoline.email
}

resource "google_service_account" "trampoline" {
  account_id   = "pull-event-trampoline"
  display_name = "Event Recorder Service Account"
}

// Run a cloud run service as the "foo" service account, and pass in the address
// of the regional ingress endpoint.
module "trampoline" {
  source = "chainguard-dev/common/infra//modules/regional-go-service"

  project_id = var.project_id
  name       = "pull-event-trampoline"
  regions    = module.networking.regional-networks

  notification_channels = []

  ingress = "INGRESS_TRAFFIC_ALL"
  egress  = "PRIVATE_RANGES_ONLY" // need to reach the issuer JWKS

  service_account = google_service_account.trampoline.email
  containers = {
    "trampoline" = {
      source = {
        working_dir = "${path.module}/.."
        importpath  = "./"
      }
      ports = [{ container_port = 8080 }]
      env = [{
        name  = "GROUP"
        value = var.group
        }, {
        name  = "ISSUER_URL"
        value = "https://issuer.enforce.dev"
      }]
      regional-env = [{
        name  = "EVENT_INGRESS_URI"
        value = { for k, v in module.trampoline-emits-events : k => v.uri }
      }]
    }
  }
}

// Who is deploying this?
data "google_client_openid_userinfo" "me" {}

// Record cloudevents of type com.example.foo and com.example.bar
module "recorder" {
  source = "chainguard-dev/common/infra//modules/cloudevent-recorder"

  deletion_protection = false

  name       = "pull-event-recorder"
  project_id = var.project_id
  regions    = module.networking.regional-networks
  broker     = module.cloudevent-broker.broker

  retention-period = 90 // days

  provisioner           = "user:${data.google_client_openid_userinfo.me.email}" // TODO: support SA deployer
  notification_channels = []

  types = {
    "dev.chainguard.registry.pull.v1" : {
      schema = file("${path.module}/pull.schema.json")
    }
  }
}

data "google_cloud_run_service" "trampoline" {
  depends_on = [module.trampoline]

  location = var.region
  name     = module.trampoline.names[var.region]
}

# Create a subscription to notify the Cloud Run service on changes under the root group.
resource "chainguard_subscription" "subscription" {
  parent_id = var.group
  sink      = data.google_cloud_run_service.trampoline.status[0].url
}
