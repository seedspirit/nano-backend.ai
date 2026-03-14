---
name: tdd-guide
description: Test-Driven Development workflow for Rust — Red-Green-Refactor cycle, scenario definition, test patterns
user-invocable: true
---

# TDD Workflow Guide

Test-Driven Development workflow for nano-backend.ai (Rust).

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
1. Valid job with available agent → Returns JobId, status=Pending
2. Job with optional fields omitted → Uses defaults, returns JobId

### Error Scenarios
1. Unknown executor type → JobError::InvalidExecutor
2. No agents available → JobError::NoAgentAvailable
3. Duplicate job ID → JobError::AlreadyExists
```

## Step 2: Write Failing Tests

Write tests BEFORE implementation.

### Unit Test Pattern (in-module)

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn submit_valid_job_returns_id() {
        let scheduler = JobScheduler::new(/* test deps */);
        let request = SubmitRequest::builder()
            .executor("python")
            .build();

        let result = scheduler.submit(request);

        assert!(result.is_ok());
        let job = result.unwrap();
        assert_eq!(job.status, JobStatus::Pending);
    }

    #[test]
    fn submit_unknown_executor_returns_error() {
        let scheduler = JobScheduler::new(/* test deps */);
        let request = SubmitRequest::builder()
            .executor("nonexistent")
            .build();

        let result = scheduler.submit(request);

        assert!(matches!(result, Err(JobError::InvalidExecutor(_))));
    }
}
```

### Async Test Pattern

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn agent_registers_successfully() {
        let registry = AgentRegistry::new_for_test().await;

        let result = registry.register("agent-1", "127.0.0.1:9000").await;

        assert!(result.is_ok());
    }
}
```

### Integration Test Pattern (`tests/` directory)

```rust
// tests/job_lifecycle.rs
use nano_backend::prelude::*;

#[tokio::test]
async fn job_completes_end_to_end() {
    let app = TestApp::spawn().await;

    let job_id = app.submit_job(test_request()).await.unwrap();
    let status = app.poll_until_done(job_id).await.unwrap();

    assert_eq!(status, JobStatus::Completed);
}
```

## Step 3: Verify Failure (Red)

Run tests and confirm they fail for the **right reason**:

```bash
cargo test -p <crate> -- <test_name>
```

**Expected failures:**
- Compile error — type/function doesn't exist yet (good)
- Assertion failure — logic not implemented (good)

**If tests pass unexpectedly:** verify the test actually exercises the target code.

## Step 4: Implement Minimum Code (Green)

Write the **simplest code** that makes tests pass:

- Each function: single purpose, < 30 lines
- Proper `Result<T, E>` returns with domain error types
- Full type annotations — no `Any` equivalents
- No features beyond what tests cover

## Step 5: Run Tests

```bash
# Run all tests for the crate
cargo test -p <crate>

# Run specific test
cargo test -p <crate> -- test_name

# Run with output
cargo test -p <crate> -- --nocapture
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
cargo fmt
cargo clippy -- -D warnings
cargo test
```

**Fix all errors — never suppress with `#[allow]`.**

## Rust-Specific Test Patterns

### Builder Pattern for Test Data

```rust
#[cfg(test)]
struct TestJobBuilder {
    executor: String,
    timeout: Option<Duration>,
}

#[cfg(test)]
impl TestJobBuilder {
    fn new() -> Self {
        Self {
            executor: "python".into(),
            timeout: None,
        }
    }
    fn executor(mut self, e: &str) -> Self { self.executor = e.into(); self }
    fn timeout(mut self, t: Duration) -> Self { self.timeout = Some(t); self }
    fn build(self) -> SubmitRequest { /* ... */ }
}
```

### Testing Error Variants

```rust
#[test]
fn rejects_invalid_input() {
    let result = parse_config("bad data");
    assert!(matches!(result, Err(ConfigError::ParseFailed { .. })));
}
```

### Testing with Traits (Dependency Injection)

```rust
#[cfg(test)]
struct MockAgentPool;

#[cfg(test)]
impl AgentPool for MockAgentPool {
    async fn acquire(&self) -> Result<AgentHandle, PoolError> {
        Ok(AgentHandle::fake())
    }
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
- **Wrong:** Add `#[allow(unused)]` or `#[allow(clippy::...)]`
- **Correct:** Fix root cause

## Summary

1. **Define scenarios** — success + error cases
2. **Write failing tests** — `#[test]` / `#[tokio::test]`
3. **Verify failure** — `cargo test` red
4. **Implement minimum** — typed, explicit errors
5. **Pass tests** — `cargo test` green
6. **Refactor** — `cargo fmt` + `cargo clippy` + keep green

**Remember: Red → Green → Refactor → Repeat**
