# S6: Skills Go 전환 — go-guide 신규 + 기존 스킬 수정

**Type**: Story
**Labels**: `story`, `infra`, `go`
**Milestone**: Epic 2: Rust → Go 마이그레이션

---

## Story

**Epic**: #TBD (Epic: Rust에서 Go로 언어 마이그레이션)
**Component**: infra
**blockedBy**: S5 (문서 전환)

### Background
`.claude/skills/` 아래 스킬들이 Rust 도구(cargo, clippy)와 패턴(trait, crate)을 참조한다. Go로 포팅 후 에이전트가 올바른 도구와 패턴을 사용하려면 스킬도 갱신되어야 한다.

### Goal
`rust-guide`를 `go-guide`로 대체하고, `tdd-guide`, `submit`, `autopilot` 등 기존 스킬 내 Rust 참조를 Go로 변경한다.

### Acceptance Criteria
- [ ] `.claude/skills/go-guide/SKILL.md` 신규 생성 — Go 코딩 컨벤션, 에러 처리, 테스트 패턴, 선호 의존성 포함
- [ ] `tdd-guide`, `submit`, `autopilot` 내 `cargo fmt/clippy/test` → `go fmt/vet/test` 변경, Rust 코드 예시 → Go 코드 예시 변경
- [ ] `analyze` 스킬의 Rust 크레이트 참조를 Go 패키지 경로로 변경

### Affected Code
- `.claude/skills/go-guide/SKILL.md` (신규)
- `.claude/skills/rust-guide/SKILL.md` (삭제 또는 deprecated 표시)
- `.claude/skills/tdd-guide/SKILL.md` (수정)
- `.claude/skills/submit/SKILL.md` (수정)
- `.claude/skills/autopilot/SKILL.md` (수정)
- `.claude/skills/analyze/SKILL.md` (수정)

### Design Notes
go-guide 주요 섹션:
- 에러 처리: `if err != nil` 패턴, sentinel errors, `errors.Is`/`errors.As`
- 타입 설계: struct embedding, 인터페이스 설계 원칙 (작은 인터페이스)
- 네이밍: exported/unexported, 패키지 이름 컨벤션
- 테스트: table-driven tests, `testify` 지양 (표준 라이브러리 선호)
- 로깅: `slog` only
- 린트: `go vet` + `staticcheck`
- 선호 의존성: `net/http`, `slog`, `google.golang.org/grpc`, `database/sql` + `pgx`

### Test Plan
- go-guide 스킬 파일이 존재하고 주요 섹션(에러 처리, 테스트, 로깅)을 포함
- tdd-guide에서 `cargo` 명령어가 더 이상 참조되지 않음
- submit에서 quality 검사가 `go fmt/vet/test`로 변경됨
