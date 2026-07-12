package projectsvc

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/project"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
	"github.com/seedspirit/nano-backend.ai/internal/manager/repository"
)

// Args configures the project service.
type Args struct {
	Repositories *repository.Repositories
}

// ProjectRepository is the persistence dependency required by the project service.
type ProjectRepository interface {
	Create(ctx context.Context, value *project.Project) error
	Get(ctx context.Context, id uuid.UUID) (project.Project, error)
}

// Service provides project use cases.
type Service struct {
	repo ProjectRepository
}

// NewService creates a project service.
func NewService(args Args) *Service {
	return &Service{repo: args.Repositories.Project}
}

// Create validates and persists a new project.
func (s *Service) Create(ctx context.Context, name, description string) (project.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return project.Project{}, errordef.Errorf(errordef.ValidationError, "project name is required")
	}
	created := project.New(name, description)
	if err := s.repo.Create(ctx, &created); err != nil {
		return project.Project{}, err
	}
	return created, nil
}

// Get returns a project by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (project.Project, error) {
	return s.repo.Get(ctx, id)
}
