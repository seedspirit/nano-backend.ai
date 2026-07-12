package db

import (
	"testing"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
)

func TestListPendingSessionsReturnsPendingSessionsWithDefinitions(t *testing.T) {
	fixture := newSessionRepositoryFixture(t)
	projectID := fixture.givenProject()
	oldestID := fixture.givenSession(projectID, "oldest", "2026-05-21T00:00:00Z")
	fixture.givenSession(projectID, "newer", "2026-05-21T00:01:00Z")
	runningID := fixture.givenSession(projectID, "running", "2026-05-21T00:02:00Z")
	if _, err := fixture.repo.db.ExecContext(fixture.ctx, `
		UPDATE sessions SET status = ? WHERE id = ?
	`, string(session.Running), runningID.String()); err != nil {
		t.Fatalf("mark session running: %v", err)
	}

	repo := &SchedulerRepository{db: fixture.repo.db}
	got, err := repo.ListPendingSessions(fixture.ctx)
	if err != nil {
		t.Fatalf("ListPendingSessions: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d sessions, want 2", len(got))
	}
	byID := make(map[string]session.Session, len(got))
	for _, item := range got {
		byID[item.ID.String()] = item
	}
	if byID[oldestID.String()].Name != "oldest" {
		t.Fatalf("oldest definition = %+v", byID[oldestID.String()].Definition)
	}
}

func TestListPendingSessionsReturnsEmptySliceWhenQueueIsEmpty(t *testing.T) {
	fixture := newSessionRepositoryFixture(t)
	repo := &SchedulerRepository{db: fixture.repo.db}

	got, err := repo.ListPendingSessions(fixture.ctx)
	if err != nil {
		t.Fatalf("ListPendingSessions: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("got sessions %+v, want empty slice", got)
	}
}
