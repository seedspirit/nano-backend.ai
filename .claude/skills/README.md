# Nano Backend.AI — Claude Code Skills

Skills provide focused guides for AI agents. Invoke with `/skill-name`.

## Available Skills

| Skill | Purpose | Use When |
|-------|---------|----------|
| `/rust-guide` | Rust coding conventions, error handling, type design, async | Writing or reviewing Rust code |
| `/tdd-guide` | TDD workflow (Red → Green → Refactor) | Implementing features or fixing bugs |
| `/submit` | Quality checks, commit, PR creation | Ready to submit changes |

## Workflow

```
rust-guide (always active)
        │
        ▼
   tdd-guide → implement → submit
```

## Related Documents

- `CLAUDE.md` — Root-level agent guidelines
- `docs/design/` — Detailed design documents
