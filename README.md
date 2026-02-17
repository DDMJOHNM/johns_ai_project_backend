# Mental Health Counselling Client Management System

A system to manage clients receiving counselling for mental health.

Customer On Boarding Flow

Architecture Overview  
The system uses a Next.js frontend that integrates an OpenAI-powered onboarding assistant. Users provide their details in natural language (voice or text), which the frontend sends directly to an OpenAI agent. The agent extracts structured fields (first name, last name, email) and returns them to the UI for user review. Once confirmed, the frontend sends the validated payload to a Go backend service exposed via AWS API Gateway. The backend performs additional validation and persists the client record in DynamoDB. The entire backend stack is deployed through GitHub Actions and provisioned using Makefile-driven AWS infrastructure.

```
┌──────────────────────────┐
│          User            │
│  (speaks info)  │
└─────────────┬────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│        Next.js Frontend (Client UI)          │
│  - Chat-style onboarding assistant            │
│  - Transcript + Detected Details              │
│  - State management + validation              │
└─────────────┬────────────────────────────────┘
              │ Prompt + conversation context
              ▼
┌──────────────────────────────────────────────┐
│          OpenAI Agent / LLM                  │
│  - Natural language → structured fields       │
│  - Extracts: first name, last name, email     │
│  - Returns JSON-like structured output        │
└─────────────┬────────────────────────────────┘
              │ Structured fields
              ▼
┌──────────────────────────────────────────────┐
│        Next.js Frontend (Review Step)        │
│  - Shows parsed fields                        │
│  - User confirms or edits                     │
└─────────────┬────────────────────────────────┘
              │ Validated payload
              ▼
┌──────────────────────────────────────────────┐
│      AWS API Gateway (HTTPS endpoint)        │
└─────────────┬────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│   Go Backend Service (johns_ai_project_backend) │
│  - Validation                                  │
│  - Business logic                              │
│  - Error handling                              │
│  - Persistence layer                           │
└─────────────┬────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│        DynamoDB (Client Records)             │
│  - First name                                 │
│  - Last name                                  │
│  - Email                                      │
│  - Metadata                                   │
└──────────────────────────────────────────────┘

┌──────────────────────────────────────────────┐
│ GitHub Actions CI/CD                         │
│ - Build & test                                │
│ - Deploy backend to AWS                       │
│ - Provision infra via Makefile/Terraform      │
└──────────────────────────────────────────────┘

```



## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Available Commands](#available-commands)
- [Environment Setup](#environment-setup)
- [Database Configuration](#database-configuration)
- [Database Schema](#database-schema)
- [API Server](#api-server)
- [API Endpoints](#api-endpoints)
- [Testing the API](#testing-the-api)
- [API Gateway](#api-gateway)
- [CI/CD with GitHub Actions](#cicd-with-github-actions)
- [AWS Deployment](#aws-deployment)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Make
- AWS CLI (for deployment)

## Quick Start

1. **Start DynamoDB Local container:**
   ```bash
   make docker-up
   ```

2. **Create DynamoDB tables:**
   ```bash
   make setup-db
   ```

3. **Seed DynamoDB with test data:**
   ```bash
   make seed-db
   ```

Or run everything at once:
```bash
make setup
```

4. **Start the API server:**
   ```bash
   make run-server
   ```

The server will be available at `http://localhost:8080`.

## Available Commands

### Debugging
- `docker stop` - john_ai_backend


### Docker Commands
- `make docker-up` - Start DynamoDB Local container
- `make docker-down` - Stop DynamoDB Local container
- `make docker-logs` - View DynamoDB container logs
- `make docker-status` - Check DynamoDB container status

### Database Commands
- `make setup-db` - Create DynamoDB tables (clients and users)
- `make seed-db` - Seed DynamoDB with test data
- `make test-db` - Run setup-db and seed-db
- `make verify` - Verify tables exist and have data

### Build Commands
- `make build` - Build all Go binaries
- `make build-create-db` - Build create-db binary
- `make build-seed-db` - Build seed-db binary
- `make build-example` - Build example client service binary
- `make run-example` - Run example client service
- `make build-server` - Build API server binary
- `make run-server` - Run API server (default port 8080)

### API Gateway Commands
- `make deploy-api-gateway` - Deploy API Gateway (requires BACKEND_URL)
- `make get-api-url` - Get API Gateway endpoint URL
- `make test-api-gateway` - Test API Gateway endpoints
- `make delete-api-gateway` - Delete API Gateway stack

### Setup & Cleanup
- `make setup` - Full setup (docker-up + test-db)
- `make clean` - Clean build artifacts and binaries

### Environment Variables
- `DYNAMODB_ENDPOINT` - DynamoDB endpoint (default: http://localhost:8000)
- `AWS_REGION` - AWS region (default: us-east-1)
- `HTTP_PORT` - HTTP server port (default: 8080)
- `JWT_SECRET` - Secret key for JWT token signing (required for authentication)
- `BACKEND_URL` - Backend server URL for API Gateway deployment

## Environment Setup

### Local Development Environment

#### Initial Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd johns_ai_project_backend
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
   JWT_SECRET=dev-secret-key-change-in-production
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

#### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DYNAMODB_ENDPOINT` | DynamoDB endpoint URL | `http://localhost:8000` | No (for local) |
| `AWS_REGION` | AWS region | `us-east-1` | Yes |
| `HTTP_PORT` | HTTP server port | `8080` | No |
| `JWT_SECRET` | JWT token signing secret | `dev-secret-key-...` | Yes (production) |
| `AWS_ACCESS_KEY_ID` | AWS access key (production) | - | Yes (production) |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key (production) | - | Yes (production) |

### Production Environment

#### AWS Setup

1. **Create DynamoDB Tables:**
   - Use AWS Console or the `create-db` script
   - Ensure tables match the schema defined in the code

2. **Configure IAM Permissions:**
   - Create IAM user/role with DynamoDB permissions
   - Attach policies: `AmazonDynamoDBFullAccess` (or custom policy)

3. **Set up JWT Secret (Required for Authentication):**
   ```bash
   # Use the automated setup script
   ./scripts/setup-jwt-secret.sh
   ```
   
   See [JWT Secret Setup Guide](docs/JWT_SECRET_SETUP.md) for detailed instructions.

4. **Set Environment Variables:**
   ```bash
   # Remove DYNAMODB_ENDPOINT to use AWS endpoint
   export AWS_REGION=us-east-1
   export AWS_ACCESS_KEY_ID=your-access-key
   export AWS_SECRET_ACCESS_KEY=your-secret-key
   export HTTP_PORT=8080
   # JWT_SECRET is automatically loaded from SSM Parameter Store on EC2
   ```

## Database Configuration

Environment variables are configured in the `.env` file. Copy `.env.example` to `.env` and modify as needed:

```bash
cp .env.example .env
```

Default DynamoDB settings (in `.env`):
- `DYNAMODB_ENDPOINT=http://localhost:8000` - DynamoDB Local endpoint
- `AWS_REGION=us-east-1` - AWS region

You can override these by:
1. Editing the `.env` file
2. Setting environment variables when running commands:
   ```bash
   DYNAMODB_ENDPOINT=http://localhost:8000 AWS_REGION=us-east-1 make setup-db
   ```

## Database Schema

### Clients Table
- **Primary Key:** `id` (String)
- **Global Secondary Indexes:**
  - `email-index` - Query by email
  - `status-index` - Query by status
- Stores client information including personal details, contact information, and emergency contacts
- Status: active, inactive, archived

### Users Table
- **Primary Key:** `id` (String)
- **Global Secondary Indexes:**
  - `username-index` - Query by username
  - `email-index` - Query by email
  - `role-index` - Query by role
- Stores system users (admin, counsellors, staff)
- Roles: admin, counsellor, staff
- Status: active, inactive, suspended

## API Server

The API server provides REST endpoints to interact with the client data.

### Starting the Server

```bash
make run-server
```

The server will start on `http://localhost:8080` (or the port specified in `HTTP_PORT` environment variable).

## API Endpoints

### Base URL

- Local: `http://localhost:8080`
- Port can be configured via `HTTP_PORT` environment variable (default: 8080)

### Health Check

**GET** `/health`

Check if the server is running.

**Response:**
```json
{
  "status": "ok",
  "message": "Server is running"
}
```

### Get All Clients

**GET** `/api/clients`

Retrieves all clients from the database.

**Response:**
```json
[
  {
    "id": "client-001",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "555-0101",
    "date_of_birth": "1985-03-15",
    "address": "123 Main St, Anytown, ST 12345",
    "emergency_contact_name": "Jane Doe",
    "emergency_contact_phone": "555-0102",
    "status": "active",
    "created_at": "2025-11-11T18:00:00Z",
    "updated_at": "2025-11-11T18:00:00Z"
  }
]
```

### Get Client by ID

**GET** `/api/clients/{id}`

Retrieves a single client by their ID.

**Parameters:**
- `id` (path parameter) - The client ID (e.g., `client-001`)

**Example:**
```
GET /api/clients/client-001
```

**Response:**
```json
{
  "id": "client-001",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "phone": "555-0101",
  "date_of_birth": "1985-03-15",
  "address": "123 Main St, Anytown, ST 12345",
  "emergency_contact_name": "Jane Doe",
  "emergency_contact_phone": "555-0102",
  "status": "active",
  "created_at": "2025-11-11T18:00:00Z",
  "updated_at": "2025-11-11T18:00:00Z"
}
```

**Error Response (404):**
```
Client not found: client-999
```

### Get Active Clients

**GET** `/api/clients/active`

Retrieves all clients with status "active".

**Response:**
```json
[
  {
    "id": "client-001",
    "first_name": "John",
    "last_name": "Doe",
    "status": "active",
    ...
  }
]
```

### Get Inactive Clients

**GET** `/api/clients/inactive`

Retrieves all clients with status "inactive".

**Response:**
```json
[
  {
    "id": "client-004",
    "first_name": "Emily",
    "last_name": "Williams",
    "status": "inactive",
    ...
  }
]
```

### Error Responses

All endpoints return appropriate HTTP status codes:

- `200 OK` - Success
- `400 Bad Request` - Invalid request (e.g., missing ID)
- `404 Not Found` - Resource not found
- `405 Method Not Allowed` - Wrong HTTP method
- `500 Internal Server Error` - Server error

Error responses include error messages in the response body.

## Testing the API

### Using cURL

#### Health Check
```bash
curl -X GET http://localhost:8080/health
```

#### Get All Clients
```bash
curl -X GET http://localhost:8080/api/clients
```

#### Get Active Clients
```bash
curl -X GET http://localhost:8080/api/clients/active
```

#### Get Inactive Clients
```bash
curl -X GET http://localhost:8080/api/clients/inactive
```

#### Get Client by ID

**Example: Get client-001**
```bash
curl -X GET http://localhost:8080/api/clients/client-001
```

**Example: Get client-002**
```bash
curl -X GET http://localhost:8080/api/clients/client-002
```

**Example: Get client-003**
```bash
curl -X GET http://localhost:8080/api/clients/client-003
```

**Example: Get client-004 (inactive)**
```bash
curl -X GET http://localhost:8080/api/clients/client-004
```

**Example: Get client-005**
```bash
curl -X GET http://localhost:8080/api/clients/client-005
```

#### Pretty Print with jq (optional)

If you have `jq` installed, you can pipe the output for better formatting:

```bash
curl -s -X GET http://localhost:8080/api/clients | jq '.'
```

### Using the Test Script

#### Local Testing

Run the test script:
```bash
chmod +x test-api.sh
./test-api.sh
```

Or set a custom base URL:
```bash
BASE_URL=http://localhost:8080 ./test-api.sh
```

#### API Gateway Testing

If you have deployed API Gateway, test against it:

```bash
# Get the API Gateway URL
API_URL=$(make get-api-url)

# Test endpoints
curl -X GET $API_URL/health
curl -X GET $API_URL/api/clients
curl -X GET $API_URL/api/clients/active
curl -X GET $API_URL/api/clients/client-001
```

Or use the test script with API Gateway URL:
```bash
./test-api.sh https://your-api-id.execute-api.us-east-1.amazonaws.com/prod
```

### Testing with Postman

1. **Start the server:**
   ```bash
   make run-server
   ```

2. **Import these requests into Postman:**

   - **Health Check:**
     - Method: `GET`
     - URL: `http://localhost:8080/health`

   - **Get All Clients:**
     - Method: `GET`
     - URL: `http://localhost:8080/api/clients`

   - **Get Client by ID:**
     - Method: `GET`
     - URL: `http://localhost:8080/api/clients/client-001`

   - **Get Active Clients:**
     - Method: `GET`
     - URL: `http://localhost:8080/api/clients/active`

   - **Get Inactive Clients:**
     - Method: `GET`
     - URL: `http://localhost:8080/api/clients/inactive`

3. **All responses are in JSON format** and can be viewed in Postman's response body.

## API Gateway

The project includes AWS API Gateway integration for production deployment. API Gateway provides a managed API endpoint that routes requests to your backend server.

### Deploy API Gateway

To deploy API Gateway to AWS:

```bash
# Set your backend server URL
export BACKEND_URL="http://your-alb-or-ec2-url:8080"
export AWS_REGION="us-east-1"

# Deploy
make deploy-api-gateway

# Get the API Gateway URL
make get-api-url

# Test the API Gateway
make test-api-gateway
```

**Note:** Make sure your backend server is running and accessible at the `BACKEND_URL` before deploying API Gateway.

### API Gateway Features

- **HTTP API** - Modern, low-latency API Gateway
- **CORS Support** - Configured for cross-origin requests
- **CloudWatch Logging** - Automatic request logging
- **Throttling** - Rate limiting (100 burst, 50 sustained)
- **All Routes** - Health check and all client endpoints configured

### Local vs API Gateway

- **Local Development:** Test directly against `http://localhost:8080` (no API Gateway needed)
- **AWS Production:** Use API Gateway URL for managed API access

## CI/CD with GitHub Actions

The project includes GitHub Actions workflows for continuous integration and deployment.

### Continuous Integration (CI)

The CI workflow (`.github/workflows/ci.yml`) runs on every push and pull request to `main` and `develop` branches:

- Runs tests
- Builds all binaries
- Checks code formatting
- Runs `go vet` for static analysis
- Uploads build artifacts

### Continuous Deployment (CD)

The deployment workflow (`.github/workflows/deploy.yml`) runs on:
- Pushes to `main` branch
- Version tags (e.g., `v1.0.0`)
- Manual workflow dispatch

#### Setting Up GitHub Secrets

To enable deployment, configure the following GitHub Secrets in your repository:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add the following secrets:

   - `AWS_ACCESS_KEY_ID`: AWS access key for deployment
   - `AWS_SECRET_ACCESS_KEY`: AWS secret key for deployment
   - `AWS_REGION`: AWS region (e.g., `us-east-1`)
   - `BACKEND_URL`: Backend server URL for API Gateway (optional)

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
   - Deploys API Gateway (if BACKEND_URL is set)
   - Uploads deployment artifacts

3. **Post-Deployment:**
   - Binary available in GitHub Actions artifacts
   - API Gateway URL displayed in deployment summary
   - Configure your infrastructure to use the binary
   - Set up service management (systemd, supervisor, etc.)

#### Optional Deployment Targets

The deployment workflow includes placeholders for:
- **EC2 Deployment**: Uncomment and configure EC2 deployment steps
- **Docker/ECS Deployment**: Uncomment and configure Docker registry
- **Lambda Deployment**: Uncomment and configure Lambda deployment

#### Production Environment

Create a GitHub environment called `production`:
1. Go to **Settings** → **Environments**
2. Create a new environment named `production`
3. Add environment-specific secrets if needed

## AWS Deployment

For AWS deployment, configure the DynamoDB endpoint to point to your AWS DynamoDB tables. Update the `DYNAMODB_ENDPOINT` environment variable (or remove it to use the default AWS endpoint) and set the appropriate `AWS_REGION` in your deployment configuration.

### EC2 + Load Balancer Deployment (Automated)

This is the recommended approach for production deployments. The infrastructure is automatically deployed via GitHub Actions when you push a version tag.

#### Quick Start

1. **Add GitHub Secrets:**

   Go to: **GitHub Repo Settings** → **Secrets and variables** → **Actions**

   | Secret | Value |
   |--------|-------|
   | `AWS_ACCESS_KEY_ID` | Your AWS access key ID |
   | `AWS_SECRET_ACCESS_KEY` | Your AWS secret access key |
   | `AWS_REGION` | `us-east-1` (or your preferred region) |
   | `EC2_INSTANCE_TYPE` | `t3.micro` (free tier) or `t3.small` |
   | `DEPLOY_EC2` | `true` |

2. **Deploy:**

   Create and push a version tag:
   ```bash
   git tag v0.2.8
   git push origin v0.2.8
   ```

   Or trigger manually in GitHub Actions.

3. **Get Your Backend URL:**

   After deployment completes, check the GitHub Actions job summary for:
   ```
   Load Balancer DNS: johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com
   Backend URL: http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com
   ```

4. **Test It:**

   ```bash
   # Health check
   curl http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com/health

   # Get clients
   curl http://johns-ai-backend-ec2-alb-123456.us-east-1.elb.amazonaws.com/api/clients
   ```

#### Infrastructure Created

✅ **Default VPC with 2 public subnets** (across 2 availability zones)  
✅ **Internet Gateway** for outbound traffic  
✅ **Security Groups** (ALB and EC2)  
✅ **Application Load Balancer** (port 80)  
✅ **Target Group** with health checks (`/health`)  
✅ **EC2 Instance** running your backend (port 8080)  
✅ **IAM Role** with DynamoDB permissions  
✅ **CloudWatch Logs** for monitoring  

#### Security Features

- ✅ IAM roles (no hardcoded credentials on EC2)
- ✅ Security groups restrict traffic
- ✅ Health checks ensure instance is healthy
- ✅ Auto-deregistration if instance fails
- ✅ No SSH access exposed (uses Systems Manager Session Manager)
- ⚠️ **HTTP only** (add HTTPS for production)

#### Cost Estimate

- EC2 t3.micro: ~$7.59/month (eligible for free tier)
- Load Balancer: ~$16.20/month
- Data transfer: ~$0.02/GB
- **Total: ~$24/month**

#### Connect to API Gateway

Once the EC2 stack is deployed:

1. Get the ALB DNS from deployment output
2. Add GitHub secret: `BACKEND_URL=http://<ALB-DNS>`
3. Push a new tag to deploy API Gateway pointing to it

#### Debugging EC2 Deployment

**Check instance status:**
```bash
aws ec2 describe-instances --region us-east-1 \
  --query 'Reservations[].Instances[?Tags[?Key==`Name`&&Value==`*backend*`]]'
```

**Access instance via AWS Systems Manager:**
```bash
# List sessions
aws ssm describe-sessions --region us-east-1

# Start a session
aws ssm start-session --target i-xxxxx --region us-east-1
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

**Instance Type Options:**
- `t3.micro` - Free tier (1GB RAM)
- `t3.small` - $18/month (2GB RAM)
- `t3.medium` - $35/month (4GB RAM)

Change by updating GitHub secret `EC2_INSTANCE_TYPE`.

**Cleanup:**

To delete all resources:
```bash
aws cloudformation delete-stack \
  --stack-name johns-ai-backend-ec2 --region us-east-1
```

### Manual AWS Deployment

1. **Set up AWS credentials:**
   ```bash
   export AWS_ACCESS_KEY_ID=your-access-key
   export AWS_SECRET_ACCESS_KEY=your-secret-key
   export AWS_REGION=us-east-1
   ```

2. **Create DynamoDB tables:**
   ```bash
   # Remove or unset DYNAMODB_ENDPOINT to use AWS
   unset DYNAMODB_ENDPOINT
   make setup-db
   ```

3. **Build and deploy the server:**
   ```bash
   make build-server
   # Deploy bin/server to your infrastructure
   ```

4. **Deploy API Gateway:**
   ```bash
   export BACKEND_URL="http://your-backend-server:8080"
   make deploy-api-gateway
   ```

### Deployment Targets

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

**API Gateway deployment fails:**
- Verify `BACKEND_URL` is accessible
- Check backend server is running
- Ensure IAM permissions for API Gateway

## Best Practices

1. **Never commit `.env` files** - They contain sensitive information
2. **Use `.env.example`** as a template for required variables
3. **Use GitHub Secrets** for production credentials
4. **Test locally** before pushing to main branch
5. **Use version tags** for production releases
6. **Monitor deployment logs** in GitHub Actions
7. **Use API Gateway** for production API access
8. **Monitor CloudWatch logs** for API Gateway requests

## Authentication

The API uses JWT-based authentication for all client endpoints. See the following documentation:

- **[Authentication Guide](docs/AUTHENTICATION.md)** - Complete authentication system documentation
- **[JWT Secret Setup](docs/JWT_SECRET_SETUP.md)** - How to configure JWT_SECRET in AWS
- **[Postman Setup](docs/POSTMAN_SETUP.md)** - Testing with Postman

### Quick Start

For local development:
```bash
export JWT_SECRET="dev-secret-key-change-in-production"
make run-server
```

For production deployment:
```bash
./scripts/setup-jwt-secret.sh
```

## Additional Resources

- [AWS DynamoDB Documentation](https://docs.aws.amazon.com/dynamodb/)
- [AWS API Gateway Documentation](https://docs.aws.amazon.com/apigateway/)
- [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Environment Variables](https://golang.org/pkg/os/#Getenv)
