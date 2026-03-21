# S5: CLAUDE.md 및 문서 Go 전환

**Type**: Story
**Labels**: `story`, `documentation`, `go`
**Milestone**: Epic 2: Rust → Go 마이그레이션

---

## Story

**Epic**: #TBD (Epic: Rust에서 Go로 언어 마이그레이션)
**Component**: docs

### Background
현재 CLAUDE.md(루트 + 서브 디렉토리)와 README.md는 Rust 중심으로 작성되어 있다. AI 에이전트가 Go 코드를 올바르게 생성하려면, 가이드라인이 Go 컨벤션을 반영해야 한다.

### Goal
CLAUDE.md(루트, common, manager)와 README.md를 Go 언어 및 도구 체인 기준으로 갱신한다.

### Acceptance Criteria
- [ ] 루트 CLAUDE.md: 언어를 Go로 변경, 도구(cargo → go), 금지 사항(unwrap → 에러 무시 등) Go 관용 표현으로 갱신
- [ ] README.md: Tech Stack을 Go 기준으로 갱신 (Go, net/http, slog, gRPC 등)
- [ ] 서브 CLAUDE.md(`internal/common/`, `internal/manager/`): 새 디렉토리 경로와 Go 규칙 반영

### Affected Code
- `CLAUDE.md` (수정)
- `README.md` (수정)
- `internal/common/CLAUDE.md` (신규 — `crates/common/CLAUDE.md` 대응)
- `internal/manager/CLAUDE.md` (신규 — `crates/manager/CLAUDE.md` 대응)

### Design Notes
주요 전환 포인트:
| Rust 컨벤션 | Go 컨벤션 |
|-------------|-----------|
| `cargo fmt` | `gofmt` / `goimports` |
| `cargo clippy -- -D warnings` | `go vet` + `staticcheck` |
| `cargo test --all` | `go test ./...` |
| `.unwrap()` / `.expect()` 금지 | `_ = err` (에러 무시) 금지, `if err != nil` 필수 |
| `unsafe` 금지 | `unsafe` 패키지 사용 금지 (동일) |
| `println!` 금지, `tracing` 사용 | `fmt.Println` 금지, `slog` 사용 |
| `thiserror` | `errors.New` / `fmt.Errorf` / sentinel errors |
| `#[cfg(test)] mod tests` | `_test.go` 파일 |

### Test Plan
- CLAUDE.md에서 "Rust", "cargo", "crate" 키워드가 Go 대응 용어로 교체되었는지 확인
- README.md Tech Stack 섹션이 Go 기준으로 갱신되었는지 확인
