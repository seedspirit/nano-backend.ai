---
name: create-issue
description: Create a GitHub issue with structured template — supports bug reports, feature requests, and tasks with component labels.
user-invocable: true
---

# Create Issue

Create a well-structured GitHub issue. Typically invoked after `/analyze` finds a bug, or when planning new work.

## Input

Accepts one of:
- **Bug report** from `/analyze` output (auto-fills from analysis)
- **Feature / task description** in plain text
- **Epic definition** with child stories
- **No input** — will ask interactively

## Workflow

### 1. Gather Information

Ask the user (skip fields already provided):

1. **Issue type**: Epic / Story / Bug / Feature / Task
2. **Title**: concise summary (imperative mood, under 70 chars)
3. **Description**: what, why, and acceptance criteria
4. **Parent Epic**: (Story only) link to parent epic issue number

### 2. Generate Issue Body

Use the template matching the issue type.

#### Epic Template

An Epic is a top-level goal containing multiple Stories. Link it to a GitHub Milestone.

```markdown
## Epic

### Goal
<Single business-level goal — what capability does the system gain when this Epic is done?>

### Motivation
<Why now? What problem exists today, or what opportunity does this unlock?
Reference user needs, architecture gaps, or upstream requirements.>

### Context
<Current state of the codebase relevant to this Epic.
Which crates/modules are involved? What exists today vs. what needs to change?
Link to design docs if available (e.g., `docs/design/xxx.md`).>

### Stories

| # | Story | Summary | Component | Depends on |
|---|-------|---------|-----------|------------|
| S1 | <title> | <one-line: what + why> | common | — |
| S2 | <title> | <one-line: what + why> | manager | S1 |
| S3 | <title> | <one-line: what + why> | manager | S1 |

### Dependency Graph
<blocks/blockedBy relationships between Stories.
Highlight which Stories can proceed in parallel.
Example: "S2 and S3 can start in parallel once S1 merges">

### Design Decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| <e.g., trait-based DI for storage> | <reason> | <alternative and why rejected> |

### Out of Scope
<What this Epic intentionally does NOT cover. Prevents scope creep.>

### Success Criteria
<How do we know the Epic is truly done? System-level observable outcomes.>
```

#### Story Template

A Story is 1 PR = one clear deliverable.

```markdown
## Story

**Epic**: #<epic-issue-number>
**Component**: manager / agent / common / ...

### Background
<Why this Story exists. What user need or architectural requirement drives it?
How does it fit into the parent Epic's goal?>

### Goal
<Single goal this Story achieves>

### Acceptance Criteria
- [ ] <Concrete, testable condition>
- [ ] <condition>
- [ ] <condition> (max 3 — split if more)

### Affected Code
<Which crates, modules, or files will be created/modified?
e.g., `crates/common/src/response.rs`, `crates/manager/src/routes/`>

### Design Notes
<Key design choices for this Story: traits to define/implement, error types, API shape.
Link to relevant existing code or design docs.
NOT full implementation details — just enough for an agent to understand the approach.>

### Test Plan
- <What scenarios will be tested?>
- <e.g., "Unit test: valid input returns Ok(...)" >
- <e.g., "Unit test: missing field returns ValidationError">
- <e.g., "Integration test: GET /endpoint returns 200 with expected JSON shape">
```

#### Bug Template

```markdown
## Bug Report

**Component**: manager / agent / common / ...
**Severity**: Highest / High / Medium / Low

### Symptom
<What was observed? Include error messages, logs, or unexpected behavior.>

### Expected Behavior
<What should happen instead?>

### Impact
<Who or what is affected? How severe is the disruption?
e.g., "Blocks all API responses" or "Cosmetic — wrong log level">

### Root Cause
<If known from /analyze: explain the code path that causes the bug.
Reference specific files/functions. Otherwise "TBD — needs investigation".>

### Reproduction Steps
1. ...
2. ...

### Affected Code
- `internal/...` or `cmd/...` — <brief description of what this file does in the bug path>

### Fix Direction
<If known: outline the approach (not full implementation).
e.g., "Add validation in `parse_request()` before passing to handler">

### Test Plan
- <How to verify the fix?>
- <e.g., "Add test: invalid input no longer panics, returns 400">
- <e.g., "Existing test X should still pass">
```

#### Feature / Task Template

```markdown
## Feature Request / Task

**Component**: manager / agent / common / ...

### Background
<Why is this needed? What problem does it solve or what does it enable?
Link to parent issue or Epic if applicable.>

### Goal
<What this achieves — stated as a concrete outcome>

### Acceptance Criteria
- [ ] <concrete, testable criterion>
- [ ] <criterion>

### Affected Code
<Which crates/modules/files will be created or changed?>

### Design Notes
<Approach, relevant patterns, constraints. Link to existing code or docs.>

### Test Plan
- <What scenarios will be tested?>
- <Both success and error/edge cases>
```

### Quality Bar

Before presenting the draft, verify every issue body meets these criteria:

- **Self-contained**: A reader (human or AI agent) can understand the issue without opening other tabs. If the issue references external context, summarize it inline rather than just linking.
- **"Why" before "What"**: Background/Motivation section explains the reason this issue exists. Never skip it.
- **Affected Code is concrete**: List actual crate/module/file paths, not vague references like "the code" or "relevant modules".
- **Test Plan is scenario-based**: Each item describes what is being verified (e.g., "Unit test: empty input returns ValidationError"), not just "add tests".
- **ACs are observable**: Each criterion can be checked by running a command or reading an output. Avoid subjective criteria like "clean code" or "well-structured".
- **No placeholder text**: Every `<angle-bracket placeholder>` in the template must be replaced with real content. If a section genuinely does not apply, remove it rather than leaving a placeholder.

### 3. User Review

Present the draft to the user via `AskUserQuestion`:
- **Create as-is**
- **Edit** — let the user modify
- **Cancel**

### 4. Create Issue

```bash
gh issue create \
  --title "<title>" \
  --body "<body>" \
  --label "<component>,<type>"
```

Label mapping:

| Type | Label |
|------|-------|
| Epic | `epic` |
| Story | `story` |
| Bug | `bug` |
| Feature | `enhancement` |
| Task | `task` |

When creating an Epic, also create a GitHub Milestone with the same name.
When creating a Story, link it to the parent Epic's Milestone.

Component labels: `manager`, `agent`, `common`, `infra`

If labels don't exist yet, create without labels and note it.

### 5. Story Decomposition Check (Epic only)

After creating the Epic, create its child Stories sequentially:

1. Create each Story as an individual issue (using the Story Template)
2. Link to the Epic's Milestone
3. Note dependency relationships in the issue body with `blocks: #N` / `blockedBy: #N`

Validate during decomposition:
- Does each Story have a single goal?
- Are ACs 3 or fewer?
- Is there a preceding trait/interface Story enabling parallel implementation Stories?
- Are changes contained within module boundaries?

### 6. Report

```markdown
## Issue Created
- **Number**: #<number>
- **Title**: <title>
- **URL**: <url>
- **Labels**: <labels>
- **Milestone**: <milestone> (Epic/Story only)

### Next Steps
- Start work: `git checkout -b issue-<number>` then `/tdd-guide`
- Delegate to worker: `/spawn-worker <number>`
```
