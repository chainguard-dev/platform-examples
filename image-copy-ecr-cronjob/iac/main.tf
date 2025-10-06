data "aws_eks_cluster" "cluster" {
  name = var.cluster_name
}

data "aws_eks_cluster_auth" "cluster" {
  name = var.cluster_name
}

data "aws_iam_openid_connect_provider" "cluster" {
  url = data.aws_eks_cluster.cluster.identity[0].oidc[0].issuer
}

locals {
  oidc_provider = replace(data.aws_eks_cluster.cluster.identity[0].oidc[0].issuer, "https://", "")
}

resource "aws_ecr_repository" "chainguard" {
  name         = var.repo_name
  force_delete = true
}

resource "aws_ecr_repository" "image_copy" {
  name         = "${var.repo_name}-image-copy"
  force_delete = true
}

locals {
  source_hash = sha1(
    join(
      "",
      [
        filesha1("${path.module}/../Dockerfile"),
        filesha1("${path.module}/../image-copy.sh"),
      ]
    )
  )
}

provider "docker" {
  registry_auth {
    address     = split("/", aws_ecr_repository.image_copy.repository_url)[0]
    config_file = pathexpand("~/.docker/config.json")
  }
}

resource "docker_registry_image" "image" {
  name          = docker_image.image.name
  keep_remotely = true
  triggers = {
    image_id = docker_image.image.image_id
  }
}

resource "docker_image" "image" {
  name = aws_ecr_repository.image_copy.repository_url
  build {
    context    = abspath("${path.module}/..")
    dockerfile = abspath("${path.module}/../Dockerfile")
    tag        = ["latest"]
    platform   = "linux/amd64"
    builder    = "default"
    build_args = {
      ORG_NAME        = var.org_name
      BASE_IMAGE_NAME = var.base_image_name
    }
  }
  triggers = {
    source_hash = local.source_hash
  }
}

resource "aws_iam_role" "image_copy" {
  name               = "${var.cluster_name}-image-copy"
  assume_role_policy = data.aws_iam_policy_document.image_copy_assume_role.json

  inline_policy {
    name   = "image-copy"
    policy = data.aws_iam_policy_document.image_copy.json
  }
}

data "aws_iam_policy_document" "image_copy_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"
    principals {
      type        = "Federated"
      identifiers = [data.aws_iam_openid_connect_provider.cluster.arn]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_provider}:sub"
      values   = ["system:serviceaccount:${var.namespace}:image-copy"]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_provider}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "image_copy" {
  statement {
    effect = "Allow"
    actions = [
      "ecr:CreateRepository",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:GetRepositoryPolicy",
      "ecr:DescribeRepositories",
      "ecr:ListImages",
      "ecr:DescribeImages",
      "ecr:BatchGetImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:PutImage"
    ]
    resources = [
      aws_ecr_repository.chainguard.arn,
      "${aws_ecr_repository.chainguard.arn}/*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
    ]
    resources = ["*"]
  }
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.cluster.token
}

resource "kubernetes_namespace" "chainguard" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_service_account" "image_copy" {
  metadata {
    name      = "image-copy"
    namespace = var.namespace
    annotations = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.image_copy.arn
    }
  }
}

data "aws_region" "current" {}

resource "kubernetes_cron_job_v1" "image_copy" {
  metadata {
    name      = "image-copy"
    namespace = var.namespace
  }
  spec {
    concurrency_policy        = "Replace"
    failed_jobs_history_limit = 5
    schedule                  = "1 0 * * *"
    job_template {
      metadata {}
      spec {
        backoff_limit = 2
        template {
          metadata {}
          spec {
            service_account_name = kubernetes_service_account.image_copy.metadata[0].name
            container {
              name  = "image-copy"
              image = "${docker_registry_image.image.name}@${docker_registry_image.image.sha256_digest}"

              env {
                name  = "ORG_NAME"
                value = var.org_name
              }

              env {
                name  = "IDENTITY_ID"
                value = chainguard_identity.image_copy.id
              }

              env {
                name  = "DST_REPO_NAME"
                value = var.repo_name
              }

              env {
                name  = "DST_REPO_URI"
                value = aws_ecr_repository.chainguard.repository_url
              }

              env {
                name  = "UPDATED_WITHIN"
                value = var.updated_within
              }

              env {
                name  = "AWS_REGION"
                value = data.aws_region.current.name
              }
            }
          }
        }
      }
    }
  }
}
