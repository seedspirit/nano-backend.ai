package scheduler

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	managerdb "github.com/seedspirit/nano-backend.ai/internal/manager/repository/db"
)

func TestListPendingSessionsReturnsPendingSessionsWithDefinitions(t *testing.T) {
	fixture := newSchedulerFixture(t)
	projectID := fixture.givenProject()
	firstID := fixture.givenSession(projectID, "first", session.Pending)
	fixture.givenSession(projectID, "second", session.Pending)
	fixture.givenSession(projectID, "running", session.Running)

	got, err := fixture.repo.ListPendingSessions(fixture.ctx)
	if err != nil {
		t.Fatalf("ListPendingSessions: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d sessions, want 2", len(got))
	}
	byID := make(map[uuid.UUID]session.Session, len(got))
	for _, item := range got {
		byID[item.ID] = item
	}
	if byID[firstID].Name != "first" {
		t.Fatalf("first definition = %+v", byID[firstID].Definition)
	}
}

func TestListPendingSessionsReturnsEmptySliceWhenQueueIsEmpty(t *testing.T) {
	fixture := newSchedulerFixture(t)

	got, err := fixture.repo.ListPendingSessions(fixture.ctx)
	if err != nil {
		t.Fatalf("ListPendingSessions: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("got sessions %+v, want empty slice", got)
	}
}

type schedulerFixture struct {
	t      *testing.T
	ctx    context.Context
	db     *sqlx.DB
	repo   *Repository
	serial int
}

func newSchedulerFixture(t *testing.T) *schedulerFixture {
	t.Helper()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "manager.db")
	dbx, err := managerdb.Open(ctx, managerdb.Args{DBPath: dbPath})
	if err != nil {
		t.Fatalf("open fixture db: %v", err)
	}
	repo, err := NewRepository(ctx, Args{DBPath: dbPath})
	if err != nil {
		_ = dbx.Close()
		t.Fatalf("new scheduler repository: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Close(); err != nil {
			t.Errorf("close scheduler repository: %v", err)
		}
		if err := dbx.Close(); err != nil {
			t.Errorf("close fixture db: %v", err)
		}
	})
	return &schedulerFixture{t: t, ctx: ctx, db: dbx, repo: repo}
}

func (f *schedulerFixture) givenProject() uuid.UUID {
	f.t.Helper()
	id := uuid.New()
	if _, err := f.db.ExecContext(f.ctx, `
		INSERT INTO projects (id, name, created_at)
		VALUES (?, ?, ?)
	`, id.String(), "scheduler-project", "2026-07-12T00:00:00Z"); err != nil {
		f.t.Fatalf("insert project: %v", err)
	}
	return id
}

func (f *schedulerFixture) givenSession(projectID uuid.UUID, name string, status session.Status) uuid.UUID {
	f.t.Helper()
	f.serial++
	id := uuid.New()
	if _, err := f.db.ExecContext(f.ctx, `
		INSERT INTO sessions (
			id, project_id, type, name, model_base_model,
			resource_gpu_count, resource_memory_limit_bytes,
			resource_timeout_duration_seconds, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id.String(), projectID.String(), string(session.Batch), name, "meta-llama/Llama-3-8B",
		1, int64(1<<30), int64(3600), string(status),
		fmt.Sprintf("2026-07-12T00:00:%02dZ", f.serial)); err != nil {
		f.t.Fatalf("insert session: %v", err)
	}
	return id
}
