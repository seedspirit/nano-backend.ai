---
name: analyze
description: Analyze error logs, symptoms, or screenshots by tracing nano-backend.ai code. Reports bugs or explains expected behavior.
user-invocable: true
---

# Analyze

Analyze error situations to determine whether they are bugs, and explain code behavior.

## Input

Accepts various forms of error/symptom reports:

| Input Type | Handling |
|------------|----------|
| **Error logs / panic output** | Extract file paths, function names, and error types from stack traces and goroutine dumps |
| **Plain symptom description** | Identify keywords, then ask follow-up questions for environment and reproduction steps |
| **Screenshots / images** | Read the image via the Read tool, interpret error messages and UI state on screen |
| **Log file paths** | Read the file and extract relevant error/warning spans (`tracing` output) |

If information is insufficient, ask the user follow-up questions. **Ask at most 2 rounds** of questions — if still insufficient, proceed with the available information.

## Workflow

### 1. Symptom Parsing

Extract the following from the input:
- **Symptoms**: The problem observed by the user
- **Error messages / logs**: panic/goroutine dumps, `slog` output, HTTP status, gRPC error codes
- **Environment info**: OS, Go version, configuration
- **Reproduction steps**: if available

### 2. Code Tracing

Explore the codebase to trace the cause of the error.
- Pinpoint the error origin based on error messages, function names, and file paths
- Follow the code flow to estimate the root cause
- Identify affected components:

| Component | Crate / Path |
|-----------|-------------|
| Manager | `cmd/manager/`, `internal/manager/` |
| Agent | `cmd/agent/`, `internal/agent/` |
| Common/Shared | `internal/common/` |
| Database | migrations (`goose`), `jmoiron/sqlx` queries |
| Redis/Valkey | `valkey-io/valkey-glide-go` client code |
| gRPC | `.proto` definitions, gRPC service impls |

### 3. Verdict and Report

Output templates differ by classification:

**If Bug:**

```markdown
## Analysis Result

- **Classification**: Bug
- **Component**: manager / agent / common / ...
- **Severity**: Highest / High / Medium / Low
- **Cause Summary**: ...
- **Related Code**: `cmd/manager/`, `internal/manager/`, etc.
- **Fix Direction**: ...
```

> Prompt: **"Would you like to create an issue? (`gh issue create`)"**

Severity criteria:

| Severity | Criteria |
|----------|----------|
| Highest | Data loss, security vulnerability, system crash / panic |
| High | Core functionality broken, no workaround |
| Medium | Degraded functionality, workaround exists |
| Low | Log message improvements, doc errors, cosmetic issues |

**If Expected Behavior:**

```markdown
## Analysis Result

- **Classification**: Expected Behavior
- **Component**: manager / agent / common / ...
- **Cause Summary**: ...
- **Related Code**: `cmd/manager/`, `internal/manager/`, etc.
- **Correct Usage**: ...
```

**If Inconclusive:**

```markdown
## Analysis Result

- **Classification**: Inconclusive
- **Most Likely Cause**: ...
- **Investigated Code**: <list of paths checked>
- **Missing Information**: <what would help narrow it down>
- **Suggested Next Steps**: ...
```
