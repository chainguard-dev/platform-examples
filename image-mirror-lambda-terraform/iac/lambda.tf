############################
# ECR repo that holds the LAMBDA CONTAINER IMAGE (not the mirrored images)
############################
resource "aws_ecr_repository" "lambda_image" {
  name                 = var.name_prefix
  image_tag_mutability = "MUTABLE"
  image_scanning_configuration { scan_on_push = true }
}

############################
# Build the container from ../cmd/fullsync and push to the repo above
############################
resource "ko_build" "image" {
  importpath  = var.go_importpath
  working_dir = ".."
  repo        = aws_ecr_repository.lambda_image.repository_url
  sbom        = "none"
}

############################
# Lambda execution role
############################
resource "aws_iam_role" "lambda" {
  name = var.name_prefix
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Sid       = "LambdaAssume",
      Effect    = "Allow",
      Action    = "sts:AssumeRole",
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}

# Basic Lambda logging
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

############################
# ECR permissions so the Lambda can create and push mirror repos/images at runtime
# (fixed deprecated region attribute by using var.aws_region)
############################
resource "aws_iam_role_policy" "ecr_pusher" {
  name = "ecr-pusher"
  role = aws_iam_role.lambda.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "ECRRepoWide",
        Effect = "Allow",
        Action = [
          "ecr:CreateRepository",
          "ecr:DescribeRepositories",
          "ecr:DescribeImages",
          "ecr:ListImages",
          "ecr:GetRepositoryPolicy",
          "ecr:SetRepositoryPolicy",
          "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:PutImage"
        ],
        Resource = "arn:aws:ecr:${var.aws_region}:${data.aws_caller_identity.current.account_id}:repository/*"
      },
      {
        Sid     = "ECRAuthToken",
        Effect  = "Allow",
        Action  = [ "ecr:GetAuthorizationToken" ],
        Resource = "*"
      }
    ]
  })
}

############################
# Lambda function (container image from ko_build)
############################
resource "aws_lambda_function" "lambda" {
  function_name = var.name_prefix
  package_type  = "Image"
  role          = aws_iam_role.lambda.arn
  image_uri     = ko_build.image.image_ref
  timeout       = 900

  # If you want strict one-at-a-time chaining, uncomment:
  # reserved_concurrent_executions = 1

  environment {
    variables = {
      SRC_REGISTRY = var.src_registry
      GROUP_NAME   = var.group_name
      DST_PREFIX   = var.dst_prefix

      # Pull-token credentials for cgr.dev (username=identity id, password=JWT)
      CGR_USERNAME = var.cgr_username
      CGR_PASSWORD = var.cgr_password

      # Optional knobs for your chaining main.go
      REPO_LIST_JSON = jsonencode(var.repo_list)
      COPY_ALL_TAGS  = tostring(var.copy_all_tags)
      REPO_TAGS_JSON = jsonencode(var.repo_tags)

      # Booleans must be strings in Lambda env
      MIRROR_DRY_RUN = tostring(var.mirror_dry_run)

      # START_INDEX         = "0"
      # MIRROR_DRY_RUN      = "false"
      # REPO_LIST_JSON      = ""
      # REPO_LIST_CSV       = ""
      # REPO_LIST_SSM_PARAM = ""
    }
  }
}

############################
# Public Function URL (easy test trigger)
############################
resource "aws_lambda_function_url" "lambda" {
  function_name      = aws_lambda_function.lambda.function_name
  authorization_type = "NONE"
}