#!/bin/bash
yum update -y

# Install Go
cd /tmp
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile

# Create app directory
mkdir -p /opt/john-ai-project
cd /opt/john-ai-project

# Create systemd service with JWT_SECRET from SSM Parameter Store
cat > /etc/systemd/system/john-ai-backend.service <<EOF
[Unit]
Description=Johns AI Project Backend
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/john-ai-project
ExecStartPre=/bin/bash -c 'export JWT_SECRET=\$(aws ssm get-parameter --name /john-ai-project/jwt-secret --with-decryption --region ${aws_region} --query Parameter.Value --output text 2>/dev/null || echo "")'
ExecStart=/bin/bash -c 'export JWT_SECRET=\$(aws ssm get-parameter --name /john-ai-project/jwt-secret --with-decryption --region ${aws_region} --query Parameter.Value --output text) && /opt/john-ai-project/server'
Restart=always
RestartSec=10
Environment="AWS_REGION=${aws_region}"
Environment="HTTP_PORT=8080"
Environment="DYNAMODB_ENDPOINT="
StandardOutput=journal
StandardError=journal
SyslogIdentifier=john-ai-backend

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable john-ai-backend.service

# Install and configure CloudWatch agent for log forwarding
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U ./amazon-cloudwatch-agent.rpm

# Configure rsyslog to forward application logs to a file
mkdir -p /var/log/john-ai-backend
cat > /etc/rsyslog.d/30-john-ai-backend.conf <<'RSYSLOGEOF'
# Forward john-ai-backend service logs to file
if $programname == 'john-ai-backend' then /var/log/john-ai-backend/application.log
& stop
RSYSLOGEOF

systemctl restart rsyslog

# Create CloudWatch agent configuration
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json <<'CWEOF'
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/john-ai-backend/application.log",
            "log_group_name": "/aws/ec2/john-ai-backend/application",
            "log_stream_name": "{instance_id}",
            "retention_in_days": 7,
            "timezone": "UTC"
          }
        ]
      }
    }
  }
}
CWEOF

# Start CloudWatch agent
/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
  -a fetch-config \
  -m ec2 \
  -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json \
  -s

# The binary will be deployed by GitHub Actions
# Service will start automatically once binary is present

