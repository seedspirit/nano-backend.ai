package projectserv

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/dto/response"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// defaultProjectSessionsLimit is the Phase 0 list size before cursor pagination.
const defaultProjectSessionsLimit = 20

type projectHandler struct {
	svc sessionService
}

type sessionService interface {
	ListProjectSessions(ctx context.Context, projectID uuid.UUID, limit int) ([]session.Session, error)
}

func newProjectHandler(args Args) (*projectHandler, error) {
	if args.Services == nil || args.Services.SessionSvc == nil {
		return nil, errordef.Errorf(errordef.InvalidInput, "session service is required")
	}
	return &projectHandler{svc: args.Services.SessionSvc}, nil
}

func (h *projectHandler) listSessions(c *echo.Context) error {
	ctx := c.Request().Context()
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, payload := errordef.Response(errordef.ErrInvalidInput, "check the project ID and retry", nil)
		return c.JSON(status, payload)
	}

	sessions, err := h.svc.ListProjectSessions(ctx, projectID, defaultProjectSessionsLimit)
	if err != nil {
		status, payload := errordef.Response(err, "retry later or contact an operator", nil)
		return c.JSON(status, payload)
	}

	data := response.ProjectSessionsData{
		Sessions: response.NewSessionSummaries(sessions),
		Limit:    defaultProjectSessionsLimit,
	}
	return c.JSON(http.StatusOK, response.OK(data))
}
