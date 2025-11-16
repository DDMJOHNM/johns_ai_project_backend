package handler

import (
	"net/http"

	"github.com/jmason/john_ai_project/internal/service"
)

// ClientHandler handles HTTP requests for client operations
type ClientHandler struct {
	service *service.ClientService
}

// NewClientHandler creates a new client handler
func NewClientHandler(service *service.ClientService) *ClientHandler {
	return &ClientHandler{
		service: service,
	}
}

// GetClientList handles GET /api/clients - returns all clients
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

// GetClientByID handles GET /api/clients/:id - returns a single client
func (h *ClientHandler) GetClientByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get ID from context (set by router) or from URL path
	var id string
	if ctxID := r.Context().Value("client_id"); ctxID != nil {
		id = ctxID.(string)
	} else {
		// Fallback: extract from path
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

// GetActiveClients handles GET /api/clients/active - returns all active clients
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

// GetInactiveClients handles GET /api/clients/inactive - returns all inactive clients
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

