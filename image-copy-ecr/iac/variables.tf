variable "group_name" {
  type        = string
  description = "The name of the Chainguard group that we are subscribing to. For instance: 'your.org.com'."
}

variable "dst_repo" {
  type        = string
  description = "The name of the destination repo where images should be copied to. For instance: 'my-repo'. This repository will be created by this terraform and images will be copied to 'my-repo/<image_name>'."
}

variable "immutable_tags" {
  type        = bool
  description = "Whether to enable immutable tags."
  default     = false
}

variable "ignore_referrers" {
  type        = bool
  description = "Whether to ignore events for signatures and attestations."
  default     = false
}
