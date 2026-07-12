package repository

import (
	"context"

	"github.com/seedspirit/nano-backend.ai/internal/manager/repository/db"
)

// Args configures repository construction.
type Args struct {
	DBPath string
}

// Repositories groups concrete manager repository instances.
type Repositories struct {
	Session *db.SessionRepository
}

// NewRepositories opens and migrates the configured persistence backends.
func NewRepositories(ctx context.Context, args Args) (*Repositories, error) {
	sessionRepo, err := db.NewSessionRepository(ctx, db.Args{
		DBPath: args.DBPath,
	})
	if err != nil {
		return nil, err
	}

	return &Repositories{
		Session: sessionRepo,
	}, nil
}

// Close releases resources owned by the repositories.
func (r *Repositories) Close() error {
	return r.Session.Close()
}
