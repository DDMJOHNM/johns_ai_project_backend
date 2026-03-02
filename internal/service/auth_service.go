package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmason/john_ai_project/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Auth validation errors
var (
	ErrAuthMissingFields     = errors.New("username, email, password, first_name, and last_name are required")
	ErrAuthInvalidPassword   = errors.New("password must be at least 8 characters long")
	ErrAuthInvalidEmail      = errors.New("invalid email format")
	ErrAuthLoginMissingFields = errors.New("login and password are required")
	ErrUserExists            = errors.New("user with this email already exists")
	ErrUsernameTaken          = errors.New("username is already taken")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAccountDisabled       = errors.New("account is disabled")
)

var authEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// UserRepository interface for dependency injection
type UserRepository interface {
	CreateUser(ctx context.Context, user *repository.User) error
	GetUserByEmail(ctx context.Context, email string) (*repository.User, error)
	GetUserByUsername(ctx context.Context, username string) (*repository.User, error)
	GetUserByID(ctx context.Context, id string) (*repository.User, error)
}

type AuthService struct {
	userRepo  UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, username, email, password, firstName, lastName string) (*repository.User, error) {
	// Validate required fields
	if username == "" || email == "" || password == "" || firstName == "" || lastName == "" {
		return nil, ErrAuthMissingFields
	}

	// Validate password length
	if len(password) < 8 {
		return nil, ErrAuthInvalidPassword
	}

	// Validate email format
	if !authEmailRegex.MatchString(email) {
		return nil, ErrAuthInvalidEmail
	}

	// Check if user already exists by email
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserExists
	}

	// Check if username is taken
	_, err = s.userRepo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, ErrUsernameTaken
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	now := time.Now().Format(time.RFC3339)
	user := &repository.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         "user",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, usernameOrEmail, password string) (string, *repository.User, error) {
	// Validate required fields
	if usernameOrEmail == "" || password == "" {
		return "", nil, ErrAuthLoginMissingFields
	}

	log.Printf("[AUTH] Login attempt for: %s", usernameOrEmail)

	// Try to get user by email first
	user, err := s.userRepo.GetUserByEmail(ctx, usernameOrEmail)
	if err != nil {
		log.Printf("[AUTH] User not found by email, trying username...")
		// If not found by email, try username
		user, err = s.userRepo.GetUserByUsername(ctx, usernameOrEmail)
		if err != nil {
			log.Printf("[AUTH] User not found by username either: %v", err)
			return "", nil, ErrInvalidCredentials
		}
		log.Printf("[AUTH] User found by username: %s", user.Username)
	} else {
		log.Printf("[AUTH] User found by email: %s", user.Email)
	}

	// Check if user is active
	if !user.IsActive {
		log.Printf("[AUTH] Account is disabled for user: %s", user.Username)
		return "", nil, ErrAccountDisabled
	}

	log.Printf("[AUTH] Comparing password hash...")
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Printf("[AUTH] Password verification failed: %v", err)
		return "", nil, ErrInvalidCredentials
	}

	log.Printf("[AUTH] Password verified successfully")
	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		log.Printf("[AUTH] Failed to generate token: %v", err)
		return "", nil, err
	}

	log.Printf("[AUTH] Login successful for user: %s", user.Username)
	return token, user, nil
}

func (s *AuthService) GenerateToken(user *repository.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "john-ai-project",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*repository.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}

