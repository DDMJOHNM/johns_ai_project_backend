package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Client represents a client in the database
type Client struct {
	ID                      string `dynamodbav:"id" json:"id"`
	FirstName               string `dynamodbav:"first_name" json:"first_name"`
	LastName                string `dynamodbav:"last_name" json:"last_name"`
	Email                   string `dynamodbav:"email" json:"email"`
	Phone                   string `dynamodbav:"phone" json:"phone"`
	DateOfBirth             string `dynamodbav:"date_of_birth" json:"date_of_birth"`
	Address                 string `dynamodbav:"address" json:"address"`
	EmergencyContactName    string `dynamodbav:"emergency_contact_name" json:"emergency_contact_name"`
	EmergencyContactPhone   string `dynamodbav:"emergency_contact_phone" json:"emergency_contact_phone"`
	Status                  string `dynamodbav:"status" json:"status"`
	CreatedAt               string `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt               string `dynamodbav:"updated_at" json:"updated_at"`
}

// ClientRepository handles database operations for clients
type ClientRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewClientRepository creates a new client repository
func NewClientRepository(client *dynamodb.Client) *ClientRepository {
	return &ClientRepository{
		client:    client,
		tableName: "clients",
	}
}

// GetClientList retrieves all clients from the database
func (r *ClientRepository) GetClientList(ctx context.Context) ([]Client, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan clients table: %w", err)
	}

	var clients []Client
	for _, item := range result.Items {
		var client Client
		if err := attributevalue.UnmarshalMap(item, &client); err != nil {
			return nil, fmt.Errorf("failed to unmarshal client: %w", err)
		}
		clients = append(clients, client)
	}

	return clients, nil
}

// GetClientByID retrieves a single client by ID
func (r *ClientRepository) GetClientByID(ctx context.Context, id string) (*Client, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("client not found: %s", id)
	}

	var client Client
	if err := attributevalue.UnmarshalMap(result.Item, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client: %w", err)
	}

	return &client, nil
}

// GetClientsByStatus retrieves clients filtered by status using the status-index GSI
func (r *ClientRepository) GetClientsByStatus(ctx context.Context, status string) ([]Client, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("status-index"),
		KeyConditionExpression: aws.String("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: status},
		},
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query clients by status: %w", err)
	}

	var clients []Client
	for _, item := range result.Items {
		var client Client
		if err := attributevalue.UnmarshalMap(item, &client); err != nil {
			return nil, fmt.Errorf("failed to unmarshal client: %w", err)
		}
		clients = append(clients, client)
	}

	return clients, nil
}

