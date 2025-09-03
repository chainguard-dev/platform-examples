# AWS region where resources will be deployed
variable "aws_region" {
  type        = string
  description = "AWS region to deploy into"
}

# AWS CLI profile (optional if using env vars or IAM role)
variable "aws_profile" {
  type        = string
  default     = null
  description = "AWS profile to use for credentials"
}

# Prefix for Lambda function and related resources
variable "name_prefix" {
  type        = string
  default     = "image-copy-all"
  description = "Base name prefix for the Lambda and ECR repo"
}

# Chainguard group/org (e.g. bannon.dev)
variable "group_name" {
  type        = string
  description = "Chainguard group/org to mirror images from"
}

# Source registry (defaults to Chainguard public)
variable "src_registry" {
  type        = string
  default     = "cgr.dev"
  description = "Source registry to mirror from"
}

# Optional prefix in ECR for repos
variable "dst_prefix" {
  type        = string
  default     = ""
  description = "Optional prefix to add before ECR repos"
}

# Chainguard pull token identity id (username)
variable "cgr_username" {
  type        = string
  description = "Chainguard pull-token identity id (username part)"
}

# Chainguard pull token JWT (sensitive, pass via CLI or env)
variable "cgr_password" {
  type        = string
  sensitive   = true
  description = "Chainguard pull-token JWT (password)"
}

# Go import path for the Lambda build
# Must match your go.mod "module" value + /cmd/fullsync
variable "go_importpath" {
  type        = string
  default     = "fullsync.local/cmd/fullsync"
  description = "Import path to the Lambda entrypoint"
}

variable "repo_list" {
  description = "List of source repos to mirror (cgr.dev/… paths)"
  type        = list(string)
  default     = []
}

variable "mirror_dry_run" {
  description = "If true, do not actually push; just log"
  type        = bool
  default     = false
}

variable "copy_all_tags" {
  description = "Mirror all tags for each repo (true) or just :latest (false)"
  type        = bool
  default     = false
}

variable "repo_tags" {
  description = "Map of repo → list of tags to mirror"
  type        = map(list(string))
  default     = {}
}