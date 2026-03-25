package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Note is one entry in a client's notes list.
type Note struct {
	Date     string `dynamodbav:"date" json:"date"`
	ClientID string `dynamodbav:"client_id" json:"client_id"`
	Note     string `dynamodbav:"note" json:"note"`
}

type Client struct {
	ID                    string `dynamodbav:"id" json:"id"`
	FirstName             string `dynamodbav:"first_name" json:"first_name"`
	LastName              string `dynamodbav:"last_name" json:"last_name"`
	Email                 string `dynamodbav:"email" json:"email"`
	Phone                 string `dynamodbav:"phone" json:"phone"`
	DateOfBirth           string `dynamodbav:"date_of_birth" json:"date_of_birth"`
	Address               string `dynamodbav:"address" json:"address"`
	EmergencyContactName  string `dynamodbav:"emergency_contact_name" json:"emergency_contact_name"`
	EmergencyContactPhone string `dynamodbav:"emergency_contact_phone" json:"emergency_contact_phone"`
	Status                string `dynamodbav:"status" json:"status"`
	RequestedCounsellor   string `dynamodbav:"requested_counsellor" json:"requested_counsellor"`
	Urgency               string `dynamodbav:"urgency" json:"urgency"`
	NextAppointment       string `dynamodbav:"next_appointment" json:"next_appointment"`
	Notes                 []Note `dynamodbav:"notes" json:"notes"`
	CreatedAt             string `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt             string `dynamodbav:"updated_at" json:"updated_at"`
}

// MarshalJSON adds display helpers: name (first + last), initial_consult_notes (first note body).
func (c Client) MarshalJSON() ([]byte, error) {
	type Alias Client
	fn := strings.TrimSpace(c.FirstName)
	ln := strings.TrimSpace(c.LastName)
	name := strings.TrimSpace(fn + " " + ln)
	initialConsult := ""
	if len(c.Notes) > 0 {
		initialConsult = c.Notes[0].Note
	}
	return json.Marshal(&struct {
		Alias
		Name                string `json:"name"`
		InitialConsultNotes string `json:"initial_consult_notes"`
	}{
		Alias:               Alias(c),
		Name:                name,
		InitialConsultNotes: initialConsult,
	})
}

// ClientRepository handles database operations for clients
type ClientRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewClientRepository(client *dynamodb.Client) *ClientRepository {
	return &ClientRepository{
		client:    client,
		tableName: "clients",
	}
}

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

// TODO: parameter validation for the id and return an error if the id is not a valid uuid
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

	// US spelling sometimes used in tools / manual edits; map onto canonical field for JSON/API.
	if client.RequestedCounsellor == "" {
		if alt := stringAttrS(result.Item, "requested_counselor"); alt != "" {
			client.RequestedCounsellor = alt
		}
	}

	return &client, nil
}

func stringAttrS(item map[string]types.AttributeValue, key string) string {
	v, ok := item[key]
	if !ok {
		return ""
	}
	s, ok := v.(*types.AttributeValueMemberS)
	if !ok {
		return ""
	}
	return s.Value
}

// idFromItem reads the table hash key from a DynamoDB item (string or number).
func idFromItem(item map[string]types.AttributeValue) string {
	v, ok := item["id"]
	if !ok {
		return ""
	}
	switch t := v.(type) {
	case *types.AttributeValueMemberS:
		return strings.TrimSpace(t.Value)
	case *types.AttributeValueMemberN:
		return strings.TrimSpace(t.Value)
	default:
		return ""
	}
}

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

func (r *ClientRepository) CreateClient(ctx context.Context, client *Client) error {
	item, err := attributevalue.MarshalMap(client)
	if err != nil {
		return fmt.Errorf("failed to marshal client: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

func (r *ClientRepository) GetClientByEmail(ctx context.Context, email string) (*Client, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("#email = :email"),
		ExpressionAttributeNames: map[string]string{
			"#email": "email",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
		Limit: aws.Int32(1),
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query client by email: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("client not found for email: %s", email)
	}

	item := result.Items[0]
	var stub Client
	if err := attributevalue.UnmarshalMap(item, &stub); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client: %w", err)
	}

	id := strings.TrimSpace(stub.ID)
	if id == "" {
		id = idFromItem(item)
	}
	if id == "" {
		return nil, fmt.Errorf("email index row missing id attribute for email %s", email)
	}

	// Load the full item by primary key. Email-index queries may omit attributes if the index
	// projection is not ALL, so a follow-up GetItem guarantees requested_counsellor, urgency, etc.
	return r.GetClientByID(ctx, id)
}

// ClientPatch lists optional fields to write. At least one field must be non-nil.
type ClientPatch struct {
	FirstName           *string
	LastName            *string
	Email               *string
	Notes               *[]Note
	RequestedCounsellor *string
	Urgency             *string
	NextAppointment     *string
}

func (r *ClientRepository) UpdateClient(ctx context.Context, id string, patch ClientPatch) error {
	if patch.FirstName == nil && patch.LastName == nil && patch.Email == nil && patch.Notes == nil &&
		patch.RequestedCounsellor == nil && patch.Urgency == nil && patch.NextAppointment == nil {
		return fmt.Errorf("no fields to update")
	}

	updatedAt := time.Now().Format(time.RFC3339)
	parts := []string{"updated_at = :ua"}
	values := map[string]types.AttributeValue{
		":ua": &types.AttributeValueMemberS{Value: updatedAt},
	}

	if patch.FirstName != nil {
		parts = append(parts, "first_name = :fn")
		values[":fn"] = &types.AttributeValueMemberS{Value: *patch.FirstName}
	}
	if patch.LastName != nil {
		parts = append(parts, "last_name = :ln")
		values[":ln"] = &types.AttributeValueMemberS{Value: *patch.LastName}
	}
	if patch.Email != nil {
		parts = append(parts, "email = :em")
		values[":em"] = &types.AttributeValueMemberS{Value: *patch.Email}
	}
	if patch.Notes != nil {
		notes := *patch.Notes
		if notes == nil {
			notes = []Note{}
		}
		notesAV, err := attributevalue.Marshal(notes)
		if err != nil {
			return fmt.Errorf("failed to marshal notes: %w", err)
		}
		parts = append(parts, "notes = :notes")
		values[":notes"] = notesAV
	}
	if patch.RequestedCounsellor != nil {
		parts = append(parts, "requested_counsellor = :rc")
		values[":rc"] = &types.AttributeValueMemberS{Value: *patch.RequestedCounsellor}
	}
	if patch.Urgency != nil {
		parts = append(parts, "urgency = :ur")
		values[":ur"] = &types.AttributeValueMemberS{Value: *patch.Urgency}
	}
	if patch.NextAppointment != nil {
		parts = append(parts, "next_appointment = :na")
		values[":na"] = &types.AttributeValueMemberS{Value: *patch.NextAppointment}
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ConditionExpression: aws.String("attribute_exists(#pk)"),
		UpdateExpression:    aws.String("SET " + strings.Join(parts, ", ")),
		ExpressionAttributeNames: map[string]string{
			"#pk": "id",
		},
		ExpressionAttributeValues: values,
	}

	_, err := r.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	return nil
}
