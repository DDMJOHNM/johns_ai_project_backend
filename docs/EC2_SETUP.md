# EC2 Backend with Load Balancer Setup Guide

## Overview

This guide explains how to deploy your Johns AI Project backend to AWS EC2 with an Application Load Balancer (ALB) for high availability and automatic health checks.

## Architecture

```
┌─────────────────────────────────────────────┐
│          Internet (clients)                  │
└────────────────────┬────────────────────────┘
                     │
          ┌──────────▼──────────┐
          │  Application Load   │
          │    Balancer (ALB)   │
          │   (Port 80/443)     │
          └──────────┬──────────┘
                     │
       ┌─────────────┴────────────────────┐
       │                                   │
       ▼                                   ▼
┌──────────────┐                  ┌──────────────┐
│  EC2 Instance│ (Subnet 1)       │  EC2 Instance│ (Subnet 2)
│  Port 8080   │ AZ 1             │  Port 8080   │ AZ 2
└──────┬───────┘                  └──────┬───────┘
       │                                   │
       └─────────────┬─────────────────────┘
                     │
                     ▼
            ┌────────────────────┐
            │  DynamoDB Tables   │
            │  (clients, users)  │
            └────────────────────┘
```

## Prerequisites

1. **AWS Account** with EC2, Load Balancer, DynamoDB, and CloudFormation permissions
2. **GitHub Secrets** configured in your repository:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_REGION` (optional, defaults to us-east-1)
   - `EC2_INSTANCE_TYPE` (optional, defaults to t3.micro)
   - `DEPLOY_EC2` (set to `true` to enable EC2 deployment)

**Note:** No EC2 key pair is required. Access to instances is managed via AWS Systems Manager Session Manager.

## Step 1: Add GitHub Secrets

In your GitHub repository:
1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add the following secrets:

| Secret | Value | Example |
|--------|-------|---------|
| `AWS_ACCESS_KEY_ID` | Your AWS access key | `AKIA...` |
| `AWS_SECRET_ACCESS_KEY` | Your AWS secret key | `aws...` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `EC2_INSTANCE_TYPE` | Instance type | `t3.micro` or `t3.small` |
| `DEPLOY_EC2` | Enable EC2 deployment | `true` |

## Step 2: Deploy

### Option A: Trigger with Git Tag

```bash
git tag v0.2.0
git push origin v0.2.0
```

### Option B: Manual Trigger

In GitHub:
1. Go to **Actions**
2. Select **Deploy to Production**
3. Click **Run workflow**

## Step 3: Monitor Deployment

1. Go to your GitHub repository → **Actions**
2. Check the deploy job logs
3. Look for the **Load Balancer DNS** in the step summary

The deployment will:
1. Validate CloudFormation templates
2. Create DynamoDB tables
3. Deploy EC2 instance(s)
4. Create Application Load Balancer
5. Configure auto-scaling and health checks
6. Seed the database

## Step 4: Use Load Balancer URL

Once deployment completes, you'll get a Load Balancer DNS like:
```
johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com
```

### Test the Backend

```bash
# Health check
curl http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com/health

# Get all clients
curl http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com/api/clients
```

### Update API Gateway

To use this backend with API Gateway:
1. Add GitHub secret: `BACKEND_URL=http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com`
2. Re-deploy with a new tag

Or update the API Gateway CloudFormation template with the ALB URL.

## Step 5: Access EC2 Instance (if needed)

You can access the instance using AWS Systems Manager Session Manager (no SSH key required):

```bash
# List EC2 instances
aws ec2 describe-instances --region us-east-1 \
  --filters "Name=tag:Name,Values=johns-ai-backend-ec2-backend" \
  --query 'Reservations[0].Instances[0].InstanceId'

# Start a session to the instance
aws ssm start-session --target i-xxxxx --region us-east-1

# Once connected, view logs
sudo journalctl -u john-ai-backend.service -f

# Check service status
sudo systemctl status john-ai-backend.service
```

**Benefits of Session Manager:**
- ✅ No SSH key management
- ✅ Encrypted connections
- ✅ CloudTrail logging of all commands
- ✅ Fine-grained IAM permissions
- ✅ Works with instances in private subnets

## Costs

**Estimated Monthly Costs (US East 1):**
- **t3.micro EC2**: $7.59/month (on-demand, eligible for free tier)
- **Application Load Balancer**: $16.20/month
- **Data Transfer**: ~$0.02 per GB
- **DynamoDB**: Pay-per-request (usually < $1/month for testing)

**Total**: ~$23-25/month (may be less with free tier)

## Scaling

To scale to multiple instances:

1. Update the CloudFormation template to use an Auto Scaling Group instead of a single instance
2. Configure target group with multiple AZs
3. Adjust health check thresholds

Example:
```yaml
InstancesPerAZ: 2
MinSize: 2
MaxSize: 4
```

## Cleanup

To delete all deployed resources:

```bash
# Delete EC2 stack
aws cloudformation delete-stack --stack-name johns-ai-backend-ec2 --region us-east-1

# Delete API Gateway stack (if deployed)
aws cloudformation delete-stack --stack-name johns-ai-api-gateway --region us-east-1

# Delete DynamoDB stack
aws cloudformation delete-stack --stack-name johns-ai-dynamodb --region us-east-1
```

## Troubleshooting

### EC2 instance not starting

```bash
# Check instance status
aws ec2 describe-instances \
  --instance-ids i-xxxxx \
  --query 'Reservations[0].Instances[0].State' \
  --region us-east-1

# Check system logs
aws ec2 get-console-output --instance-id i-xxxxx --region us-east-1
```

### Load Balancer shows unhealthy targets

```bash
# Check target health
aws elbv2 describe-target-health \
  --target-group-arn arn:aws:elasticloadbalancing:... \
  --region us-east-1

# Use Session Manager to check instance logs
aws ssm start-session --target i-xxxxx --region us-east-1
# Then inside the session:
sudo journalctl -u john-ai-backend.service -n 50
```

### API Gateway can't reach backend

1. Verify ALB DNS is correct: `curl http://<ALB-DNS>/health`
2. Check API Gateway CloudFormation backend URL
3. Verify security groups allow traffic from ALB to EC2

## Security Considerations

For production:
1. **Enable HTTPS** with a certificate (update ALB listener)
2. **Restrict SSH**: Update security group to allow SSH only from specific IPs
3. **Use RDS** for data instead of local storage
4. **Set up CloudWatch alarms** for performance monitoring
5. **Enable VPC Flow Logs** for network monitoring
6. **Rotate credentials** regularly
7. **Use IAM roles** instead of access keys (already configured)

## Next Steps

1. Monitor CloudWatch logs: `/aws/ec2/john-ai-backend`
2. Set up alarms for CPU and memory usage
3. Configure auto-scaling policies
4. Add HTTPS with AWS Certificate Manager
5. Enable AWS WAF for DDoS protection
