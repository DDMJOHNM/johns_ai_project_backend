
variable "aws_region" {
    description = "The AWS region to deploy the infrastructure to"
    type = string
    default = "us-east-1"
}

variable "stack_name" {
    description = "The name of the stack to deploy the infrastructure to"
    type = string
    default = "john-ai-project"
}