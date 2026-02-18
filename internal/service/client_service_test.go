package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmason/john_ai_project/internal/repository"
)

// Mock ClientRepository
type MockClientRepository struct {
	CreateClientFunc       func(ctx context.Context, client *repository.Client) error
	GetClientListFunc      func(ctx context.Context) ([]repository.Client, error)
	GetClientByIDFunc      func(ctx context.Context, id string) (*repository.Client, error)
	GetClientsByStatusFunc func(ctx context.Context, status string) ([]repository.Client, error)
}

func (m *MockClientRepository) CreateClient(ctx context.Context, client *repository.Client) error {
	if m.CreateClientFunc != nil {
		return m.CreateClientFunc(ctx, client)
	}
	return nil
}

func (m *MockClientRepository) GetClientList(ctx context.Context) ([]repository.Client, error) {
	if m.GetClientListFunc != nil {
		return m.GetClientListFunc(ctx)
	}
	return nil, nil
}

func (m *MockClientRepository) GetClientByID(ctx context.Context, id string) (*repository.Client, error) {
	if m.GetClientByIDFunc != nil {
		return m.GetClientByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockClientRepository) GetClientsByStatus(ctx context.Context, status string) ([]repository.Client, error) {
	if m.GetClientsByStatusFunc != nil {
		return m.GetClientsByStatusFunc(ctx, status)
	}
	return nil, nil
}

func TestClientService_CreateClient(t *testing.T) {
	tests := []struct {
		name          string
		client        *repository.Client
		mockSetup     func(*MockClientRepository)
		expectedError string
	}{
		{
			name: "Success - Create client",
			client: &repository.Client{
				ID:        "client-123",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Status:    "active",
			},
			mockSetup: func(m *MockClientRepository) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name: "Failure - Repository error",
			client: &repository.Client{
				ID:        "client-456",
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane@example.com",
			},
			mockSetup: func(m *MockClientRepository) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return errors.New("dynamodb error")
				}
			},
			expectedError: "failed to create client",
		},
		{
			name: "Success - Client with all fields",
			client: &repository.Client{
				ID:                    "client-789",
				FirstName:             "Full",
				LastName:              "Fields",
				Email:                 "full@example.com",
				Phone:                 "+1234567890",
				DateOfBirth:           "1990-01-01",
				Address:               "123 Street",
				EmergencyContactName:  "Emergency Person",
				EmergencyContactPhone: "+0987654321",
				Status:                "active",
				CreatedAt:             "2024-01-01T00:00:00Z",
				UpdatedAt:             "2024-01-01T00:00:00Z",
			},
			mockSetup: func(m *MockClientRepository) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					// Verify all fields are passed through
					if diff := cmp.Diff("Emergency Person", client.EmergencyContactName); diff != "" {
						t.Errorf("EmergencyContactName mismatch (-want +got):\n%s", diff)
					}
					return nil
				}
			},
			expectedError: "",
		},
		{
			name: "Success - Client with minimal fields",
			client: &repository.Client{
				ID:        "client-min",
				FirstName: "Min",
				LastName:  "Test",
				Email:     "min@test.com",
			},
			mockSetup: func(m *MockClientRepository) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					// Verify phone field is empty
					if diff := cmp.Diff("", client.Phone); diff != "" {
						t.Errorf("Phone mismatch (-want +got):\n%s", diff)
					}
					return nil
				}
			},
			expectedError: "",
		},
		{
			name: "Failure - Network error",
			client: &repository.Client{
				ID:        "client-net",
				FirstName: "Network",
				LastName:  "Error",
				Email:     "network@error.com",
			},
			mockSetup: func(m *MockClientRepository) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return errors.New("network timeout")
				}
			},
			expectedError: "failed to create client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockClientRepository{}
			tt.mockSetup(mockRepo)

			service := NewClientService(mockRepo)
			ctx := context.Background()

			err := service.CreateClient(ctx, tt.client)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

