package service

import (
	"github.com/seedspirit/nano-backend.ai/internal/manager/service/sessionsvc"
)

// Services groups manager service dependencies for handlers.
type Services struct {
	SessionSvc *sessionsvc.Service
}

// NewServices creates an empty service registry.
func NewServices() *Services {
	return &Services{}
}

// WithSessionService registers the session service.
func (s *Services) WithSessionService(args sessionsvc.Args) *Services {
	s.SessionSvc = sessionsvc.NewService(args)
	return s
}
