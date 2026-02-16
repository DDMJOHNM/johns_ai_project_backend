# DynamoDB Table - Clients
resource "aws_dynamodb_table" "clients" {
  name           = "clients"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  tags = {
    Name        = "clients"
    Project     = var.stack_name
    Environment = var.environment
  }
}

# DynamoDB Table - Users
resource "aws_dynamodb_table" "users" {
  name           = "users"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  tags = {
    Name        = "users"
    Project     = var.stack_name
    Environment = var.environment
  }
}

