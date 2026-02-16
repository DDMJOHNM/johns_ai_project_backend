# API Gateway HTTP API
resource "aws_apigatewayv2_api" "main" {
  name          = "${var.stack_name}-api-${var.environment}"
  protocol_type = "HTTP"
  description   = "API Gateway for ${var.stack_name}"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"]
    allow_headers = ["*"]
    max_age       = 300
  }

  tags = {
    Name        = "${var.stack_name}-api"
    Project     = var.stack_name
    Environment = var.environment
  }
}

# Integration with EC2 backend
resource "aws_apigatewayv2_integration" "backend" {
  api_id           = aws_apigatewayv2_api.main.id
  integration_type = "HTTP_PROXY"
  integration_uri  = var.backend_url

  integration_method = "ANY"
  payload_format_version = "1.0"
}

# Default route - catch all
resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.backend.id}"
}

# Specific route for /health
resource "aws_apigatewayv2_route" "health" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /health"
  target    = "integrations/${aws_apigatewayv2_integration.backend.id}"
}

# Route for auth endpoints
resource "aws_apigatewayv2_route" "auth" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "ANY /api/auth/{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.backend.id}"
}

# Route for clients endpoints
resource "aws_apigatewayv2_route" "clients" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "ANY /api/clients/{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.backend.id}"
}

# Stage
resource "aws_apigatewayv2_stage" "prod" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = var.stage_name
  auto_deploy = true

  access_log_settings {
    destination_arn = var.cloudwatch_log_group_arn
    format = jsonencode({
      httpMethod     = "$context.httpMethod"
      ip             = "$context.identity.sourceIp"
      protocol       = "$context.protocol"
      requestId      = "$context.requestId"
      requestTime    = "$context.requestTime"
      responseLength = "$context.responseLength"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      error = {
        message       = "$context.error.message"
        messageString = "$context.error.messageString"
      }
    })
  }

  tags = {
    Name        = "${var.stack_name}-${var.stage_name}"
    Environment = var.environment
  }
}

