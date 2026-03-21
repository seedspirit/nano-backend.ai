# S2: common 패키지 포팅 — ApiResponse 및 에러 타입

**Type**: Story
**Labels**: `story`, `common`, `go`
**Milestone**: Epic 2: Rust → Go 마이그레이션

---

## Story

**Epic**: #TBD (Epic: Rust에서 Go로 언어 마이그레이션)
**Component**: common

### Background
Rust `common` 크레이트의 `ApiResponse`와 `CommonError`는 모든 API 응답의 계약이다. Manager와 Agent 포팅 전에 이 공유 타입이 Go로 먼저 전환되어야 한다.

### Goal
`ApiResponse` struct와 에러 타입을 Go로 재구현하고, JSON 직렬화/역직렬화 테스트를 포함한다.

### Acceptance Criteria
- [ ] `internal/common/response.go`에 `ApiResponse` struct 정의 + `NewApiResponse`, `Ok`, `Error` 생성 함수 + JSON 태그
- [ ] `internal/common/error.go`에 `CommonError` 타입 정의 (sentinel error 또는 커스텀 error 타입)
- [ ] 기존 Rust 테스트 5개 시나리오를 Go 테스트로 대응 (생성, Ok/Error 상태, 직렬화, 역직렬화)

### Affected Code
- `internal/common/response.go` (신규)
- `internal/common/response_test.go` (신규)
- `internal/common/error.go` (신규)
- `internal/common/error_test.go` (신규)

### Design Notes
- `ApiResponse` JSON 필드: `status`, `reason`, `next_action_hint` — 기존 계약 유지
- Go의 `encoding/json` 사용, 외부 라이브러리 불필요
- 에러 타입: `fmt.Errorf` wrapping 또는 sentinel error 패턴
- Rust의 `pub use` re-export → Go는 패키지 수준 export로 자연스럽게 대응

### Test Plan
- Unit test: `NewApiResponse`가 모든 필드를 올바르게 설정
- Unit test: `Ok` 함수가 status="ok" 설정
- Unit test: `Error` 함수가 status="error" 설정
- Unit test: `ApiResponse`를 JSON으로 직렬화 시 기대 구조와 일치
- Unit test: JSON 문자열을 `ApiResponse`로 역직렬화 성공
