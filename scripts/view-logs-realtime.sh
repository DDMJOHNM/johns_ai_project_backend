#!/bin/bash
# Script to view backend server logs in real-time using SSM Session Manager
# Note: This requires AWS Session Manager plugin to be installed

INSTANCE_ID=$(aws cloudformation describe-stacks \
  --stack-name johns-ai-backend-ec2 \
  --query 'Stacks[0].Outputs[?OutputKey==`InstanceId`].OutputValue' \
  --output text \
  --region ${AWS_REGION:-us-east-1} 2>/dev/null)

if [ -z "$INSTANCE_ID" ]; then
  echo "Error: Could not find EC2 instance. Is the stack deployed?"
  exit 1
fi

echo "Connecting to instance: $INSTANCE_ID"
echo "To view logs in real-time, run: journalctl -u john-ai-backend.service -f"
echo "Press Ctrl+D to exit"
echo ""

# Use SSM Session Manager to connect interactively
aws ssm start-session \
  --target $INSTANCE_ID \
  --region ${AWS_REGION:-us-east-1} \
  --document-name AWS-StartInteractiveCommand \
  --parameters command="journalctl -u john-ai-backend.service -f"

