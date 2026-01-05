package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmason/john_ai_project/internal/repository"
	"github.com/jmason/john_ai_project/internal/service"
)

// ClientHandler handles HTTP requests for client operations
type ClientHandler struct {
	service *service.ClientService
}

func NewClientHandler(service *service.ClientService) *ClientHandler {
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
	if ctxID := r.Context().Value("client_id"); ctxID != nil {
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

func (h *ClientHandler) AddClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newClient repository.Client
	//populate the new client with the current timestamp
	now := time.Now().Format(time.RFC3339)
	newClient.CreatedAt = now
	newClient.UpdatedAt = now
	newClient.Status = "active"

	if err := json.NewDecoder(r.Body).Decode(&newClient); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	newClient.ID = uuid.New().String()
	err := h.service.AddClient(r.Context(), &newClient)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, http.StatusCreated, "Client added successfully")
}
