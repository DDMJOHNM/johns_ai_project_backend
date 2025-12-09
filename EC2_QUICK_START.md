# EC2 + Load Balancer Deployment Checklist

## ✅ Quick Start

### 1. Create EC2 Key Pair (if you don't have one)

```bash
# Create and save the key
aws ec2 create-key-pair --key-name johns-ai-backend-key --region us-east-1 \
  --query 'KeyMaterial' --output text > ~/johns-ai-backend-key.pem
chmod 400 ~/johns-ai-backend-key.pem
```

### 2. Add GitHub Secrets

Go to: **GitHub Repo Settings** → **Secrets and variables** → **Actions**

| Secret | Value |
|--------|-------|
| `AWS_ACCESS_KEY_ID` | Your AWS access key ID |
| `AWS_SECRET_ACCESS_KEY` | Your AWS secret access key |
| `AWS_REGION` | `us-east-1` (or your preferred region) |
| `EC2_KEY_NAME` | `johns-ai-backend-key` |
| `EC2_INSTANCE_TYPE` | `t3.micro` (free tier) or `t3.small` |
| `DEPLOY_EC2` | `true` |

### 3. Deploy

Create and push a tag:
```bash
git tag v0.2.0
git push origin v0.2.0
```

Or trigger manually in GitHub Actions.

### 4. Get Your Backend URL

After deployment completes, check the GitHub Actions job summary for:
```
Load Balancer DNS: johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com
Backend URL: http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com
```

### 5. Test It

```bash
# Health check
curl http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com/health

# Get clients
curl http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com/api/clients
```

## 📊 What Gets Created

✅ **VPC with 2 public subnets** (across 2 availability zones)  
✅ **Internet Gateway** for outbound traffic  
✅ **Security Groups** (ALB and EC2)  
✅ **Application Load Balancer** (port 80)  
✅ **Target Group** with health checks (`/health`)  
✅ **EC2 Instance** running your backend (port 8080)  
✅ **IAM Role** with DynamoDB permissions  
✅ **CloudWatch Logs** for monitoring  

## 🔐 Security Features

- ✅ IAM roles (no hardcoded credentials on EC2)
- ✅ Security groups restrict traffic
- ✅ Health checks ensure instance is healthy
- ✅ Auto-deregistration if instance fails
- ⚠️ **HTTP only** (add HTTPS for production)

## 💰 Estimated Cost

- EC2 t3.micro: ~$7.59/month (eligible for free tier)
- Load Balancer: ~$16.20/month
- Data transfer: ~$0.02/GB
- **Total: ~$24/month**

## 🔗 Connect to API Gateway

Once the EC2 stack is deployed:

1. Get the ALB DNS from deployment output
2. Add GitHub secret: `BACKEND_URL=http://<ALB-DNS>`
3. Push a new tag to deploy API Gateway pointing to it

Or update `infra/api-gateway.yml` manually with the ALB URL.

## 🐛 Debugging

**Check instance status:**
```bash
aws ec2 describe-instances --region us-east-1 \
  --query 'Reservations[].Instances[?Tags[?Key==`Name`&&Value==`*backend*`]]'
```

**SSH into instance:**
```bash
# Get instance IP
INSTANCE_ID=$(aws ec2 describe-instances --region us-east-1 \
  --filters "Name=tag:Name,Values=*backend*" \
  --query 'Reservations[0].Instances[0].PublicIpAddress' --output text)

ssh -i ~/johns-ai-backend-key.pem ec2-user@$INSTANCE_ID
```

**View backend logs:**
```bash
sudo journalctl -u john-ai-backend.service -f
```

**Check ALB targets:**
```bash
aws elbv2 describe-target-health \
  --target-group-arn $(aws elbv2 describe-target-groups \
    --query 'TargetGroups[0].TargetGroupArn' --output text) \
  --region us-east-1
```

## ⚙️ Configuration

**Instance Type Options:**
- `t3.micro` - Free tier (1GB RAM)
- `t3.small` - $18/month (2GB RAM)
- `t3.medium` - $35/month (4GB RAM)

**Change by updating GitHub secret** `EC2_INSTANCE_TYPE`

## 🗑️ Cleanup

To delete all resources:
```bash
aws cloudformation delete-stack \
  --stack-name johns-ai-backend-ec2 --region us-east-1
```

## 📚 Full Documentation

See `docs/EC2_SETUP.md` for detailed setup and troubleshooting.
