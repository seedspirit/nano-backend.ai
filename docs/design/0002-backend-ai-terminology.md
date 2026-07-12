# Backend.AI-aligned domain terminology

This repository uses Backend.AI terminology for compute infrastructure and
keeps product terminology for model-development orchestration.

## Canonical hierarchy

```text
Project
  └─ AgentTask              future: a long-running goal delegated to an AI agent
      └─ Experiment         future: related development attempts and comparisons
          └─ Session        a user-visible compute environment or execution
              └─ Kernel     an isolated unit created by an Agent
```

`AgentTask` and `Experiment` are reserved domain boundaries, not yet persisted
entities. They should be introduced only with concrete lifecycle and acceptance
requirements.

## Current terms

| Term | Meaning |
|------|---------|
| Project | Namespace for sessions, artifacts, and future experiments. |
| Session | User-visible compute lifecycle. Types are interactive, batch, inference, and system. |
| Session spec draft | User and preset input before validation and finalization. |
| Session spec | Immutable logical definition of a session. |
| Kernel | Agent-side isolated compute unit, commonly a container or process. |
| Kernel creation spec | Fully bound manager-to-agent contract used to create a kernel. |
| Resource slots | Requested CPU, memory, and accelerator capacity. |
| Allocation | Provisioner's selected Agent and physical accelerator devices. |
| Runtime | Agent implementation that materializes a kernel as a process, container, VM, or pod. |

## Status and result

Session lifecycle and outcome are separate:

```text
pending → preparing → running → terminated
                                  └─ success | failure
```

Do not use `succeeded` or `failed` as terminal lifecycle states. Failure detail
belongs in `failure_reason` while the result remains `failure`.

## Product-specific specifications

Fine-tuning, notebook development, evaluation, and model serving are purposes
that produce sessions; they are not alternative names for a session or kernel.
Future product types such as `TrainingSpec` may compile into a batch
`SessionSpec`, while a notebook request compiles into an interactive one.
