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

	"github.com/jmason/john_ai_project/internal/repository"
	"github.com/jmason/john_ai_project/internal/service"
)

// Mock AuthService
type MockAuthService struct {
	RegisterFunc      func(ctx context.Context, username, email, password, firstName, lastName string) (*repository.User, error)
	LoginFunc         func(ctx context.Context, usernameOrEmail, password string) (string, *repository.User, error)
	GetUserByIDFunc   func(ctx context.Context, userID string) (*repository.User, error)
	GenerateTokenFunc func(user *repository.User) (string, error)
	ValidateTokenFunc func(tokenString string) (*service.Claims, error)
}

func (m *MockAuthService) Register(ctx context.Context, username, email, password, firstName, lastName string) (*repository.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, username, email, password, firstName, lastName)
	}
	return nil, nil
}

func (m *MockAuthService) Login(ctx context.Context, usernameOrEmail, password string) (string, *repository.User, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, usernameOrEmail, password)
	}
	return "", nil, nil
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockAuthService) GenerateToken(user *repository.User) (string, error) {
	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(user)
	}
	return "", nil
}

func (m *MockAuthService) ValidateToken(tokenString string) (*service.Claims, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(tokenString)
	}
	return nil, nil
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(*MockAuthService)
		expectedStatus   int
		expectedError    string
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Full registration",
			requestBody: RegisterRequest{
				Username:  "johndoe",
				Email:     "john@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(m *MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, username, email, password, firstName, lastName string) (*repository.User, error) {
					return &repository.User{
						ID:        "user-123",
						Username:  username,
						Email:     email,
						FirstName: firstName,
						LastName:  lastName,
					}, nil
				}
				m.GenerateTokenFunc = func(user *repository.User) (string, error) {
					return "jwt-token-123", nil
				}
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp AuthResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Token != "jwt-token-123" {
					t.Errorf("Expected token 'jwt-token-123', got %q", resp.Token)
				}
			},
		},
		{
			name:           "Failure - Invalid JSON",
			requestBody:    `{"username": "test`,
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "Failure - Service validation (missing fields)",
			requestBody: RegisterRequest{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, username, email, password, firstName, lastName string) (*repository.User, error) {
					return nil, service.ErrAuthMissingFields
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username, email, password, first_name, and last_name are required",
		},
		{
			name: "Failure - Service validation (invalid password)",
			requestBody: RegisterRequest{
				Username:  "johndoe",
				Email:     "john@example.com",
				Password:  "short",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(m *MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, username, email, password, firstName, lastName string) (*repository.User, error) {
					return nil, service.ErrAuthInvalidPassword
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 8 characters long",
		},
		{
			name: "Failure - Service validation (email exists)",
			requestBody: RegisterRequest{
				Username:  "johndoe",
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(m *MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, username, email, password, firstName, lastName string) (*repository.User, error) {
					return nil, service.ErrUserExists
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "user with this email already exists",
		},
		{
			name: "Failure - GenerateToken error",
			requestBody: RegisterRequest{
				Username:  "johndoe",
				Email:     "john@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(m *MockAuthService) {
				m.RegisterFunc = func(ctx context.Context, username, email, password, firstName, lastName string) (*repository.User, error) {
					return &repository.User{ID: "user-123"}, nil
				}
				m.GenerateTokenFunc = func(user *repository.User) (string, error) {
					return "", context.DeadlineExceeded
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to generate token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockAuthService{}
			tt.mockSetup(mockSvc)

			handler := NewAuthHandler(mockSvc)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					bodyStr := w.Body.String()
					if !strings.Contains(bodyStr, tt.expectedError) {
						t.Errorf("Expected error containing '%s', got body: %s", tt.expectedError, bodyStr)
					}
				} else {
					if !strings.Contains(errResp.Error, tt.expectedError) && !strings.Contains(errResp.Message, tt.expectedError) {
						t.Errorf("Expected error containing '%s', got Error='%s', Message='%s'",
							tt.expectedError, errResp.Error, errResp.Message)
					}
				}
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Success - Login",
			requestBody: LoginRequest{
				Login:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAuthService) {
				m.LoginFunc = func(ctx context.Context, usernameOrEmail, password string) (string, *repository.User, error) {
					return "jwt-token", &repository.User{ID: "user-123", Username: "johndoe"}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure - Invalid JSON",
			requestBody:    `{"login": "test`,
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "Failure - Missing login/password (400)",
			requestBody: LoginRequest{
				Login:    "",
				Password: "password123",
			},
			mockSetup: func(m *MockAuthService) {
				m.LoginFunc = func(ctx context.Context, usernameOrEmail, password string) (string, *repository.User, error) {
					return "", nil, service.ErrAuthLoginMissingFields
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "login and password are required",
		},
		{
			name: "Failure - Invalid credentials (401)",
			requestBody: LoginRequest{
				Login:    "john@example.com",
				Password: "wrong",
			},
			mockSetup: func(m *MockAuthService) {
				m.LoginFunc = func(ctx context.Context, usernameOrEmail, password string) (string, *repository.User, error) {
					return "", nil, service.ErrInvalidCredentials
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid credentials",
		},
		{
			name: "Failure - Account disabled (401)",
			requestBody: LoginRequest{
				Login:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAuthService) {
				m.LoginFunc = func(ctx context.Context, usernameOrEmail, password string) (string, *repository.User, error) {
					return "", nil, service.ErrAccountDisabled
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "account is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockAuthService{}
			tt.mockSetup(mockSvc)

			handler := NewAuthHandler(mockSvc)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					bodyStr := w.Body.String()
					if !strings.Contains(bodyStr, tt.expectedError) {
						t.Errorf("Expected error containing '%s', got body: %s", tt.expectedError, bodyStr)
					}
				} else {
					if !strings.Contains(errResp.Error, tt.expectedError) && !strings.Contains(errResp.Message, tt.expectedError) {
						t.Errorf("Expected error containing '%s', got Error='%s', Message='%s'",
							tt.expectedError, errResp.Error, errResp.Message)
					}
				}
			}
		})
	}
}

func TestAuthHandler_Me(t *testing.T) {
	tests := []struct {
		name           string
		contextUserID  interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "Success - User in context",
			contextUserID: "user-123",
			mockSetup: func(m *MockAuthService) {
				m.GetUserByIDFunc = func(ctx context.Context, userID string) (*repository.User, error) {
					return &repository.User{
						ID:       userID,
						Username: "johndoe",
						Email:    "john@example.com",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure - No user_id in context",
			contextUserID:  nil,
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not found in context",
		},
		{
			name:          "Failure - User not found",
			contextUserID: "nonexistent",
			mockSetup: func(m *MockAuthService) {
				m.GetUserByIDFunc = func(ctx context.Context, userID string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockAuthService{}
			tt.mockSetup(mockSvc)

			handler := NewAuthHandler(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			if tt.contextUserID != nil {
				ctx := context.WithValue(req.Context(), "user_id", tt.contextUserID)
				req = req.WithContext(ctx)
			}
			w := httptest.NewRecorder()

			handler.Me(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					bodyStr := w.Body.String()
					if !strings.Contains(bodyStr, tt.expectedError) {
						t.Errorf("Expected error containing '%s', got body: %s", tt.expectedError, bodyStr)
					}
				} else {
					if !strings.Contains(errResp.Error, tt.expectedError) && !strings.Contains(errResp.Message, tt.expectedError) {
						t.Errorf("Expected error containing '%s', got Error='%s' Message='%s'", tt.expectedError, errResp.Error, errResp.Message)
					}
				}
			}
		})
	}
}

func TestAuthHandler_AuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedCalled bool
	}{
		{
			name:           "Failure - No Authorization header",
			authHeader:     "",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Failure - Invalid format (no Bearer)",
			authHeader:     "token123",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Failure - Invalid token",
			authHeader: "Bearer invalid-token",
			mockSetup: func(m *MockAuthService) {
				m.ValidateTokenFunc = func(tokenString string) (*service.Claims, error) {
					return nil, context.DeadlineExceeded
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Success - Valid token passes to next handler",
			authHeader: "Bearer valid-token",
			mockSetup: func(m *MockAuthService) {
				m.ValidateTokenFunc = func(tokenString string) (*service.Claims, error) {
					return &service.Claims{UserID: "user-123", Username: "johndoe", Email: "john@example.com", Role: "user"}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			mockSvc := &MockAuthService{}
			tt.mockSetup(mockSvc)

			handler := NewAuthHandler(mockSvc)
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.AuthMiddleware(nextHandler.ServeHTTP).ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedCalled && !called {
				t.Error("Expected next handler to be called")
			}
		})
	}
}
