
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

variable "environment" {
    description = "The environment (e.g., production, development)"
    type = string
    default = "production"
}

variable "instance_type" {
    description = "EC2 instance type"
    type = string
    default = "t3.micro"
}

variable "allowed_cidr_blocks" {
    description = "CIDR block allowed to access SSH on EC2 instance"
    type = string
    default = "161.29.129.153/32"
}