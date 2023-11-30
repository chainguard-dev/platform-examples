terraform {
  required_providers {
    aws        = { source = "hashicorp/aws" }
    chainguard = { source = "chainguard/chainguard" }
    ko         = { source = "ko-build/ko" }
  }
}

provider "aws" {}

provider "ko" {}

provider "chainguard" {}
