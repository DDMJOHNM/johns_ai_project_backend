output "api_id" {
  description = "API Gateway ID"
  value       = aws_apigatewayv2_api.main.id
}

output "api_endpoint" {
  description = "API Gateway endpoint URL"
  value       = aws_apigatewayv2_api.main.api_endpoint
}

output "api_endpoint_with_stage" {
  description = "Full API Gateway endpoint with stage"
  value       = "${aws_apigatewayv2_api.main.api_endpoint}/${aws_apigatewayv2_stage.prod.name}"
}

output "stage_name" {
  description = "API Gateway stage name"
  value       = aws_apigatewayv2_stage.prod.name
}

output "integration_id" {
  description = "API Gateway integration ID"
  value       = aws_apigatewayv2_integration.backend.id
}

