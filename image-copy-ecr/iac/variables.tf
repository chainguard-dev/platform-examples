variable "group" {
  type        = string
  description = "The Chainguard group that we are subscribing to."
}

variable "dst_repo" {
  type        = string
  description = "The destination repo where images should be copied to."
}
