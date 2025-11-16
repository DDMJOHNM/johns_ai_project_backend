# Mental Health Counselling Client Management System

A system to manage clients receiving counselling for mental health.

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Make

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

## Available Commands

### Docker Commands
- `make docker-up` - Start DynamoDB Local container
- `make docker-down` - Stop DynamoDB Local container
- `make docker-logs` - View DynamoDB container logs
- `make docker-status` - Check DynamoDB container status

### Database Commands
- `make setup-db` - Create DynamoDB tables
- `make seed-db` - Seed DynamoDB with test data
- `make test-db` - Run setup-db and seed-db
- `make verify` - Verify tables exist

### Server Commands
- `make run-server` - Start the API server (default port 8080)
- `make build-server` - Build the API server binary

### Other Commands
- `make setup` - Full setup (docker + database)
- `make build` - Build all binaries
- `make clean` - Clean build artifacts

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

## Local Development

The system uses Docker Compose to run DynamoDB Local. The database is accessible at `http://localhost:8000`.

## API Server

The API server provides REST endpoints to interact with the client data.

### Starting the Server

```bash
make run-server
```

The server will start on `http://localhost:8080` (or the port specified in `HTTP_PORT` environment variable).

### Available Endpoints

- `GET /health` - Health check endpoint
- `GET /api/clients` - Get all clients
- `GET /api/clients/{id}` - Get a specific client by ID
- `GET /api/clients/active` - Get all active clients
- `GET /api/clients/inactive` - Get all inactive clients

See [API_ENDPOINTS.md](API_ENDPOINTS.md) for detailed API documentation and Postman examples.

### Testing with Postman

1. Start the server: `make run-server`
2. Import the endpoints from `API_ENDPOINTS.md` into Postman
3. All endpoints return JSON responses

## Environment Setup

### Local Development

1. **Copy the environment template:**
   ```bash
   cp .env.example .env
   ```

2. **Configure your local environment:**
   Edit `.env` with your local settings:
   ```bash
   DYNAMODB_ENDPOINT=http://localhost:8000
   AWS_REGION=us-east-1
   HTTP_PORT=8080
   ```

3. **Start the development environment:**
   ```bash
   make setup
   ```

### Production Environment

For production deployment, you'll need to configure the following:

1. **AWS Credentials:**
   - Set up AWS IAM user/role with DynamoDB permissions
   - Configure AWS credentials via environment variables or IAM roles

2. **Environment Variables:**
   - `DYNAMODB_ENDPOINT`: Leave empty or remove to use AWS DynamoDB endpoint
   - `AWS_REGION`: Your AWS region (e.g., `us-east-1`)
   - `HTTP_PORT`: Server port (default: `8080`)

3. **DynamoDB Tables:**
   - Create tables in AWS DynamoDB (or use the deployment script)
   - Ensure proper IAM permissions for table operations

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

