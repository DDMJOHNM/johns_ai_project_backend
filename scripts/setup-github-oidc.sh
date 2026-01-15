#!/bin/bash
# Script to set up GitHub Actions OIDC authentication with AWS
set -e

echo "=== GitHub Actions OIDC Setup for AWS ==="
echo ""

# Prompt for required information
read -p "Enter your AWS Account ID (12 digits): " AWS_ACCOUNT_ID
read -p "Enter your GitHub repository (username/repo, e.g., johnmason/john_ai_project): " GITHUB_REPO
read -p "Enter the IAM role name [GitHubActionsDeployRole]: " ROLE_NAME
ROLE_NAME=${ROLE_NAME:-GitHubActionsDeployRole}

# Validate inputs
if [[ ! $AWS_ACCOUNT_ID =~ ^[0-9]{12}$ ]]; then
  echo "Error: AWS Account ID must be 12 digits"
  exit 1
fi

if [[ ! $GITHUB_REPO =~ ^[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$ ]]; then
  echo "Error: GitHub repository must be in format username/repo"
  exit 1
fi

echo ""
echo "Configuration:"
echo "  AWS Account ID: $AWS_ACCOUNT_ID"
echo "  GitHub Repo: $GITHUB_REPO"
echo "  IAM Role Name: $ROLE_NAME"
echo ""
read -p "Continue? (y/n): " CONFIRM

if [[ $CONFIRM != "y" && $CONFIRM != "Y" ]]; then
  echo "Aborted."
  exit 0
fi

echo ""
echo "Step 1: Creating OIDC provider (if not exists)..."
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1 \
  2>/dev/null && echo "✓ OIDC provider created" || echo "✓ OIDC provider already exists"

echo ""
echo "Step 2: Creating trust policy..."
cat > /tmp/github-actions-trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:${GITHUB_REPO}:*"
        }
      }
    }
  ]
}
EOF
echo "✓ Trust policy created at /tmp/github-actions-trust-policy.json"

echo ""
echo "Step 3: Creating IAM role..."
if aws iam create-role \
  --role-name $ROLE_NAME \
  --assume-role-policy-document file:///tmp/github-actions-trust-policy.json \
  --description "Role for GitHub Actions to deploy $GITHUB_REPO" \
  2>/dev/null; then
  echo "✓ IAM role created: $ROLE_NAME"
else
  echo "⚠ Role already exists, updating trust policy..."
  aws iam update-assume-role-policy \
    --role-name $ROLE_NAME \
    --policy-document file:///tmp/github-actions-trust-policy.json
  echo "✓ Trust policy updated"
fi

echo ""
echo "Step 4: Creating permissions policy..."
cat > /tmp/github-actions-permissions.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudformation:*",
        "dynamodb:*",
        "ec2:*",
        "s3:*",
        "ssm:*",
        "iam:GetRole",
        "iam:PassRole",
        "iam:CreateRole",
        "iam:DeleteRole",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy",
        "iam:PutRolePolicy",
        "iam:DeleteRolePolicy",
        "iam:GetRolePolicy",
        "logs:*",
        "apigateway:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF

aws iam put-role-policy \
  --role-name $ROLE_NAME \
  --policy-name GitHubActionsDeployPolicy \
  --policy-document file:///tmp/github-actions-permissions.json

echo "✓ Permissions policy attached"

echo ""
echo "Step 5: Getting role ARN..."
ROLE_ARN=$(aws iam get-role --role-name $ROLE_NAME --query 'Role.Arn' --output text)

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✅ Setup complete!"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "Next steps:"
echo ""
echo "1. Add this secret to your GitHub repository:"
echo ""
echo "   Name:  AWS_ROLE_ARN"
echo "   Value: $ROLE_ARN"
echo ""
echo "2. Go to: https://github.com/$GITHUB_REPO/settings/secrets/actions"
echo "3. Click 'New repository secret'"
echo "4. Paste the values above"
echo ""
echo "5. (Optional) You can now delete these secrets if they exist:"
echo "   - AWS_ACCESS_KEY_ID"
echo "   - AWS_SECRET_ACCESS_KEY"
echo ""
echo "════════════════════════════════════════════════════════════════"
echo ""

# Cleanup
rm -f /tmp/github-actions-trust-policy.json /tmp/github-actions-permissions.json

echo "Role ARN saved to clipboard (if pbcopy is available)..."
echo "$ROLE_ARN" | pbcopy 2>/dev/null || true

