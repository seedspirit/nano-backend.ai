package service

import (
	"github.com/seedspirit/nano-backend.ai/internal/manager/service/projectsvc"
	"github.com/seedspirit/nano-backend.ai/internal/manager/service/sessionsvc"
)

// Services groups manager service dependencies for handlers.
type Services struct {
	ProjectSvc *projectsvc.Service
	SessionSvc *sessionsvc.Service
}

// NewServices creates an empty service registry.
func NewServices() *Services {
	return &Services{}
}

// WithProjectService registers the project service.
func (s *Services) WithProjectService(args projectsvc.Args) *Services {
	s.ProjectSvc = projectsvc.NewService(args)
	return s
}

// WithSessionService registers the session service.
func (s *Services) WithSessionService(args sessionsvc.Args) *Services {
	s.SessionSvc = sessionsvc.NewService(args)
	return s
}
