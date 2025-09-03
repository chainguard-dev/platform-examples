# iam-self-invoke.tf — minimal, references resources defined in lambda.tf

# Allow the function to invoke itself (for chained indexing)
resource "aws_iam_role_policy" "allow_self_invoke" {
  name = "${var.name_prefix}-allow-self-invoke"
  role = aws_iam_role.lambda.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "AllowSelfInvoke",
        Effect = "Allow",
        Action = ["lambda:InvokeFunction"],
        # unqualified + qualified arns (versions/aliases)
        Resource = [
          aws_lambda_function.lambda.arn,
          "${aws_lambda_function.lambda.arn}:*"
        ]
      }
    ]
  })
}