# JWT Secret Setup Guide

This guide explains how to securely configure the JWT_SECRET for your Johns AI Project backend using AWS Systems Manager Parameter Store.

## Overview

The JWT_SECRET is a critical security component used to sign and verify JWT tokens for user authentication. In production, this secret is:
- **Stored securely** in AWS Systems Manager Parameter Store as an encrypted SecureString
- **Automatically retrieved** by the EC2 instance at startup via IAM permissions
- **Never exposed** in code, logs, or CloudFormation templates

## Quick Setup

### Option 1: Automated Script (Recommended)

Run the setup script which will guide you through the process:

```bash
./scripts/setup-jwt-secret.sh
```

This script will:
1. Check your AWS credentials
2. Generate a secure random JWT secret (or let you provide your own)
3. Store it in AWS SSM Parameter Store
4. Provide next steps for deployment

### Option 2: Manual Setup

#### Step 1: Generate a Secure Secret

```bash
openssl rand -base64 48
```

Save this value - you'll need it in the next step.

#### Step 2: Store in AWS SSM Parameter Store

```bash
aws ssm put-parameter \
  --name "/john-ai-project/jwt-secret" \
  --value "YOUR_GENERATED_SECRET_HERE" \
  --type "SecureString" \
  --description "JWT secret for Johns AI Project authentication" \
  --region us-east-1
```

#### Step 3: Deploy CloudFormation Stack

The CloudFormation template (`infra/ec2-backend.yml`) is pre-configured to:
- Grant SSM read permissions to the EC2 instance
- Fetch the JWT secret at service startup
- Inject it as an environment variable

Deploy the stack:

```bash
aws cloudformation create-stack \
  --stack-name john-ai-backend \
  --template-body file://infra/ec2-backend.yml \
  --capabilities CAPABILITY_IAM \
  --region us-east-1
```

Or update an existing stack:

```bash
aws cloudformation update-stack \
  --stack-name john-ai-backend \
  --template-body file://infra/ec2-backend.yml \
  --capabilities CAPABILITY_IAM \
  --region us-east-1
```

## How It Works

### Architecture

```
┌─────────────────────────────────────┐
│   AWS Systems Manager               │
│   Parameter Store                   │
│                                     │
│   /john-ai-project/jwt-secret      │
│   (SecureString - Encrypted)       │
└──────────────┬──────────────────────┘
               │ IAM Permission
               │ ssm:GetParameter
               ▼
┌─────────────────────────────────────┐
│   EC2 Instance                      │
│   ┌───────────────────────────────┐ │
│   │ systemd service               │ │
│   │ john-ai-backend.service       │ │
│   │                               │ │
│   │ ExecStart:                    │ │
│   │ 1. Fetch JWT_SECRET from SSM  │ │
│   │ 2. Export as env variable     │ │
│   │ 3. Start Go server            │ │
│   └───────────────────────────────┘ │
└─────────────────────────────────────┘
```

### IAM Permissions

The EC2 instance role includes this policy:

```yaml
- PolicyName: SSMParameterAccess
  PolicyDocument:
    Version: '2012-10-17'
    Statement:
      - Effect: Allow
        Action:
          - ssm:GetParameter
          - ssm:GetParameters
        Resource: !Sub 'arn:aws:ssm:${AWS::Region}:${AWS::AccountId}:parameter/john-ai-project/*'
```

### Systemd Service Configuration

The service automatically fetches the secret on startup:

```ini
[Service]
ExecStart=/bin/bash -c 'export JWT_SECRET=$(aws ssm get-parameter --name /john-ai-project/jwt-secret --with-decryption --region us-east-1 --query Parameter.Value --output text) && /opt/john-ai-project/bin/server'
```

## Managing the Secret

### View Parameter Info (Without Value)

```bash
aws ssm get-parameter \
  --name "/john-ai-project/jwt-secret" \
  --region us-east-1
```

### Update the Secret

```bash
aws ssm put-parameter \
  --name "/john-ai-project/jwt-secret" \
  --value "NEW_SECRET_HERE" \
  --type "SecureString" \
  --overwrite \
  --region us-east-1
```

After updating, restart the service on EC2:

```bash
# SSH into your EC2 instance
ssh ec2-user@YOUR_EC2_IP

# Restart the service to pick up new secret
sudo systemctl restart john-ai-backend

# Verify it's running
sudo systemctl status john-ai-backend
```

### Delete the Secret

```bash
aws ssm delete-parameter \
  --name "/john-ai-project/jwt-secret" \
  --region us-east-1
```

⚠️ **Warning:** Deleting the secret will cause authentication to fail. Only delete if you're decommissioning the application.

## Verification

### Verify Parameter Exists

```bash
aws ssm get-parameter \
  --name "/john-ai-project/jwt-secret" \
  --region us-east-1
```

### Verify EC2 Can Access Parameter

SSH into your EC2 instance and run:

```bash
aws ssm get-parameter \
  --name /john-ai-project/jwt-secret \
  --with-decryption \
  --region us-east-1 \
  --query Parameter.Value \
  --output text
```

If this returns the secret value, the permissions are correct.

### Verify Service is Running

```bash
sudo systemctl status john-ai-backend
```

Check logs:

```bash
sudo journalctl -u john-ai-backend -f
```

## Troubleshooting

### "Parameter not found" Error

**Problem:** The EC2 service can't find the JWT secret parameter.

**Solution:**
1. Verify the parameter exists:
   ```bash
   aws ssm get-parameter --name /john-ai-project/jwt-secret --region us-east-1
   ```
2. Ensure you're using the correct region
3. Check the parameter name matches exactly

### "Access Denied" Error

**Problem:** EC2 instance doesn't have permission to read the parameter.

**Solution:**
1. Verify the EC2 instance has the correct IAM role attached
2. Check the IAM role has the SSMParameterAccess policy
3. Redeploy the CloudFormation stack if needed

### Service Won't Start

**Problem:** The systemd service fails to start.

**Solution:**
1. Check service logs:
   ```bash
   sudo journalctl -u john-ai-backend -n 50
   ```
2. Verify the JWT secret is being retrieved:
   ```bash
   aws ssm get-parameter --name /john-ai-project/jwt-secret --with-decryption --region us-east-1
   ```
3. Manually test the ExecStart command:
   ```bash
   export JWT_SECRET=$(aws ssm get-parameter --name /john-ai-project/jwt-secret --with-decryption --region us-east-1 --query Parameter.Value --output text)
   echo "JWT_SECRET length: ${#JWT_SECRET}"
   ```

## Security Best Practices

✅ **DO:**
- Use AWS Systems Manager Parameter Store for production secrets
- Generate cryptographically strong random secrets (48+ bytes)
- Rotate secrets periodically
- Use IAM policies to restrict access to the parameter
- Monitor access to the parameter using CloudTrail

❌ **DON'T:**
- Hard-code secrets in source code
- Commit secrets to version control
- Share secrets via email or chat
- Use weak or predictable secrets
- Store secrets in plain text files

## Local Development

For local development, you can set the JWT_SECRET directly as an environment variable:

```bash
export JWT_SECRET="dev-secret-key-change-in-production"
./bin/server
```

This bypasses SSM Parameter Store and uses the environment variable directly.

## Related Documentation

- [Authentication System](./AUTHENTICATION.md) - Complete authentication documentation
- [EC2 Setup](./EC2_SETUP.md) - EC2 instance setup and configuration
- [Deployment Scripts](../scripts/) - Automated deployment tools

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Review CloudFormation stack events
3. Check EC2 system logs and service logs
4. Verify IAM permissions are correct

