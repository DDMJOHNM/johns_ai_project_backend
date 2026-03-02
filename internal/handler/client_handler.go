package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jmason/john_ai_project/internal/repository"
	"github.com/jmason/john_ai_project/internal/service"
)

type ContextKey string

const ClientIDKey ContextKey = "client_id"

// ClientService interface for dependency injection
type ClientService interface {
	GetClientList(ctx context.Context) ([]repository.Client, error)
	GetClientByID(ctx context.Context, id string) (*repository.Client, error)
	GetActiveClients(ctx context.Context) ([]repository.Client, error)
	GetInactiveClients(ctx context.Context) ([]repository.Client, error)
	CreateClient(ctx context.Context, client *repository.Client) error
}

type ClientHandler struct {
	service ClientService
}

func NewClientHandler(service ClientService) *ClientHandler {
	return &ClientHandler{
		service: service,
	}
}

func (h *ClientHandler) GetClientList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clients, err := h.service.GetClientList(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, http.StatusOK, clients)
}

func (h *ClientHandler) GetClientByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var id string
	if ctxID := r.Context().Value(ClientIDKey); ctxID != nil {
		id = ctxID.(string)
	} else {
		path := r.URL.Path
		if len(path) > len("/api/clients/") {
			id = path[len("/api/clients/"):]
		}
	}

	if id == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	client, err := h.service.GetClientByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	RespondJSON(w, http.StatusOK, client)
}

func (h *ClientHandler) GetActiveClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clients, err := h.service.GetActiveClients(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, http.StatusOK, clients)
}

func (h *ClientHandler) GetInactiveClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clients, err := h.service.GetInactiveClients(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, http.StatusOK, clients)
}

type CreateClientRequest struct {
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	Email                 string `json:"email"`
	Phone                 string `json:"phone"`
	DateOfBirth           string `json:"date_of_birth"`
	Address               string `json:"address"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	Status                string `json:"status"`
}

func (h *ClientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Map request to domain model - service handles validation, defaults, ID, timestamps
	client := &repository.Client{
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Email:                 req.Email,
		Phone:                 req.Phone,
		DateOfBirth:           req.DateOfBirth,
		Address:               req.Address,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		Status:                req.Status,
	}

	if err := h.service.CreateClient(r.Context(), client); err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrMissingRequiredFields || err == service.ErrInvalidEmail {
			statusCode = http.StatusBadRequest
		}
		RespondJSON(w, statusCode, ErrorResponse{
			Error:   "Failed to create client",
			Message: err.Error(),
		})
		return
	}

	RespondJSON(w, http.StatusCreated, client)
}
