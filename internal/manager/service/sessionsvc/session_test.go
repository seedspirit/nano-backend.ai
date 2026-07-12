package sessionsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/aggregate"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/draft"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/spec"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

type stubBuilder struct {
	spec spec.Spec
	err  error
}

func (b *stubBuilder) Build(ctx context.Context, d *draft.Draft) (spec.Spec, error) {
	if b.err != nil {
		return spec.Spec{}, b.err
	}
	return b.spec, nil
}

type stubRunRepo struct {
	projectExistsErr error
	createRunErr     error
	created          bool
	spec             *spec.Spec
}

func (r *stubRunRepo) ProjectExists(ctx context.Context, projectID uuid.UUID) error {
	return r.projectExistsErr
}

func (r *stubRunRepo) CreateSession(ctx context.Context, target *aggregate.Session) error {
	r.created = true
	r.spec = &target.Definition
	return r.createRunErr
}

func (r *stubRunRepo) GetSpec(ctx context.Context, id uuid.UUID) (spec.Spec, error) {
	return spec.Spec{}, errordef.ErrNotFound
}

func (r *stubRunRepo) ListProjectSessions(ctx context.Context, projectID uuid.UUID, limit int) ([]session.Session, error) {
	return nil, nil
}

func TestSubmitReturnsPendingSessionOnSuccess(t *testing.T) {
	projectID := uuid.New()
	specID := uuid.New()
	builder := &stubBuilder{spec: spec.Spec{ID: specID, ProjectID: projectID}}
	repo := &stubRunRepo{}
	svc := &Service{repo: repo, specBuilder: builder}

	d := draft.Draft{ID: uuid.New(), ProjectID: projectID}
	got, err := svc.Submit(context.Background(), &d)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if repo.spec == nil || repo.spec.ID != got.ID || got.ID != specID {
		t.Fatalf("persisted definition id = %v; want session and builder id %s", repo.spec, specID)
	}
	if got.Lifecycle.Status != session.Pending {
		t.Fatalf("got status %s, want pending", got.Lifecycle.Status)
	}
	if !repo.created {
		t.Fatal("repo.CreateSession not called")
	}
}

func TestSubmitReturnsNotFoundForMissingProject(t *testing.T) {
	builder := &stubBuilder{}
	repo := &stubRunRepo{projectExistsErr: errordef.ErrNotFound}
	svc := &Service{repo: repo, specBuilder: builder}

	_, err := svc.Submit(context.Background(), &draft.Draft{ID: uuid.New(), ProjectID: uuid.New()})
	if !errors.Is(err, errordef.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
