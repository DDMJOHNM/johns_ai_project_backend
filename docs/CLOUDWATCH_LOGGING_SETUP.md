# CloudWatch Logging Setup - Complete

This document describes the per-instance CloudWatch Logs integration for the john-ai-backend service.

## Overview

The backend service now sends all logs to AWS CloudWatch Logs with **per-instance stream naming**. Each EC2 instance creates its own log stream named `api-server-{instance-id}`, making it easy to filter and analyze logs by instance.

## Components

### 1. **Logger Enhancement** (`internal/logger/cloudwatch.go`)

Enhanced the CloudWatch logger to automatically query EC2 instance metadata and append the instance ID to log stream names:

- **Metadata Lookup**: Queries EC2 metadata endpoint (`169.254.169.254/latest/meta-data/instance-id`)
- **Timeout**: Uses 750ms timeout to avoid hanging in non-EC2 environments
- **Fallback**: Falls back to `os.Hostname()` if metadata is unavailable
- **Stream Naming**: Appends instance ID to base stream name → `api-server-{instance-id}`
- **Example**: `api-server-i-0e0eb8403d8afce62`

**Key Code**:
```go
// Attempt to enrich stream name with EC2 instance-id when available
instanceID := ""
mdCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
defer cancel()

req, err := http.NewRequestWithContext(mdCtx, http.MethodGet, "http://169.254.169.254/latest/meta-data/instance-id", nil)
if err == nil {
    client := &http.Client{Timeout: 750 * time.Millisecond}
    resp, err := client.Do(req)
    if err == nil && resp != nil {
        defer resp.Body.Close()
        if resp.StatusCode == http.StatusOK {
            if body, err := io.ReadAll(resp.Body); err == nil {
                instanceID = strings.TrimSpace(string(body))
            }
        }
    }
}

// If we couldn't get instance-id, fall back to hostname (useful for local dev)
if instanceID == "" {
    instanceID, _ = os.Hostname()
}

logStreamName = fmt.Sprintf("%s-%s", logStreamName, instanceID)
```

### 2. **CloudWatch Agent Installation** (`scripts/install-cloudwatch-agent.sh`)

Automated bash script that deploys CloudWatch Agent to EC2 instances via AWS Systems Manager (SSM).

**What it does**:
1. Queries CloudFormation stack for EC2 instance ID
2. Uses SSM RunCommand to execute remote commands on the instance
3. Downloads and installs CloudWatch Agent (Amazon Linux 2 RPM)
4. Creates `/var/log/john-ai-backend.log` for service output
5. Writes CloudWatch Agent configuration to `/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json`
6. Configures agent to tail `/var/log/john-ai-backend.log` with per-instance stream naming
7. Starts/restarts CloudWatch Agent

**Usage**:
```bash
STACK_NAME=johns-ai-backend-ec2 AWS_REGION=us-east-1 ./scripts/install-cloudwatch-agent.sh
```

**Default Stack**: `johns-ai-backend-ec2` (configured via `STACK_NAME` env var)

### 3. **Agent Configuration** (`infra/cloudwatch-agent-config.json`)

Reference CloudWatch Agent configuration template:

```json
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/john-ai-backend.log",
            "log_group_name": "/aws/ec2/john-ai-backend",
            "log_stream_name": "api-server-{instance_id}",
            "timezone": "UTC"
          }
        ]
      }
    }
  }
}
```

**Key Details**:
- **Log Group**: `/aws/ec2/john-ai-backend`
- **Stream Name Template**: `api-server-{instance_id}` — Agent expands `{instance_id}` to actual EC2 instance ID
- **Source File**: `/var/log/john-ai-backend.log` (service stdout/stderr)
- **Timezone**: UTC

## Deployment Steps

### 1. Build the Binary

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/server ./cmd/server
```

### 2. Deploy CloudFormation Stack

Ensure the EC2 instance is running with the correct IAM role that includes CloudWatch Logs permissions:

```bash
aws cloudformation deploy \
  --template-file infra/ec2-backend.yml \
  --stack-name johns-ai-backend-ec2 \
  --region us-east-1 \
  --capabilities CAPABILITY_NAMED_IAM
```

### 3. Deploy Backend Binary

Copy the binary to S3 and deploy to EC2 (existing process).

### 4. Install CloudWatch Agent

```bash
./scripts/install-cloudwatch-agent.sh
```

This script:
- Finds the EC2 instance from CloudFormation stack outputs
- Sends SSM command to install and configure agent
- Waits for completion and displays status

## Log Stream Visibility

Once the agent is running and the service is writing logs:

### View in AWS Console

1. Go to **CloudWatch → Logs → Log Groups**
2. Find `/aws/ec2/john-ai-backend`
3. Streams are named like `api-server-i-0e0eb8403d8afce62`

### View via AWS CLI

```bash
# List streams
aws logs describe-log-streams \
  --log-group-name /aws/ec2/john-ai-backend \
  --region us-east-1

