# Rust Programming

## Async fn in Trait

Rust 1.75부터 trait에서 `async fn`을 직접 사용할 수 있다. 이전에는 `async-trait` 크레이트가 필수였다.

### 네이티브 방식 (Rust 1.75+)

```rust
trait KernelRuntime {
    async fn create(&self, spec: KernelSpec) -> Result<KernelID, KernelError>;
}
```

이 코드는 내부적으로 다음과 동일하다:

```rust
trait KernelRuntime {
    fn create(&self, spec: KernelSpec) -> impl Future<Output = Result<KernelID, KernelError>>;
}
```

### 문제: Send bound

멀티스레드 비동기 런타임(Tokio)에서는 Future가 스레드 간 이동 가능해야 한다. 즉 `Send`여야 한다. `async fn` 문법으로는 이 bound를 직접 붙일 수 없다:

```rust
// ❌ 이렇게 쓸 수 없음
trait KernelRuntime {
    async fn create(&self, spec: KernelSpec) -> Result<KernelID, KernelError> + Send;
}
```

그래서 `impl Future` 반환 스타일을 사용한다:

```rust
// ✅ Send bound 명시 가능
trait KernelRuntime {
    fn create(
        &self,
        spec: KernelSpec,
    ) -> impl Future<Output = Result<KernelID, KernelError>> + Send;
}
```

이 방식의 장점:
- 외부 크레이트(`async-trait`) 불필요
- `Send` bound를 명시적으로 제어 가능
- 불필요한 `Box` 할당 없음 (async-trait은 내부적으로 `Box::pin` 사용)

### Object Safety 제한

`impl Future`를 반환하는 trait은 **object-safe하지 않다**. 즉 `dyn KernelRuntime`으로 사용할 수 없다:

```rust
// ❌ 컴파일 에러
let runtime: Box<dyn KernelRuntime> = Box::new(LocalProcessRuntime::new());
```

설계 문서에서는 `Box<dyn KernelRuntime>`을 사용하는 것으로 되어 있으나, 이는 S3 구현 시 해결할 문제이다. 해결 방법:
1. `async-trait` 크레이트 사용 → `Box::pin` 기반으로 object-safe하게 변환
2. 제네릭 사용 → `struct Agent<R: KernelRuntime>` (정적 디스패치)
3. `trait-variant` 크레이트 사용 → object-safe 버전의 trait을 자동 생성

참고: [Rust RFC 3185 - Static async fn in traits](https://rust-lang.github.io/rfcs/3185-static-async-fn-in-trait.html)

## `#[derive]`로 구현되는 trait들

`KernelID`, `KernelSpec`, `KernelStatus`에 여러 trait을 derive했다. 각각의 역할:

| Derive | 용도 | 예시 |
|--------|------|------|
| `Debug` | `{:?}` 포맷으로 출력. 디버깅/로깅 필수 | `println!("{:?}", id)` → `KernelID("k-001")` |
| `Clone` | `.clone()`으로 값 복사. 소유권 이동 없이 사본 생성 | `let id2 = id.clone();` |
| `PartialEq`, `Eq` | `==` 비교 가능. `Eq`는 반사성 보장 (자기 자신과 같음) | `assert_eq!(id1, id2)` |
| `Hash` | `HashMap`/`HashSet`의 키로 사용 가능 | `HashMap<KernelID, Child>` |
| `Serialize` | serde로 JSON 등으로 변환 | `serde_json::to_string(&spec)` |
| `Deserialize` | JSON 등에서 Rust 타입으로 역변환 | `serde_json::from_str::<KernelSpec>(json)` |

### `PartialEq` vs `Eq`

- `PartialEq`: 두 값을 비교할 수 있음. `NaN != NaN` 같은 경우를 허용 (float)
- `Eq`: `PartialEq` + 반사성(자기 자신과 항상 같음). `HashMap` 키로 쓰려면 `Eq` 필요
- 정수, 문자열 등은 `Eq` 가능. `f64`는 `NaN` 때문에 `Eq` 불가

## `thiserror`로 에러 타입 정의

`thiserror` 크레이트는 `std::error::Error` + `Display` 구현을 자동 생성한다:

```rust
#[derive(Debug, thiserror::Error)]
pub enum KernelError {
    #[error("kernel not found: {0}")]
    NotFound(KernelID),

    #[error("kernel already exists: {0}")]
    AlreadyExists(KernelID),

    #[error("runtime error: {0}")]
    Runtime(String),
}
```

- `#[error("...")]`: `Display` trait의 `fmt` 메서드를 생성. `{0}`은 첫 번째 필드 참조
- `{0}`이 작동하려면 해당 타입이 `Display`를 구현해야 함 → `KernelID`에 수동으로 `Display` 구현
- `#[from]`을 쓰면 `From` trait도 자동 생성 (이번에는 미사용)

`?` 연산자와 함께 사용:

```rust
fn get_status(id: KernelID) -> Result<KernelStatus, KernelError> {
    let kernel = self.kernels.get(&id)
        .ok_or(KernelError::NotFound(id))?;  // None이면 에러 반환
    Ok(kernel.status())
}
```

참고: [thiserror 공식 문서](https://docs.rs/thiserror/latest/thiserror/)
