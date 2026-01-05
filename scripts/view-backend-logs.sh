#!/bin/bash
# Script to view backend server logs on EC2

INSTANCE_ID=$(aws cloudformation describe-stacks \
  --stack-name johns-ai-backend-ec2 \
  --query 'Stacks[0].Outputs[?OutputKey==`InstanceId`].OutputValue' \
  --output text \
  --region ${AWS_REGION:-us-east-1} 2>/dev/null)

if [ -z "$INSTANCE_ID" ]; then
  echo "Error: Could not find EC2 instance. Is the stack deployed?"
  exit 1
fi

echo "Viewing logs for instance: $INSTANCE_ID"
echo "Press Ctrl+C to stop"
echo ""

# View logs in real-time
aws ssm start-session \
  --target $INSTANCE_ID \
  --region ${AWS_REGION:-us-east-1} \
  --document-name AWS-StartInteractiveCommand \
  --parameters command="journalctl -u john-ai-backend.service -f" 2>/dev/null || \
  aws ssm send-command \
    --instance-ids $INSTANCE_ID \
    --document-name "AWS-RunShellScript" \
    --parameters "commands=[\"journalctl -u john-ai-backend.service -n 100 --no-pager\"]" \
    --region ${AWS_REGION:-us-east-1} \
    --query 'Command.CommandId' \
    --output text | xargs -I {} sh -c 'sleep 3 && aws ssm get-command-invocation --command-id {} --instance-id $INSTANCE_ID --region ${AWS_REGION:-us-east-1} --query "StandardOutputContent" --output text'

