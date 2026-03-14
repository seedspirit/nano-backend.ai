---
name: rust-guide
description: Rust coding conventions for nano-backend.ai — error handling, type design, async patterns, quality enforcement
user-invocable: true
---

# Rust Coding Guide

Condensed conventions for all Rust code in this project.
References: [Microsoft Rust Guidelines](https://microsoft.github.io/rust-guidelines), Rust API Guidelines.

## Error Handling

- Return `Result<T, E>` for all fallible operations — propagate with `?`
- Define domain error types with `thiserror::Error`
- **No `.unwrap()` / `.expect()` in production code** — allowed only in tests
- No silent failures: never return `Option::None` when an error should be raised

```rust
// Good
#[derive(Debug, thiserror::Error)]
pub enum JobError {
    #[error("job {0} not found")]
    NotFound(JobId),
    #[error("agent unreachable: {0}")]
    AgentUnreachable(String),
}

// Bad — primitive string errors
fn do_thing() -> Result<(), String> { ... }
```

## Type Design

- **Avoid primitive obsession** — wrap IDs, keys, tokens in newtypes
- All public types must implement `Debug`
- Types shown to users should also implement `Display`

```rust
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct SessionId(pub Uuid);
```

## Naming (M-CONCISE-NAMES)

- Avoid generic weasel words: ~~Manager~~, ~~Service~~, ~~Factory~~, ~~Handler~~
- Use names that describe *what it does*: `JobScheduler`, `HeartbeatMonitor`, `SessionStore`

## Panic Policy (M-PANIC-IS-STOP / M-PANIC-ON-BUG)

- Panics = program termination, not control flow
- Use panics only for **unrecoverable programming errors** (invariant violations)
- Never panic in library code — return `Result`

## Unsafe Policy (M-UNSAFE)

Valid reasons: FFI calls, novel low-level abstractions, proven performance need.

- Must include `// SAFETY: ...` comment explaining why
- Keep unsafe blocks minimal and isolated
- Run `cargo miri test` when possible

## Async Patterns

- Runtime: **Tokio** (multi-threaded)
- All async task boundaries require `Send + 'static`
- **No blocking in async** — use `tokio::task::spawn_blocking` for CPU-heavy or blocking I/O
- Prefer structured concurrency (`tokio::select!`, `JoinSet`) over loose spawns

## Logging (M-LOG-STRUCTURED)

- Use `tracing` crate — no `println!` or `eprintln!`
- Structured fields: `tracing::info!(job_id = %id, status = "started")`
- Redact sensitive data (tokens, passwords) — never log secrets

## Lint & Formatting (M-STATIC-VERIFICATION)

```bash
cargo fmt                        # format all code
cargo clippy -- -D warnings      # zero warnings policy
cargo audit                      # dependency vulnerability check
```

- Use `#[expect(lint)]` instead of `#[allow(lint)]` to catch stale suppressions (M-LINT-OVERRIDE-EXPECT)
- Never suppress lints to avoid fixing issues

## Crate Design (M-SMALLER-CRATES)

- Favor multiple focused crates over a monolith
- Each crate has a single clear responsibility
- No circular dependencies between crates
- Explicit `pub` boundaries — expose only what is needed

## Magic Values (M-DOCUMENTED-MAGIC)

- All hardcoded values must have a comment explaining *why* that value
- Include references to external specs or system requirements

```rust
/// Maximum heartbeat interval before an agent is considered dead.
/// Chosen to tolerate 2 missed beats at the 5s send interval.
const HEARTBEAT_TIMEOUT: Duration = Duration::from_secs(15);
```

## Dependencies

Preferred crates:

| Purpose | Crate |
|---------|-------|
| Async runtime | `tokio` |
| HTTP server | `axum` |
| Serialization | `serde` + `serde_json` |
| Database | `sqlx` (PostgreSQL) |
| Redis | `glide` (Valkey) |
| gRPC | `tonic` |
| Error types | `thiserror` |
| Error context | `anyhow` (binaries only) |
| Logging | `tracing` + `tracing-subscriber` |
| CLI | `clap` |
