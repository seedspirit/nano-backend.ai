package sessionserv

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/draft"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/spec"
	"github.com/seedspirit/nano-backend.ai/internal/common/dto/request"
	"github.com/seedspirit/nano-backend.ai/internal/common/dto/response"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

type sessionHandler struct {
	svc sessionService
}

type sessionService interface {
	GetSpec(ctx context.Context, sessionID uuid.UUID) (spec.Spec, error)
	Submit(ctx context.Context, sessionDraft *draft.Draft) (session.Session, error)
}

func newSessionHandler(args Args) (*sessionHandler, error) {
	if args.Services == nil || args.Services.SessionSvc == nil {
		return nil, errordef.Errorf(errordef.InvalidInput, "session service is required")
	}
	return &sessionHandler{svc: args.Services.SessionSvc}, nil
}

func (h *sessionHandler) getSpec(c *echo.Context) error {
	ctx := c.Request().Context()
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, payload := errordef.Response(errordef.ErrInvalidSessionID, "check the session ID and retry", nil)
		return c.JSON(status, payload)
	}
	sessionSpec, err := h.svc.GetSpec(ctx, sessionID)
	if err != nil {
		status, payload := errordef.Response(err, "retry later or contact an operator", nil)
		return c.JSON(status, payload)
	}
	return c.JSON(http.StatusOK, response.OK(sessionSpec))
}

func (h *sessionHandler) submit(c *echo.Context) error {
	ctx := c.Request().Context()
	var req request.SessionSpecDraftReq
	if err := c.Bind(&req); err != nil {
		status, payload := errordef.Response(
			errordef.Errorf(errordef.ValidationError, "invalid request body: %s", err.Error()),
			"check the request body shape and retry",
			nil,
		)
		return c.JSON(status, payload)
	}

	sessionDraft := req.ToDraft(uuid.New())

	created, err := h.svc.Submit(ctx, &sessionDraft)
	if err != nil {
		hint := "retry later or contact an operator"
		if errors.Is(err, errordef.ErrNotFound) {
			hint = "check project_id and retry"
		}
		status, payload := errordef.Response(err, hint, nil)
		return c.JSON(status, payload)
	}

	summary := response.NewSessionSummary(&created)
	return c.JSON(http.StatusCreated, response.Success(http.StatusCreated, summary))
}
