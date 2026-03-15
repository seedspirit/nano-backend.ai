# Common Crate — Agent Guidelines

Library crate that defines shared types and contracts used across the entire workspace.

## Role

- Provide shared types depended on by all other crates (`ApiResponse`, `CommonError`)
- Single source of truth for the API response format
- No runtime logic — only pure data structures and traits

## Rules

- Changes to `ApiResponse` affect all consumers — verify downstream crate tests before merging
- Must not depend on any other workspace crate (leaf of the dependency graph)
- New public types must be re-exported via `pub use` in `lib.rs`
- Error types are infrastructure-level only — domain errors belong in each crate
