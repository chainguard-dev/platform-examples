# Assumes you already have:
# - provider "aws" with var.aws_region
# - data "aws_caller_identity" "current"
# - var.dst_prefix (e.g., "bannon.dev" or "bannon.dev/<subprefix>")
# - aws_iam_role.lambda (your function role)

locals {
  # Two resource ARNs: the prefix repo itself *and* its children
  ecr_repo_arn_root   = "arn:aws:ecr:${var.aws_region}:${data.aws_caller_identity.current.account_id}:repository/${var.dst_prefix}"
  ecr_repo_arn_prefix = "arn:aws:ecr:${var.aws_region}:${data.aws_caller_identity.current.account_id}:repository/${var.dst_prefix}/*"
}

data "aws_iam_policy_document" "lambda_ecr_fullsync" {
  statement {
    sid     = "EcrWriteAndDescribeForMirror"
    effect  = "Allow"
    actions = [
      # pushing/copying
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchGetImage",
      "ecr:CompleteLayerUpload",
      "ecr:CreateRepository",
      "ecr:DescribeImages",
      "ecr:DescribeRepositories",
      "ecr:GetDownloadUrlForLayer",
      "ecr:GetRepositoryPolicy",
      "ecr:InitiateLayerUpload",
      "ecr:ListImages",
      "ecr:PutImage",
      "ecr:UploadLayerPart",

      # repository policy management on your namespace
      "ecr:SetRepositoryPolicy",
      "ecr:DeleteRepositoryPolicy",

      # optional quality-of-life on repos you own
      "ecr:ListTagsForResource",
      "ecr:TagResource",
      "ecr:UntagResource",

      # optional: uncomment if you want the lambda to tune repo settings
      # "ecr:PutImageTagMutability",
      # "ecr:PutImageScanningConfiguration",
      # "ecr:PutLifecyclePolicy",
      # "ecr:DeleteLifecyclePolicy",
    ]
    resources = [
      local.ecr_repo_arn_root,
      local.ecr_repo_arn_prefix,
    ]
  }

  # AWS requires GetAuthorizationToken on "*"
  statement {
    sid     = "AuthToken"
    effect  = "Allow"
    actions = ["ecr:GetAuthorizationToken"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "lambda_ecr_fullsync" {
  name        = "${var.name_prefix}-ecr-fullsync"
  description = "Allow Lambda to push images and manage repo policies for ${var.dst_prefix}"
  policy      = data.aws_iam_policy_document.lambda_ecr_fullsync.json
}

resource "aws_iam_role_policy_attachment" "lambda_ecr_fullsync" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.lambda_ecr_fullsync.arn
}
