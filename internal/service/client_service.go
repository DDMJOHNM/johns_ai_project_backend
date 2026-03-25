package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmason/john_ai_project/internal/repository"
)

// Validation errors for CreateClient
var (
	ErrMissingRequiredFields = errors.New("first_name, last_name, and email are required")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrMissingEmail          = errors.New("email is required")
	ErrMissingClientID       = errors.New("client id is required")
	ErrNoFieldsToUpdate      = errors.New("provide at least one field to update")
	ErrEmailAlreadyExists    = errors.New("a client with this email already exists")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ClientRepository interface for dependency injection
type ClientRepository interface {
	GetClientList(ctx context.Context) ([]repository.Client, error)
	GetClientByID(ctx context.Context, id string) (*repository.Client, error)
	GetClientByEmail(ctx context.Context, email string) (*repository.Client, error)
	GetClientsByStatus(ctx context.Context, status string) ([]repository.Client, error)
	CreateClient(ctx context.Context, client *repository.Client) error
	UpdateClient(ctx context.Context, clientID string, patch repository.ClientPatch) error
}

// ClientUpdateInput is a partial update: any non-nil field is applied.
type ClientUpdateInput struct {
	FirstName   *string
	LastName    *string
	Email       *string
	InitialNote *repository.Note
	// NotesList replaces the entire notes list when non-nil (including empty slice to clear).
	NotesList           *[]repository.Note
	RequestedCounsellor *string
	Urgency             *string
	NextAppointment     *string
}

type ClientService struct {
	repo ClientRepository
}

func NewClientService(repo ClientRepository) *ClientService {
	return &ClientService{
		repo: repo,
	}
}

func (s *ClientService) GetClientList(ctx context.Context) ([]repository.Client, error) {
	clients, err := s.repo.GetClientList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client list: %w", err)
	}
	return clients, nil
}

func (s *ClientService) GetClientByID(ctx context.Context, id string) (*repository.Client, error) {
	client, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client by ID: %w", err)
	}
	return client, nil
}

func (s *ClientService) GetClientByEmail(ctx context.Context, email string) (*repository.Client, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, ErrMissingEmail
	}
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	// Prefer lowercase (matches normalized Create/Update); retry exact casing for legacy rows.
	lower := strings.ToLower(email)
	client, err := s.repo.GetClientByEmail(ctx, lower)
	if err != nil && strings.Contains(err.Error(), "not found") && lower != email {
		client, err = s.repo.GetClientByEmail(ctx, email)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client by email: %w", err)
	}
	return client, nil
}

func (s *ClientService) GetActiveClients(ctx context.Context) ([]repository.Client, error) {
	clients, err := s.repo.GetClientsByStatus(ctx, "active")
	if err != nil {
		return nil, fmt.Errorf("failed to get active clients: %w", err)
	}
	return clients, nil
}

func (s *ClientService) GetInactiveClients(ctx context.Context) ([]repository.Client, error) {
	clients, err := s.repo.GetClientsByStatus(ctx, "inactive")
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive clients: %w", err)
	}
	return clients, nil
}

func (s *ClientService) CreateClient(ctx context.Context, client *repository.Client) error {
	client.Email = strings.TrimSpace(strings.ToLower(client.Email))

	// Validate required fields
	if client.FirstName == "" || client.LastName == "" || client.Email == "" {
		return ErrMissingRequiredFields
	}

	// Validate email format
	if !emailRegex.MatchString(client.Email) {
		return ErrInvalidEmail
	}

	_, err := s.repo.GetClientByEmail(ctx, client.Email)
	if err == nil {
		return ErrEmailAlreadyExists
	}
	if !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to check existing client by email: %w", err)
	}

	// Set default status if not provided
	if client.Status == "" {
		client.Status = "active"
	}

	// Generate ID if not provided
	if client.ID == "" {
		client.ID = "client-" + uuid.New().String()
	}

	// Set timestamps
	now := time.Now().Format(time.RFC3339)
	if client.CreatedAt == "" {
		client.CreatedAt = now
	}
	client.UpdatedAt = now

	if err := s.repo.CreateClient(ctx, client); err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return nil
}

func (s *ClientService) UpdateClient(ctx context.Context, clientID string, in ClientUpdateInput) error {
	if clientID == "" {
		return ErrMissingClientID
	}
	hasInitialNote := in.InitialNote != nil && strings.TrimSpace(in.InitialNote.Note) != ""
	hasNotesList := in.NotesList != nil && len(*in.NotesList) > 0
	if in.FirstName == nil && in.LastName == nil && in.Email == nil && !hasInitialNote &&
		!hasNotesList &&
		in.RequestedCounsellor == nil && in.Urgency == nil && in.NextAppointment == nil {
		return ErrNoFieldsToUpdate
	}

	existing, err := s.repo.GetClientByID(ctx, clientID)
	if err != nil {
		return fmt.Errorf("failed to load client: %w", err)
	}

	patch := repository.ClientPatch{}

	if in.FirstName != nil {
		v := strings.TrimSpace(*in.FirstName)
		if v == "" {
			return ErrMissingRequiredFields
		}
		patch.FirstName = &v
	}
	if in.LastName != nil {
		v := strings.TrimSpace(*in.LastName)
		if v == "" {
			return ErrMissingRequiredFields
		}
		patch.LastName = &v
	}
	if in.Email != nil {
		em := strings.TrimSpace(strings.ToLower(*in.Email))
		if !emailRegex.MatchString(em) {
			return ErrInvalidEmail
		}
		patch.Email = &em
	}
	// Non-empty notes_list replaces the list. Empty slice is ignored so partial updates (e.g. only
	// counsellor/urgency) do not clear notes when the UI sends notes_list: [].
	if in.NotesList != nil && len(*in.NotesList) > 0 {
		patch.Notes = in.NotesList
	} else if in.InitialNote != nil && strings.TrimSpace(in.InitialNote.Note) != "" {
		note := *in.InitialNote
		if note.ClientID == "" {
			note.ClientID = clientID
		}
		if note.Date == "" {
			note.Date = time.Now().Format(time.RFC3339)
		}
		notes := make([]repository.Note, len(existing.Notes))
		copy(notes, existing.Notes)
		if len(notes) == 0 {
			notes = []repository.Note{note}
		} else {
			notes[0] = note
		}
		patch.Notes = &notes
	}
	if in.RequestedCounsellor != nil {
		v := strings.TrimSpace(*in.RequestedCounsellor)
		patch.RequestedCounsellor = &v
	}
	if in.Urgency != nil {
		v := strings.TrimSpace(*in.Urgency)
		patch.Urgency = &v
	}
	if in.NextAppointment != nil {
		v := strings.TrimSpace(*in.NextAppointment)
		patch.NextAppointment = &v
	}

	if err := s.repo.UpdateClient(ctx, clientID, patch); err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}
	return nil
}
