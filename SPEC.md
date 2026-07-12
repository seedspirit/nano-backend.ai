# nano-backend.ai Product and Architecture Direction

> Status: Draft
> Current scope: Phase 0 — agent-native model development on a single GPU node

## 1. Vision

nano-backend.ai is not intended to remain a fine-tuning runner. Its long-term
goal is to become a small Backend.AI where AI agents can receive model
development goals, write and submit code, explore data in notebooks, iterate on
training and evaluation, compare outcomes, and decide what to do next.

Users and agents should express intent instead of Docker commands or physical
GPU indices:

- open a development environment and explore data or models;
- submit finite code for execution;
- iterate over training and evaluation;
- open a completed model as an inference environment;
- reproduce an earlier attempt from its inputs and outputs.

The platform turns this intent into validated execution specifications,
allocates resources, observes execution, and keeps a durable record of both the
process and its results.

## 2. Current Phase and Long-term Direction

The long-term vision is broad, but Phase 0 deliberately starts small:

- one machine with 2× RTX 3090 GPUs;
- one GPU per session;
- preset-validated LoRA/SFT batch sessions;
- a SQLite session ledger;
- local filesystem artifacts;
- at most two concurrent sessions.

Fine-tuning is the first vertical slice used to validate the complete path from
session submission to artifact preservation. It does not define the permanent
product boundary.

## 3. Domain Model

```text
Project
  └─ AgentTask
      └─ Experiment
          └─ Session
              └─ Kernel
```

### Project

A namespace for related sessions, experiments, artifacts, and assets. Over
time, it becomes the shared model-development context for people and agents.

### AgentTask

A long-running goal delegated to an AI agent, such as “develop an adapter that
beats the baseline on this dataset.” One AgentTask may create multiple
Experiments and Sessions.

AgentTask is introduced only when its lifecycle and completion conditions are
concrete. Phase 0 reserves the term and responsibility boundary.

### Experiment

Groups development attempts and results around a hypothesis or comparison. It
connects training and evaluation sessions with metrics, reports, and model
artifacts.

Experiment is not persisted in Phase 0. It is defined now to keep experiment
orchestration responsibilities out of Session.

### Session

The user-visible compute lifecycle. An immutable SessionSpec describes what to
execute, required resources, and referenced assets and mounts.

Session types follow Backend.AI terminology:

- `interactive`: Jupyter, shell, or IDE development environments;
- `batch`: finite training, evaluation, preprocessing, or submitted code;
- `inference`: model-serving environments;
- `system`: platform-internal computation.

Phase 0 fine-tuning is a `batch` session.

### Kernel

The isolated compute unit materialized by an Agent. Docker containers are the
initial implementation, but a Kernel may also be backed by a process, VM, or
Kubernetes Pod.

The Manager owns Sessions; Agents materialize Kernels. Docker is a Kernel
implementation detail, not a public domain contract.

## 4. Core Design Principles

### Agent-native interfaces

APIs prioritize reliable agent decisions over prose-first responses:

- stable error codes;
- explicit status and result;
- pollable long-running resources;
- next-action context for recovery;
- deterministic specification normalization.

### Declarative input

Users do not describe how to create containers. They submit a SessionSpecDraft
with preset references and option parameters. The platform validates and
finalizes it into an immutable SessionSpec.

### Reproducibility first

Every terminal session preserves its inputs, resolved configuration, logs,
metrics, reports, and available outputs. Failure is also a result that must be
inspectable and reproducible.

### Infrastructure detail isolation

Schedulers and services do not depend on Docker SDK types or transport DTOs.
The shared Kernel contract between Manager and Agent remains independent of
transport and runtime implementation.

### Separate lifecycle from outcome

Session status represents lifecycle position; result represents terminal
outcome.

```text
pending → preparing → running → terminated
                                  └─ success | failure
```

A failed execution still has `status=terminated` and `result=failure`.
`failure_reason` records the concrete machine-readable cause.

### Start small with extensible boundaries

Phase 0 does not require a generic scheduling framework or workflow engine. It
does require clear Session, Kernel, Runtime, and Artifact boundaries so future
features can be added without redefining existing concepts.

## 5. Roles of Specs and Presets

### SessionSpecDraft

Unfinalized intent submitted by a user or agent. Values may be omitted and
preset defaults may not yet be applied.

### SessionSpec

An immutable execution definition after validation and default resolution. The
same SessionSpec should be reusable for reproduction.

### Preset

A Preset is a validated behavioral contract between the platform and a runtime,
not merely a configuration bundle. It defines accepted input, defaults, and
required outputs.

