package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmason/john_ai_project/internal/repository"
	"github.com/jmason/john_ai_project/internal/service"
)

type ContextKey string

const ClientIDKey ContextKey = "client_id"

// ClientService interface for dependency injection
type ClientService interface {
	GetClientList(ctx context.Context) ([]repository.Client, error)
	GetClientByID(ctx context.Context, id string) (*repository.Client, error)
	GetClientByEmail(ctx context.Context, email string) (*repository.Client, error)
	GetActiveClients(ctx context.Context) ([]repository.Client, error)
	GetInactiveClients(ctx context.Context) ([]repository.Client, error)
	CreateClient(ctx context.Context, client *repository.Client) error
	UpdateClient(ctx context.Context, clientID string, in service.ClientUpdateInput) error
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

func (h *ClientHandler) GetClientByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := strings.TrimSpace(r.URL.Query().Get("email"))
	client, err := h.service.GetClientByEmail(r.Context(), email)
	if err != nil {
		switch err {
		case service.ErrMissingEmail, service.ErrInvalidEmail:
			RespondJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid email",
				Message: err.Error(),
			})
			return
		default:
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
	FirstName             string            `json:"first_name"`
	LastName              string            `json:"last_name"`
	Email                 string            `json:"email"`
	Phone                 string            `json:"phone"`
	DateOfBirth           string            `json:"date_of_birth"`
	Address               string            `json:"address"`
	EmergencyContactName  string            `json:"emergency_contact_name"`
	EmergencyContactPhone string            `json:"emergency_contact_phone"`
	Status                string            `json:"status"`
	RequestedCounsellor   string            `json:"requested_counsellor"`
	Urgency               string            `json:"urgency"`
	NextAppointment       string            `json:"next_appointment"`
	Notes                 []repository.Note `json:"notes,omitempty"`
}

// UnmarshalJSON accepts the same alternate keys as UpdateClientRequest for create flows from JS clients.
func (r *CreateClientRequest) UnmarshalJSON(data []byte) error {
	type Alias CreateClientRequest
	aux := &struct {
		*Alias
		RequestedCounsellorCamel string `json:"requestedCounsellor,omitempty"`
		RequestedCounselorSnake  string `json:"requested_counselor,omitempty"`
		CounsellorIDSnake        string `json:"counsellor_id,omitempty"`
		CounsellorIDCamel        string `json:"counsellorId,omitempty"`
		UrgencyLevel             string `json:"urgencyLevel,omitempty"`
		UrgencyLevelSnake        string `json:"urgency_level,omitempty"`
		NextAppointmentCamel     string `json:"nextAppointment,omitempty"`
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.RequestedCounsellor == "" {
		switch {
		case aux.RequestedCounsellorCamel != "":
			r.RequestedCounsellor = aux.RequestedCounsellorCamel
		case aux.RequestedCounselorSnake != "":
			r.RequestedCounsellor = aux.RequestedCounselorSnake
		case aux.CounsellorIDSnake != "":
			r.RequestedCounsellor = aux.CounsellorIDSnake
		case aux.CounsellorIDCamel != "":
			r.RequestedCounsellor = aux.CounsellorIDCamel
		}
	}
	if r.Urgency == "" {
		switch {
		case aux.UrgencyLevel != "":
			r.Urgency = aux.UrgencyLevel
		case aux.UrgencyLevelSnake != "":
			r.Urgency = aux.UrgencyLevelSnake
		}
	}
	if r.NextAppointment == "" && aux.NextAppointmentCamel != "" {
		r.NextAppointment = aux.NextAppointmentCamel
	}
	return nil
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
		RequestedCounsellor:   req.RequestedCounsellor,
		Urgency:               req.Urgency,
		NextAppointment:       req.NextAppointment,
		Notes:                 req.Notes,
	}

	if err := h.service.CreateClient(r.Context(), client); err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrMissingRequiredFields || err == service.ErrInvalidEmail {
			statusCode = http.StatusBadRequest
		}
		if err == service.ErrEmailAlreadyExists {
			statusCode = http.StatusConflict
		}
		RespondJSON(w, statusCode, ErrorResponse{
			Error:   "Failed to create client",
			Message: err.Error(),
		})
		return
	}

	RespondJSON(w, http.StatusCreated, client)
}

// ReservedClientPathID reports whether id is a fixed route segment under /api/clients/, not a client id.
func ReservedClientPathID(id string) bool {
	switch id {
	case "active", "inactive", "add", "by-email":
		return true
	default:
		return false
	}
}

// UpdateClientRequest is a partial update: include only fields to change.
// Notes is a single note object; it updates the first entry in the client's notes list (same as initial_note).
// NotesList replaces the entire notes array (use [] to clear). If set, it takes precedence over initial_note/notes.
//
// JSON also accepts common frontend aliases (camelCase, counsellor id, US spelling) — see UnmarshalJSON.
type UpdateClientRequest struct {
	FirstName           *string            `json:"first_name,omitempty"`
	LastName            *string            `json:"last_name,omitempty"`
	Email               *string            `json:"email,omitempty"`
	InitialNote         *repository.Note   `json:"initial_note,omitempty"`
	Notes               *repository.Note   `json:"notes,omitempty"`
	NotesList           *[]repository.Note `json:"notes_list,omitempty"`
	RequestedCounsellor *string            `json:"requested_counsellor,omitempty"`
	Urgency             *string            `json:"urgency,omitempty"`
	NextAppointment     *string            `json:"next_appointment,omitempty"`
}

