variable "name" {
  description = "Name to prefix to created resources."
}

variable "project_id" {
  type        = string
  description = "The project that will host this solution."
}

variable "location" {
  type        = string
  default     = "us-central1"
  description = "Where to run the Cloud Run service."
}

variable "env" {
  type        = string
  default     = "enforce.dev"
  description = "The Chainguard environment against which this solution is configured."
}

variable "group" {
  type        = string
  description = "The Chainguard group that we are subscribing to policy violations under."
}

variable "github_org" {
  type        = string
  description = "The github organization (or user) under which the repository lives."
}

variable "github_repo" {
  type        = string
  description = "The github repository in which to run the workflow."
}

variable "github_workflow_id" {
  type        = string
  description = "The github workflow ID to invoke."
}
