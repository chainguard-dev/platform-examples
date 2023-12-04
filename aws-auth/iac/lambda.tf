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

data "aws_iam_policy" "lambda" {
  name = "AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role" "lambda" {
  name                = "aws-auth"
  assume_role_policy  = data.aws_iam_policy_document.lambda.json
  managed_policy_arns = [data.aws_iam_policy.lambda.arn]
}

resource "aws_ecr_repository" "app-repo" {
  name = "aws-auth"
}

resource "ko_build" "image" {
  repo        = aws_ecr_repository.app-repo.repository_url
  importpath  = "github.com/chainguard-dev/enforce-events/aws-auth/cmd/app"
  working_dir = path.module
  // Disable SBOM generation due to
  // https://github.com/ko-build/ko/issues/878
  sbom = "none"
}

locals {
  // Using a local for the lambda breaks a cyclic dependency between
  // chainguard_identity.aws and aws_lambda_function.lambda
  lambda_name = "aws-auth"
}

data "aws_region" "current" {}

resource "aws_lambda_function" "lambda" {
  function_name = local.lambda_name
  role          = aws_iam_role.lambda.arn

  package_type = "Image"
  image_uri    = ko_build.image.image_ref

  environment {
    variables = {
      GROUP        = var.group
      IDENTITY     = chainguard_identity.aws.id
      ISSUER_URL   = "https://issuer.enforce.dev"
      API_ENDPOINT = "https://console-api.enforce.dev"
    }
  }
}

// Allow anyone to invoke the lambda.
resource "aws_lambda_function_url" "lambda" {
  function_name      = aws_lambda_function.lambda.function_name
  authorization_type = "NONE"
}
