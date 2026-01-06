#!/bin/bash
# Quick script to view API Gateway access logs

set -e

AWS_REGION=${AWS_REGION:-us-east-1}
SINCE=${1:-10m}  # Default: last 10 minutes

# Get API Gateway ID
API_ID=$(aws cloudformation describe-stacks \
  --stack-name johns-ai-api-gateway \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiId`].OutputValue' \
  --output text \
  --region $AWS_REGION 2>/dev/null || echo "")

if [ -z "$API_ID" ]; then
  echo "Error: API Gateway stack not found"
  exit 1
fi

LOG_GROUP="/aws/apigateway/$API_ID"

echo "📡 API Gateway Access Logs"
echo "Log Group: $LOG_GROUP"
echo "Since: $SINCE"
echo "Region: $AWS_REGION"
echo ""
echo "Press Ctrl+C to stop following logs"
echo ""

# View logs with follow mode
aws logs tail "$LOG_GROUP" \
  --region $AWS_REGION \
  --since "$SINCE" \
  --format short \
  --follow

