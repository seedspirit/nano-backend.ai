package projectserv

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/project"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/dto/request"
	"github.com/seedspirit/nano-backend.ai/internal/common/dto/response"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// defaultProjectSessionsLimit is the Phase 0 list size before cursor pagination.
const defaultProjectSessionsLimit = 20

type projectHandler struct {
	projects projectService
	sessions sessionService
}

type projectService interface {
	Create(ctx context.Context, name, description string) (project.Project, error)
	Get(ctx context.Context, id uuid.UUID) (project.Project, error)
}

type sessionService interface {
	ListProjectSessions(ctx context.Context, projectID uuid.UUID, limit int) ([]session.Session, error)
}

func newProjectHandler(args Args) (*projectHandler, error) {
	if args.Services == nil || args.Services.ProjectSvc == nil {
		return nil, errordef.Errorf(errordef.InvalidInput, "project service is required")
	}
	if args.Services.SessionSvc == nil {
		return nil, errordef.Errorf(errordef.InvalidInput, "session service is required")
	}
	return &projectHandler{projects: args.Services.ProjectSvc, sessions: args.Services.SessionSvc}, nil
}

func (h *projectHandler) create(c *echo.Context) error {
	ctx := c.Request().Context()
	var req request.CreateProjectReq
	if err := c.Bind(&req); err != nil {
		status, payload := errordef.Response(
			errordef.Errorf(errordef.ValidationError, "invalid request body: %s", err.Error()),
			"check the request body shape and retry",
			nil,
		)
		return c.JSON(status, payload)
	}

	created, err := h.projects.Create(ctx, req.Name, req.Description)
	if err != nil {
		status, payload := errordef.Response(err, projectErrorHint(err), nil)
		return c.JSON(status, payload)
	}
	return c.JSON(http.StatusCreated, response.Success(http.StatusCreated, created))
}

func (h *projectHandler) get(c *echo.Context) error {
	ctx := c.Request().Context()
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, payload := errordef.Response(errordef.ErrInvalidProjectID, "check the project ID and retry", nil)
		return c.JSON(status, payload)
	}

	value, err := h.projects.Get(ctx, projectID)
	if err != nil {
		status, payload := errordef.Response(err, "check the project ID or retry later", nil)
		return c.JSON(status, payload)
	}
	return c.JSON(http.StatusOK, response.OK(value))
}

func (h *projectHandler) listSessions(c *echo.Context) error {
	ctx := c.Request().Context()
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, payload := errordef.Response(errordef.ErrInvalidInput, "check the project ID and retry", nil)
		return c.JSON(status, payload)
	}

	sessions, err := h.sessions.ListProjectSessions(ctx, projectID, defaultProjectSessionsLimit)
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

func projectErrorHint(err error) string {
	if errors.Is(err, errordef.ErrValidation) {
		return "provide a non-empty project name and retry"
	}
	if errors.Is(err, errordef.ErrProjectNameConflict) {
		return "choose a different project name and retry"
	}
	return "retry later or contact an operator"
}
