output "lambda_function_name" {
  value = aws_lambda_function.lambda.function_name
}

output "lambda_function_url" {
  value = aws_lambda_function_url.lambda.function_url
}

output "lambda_image_repository_url" {
  value = aws_ecr_repository.lambda_image.repository_url
}