# Cargo Workspace 구성과 첫 번째 Health API

PR: #2
Date: 2026-03-14

## What was done

- Cargo workspace로 멀티 크레이트 프로젝트 초기 구성 (`crates/common`, `crates/manager`)
- `nano-common` 크레이트에 `ApiResponse` 공유 타입 정의
- `nano-manager` 크레이트에 axum 기반 `GET /health` 엔드포인트 구현

## Concepts learned

### Rust 기초

- **Cargo workspace**: 루트 `Cargo.toml`에 `[workspace]`를 선언하면 여러 크레이트를 하나의 프로젝트로 관리할 수 있다. `workspace.dependencies`로 의존성 버전을 중앙 관리하면 크레이트 간 버전 불일치를 방지할 수 있다.
  ```toml
  # 루트 Cargo.toml
  [workspace.dependencies]
  tokio = { version = "1", features = ["full"] }

  # 각 크레이트의 Cargo.toml
  [dependencies]
  tokio = { workspace = true }  # 버전을 루트에서 상속
  ```

- **모듈 시스템**: `mod`로 모듈을 선언하고, `pub`으로 외부에 노출한다. `lib.rs`에서 `pub use`로 재수출(re-export)하면 외부에서 `nano_common::ApiResponse`처럼 짧은 경로로 접근 가능.
  ```rust
  // crates/common/src/lib.rs
  pub mod response;
  pub use response::ApiResponse;  // 재수출

  // 외부에서 사용할 때
  use nano_common::ApiResponse;   // response 모듈을 거치지 않고 바로 접근
  ```

- **`impl Into<String>` 패턴**: 함수 파라미터에 `impl Into<String>`을 쓰면 `&str`과 `String` 모두 받을 수 있다. 호출 시 `.into()` 없이 `"문자열 리터럴"`을 직접 넘길 수 있어 편리.
  ```rust
  // 이렇게 정의하면
  pub fn new(status: impl Into<String>) -> Self { ... }

  // 두 가지 다 가능
  ApiResponse::new("healthy");              // &str
  ApiResponse::new(my_string_variable);     // String
  ```

- **Derive 매크로**: `#[derive(Debug, Clone, Serialize, Deserialize)]`로 구조체에 자동으로 트레이트 구현을 붙인다. `Serialize`/`Deserialize`는 serde 크레이트가 제공하며, JSON 변환을 자동 처리.

- **에러 처리 (`thiserror`)**: `#[derive(thiserror::Error)]`로 에러 타입을 선언적으로 정의. `#[error("메시지")]`로 `Display` 구현을, `#[from]`으로 자동 변환을 생성.
  ```rust
  #[derive(Debug, thiserror::Error)]
  pub enum ManagerError {
      #[error("failed to bind server: {0}")]
      Bind(std::io::Error),       // io::Error를 감싸서 컨텍스트 제공
  }
  ```

- **`?` 연산자와 `.map_err()`**: `?`는 `Result`가 `Err`이면 함수에서 즉시 리턴한다. 에러 타입이 다를 때 `.map_err()`로 변환해서 `?`를 쓸 수 있다.
  ```rust
  // TcpListener::bind는 io::Error를 반환하지만
  // 함수 리턴 타입은 Result<(), ManagerError>
  let listener = TcpListener::bind(addr)
      .await
      .map_err(ManagerError::Bind)?;  // io::Error → ManagerError::Bind로 변환
  ```

### axum (웹 프레임워크)

- **Router + Handler 패턴**: `Router::new().route("/path", get(handler_fn))`으로 라우트 등록. 핸들러는 `async fn`이고, 리턴 타입이 `IntoResponse`를 구현하면 된다. `Json<T>`는 자동으로 `Content-Type: application/json` 응답.
  ```rust
  pub async fn check() -> Json<ApiResponse> {
      Json(ApiResponse::new("healthy", "manager is running", "proceed"))
  }
  ```

- **테스트 방법**: `tower::ServiceExt::oneshot()`으로 실제 TCP 없이 라우터를 테스트할 수 있다. 요청을 한 번만 보내고 응답을 받는 패턴.
  ```rust
  let app = build_router();
  let response = app
      .oneshot(Request::builder().uri("/health").body(Body::empty()).unwrap())
      .await
      .unwrap();
  assert_eq!(response.status(), StatusCode::OK);
  ```

### tracing (로깅)

- **`tracing` vs `println!`**: `tracing`은 구조화된 로그를 제공한다. `tracing::info!(%addr, "message")`처럼 필드를 붙이면 JSON 로그 등에서 파싱 가능. `println!`은 비구조화 텍스트라 프로덕션에서 쓰지 않는다.

- **`EnvFilter`**: `RUST_LOG` 환경변수로 로그 레벨을 런타임에 제어. `RUST_LOG=debug cargo run`하면 debug 레벨까지 출력.

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| `crates/common` 분리 | `ApiResponse`는 모든 크레이트가 공유하는 타입. 나중에 agent 크레이트도 사용 예정 | manager 안에 두고 나중에 분리 — 초기 리팩토링 비용 발생 |
| `workspace.dependencies` 사용 | 크레이트 간 의존성 버전 불일치 방지 | 각 크레이트에서 개별 지정 — 버전 드리프트 위험 |
| `thiserror` (not `anyhow`) | 에러 변형이 명확하고 적어서 타입 안전이 유리 | `anyhow` — 바이너리에선 가능하나 에러 종류가 적을 때 과잉 |
| `tower::ServiceExt` 테스트 | axum의 내부 의존성이라 추가 크레이트 불필요 | `axum-test` — 별도 dev-dependency 추가 필요 |

## Further study

- [ ] Rust 소유권(ownership)과 빌림(borrowing) 기본 개념 — [The Book Ch.4](https://doc.rust-lang.org/book/ch04-00-understanding-ownership.html)
- [ ] `async`/`await` 동작 원리와 Tokio 런타임 — [Tokio Tutorial](https://tokio.rs/tokio/tutorial)
- [ ] axum의 extractor 패턴 (`Path`, `Query`, `State`) — 다음 엔드포인트 구현 시 필요
- [ ] `Result`와 `Option` 체이닝 메서드들 (`and_then`, `or_else`, `unwrap_or`) — [std::result](https://doc.rust-lang.org/std/result/)
- [ ] `#[cfg(test)]` 모듈과 통합 테스트(`tests/`) 차이점
