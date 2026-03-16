# Rust Programming

## 바이너리 크레이트 vs 라이브러리 크레이트

Rust에서 크레이트(crate)는 컴파일 단위이다. 두 종류가 있다:

- **바이너리 크레이트**: `main.rs`에 `fn main()`이 있어 실행 가능한 프로그램을 생성. `cargo run -p <name>`으로 실행
- **라이브러리 크레이트**: `lib.rs`가 진입점. 다른 크레이트가 `use`로 가져다 쓸 수 있는 코드 제공. 단독 실행 불가

이 프로젝트에서:
- `common` = 라이브러리 크레이트 (`lib.rs`) — 공유 타입 제공
- `manager`, `agent` = 바이너리 크레이트 (`main.rs`) — 각각 독립 실행되는 서비스

`Cargo.toml`에서 `[[bin]]` 섹션으로 바이너리를 명시적으로 선언할 수 있다:

```toml
[[bin]]
name = "agent"
path = "src/main.rs"
```

## Workspace에 크레이트 추가하기

Cargo workspace는 여러 크레이트를 하나의 프로젝트로 묶는 구조이다. 루트 `Cargo.toml`의 `members`에 크레이트 경로를 추가하면 된다:

```toml
[workspace]
members = [
    "crates/common",
    "crates/manager",
    "crates/agent",    # 새로 추가
]
```

workspace의 장점:
- **공유 `Cargo.lock`**: 모든 크레이트가 동일한 의존성 버전 사용 → 호환성 보장
- **공유 `target/`**: 빌드 아티팩트를 한 곳에서 관리 → 디스크 절약, 빌드 캐시 공유
- **`[workspace.dependencies]`**: 의존성 버전을 한 곳에서 관리. 각 크레이트의 `Cargo.toml`에서 `{ workspace = true }`로 참조

```toml
# 루트 Cargo.toml
[workspace.dependencies]
tokio = { version = "1", features = ["full"] }

# 크레이트 Cargo.toml
[dependencies]
tokio = { workspace = true }  # 버전을 여기서 반복하지 않음
```

## `#[tokio::main]` 매크로

Rust의 `async fn main()`은 직접 실행할 수 없다. 비동기 코드를 실행하려면 런타임(runtime)이 필요하다. `#[tokio::main]`은 이를 자동으로 설정해주는 매크로이다:

```rust
#[tokio::main]
async fn main() -> Result<(), AgentError> {
    // 비동기 코드 사용 가능
    Ok(())
}
```

이 매크로는 내부적으로 다음과 같이 변환된다:

```rust
fn main() -> Result<(), AgentError> {
    tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .unwrap()
        .block_on(async {
            // async fn main()의 본문
            Ok(())
        })
}
```

- `new_multi_thread()`: 멀티스레드 런타임 생성 (CPU 코어 수만큼 워커 스레드)
- `enable_all()`: I/O, 타이머 등 모든 기능 활성화
- `block_on()`: 비동기 코드를 동기적으로 실행 (메인 스레드를 블록)

## `tracing_subscriber` 초기화 패턴

`tracing`은 Rust의 구조화된 로깅 프레임워크이다. `tracing-subscriber`는 로그를 실제로 출력하는 구현체이다.

```rust
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

tracing_subscriber::registry()
    .with(EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")))
    .with(tracing_subscriber::fmt::layer())
    .init();
```

각 부분의 역할:
- **`registry()`**: 빈 subscriber를 생성. Layer를 쌓아올릴 기반
- **`EnvFilter`**: `RUST_LOG` 환경변수로 로그 레벨을 제어. 미설정 시 `"info"` 기본값
  - 예: `RUST_LOG=debug cargo run -p agent` → 디버그 레벨까지 출력
  - 예: `RUST_LOG=agent=trace` → agent 크레이트만 trace 레벨
- **`fmt::layer()`**: 로그를 사람이 읽기 좋은 형태로 stdout에 출력
- **`.init()`**: 전역 subscriber로 등록. 프로세스당 한 번만 호출 가능

이 패턴은 프로젝트의 manager 크레이트와 동일하게 사용하여 일관성을 유지한다.

## `thiserror`를 활용한 에러 타입 정의

`thiserror` 크레이트는 `std::error::Error` 트레이트 구현을 자동으로 생성해준다:

```rust
#[derive(Debug, thiserror::Error)]
pub enum AgentError {
    #[error("failed to bind server: {0}")]
    Bind(std::io::Error),

    #[error("server error: {0}")]
    Serve(std::io::Error),
}
```

- `#[derive(thiserror::Error)]`: `std::error::Error`와 `Display` 트레이트를 자동 구현
- `#[error("...")]`: `Display` 출력 형식 지정. `{0}`은 첫 번째 필드를 포맷팅
- `#[from]`: 다른 에러 타입에서 자동 변환 가능 (여기서는 미사용, `CommonError`에서 사용)

`main()`이 `Result<(), AgentError>`를 반환하면, 에러 발생 시 Rust가 자동으로 `Display`를 호출하여 에러 메시지를 출력한다.
