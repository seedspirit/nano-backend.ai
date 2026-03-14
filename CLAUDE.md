# Nano Backend.AI — Agent Guidelines (Root)

See `README.md` for project overview, architecture, and tech stack.

## Documentation Hierarchy

- **Root CLAUDE.md** (this file): global principles applicable everywhere
- **Sub-directory CLAUDE.md**: local rules scoped to that directory only
- **`docs/design/`**: detailed design documents and rationale

## CLAUDE.md Authoring Rules

- Policy and role only — no verbose implementation details
- Root document = global principles; sub-documents = local rules
- Keep each file short so agent context is not overwhelmed

## Language & Conventions

- Rust (latest stable), async runtime: Tokio
- Format: `cargo fmt` — all code must pass before commit
- Lint: `cargo clippy -- -D warnings` — treat all warnings as errors
- Write English comments; Korean is acceptable in design docs under `docs/`

## API Response Conventions

All external API responses use structured JSON:

```json
{ "status": "...", "reason": "...", "next_action_hint": "..." }
```

No unstructured text in API responses. Long-running operations return a pollable job ID.

## Dependency Rules

- No circular dependencies between crates
- Explicit `pub` boundaries — expose only what is needed
- Internal crate communication via defined interfaces, not reaching into internals

## Test Principles

- Unit tests: `#[cfg(test)] mod tests` alongside source
- Integration tests: top-level `tests/` directory
- Every public function should have at least one test

## Prohibitions

- No `.unwrap()` / `.expect()` in production code — use `Result` or `?`
- No `unsafe` without a comment justifying why it is necessary
- No panicking in library code
- No `println!` for logging — use `tracing` crate
