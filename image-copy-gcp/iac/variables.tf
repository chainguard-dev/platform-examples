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

variable "group_name" {
  type        = string
  description = "The name of the Chainguard group that we are subscribing to events under."
}

variable "dst_repo" {
  type        = string
  description = "The destination repo where images should be copied to."
}

variable "ignore_referrers" {
  type        = bool
  description = "Whether to ignore events for signatures and attestations."
  default     = false
}

variable "verify_signatures" {
  type        = bool
  description = "Whether to verify signatures before copying images."
  default     = false
}
