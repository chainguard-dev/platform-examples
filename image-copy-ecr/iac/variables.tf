variable "group" {
  type        = string
  description = "The Chainguard group that we are subscribing to."
}

variable "dst_repo" {
  type        = string
  description = "The destination repo where images should be copied to."
}

variable "immutable_tags" {
  type        = bool
  description = "Whether to enable immutable tags."
  default     = false
}
