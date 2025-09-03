terraform {
  required_version = ">= 1.5.0"  # compatible with your 1.5.7

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
    ko = {
      source  = "ko-build/ko"
      version = ">= 0.0.16"
    }
  }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile
}

# Single source of truth for these datasources
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}