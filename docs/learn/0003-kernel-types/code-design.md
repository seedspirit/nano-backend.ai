# Code Design

## Newtype Pattern

**Newtype**은 기존 타입을 단일 필드 struct로 감싸서 새로운 타입을 만드는 패턴이다.

```rust
pub struct KernelId(pub String);
```

### 왜 `String` 대신 `KernelId`를 쓰는가?

```rust
// ❌ 이렇게 하면 커널 ID와 사용자 ID가 타입 수준에서 구분 불가
fn destroy(kernel_id: String, user_id: String) { ... }
destroy(user_id, kernel_id);  // 인자 순서 뒤바뀌어도 컴파일 통과!

// ✅ Newtype을 쓰면 컴파일러가 잡아줌
fn destroy(kernel_id: KernelId, user_id: UserId) { ... }
destroy(user_id, kernel_id);  // 컴파일 에러!
```

### Newtype의 비용

- **런타임 비용 없음**: Rust 컴파일러가 최적화 시 wrapper를 제거함 (zero-cost abstraction)
- **derive 필요**: 내부 타입의 기능(비교, 해시, 직렬화 등)을 쓰려면 `derive`로 명시 필요

```rust
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct KernelId(pub String);
```

- `Hash` — HashMap의 키로 사용하기 위해 필요 (S3에서 `HashMap<KernelId, Child>` 예정)
- `Display` — 로그 출력, 에러 메시지에서 사용하기 위해 수동 구현

참고: `crates/common/src/kernel.rs`

## Strategy Pattern과 Trait 기반 추상화

`KernelRuntime` trait은 **Strategy Pattern**을 Rust 방식으로 구현한 것이다.

```
Agent (Context)
  └── runtime: impl KernelRuntime (Strategy)
        ├── LocalProcessRuntime
        ├── DockerRuntime
        └── K8sRuntime
```

GoF의 Strategy Pattern에서:
- **Context** = Agent — 공통 로직(gRPC 서버, heartbeat)을 가짐
- **Strategy** = KernelRuntime trait — 교체 가능한 알고리즘 인터페이스
- **ConcreteStrategy** = LocalProcessRuntime 등 — 실제 구현체

### OOP vs Rust의 차이

OOP에서는 Strategy를 인터페이스 + 클래스 상속으로 구현한다. Rust에서는 **trait + 제네릭 또는 trait object**로 구현한다:

```rust
// 제네릭 (정적 디스패치, 컴파일 타임에 결정)
struct Agent<R: KernelRuntime> {
    runtime: R,
}

// trait object (동적 디스패치, 런타임에 결정)
struct Agent {
    runtime: Box<dyn KernelRuntime>,
}
```

현재는 trait만 정의한 상태. 설계 문서(`docs/design/0001-session-kernel-pipeline.md`)에서는 `Box<dyn KernelRuntime>`(동적 디스패치)을 사용하지만, 구현 시 object safety 문제가 있을 수 있어 S3에서 결정 예정.

## 도메인 에러 분리

`KernelError`를 기존 `CommonError`에 넣지 않고 별도로 정의한 이유:

```rust
// CommonError — 인프라 수준 에러 (JSON 파싱 등)
pub enum CommonError {
    Json(serde_json::Error),
}

// KernelError — 도메인 수준 에러 (커널 조작 실패)
pub enum KernelError {
    NotFound(KernelId),
    AlreadyExists(KernelId),
    Runtime(String),
}
```

- **관심사 분리**: 인프라 에러와 비즈니스 에러는 처리 방식이 다름
- **확장성**: 새 도메인(세션, 자원 등)이 생기면 각각 별도 에러 타입을 만듦
- common 크레이트의 CLAUDE.md 규칙: "Error types are infrastructure-level only — domain errors belong in each crate"
  - 다만 `KernelError`는 trait 계약의 일부이므로 trait과 함께 common에 배치

참고: `crates/common/src/kernel.rs`, `crates/common/src/error.rs`
