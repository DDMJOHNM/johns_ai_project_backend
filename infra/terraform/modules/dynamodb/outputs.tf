output "clients_table_name" {
  description = "Name of the Clients DynamoDB table"
  value       = aws_dynamodb_table.clients.name
}

output "clients_table_arn" {
  description = "ARN of the Clients DynamoDB table"
  value       = aws_dynamodb_table.clients.arn
}

output "users_table_name" {
  description = "Name of the Users DynamoDB table"
  value       = aws_dynamodb_table.users.name
}

output "users_table_arn" {
  description = "ARN of the Users DynamoDB table"
  value       = aws_dynamodb_table.users.arn
}

