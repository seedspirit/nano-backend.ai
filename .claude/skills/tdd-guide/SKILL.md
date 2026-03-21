---
name: tdd-guide
description: Test-Driven Development workflow for Go — Red-Green-Refactor cycle, scenario definition, test patterns
user-invocable: true
---

# TDD Workflow Guide

Test-Driven Development workflow for nano-backend.ai (Go).

## TDD Cycle

```
1. Define Scenarios → 2. Write Tests → 3. Verify Failure → 4. Implement → 5. Pass → 6. Refactor
                                                                            ↑_____________|
```

**Key Principle: Red → Green → Refactor**

## Step 1: Define Test Scenarios

Before writing any code, document success and error cases:

```markdown
## Test Target: {Feature Name}

### Success Scenarios
1. {Primary success case}
2. {Edge case that should succeed}

### Error Scenarios
1. {Invalid input} → Expected: {ErrorType}
2. {Resource not found} → Expected: {ErrorType}
3. {Boundary condition} → Expected: {behavior}
```

### Example

```markdown
## Test Target: Job Submission

### Success Scenarios
1. Valid job with available agent → Returns JobID, status=Pending
2. Job with optional fields omitted → Uses defaults, returns JobID

### Error Scenarios
1. Unknown executor type → ErrInvalidExecutor
2. No agents available → ErrNoAgentAvailable
3. Duplicate job ID → ErrAlreadyExists
```

## Step 2: Write Failing Tests

Write tests BEFORE implementation.

### Unit Test Pattern (same package)

```go
package scheduler

import "testing"

func TestSubmitValidJobReturnsID(t *testing.T) {
    sched := NewJobScheduler( /* test deps */ )
    request := SubmitRequest{
        Executor: "python",
    }

    job, err := sched.Submit(request)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if job.Status != StatusPending {
        t.Errorf("got status %v, want %v", job.Status, StatusPending)
    }
}

func TestSubmitUnknownExecutorReturnsError(t *testing.T) {
    sched := NewJobScheduler( /* test deps */ )
    request := SubmitRequest{
        Executor: "nonexistent",
    }

    _, err := sched.Submit(request)

    if !errors.Is(err, ErrInvalidExecutor) {
        t.Errorf("got error %v, want ErrInvalidExecutor", err)
    }
}
```

### Table-Driven Test Pattern

```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Config
        wantErr error
    }{
        {
            name:  "valid config",
            input: `{"port": 8080}`,
            want:  Config{Port: 8080},
        },
        {
            name:    "invalid JSON",
            input:   "bad data",
            wantErr: ErrParseFailed,
        },
        {
            name:    "missing required field",
            input:   `{}`,
            wantErr: ErrValidation,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseConfig(tt.input)
            if tt.wantErr != nil {
                if !errors.Is(err, tt.wantErr) {
                    t.Errorf("got error %v, want %v", err, tt.wantErr)
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.want {
                t.Errorf("got %+v, want %+v", got, tt.want)
            }
        })
    }
}
```

### HTTP Handler Test Pattern

```go
package manager_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHealthEndpoint(t *testing.T) {
    handler := NewRouter()
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
    }
}
```

### Integration Test Pattern (`_test.go` with build tag)

```go
//go:build integration

package integration

import "testing"

func TestJobCompletesEndToEnd(t *testing.T) {
    app := SpawnTestApp(t)

    jobID, err := app.SubmitJob(testRequest())
    if err != nil {
        t.Fatalf("submit failed: %v", err)
    }

    status, err := app.PollUntilDone(jobID)
    if err != nil {
        t.Fatalf("poll failed: %v", err)
    }
    if status != StatusCompleted {
        t.Errorf("got status %v, want Completed", status)
    }
}
```

## Step 3: Verify Failure (Red)

Run tests and confirm they fail for the **right reason**:

```bash
go test ./internal/<package>/...
```

**Expected failures:**
- Compile error — type/function doesn't exist yet (good)
- Assertion failure — logic not implemented (good)

**If tests pass unexpectedly:** verify the test actually exercises the target code.

## Step 4: Implement Minimum Code (Green)

Write the **simplest code** that makes tests pass:

- Each function: single purpose, < 30 lines
- Proper `error` returns with domain error types
- Full type annotations
- No features beyond what tests cover

## Step 5: Run Tests

```bash
# Run all tests for a package
go test ./internal/<package>/...

# Run specific test
go test ./internal/<package>/... -run TestName

# Run with verbose output
go test ./internal/<package>/... -v
```

All tests must pass. If they fail, fix the implementation — do NOT skip tests.

## Step 6: Refactor

Improve code while keeping tests green:

- [ ] Extract functions > 30 lines
- [ ] Remove duplication
- [ ] Improve naming
- [ ] Simplify complex conditionals

### Quality Checks (mandatory after refactor)

```bash
gofmt -w .
golangci-lint run ./...
go test ./...
```

**Fix all errors — never suppress with `//nolint`.**

## Go-Specific Test Patterns

### Test Helpers

```go
func newTestJob(t *testing.T, executor string) SubmitRequest {
    t.Helper()
    return SubmitRequest{
        Executor: executor,
        Timeout:  5 * time.Second,
    }
}
```

### Testing Error Types

```go
func TestRejectsInvalidInput(t *testing.T) {
    _, err := parseConfig("bad data")

    var configErr *ConfigError
    if !errors.As(err, &configErr) {
        t.Fatalf("got %T, want *ConfigError", err)
    }
    if configErr.Op != "parse" {
        t.Errorf("got op %q, want %q", configErr.Op, "parse")
    }
}
```

### Testing with Interfaces (Dependency Injection)

```go
type mockAgentPool struct{}

func (m *mockAgentPool) Acquire(ctx context.Context) (AgentHandle, error) {
    return AgentHandle{ID: "fake-agent"}, nil
}

func TestSchedulerWithMockPool(t *testing.T) {
    sched := NewJobScheduler(&mockAgentPool{})
    // ...
}
```

### Cleanup with t.Cleanup

```go
func TestWithTempFile(t *testing.T) {
    f, err := os.CreateTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { os.Remove(f.Name()) })

    // use f...
}
```

## Common TDD Mistakes

### 1. Writing Implementation First
- **Wrong:** Implement → Write tests to verify
- **Correct:** Write tests → Implement to pass

### 2. Testing Implementation Details
- **Wrong:** Assert internal method calls, private state
- **Correct:** Test public behavior and outcomes

### 3. Large Cycles
- **Wrong:** Write 20 tests → Implement everything
- **Correct:** One test → Small implementation → Repeat

### 4. Skipping Red Phase
- **Wrong:** Write test → Implement → Run (pass)
- **Correct:** Write test → Run (fail) → Implement → Run (pass)

### 5. Suppressing Quality Errors
- **Wrong:** Add `//nolint` directives
- **Correct:** Fix root cause

## Summary

1. **Define scenarios** — success + error cases
2. **Write failing tests** — `func TestXxx(t *testing.T)`
3. **Verify failure** — `go test` red
4. **Implement minimum** — typed, explicit errors
5. **Pass tests** — `go test` green
6. **Refactor** — `gofmt` + `golangci-lint` + keep green

**Remember: Red → Green → Refactor → Repeat**
