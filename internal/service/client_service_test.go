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
	GetClientByEmailFunc   func(ctx context.Context, email string) (*repository.Client, error)
	GetClientsByStatusFunc func(ctx context.Context, status string) ([]repository.Client, error)
	UpdateClientFunc       func(ctx context.Context, id string, patch repository.ClientPatch) error
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

func (m *MockClientRepository) GetClientByEmail(ctx context.Context, email string) (*repository.Client, error) {
	if m.GetClientByEmailFunc != nil {
		return m.GetClientByEmailFunc(ctx, email)
	}
	return nil, errors.New("client not found for email: " + email)
}

func (m *MockClientRepository) GetClientsByStatus(ctx context.Context, status string) ([]repository.Client, error) {
	if m.GetClientsByStatusFunc != nil {
		return m.GetClientsByStatusFunc(ctx, status)
	}
	return nil, nil
}

func (m *MockClientRepository) UpdateClient(ctx context.Context, id string, patch repository.ClientPatch) error {
	if m.UpdateClientFunc != nil {
		return m.UpdateClientFunc(ctx, id, patch)
	}
	return nil
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
		{
			name: "Failure - Email already exists",
			client: &repository.Client{
				FirstName: "Dup",
				LastName:  "User",
				Email:     "dup@example.com",
			},
			mockSetup: func(m *MockClientRepository) {
				m.GetClientByEmailFunc = func(ctx context.Context, email string) (*repository.Client, error) {
					return &repository.Client{ID: "existing", Email: email}, nil
				}
			},
			expectedError: "a client with this email already exists",
		},
		{
			name: "Failure - Missing first_name",
			client: &repository.Client{
				LastName: "Doe",
				Email:    "test@example.com",
			},
			mockSetup:     func(m *MockClientRepository) {},
			expectedError: "first_name, last_name, and email are required",
		},
		{
			name: "Failure - Missing last_name",
			client: &repository.Client{
				FirstName: "John",
				Email:     "test@example.com",
			},
			mockSetup:     func(m *MockClientRepository) {},
			expectedError: "first_name, last_name, and email are required",
		},
		{
			name: "Failure - Missing email",
			client: &repository.Client{
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup:     func(m *MockClientRepository) {},
			expectedError: "first_name, last_name, and email are required",
		},
		{
			name: "Failure - Invalid email format",
			client: &repository.Client{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "not-an-email",
			},
			mockSetup:     func(m *MockClientRepository) {},
			expectedError: "invalid email format",
		},
		{
			name: "Success - Default status when empty",
			client: &repository.Client{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane@example.com",
				Status:    "",
			},
			mockSetup: func(m *MockClientRepository) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					if client.Status != "active" {
						t.Errorf("Expected default status 'active', got '%s'", client.Status)
					}
					return nil
				}
			},
			expectedError: "",
		},
		{
			name: "Success - Generates ID when empty",
			client: &repository.Client{
				FirstName: "New",
				LastName:  "User",
				Email:     "new@example.com",
			},
			mockSetup: func(m *MockClientRepository) {
				m.CreateClientFunc = func(ctx context.Context, client *repository.Client) error {
					if client.ID == "" {
						t.Error("Expected ID to be generated")
					}
					if !strings.HasPrefix(client.ID, "client-") {
						t.Errorf("Expected ID to start with 'client-', got %s", client.ID)
					}
					return nil
				}
			},
			expectedError: "",
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

func TestClientService_GetClientByEmail(t *testing.T) {
	want := &repository.Client{
		ID:        "client-1",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo := &MockClientRepository{
			GetClientByEmailFunc: func(ctx context.Context, email string) (*repository.Client, error) {
				if email != want.Email {
					t.Errorf("email = %q, want %q", email, want.Email)
				}
				return want, nil
			},
		}
		svc := NewClientService(mockRepo)
		got, err := svc.GetClientByEmail(context.Background(), want.Email)
		if err != nil {
			t.Fatalf("GetClientByEmail: %v", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("client mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("empty email", func(t *testing.T) {
		svc := NewClientService(&MockClientRepository{})
		_, err := svc.GetClientByEmail(context.Background(), "")
		if err != ErrMissingEmail {
			t.Fatalf("err = %v, want ErrMissingEmail", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		svc := NewClientService(&MockClientRepository{})
		_, err := svc.GetClientByEmail(context.Background(), "not-email")
		if err != ErrInvalidEmail {
			t.Fatalf("err = %v, want ErrInvalidEmail", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &MockClientRepository{
			GetClientByEmailFunc: func(ctx context.Context, email string) (*repository.Client, error) {
				return nil, errors.New("db down")
			},
		}
		svc := NewClientService(mockRepo)
		_, err := svc.GetClientByEmail(context.Background(), "a@b.co")
		if err == nil || !strings.Contains(err.Error(), "failed to get client by email") {
			t.Fatalf("err = %v, want wrapped failure", err)
		}
	})
}

func TestClientService_UpdateClient(t *testing.T) {
	fn := "Jane"
	ln := "Doe"
	em := "jane@example.com"
	base := &repository.Client{
		ID:        "client-1",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Status:    "active",
	}

	t.Run("success first name only", func(t *testing.T) {
		mockRepo := &MockClientRepository{
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return base, nil
			},
			UpdateClientFunc: func(ctx context.Context, id string, patch repository.ClientPatch) error {
				if patch.FirstName == nil || *patch.FirstName != "Janet" {
					t.Fatalf("patch.FirstName = %v", patch.FirstName)
				}
				return nil
			},
		}
		svc := NewClientService(mockRepo)
		janet := "Janet"
		if err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{FirstName: &janet}); err != nil {
			t.Fatalf("UpdateClient: %v", err)
		}
	})

	t.Run("initial note with blank body does not overwrite notes", func(t *testing.T) {
		blanks := repository.Note{Note: "  "}
		rc := "Dr. Z"
		mockRepo := &MockClientRepository{
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return &repository.Client{ID: id, Notes: []repository.Note{{Note: "keep"}}}, nil
			},
			UpdateClientFunc: func(ctx context.Context, id string, patch repository.ClientPatch) error {
				if patch.Notes != nil {
					t.Fatalf("patch.Notes = %v, want nil", patch.Notes)
				}
				if patch.RequestedCounsellor == nil || *patch.RequestedCounsellor != rc {
					t.Fatalf("patch.RequestedCounsellor = %v", patch.RequestedCounsellor)
				}
				return nil
			},
		}
		svc := NewClientService(mockRepo)
		if err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{
			InitialNote:         &blanks,
			RequestedCounsellor: &rc,
		}); err != nil {
			t.Fatalf("UpdateClient: %v", err)
		}
	})

	t.Run("success initial note", func(t *testing.T) {
		mockRepo := &MockClientRepository{
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return &repository.Client{ID: id, Notes: []repository.Note{}}, nil
			},
			UpdateClientFunc: func(ctx context.Context, id string, patch repository.ClientPatch) error {
				if patch.Notes == nil || len(*patch.Notes) != 1 || (*patch.Notes)[0].Note != "hello" {
					t.Fatalf("patch.Notes = %v", patch.Notes)
				}
				return nil
			},
		}
		svc := NewClientService(mockRepo)
		n := repository.Note{Note: "hello"}
		if err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{InitialNote: &n}); err != nil {
			t.Fatalf("UpdateClient: %v", err)
		}
	})

	t.Run("success notes list replaces entire list", func(t *testing.T) {
		full := []repository.Note{
			{Date: "2025-01-01T00:00:00Z", ClientID: "client-1", Note: "a"},
			{Date: "2025-01-02T00:00:00Z", ClientID: "client-1", Note: "b"},
		}
		mockRepo := &MockClientRepository{
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return base, nil
			},
			UpdateClientFunc: func(ctx context.Context, id string, patch repository.ClientPatch) error {
				if patch.Notes == nil || len(*patch.Notes) != 2 {
					t.Fatalf("patch.Notes = %v", patch.Notes)
				}
				return nil
			},
		}
		svc := NewClientService(mockRepo)
		if err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{NotesList: &full}); err != nil {
			t.Fatalf("UpdateClient: %v", err)
		}
	})

	t.Run("missing id", func(t *testing.T) {
		svc := NewClientService(&MockClientRepository{})
		err := svc.UpdateClient(context.Background(), "", ClientUpdateInput{FirstName: &fn})
		if err != ErrMissingClientID {
			t.Fatalf("err = %v, want ErrMissingClientID", err)
		}
	})

	t.Run("no fields", func(t *testing.T) {
		svc := NewClientService(&MockClientRepository{})
		err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{})
		if err != ErrNoFieldsToUpdate {
			t.Fatalf("err = %v, want ErrNoFieldsToUpdate", err)
		}
	})

	t.Run("empty first name when provided", func(t *testing.T) {
		svc := NewClientService(&MockClientRepository{})
		empty := "  "
		err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{FirstName: &empty})
		if err != ErrMissingRequiredFields {
			t.Fatalf("err = %v, want ErrMissingRequiredFields", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		svc := NewClientService(&MockClientRepository{
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return base, nil
			},
		})
		bad := "bad"
		err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{Email: &bad})
		if err != ErrInvalidEmail {
			t.Fatalf("err = %v, want ErrInvalidEmail", err)
		}
	})

	t.Run("success requested counsellor urgency next appointment", func(t *testing.T) {
		rc := "Sarah Johnson, MA, LPC"
		urg := "soon"
		next := "2025-04-01T10:00:00Z"
		mockRepo := &MockClientRepository{
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return base, nil
			},
			UpdateClientFunc: func(ctx context.Context, id string, patch repository.ClientPatch) error {
				if patch.RequestedCounsellor == nil || *patch.RequestedCounsellor != rc {
					t.Fatalf("patch.RequestedCounsellor = %v", patch.RequestedCounsellor)
				}
				if patch.Urgency == nil || *patch.Urgency != urg {
					t.Fatalf("patch.Urgency = %v", patch.Urgency)
				}
				if patch.NextAppointment == nil || *patch.NextAppointment != next {
					t.Fatalf("patch.NextAppointment = %v", patch.NextAppointment)
				}
				return nil
			},
		}
		svc := NewClientService(mockRepo)
		err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{
			RequestedCounsellor: &rc,
			Urgency:             &urg,
			NextAppointment:     &next,
		})
		if err != nil {
			t.Fatalf("UpdateClient: %v", err)
		}
	})

	t.Run("repository error on update", func(t *testing.T) {
		mockRepo := &MockClientRepository{
			GetClientByIDFunc: func(ctx context.Context, id string) (*repository.Client, error) {
				return base, nil
			},
			UpdateClientFunc: func(ctx context.Context, id string, patch repository.ClientPatch) error {
				return errors.New("db down")
			},
		}
		svc := NewClientService(mockRepo)
		err := svc.UpdateClient(context.Background(), "client-1", ClientUpdateInput{FirstName: &fn, LastName: &ln, Email: &em})
		if err == nil || !strings.Contains(err.Error(), "failed to update client") {
			t.Fatalf("err = %v, want wrapped failure", err)
		}
	})
}
