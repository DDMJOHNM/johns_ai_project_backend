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

