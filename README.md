# Nano Backend.AI

A minimal reimplementation of [Backend.AI](https://backend.ai) — designed for learning,
on-premise deployment, and AI-agent-first interaction.

## Goals

- Relearn Backend.AI's core architecture through a small-scale redesign
- Practice production-grade design decisions
- Experiment with AI-agent-first codebase operations
- Prioritize easy install, upgrade, and migration for on-premise use

## API Design Philosophy

AI agents are the primary consumer. Every response is machine-readable first:
structured JSON with explicit `status`, `reason`, and `next_action_hint` fields.
Long-running operations expose a pollable job model.
Human users interact through a `/v1/chat/completions`-compatible conversational layer.

## Architecture

```
┌─────────────────┐
│  Reverse Proxy  │  ← TLS termination, static/API proxy
└───────┬─────────┘
        │
┌───────▼───────────┐      gRPC      ┌─────────────────┐
│     Manager       │ ◄────────────► │    Agent(s)     │
│  (control plane)  │                │(execution plane)│
└──┬──────────┬─────┘                └─────────────────┘
   │          │
   ▼          ▼
┌────────┐  ┌───────┐
│Postgres│  │ Redis │
│(state) │  │(coord)│
└────────┘  └───────┘
```

| Component     | Role                                          |
|---------------|-----------------------------------------------|
| Manager       | External API, scheduling, metadata, state mgmt |
| Agent         | Job execution, heartbeat, local executor       |
| PostgreSQL    | Durable state, source of truth                 |
| Redis         | Ephemeral coordination, event bus, cache        |
| Reverse Proxy | Ingress, TLS, static & API proxy               |

## Tech Stack

- **Language:** Go
- **External API:** HTTP + JSON REST (net/http + chi)
- **Internal API:** gRPC
- **Database:** PostgreSQL
- **Cache / Coordination:** Redis

## Non-Goals (v0)

Multi-tenancy · Auth · GPU scheduling · Image build pipeline · Advanced proxy routing · Distributed DB

## Project Layout

```
├── CLAUDE.md          # AI agent guidelines (root-level principles)
├── cmd/               # Binary entry points (manager, agent)
├── internal/          # Private packages (common, manager, agent)
├── docs/design/       # Detailed design documents
└── Makefile           # Build, test, lint, fmt targets
```

## License

TBD
