#!/usr/bin/env bash
# Installs and configures the Amazon CloudWatch Agent on the backend EC2 instance(s).
# This script sends an SSM RunCommand to the instance returned by the CloudFormation stack output.
# It will:
#  - download and install the CloudWatch Agent RPM (if not already installed)
#  - create /var/log/john-ai-backend.log and update the systemd unit to write stdout/stderr there
#  - write a CloudWatch Agent config to /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json
#  - start/restart the CloudWatch Agent

set -euo pipefail
STACK_NAME=${STACK_NAME:-johns-ai-backend-ec2}
REGION=${AWS_REGION:-us-east-1}

INSTANCE_ID=$(aws cloudformation describe-stacks \
  --stack-name "$STACK_NAME" \
  --query 'Stacks[0].Outputs[?OutputKey==`InstanceId`].OutputValue' \
  --output text \
  --region "$REGION" 2>/dev/null)

if [ -z "$INSTANCE_ID" ]; then
  echo "Error: Could not find EC2 instance. Is the stack deployed?"
  exit 1
fi

echo "Will configure CloudWatch Agent on instance: $INSTANCE_ID (region: $REGION)"

# The shell script to run on the instance via SSM
# First write a temporary local script file that will be uploaded via SSM parameters
cat > /tmp/john_ai_install_cw_agent.sh <<'REMOTE'
#!/usr/bin/env bash
set -euo pipefail
# Download & install CloudWatch agent if not present
AGENT_CTL="/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl"
if [ ! -x "$AGENT_CTL" ]; then
  echo "Installing CloudWatch Agent..."
  curl -sS -o /tmp/amazon-cloudwatch-agent.rpm https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
  sudo rpm -Uvh /tmp/amazon-cloudwatch-agent.rpm
fi

# Ensure log file exists and has permissive perms
sudo mkdir -p /var/log
sudo touch /var/log/john-ai-backend.log
sudo chown root:root /var/log/john-ai-backend.log
sudo chmod 0644 /var/log/john-ai-backend.log

# Update systemd unit to redirect stdout/stderr to the log file (append)
UNIT_PATH="/etc/systemd/system/john-ai-backend.service"
if [ -f "$UNIT_PATH" ]; then
  if ! grep -q "StandardOutput" "$UNIT_PATH"; then
    sudo sed -i '/\[Service\]/a StandardOutput=append:/var/log/john-ai-backend.log\nStandardError=append:/var/log/john-ai-backend.log' "$UNIT_PATH"
    sudo systemctl daemon-reload || true
    sudo systemctl restart john-ai-backend.service || true
  else
    echo "systemd unit already configured to redirect output."
  fi
else
  echo "Warning: systemd unit $UNIT_PATH not found. Make sure your service logs to /var/log/john-ai-backend.log or journal."
fi

# Write CloudWatch Agent config
sudo mkdir -p /opt/aws/amazon-cloudwatch-agent/etc
cat <<'CWCONF' | sudo tee /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json >/dev/null
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
CWCONF

# Start or reload the agent with this configuration
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json -s || true

# Confirm agent status
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -m ec2 -a status || true

echo "CloudWatch Agent configured (if the instance has network + IAM permissions)."
REMOTE

# Build the SSM parameters JSON from the local script (split into lines)
python3 - <<'PY' > /tmp/ssm_params.json
import json
with open('/tmp/john_ai_install_cw_agent.sh','r') as f:
    lines = f.read().splitlines()
print(json.dumps({'commands': lines}))
PY

# Send the command via SSM using a parameters file
COMMAND_ID=$(aws ssm send-command \
  --instance-ids "$INSTANCE_ID" \
  --document-name "AWS-RunShellScript" \
  --comment "Install and configure CloudWatch Agent for john-ai-backend" \
  --parameters file:///tmp/ssm_params.json \
  --timeout-seconds 600 \
  --region "$REGION" \
  --query Command.CommandId --output text)

if [ -z "$COMMAND_ID" ]; then
  echo "Failed to create SSM command"
  exit 2
fi

echo "SSM command sent (CommandId: $COMMAND_ID). Waiting for completion..."
# Wait for command to finish
aws ssm wait command-executed --command-id "$COMMAND_ID" --instance-id "$INSTANCE_ID" --region "$REGION"

# Fetch command output
aws ssm get-command-invocation --command-id "$COMMAND_ID" --instance-id "$INSTANCE_ID" --region "$REGION" --query '{Status:Status,StandardOutput:StandardOutputContent,StandardError:StandardErrorContent}' --output json

echo "Done. If the instance has IAM permissions (logs:CreateLogStream, logs:PutLogEvents) you should see new log streams under /aws/ec2/john-ai-backend named like api-server-<instance-id>."
