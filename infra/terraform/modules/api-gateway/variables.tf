variable "stack_name" {
  description = "The name of the stack/project"
  type        = string
  default     = "john-ai-project"
}

variable "environment" {
  description = "The environment (e.g., production, development)"
  type        = string
  default     = "production"
}

variable "stage_name" {
  description = "API Gateway stage name"
  type        = string
  default     = "prod"
}

variable "backend_url" {
  description = "Backend URL for API Gateway integration"
  type        = string
}

variable "cloudwatch_log_group_arn" {
  description = "CloudWatch Log Group ARN for API Gateway access logs"
  type        = string
  default     = ""
}

