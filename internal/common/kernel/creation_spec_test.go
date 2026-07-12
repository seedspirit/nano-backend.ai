package kernel

import (
	"testing"

	"github.com/google/uuid"

	"github.com/seedspirit/nano-backend.ai/internal/common/agent"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
)

func mustImage(t *testing.T, s string) ImageRef {
	t.Helper()
	ref, err := ParseImageRef(s)
	if err != nil {
		t.Fatalf("ParseImageRef(%q): %v", s, err)
	}
	return ref
}

func mustPath(t *testing.T, s string) AgentPath {
	t.Helper()
	p, err := NewAgentPath(s)
	if err != nil {
		t.Fatalf("NewAgentPath(%q): %v", s, err)
	}
	return p
}

// validParts returns a complete, valid kernel creation spec for tests.
func validParts(t *testing.T) CreationSpecArgs {
	t.Helper()
	id, err := NewIdentifiers(uuid.New(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("NewIdentifiers: %v", err)
	}
	ex, err := NewExecution(&ExecutionArgs{Image: mustImage(t, "ghcr.io/org/img:v1")})
	if err != nil {
		t.Fatalf("NewExecution: %v", err)
	}
	res := NewResourceSlots(session.CPUOptions{Cores: 4}, session.MemoryOptions{LimitBytes: 1 << 30})
	as, err := NewAllocation(agent.NewID(), []GPUIndex{0})
	if err != nil {
		t.Fatalf("NewAllocation: %v", err)
	}
	io := NewIOBindings(&IOArgs{ArtifactPath: mustPath(t, "/agent/artifacts")})
	return CreationSpecArgs{Identifiers: id, Execution: ex, ResourceSlots: res, Allocation: as, IO: io}
}

func TestNewCreationSpecValid(t *testing.T) {
	args := validParts(t)
	if _, err := NewCreationSpec(&args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCreationSpecRejectsMissingParts(t *testing.T) {
	base := validParts(t)

	tests := []struct {
		name   string
		mutate func(*CreationSpecArgs)
	}{
		{name: "zero identifiers", mutate: func(a *CreationSpecArgs) { a.Identifiers = Identifiers{} }},
		{name: "zero execution image", mutate: func(a *CreationSpecArgs) { a.Execution = Execution{} }},
		{name: "empty agent id", mutate: func(a *CreationSpecArgs) { a.Allocation = Allocation{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := base
			tt.mutate(&args)
			if _, err := NewCreationSpec(&args); err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}

func TestNewExecutionCopiesInputs(t *testing.T) {
	entrypoint := []string{"python"}
	command := []string{"train.py"}
	env := map[string]string{"BASE_MODEL": "llama"}

	ex, err := NewExecution(&ExecutionArgs{
		Image:      mustImage(t, "img:v1"),
		Entrypoint: entrypoint,
		Command:    command,
		Env:        env,
	})
	if err != nil {
		t.Fatalf("NewExecution: %v", err)
	}

	// Mutate the caller's inputs after construction.
	entrypoint[0] = "MUTATED"
	command[0] = "MUTATED"
	env["BASE_MODEL"] = "MUTATED"
	env["EXTRA"] = "added"

	if ex.Entrypoint[0] != "python" {
		t.Errorf("entrypoint leaked mutation: %v", ex.Entrypoint)
	}
	if ex.Command[0] != "train.py" {
		t.Errorf("command leaked mutation: %v", ex.Command)
	}
	if ex.Env["BASE_MODEL"] != "llama" {
		t.Errorf("env value leaked mutation: %v", ex.Env)
	}
	if _, ok := ex.Env["EXTRA"]; ok {
		t.Errorf("env gained a key from caller mutation: %v", ex.Env)
	}
}

func TestNewAllocationCopiesGPUs(t *testing.T) {
	gpus := []GPUIndex{0, 1}
	allocation, err := NewAllocation(agent.NewID(), gpus)
	if err != nil {
		t.Fatalf("NewAllocation: %v", err)
	}

	gpus[0] = 99

	if allocation.GPUs[0] != 0 {
		t.Errorf("GPUs leaked mutation: %v", allocation.GPUs)
	}
}

func TestNewIOBindingsCopiesMounts(t *testing.T) {
	mounts := []Mount{{Source: mustPath(t, "/data"), Target: "/mnt/data"}}
	io := NewIOBindings(&IOArgs{Mounts: mounts})

	mounts[0].Target = "MUTATED"

	if io.Mounts[0].Target != "/mnt/data" {
		t.Errorf("mounts leaked mutation: %v", io.Mounts)
	}
}
