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
//   - DYNAMODB_ENDPOINT: DynamoDB endpoint (if empty, uses AWS DynamoDB)
//   - AWS_REGION: AWS region (default: us-east-1)
//   - AWS_ACCESS_KEY_ID: AWS access key (optional, uses IAM role if not set)
//   - AWS_SECRET_ACCESS_KEY: AWS secret key (optional, uses IAM role if not set)
func NewClient(ctx context.Context) (*Client, error) {
	endpoint := getEnv("DYNAMODB_ENDPOINT", "")
	region := getEnv("AWS_REGION", "us-east-1")
	accessKey := getEnv("AWS_ACCESS_KEY_ID", "")
	secretKey := getEnv("AWS_SECRET_ACCESS_KEY", "")

	// Build config options
	cfgOpts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	// Only use static credentials if both access key and secret are provided (for local dev)
	if accessKey != "" && secretKey != "" {
		cfgOpts = append(cfgOpts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	}
	// Otherwise, LoadDefaultConfig will automatically use IAM role credentials on EC2

	// Add custom endpoint resolver for local DynamoDB
	cfgOpts = append(cfgOpts, config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
		func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID && endpoint != "" {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: reg,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		},
	)))

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
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

