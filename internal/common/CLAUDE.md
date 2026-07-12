# Common Package Index

This directory contains shared code that can be used across binaries and manager/agent packages.

## Directory Index

- `agent/` — shared agent identity types.
- `data/` — pure application data types used by business logic.
- `dto/` — API boundary types used for serialization and deserialization.
- `encoding/` — small shared encoding helpers.
- `errordef/` — stable error codes and response-envelope mapping.
- `kernel/` — runtime-facing kernel types and ports.
- `kernel/` — manager-agent kernel contract types (CreationSpec, ID, value types).

## Constraints

- Keep role boundaries explicit. Do not place API DTOs or persistence entities in `data`.
- Shared packages must not import manager, agent, server, service, or repository implementation packages.
- Add new shared packages only when more than one higher-level package genuinely needs them.
