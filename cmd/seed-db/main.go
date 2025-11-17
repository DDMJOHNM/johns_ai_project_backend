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
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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

	log.Println("Connected successfully")

	// Seed clients
	if err := seedClients(ctx, client); err != nil {
		log.Fatalf("Failed to seed clients: %v", err)
	}

	// Seed users
	// if err := seedUsers(ctx, client); err != nil {
	// 	log.Fatalf("Failed to seed users: %v", err)
	// }

	log.Println("Database seeding completed successfully!")
}

func seedClients(ctx context.Context, client *dynamodb.Client) error {
	log.Println("Seeding clients table...")

	clients := []map[string]interface{}{
		{
			"id":                      "client-001",
			"first_name":              "John",
			"last_name":               "Doe",
			"email":                   "john.doe@example.com",
			"phone":                   "555-0101",
			"date_of_birth":           "1985-03-15",
			"address":                 "123 Main St, Anytown, ST 12345",
			"emergency_contact_name":  "Jane Doe",
			"emergency_contact_phone": "555-0102",
			"status":                  "active",
			"created_at":              time.Now().Format(time.RFC3339),
			"updated_at":              time.Now().Format(time.RFC3339),
		},
		{
			"id":                      "client-002",
			"first_name":              "Sarah",
			"last_name":               "Smith",
			"email":                   "sarah.smith@example.com",
			"phone":                   "555-0201",
			"date_of_birth":           "1990-07-22",
			"address":                 "456 Oak Ave, Somewhere, ST 67890",
			"emergency_contact_name":  "Bob Smith",
			"emergency_contact_phone": "555-0202",
			"status":                  "active",
			"created_at":              time.Now().Format(time.RFC3339),
			"updated_at":              time.Now().Format(time.RFC3339),
		},
		{
			"id":                      "client-003",
			"first_name":              "Michael",
			"last_name":               "Johnson",
			"email":                   "michael.johnson@example.com",
			"phone":                   "555-0301",
			"date_of_birth":           "1988-11-08",
			"address":                 "789 Pine Rd, Elsewhere, ST 11111",
			"emergency_contact_name":  "Mary Johnson",
			"emergency_contact_phone": "555-0302",
			"status":                  "active",
			"created_at":              time.Now().Format(time.RFC3339),
			"updated_at":              time.Now().Format(time.RFC3339),
		},
		{
			"id":                      "client-004",
			"first_name":              "Emily",
			"last_name":               "Williams",
			"email":                   "emily.williams@example.com",
			"phone":                   "555-0401",
			"date_of_birth":           "1992-05-30",
			"address":                 "321 Elm St, Nowhere, ST 22222",
			"emergency_contact_name":  "David Williams",
			"emergency_contact_phone": "555-0402",
			"status":                  "inactive",
			"created_at":              time.Now().Format(time.RFC3339),
			"updated_at":              time.Now().Format(time.RFC3339),
		},
		{
			"id":                      "client-005",
			"first_name":              "James",
			"last_name":               "Brown",
			"email":                   "james.brown@example.com",
			"phone":                   "555-0501",
			"date_of_birth":           "1987-09-14",
			"address":                 "654 Maple Dr, Anywhere, ST 33333",
			"emergency_contact_name":  "Lisa Brown",
			"emergency_contact_phone": "555-0502",
			"status":                  "active",
			"created_at":              time.Now().Format(time.RFC3339),
			"updated_at":              time.Now().Format(time.RFC3339),
		},
	}

	for _, clientData := range clients {
		item, err := attributevalue.MarshalMap(clientData)
		if err != nil {
			return fmt.Errorf("failed to marshal client data: %w", err)
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("clients"),
			Item:      item,
		})
		if err != nil {
			return fmt.Errorf("failed to put client item: %w", err)
		}

		log.Printf("  ✓ Seeded client: %s %s (%s)", clientData["first_name"], clientData["last_name"], clientData["email"])
	}

	log.Printf("✓ Seeded %d clients", len(clients))
	return nil
}

func seedUsers(ctx context.Context, client *dynamodb.Client) error {
	log.Println("Seeding users table...")

	// Note: In production, passwords should be properly hashed using bcrypt or similar
	// For test data, we're using simple placeholders
	users := []map[string]interface{}{
		{
			"id":            "user-001",
			"username":      "admin",
			"email":         "admin@example.com",
			"password_hash": "$2a$10$dummyhashforadmin", // Placeholder - use bcrypt in production
			"first_name":    "Admin",
			"last_name":     "User",
			"role":          "admin",
			"status":        "active",
			"created_at":    time.Now().Format(time.RFC3339),
			"updated_at":    time.Now().Format(time.RFC3339),
		},
		{
			"id":            "user-002",
			"username":      "counsellor1",
			"email":         "counsellor1@example.com",
			"password_hash": "$2a$10$dummyhashforcounsellor1", // Placeholder
			"first_name":    "Alice",
			"last_name":     "Counsellor",
			"role":          "counsellor",
			"status":        "active",
			"created_at":    time.Now().Format(time.RFC3339),
			"updated_at":    time.Now().Format(time.RFC3339),
		},
		{
			"id":            "user-003",
			"username":      "counsellor2",
			"email":         "counsellor2@example.com",
			"password_hash": "$2a$10$dummyhashforcounsellor2", // Placeholder
			"first_name":    "Bob",
			"last_name":     "Therapist",
			"role":          "counsellor",
			"status":        "active",
			"created_at":    time.Now().Format(time.RFC3339),
			"updated_at":    time.Now().Format(time.RFC3339),
		},
		{
			"id":            "user-004",
			"username":      "staff1",
			"email":         "staff1@example.com",
			"password_hash": "$2a$10$dummyhashforstaff1", // Placeholder
			"first_name":    "Charlie",
			"last_name":     "Staff",
			"role":          "staff",
			"status":        "active",
			"created_at":    time.Now().Format(time.RFC3339),
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	for _, userData := range users {
		item, err := attributevalue.MarshalMap(userData)
		if err != nil {
			return fmt.Errorf("failed to marshal user data: %w", err)
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("users"),
			Item:      item,
		})
		if err != nil {
			return fmt.Errorf("failed to put user item: %w", err)
		}

		log.Printf("  ✓ Seeded user: %s (%s) - %s", userData["username"], userData["email"], userData["role"])
	}

	log.Printf("✓ Seeded %d users", len(users))
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
