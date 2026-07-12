package sessionsvc

import (
	"context"

	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/draft"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/spec"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
	"github.com/seedspirit/nano-backend.ai/internal/manager/repository"
)

// Args configures the session service.
type Args struct {
	Repositories *repository.Repositories
	SpecBuilder  SpecBuilder
}

// SessionRepository is the persistence dependency required by the session service.
type SessionRepository interface {
	GetSpec(ctx context.Context, id uuid.UUID) (spec.Spec, error)
	ListProjectSessions(ctx context.Context, projectID uuid.UUID, limit int) ([]session.Session, error)
	ProjectExists(ctx context.Context, projectID uuid.UUID) error
	CreateSession(ctx context.Context, target *session.Session) error
}

// SpecBuilder finalizes a submitted draft into an immutable spec.
type SpecBuilder interface {
	Build(ctx context.Context, d *draft.Draft) (spec.Spec, error)
}

// Service provides session use cases.
type Service struct {
	repo        SessionRepository
	specBuilder SpecBuilder
}

// NewService creates a session service.
func NewService(args Args) *Service {
	return &Service{
		repo:        args.Repositories.Session,
		specBuilder: args.SpecBuilder,
	}
}

// GetSpec returns the spec associated with a session ID.
func (s *Service) GetSpec(ctx context.Context, id uuid.UUID) (spec.Spec, error) {
	return s.repo.GetSpec(ctx, id)
}

// ListProjectSessions returns the most recent sessions associated with a project.
func (s *Service) ListProjectSessions(ctx context.Context, projectID uuid.UUID, limit int) ([]session.Session, error) {
	return s.repo.ListProjectSessions(ctx, projectID, limit)
}

// Submit validates a draft, finalizes a spec, and persists spec + pending session record.
func (s *Service) Submit(ctx context.Context, sessionDraft *draft.Draft) (session.Session, error) {
	if sessionDraft == nil {
		return session.Session{}, errordef.Errorf(errordef.InvalidInput, "draft is nil")
	}
	if err := s.repo.ProjectExists(ctx, sessionDraft.ProjectID); err != nil {
		return session.Session{}, err
	}

	built, err := s.specBuilder.Build(ctx, sessionDraft)
	if err != nil {
		return session.Session{}, err
	}

	target := session.NewPending(built)
	if err := s.repo.CreateSession(ctx, &target); err != nil {
		return session.Session{}, err
	}
	return target, nil
}
