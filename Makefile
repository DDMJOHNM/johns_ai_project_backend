.PHONY: help setup-db seed-db docker-up docker-down docker-logs docker-status clean test-db setup verify build build-create-db build-seed-db build-example run-example build-server run-server deploy-api-gateway get-api-url delete-api-gateway test-api-gateway

# Variables with defaults (can be overridden by .env file or environment)
# The .env file is automatically loaded by docker-compose and Go programs
DYNAMODB_ENDPOINT ?= http://localhost:8000
AWS_REGION ?= us-east-1
BINARY_DIR = bin

# Helper to load .env file for shell commands
ENV_LOAD = if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi;

help:
	@echo "Mental Health Counselling Client Management System - Makefile"
	@echo ""
	@echo "Available commands:"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-up       - Start DynamoDB Local container"
	@echo "  make docker-down     - Stop DynamoDB Local container"
	@echo "  make docker-logs     - View DynamoDB container logs"
	@echo "  make docker-status   - Check DynamoDB container status"
	@echo ""
	@echo "Database Commands:"
	@echo "  make setup-db        - Create DynamoDB tables (clients and users)"
	@echo "  make seed-db         - Seed DynamoDB with test data"
	@echo "  make test-db         - Run setup-db and seed-db"
	@echo "  make verify          - Verify tables exist and have data"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build           - Build all Go binaries"
	@echo "  make build-create-db - Build create-db binary"
	@echo "  make build-seed-db   - Build seed-db binary"
	@echo "  make build-example   - Build example client service binary"
	@echo "  make run-example     - Run example client service"
	@echo "  make build-server    - Build API server binary"
	@echo "  make run-server      - Run API server (default port 8080)"
	@echo ""
	@echo "Setup & Cleanup:"
	@echo "  make setup           - Full setup (docker-up + test-db)"
	@echo "  make clean           - Clean build artifacts and binaries"
	@echo ""
	@echo "API Gateway Commands:"
	@echo "  make deploy-api-gateway - Deploy API Gateway (requires BACKEND_URL)"
	@echo "  make get-api-url        - Get API Gateway endpoint URL"
	@echo "  make test-api-gateway   - Test API Gateway endpoints"
	@echo "  make delete-api-gateway - Delete API Gateway stack"
	@echo ""
	@echo "Environment Variables:"
	@echo "  DYNAMODB_ENDPOINT    - DynamoDB endpoint (default: http://localhost:8000)"
	@echo "  AWS_REGION           - AWS region (default: us-east-1)"

# Docker commands
docker-up:
	@echo "Starting DynamoDB Local container..."
	@docker compose up -d
	@echo "Waiting for DynamoDB to be ready..."
	@sleep 5
	@echo "✓ DynamoDB Local container is running on port 8000"
	@echo "  Endpoint: $(DYNAMODB_ENDPOINT)"

docker-down:
	@echo "Stopping DynamoDB Local container..."
	@docker compose down
	@echo "✓ DynamoDB Local container stopped"

docker-logs:
	@docker compose logs -f dynamodb

docker-status:
	@echo "Checking DynamoDB container status..."
	@docker compose ps dynamodb || echo "Container is not running"

# Database setup
setup-db:
	@echo "Creating DynamoDB tables..."
	@if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi; \
	 DYNAMODB_ENDPOINT=$${DYNAMODB_ENDPOINT:-$(DYNAMODB_ENDPOINT)} \
	 AWS_REGION=$${AWS_REGION:-$(AWS_REGION)} \
	 go run cmd/create-db/main.go

seed-db:
	@echo "Seeding DynamoDB with test data..."
	@if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi; \
	 DYNAMODB_ENDPOINT=$${DYNAMODB_ENDPOINT:-$(DYNAMODB_ENDPOINT)} \
	 AWS_REGION=$${AWS_REGION:-$(AWS_REGION)} \
	 go run cmd/seed-db/main.go

test-db: setup-db seed-db
	@echo "✓ DynamoDB setup and seeding completed!"

verify:
	@echo "Verifying DynamoDB setup..."
	@echo "Checking if tables exist..."
	@aws dynamodb list-tables \
		--endpoint-url $(DYNAMODB_ENDPOINT) \
		--region $(AWS_REGION) \
		2>/dev/null || (echo "Error: Could not connect to DynamoDB. Is it running?" && exit 1)
	@echo "✓ Verification complete"

# Build commands
build: build-create-db build-seed-db build-example build-server
	@echo "✓ All binaries built successfully"

