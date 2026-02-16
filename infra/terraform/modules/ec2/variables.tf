variable "stack_name" {
  description = "The name of the stack/project"
  type        = string
  default     = "john-ai-project"
}

variable "aws_region" {
  description = "The AWS region"
  type        = string
  default     = "us-east-1"
}

variable "aws_account_id" {
  description = "The AWS account ID"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "environment" {
  description = "The environment (e.g., production, development)"
  type        = string
  default     = "production"
}

variable "allowed_cidr_blocks" {
  description = "CIDR block allowed to access SSH"
  type        = string
  default     = "161.29.129.153/32"
}
