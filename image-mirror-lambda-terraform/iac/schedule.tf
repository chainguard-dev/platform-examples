# Run every 4 hours
resource "aws_cloudwatch_event_rule" "every_4_hours" {
  name                = "chainguard-mirror-every-4-hours"
  description         = "Invoke the image-copy-all Lambda every 4 hours"
  schedule_expression = "rate(4 hours)"
}

# Give EventBridge permission to invoke the Lambda
resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.every_4_hours.arn
}

# Connect the rule to your Lambda, and start at index 0
resource "aws_cloudwatch_event_target" "lambda_target" {
  rule      = aws_cloudwatch_event_rule.every_4_hours.name
  target_id = "lambda"
  arn       = aws_lambda_function.lambda.arn

  # 👇 This kicks off the self-chaining at the first repo.
  input = jsonencode({
    index = 0
  })
}