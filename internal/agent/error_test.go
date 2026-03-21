package agent

import (
	"errors"
	"fmt"
	"testing"
)

func TestStartupErrorWrapping(t *testing.T) {
	err := fmt.Errorf("bind :9090: %w", ErrStartup)

	if !errors.Is(err, ErrStartup) {
		t.Errorf("expected errors.Is(err, ErrStartup) to be true")
	}
}

func TestOperationErrorWrapping(t *testing.T) {
	err := fmt.Errorf("heartbeat failed: %w", ErrOperation)

	if !errors.Is(err, ErrOperation) {
		t.Errorf("expected errors.Is(err, ErrOperation) to be true")
	}
}
