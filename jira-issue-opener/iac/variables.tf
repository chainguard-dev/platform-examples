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

variable "jira_project" {
  type        = string
  description = "The jira project where issues will be filed."
}

variable "jira_url" {
  type        = string
  description = "The URL for jira."
}

variable "jira_user" {
  type        = string
  description = "The jira user to login as."
}

variable "issue_type" {
  type        = string
  description = "The type of issue to file (task, bug etc.)."
  default     = "Task"
}

