# GitHub Actions AWS OIDC Setup Guide

This guide explains how to configure AWS OIDC authentication for GitHub Actions, which is more secure than using long-lived AWS access keys.

## Overview

The workflow has been updated to use **OpenID Connect (OIDC)** to authenticate with AWS. This method:
- ✅ Uses temporary credentials that expire automatically
- ✅ No need to store long-lived AWS access keys in GitHub Secrets
- ✅ More secure and follows AWS best practices
- ✅ Uses IAM role assumption

## Setup Steps

### 1. Create an OIDC Provider in AWS

First, create an OIDC identity provider in your AWS account to trust GitHub Actions:

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1 \
  --region us-east-1
```

**Note**: You only need to do this **once per AWS account**.

### 2. Create an IAM Role for GitHub Actions

Create a file called `github-actions-trust-policy.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::YOUR_AWS_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:YOUR_GITHUB_USERNAME/YOUR_REPO_NAME:*"
        }
      }
    }
  ]
}
```

**Replace**:
- `YOUR_AWS_ACCOUNT_ID` with your AWS account ID (12-digit number)
- `YOUR_GITHUB_USERNAME` with your GitHub username
- `YOUR_REPO_NAME` with your repository name (e.g., `john_ai_project`)

Example:
```json
"token.actions.githubusercontent.com:sub": "repo:johnmason/john_ai_project:*"
```

**For more security**, you can restrict to specific branches:
```json
"token.actions.githubusercontent.com:sub": "repo:johnmason/john_ai_project:ref:refs/heads/main"
```

### 3. Create the IAM Role

```bash
aws iam create-role \
  --role-name GitHubActionsDeployRole \
  --assume-role-policy-document file://github-actions-trust-policy.json \
  --description "Role for GitHub Actions to deploy john_ai_project"
```

### 4. Attach Permissions to the Role

Create a permissions policy file called `github-actions-permissions.json`:

```json
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
```

Attach the policy:

```bash
aws iam put-role-policy \
  --role-name GitHubActionsDeployRole \
  --policy-name GitHubActionsDeployPolicy \
  --policy-document file://github-actions-permissions.json
```

### 5. Get the Role ARN

```bash
aws iam get-role --role-name GitHubActionsDeployRole --query 'Role.Arn' --output text
```

This will output something like:
```
arn:aws:iam::123456789012:role/GitHubActionsDeployRole
```

### 6. Add GitHub Secret

1. Go to your GitHub repository
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add the following secret:

   - **Name**: `AWS_ROLE_ARN`
   - **Value**: The ARN from step 5 (e.g., `arn:aws:iam::123456789012:role/GitHubActionsDeployRole`)

5. Ensure you also have the `AWS_REGION` secret set (or it will default to `us-east-1`)

### 7. Remove Old Secrets (Optional but Recommended)

Since you're now using OIDC, you can **delete** these secrets:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

## Verification

To verify the setup is working:

1. Push a tag to trigger the deployment workflow:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. Or manually trigger the workflow from GitHub:
   - Go to **Actions** → **Deploy to Production** → **Run workflow**

3. Check the "Configure AWS credentials" step in the workflow logs. It should show:
   ```
   Assuming role with OIDC
   Role ARN: arn:aws:iam::123456789012:role/GitHubActionsDeployRole
   ```

## Troubleshooting

### Error: "The security token included in the request is invalid"

**Causes**:
1. **OIDC provider not created** - Make sure you created the OIDC provider in step 1
2. **Trust policy mismatch** - Verify the repository name in the trust policy matches your GitHub repo exactly
3. **Wrong AWS account ID** - Check the account ID in the trust policy and role ARN
4. **Role ARN secret not set** - Verify `AWS_ROLE_ARN` secret is set correctly in GitHub

### Error: "User is not authorized to perform: sts:AssumeRoleWithWebIdentity"

**Solution**: The trust policy in the IAM role needs to include the correct GitHub repository. Check step 2.

### Error: "Not authorized to perform: cloudformation:CreateStack"

**Solution**: The role needs more permissions. Review and update the permissions policy in step 4.

## Quick Setup Script

Here's a complete script to automate the setup:

```bash
#!/bin/bash
set -e

# Configuration - EDIT THESE VALUES
AWS_ACCOUNT_ID="YOUR_AWS_ACCOUNT_ID"
GITHUB_REPO="YOUR_GITHUB_USERNAME/YOUR_REPO_NAME"  # e.g., "johnmason/john_ai_project"
ROLE_NAME="GitHubActionsDeployRole"

echo "Setting up GitHub Actions OIDC for AWS..."

# 1. Create OIDC provider (if not exists)
echo "Creating OIDC provider..."
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1 \
  2>/dev/null || echo "OIDC provider already exists"

# 2. Create trust policy
cat > /tmp/trust-policy.json <<EOF
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

# 3. Create IAM role
echo "Creating IAM role..."
aws iam create-role \
  --role-name $ROLE_NAME \
  --assume-role-policy-document file:///tmp/trust-policy.json \
  --description "Role for GitHub Actions to deploy" \
  2>/dev/null || echo "Role already exists"

# 4. Attach permissions
echo "Attaching permissions..."
cat > /tmp/permissions.json <<EOF
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
  --policy-document file:///tmp/permissions.json

# 5. Get role ARN
ROLE_ARN=$(aws iam get-role --role-name $ROLE_NAME --query 'Role.Arn' --output text)

echo ""
echo "✅ Setup complete!"
echo ""
echo "Add this to your GitHub repository secrets:"
echo "Name: AWS_ROLE_ARN"
echo "Value: $ROLE_ARN"
echo ""
echo "GitHub Settings → Secrets and variables → Actions → New repository secret"
```

Save this as `setup-github-oidc.sh`, edit the configuration values at the top, and run:

```bash
chmod +x setup-github-oidc.sh
./setup-github-oidc.sh
```

## References

- [GitHub Actions OIDC Documentation](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
- [AWS IAM OIDC Identity Providers](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html)
- [aws-actions/configure-aws-credentials](https://github.com/aws-actions/configure-aws-credentials)

