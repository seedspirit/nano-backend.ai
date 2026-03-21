---
name: autopilot
description: Fully automated pipeline — triggers /create-issue, /tdd-guide, /submit sequentially with self-review
user-invocable: true
---

# /autopilot — Automated Development Pipeline

Orchestrates existing skills end-to-end. Each phase **triggers the actual skill** via `Skill` tool and proceeds on completion.

## Input

- **Instruction**: What to build, fix, or change (plain text)
- **issue** (optional): Existing GitHub issue number — skip Phase 1

## Phases

### Phase 1: Issue Creation → trigger `/create-issue`

> Skip if issue number provided.

Invoke `Skill("create-issue")` with the user's instruction as context.

**Completion signal**: Issue number and URL are returned.
**Capture**: `ISSUE_NUMBER`, `ISSUE_TITLE`

### Phase 2: Branch Setup

```bash
git checkout -b issue-<ISSUE_NUMBER> main
```

If branch exists: ask **Reuse** / **Recreate** / **Cancel**.

No skill trigger — direct git command.

### Phase 3: Planning

1. Explore codebase with `Agent(subagent_type=Explore)` to identify relevant files
2. Generate plan:

```markdown
## Success Criteria
### <Feature Area>
- [ ] <scenario: input/action → expected result>
### Common
- [ ] go test ./... passes
- [ ] golangci-lint run ./... clean

## Tasks
- [ ] Task 1: <description>
- [ ] Task 2: <description>
```

**Checkpoint**: Present plan. **Wait for user approval.** Re-generate on feedback.

No skill trigger — internal planning.

### Phase 4: Implementation → trigger `/tdd-guide`

Invoke `Skill("tdd-guide")` for each task from the plan.

For each task:
1. `/tdd-guide` drives: scenarios → failing tests → implement → refactor
2. Verify: `go test ./internal/<pkg>/...` + `golangci-lint run ./...`

**Completion signal**: All plan tasks checked off, tests green.

### Phase 5: Test Verification + Self-Review → reference `/go-guide`

Invoke `Skill("go-guide")` to load conventions, then:

#### 5a. Test Coverage Check

1. List all tests in changed packages. For each public function or endpoint, confirm:
   - **Success scenario** exists (valid input → expected output)
   - **Error/edge scenario** exists (invalid input, boundary → expected error)

2. Report coverage:
   ```
   ## Test Coverage Report
   ### <package>::<file>
   - ✅ success: <description>
   - ✅ error: <description>
   - ❌ missing: <what's not tested>
   ```

3. Write any missing tests before proceeding.

#### 5b. Code Review

1. `git diff main..HEAD` — read every changed file
2. Verify against checklist:

| Check | Source |
|-------|--------|
| No `panic()` in library code | CLAUDE.md |
| `log/slog` only, no `fmt.Println` | CLAUDE.md |
| External endpoints return `ApiResponse` | CLAUDE.md |
| Every public fn has success + error tests | CLAUDE.md |
| Domain error types with `error` interface | /go-guide |
| Descriptive naming | /go-guide |
| Minimal exported API | /go-guide |

3. Fix findings immediately
4. Run quality checks:
   ```bash
   gofmt -l .
   golangci-lint run ./...
   go test ./...
   ```
5. Repeat until clean (max 3 iterations). Stop and report if still failing.

### Phase 6: Submit → trigger `/submit`

Invoke `Skill("submit", args="issue <ISSUE_NUMBER>")`.

`/submit` handles: quality enforcement → learning notes → commit → push → PR.

**Completion signal**: PR URL returned.

## Phase Transitions

```
Phase 1 ──completed(ISSUE_NUMBER)──→ Phase 2
Phase 2 ──branch ready──→ Phase 3
Phase 3 ──user approved──→ Phase 4
Phase 4 ──tests green──→ Phase 5
Phase 5 ──review clean──→ Phase 6
Phase 6 ──PR created──→ Done
```

All transitions are automatic except Phase 3→4 (plan approval).

## Error Handling

| Error | Action |
|-------|--------|
| `/create-issue` fails | Report error, ask user to create manually |
| Tests fail after impl | Diagnose, fix, retry (max 3 per task) |
| Self-review fails 3x | Stop, report findings to user |
| `/submit` fails | Report error, show manual recovery steps |

## Example

```
/autopilot manager에 GET /sessions endpoint 추가. 빈 배열 반환하면 됨
/autopilot issue 5
```
