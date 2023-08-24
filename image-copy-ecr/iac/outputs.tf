output "url" {
  value = aws_lambda_function_url.lambda.function_url
}

output "dst_repo" {
  value = aws_ecr_repository.repo.repository_url
}
