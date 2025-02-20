data "aws_caller_identity" "current" {}

data "chainguard_group" "group" {
  name = var.group_name
}

resource "chainguard_identity" "aws" {
  parent_id   = data.chainguard_group.group.id
  name        = "aws-lambda-identity"
  description = "Identity for AWS Lambda"

  aws_identity {
    aws_account         = data.aws_caller_identity.current.account_id
    aws_user_id_pattern = "^AROA(.*):${local.lambda_name}$"

    // NB: This role will be assumed so can't use the role ARN directly. We must used the ARN of the assumed role
    aws_arn = "arn:aws:sts::${data.aws_caller_identity.current.account_id}:assumed-role/${aws_iam_role.lambda.name}/${local.lambda_name}"
  }
}

# Look up the registry.pull role to grant the identity.
data "chainguard_role" "puller" {
  name = "registry.pull"
}

resource "chainguard_rolebinding" "puller" {
  identity = chainguard_identity.aws.id
  role     = data.chainguard_role.puller.items[0].id
  group    = data.chainguard_group.group.id
}

# Create a subscription to notify the Lambda function on changes under the root group.
resource "chainguard_subscription" "subscription" {
  parent_id = data.chainguard_group.group.id
  sink      = aws_lambda_function_url.lambda.function_url
}
