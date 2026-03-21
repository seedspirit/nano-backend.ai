# S4: agent 스캐폴드 포팅

**Type**: Story
**Labels**: `story`, `agent`, `go`
**Milestone**: Epic 2: Rust → Go 마이그레이션

---

## Story

**Epic**: #TBD (Epic: Rust에서 Go로 언어 마이그레이션)
**Component**: agent

### Background
Rust agent는 현재 스캐폴드 상태로, tracing 초기화 후 로그를 출력하고 종료한다. 동일한 수준의 Go 스캐폴드를 구축한다.

### Goal
Agent 바이너리 진입점을 Go로 구현하고, 구조적 로깅을 초기화하며, 에러 타입을 정의한다.

### Acceptance Criteria
- [ ] `cmd/agent/main.go`에서 `slog` 로거 초기화 + "agent started" 로그 출력 후 정상 종료
- [ ] `internal/agent/error.go`에 `AgentError` 타입 정의 (Bind, Serve 에러)
- [ ] `go build ./cmd/agent` 성공, `go vet` 경고 없음

### Affected Code
- `cmd/agent/main.go` (수정)
- `internal/agent/error.go` (신규)

### Design Notes
- `slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))` 패턴
- `AgentError`: Bind/Serve 에러를 `fmt.Errorf`로 wrapping하거나 커스텀 타입 정의
- gRPC 서버는 이 Story 범위 밖 — 로그 출력 후 종료만 구현

### Test Plan
- `go build ./cmd/agent` 성공 확인
- `go vet ./internal/agent/...` 경고 없음
