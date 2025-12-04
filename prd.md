### PRD for the development of a system to manage clients recieiving counselling for mental health 

Section 1 : Database Creation

1. ✅ Create a DynamoDB database creation script in Go using AWS SDK v2
2. ✅ This database must be an AWS DynamoDB database that can be accessed for local development via Docker Desktop (using amazon/dynamodb-local:latest)
3. ✅ The database must use the AWS SDK for Go v2 (github.com/aws/aws-sdk-go-v2) for DynamoDB operations
4. ✅ The database initial tables will be clients and users with appropriate Global Secondary Indexes (GSIs)
5. ✅ There must be a script in Go to seed the database with test data using DynamoDB PutItem operations
6. ✅ Include all relevant commands in a Makefile (docker-up, docker-down, setup-db, seed-db, test-db, setup)

Section 2 : DynamoDB Local Setup
1. ✅ Create a DynamoDB that can be run locally on Docker using amazon/dynamodb-local:latest
2. ✅ Configure Docker Compose to run DynamoDB Local on port 8000
3. ✅ Set up environment variables for DynamoDB endpoint (DYNAMODB_ENDPOINT) and AWS region (AWS_REGION)
4. ✅ Implement table creation with proper key schemas and Global Secondary Indexes for efficient querying

Section 3: Environment Configuration
1. ✅ Put all environment variables in a .env file
2. ✅ Create .env.example template file
3. ✅ Update Makefile to load variables from .env file
4. ✅ Update docker-compose.yaml to use .env file

Section 4: Application Layer
1. ✅ Create a connection script in Go (internal/db/connection.go)
2. ✅ Make a repository and a service that gets the client list (internal/repository/client_repository.go, internal/service/client_service.go)
3. ✅ Create a router in Go and create some routes that can be hit via Postman and return the data (internal/router/router.go, cmd/server/main.go)
  
Section 5: Deployment 
1. Set Up the Project environments local development amd deployment to production via github actions     
2. Add Amazon api gateway via a cloud formation stack to deploy via github actions  
3. Also add cloud watch logging for the endpoint 

//TODO:
Section 6: Set up tests in go 
Section 7: Create Client
Section 8: Static React Site deployable to bucket and aws gateway set up
Section 9: Set Up React Test Library and Jest Tests  
Section 10: Set Up Authentication Login and Register and Logout
Section 11: Setup Error logging 
Section 12: Security Audit

