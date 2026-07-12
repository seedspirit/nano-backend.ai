# Nano Backend.AI

A small Go backend for an agent-native fine-tuning ledger.

Phase 0 is intentionally narrow: it targets a single machine with 2x RTX 3090 GPUs and executes one-GPU LoRA fine-tuning sessions from validated presets. The goal is not to expose a generic job runner; the goal is to make training sessions submit-able, inspectable, and reproducible by AI agents.

See [`SPEC.md`](SPEC.md) for the full MVP contract.

## Agent Guidance

`CLAUDE.md` is the canonical agent instruction file for this repository. Agents
that start from `AGENTS.md` should treat it as a pointer to `CLAUDE.md` and then
follow the same shared rules.

Use `.claude/skills/README.md` for the available project workflows and skills.

## MVP Goals

- Accept declarative session spec drafts built from preset refs and option parameters
- Validate presets and option policies before consuming queue or GPU capacity
- Persist every session in a local SQLite ledger
- Execute at most two single-GPU sessions concurrently
- Keep Docker behind an agent-side kernel runtime
- Preserve logs, config, metrics, and artifacts for every terminal session
- Make failures machine-readable through explicit session status and result and `failure_reason`

## API Design Philosophy

AI agents are the primary consumer. Responses should be machine-readable first:
structured JSON envelopes, endpoint-specific `data` payloads, explicit statuses,
stable `error.code` values, and clear next-step hints where useful.

Long-running operations expose pollable resources. For Phase 0, logs use cursor-based polling rather than WebSockets.

## Phase 0 Architecture

```text
SessionSpecDraft
  -> API preflight validation
  -> preset registry / spec builder
  -> immutable session spec.Spec
  -> SQLite session ledger
  -> SchedulerCoordinator
  -> SessionProvisioner / GPU claim
  -> KernelCreationSpec
  -> KernelLauncher
  -> DockerRuntime
  -> local artifact store
```

| Component | Role |
|-----------|------|
| HTTP API | Submit and inspect sessions, logs, and artifacts |
| Spec builder | Resolve preset refs, validate option parameters, and finalize immutable specs |
| SchedulerCoordinator | Own session lifecycle transitions and terminal reconciliation |
| SessionProvisioner | FIFO scheduling, 2-GPU assignment, and kernel creation spec construction |
| KernelLauncher | Manager-side port for prepare/start/cleanup calls |
| DockerRuntime | Agent-side Docker container materialization and observation |
| SQLite | Durable source of truth for projects, sessions, and artifact metadata |
| Local artifact store | Stores specs, resolved configs, logs, metrics, reports, adapters |

## Tech Stack

- **Language:** Go
- **External API:** HTTP + JSON REST
- **Database:** SQLite for Phase 0
- **Kernel substrate:** Agent-side Docker backend behind REST/HTTP and a Go port
- **Storage:** Local filesystem artifact store

## Future Architecture Notes

Postgres, Redis/Valkey hints, alternate manager-agent transports, multi-node scheduling, and richer cancellation/orphan cleanup semantics are future architecture directions, not Phase 0 requirements.

## Non-Goals (MVP)

- Multi-tenant quota or policy enforcement
- Distributed training
- Kubernetes native integration
- Real-time serving orchestration
- Web UI or dashboard
- Advanced scheduling or bin-packing
- Webhook or notification system
- W&B SaaS integration
- Cancel API implementation, deferred to Phase 2

## Project Layout

```text
├── CLAUDE.md          # Canonical AI agent guidelines
├── AGENTS.md          # Pointer for agents that read AGENTS.md first
├── cmd/               # Binary entry points
├── internal/          # Private packages
├── docs/              # Design, education, and learning notes
├── SPEC.md            # Phase 0 MVP specification
└── Makefile           # Build, test, lint, fmt targets
```

## License

TBD
