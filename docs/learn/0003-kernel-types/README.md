# Kernel 공통 타입 및 KernelRuntime trait 정의

PR: #22
Date: 2026-03-15

## What was done

- `KernelID`, `KernelSpec`, `KernelStatus` 타입을 common 크레이트에 정의
- `KernelRuntime` trait으로 커널 라이프사이클 추상화 (create/destroy/status)
- `KernelError` 도메인 에러 타입 정의 및 10개 단위 테스트 작성

## Categories

- [Code Design](./code-design.md)
- [Rust Programming](./rust.md)
- [Backend.AI Architecture](./backend-ai.md)

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| `KernelID`를 newtype(`String` wrapper)으로 정의 | 타입 안전성 확보 — 일반 String과 혼동 방지 | 단순 `String` typedef — 컴파일 타임 구분 불가 |
| `KernelRuntime` trait을 common 크레이트에 배치 | Manager도 타입 참조가 필요하고, trait이 계약(contract) 역할 | agent 크레이트에 배치 — Manager가 참조 불가 |
| Rust 네이티브 `async fn in trait` 사용 | Rust 1.75+에서 지원, 외부 의존성 불필요 | `async-trait` 크레이트 — object safety 자동 지원이지만 불필요한 의존성 추가 |
| `impl Future` 반환 스타일 사용 | `+ Send` bound를 명시적으로 붙일 수 있어 멀티스레드 환경에서 안전 | `async fn` 직접 사용 — Send bound 제어 불가 |

## Further study

- [ ] Rust의 `async fn in trait` vs `impl Future` 반환 패턴의 object safety 차이
- [ ] `dyn KernelRuntime` (trait object) 사용 시 `async-trait` 또는 `trait-variant` 필요성 확인 (S3에서)
- [ ] serde의 `#[serde(tag = "type")]` 등 enum 직렬화 전략 비교
- [ ] Backend.AI 실제 코드에서 커널 상태 머신 구현 확인: [backend.ai/src/ai/backend/agent/kernel.py](https://github.com/lablup/backend.ai)