// UnmarshalJSON maps alternate keys used by JS clients and coerces shapes that would otherwise
// fail or be ignored (e.g. urgency as number, counsellor as nested object).
func (r *UpdateClientRequest) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m == nil {
		return nil
	}
	normalizeNotesInClientUpdateMap(m)
	normalizeClientUpdateMap(m)
	flat, err := json.Marshal(m)
	if err != nil {
		return err
	}
	type Alias UpdateClientRequest
	return json.Unmarshal(flat, (*Alias)(r))
}

// normalizeNotesInClientUpdateMap maps "notes" JSON arrays to notes_list (full replace) and drops
// empty arrays so we do not clear stored notes when the UI sends notes: [] by default.
func normalizeNotesInClientUpdateMap(m map[string]json.RawMessage) {
	if raw, ok := m["notes"]; ok {
		s := bytes.TrimSpace(raw)
		if len(s) > 0 && s[0] == '[' {
			if bytes.Equal(s, []byte("[]")) {
				delete(m, "notes")
			} else {
				m["notes_list"] = raw
				delete(m, "notes")
			}
		}
	}
	if raw, ok := m["notesList"]; ok {
		s := bytes.TrimSpace(raw)
		if len(s) > 0 && s[0] == '[' {
			if bytes.Equal(s, []byte("[]")) {
				delete(m, "notesList")
			} else {
				m["notes_list"] = raw
				delete(m, "notesList")
			}
		}
	}
}

func normalizeClientUpdateMap(m map[string]json.RawMessage) {
	// Urgency is often sent as a number (dropdown index) or bool; store as string for DynamoDB.
	for _, k := range []string{"urgency", "urgencyLevel", "urgency_level", "priority"} {
		if raw, ok := m[k]; ok {
			if p := rawJSONStringFromAny(raw); p != nil {
				b, _ := json.Marshal(*p)
				m[k] = b
			}
		}
	}
	if _, has := m["urgency"]; !has {
		for _, k := range []string{"urgencyLevel", "urgency_level", "priority"} {
			if raw, ok := m[k]; ok {
				if p := rawJSONStringFromAny(raw); p != nil {
					b, _ := json.Marshal(*p)
					m["urgency"] = b
					break
				}
			}
		}
	}

	if _, has := m["requested_counsellor"]; !has {
		if raw, ok := m["counsellor"]; ok {
			if s := extractCounsellorDisplay(raw); s != "" {
				b, _ := json.Marshal(s)
				m["requested_counsellor"] = b
			}
		}
	}
	if _, has := m["requested_counsellor"]; !has {
		for _, k := range []string{
			"requestedCounsellor", "requested_counselor", "counsellor_id", "counsellorId",
			"assignedCounsellor", "assigned_counsellor", "selectedCounsellor", "selected_counsellor",
			"counsellorName", "counsellor_name",
		} {
			if raw, ok := m[k]; ok {
				if s := extractCounsellorDisplay(raw); s != "" {
					b, _ := json.Marshal(s)
					m["requested_counsellor"] = b
					break
				}
				if p := rawJSONStringFromAny(raw); p != nil {
					b, _ := json.Marshal(*p)
					m["requested_counsellor"] = b
					break
				}
			}
		}
	}

	if raw, ok := m["nextAppointment"]; ok && m["next_appointment"] == nil {
		m["next_appointment"] = raw
	}
}

func rawJSONStringFromAny(raw json.RawMessage) *string {
	if raw == nil {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return &s
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		ss := strconv.FormatFloat(f, 'f', -1, 64)
		return &ss
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		ss := strconv.Itoa(n)
		return &ss
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		ss := strconv.FormatBool(b)
		return &ss
	}
	return nil
}

func extractCounsellorDisplay(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return strings.TrimSpace(s)
	}
	var obj struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		FullName    string `json:"fullName"`
		Label       string `json:"label"`
		Title       string `json:"title"`
		Email       string `json:"email"`
	}
	if json.Unmarshal(raw, &obj) != nil {
		return ""
	}
	for _, v := range []string{obj.Name, obj.DisplayName, obj.FullName, obj.Label, obj.Title, obj.Email, obj.ID} {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}

func (h *ClientHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
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

	if id == "" || ReservedClientPathID(id) {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid client ID",
			Message: "A valid client id is required in the URL path",
		})
		return
	}

	var req UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	initialNote := req.InitialNote
	if initialNote == nil {
		initialNote = req.Notes
	}
	// Empty {} decodes as non-nil Note with blank body; do not overwrite stored initial consult text.
	if initialNote != nil && strings.TrimSpace(initialNote.Note) == "" {
		initialNote = nil
	}
	in := service.ClientUpdateInput{
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Email:               req.Email,
		InitialNote:         initialNote,
		NotesList:           req.NotesList,
		RequestedCounsellor: req.RequestedCounsellor,
		Urgency:             req.Urgency,
		NextAppointment:     req.NextAppointment,
	}

	if err := h.service.UpdateClient(r.Context(), id, in); err != nil {
		statusCode := http.StatusInternalServerError
		switch err {
		case service.ErrMissingClientID, service.ErrMissingRequiredFields, service.ErrInvalidEmail, service.ErrNoFieldsToUpdate:
			statusCode = http.StatusBadRequest
		}
		if strings.Contains(err.Error(), "failed to load client") {
			statusCode = http.StatusNotFound
		}
		RespondJSON(w, statusCode, ErrorResponse{
			Error:   "Failed to update client",
			Message: err.Error(),
		})
		return
	}

	updated, err := h.service.GetClientByID(r.Context(), id)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to load updated client",
			Message: err.Error(),
		})
		return
	}

	RespondJSON(w, http.StatusOK, updated)
}
