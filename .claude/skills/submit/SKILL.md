---
name: submit
description: Quality checks, learning notes, commit, PR creation workflow for nano-backend.ai
user-invocable: true
---

# Submit Workflow

Post-implementation submission pipeline: quality enforcement, learning notes, commit, and PR creation.

## Parameters

- **issue** (optional): GitHub issue number (e.g., `#12` or `12`). Auto-detected from branch name if pattern `issue-\d+` exists.
- **base_branch** (optional): Target branch for PR. Defaults to `main`.

## Workflow

### Phase 1: Pre-flight

1. **Detect issue number**
   - Check if user provided `issue`
   - Otherwise extract from branch name (pattern: `issue-(\d+)`)
   - If not found, ask user (allow "none" for no linked issue)

2. **Fetch issue details** (if issue found)
   ```bash
   gh issue view <number> --json title,body,labels
   ```

3. **Review changes**
   - `git status` — see changed/untracked files
   - `git diff` and `git diff --staged` — review content
   - `git log {base_branch}..HEAD` — existing commits on branch
   - Summarize changes to user before proceeding

### Phase 2: Quality Enforcement

**Mandatory — never skip.**

Run sequentially, stop on first failure:

```bash
cargo fmt --check
cargo clippy -- -D warnings
cargo test
```

- If `fmt` needs changes, run `cargo fmt` and stage results
- If `clippy` fails, fix the issues and re-run
- If tests fail, fix and re-run
- **All three must pass before continuing**

### Phase 3: Learning Notes

**Mandatory — every PR must include a learning document.**

Generate `docs/learn/NNNN-<slug>.md` where `NNNN` is a zero-padded sequence number and `<slug>` summarizes the PR topic.

#### Template

```markdown
# <Title matching PR topic>

PR: #{number} (or "pending" if not yet created)
Date: YYYY-MM-DD

## What was done

<1-3 bullet summary of the implementation>

## Concepts learned

### Rust
- <Language features, patterns, or idioms used for the first time or reinforced>
- <Ownership / lifetime / async nuances encountered>

### Architecture / Design
- <Backend.AI architecture concepts touched (e.g., Manager-Agent split, job lifecycle)>
- <Design patterns applied (e.g., trait-based DI, builder pattern, newtype)>

### Systems / Infra
- <Database, Redis, gRPC, networking concepts if relevant>

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| ... | ... | ... |

## Further study

- [ ] <Topic or resource to dive deeper into>
- [ ] <Related Backend.AI code to read: path or link>
```

#### Rules

- Write in **Korean** (이 문서는 학습용이므로 한국어로 작성)
- Keep it concise — focus on *what you learned*, not restating the code
- Link to relevant source files, docs, or external references
- The "Further study" checklist should be actionable — specific topics, not vague
- Present the draft to the user for review before committing

### Phase 4: Commit

1. **Stage changes** (including the learning doc)
   - `git add` specific files — avoid `-A`
   - Never stage `.env`, credentials, or other sensitive files

2. **Commit message**
   - Conventional commit style: `type(scope): description`
   - Types: `feat`, `fix`, `refactor`, `test`, `docs`, `ci`, `chore`, `perf`
   - Keep first line under 80 characters

3. **Create commit** — present draft message to user for approval

### Phase 5: PR Creation

1. **Push branch**
   ```bash
   git push -u origin {branch_name}
   ```

2. **Create PR**

   If a linked issue exists, the PR body must include the issue context and how it was resolved:

   ```bash
   gh pr create --title "type(scope): description" --body "$(cat <<'EOF'
   ## Issue

   Resolves #<number>

   **Problem**: <1-2 sentence summary of the issue>

   ## Solution

   <How the issue was resolved — approach taken, key changes>

   ## Summary
   <1-3 bullet points of what changed>

   ## What I learned
   <1-2 sentences linking to the learning doc>

   ## Test plan
   - [ ] <test items>
   EOF
   )"
   ```

   If no linked issue:

   ```bash
   gh pr create --title "type(scope): description" --body "$(cat <<'EOF'
   ## Summary
   <1-3 bullet points>

   ## What I learned
   <1-2 sentences linking to the learning doc>

   ## Test plan
   - [ ] <test items>
   EOF
   )"
   ```

3. Update the learning doc's PR number if it was "pending"

4. Report PR URL to user

### Phase 6: Summary

```
Submission Complete

  PR:        #{number} - {title}
  URL:       {url}
  Branch:    {branch_name}
  Learn doc: docs/learn/NNNN-<slug>.md

Quality: All passed (fmt, clippy, test)
Commits: {count} commit(s)
```

## Error Handling

### Quality check failure
Fix the issue, re-run the failing check. Never suppress lint or test failures.

### No changes to commit
Report clean working tree. Nothing to submit.

### PR already exists
Report existing PR URL instead of creating a duplicate.
