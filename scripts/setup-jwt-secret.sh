#!/bin/bash
# Script to set up JWT_SECRET in AWS Systems Manager Parameter Store

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
PARAMETER_NAME="/john-ai-project/jwt-secret"
AWS_REGION="us-east-1"

echo -e "${GREEN}JWT Secret Setup for Johns AI Project${NC}"
echo "========================================"
echo ""

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}Error: AWS CLI is not installed.${NC}"
    echo "Please install it: https://aws.amazon.com/cli/"
    exit 1
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}Error: AWS credentials not configured.${NC}"
    echo "Please run: aws configure"
    exit 1
fi

echo -e "${YELLOW}Current AWS Identity:${NC}"
aws sts get-caller-identity
echo ""

# Ask user if they want to generate a new secret or provide their own
echo "Choose an option:"
echo "1) Generate a secure random JWT secret (recommended)"
echo "2) Provide my own JWT secret"
read -p "Enter choice [1 or 2]: " choice

case $choice in
    1)
        # Generate a secure random secret
        JWT_SECRET=$(openssl rand -base64 48)
        echo -e "${GREEN}Generated secure JWT secret!${NC}"
        ;;
    2)
        # User provides their own secret
        read -sp "Enter your JWT secret: " JWT_SECRET
        echo ""
        if [ -z "$JWT_SECRET" ]; then
            echo -e "${RED}Error: JWT secret cannot be empty${NC}"
            exit 1
        fi
        ;;
    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${YELLOW}Storing JWT secret in AWS SSM Parameter Store...${NC}"
echo "Parameter name: $PARAMETER_NAME"
echo "Region: $AWS_REGION"
echo ""

# Check if parameter already exists
if aws ssm get-parameter --name "$PARAMETER_NAME" --region "$AWS_REGION" &> /dev/null; then
    echo -e "${YELLOW}Warning: Parameter already exists.${NC}"
    read -p "Do you want to overwrite it? [y/N]: " overwrite
    
    if [[ $overwrite =~ ^[Yy]$ ]]; then
        aws ssm put-parameter \
            --name "$PARAMETER_NAME" \
            --value "$JWT_SECRET" \
            --type "SecureString" \
            --description "JWT secret for Johns AI Project authentication" \
            --overwrite \
            --region "$AWS_REGION"
        echo -e "${GREEN}✓ JWT secret updated successfully!${NC}"
    else
        echo -e "${YELLOW}Aborted. Existing parameter not modified.${NC}"
        exit 0
    fi
else
    # Create new parameter
    aws ssm put-parameter \
        --name "$PARAMETER_NAME" \
        --value "$JWT_SECRET" \
        --type "SecureString" \
        --description "JWT secret for Johns AI Project authentication" \
        --region "$AWS_REGION"
    echo -e "${GREEN}✓ JWT secret stored successfully!${NC}"
fi

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Deploy or update your CloudFormation stack:"
echo "   aws cloudformation create-stack \\"
echo "     --stack-name john-ai-backend \\"
echo "     --template-body file://infra/ec2-backend.yml \\"
echo "     --capabilities CAPABILITY_IAM \\"
echo "     --region $AWS_REGION"
echo ""
echo "2. If the stack already exists and EC2 is running, restart the service:"
echo "   ssh ec2-user@YOUR_EC2_IP"
echo "   sudo systemctl restart john-ai-backend"
echo ""
echo "To verify the parameter:"
echo "aws ssm get-parameter --name $PARAMETER_NAME --region $AWS_REGION"
echo ""
echo -e "${YELLOW}Note: The secret value is encrypted and secure.${NC}"

