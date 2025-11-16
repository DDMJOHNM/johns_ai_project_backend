package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	// Get configuration from environment
	endpoint := getEnv("DYNAMODB_ENDPOINT", "")
	region := getEnv("AWS_REGION", "us-east-1")
	accessKey := getEnv("AWS_ACCESS_KEY_ID", "")
	secretKey := getEnv("AWS_SECRET_ACCESS_KEY", "")

	// Determine connection target
	if endpoint != "" {
		log.Printf("Connecting to DynamoDB at %s (region: %s)...", endpoint, region)
	} else {
		log.Printf("Connecting to AWS DynamoDB (region: %s)...", region)
		log.Printf("Using credentials: accessKey=%s, secretKey=%s", maskString(accessKey), maskString(secretKey))
	}

	// Load AWS config
	ctx := context.Background()
	
	configOpts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	
	// For production (real AWS), always use explicit credentials from environment
	// For local, use provided credentials or defaults
	if accessKey != "" && secretKey != "" {
		configOpts = append(configOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	} else if endpoint == "" {
		// Real AWS without credentials - this will fail, so error early
		log.Fatalf("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set for AWS DynamoDB access")
	}
	
	// Use custom endpoint if provided (for local DynamoDB)
	if endpoint != "" {
		configOpts = append(configOpts, config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == dynamodb.ServiceID {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: region,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			},
		)))
	}
	
	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Wait for DynamoDB to be ready
	if err := waitForDynamoDB(ctx, client); err != nil {
		log.Fatalf("DynamoDB is not ready: %v", err)
	}

	log.Println("Connected to DynamoDB successfully")

	// Create tables
	if err := createTables(ctx, client); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	log.Println("Database setup completed successfully!")
}

func waitForDynamoDB(ctx context.Context, client *dynamodb.Client) error {
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		_, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
		if err == nil {
			return nil
		}
		log.Printf("Waiting for DynamoDB to be ready (attempt %d/%d)...", i+1, maxRetries)
		time.Sleep(retryDelay)
	}
	return fmt.Errorf("DynamoDB did not become ready after %d attempts", maxRetries)
}

func createTables(ctx context.Context, client *dynamodb.Client) error {
	log.Println("Creating tables...")

	// Create clients table
	if err := createClientsTable(ctx, client); err != nil {
		return fmt.Errorf("failed to create clients table: %w", err)
	}

	// Create users table
	if err := createUsersTable(ctx, client); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}

func createClientsTable(ctx context.Context, client *dynamodb.Client) error {
	log.Println("Creating clients table...")

	tableName := "clients"
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("email"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("status"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("email-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("email"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
			{
				IndexName: aws.String("status-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("status"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := client.CreateTable(ctx, input)
	if err != nil {
		// Check if table already exists
		var resourceInUseException *types.ResourceInUseException
		if err != nil && err.Error() != "" {
			// Try to describe table to see if it exists
			_, describeErr := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if describeErr == nil {
				log.Printf("  ✓ Clients table already exists")
				return nil
			}
		}
		if resourceInUseException != nil {
			log.Printf("  ✓ Clients table already exists")
			return nil
		}
		return err
	}

	log.Println("  ✓ Created clients table")
	return nil
}

func createUsersTable(ctx context.Context, client *dynamodb.Client) error {
	log.Println("Creating users table...")

	tableName := "users"
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("username"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("email"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("role"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("username-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("username"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
			{
				IndexName: aws.String("email-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("email"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
			{
				IndexName: aws.String("role-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("role"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	_, err := client.CreateTable(ctx, input)
	if err != nil {
		// Check if table already exists
		_, describeErr := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if describeErr == nil {
			log.Printf("  ✓ Users table already exists")
			return nil
		}
		return err
	}

	log.Println("  ✓ Created users table")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}
