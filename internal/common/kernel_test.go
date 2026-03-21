package common

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewKernelIDGeneratesUniqueIDs(t *testing.T) {
	a := NewKernelID()
	b := NewKernelID()

	if a.String() == b.String() {
		t.Errorf("expected unique IDs, got %q and %q", a, b)
	}
}

func TestNewKernelIDIsNotZero(t *testing.T) {
	id := NewKernelID()
	if id.IsZero() {
		t.Error("expected non-zero KernelID from NewKernelID()")
	}
}

func TestKernelIDZeroValue(t *testing.T) {
	var id KernelID
	if !id.IsZero() {
		t.Error("expected zero-value KernelID to be zero")
	}
}

func TestParseKernelIDValid(t *testing.T) {
	// Generate a valid UUID string via NewKernelID
	original := NewKernelID()
	parsed, err := ParseKernelID(original.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != original.String() {
		t.Errorf("got %q, want %q", parsed.String(), original.String())
	}
}

func TestParseKernelIDInvalid(t *testing.T) {
	_, err := ParseKernelID("not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID string")
	}
	if !errors.Is(err, ErrInvalidKernelID) {
		t.Errorf("expected ErrInvalidKernelID, got %v", err)
	}
}

func TestParseKernelIDEmpty(t *testing.T) {
	_, err := ParseKernelID("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
	if !errors.Is(err, ErrInvalidKernelID) {
		t.Errorf("expected ErrInvalidKernelID, got %v", err)
	}
}

func TestKernelIDEquality(t *testing.T) {
	id := NewKernelID()
	a, _ := ParseKernelID(id.String())
	b, _ := ParseKernelID(id.String())

	if a != b {
		t.Errorf("expected %q == %q", a, b)
	}

	c := NewKernelID()
	if a == c {
		t.Errorf("expected %q != %q", a, c)
	}
}

func TestKernelIDJSONRoundtrip(t *testing.T) {
	original := NewKernelID()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded KernelID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if decoded != original {
		t.Errorf("got %q, want %q", decoded, original)
	}
}

func TestKernelIDJSONInvalid(t *testing.T) {
	var id KernelID
	err := json.Unmarshal([]byte(`"not-a-uuid"`), &id)
	if err == nil {
		t.Fatal("expected error for invalid UUID in JSON")
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
		t.Errorf("got type %q, want %q", decoded.Type, StatusRunning)
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
		t.Errorf("got type %q, want %q", decoded.Type, StatusExited)
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
		t.Errorf("got type %q, want %q", decoded.Type, StatusFailed)
	}
	if decoded.Reason != "out of memory" {
		t.Errorf("got reason %q, want %q", decoded.Reason, "out of memory")
	}
}

func TestKernelErrorNotFound(t *testing.T) {
	id := NewKernelID()
	err := &KernelError{Op: "status", ID: id, Err: ErrKernelNotFound}

	if !errors.Is(err, ErrKernelNotFound) {
		t.Errorf("expected errors.Is(err, ErrKernelNotFound) to be true")
	}

	want := "kernel status " + id.String() + ": kernel not found"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestKernelErrorAlreadyExists(t *testing.T) {
	id := NewKernelID()
	err := &KernelError{Op: "create", ID: id, Err: ErrKernelAlreadyExists}

	if !errors.Is(err, ErrKernelAlreadyExists) {
		t.Errorf("expected errors.Is(err, ErrKernelAlreadyExists) to be true")
	}

	want := "kernel create " + id.String() + ": kernel already exists"
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

func TestKernelErrorWithZeroID(t *testing.T) {
	var zeroID KernelID
	err := &KernelError{Op: "destroy", ID: zeroID, Err: ErrKernelNotFound}

	// Zero ID should be omitted from error message
	want := "kernel destroy: kernel not found"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}
