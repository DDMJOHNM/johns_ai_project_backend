package db

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Client wraps the DynamoDB client and provides connection management
type Client struct {
	DynamoDB *dynamodb.Client
	Region   string
	Endpoint string
}

// NewClient creates a new DynamoDB client connection
// It reads configuration from environment variables:
//   - DYNAMODB_ENDPOINT: DynamoDB endpoint (default: http://localhost:8000 for local)
//   - AWS_REGION: AWS region (default: us-east-1)
//   - AWS_ACCESS_KEY_ID: AWS access key (default: test for local)
//   - AWS_SECRET_ACCESS_KEY: AWS secret key (default: test for local)
func NewClient(ctx context.Context) (*Client, error) {
	endpoint := getEnv("DYNAMODB_ENDPOINT", "http://localhost:8000")
	region := getEnv("AWS_REGION", "us-east-1")
	accessKey := getEnv("AWS_ACCESS_KEY_ID", "test")
	secretKey := getEnv("AWS_SECRET_ACCESS_KEY", "test")

	// Load AWS config with custom endpoint for local DynamoDB
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
				if service == dynamodb.ServiceID {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: reg,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			},
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	return &Client{
		DynamoDB: client,
		Region:   region,
		Endpoint: endpoint,
	}, nil
}

// Ping checks if the DynamoDB connection is working by listing tables
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.DynamoDB.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return fmt.Errorf("failed to ping DynamoDB: %w", err)
	}
	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

