package service

import (
	"context"

	"poker-club/backend/internal/domain"
)

// Service provides business logic operations.
// It depends on repositories for data persistence.
type Service struct {
	repos *domain.Repositories
}

// New creates a new Service instance.
func New(repos *domain.Repositories) *Service {
	return &Service{
		repos: repos,
	}
}

// HealthCheck verifies that all dependencies are accessible.
func (s *Service) HealthCheck(ctx context.Context) error {
	return s.repos.Clubs.Ping(ctx)
}
