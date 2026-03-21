# S1: Go 프로젝트 스캐폴드 및 CI 설정

**Type**: Story
**Labels**: `story`, `infra`, `go`
**Milestone**: Epic 2: Rust → Go 마이그레이션

---

## Story

**Epic**: #TBD (Epic: Rust에서 Go로 언어 마이그레이션)
**Component**: infra

### Background
Rust Cargo workspace를 Go 프로젝트 구조로 전환하는 첫 번째 단계. 이후 모든 Story가 이 스캐폴드 위에 코드를 작성하므로 선행 필수.

### Goal
Go module을 초기화하고 manager/agent/common 패키지 디렉토리를 생성하며, CI 파이프라인을 Go 도구 체인으로 전환한다.

### Acceptance Criteria
- [ ] `go build ./...` 성공 (빈 main 패키지 포함)
- [ ] `.github/workflows/ci.yml`이 `go fmt`, `go vet`, `staticcheck`, `go test ./...` 실행
- [ ] 프로젝트 디렉토리 구조: `cmd/manager/`, `cmd/agent/`, `internal/common/`, `internal/manager/`, `internal/agent/`

### Affected Code
- `go.mod` (신규)
- `cmd/manager/main.go` (신규 — 빈 진입점)
- `cmd/agent/main.go` (신규 — 빈 진입점)
- `internal/common/` (신규 — 빈 패키지)
- `internal/manager/` (신규 — 빈 패키지)
- `internal/agent/` (신규 — 빈 패키지)
- `.github/workflows/ci.yml` (수정)

### Design Notes
- Go 단일 모듈 구조 채택 (`github.com/seedspirit/nano-backend.ai`)
- `internal/` 패키지로 Rust crate의 가시성 제어를 재현
- `cmd/` 디렉토리에 바이너리 진입점 배치 (Go 관용 패턴)
- CI에서 `staticcheck`를 clippy 대응으로 사용
- Go 1.22+ 타겟

### Test Plan
- CI 파이프라인이 `go build ./...`를 성공적으로 실행
- `go vet ./...`에서 경고 없음
- `go fmt` 검사 통과 (포맷 차이 없음)
