---
name: pilot
description: User-driven step-by-step pipeline — same phases as /autopilot but you control every transition
user-invocable: true
---

# /pilot — User-Controlled Development Pipeline

Same pipeline as `/autopilot`, but **you decide when to proceed** to each next phase.

After every phase, shows a status dashboard and asks what to do next.

## Input

- **Instruction**: What to build, fix, or change (plain text)
- **issue** (optional): Existing GitHub issue number — skip Phase 1

## Phases

Same 6 phases as `/autopilot`. Each phase triggers the same underlying skill.

| # | Phase | Triggers | Produces |
|---|-------|----------|----------|
| 1 | Issue | `/create-issue` | Issue number + URL |
| 2 | Branch | git commands | Branch `issue-<N>` |
| 3 | Plan | Explore agents | Plan with success criteria |
| 4 | Implement | `/tdd-guide` | Working code + tests |
| 5 | Review | `/go-guide` ref | Clean diff |
| 6 | Submit | `/submit` | PR URL |

## Dashboard

After each phase completes, display:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Phase 3/6 ✓ Plan approved

  [1] Issue    ✓ #12 — Add sessions endpoint
  [2] Branch   ✓ issue-12
  [3] Plan     ✓ 3 tasks, 5 criteria
  [4] Implement  ⬜ pending
  [5] Review     ⬜ pending
  [6] Submit     ⬜ pending
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Then ask via `AskUserQuestion`:
- **Next** — proceed to next phase
- **Redo** — re-run current phase
- **Skip** — skip to a specific phase
- **Stop** — pause here (can resume later with `/pilot resume`)

## Key Difference from /autopilot

| | /autopilot | /pilot |
|---|-----------|--------|
| User checkpoints | Plan approval only | Every phase |
| Pace | Fast, uninterrupted | User-controlled |
| Best for | Routine tasks, trusted flow | Learning, complex tasks, first-time features |

## Example

```
/pilot manager에 GET /sessions endpoint 추가
```

```
> Phase 1 complete: Created #12 — Add GET /sessions endpoint
> Next phase: Branch Setup. Proceed?
```

```
/pilot issue 5
```

```
> Skipping Phase 1 (issue #5 already exists)
> Phase 2: Branch Setup. Proceed?
```
