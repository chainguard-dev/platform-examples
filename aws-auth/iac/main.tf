terraform {
  required_providers {
    aws        = { source = "hashicorp/aws" }
    chainguard = { source = "chainguard-dev/chainguard" }
    ko         = { source = "ko-build/ko" }
  }
}
