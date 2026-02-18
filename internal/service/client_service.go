package service

import (
	"context"
	"fmt"

	"github.com/jmason/john_ai_project/internal/repository"
)

//TODO:tests for this service

// ClientRepository interface for dependency injection
type ClientRepository interface {
	GetClientList(ctx context.Context) ([]repository.Client, error)
	GetClientByID(ctx context.Context, id string) (*repository.Client, error)
	GetClientsByStatus(ctx context.Context, status string) ([]repository.Client, error)
	CreateClient(ctx context.Context, client *repository.Client) error
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
	if err := s.repo.CreateClient(ctx, client); err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return nil
}
