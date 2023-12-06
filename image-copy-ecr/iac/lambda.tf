data "aws_iam_policy_document" "lambda" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "ecr-pusher" {
  // Permissions needed to push to the ECR repository.
  // https://docs.aws.amazon.com/AmazonECR/latest/userguide/image-push.html#image-push-iam

  statement {
    effect = "Allow"
    actions = [
      "ecr:CreateRepository", // Also need to create repositories under the target repository.

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
      aws_ecr_repository.repo.arn,
      "${aws_ecr_repository.repo.arn}/*"
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

data "aws_iam_policy" "lambda" {
  name = "AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role" "lambda" {
  name               = "image-copy"
  assume_role_policy = data.aws_iam_policy_document.lambda.json
  inline_policy {
    name   = "ecr-pusher"
    policy = data.aws_iam_policy_document.ecr-pusher.json
  }
  managed_policy_arns = [data.aws_iam_policy.lambda.arn]
}

resource "aws_ecr_repository_policy" "policy" {
  repository = aws_ecr_repository.repo.name
  policy     = data.aws_iam_policy_document.lambda.json
}

resource "aws_ecr_repository" "repo" {
  name                 = var.dst_repo
  image_tag_mutability = var.immutable_tags ? "IMMUTABLE" : "MUTABLE"

  image_scanning_configuration {
    scan_on_push = false
  }
}

// Create a sub-repository for the image-copy lambda code.
resource "aws_ecr_repository" "copier-repo" {
  name                 = "${var.dst_repo}/image-copy"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = false
  }
}

resource "ko_build" "image" {
  repo        = aws_ecr_repository.copier-repo.repository_url
  importpath  = "github.com/chainguard-dev/platform-examples/image-copy-ecr"
  working_dir = path.module
  // Disable SBOM generation due to
  // https://github.com/ko-build/ko/issues/878
  sbom = "none"
}

locals {
  // Using a local for the lambda breaks a cyclic dependency between
  // chainguard_identity.aws and aws_lambda_function.lambda
  lambda_name = "image-copy"
}

data "aws_region" "current" {}

resource "aws_lambda_function" "lambda" {
  function_name = local.lambda_name
  role          = aws_iam_role.lambda.arn

  package_type = "Image"
  image_uri    = ko_build.image.image_ref

  environment {
    variables = {
      GROUP          = var.group
      IDENTITY       = chainguard_identity.aws.id
      ISSUER_URL     = "https://issuer.enforce.dev"
      DST_REPO       = var.dst_repo
      FULL_DST_REPO  = aws_ecr_repository.repo.repository_url
      REGION         = data.aws_region.current.name
      IMMUTABLE_TAGS = var.immutable_tags
    }
  }
}

resource "aws_lambda_function_url" "lambda" {
  function_name      = aws_lambda_function.lambda.function_name
  authorization_type = "NONE"
}
