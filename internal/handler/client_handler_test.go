package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jmason/john_ai_project/internal/repository"
)

// Mock ClientService
type MockClientService struct {
	CreateClientFunc       func(ctx context.Context, client *repository.Client) error
	GetClientListFunc      func(ctx context.Context) ([]repository.Client, error)
	GetClientByIDFunc      func(ctx context.Context, id string) (*repository.Client, error)
	GetActiveClientsFunc   func(ctx context.Context) ([]repository.Client, error)
	GetInactiveClientsFunc func(ctx context.Context) ([]repository.Client, error)
}

func (m *MockClientService) CreateClient(ctx context.Context, client *repository.Client) error {
	if m.CreateClientFunc != nil {
		return m.CreateClientFunc(ctx, client)
	}
	return nil
}

func (m *MockClientService) GetClientList(ctx context.Context) ([]repository.Client, error) {
	if m.GetClientListFunc != nil {
		return m.GetClientListFunc(ctx)
	}
	return nil, nil
}

func (m *MockClientService) GetClientByID(ctx context.Context, id string) (*repository.Client, error) {
	if m.GetClientByIDFunc != nil {
		return m.GetClientByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockClientService) GetActiveClients(ctx context.Context) ([]repository.Client, error) {
	if m.GetActiveClientsFunc != nil {
		return m.GetActiveClientsFunc(ctx)
	}
	return nil, nil
}

func (m *MockClientService) GetInactiveClients(ctx context.Context) ([]repository.Client, error) {
	if m.GetInactiveClientsFunc != nil {
		return m.GetInactiveClientsFunc(ctx)
	}
	return nil, nil
}

func TestClientHandler_CreateClient(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		mockSetup      func(*MockClientService)
		expectedStatus int
		expectedError  string
		validateBody   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "Success - Full client data",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName:             "John",
				LastName:              "Doe",
				Email:                 "john.doe@example.com",
				Phone:                 "+1234567890",
				DateOfBirth:           "1990-01-01",
				Address:               "123 Main St",
				EmergencyContactName:  "Jane Doe",
				EmergencyContactPhone: "+0987654321",
				Status:                "active",
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var got repository.Client
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Verify all fields using cmp.Diff
				want := repository.Client{
					FirstName:             "John",
					LastName:              "Doe",
					Email:                 "john.doe@example.com",
					Phone:                 "+1234567890",
					DateOfBirth:           "1990-01-01",
					Address:               "123 Main St",
					EmergencyContactName:  "Jane Doe",
					EmergencyContactPhone: "+0987654321",
					Status:                "active",
				}

				// Ignore dynamic fields (ID, timestamps)
				opts := cmpopts.IgnoreFields(repository.Client{}, "ID", "CreatedAt", "UpdatedAt")
				if diff := cmp.Diff(want, got, opts); diff != "" {
					t.Errorf("Client mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:   "Success - Default status when not provided",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane@example.com",
				// Status not provided
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					// Verify default status is set to "active"
					if diff := cmp.Diff("active", client.Status); diff != "" {
						t.Errorf("Status mismatch (-want +got):\n%s", diff)
					}
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var got repository.Client
				json.NewDecoder(w.Body).Decode(&got)

				if got.Status != "active" {
					t.Errorf("Status: want 'active', got '%s'", got.Status)
				}
			},
		},
		{
			name:   "Failure - Missing first_name",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				// FirstName missing
				LastName: "Doe",
				Email:    "test@example.com",
			},
			mockSetup:      func(m *MockClientService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "first_name, last_name, and email are required",
		},
		{
			name:   "Failure - Missing last_name",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName: "John",
				// LastName missing
				Email: "test@example.com",
			},
			mockSetup:      func(m *MockClientService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "first_name, last_name, and email are required",
		},
		{
			name:   "Failure - Missing email",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName: "John",
				LastName:  "Doe",
				// Email missing
			},
			mockSetup:      func(m *MockClientService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "first_name, last_name, and email are required",
		},
		{
			name:   "Failure - Missing all required fields",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				Phone: "+1234567890", // Only optional field
			},
			mockSetup:      func(m *MockClientService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "first_name, last_name, and email are required",
		},
		{
			name:           "Failure - Invalid JSON",
			method:         http.MethodPost,
			requestBody:    `{"invalid json`,
			mockSetup:      func(m *MockClientService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:   "Failure - Service error",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return errors.New("database connection failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create client",
		},
		{
			name:   "Failure - Wrong HTTP method (GET)",
			method: http.MethodGet,
			requestBody: CreateClientRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			mockSetup:      func(m *MockClientService) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Success - Minimal required fields only",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName: "Min",
				LastName:  "Fields",
				Email:     "min@example.com",
				// All optional fields omitted
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var got repository.Client
				json.NewDecoder(w.Body).Decode(&got)

				want := repository.Client{
					FirstName: "Min",
					LastName:  "Fields",
					Email:     "min@example.com",
					Phone:     "", // Should be empty
					Status:    "active",
				}

				opts := cmpopts.IgnoreFields(repository.Client{}, "ID", "CreatedAt", "UpdatedAt",
					"DateOfBirth", "Address", "EmergencyContactName", "EmergencyContactPhone")
				if diff := cmp.Diff(want, got, opts); diff != "" {
					t.Errorf("Client mismatch (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:   "Success - Custom inactive status",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName: "Inactive",
				LastName:  "Client",
				Email:     "inactive@example.com",
				Status:    "inactive",
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var got repository.Client
				json.NewDecoder(w.Body).Decode(&got)

				want := repository.Client{
					FirstName: "Inactive",
					LastName:  "Client",
					Email:     "inactive@example.com",
					Status:    "inactive",
				}

				opts := cmpopts.IgnoreFields(repository.Client{}, "ID", "CreatedAt", "UpdatedAt",
					"Phone", "DateOfBirth", "Address", "EmergencyContactName", "EmergencyContactPhone")
				if diff := cmp.Diff(want, got, opts); diff != "" {
					t.Errorf("Client mismatch (-want +got):\n%s", diff)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := &MockClientService{}
			tt.mockSetup(mockService)

			handler := NewClientHandler(mockService)

			// Prepare request body
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			// Create request
			req := httptest.NewRequest(tt.method, "/api/clients/add", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute handler
			handler.CreateClient(w, req)

			// Assert status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Assert error message if expected
			if tt.expectedError != "" {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					// If we can't decode as ErrorResponse, check the body directly
					bodyStr := w.Body.String()
					if !strings.Contains(bodyStr, tt.expectedError) {
						t.Errorf("Expected error containing '%s', got body: %s", tt.expectedError, bodyStr)
					}
				} else {
					if !strings.Contains(errResp.Message, tt.expectedError) && !strings.Contains(errResp.Error, tt.expectedError) {
						t.Errorf("Expected error containing '%s', got Error='%s', Message='%s'",
							tt.expectedError, errResp.Error, errResp.Message)
					}
				}
			}

			// Custom body validation
			if tt.validateBody != nil {
				tt.validateBody(t, w)
			}
		})
	}
}
