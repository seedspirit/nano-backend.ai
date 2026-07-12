package sessionserv

import (
	"github.com/labstack/echo/v5"
	"github.com/seedspirit/nano-backend.ai/internal/manager/service"
)

// Args configures the session API routes.
type Args struct {
	Services *service.Services
}

// WithSubServer registers session API routes below the given group.
func WithSubServer(g *echo.Group, args Args) error {
	handler, err := newSessionHandler(args)
	if err != nil {
		return err
	}
	sessionGroup := g.Group("/sessions")
	sessionGroup.POST("", handler.submit)
	sessionGroup.GET("/:id/spec", handler.getSpec)
	return nil
}
