package common

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestKernelIDEquality(t *testing.T) {
	a := KernelID("test-id-1")
	b := KernelID("test-id-1")
	c := KernelID("test-id-2")

	if a != b {
		t.Errorf("expected %q == %q", a, b)
	}
	if a == c {
		t.Errorf("expected %q != %q", a, c)
	}
}

func TestKernelIDDisplay(t *testing.T) {
	id := KernelID("abc-123")
	if id.String() != "abc-123" {
		t.Errorf("got %q, want %q", id.String(), "abc-123")
	}
}

func TestNewKernelIDGeneratesUniqueIDs(t *testing.T) {
	a := NewKernelID()
	b := NewKernelID()

	if a == b {
		t.Errorf("expected unique IDs, got %q and %q", a, b)
	}
}

func TestKernelSpecSerialize(t *testing.T) {
	spec := KernelSpec{Command: []string{"python", "-c", "print('hello')"}}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded KernelSpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(decoded.Command) != 3 {
		t.Fatalf("got %d commands, want 3", len(decoded.Command))
	}
	if decoded.Command[0] != "python" {
		t.Errorf("got command[0] %q, want %q", decoded.Command[0], "python")
	}
}

func TestKernelSpecDeserialize(t *testing.T) {
	input := `{"command":["echo","hello"]}`

	var spec KernelSpec
	if err := json.Unmarshal([]byte(input), &spec); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(spec.Command) != 2 {
		t.Fatalf("got %d commands, want 2", len(spec.Command))
	}
	if spec.Command[1] != "hello" {
		t.Errorf("got command[1] %q, want %q", spec.Command[1], "hello")
	}
}

func TestKernelStatusRunningRoundtrip(t *testing.T) {
	status := Running()

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded KernelStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if decoded.Type != StatusRunning {
		t.Errorf("got type %d, want %d (StatusRunning)", decoded.Type, StatusRunning)
	}
}

func TestKernelStatusExitedRoundtrip(t *testing.T) {
	status := Exited(42)

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded KernelStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if decoded.Type != StatusExited {
		t.Errorf("got type %d, want %d (StatusExited)", decoded.Type, StatusExited)
	}
	if decoded.Code != 42 {
		t.Errorf("got code %d, want 42", decoded.Code)
	}
}

func TestKernelStatusFailedRoundtrip(t *testing.T) {
	status := Failed("out of memory")

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded KernelStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if decoded.Type != StatusFailed {
		t.Errorf("got type %d, want %d (StatusFailed)", decoded.Type, StatusFailed)
	}
	if decoded.Reason != "out of memory" {
		t.Errorf("got reason %q, want %q", decoded.Reason, "out of memory")
	}
}

func TestKernelErrorNotFound(t *testing.T) {
	id := KernelID("missing-id")
	err := &KernelError{Op: "status", ID: id, Err: ErrKernelNotFound}

	if !errors.Is(err, ErrKernelNotFound) {
		t.Errorf("expected errors.Is(err, ErrKernelNotFound) to be true")
	}

	want := "kernel status missing-id: kernel not found"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestKernelErrorAlreadyExists(t *testing.T) {
	id := KernelID("dup-id")
	err := &KernelError{Op: "create", ID: id, Err: ErrKernelAlreadyExists}

	if !errors.Is(err, ErrKernelAlreadyExists) {
		t.Errorf("expected errors.Is(err, ErrKernelAlreadyExists) to be true")
	}

	want := "kernel create dup-id: kernel already exists"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestKernelErrorRuntime(t *testing.T) {
	err := &KernelError{Op: "create", Err: ErrKernelRuntime}

	if !errors.Is(err, ErrKernelRuntime) {
		t.Errorf("expected errors.Is(err, ErrKernelRuntime) to be true")
	}

	want := "kernel create: kernel runtime error"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}
