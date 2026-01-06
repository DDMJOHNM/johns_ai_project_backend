#!/bin/bash
# Script to get the API Gateway URL

set -e

AWS_REGION=${AWS_REGION:-us-east-1}

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

echo "$API_URL"

