#!/bin/bash
# Script to update existing EC2 instance with JWT_SECRET support

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

REGION=${AWS_REGION:-us-east-1}
STACK_NAME=${STACK_NAME:-johns-ai-backend-ec2}
PARAMETER_NAME="/john-ai-project/jwt-secret"

echo -e "${GREEN}Updating EC2 Backend with JWT_SECRET Support${NC}"
echo "=============================================="
echo ""

# Step 1: Check if JWT secret exists in SSM
echo -e "${YELLOW}Step 1: Checking JWT secret in SSM Parameter Store...${NC}"
if aws ssm get-parameter --name "$PARAMETER_NAME" --region "$REGION" &>/dev/null; then
    echo -e "${GREEN}✓ JWT secret exists in Parameter Store${NC}"
else
    echo -e "${YELLOW}JWT secret not found. Creating one...${NC}"
    JWT_SECRET=$(openssl rand -base64 48)
    aws ssm put-parameter \
        --name "$PARAMETER_NAME" \
        --value "$JWT_SECRET" \
        --type "SecureString" \
        --description "JWT secret for Johns AI Project authentication" \
        --region "$REGION"
    echo -e "${GREEN}✓ JWT secret created${NC}"
fi
echo ""

# Step 2: Get Instance ID
echo -e "${YELLOW}Step 2: Finding EC2 instance...${NC}"
INSTANCE_ID=$(aws cloudformation describe-stacks \
    --stack-name "$STACK_NAME" \
    --query 'Stacks[0].Outputs[?OutputKey==`InstanceId`].OutputValue' \
    --output text \
    --region "$REGION" 2>/dev/null)

if [ -z "$INSTANCE_ID" ]; then
    echo -e "${RED}Error: Could not find instance. Is the stack deployed?${NC}"
    echo "Stack name: $STACK_NAME"
    exit 1
fi

echo -e "${GREEN}✓ Found instance: $INSTANCE_ID${NC}"
echo ""

# Step 3: Update systemd service
echo -e "${YELLOW}Step 3: Updating systemd service on EC2...${NC}"
COMMAND_ID=$(aws ssm send-command \
    --instance-ids "$INSTANCE_ID" \
    --document-name "AWS-RunShellScript" \
    --parameters 'commands=[
        "echo \"Stopping service...\"",
        "sudo systemctl stop john-ai-backend.service || true",
        "echo \"Updating systemd service file...\"",
        "sudo bash -c '\''cat > /etc/systemd/system/john-ai-backend.service <<\"SYSTEMDEOF\"
[Unit]
Description=Johns AI Project Backend
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/john-ai-project
ExecStartPre=/bin/bash -c \"export JWT_SECRET=\\$(aws ssm get-parameter --name /john-ai-project/jwt-secret --with-decryption --region '$REGION' --query Parameter.Value --output text 2>/dev/null || echo \\\"\\\")\"
ExecStart=/bin/bash -c \"export JWT_SECRET=\\$(aws ssm get-parameter --name /john-ai-project/jwt-secret --with-decryption --region '$REGION' --query Parameter.Value --output text) && /opt/john-ai-project/server\"
Restart=always
RestartSec=10
Environment=\"AWS_REGION='$REGION'\"
Environment=\"HTTP_PORT=8080\"
Environment=\"DYNAMODB_ENDPOINT=\"
StandardOutput=journal
StandardError=journal
SyslogIdentifier=john-ai-backend

[Install]
WantedBy=multi-user.target
SYSTEMDEOF'\''",
        "echo \"Reloading systemd...\"",
        "sudo systemctl daemon-reload",
        "echo \"Starting service...\"",
        "sudo systemctl start john-ai-backend.service",
        "sleep 3",
        "echo \"Service status:\"",
        "sudo systemctl status john-ai-backend.service --no-pager -l || true",
        "echo \"\"",
        "echo \"Recent logs:\"",
        "sudo journalctl -u john-ai-backend.service -n 10 --no-pager"
    ]' \
    --region "$REGION" \
    --query 'Command.CommandId' \
    --output text)

echo -e "${GREEN}✓ Command sent: $COMMAND_ID${NC}"
echo ""
echo "Waiting for command to complete..."
sleep 5

# Step 4: Get command output
echo ""
echo -e "${YELLOW}Step 4: Checking command results...${NC}"
aws ssm get-command-invocation \
    --command-id "$COMMAND_ID" \
    --instance-id "$INSTANCE_ID" \
    --region "$REGION" \
    --query 'StandardOutputContent' \
    --output text

echo ""
echo -e "${GREEN}=============================================="
echo "✓ Update Complete!"
echo "==============================================
${NC}"
echo ""
echo "Next steps:"
echo "1. Test the health endpoint:"
BACKEND_URL=$(aws cloudformation describe-stacks \
    --stack-name "$STACK_NAME" \
    --query 'Stacks[0].Outputs[?OutputKey==`BackendURL`].OutputValue' \
    --output text \
    --region "$REGION" 2>/dev/null)

if [ -n "$BACKEND_URL" ]; then
    echo "   curl $BACKEND_URL/health"
fi
echo ""
echo "2. View logs:"
echo "   ./scripts/view-backend-logs.sh"
echo ""
echo "3. Test authentication:"
echo "   See docs/AUTHENTICATION.md"

