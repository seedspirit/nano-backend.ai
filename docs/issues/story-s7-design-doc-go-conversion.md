# S7: 설계 문서 Go 전환

**Type**: Story
**Labels**: `story`, `documentation`, `go`
**Milestone**: Epic 2: Rust → Go 마이그레이션

---

## Story

**Epic**: #TBD (Epic: Rust에서 Go로 언어 마이그레이션)
**Component**: docs
**blockedBy**: S2 (common 패키지 — Go interface 확정 후)

### Background
`docs/design/0001-session-kernel-pipeline.md`는 KernelRuntime trait 설계를 Rust 코드 예시로 설명한다. Go 포팅 후 설계 문서의 코드가 실제 구현과 괴리되면 혼란을 야기한다.

### Goal
설계 문서 내 Rust 코드 예시를 Go interface/struct 예시로 교체하고, Rust 특화 용어(trait, crate, async fn)를 Go 대응 용어(interface, package, goroutine)로 갱신한다.

### Acceptance Criteria
- [ ] `docs/design/0001-session-kernel-pipeline.md` 내 모든 Rust 코드 블록이 Go 코드 블록으로 교체
- [ ] `KernelRuntime` trait → `KernelRuntime` interface (Go), `LocalProcessRuntime` struct 예시 포함
- [ ] Option A/B/C 비교 표에서 Rust 특화 장단점을 Go 관점으로 재평가하여 갱신

### Affected Code
- `docs/design/0001-session-kernel-pipeline.md` (수정)

### Design Notes
Rust → Go 코드 전환 예시:
```go
// Agent는 concrete — 모든 환경에서 동일한 코드
type Agent struct {
    runtime KernelRuntime // 이것만 교체
}

// 커널 라이프사이클만 추상화
type KernelRuntime interface {
    Create(ctx context.Context, spec KernelSpec) (KernelID, error)
    Destroy(ctx context.Context, id KernelID) error
    Status(ctx context.Context, id KernelID) (KernelStatus, error)
}

// 구현체는 각각 독립 패키지
type LocalProcessRuntime struct { /* ... */ }
type DockerRuntime struct { /* ... */ }
type K8sRuntime struct { /* ... */ }
```

Go에서는 Rust의 async trait 문제가 없으므로(Go interface에 goroutine 제약 없음), 미결 사항 중 "async trait 방식" 항목은 제거.

### Test Plan
- 설계 문서 내 코드 블록이 모두 Go 문법으로 유효한지 확인
- "trait", "crate", "async fn" 등 Rust 특화 용어가 남아있지 않은지 검색
