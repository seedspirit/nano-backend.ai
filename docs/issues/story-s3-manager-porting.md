# S3: manager 포팅 — HTTP 서버 및 health 엔드포인트

**Type**: Story
**Labels**: `story`, `manager`, `go`
**Milestone**: Epic 2: Rust → Go 마이그레이션

---

## Story

**Epic**: #TBD (Epic: Rust에서 Go로 언어 마이그레이션)
**Component**: manager
**blockedBy**: S2 (common 패키지)

### Background
Rust manager는 axum 기반 HTTP 서버로 `GET /health` 엔드포인트를 제공한다. Go의 `net/http` 표준 라이브러리로 동일한 기능을 재구현한다.

### Goal
Go `net/http` 기반 HTTP 서버를 구축하고, `GET /health` 엔드포인트가 `ApiResponse` JSON을 반환하도록 한다.

### Acceptance Criteria
- [ ] `cmd/manager/main.go`에서 HTTP 서버가 `127.0.0.1:8080`에서 시작, `slog` 로거 초기화
- [ ] `GET /health` → 200 OK + `{"status":"healthy","reason":"manager is running","next_action_hint":"proceed with API requests"}`
- [ ] 알 수 없는 경로 → 404 응답

### Affected Code
- `cmd/manager/main.go` (수정 — 서버 부트스트랩)
- `internal/manager/app.go` (신규 — 라우터 구성)
- `internal/manager/health.go` (신규 — health 핸들러)
- `internal/manager/error.go` (신규 — ManagerError 타입)
- `internal/manager/app_test.go` (신규)
- `internal/manager/health_test.go` (신규)

### Design Notes
- `http.NewServeMux()`로 라우팅 (Go 1.22+ 패턴 매칭 지원: `GET /health`)
- 핸들러 시그니처: `func(w http.ResponseWriter, r *http.Request)`
- `slog.Debug("health check requested")` — tracing 대응
- `ManagerError`는 서버 바인드/서빙 에러만 포함 (핸들러 에러는 각 핸들러 모듈)
- 테스트: `httptest.NewServer` 또는 `httptest.NewRecorder` 사용

### Test Plan
- Integration test: `GET /health` → StatusCode 200 + JSON 응답 검증
- Integration test: `GET /nonexistent` → StatusCode 404
- Unit test: health 핸들러가 올바른 `ApiResponse` 반환 (status="healthy")
