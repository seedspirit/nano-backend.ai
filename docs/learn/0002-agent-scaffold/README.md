# Agent Crate Scaffold

PR: #20
Date: 2026-03-15

## What was done

- Agent 바이너리 크레이트를 workspace에 추가 (`crates/agent`)
- tracing 초기화 후 info 로그 출력, 정상 종료하는 최소 구조
- `AgentError` enum 정의 (향후 gRPC 서버 에러 대비)

## Categories

- [Rust Programming](./rust.md)

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| HTTP 관련 의존성(axum, serde 등) 제외 | Agent는 gRPC 서버가 될 예정이라 HTTP 프레임워크 불필요 | axum 포함 후 나중에 교체 — 불필요한 의존성 증가 |
| `AgentError`에 `#[allow(dead_code)]` 사용 | Scaffold 단계에서 Bind/Serve variant가 아직 미사용이지만 이슈에서 명시적 요구 | variant 제거 후 나중에 추가 — 이슈 AC 미충족 |
| Manager와 동일한 tracing 초기화 패턴 | 프로젝트 전체 일관성 유지, `RUST_LOG` 환경변수로 로그 레벨 제어 가능 | 단순 `tracing_subscriber::fmt::init()` — EnvFilter 없어 유연성 부족 |

## Further study

- [ ] Rust에서 gRPC 서버 구현: `tonic` 크레이트 학습
- [ ] `tokio::main` 매크로가 내부적으로 하는 일 (런타임 빌더, 스레드 풀 구성)
- [ ] `tracing` vs `log` 크레이트 차이점과 structured logging 개념
- [ ] Cargo workspace에서 크레이트 간 의존성 관리 전략
