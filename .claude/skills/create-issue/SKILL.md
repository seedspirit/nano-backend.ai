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
- **No input** — will ask interactively

## Workflow

### 1. Gather Information

Ask the user (skip fields already provided):

1. **Issue type**: Bug / Feature / Task
2. **Title**: concise summary (imperative mood, under 70 chars)
3. **Description**: what, why, and acceptance criteria

### 2. Generate Issue Body

#### Bug Template

```markdown
## Bug Report

**Component**: manager / agent / common / ...
**Severity**: Highest / High / Medium / Low

### Symptom
<What was observed>

### Expected Behavior
<What should happen>

### Root Cause
<If known from /analyze, otherwise "TBD">

### Reproduction Steps
1. ...
2. ...

### Related Code
- `crates/...`
```

#### Feature / Task Template

```markdown
## Feature Request / Task

**Component**: manager / agent / common / ...

### Goal
<What this achieves and why>

### Acceptance Criteria
- [ ] <concrete, testable criterion>
- [ ] <criterion>

### Notes
<Design considerations, references, constraints>
```

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
| Bug | `bug` |
| Feature | `enhancement` |
| Task | `task` |

Component labels: `manager`, `agent`, `common`, `infra`

If labels don't exist yet, create without labels and note it.

### 5. Report

```markdown
## Issue Created
- **Number**: #<number>
- **Title**: <title>
- **URL**: <url>
- **Labels**: <labels>

### Next Steps
- Start work: `git checkout -b issue-<number>` then `/tdd-guide`
- Delegate to worker: `/spawn-worker <number>`
```
