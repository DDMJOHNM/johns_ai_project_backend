package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmason/john_ai_project/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
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
	// Check if user already exists by email
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Check if username is taken
	_, err = s.userRepo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, fmt.Errorf("username is already taken")
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
	log.Printf("[AUTH] Login attempt for: %s", usernameOrEmail)
	
	// Try to get user by email first
	user, err := s.userRepo.GetUserByEmail(ctx, usernameOrEmail)
	if err != nil {
		log.Printf("[AUTH] User not found by email, trying username...")
		// If not found by email, try username
		user, err = s.userRepo.GetUserByUsername(ctx, usernameOrEmail)
		if err != nil {
			log.Printf("[AUTH] User not found by username either: %v", err)
			return "", nil, fmt.Errorf("invalid credentials")
		}
		log.Printf("[AUTH] User found by username: %s", user.Username)
	} else {
		log.Printf("[AUTH] User found by email: %s", user.Email)
	}

	// Check if user is active
	if !user.IsActive {
		log.Printf("[AUTH] Account is disabled for user: %s", user.Username)
		return "", nil, fmt.Errorf("account is disabled")
	}

	log.Printf("[AUTH] Comparing password hash...")
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Printf("[AUTH] Password verification failed: %v", err)
		return "", nil, fmt.Errorf("invalid credentials")
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

