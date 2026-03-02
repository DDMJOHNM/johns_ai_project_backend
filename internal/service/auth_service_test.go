package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jmason/john_ai_project/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Mock UserRepository
type MockUserRepository struct {
	CreateUserFunc        func(ctx context.Context, user *repository.User) error
	GetUserByEmailFunc    func(ctx context.Context, email string) (*repository.User, error)
	GetUserByUsernameFunc func(ctx context.Context, username string) (*repository.User, error)
	GetUserByIDFunc       func(ctx context.Context, id string) (*repository.User, error)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *repository.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	if m.GetUserByUsernameFunc != nil {
		return m.GetUserByUsernameFunc(ctx, username)
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*repository.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, errors.New("user not found")
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		email         string
		password      string
		firstName     string
		lastName      string
		mockSetup     func(*MockUserRepository)
		expectedError string
	}{
		{
			name:      "Success - Create user",
			username:  "johndoe",
			email:     "john@example.com",
			password:  "password123",
			firstName: "John",
			lastName:  "Doe",
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
				m.GetUserByUsernameFunc = func(ctx context.Context, username string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
				m.CreateUserFunc = func(ctx context.Context, user *repository.User) error {
					return nil
				}
			},
			expectedError: "",
		},
		{
			name:          "Failure - Missing username",
			username:      "",
			email:         "john@example.com",
			password:      "password123",
			firstName:     "John",
			lastName:      "Doe",
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "username, email, password, first_name, and last_name are required",
		},
		{
			name:          "Failure - Missing email",
			username:      "johndoe",
			email:         "",
			password:      "password123",
			firstName:     "John",
			lastName:      "Doe",
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "username, email, password, first_name, and last_name are required",
		},
		{
			name:          "Failure - Missing password",
			username:      "johndoe",
			email:         "john@example.com",
			password:      "",
			firstName:     "John",
			lastName:      "Doe",
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "username, email, password, first_name, and last_name are required",
		},
		{
			name:          "Failure - Missing first_name",
			username:      "johndoe",
			email:         "john@example.com",
			password:      "password123",
			firstName:     "",
			lastName:      "Doe",
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "username, email, password, first_name, and last_name are required",
		},
		{
			name:          "Failure - Missing last_name",
			username:      "johndoe",
			email:         "john@example.com",
			password:      "password123",
			firstName:     "John",
			lastName:      "",
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "username, email, password, first_name, and last_name are required",
		},
		{
			name:          "Failure - Password too short",
			username:      "johndoe",
			email:         "john@example.com",
			password:      "short",
			firstName:     "John",
			lastName:      "Doe",
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "password must be at least 8 characters long",
		},
		{
			name:          "Failure - Invalid email format",
			username:      "johndoe",
			email:         "not-an-email",
			password:      "password123",
			firstName:     "John",
			lastName:      "Doe",
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "invalid email format",
		},
		{
			name:      "Failure - Email already exists",
			username:  "johndoe",
			email:     "existing@example.com",
			password:  "password123",
			firstName: "John",
			lastName:  "Doe",
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return &repository.User{Email: email}, nil
				}
			},
			expectedError: "user with this email already exists",
		},
		{
			name:      "Failure - Username already taken",
			username:  "takenuser",
			email:     "new@example.com",
			password:  "password123",
			firstName: "John",
			lastName:  "Doe",
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
				m.GetUserByUsernameFunc = func(ctx context.Context, username string) (*repository.User, error) {
					return &repository.User{Username: username}, nil
				}
			},
			expectedError: "username is already taken",
		},
		{
			name:      "Failure - Repository error on create",
			username:  "johndoe",
			email:     "john@example.com",
			password:  "password123",
			firstName: "John",
			lastName:  "Doe",
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
				m.GetUserByUsernameFunc = func(ctx context.Context, username string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
				m.CreateUserFunc = func(ctx context.Context, user *repository.User) error {
					return errors.New("dynamodb error")
				}
			},
			expectedError: "dynamodb error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.mockSetup(mockRepo)

			svc := NewAuthService(mockRepo, "test-jwt-secret")
			ctx := context.Background()

			user, err := svc.Register(ctx, tt.username, tt.email, tt.password, tt.firstName, tt.lastName)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if user == nil {
					t.Error("Expected user, got nil")
				}
				if user != nil && user.Username != tt.username {
					t.Errorf("Expected username %q, got %q", tt.username, user.Username)
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

func TestAuthService_Login(t *testing.T) {
	validPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		login         string
		password      string
		mockSetup     func(*MockUserRepository)
		expectToken   bool
		expectUser    bool
		expectedError string
	}{
		{
			name:     "Success - Login by email",
			login:    "john@example.com",
			password: validPassword,
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return &repository.User{
						ID:           "user-123",
						Username:     "johndoe",
						Email:        email,
						PasswordHash: string(hashedPassword),
						IsActive:     true,
					}, nil
				}
			},
			expectToken: true,
			expectUser:  true,
		},
		{
			name:     "Success - Login by username",
			login:    "johndoe",
			password: validPassword,
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
				m.GetUserByUsernameFunc = func(ctx context.Context, username string) (*repository.User, error) {
					return &repository.User{
						ID:           "user-123",
						Username:     username,
						Email:        "john@example.com",
						PasswordHash: string(hashedPassword),
						IsActive:     true,
					}, nil
				}
			},
			expectToken: true,
			expectUser:  true,
		},
		{
			name:          "Failure - Missing login",
			login:         "",
			password:      validPassword,
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "login and password are required",
		},
		{
			name:          "Failure - Missing password",
			login:         "johndoe",
			password:      "",
			mockSetup:     func(m *MockUserRepository) {},
			expectedError: "login and password are required",
		},
		{
			name:     "Failure - User not found",
			login:    "nobody@example.com",
			password: validPassword,
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
				m.GetUserByUsernameFunc = func(ctx context.Context, username string) (*repository.User, error) {
					return nil, errors.New("user not found")
				}
			},
			expectedError: "invalid credentials",
		},
		{
			name:     "Failure - Wrong password",
			login:    "john@example.com",
			password: "wrongpassword",
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return &repository.User{
						ID:           "user-123",
						Email:        email,
						PasswordHash: string(hashedPassword),
						IsActive:     true,
					}, nil
				}
			},
			expectedError: "invalid credentials",
		},
		{
			name:     "Failure - Account disabled",
			login:    "john@example.com",
			password: validPassword,
			mockSetup: func(m *MockUserRepository) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (*repository.User, error) {
					return &repository.User{
						ID:           "user-123",
						Email:        email,
						PasswordHash: string(hashedPassword),
						IsActive:     false,
					}, nil
				}
			},
			expectedError: "account is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			tt.mockSetup(mockRepo)

			svc := NewAuthService(mockRepo, "test-jwt-secret")
			ctx := context.Background()

			token, user, err := svc.Login(ctx, tt.login, tt.password)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if tt.expectToken && token == "" {
					t.Error("Expected token, got empty")
				}
				if tt.expectUser && user == nil {
					t.Error("Expected user, got nil")
				}
			}
		})
	}
}

