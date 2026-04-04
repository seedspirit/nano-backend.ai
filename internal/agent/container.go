package agent

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

// ContainerClient abstracts Docker container operations for testability.
// Implementations wrap a real Docker client; tests use a mock.
type ContainerClient interface {
	CreateContainer(ctx context.Context, config *container.Config, name string) (container.CreateResponse, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout *int) error
	RemoveContainer(ctx context.Context, containerID string) error
	InspectContainer(ctx context.Context, containerID string) (container.InspectResponse, error)
}
