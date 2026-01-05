package service

import (
	"context"
	"fmt"

	"github.com/jmason/john_ai_project/internal/repository"
)

// ClientService handles business logic for client operations
type ClientService struct {
	repo *repository.ClientRepository
}

// NewClientService creates a new client service
func NewClientService(repo *repository.ClientRepository) *ClientService {
	return &ClientService{
		repo: repo,
	}
}

// GetClientList retrieves all clients
func (s *ClientService) GetClientList(ctx context.Context) ([]repository.Client, error) {
	clients, err := s.repo.GetClientList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client list: %w", err)
	}
	return clients, nil
}

// GetClientByID retrieves a single client by ID
func (s *ClientService) GetClientByID(ctx context.Context, id string) (*repository.Client, error) {
	client, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client by ID: %w", err)
	}
	return client, nil
}

// GetActiveClients retrieves all active clients
func (s *ClientService) GetActiveClients(ctx context.Context) ([]repository.Client, error) {
	clients, err := s.repo.GetClientsByStatus(ctx, "active")
	if err != nil {
		return nil, fmt.Errorf("failed to get active clients: %w", err)
	}
	return clients, nil
}

// GetInactiveClients retrieves all inactive clients
func (s *ClientService) GetInactiveClients(ctx context.Context) ([]repository.Client, error) {
	clients, err := s.repo.GetClientsByStatus(ctx, "inactive")
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive clients: %w", err)
	}
	return clients, nil
}

// Add a new client to the database
func (s *ClientService) AddClient(ctx context.Context, client *repository.Client) error {

	err := s.repo.AddClient(ctx, *client)
	if err != nil {
		return fmt.Errorf("failed to add new client: %w", err)
	}
	return nil
}
