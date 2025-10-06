variable "org_name" {
  type        = string
  description = "The name of the Chainguard organization that we are copying images from. For instance: 'your.org.com'."
}

variable "cluster_name" {
  type        = string
  description = "The name of the AWS EKS cluster to run the job in."
}

variable "repo_name" {
  type        = string
  description = "The name of the destination repo where images should be copied to. For instance: 'my-repo'. This repository will be created by this terraform and images will be copied to 'my-repo/<image_name>'."
}

variable "namespace" {
  type        = string
  description = "The name of the namespace to deploy the job to."
  default     = "chainguard"
}

variable "updated_within" {
  type        = string
  description = "Copy tags images that were updated within this duration."
  default     = "72h"
}

variable "image_platform" {
  type        = string
  description = "The platform of the image to build."
  default     = "linux/amd64"
}

variable "base_image_name" {
  type        = string
  description = "The name of the Chainguard image to build the image-copy image from. Must have apk-tools and a shell."
  default     = "chainguard-base:latest"
}
