#!/bin/bash
# Script to test API Gateway endpoints

set -e

AWS_REGION=${AWS_REGION:-us-east-1}

# Get API URL
API_URL=$(aws cloudformation describe-stacks \
  --stack-name johns-ai-api-gateway \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiEndpoint`].OutputValue' \
  --output text \
  --region $AWS_REGION 2>/dev/null)

if [ -z "$API_URL" ]; then
  echo "Error: API Gateway stack 'johns-ai-api-gateway' not found in region $AWS_REGION."
  echo "Please ensure the stack is deployed."
  exit 1
fi

echo "=== API Gateway Testing ==="
echo "API URL: $API_URL"
echo ""

# Test health endpoint
echo "1. Testing GET /health:"
curl -s "$API_URL/health" | jq '.' || curl -s "$API_URL/health"
echo ""
echo ""

# Test POST /api/clients/add
echo "2. Testing POST /api/clients/add:"
curl -s -X POST "$API_URL/api/clients/add" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "User",
    "email": "test@example.com"
  }' | jq '.' || curl -s -X POST "$API_URL/api/clients/add" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "User",
    "email": "test@example.com"
  }'
echo ""
echo ""

# Test GET /api/clients
echo "3. Testing GET /api/clients:"
curl -s "$API_URL/api/clients" | jq '.' || curl -s "$API_URL/api/clients"
echo ""

