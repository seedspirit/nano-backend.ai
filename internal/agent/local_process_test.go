package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/seedspirit/nano-backend.ai/internal/common"
)

func TestCreateSuccess(t *testing.T) {
	lp := NewLocalProcess()
	ctx := context.Background()

	id, err := lp.Create(ctx, common.KernelSpec{Command: []string{"sleep", "3600"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.IsZero() {
		t.Error("expected non-zero KernelID")
	}

	// cleanup
	_ = lp.Destroy(ctx, id)
}

func TestCreateReturnsUniqueIDs(t *testing.T) {
	lp := NewLocalProcess()
	ctx := context.Background()

	id1, err := lp.Create(ctx, common.KernelSpec{Command: []string{"sleep", "3600"}})
	if err != nil {
		t.Fatalf("unexpected error on first create: %v", err)
	}
	id2, err := lp.Create(ctx, common.KernelSpec{Command: []string{"sleep", "3600"}})
	if err != nil {
		t.Fatalf("unexpected error on second create: %v", err)
	}

	if id1.String() == id2.String() {
		t.Errorf("expected unique IDs, got %q and %q", id1, id2)
	}

	// cleanup
	_ = lp.Destroy(ctx, id1)
	_ = lp.Destroy(ctx, id2)
}

func TestCreateEmptyCommand(t *testing.T) {
	lp := NewLocalProcess()
	ctx := context.Background()

	_, err := lp.Create(ctx, common.KernelSpec{Command: []string{}})
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if !errors.Is(err, common.ErrKernelRuntime) {
		t.Errorf("expected ErrKernelRuntime, got %v", err)
	}

	var ke *common.KernelError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *KernelError, got %T", err)
	}
	if ke.Op != "create" {
		t.Errorf("expected op %q, got %q", "create", ke.Op)
	}
}

func TestCreateNonexistentBinary(t *testing.T) {
	lp := NewLocalProcess()
	ctx := context.Background()

	_, err := lp.Create(ctx, common.KernelSpec{Command: []string{"nonexistent-binary-xyz-999"}})
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
	if !errors.Is(err, common.ErrKernelRuntime) {
		t.Errorf("expected ErrKernelRuntime, got %v", err)
	}
}

func TestDestroySuccess(t *testing.T) {
	lp := NewLocalProcess()
	ctx := context.Background()

	id, err := lp.Create(ctx, common.KernelSpec{Command: []string{"sleep", "3600"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := lp.Destroy(ctx, id); err != nil {
		t.Fatalf("unexpected error on destroy: %v", err)
	}

	// Give the wait goroutine a moment to complete
	time.Sleep(50 * time.Millisecond)
}

func TestDestroyNotFound(t *testing.T) {
	lp := NewLocalProcess()
	ctx := context.Background()

	fakeID := common.NewKernelID()
	err := lp.Destroy(ctx, fakeID)
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
	if !errors.Is(err, common.ErrKernelNotFound) {
		t.Errorf("expected ErrKernelNotFound, got %v", err)
	}

	var ke *common.KernelError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *KernelError, got %T", err)
	}
	if ke.Op != "destroy" {
		t.Errorf("expected op %q, got %q", "destroy", ke.Op)
	}
	if ke.ID.IsZero() {
		t.Error("expected non-zero ID in error")
	}
}

func TestDestroyTwice(t *testing.T) {
	lp := NewLocalProcess()
	ctx := context.Background()

	id, err := lp.Create(ctx, common.KernelSpec{Command: []string{"sleep", "3600"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := lp.Destroy(ctx, id); err != nil {
		t.Fatalf("unexpected error on first destroy: %v", err)
	}

	err = lp.Destroy(ctx, id)
	if !errors.Is(err, common.ErrKernelNotFound) {
		t.Errorf("expected ErrKernelNotFound on second destroy, got %v", err)
	}
}
