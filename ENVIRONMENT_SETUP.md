# Environment Setup Guide

This guide explains how to set up the project for local development and production deployment.

## Local Development Environment

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Make

### Initial Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd john_ai_project
   ```

2. **Copy environment template:**
   ```bash
   cp .env.example .env
   ```

3. **Configure local environment:**
   Edit `.env` file with your local settings:
   ```bash
   DYNAMODB_ENDPOINT=http://localhost:8000
   AWS_REGION=us-east-1
   HTTP_PORT=8080
   ```

4. **Start the development environment:**
   ```bash
   make setup
   ```
   This will:
   - Start DynamoDB Local in Docker
   - Create database tables
   - Seed the database with test data

5. **Start the API server:**
   ```bash
   make run-server
   ```

The server will be available at `http://localhost:8080`.

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DYNAMODB_ENDPOINT` | DynamoDB endpoint URL | `http://localhost:8000` | No (for local) |
| `AWS_REGION` | AWS region | `us-east-1` | Yes |
| `HTTP_PORT` | HTTP server port | `8080` | No |
| `AWS_ACCESS_KEY_ID` | AWS access key (production) | - | Yes (production) |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key (production) | - | Yes (production) |

## Production Environment

### AWS Setup

1. **Create DynamoDB Tables:**
   - Use AWS Console or the `create-db` script
   - Ensure tables match the schema defined in the code

2. **Configure IAM Permissions:**
   - Create IAM user/role with DynamoDB permissions
   - Attach policies: `AmazonDynamoDBFullAccess` (or custom policy)

3. **Set Environment Variables:**
   ```bash
   # Remove DYNAMODB_ENDPOINT to use AWS endpoint
   export AWS_REGION=us-east-1
   export AWS_ACCESS_KEY_ID=your-access-key
   export AWS_SECRET_ACCESS_KEY=your-secret-key
   export HTTP_PORT=8080
   ```

### GitHub Actions Deployment

#### Required GitHub Secrets

Configure these in **Settings** → **Secrets and variables** → **Actions**:

- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret key
- `AWS_REGION`: AWS region (optional, defaults to `us-east-1`)

#### Optional Secrets (for specific deployment targets)

- `EC2_HOST`: EC2 instance hostname/IP
- `EC2_USER`: EC2 SSH user (default: `ec2-user`)
- `EC2_SSH_KEY`: SSH private key for EC2
- `DOCKER_REGISTRY`: Docker registry URL (for container deployment)

#### Deployment Process

1. **Automatic Deployment:**
   - Push to `main` branch triggers deployment
   - Version tags (e.g., `v1.0.0`) trigger deployment
   - Manual trigger via GitHub Actions UI

2. **Deployment Steps:**
   - Runs tests
   - Builds production binary
   - Configures AWS credentials
   - Creates/updates DynamoDB tables
   - Uploads deployment artifacts

3. **Post-Deployment:**
   - Binary available in GitHub Actions artifacts
   - Configure your infrastructure to use the binary
   - Set up service management (systemd, supervisor, etc.)

### Deployment Targets

The deployment workflow supports multiple targets (configure as needed):

#### EC2 Deployment

1. Uncomment EC2 deployment steps in `.github/workflows/deploy.yml`
2. Configure EC2 secrets in GitHub
3. Set up SSH access to EC2 instance
4. Configure systemd service on EC2

#### Docker/ECS Deployment

1. Uncomment Docker build/push steps
2. Configure `DOCKER_REGISTRY` secret
3. Set up ECS task definition
4. Update ECS service

#### Lambda Deployment

1. Uncomment Lambda deployment steps
2. Create Lambda function in AWS
3. Configure Lambda environment variables
4. Set up API Gateway if needed

## Troubleshooting

### Local Development Issues

**DynamoDB Local not starting:**
```bash
make docker-down
make docker-up
```

**Connection errors:**
- Verify `.env` file exists and has correct `DYNAMODB_ENDPOINT`
- Check Docker container is running: `make docker-status`
- View logs: `make docker-logs`

**Table creation fails:**
- Ensure DynamoDB Local is running
- Check environment variables are loaded
- Verify AWS_REGION is set

### Production Deployment Issues

**AWS credentials not working:**
- Verify secrets are set in GitHub
- Check IAM permissions
- Ensure AWS_REGION matches your resources

**DynamoDB connection fails:**
- Remove or unset `DYNAMODB_ENDPOINT` for AWS
- Verify IAM permissions for DynamoDB
- Check AWS region configuration

**Build failures:**
- Check Go version (requires 1.21+)
- Verify all dependencies are in `go.mod`
- Review GitHub Actions logs

## Best Practices

1. **Never commit `.env` files** - They contain sensitive information
2. **Use `.env.example`** as a template for required variables
3. **Use GitHub Secrets** for production credentials
4. **Test locally** before pushing to main branch
5. **Use version tags** for production releases
6. **Monitor deployment logs** in GitHub Actions

## Additional Resources

- [AWS DynamoDB Documentation](https://docs.aws.amazon.com/dynamodb/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Environment Variables](https://golang.org/pkg/os/#Getenv)


