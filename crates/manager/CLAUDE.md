# Manager Crate — Agent Guidelines

Binary crate serving as the HTTP API server and service entry point.

## Role

- Receive external HTTP requests and route them to handlers
- Return structured JSON responses via `common::ApiResponse`
- Manage server bootstrap and lifecycle

## Rules

- `main.rs` is bootstrap only — move business logic into modules
- Adding a new endpoint: register the route in `app.rs` + create a dedicated handler module
- Every handler must have both success and error test cases
- Use only `tracing` macros for logging (`debug!`, `info!`, `warn!`, `error!`)
- Handler-specific errors belong in each handler module, not in `error.rs`
- `error.rs` is reserved for server startup/operation errors (`Bind`, `Serve`)
