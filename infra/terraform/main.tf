terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "duskaotearoa-terraform-state"
    key            = "john-ai-project-backend/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-state-lock"
    encrypt        = true
  }
}

provider "aws" {
    region = var.aws_region
}

# Data source for AWS account ID
data "aws_caller_identity" "current" {}

module "iam" {
  source = "./modules/iam"
  stack_name = var.stack_name
  aws_region = var.aws_region
}

module "dynamodb" {
  source = "./modules/dynamodb"
  stack_name = var.stack_name
  environment = var.environment
}

module "ec2" {
  source = "./modules/ec2"
  stack_name = var.stack_name
  aws_region = var.aws_region
  aws_account_id = data.aws_caller_identity.current.account_id
  instance_type = var.instance_type
  environment = var.environment
  allowed_cidr_blocks = var.allowed_cidr_blocks
}

module "api_gateway" {
  source = "./modules/api-gateway"
  stack_name = var.stack_name
  environment = var.environment
  stage_name = "prod"
  backend_url = module.ec2.backend_url
  cloudwatch_log_group_arn = "arn:aws:logs:us-east-1:051826704696:log-group:/aws/apigateway/mos5j2g72f"
}