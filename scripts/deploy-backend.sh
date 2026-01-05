#!/bin/bash
# Script to deploy backend server binary to EC2

set -e

INSTANCE_ID=$(aws cloudformation describe-stacks \
  --stack-name johns-ai-backend-ec2 \
  --query 'Stacks[0].Outputs[?OutputKey==`InstanceId`].OutputValue' \
  --output text \
  --region ${AWS_REGION:-us-east-1} 2>/dev/null)

if [ -z "$INSTANCE_ID" ]; then
  echo "Error: Could not find EC2 instance. Is the stack deployed?"
  exit 1
fi

if [ ! -f "bin/server" ]; then
  echo "Error: bin/server not found. Run 'make build-server' first."
  exit 1
fi

REGION=${AWS_REGION:-us-east-1}
BUCKET_NAME="johns-ai-backend-deploy-$(date +%s)"

echo "Deploying to instance: $INSTANCE_ID"
echo "1. Creating temporary S3 bucket..."

# Create temporary S3 bucket
aws s3 mb "s3://$BUCKET_NAME" --region $REGION

echo "2. Uploading binary..."
aws s3 cp bin/server "s3://$BUCKET_NAME/server" --region $REGION

echo "3. Deploying to EC2..."
aws ssm send-command \
  --instance-ids $INSTANCE_ID \
  --document-name "AWS-RunShellScript" \
  --parameters "commands=[
    \"systemctl stop john-ai-backend.service || true\",
    \"aws s3 cp s3://$BUCKET_NAME/server /opt/john-ai-project/server --region $REGION\",
    \"chmod +x /opt/john-ai-project/server\",
    \"systemctl start john-ai-backend.service\",
    \"sleep 3\",
    \"systemctl status john-ai-backend.service --no-pager -l\"
  ]" \
  --region $REGION \
  --output text

echo "4. Cleaning up S3 bucket..."
aws s3 rb "s3://$BUCKET_NAME" --force --region $REGION

echo ""
echo "✓ Deployment complete!"
echo ""
echo "View logs with: ./scripts/view-backend-logs.sh"
echo "Or run: aws ssm send-command --instance-ids $INSTANCE_ID --document-name 'AWS-RunShellScript' --parameters 'commands=[\"journalctl -u john-ai-backend.service -n 50 --no-pager\"]' --region $REGION --query 'Command.CommandId' --output text"

