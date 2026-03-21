# Epic: Rust에서 Go로 언어 마이그레이션

**Type**: Epic
**Labels**: `epic`, `migration`, `go`
**Milestone**: Epic 2: Rust → Go 마이그레이션

---

## Epic

### Goal
현재 Rust로 구현된 nano-backend.ai 코드베이스(common, manager, agent)를 Go로 완전히 포팅하여, 동일한 아키텍처와 API 계약을 Go 생태계 위에서 재현한다.

### Motivation
- Go는 빌드 속도, 배포 단순성, 동시성 모델에서 on-premise 환경에 적합
- 러닝 커브가 낮아 AI 에이전트(Claude)가 더 빠르고 정확하게 코드를 생성/수정 가능
- 기존 Backend.AI 생태계(Python)와의 연동 시 Go의 gRPC/HTTP 생태계가 성숙
- Rust의 생산성(컴파일 시간, 복잡한 라이프타임 관리)이 학습 프로젝트 진행 속도를 저해

### Context
현재 코드베이스 상태:
- **Cargo workspace**: `common`, `manager`, `agent` 3개 크레이트
- **common** (`crates/common/`): `ApiResponse`, `CommonError` 공유 타입 (82 LOC)
- **manager** (`crates/manager/`): axum 기반 HTTP 서버, `GET /health` 엔드포인트 (117 LOC)
- **agent** (`crates/agent/`): 스캐폴드만 존재, tracing 초기화 후 종료 (25 LOC)
- **CI** (`.github/workflows/ci.yml`): fmt → clippy → test 파이프라인
- **CLAUDE.md**: Rust 기반 컨벤션, 금지 사항, 워크 분해 원칙
- **Skills** (`.claude/skills/`): `rust-guide`, `tdd-guide`, `submit`, `create-issue`, `autopilot`, `pilot`, `spawn-worker`, `analyze`
- **Design doc** (`docs/design/0001-session-kernel-pipeline.md`): KernelRuntime trait 설계 (Rust 기준)

Go 포팅 시 유지할 것:
- 아키텍처: Manager(control plane) + Agent(execution plane) + 공유 패키지
- API 응답 계약: `{"status", "reason", "next_action_hint"}`
- 테스트 원칙: 모든 공개 함수에 성공/에러 테스트
- 설계 결정: KernelRuntime 인터페이스 추상화 (Option A)

### Stories

| # | Story | Summary | Component | Depends on |
|---|-------|---------|-----------|------------|
| S1 | Go 프로젝트 스캐폴드 및 CI 설정 | Go module 초기화, 디렉토리 구조, CI 파이프라인(fmt/vet/test) 구축 | infra | — |
| S2 | common 패키지 포팅 — ApiResponse 및 에러 타입 | ApiResponse, CommonError를 Go struct/error로 재구현 + 직렬화 테스트 | common | S1 |
| S3 | manager 포팅 — HTTP 서버 및 health 엔드포인트 | Go net/http 기반 서버 + GET /health + 라우터 테스트 | manager | S2 |
| S4 | agent 스캐폴드 포팅 | Agent 바이너리 진입점 + 로깅 초기화 + 에러 타입 | agent | S1 |
| S5 | CLAUDE.md 및 문서 Go 전환 | 루트/서브 CLAUDE.md, README를 Go 컨벤션으로 갱신 | docs | S1 |
| S6 | Skills Go 전환 — go-guide 신규 + 기존 스킬 수정 | rust-guide → go-guide, tdd-guide/submit/autopilot 내 Rust 참조를 Go로 변경 | infra | S5 |
| S7 | 설계 문서 Go 전환 | KernelRuntime 설계를 Go interface로 갱신, 코드 예시 교체 | docs | S2 |

### Dependency Graph

```
S1 ─┬→ S2 ─┬→ S3
    │      └→ S7
    ├→ S4
    └→ S5 → S6
```

- **S2, S4, S5**는 S1 완료 후 **병렬 진행 가능**
- **S3**은 S2(common 패키지) 필요
- **S6**은 S5(문서 전환) 필요
- **S7**은 S2(Go interface 확정) 필요

### Design Decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| 표준 라이브러리 `net/http` + 경량 라우터 사용 | axum 대비 Go 표준 라이브러리가 충분히 강력, 외부 의존성 최소화 | Gin, Echo, Chi — 프로젝트 규모 대비 과잉 |
| `log/slog` (구조적 로깅) 사용 | tracing 대응, Go 1.21+ 표준 라이브러리 | zerolog, zap — 외부 의존성 |
| Go 인터페이스로 KernelRuntime 추상화 | Rust trait과 1:1 대응, Go의 암묵적 인터페이스가 자연스러움 | 코드 생성 기반 추상화 — 불필요한 복잡성 |
| `errors` 패키지 + sentinel errors | thiserror 대응, Go 관용적 에러 처리 | pkg/errors — deprecated 추세 |
| Cargo workspace → Go multi-module 또는 단일 모듈 + 내부 패키지 | Go의 internal/ 패키지 패턴이 crate 가시성 제어와 유사 | Go workspace — 이 규모에서는 과잉 |

### Out of Scope
- gRPC 통신 구현 (Epic 3 이후)
- PostgreSQL/Redis 연동
- Docker/K8s 런타임 구현
- 기존 Rust 코드 삭제 (포팅 검증 후 별도 Story)
- 성능 벤치마킹

### Success Criteria
- `go build ./...` 성공 (manager, agent 바이너리 생성)
- `go test ./...` 전체 통과 (기존 8개 테스트 시나리오 대응)
- `GET /health` 응답이 Rust 버전과 동일한 JSON 구조
- CI 파이프라인(`go fmt`, `go vet`, `staticcheck`, `go test`)이 모두 통과
- CLAUDE.md와 Skills가 Go 기준으로 갱신되어 후속 에이전트 작업에 즉시 활용 가능
