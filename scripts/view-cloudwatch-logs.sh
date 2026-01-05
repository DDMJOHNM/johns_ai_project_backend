#!/bin/bash
# Script to view CloudWatch logs for API Gateway and EC2 backend

set -e

AWS_REGION=${AWS_REGION:-us-east-1}
LOG_TYPE=${1:-all}  # Options: api-gateway, ec2, all

# Get API Gateway ID
API_ID=$(aws cloudformation describe-stacks \
  --stack-name johns-ai-api-gateway \
  --query 'Stacks[0].Outputs[?OutputKey==`ApiId`].OutputValue' \
  --output text \
  --region $AWS_REGION 2>/dev/null || echo "")

# Get EC2 Instance ID
INSTANCE_ID=$(aws cloudformation describe-stacks \
  --stack-name johns-ai-backend-ec2 \
  --query 'Stacks[0].Outputs[?OutputKey==`InstanceId`].OutputValue' \
  --output text \
  --region $AWS_REGION 2>/dev/null || echo "")

echo "=== CloudWatch Logs Viewer ==="
echo "Region: $AWS_REGION"
echo ""

if [ "$LOG_TYPE" = "api-gateway" ] || [ "$LOG_TYPE" = "all" ]; then
  if [ -n "$API_ID" ]; then
    echo "📡 API Gateway Logs (Access Logs):"
    echo "Log Group: /aws/apigateway/$API_ID/access"
    echo ""
    echo "Recent access logs:"
    aws logs tail "/aws/apigateway/$API_ID/access" \
      --region $AWS_REGION \
      --since 10m \
      --format short \
      --follow 2>/dev/null || echo "  No logs found or log group doesn't exist yet"
    echo ""
  else
    echo "⚠️  API Gateway stack not found"
    echo ""
  fi
fi

if [ "$LOG_TYPE" = "ec2" ] || [ "$LOG_TYPE" = "all" ]; then
  if [ -n "$INSTANCE_ID" ]; then
    echo "🖥️  EC2 Backend Application Logs:"
    echo "Log Group: /aws/ec2/john-ai-backend/application"
    echo "Instance ID: $INSTANCE_ID"
    echo ""
    echo "Recent application logs:"
    aws logs tail "/aws/ec2/john-ai-backend/application" \
      --region $AWS_REGION \
      --since 10m \
      --format short \
      --follow 2>/dev/null || echo "  No logs found or log group doesn't exist yet"
    echo ""
  else
    echo "⚠️  EC2 backend stack not found"
    echo ""
  fi
fi

echo ""
echo "Usage:"
echo "  $0 api-gateway  # View only API Gateway logs"
echo "  $0 ec2         # View only EC2 application logs"
echo "  $0 all         # View all logs (default)"
echo ""
echo "To view logs in AWS Console:"
if [ -n "$API_ID" ]; then
  echo "  API Gateway: https://console.aws.amazon.com/cloudwatch/home?region=$AWS_REGION#logsV2:log-groups/log-group/%2Faws%2Fapigateway%2F$API_ID"
fi
if [ -n "$INSTANCE_ID" ]; then
  echo "  EC2 Backend: https://console.aws.amazon.com/cloudwatch/home?region=$AWS_REGION#logsV2:log-groups/log-group/%2Faws%2Fec2%2Fjohn-ai-backend%2Fapplication"
fi