TrainerPreset is central to Phase 0. Resource and output policies should become
independently composable as their requirements mature.

Fine-tuning, notebooks, evaluation, and serving are purposes that resolve into
SessionSpecs rather than unrelated top-level execution objects.

## 6. From Session to Kernel

```text
SessionSpecDraft
  → validation and preset resolution
  → immutable SessionSpec
  → Session persistence
  → scheduling and resource allocation
  → kernel.CreationSpec
  → Agent runtime
  → Kernel
  → observation and artifact reconciliation
```

- The Manager owns session lifecycle and scheduling decisions.
- A Provisioner selects the Agent and physical GPU device.
- `kernel.CreationSpec` is the bound Manager-Agent execution contract.
- The Agent runtime translates it into Docker or another substrate.
- Terminal observation returns to the Manager reconciliation flow, which
  finalizes Session status, result, and artifacts.

FIFO scheduling and explicit GPU allocation are sufficient for Phase 0.
Fairness, priorities, and multi-node placement should be added only when their
requirements become real.

## 7. Artifacts and Reproducibility

Artifacts are a core output of the session ledger, not an optional add-on.

The platform preserves at least these categories:

- submitted and finalized specifications;
- runtime-resolved configuration;
- stdout and stderr;
- structured metrics;
- human-readable reports;
- successful outputs or partial outputs;
- execution environment and resource allocation metadata.

ArtifactIndex provides metadata such as location, size, and checksum. Phase 0
uses a local filesystem, but the Session domain must not depend on a particular
storage backend.

## 8. API Direction

Session is the primary user-facing API resource:

- submit a session spec draft;
- inspect a session and its finalized spec;
- list sessions by project;
- poll logs and artifacts;
- eventually request termination or cancellation.

Kernel is the primary Manager-Agent API resource:

- prepare or create a kernel;
- start a kernel;
- observe kernel status;
- clean up or destroy a kernel.

External APIs and internal Agent APIs do not share transport DTOs. Domain types
remain transport-independent.

## 9. Failure and Recovery Direction

Failures are exposed through a stable taxonomy instead of raw runtime messages.
Examples include:

- image pull failure;
- model or dataset staging failure;
- resource exhaustion or OOM;
- timeout;
- trainer failure;
- artifact reconciliation failure.

Every failure preserves logs and partial artifacts where possible. An agent
uses `failure_reason` and session context to choose between retrying, modifying
the specification, or reporting to a user.

Idempotency prevents network retries from creating duplicate sessions. The
same key and normalized spec return the existing session; a different spec
with the same key produces a conflict.

## 10. Evolution by Phase

### Phase 0 — Fine-tuning vertical slice

- preset-backed batch sessions;
- single-node GPU allocation;
- session ledger and basic query APIs;
- Kernel contract and Docker runtime direction;
- preservation of logs, metrics, and model artifacts.

### Phase 1 — Complete the execution path

- Manager-Agent Kernel API;
- Docker runtime materialization;
- asset staging and caching;
- terminal observation and reconciliation;
- mapping runtime failures into the stable taxonomy.

### Phase 2 — Development environments and control

- interactive notebook and shell sessions;
- submitted-code batch sessions;
- cancellation, termination, and timeout;
- orphan Kernel cleanup;
- richer log and artifact APIs.

### Phase 3 — Agent-driven model development

- Experiment lifecycle;
- AgentTask and completion criteria;
- training and evaluation comparison loops;
- lineage and model promotion;
- agents planning subsequent sessions from observed results.

### Later extensions

- inference sessions and model serving;
- multi-node or multi-GPU sessions;
- resource groups and richer scheduling policies;
- remote artifact storage;
- multi-tenant quotas and policies.

## 11. MVP Non-goals

- a generic workflow engine;
- a distributed scheduler framework;
- Kubernetes-native orchestration;
- multi-tenant quota or billing;
- complex priorities and bin-packing;
- a real-time dashboard;
- mandatory dependency on external experiment-tracking SaaS.

These are not permanent non-goals. They are unnecessary for validating the
core boundaries in Phase 0.

## 12. Success Criteria

Phase 0 succeeds when all of the following are true:

- Can an agent submit a training session without knowing Docker details?
- Is invalid input rejected before it occupies a GPU?
- Can session state and failure cause be understood through APIs alone?
- Can every terminal session be reproduced from its recorded inputs and
  artifacts?
- Can the scheduler safely allocate and release both GPUs?
- Can another runtime or session type be introduced without redefining the
  core domain?

Concrete types and package layout are sourced from the code and `docs/design/`.
This document defines the product and architecture direction they should serve.
