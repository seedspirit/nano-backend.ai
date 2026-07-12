package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"modernc.org/sqlite"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/project"
	"github.com/seedspirit/nano-backend.ai/internal/common/encoding"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// ProjectRepository persists and retrieves projects in SQLite.
type ProjectRepository struct {
	db *sqlx.DB
}

// NewProjectRepository opens, migrates, and returns a SQLite project repository.
func NewProjectRepository(ctx context.Context, args Args) (*ProjectRepository, error) {
	dbx, err := Open(ctx, args)
	if err != nil {
		return nil, err
	}
	return &ProjectRepository{db: dbx}, nil
}

// Close releases the repository database handle.
func (r *ProjectRepository) Close() error {
	return r.db.Close()
}

// Create inserts a new project.
func (r *ProjectRepository) Create(ctx context.Context, value *project.Project) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, created_at)
		VALUES (?, ?, ?, ?)
	`, value.ID.String(), value.Name, value.Description, encoding.FormatTime(value.CreatedAt))
	if err == nil {
		return nil
	}

	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && strings.Contains(sqliteErr.Error(), "projects.name") {
		return errordef.ErrProjectNameConflict
	}
	return fmt.Errorf("insert project %s: %w", value.ID, err)
}

// Get returns a project by ID.
func (r *ProjectRepository) Get(ctx context.Context, id uuid.UUID) (project.Project, error) {
	var row struct {
		ID          string `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		CreatedAt   string `db:"created_at"`
	}
	if err := r.db.GetContext(ctx, &row, `
		SELECT id, name, description, created_at
		FROM projects
		WHERE id = ?
	`, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return project.Project{}, errordef.ErrNotFound
		}
		return project.Project{}, fmt.Errorf("get project %s: %w", id, err)
	}

	projectID, err := uuid.Parse(row.ID)
	if err != nil {
		return project.Project{}, fmt.Errorf("parse project id %q: %w", row.ID, err)
	}
	createdAt, err := encoding.ParseTime(row.CreatedAt)
	if err != nil {
		return project.Project{}, err
	}
	return project.Project{
		ID:          projectID,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   createdAt,
	}, nil
}