# Tail logs
aws logs tail /aws/ec2/john-ai-backend --follow --region us-east-1

# Tail specific stream
aws logs tail /aws/ec2/john-ai-backend/api-server-i-0e0eb8403d8afce62 --follow
```

## IAM Permissions Required

The EC2 instance IAM role must include CloudWatch Logs permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:CreateLogGroup",
        "logs:DescribeLogStreams",
        "logs:DescribeLogGroups",
        "logs:PutRetentionPolicy"
      ],
      "Resource": "arn:aws:logs:*:*:log-group:/aws/ec2/john-ai-backend*"
    }
  ]
}
```

The CloudFormation template should include `CloudWatchAgentServerPolicy` managed policy (or inline equivalent).

## Service Integration

The systemd service (`/etc/systemd/system/john-ai-backend.service`) is configured to:

1. Redirect service stdout/stderr to `/var/log/john-ai-backend.log`:
   ```ini
   ExecStart=/bin/sh -c "exec /opt/john-ai-project/server >> /var/log/john-ai-backend.log 2>&1"
   ```

2. This allows the CloudWatch Agent to tail the file and push events to CloudWatch Logs

3. Log entries include:
   - Request logs (via `logAndStripHandler` middleware in `internal/router/router.go`)
   - Error logs (from service operations)
   - Startup messages

## Environment Variables

The service respects these env vars (set via systemd service file):

- `AWS_REGION`: CloudWatch region (default: `us-east-1`)
- `HTTP_PORT`: Service port (default: `8080`)
- `DYNAMODB_ENDPOINT`: DynamoDB endpoint (empty = AWS service)
- `CLOUDWATCH_LOG_GROUP`: CloudWatch log group name (default: `/aws/ec2/john-ai-backend`)
- `CLOUDWATCH_LOG_STREAM`: Base stream name (default: `api-server`, expanded with instance ID)

## Troubleshooting

### No Log Streams Visible

1. **Check agent status on instance**:
   ```bash
   aws ssm send-command \
     --instance-ids i-0e0eb8403d8afce62 \
     --document-name "AWS-RunShellScript" \
     --parameters 'commands=["/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -m ec2 -a status"]' \
     --region us-east-1
   ```

2. **Check log file exists and has content**:
   ```bash
   aws ssm send-command \
     --instance-ids i-0e0eb8403d8afce62 \
     --document-name "AWS-RunShellScript" \
     --parameters 'commands=["ls -la /var/log/john-ai-backend.log","tail -20 /var/log/john-ai-backend.log"]' \
     --region us-east-1
   ```

3. **Verify IAM role has permissions**:
   - Check EC2 instance IAM role includes `CloudWatchAgentServerPolicy`

4. **Check service is running**:
   ```bash
   aws ssm send-command \
     --instance-ids i-0e0eb8403d8afce62 \
     --document-name "AWS-RunShellScript" \
     --parameters 'commands=["systemctl status john-ai-backend.service"]' \
     --region us-east-1
   ```

### Log Stream Naming

- Expected format: `api-server-{instance-id}`
- Example: `api-server-i-0e0eb8403d8afce62`

If the instance ID is not appearing:
1. Check EC2 instance can reach metadata endpoint (port 80 from instance)
2. Verify security groups allow 169.254.169.254 traffic
3. Check CloudWatch Agent logs for metadata fetch errors

## Maintenance

### Update Agent Configuration

If you need to change the CloudWatch Agent config:

1. Update `/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json` on instance
2. Reload agent:
   ```bash
   sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
     -a fetch-config \
     -m ec2 \
     -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json \
     -s
   ```

### Log Retention

CloudWatch Logs retention can be set via CloudFormation or CLI:

```bash
aws logs put-retention-policy \
  --log-group-name /aws/ec2/john-ai-backend \
  --retention-in-days 30
```

## Files Modified/Created

- ✅ `internal/logger/cloudwatch.go` — Enhanced to query instance ID
- ✅ `scripts/install-cloudwatch-agent.sh` — New installation script
- ✅ `infra/cloudwatch-agent-config.json` — New config template

## References

- [AWS CloudWatch Agent Documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Install-CloudWatch-Agent.html)
- [EC2 Instance Metadata Service](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html)
- [CloudWatch Logs Quotas](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/cloudwatch_limits_cwl.html)
