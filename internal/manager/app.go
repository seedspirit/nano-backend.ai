package manager

import (
	"context"

	"github.com/seedspirit/nano-backend.ai/internal/manager/repository"
	"github.com/seedspirit/nano-backend.ai/internal/manager/servers"
	"github.com/seedspirit/nano-backend.ai/internal/manager/service"
	"github.com/seedspirit/nano-backend.ai/internal/manager/service/sessionsvc"
	"github.com/seedspirit/nano-backend.ai/internal/manager/sessionspec/preset"
	"github.com/seedspirit/nano-backend.ai/internal/manager/sessionspec/specbuilder"
	"github.com/seedspirit/nano-backend.ai/internal/manager/sessionspec/validator"
)

// Args configures the manager application.
type Args struct {
	Addr   string
	DBPath string
}

// App owns the manager runtime dependencies.
type App struct {
	repositories *repository.Repositories
	server       *servers.Server
}

// NewApp wires repositories, services, and servers for the manager process.
func NewApp(ctx context.Context, args Args) (*App, error) {
	repositories, err := repository.NewRepositories(ctx, repository.Args{
		DBPath: args.DBPath,
	})
	if err != nil {
		return nil, err
	}

	registry := preset.NewStaticRegistry(preset.Phase0Presets()...)
	builder := specbuilder.PresetBacked{
		Registry:  registry,
		Validator: validator.Noop{},
	}

	services := service.NewServices().WithSessionService(sessionsvc.Args{
		Repositories: repositories,
		SpecBuilder:  builder,
	})
	server, err := servers.NewServer(servers.ServerArgs{
		Addr:     args.Addr,
		Services: services,
	})
	if err != nil {
		_ = repositories.Close()
		return nil, err
	}

	return &App{
		repositories: repositories,
		server:       server,
	}, nil
}

// Start sessions the manager server.
func (a *App) Start(ctx context.Context) error {
	return a.server.Start(ctx)
}

// Close releases resources owned by the application.
func (a *App) Close() error {
	return a.repositories.Close()
}
