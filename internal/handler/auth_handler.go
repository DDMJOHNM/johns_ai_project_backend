package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jmason/john_ai_project/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Login    string `json:"login"`    // Can be email or username
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Validate
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required fields",
			Message: "username, email, password, first_name, and last_name are required",
		})
		return
	}

	// Basic password validation
	if len(req.Password) < 8 {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid password",
			Message: "Password must be at least 8 characters long",
		})
		return
	}

	user, err := h.authService.Register(r.Context(), req.Username, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Registration failed",
			Message: err.Error(),
		})
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate token",
			Message: err.Error(),
		})
		return
	}

	RespondJSON(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	if req.Login == "" || req.Password == "" {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required fields",
			Message: "login and password are required",
		})
		return
	}

	token, user, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		RespondJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error:   "Login failed",
			Message: err.Error(),
		})
		return
	}

	RespondJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		RespondJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not found in context",
		})
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		RespondJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "User not found",
			Message: err.Error(),
		})
		return
	}

	RespondJSON(w, http.StatusOK, user)
}

// AuthMiddleware protects routes by requiring a valid JWT token
func (h *AuthHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			RespondJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authorization header required",
			})
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			RespondJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid authorization format. Use: Bearer <token>",
			})
			return
		}

		token := parts[1]

		// Validate token
		claims, err := h.authService.ValidateToken(token)
		if err != nil {
			RespondJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid or expired token",
			})
			return
		}

		// Add user info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_username", claims.Username)
		ctx = context.WithValue(ctx, "user_role", claims.Role)

		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