func TestAuthService_GetUserByID(t *testing.T) {
	mockRepo := &MockUserRepository{}
	mockRepo.GetUserByIDFunc = func(ctx context.Context, id string) (*repository.User, error) {
		if id == "user-123" {
			return &repository.User{
				ID:       "user-123",
				Username: "johndoe",
				Email:    "john@example.com",
			}, nil
		}
		return nil, errors.New("user not found")
	}

	svc := NewAuthService(mockRepo, "test-secret")
	ctx := context.Background()

	user, err := svc.GetUserByID(ctx, "user-123")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if user.ID != "user-123" || user.Username != "johndoe" {
		t.Errorf("Expected user-123/johndoe, got %s/%s", user.ID, user.Username)
	}

	_, err = svc.GetUserByID(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent user")
	}
}

func TestAuthService_GenerateTokenAndValidateToken(t *testing.T) {
	svc := NewAuthService(&MockUserRepository{}, "test-jwt-secret")
	user := &repository.User{
		ID:       "user-123",
		Username: "johndoe",
		Email:    "john@example.com",
		Role:     "user",
	}

	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != user.ID || claims.Username != user.Username || claims.Email != user.Email {
		t.Errorf("Claims mismatch: got UserID=%s Username=%s Email=%s", claims.UserID, claims.Username, claims.Email)
	}

	// Invalid token
	_, err = svc.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
