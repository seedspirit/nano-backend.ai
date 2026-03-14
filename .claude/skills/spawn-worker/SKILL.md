---
name: spawn-worker
description: Spawn an autonomous Claude Code worker in a tmux window for a GitHub issue — creates worktree, generates plan with success criteria, and launches iterative session.
user-invocable: true
---

# Spawn Worker

Create a worktree for a GitHub issue, generate an execution plan with success criteria, get user approval, then launch an iterative Claude Code session in a tmux window.

## Input

- **GitHub issue number** (required): e.g., `#12` or `12`
- **Base branch** (optional, default: `main`): branch to base the worktree on

If no issue number is provided, ask the user.

## Prerequisites

Verify before proceeding. On failure, report the cause and stop.

- `tmux` is installed: `command -v tmux`
- `claude` CLI is available: `command -v claude`
- Inside the nano-backend.ai repo: `test -d .git`

## Workflow

On command failure at any step, report the cause to the user and stop.

### 1. Fetch GitHub Issue

```bash
gh issue view <number> --json title,body,labels,state,assignees
```

Display a summary and ask user to approve and select options:

```markdown
## Issue: #12 — <title>
- **Labels**: bug / feature / ...
- **State**: open / closed
- **Assignees**: ...

> <first 200 chars of body>
```

Ask two questions simultaneously using `AskUserQuestion`:

1. **Base branch**: `main` (default) or other
2. **Model**: sonnet (Recommended) / opus (if plan requires complex cross-module reasoning)

### Model Selection

Default is **sonnet** — the plan is created by the main session (opus), so the worker only executes well-defined tasks.

Use **opus** when the plan involves ambiguous exploration or heavy cross-module changes that sonnet may struggle with.

### 2. Create Worktree

```bash
git fetch origin <base-branch>
git worktree add -b issue-<number> .worktrees/issue-<number> origin/<base-branch>
```

If the worktree already exists, ask the user:

- **Reuse**: use existing worktree as-is
- **Recreate**: remove and create fresh
- **Cancel**: abort

### 3. Create Plan

Performed directly by the spawn-worker skill (outside the loop, in the main Claude session):

1. Extract success criteria from the issue body (if present)
2. Explore the codebase using an Explore subagent to identify relevant code
3. Write `plan.md`:

```markdown
# Plan: #12 — <title>

## Success Criteria
### <Feature Area 1>
- [ ] <scenario: input/action → expected result>
- [ ] <scenario>
### <Feature Area 2>
- [ ] <scenario>
### Common
- [ ] `cargo test` passes for affected crates
- [ ] `cargo clippy -- -D warnings` clean

## Tasks
- [ ] Task 1: <description>
- [ ] Task 2: <description>
- [ ] Task 3: <description>
```

**Scenario-level**: specify concrete input/action/expected result, not "write tests". Group by feature with `###` subheadings.

**If success criteria are missing from the issue**: generate scenario-level criteria based on the issue content and code analysis.

**Always include as the last criteria**:
- `cargo test` passes for affected crates
- `cargo clippy -- -D warnings` clean

### 4. User Review Plan

Present plan.md to the user and request approval via `AskUserQuestion`:

```markdown
## Plan: #12 — Fix heartbeat timeout handling

### Success Criteria
#### Heartbeat
- [ ] agent sends heartbeat every 5s → manager tracks last_seen
- [ ] agent misses 3 beats → manager marks agent as lost
- [ ] lost agent reconnects → manager restores status
#### Common
- [ ] `cargo test` passes for affected crates
- [ ] `cargo clippy -- -D warnings` clean

### Tasks (3 tasks, estimated 3+1 iterations)
- [ ] Task 1: Implement heartbeat timeout detection in manager
- [ ] Task 2: Add reconnection logic for lost agents
- [ ] Task 3: Add integration test for timeout → reconnect flow

Proceed with this plan? (Let me know if you want any changes)
```

If the user requests changes, apply them and re-present. Proceed on approval.

### 5. Create Worker Files

Worker files are stored in `.workers/issue-<number>/`, **not** inside the worktree. This prevents accidental commits.

```bash
mkdir -p .workers/issue-<number>
```

Determine the absolute path of the current working directory:

```bash
REPO_ROOT="$(pwd)"
```

#### `.workers/issue-<number>/env`

Create this file using the Write tool with the following content (5 lines, no blank lines):

```
ISSUE_NUMBER=12
REPO_ROOT=/absolute/path/to/nano-backend.ai
MODEL=sonnet
BASE_BRANCH=main
MAX_ITERATIONS=5
```

**MAX_ITERATIONS** = number of tasks in the plan + 2 (submit iteration + buffer).

#### `.workers/issue-<number>/issue.md`

Create this file using the Write tool with the GitHub issue data:

```markdown
# #12
- **Title**: Fix heartbeat timeout handling
- **Labels**: bug
- **State**: open

## Description
<issue body as-is>
```

#### `.workers/issue-<number>/plan.md`

Write the plan created in step 3 to this file.

Write all three files using the Write tool.

### 6. Launch tmux Window

#### tmux Session

Check if session `nanodev` exists. If not, create it:

```bash
tmux has-session -t nanodev 2>/dev/null || tmux new-session -d -s nanodev
```

#### Worker Conflict (PID-based)

Check if a worker is already running for this issue using the PID file:

```bash
PID_FILE=".workers/issue-<number>/pid"
if [ -f "$PID_FILE" ]; then
  PID=$(cat "$PID_FILE")
  if kill -0 "$PID" 2>/dev/null; then
    echo "RUNNING"  # worker is alive
  else
    rm -f "$PID_FILE"  # stale pid file — clean up silently
  fi
fi
```

If the worker is **running** (pid alive), ask the user:
- **Reuse**: attach to the existing tmux window without launching a new process
- **Kill & Replace**: kill the existing worker (`kill $PID`), then launch a new one
- **Cancel**: abort

If the pid file does not exist (or was stale), proceed to launch normally.

#### Launch Command

```bash
tmux new-window -t nanodev -n "issue-<number>"
tmux send-keys -t "nanodev:issue-<number>" \
  "bash ${REPO_ROOT}/.claude/skills/spawn-worker/launch.sh ${REPO_ROOT}/.workers/issue-<number>" Enter
```

### 7. Report

```markdown
## Worker Spawned
- **Issue**: #<number> — <title>
- **Worktree**: `.worktrees/issue-<number>`
- **Worker files**: `.workers/issue-<number>/`
- **Plan**: <N> tasks, <M> criteria
- **Iterations**: max <MAX_ITERATIONS>
- **tmux**: session `nanodev`, window `issue-<number>`

### Commands
- Attach: `tmux attach -t nanodev:issue-<number>`
- Status: `/worker-status <number>`
```