build-create-db:
	@echo "Building create-db binary..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/create-db ./cmd/create-db
	@echo "✓ Created $(BINARY_DIR)/create-db"

build-seed-db:
	@echo "Building seed-db binary..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/seed-db ./cmd/seed-db
	@echo "✓ Created $(BINARY_DIR)/seed-db"

build-example:
	@echo "Building example binary..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/example ./cmd/example
	@echo "✓ Created $(BINARY_DIR)/example"

run-example:
	@echo "Running example client service..."
	@if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi; \
	 DYNAMODB_ENDPOINT=$${DYNAMODB_ENDPOINT:-$(DYNAMODB_ENDPOINT)} \
	 AWS_REGION=$${AWS_REGION:-$(AWS_REGION)} \
	 go run ./cmd/example

build-server:
	@echo "Building server binary..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/server ./cmd/server
	@echo "✓ Created $(BINARY_DIR)/server"

run-server:
	@echo "Starting API server..."
	@echo "Server will be available at http://localhost:$${HTTP_PORT:-8080}"
	@echo "Press Ctrl+C to stop"
	@if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi; \
	 DYNAMODB_ENDPOINT=$${DYNAMODB_ENDPOINT:-$(DYNAMODB_ENDPOINT)} \
	 AWS_REGION=$${AWS_REGION:-$(AWS_REGION)} \
	 HTTP_PORT=$${HTTP_PORT:-8080} \
	 go run ./cmd/server

# Cleanup
clean:
	@echo "Cleaning build artifacts..."
	@go clean
	@rm -rf build/ $(BINARY_DIR)/
	@echo "✓ Cleanup complete"

# Full setup (docker + database)
setup: docker-up test-db
	@echo ""
	@echo "✓ Full setup completed!"
	@echo ""
	@echo "DynamoDB is ready at: $(DYNAMODB_ENDPOINT)"
	@echo "Tables created: clients, users"
	@echo ""
	@echo "Next steps:"
	@echo "  - Run 'make verify' to verify the setup"
	@echo "  - Check logs with 'make docker-logs'"

# API Gateway commands
deploy-api-gateway:
	@echo "Deploying API Gateway..."
	@if [ -z "$(BACKEND_URL)" ]; then \
		echo "Error: BACKEND_URL must be set (e.g., http://your-alb-url:8080)"; \
		echo "Usage: make deploy-api-gateway BACKEND_URL=http://your-backend-url:8080"; \
		exit 1; \
	fi
	@aws cloudformation deploy \
		--stack-name johns-ai-api-gateway \
		--template-file infra/api-gateway.yml \
		--parameter-overrides \
			BackendUrl=$(BACKEND_URL) \
			StageName=$${STAGE_NAME:-prod} \
			Environment=$${ENVIRONMENT:-production} \
		--capabilities CAPABILITY_IAM \
		--region $(AWS_REGION)

get-api-url:
	@aws cloudformation describe-stacks \
		--stack-name johns-ai-api-gateway \
		--query 'Stacks[0].Outputs[?OutputKey==`ApiEndpoint`].OutputValue' \
		--output text \
		--region $(AWS_REGION) 2>/dev/null || echo "Stack not found. Deploy with: make deploy-api-gateway BACKEND_URL=..."

delete-api-gateway:
	@echo "Deleting API Gateway stack..."
	@aws cloudformation delete-stack \
		--stack-name johns-ai-api-gateway \
		--region $(AWS_REGION)
	@echo "Stack deletion initiated. Check status with: aws cloudformation describe-stacks --stack-name johns-ai-api-gateway"

test-api-gateway:
	@echo "Testing API Gateway endpoints..."
	@API_URL=$$(make get-api-url 2>/dev/null | grep -v "Stack not found"); \
	if [ -z "$$API_URL" ] || [ "$$API_URL" = "Stack not found. Deploy with: make deploy-api-gateway BACKEND_URL=..." ]; then \
		echo "Error: API Gateway not deployed. Run: make deploy-api-gateway BACKEND_URL=..."; \
		exit 1; \
	fi; \
	echo "Testing against: $$API_URL"; \
	echo ""; \
	echo "1. Health Check:"; \
	curl -s $$API_URL/health | jq '.' || curl -s $$API_URL/health; \
	echo ""; \
	echo "2. Get All Clients:"; \
	curl -s $$API_URL/api/clients | jq '.' || curl -s $$API_URL/api/clients; \
	echo ""


