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
	"github.com/jmason/john_ai_project/internal/service"
)

// Mock ClientService
type MockClientService struct {
	CreateClientFunc       func(ctx context.Context, client *repository.Client) error
	GetClientListFunc      func(ctx context.Context) ([]repository.Client, error)
	GetClientByIDFunc      func(ctx context.Context, id string) (*repository.Client, error)
	GetClientByEmailFunc   func(ctx context.Context, email string) (*repository.Client, error)
	GetActiveClientsFunc   func(ctx context.Context) ([]repository.Client, error)
	GetInactiveClientsFunc func(ctx context.Context) ([]repository.Client, error)
	UpdateClientFunc       func(ctx context.Context, clientID string, in service.ClientUpdateInput) error
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

func (m *MockClientService) GetClientByEmail(ctx context.Context, email string) (*repository.Client, error) {
	if m.GetClientByEmailFunc != nil {
		return m.GetClientByEmailFunc(ctx, email)
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

func (m *MockClientService) UpdateClient(ctx context.Context, clientID string, in service.ClientUpdateInput) error {
	if m.UpdateClientFunc != nil {
		return m.UpdateClientFunc(ctx, clientID, in)
	}
	return nil
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
				// Status not provided - service sets default
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					client.Status = "active" // Simulate service setting default
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
			name:   "Failure - Missing first_name (service validation)",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				LastName: "Doe",
				Email:    "test@example.com",
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return service.ErrMissingRequiredFields
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "first_name, last_name, and email are required",
		},
		{
			name:   "Failure - Missing last_name (service validation)",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName: "John",
				Email:     "test@example.com",
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return service.ErrMissingRequiredFields
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "first_name, last_name, and email are required",
		},
		{
			name:   "Failure - Missing email (service validation)",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return service.ErrMissingRequiredFields
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "first_name, last_name, and email are required",
		},
		{
			name:   "Failure - Missing all required fields (service validation)",
			method: http.MethodPost,
			requestBody: CreateClientRequest{
				Phone: "+1234567890",
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					return service.ErrMissingRequiredFields
				}
			},
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
			},
			mockSetup: func(m *MockClientService) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					client.Status = "active" // Simulate service setting default
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
					Phone:     "",
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

func TestClientHandler_GetClientByEmail(t *testing.T) {
	tests := []struct {
		name           string
		rawQuery       string
		mockSetup      func(*MockClientService)
		expectedStatus int
	}{
		{
			name:     "success",
			rawQuery: "email=jane%40example.com",
			mockSetup: func(m *MockClientService) {
				m.GetClientByEmailFunc = func(ctx context.Context, email string) (*repository.Client, error) {
					return &repository.Client{ID: "client-1", Email: email, FirstName: "Jane"}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "validation error from service",
			rawQuery: "email=",
			mockSetup: func(m *MockClientService) {
				m.GetClientByEmailFunc = func(ctx context.Context, email string) (*repository.Client, error) {
					return nil, service.ErrMissingEmail
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "not found",
			rawQuery: "email=no%40one.com",
			mockSetup: func(m *MockClientService) {
				m.GetClientByEmailFunc = func(ctx context.Context, email string) (*repository.Client, error) {
					return nil, errors.New("failed to get client by email: client not found for email: no@one.com")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClientService{}
			tt.mockSetup(mock)
			h := NewClientHandler(mock)

			url := "/api/clients/by-email"
			if tt.rawQuery != "" {
				url += "?" + tt.rawQuery
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			h.GetClientByEmail(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}

func TestClientJSON_responseIncludesExtendedFields(t *testing.T) {
	c := repository.Client{
		ID:                  "c1",
		FirstName:           "A",
		LastName:            "B",
		Email:               "a@b.co",
		RequestedCounsellor: "Dr. X",
		Urgency:             "high",
		NextAppointment:     "2025-04-01T10:00:00Z",
		Notes:               []repository.Note{{Date: "2025-01-01T00:00:00Z", ClientID: "c1", Note: "n"}},
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"requested_counsellor", "urgency", "next_appointment", "notes"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %q in JSON: %s", k, string(b))
		}
	}
}

func TestUpdateClientRequest_camelCaseAndCounsellorIDJSON(t *testing.T) {
	raw := []byte(`{
		"requestedCounsellor": "Dr. A",
		"urgencyLevel": "high",
		"nextAppointment": "2025-06-01T12:00:00Z"
	}`)
	var req UpdateClientRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.RequestedCounsellor == nil || *req.RequestedCounsellor != "Dr. A" {
		t.Fatalf("requested_counsellor = %v", req.RequestedCounsellor)
	}
	if req.Urgency == nil || *req.Urgency != "high" {
		t.Fatalf("urgency = %v", req.Urgency)
	}
	if req.NextAppointment == nil || *req.NextAppointment != "2025-06-01T12:00:00Z" {
		t.Fatalf("next_appointment = %v", req.NextAppointment)
	}

	raw2 := []byte(`{"counsellor_id": "user-counsellor-1", "urgency_level": "low"}`)
	var req2 UpdateClientRequest
	if err := json.Unmarshal(raw2, &req2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req2.RequestedCounsellor == nil || *req2.RequestedCounsellor != "user-counsellor-1" {
		t.Fatalf("counsellor_id -> requested_counsellor = %v", req2.RequestedCounsellor)
	}
	if req2.Urgency == nil || *req2.Urgency != "low" {
		t.Fatalf("urgency_level = %v", req2.Urgency)
	}
}

func TestUpdateClientRequest_notesEmptyArrayOmitted(t *testing.T) {
	raw := []byte(`{"requestedCounsellor":"Dr. X","urgency":"high","notes":[],"initial_note":{}}`)
	var req UpdateClientRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.NotesList != nil {
		t.Fatalf("NotesList = %v, want nil (empty notes must not clear DB)", req.NotesList)
	}
}

func TestUpdateClientRequest_numericUrgencyAndCounsellorObject(t *testing.T) {
	raw := []byte(`{"urgency":2,"counsellor":{"name":"Dr. Smith","id":"u-1"}}`)
	var req UpdateClientRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Urgency == nil || *req.Urgency != "2" {
		t.Fatalf("urgency = %v", req.Urgency)
	}
	if req.RequestedCounsellor == nil || *req.RequestedCounsellor != "Dr. Smith" {
		t.Fatalf("requested_counsellor = %v", req.RequestedCounsellor)
	}
}

func TestUpdateClientRequest_notesObjectJSON(t *testing.T) {
	raw := []byte(`{
    "notes" : {"date":"2026-03-25T12:24:15+13:00", "client_id":"client-cf186df5-d4e5-4281-84ed-77ac7708cc38", "note":"test note"}
}`)
	var req UpdateClientRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.Notes == nil {
		t.Fatal("expected req.Notes set")
	}
	if req.Notes.Note != "test note" || req.Notes.ClientID != "client-cf186df5-d4e5-4281-84ed-77ac7708cc38" {
		t.Fatalf("Notes = %+v", req.Notes)
	}
}

func TestClientHandler_UpdateClient(t *testing.T) {
	bodyJSON := []byte(`{"first_name":"Jane"}`)

	t.Run("success", func(t *testing.T) {
		var gotID string
		mock := &MockClientService{
			UpdateClientFunc: func(ctx context.Context, id string, in service.ClientUpdateInput) error {
				gotID = id
				if in.FirstName == nil || *in.FirstName != "Jane" {
					t.Fatalf("FirstName = %v", in.FirstName)
				}
				return nil
			},
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return &repository.Client{ID: id, FirstName: "Jane", LastName: "Doe", Email: "jane@example.com"}, nil
			},
		}
		h := NewClientHandler(mock)
		req := httptest.NewRequest(http.MethodPut, "/api/clients/client-99", bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), ClientIDKey, "client-99"))
		w := httptest.NewRecorder()
		h.UpdateClient(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if gotID != "client-99" {
			t.Errorf("id = %q, want client-99", gotID)
		}
	})

	t.Run("success PATCH", func(t *testing.T) {
		mock := &MockClientService{
			UpdateClientFunc: func(ctx context.Context, id string, in service.ClientUpdateInput) error {
				return nil
			},
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return &repository.Client{ID: id, FirstName: "Jane", LastName: "Doe", Email: "jane@example.com"}, nil
			},
		}
		h := NewClientHandler(mock)
		req := httptest.NewRequest(http.MethodPatch, "/api/clients/client-99", bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), ClientIDKey, "client-99"))
		w := httptest.NewRecorder()
		h.UpdateClient(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("PATCH status = %d", w.Code)
		}
	})

	t.Run("reserved path id", func(t *testing.T) {
		mock := &MockClientService{}
		h := NewClientHandler(mock)
		req := httptest.NewRequest(http.MethodPut, "/api/clients/active", bytes.NewReader(bodyJSON))
		req = req.WithContext(context.WithValue(req.Context(), ClientIDKey, "active"))
		w := httptest.NewRecorder()
		h.UpdateClient(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		mock := &MockClientService{
			UpdateClientFunc: func(ctx context.Context, clientID string, in service.ClientUpdateInput) error {
				return service.ErrInvalidEmail
			},
		}
		h := NewClientHandler(mock)
		req := httptest.NewRequest(http.MethodPut, "/api/clients/c1", bytes.NewReader(bodyJSON))
		req = req.WithContext(context.WithValue(req.Context(), ClientIDKey, "c1"))
		w := httptest.NewRecorder()
		h.UpdateClient(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})
}
