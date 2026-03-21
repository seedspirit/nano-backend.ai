# Go Programming

## Opaque Struct 패턴: unexported 필드로 불변성 강제

Go에서는 struct의 필드를 소문자로 시작하면 패키지 외부에서 접근할 수 없다. 이를 활용하면 **생성자를 통해서만 유효한 값을 만들 수 있는** opaque type을 만들 수 있다.

```go
// uuid 필드가 unexported → 외부에서 KernelID{uuid: ...} 불가
type KernelID struct {
    uuid uuid.UUID
}

func NewKernelID() KernelID              { return KernelID{uuid: uuid.New()} }
func ParseKernelID(s string) (KernelID, error) { ... } // UUID 검증 포함
```

**Named type(`type KernelID string`)과의 비교:**

Named type은 타입 구분은 해주지만 캐스팅으로 우회 가능하다:
```go
type KernelID string
id := KernelID("not-a-uuid") // 컴파일 성공 — 검증 없음
```

Opaque struct는 외부에서 직접 생성이 불가능하다:
```go
type KernelID struct { uuid uuid.UUID }
id := KernelID{uuid: ...} // 컴파일 에러 — uuid가 unexported
```

**`fmt.Stringer` 인터페이스:**

`String()` 메서드를 구현하면 `fmt.Printf("%s", id)` 등에서 자동 호출된다:
```go
func (id KernelID) String() string { return id.uuid.String() }
```

참고: `internal/common/kernel.go:16-62`

## iota를 사용한 열거형 패턴

Go에는 `enum` 키워드가 없다. 대신 `const` 블록에서 `iota`를 사용한다.

```go
type KernelStatusType int

const (
    StatusRunning KernelStatusType = iota // 0
    StatusExited                          // 1
    StatusFailed                          // 2
)
```

**iota 동작:**
- `const` 블록 내에서 0부터 시작, 줄마다 1씩 증가
- 첫 상수에만 타입과 `iota`를 명시하면 나머지는 자동 적용

**주의점:**
- zero value가 `0`이므로 `StatusRunning`이 기본값이 됨 — 초기화되지 않은 상태가 Running으로 해석될 수 있음
- 이를 피하려면 `StatusUnknown = iota`를 0번에 두는 패턴도 있음
- JSON 직렬화 시 정수로 표현됨 — 문자열이 필요하면 `go generate` + `stringer` 사용

참고: `internal/common/kernel.go:70-80`

## Sentinel Error 패턴

패키지 수준 변수로 에러를 미리 정의하고, `errors.Is()`로 비교하는 Go 관용 패턴이다.

```go
var ErrKernelNotFound = errors.New("kernel not found")

// 사용
if errors.Is(err, ErrKernelNotFound) {
    // 404 반환
}
```

**`errors.Is()` vs `==` 비교:**
- `==`는 wrapping된 에러를 찾지 못함
- `errors.Is()`는 `Unwrap()` 체인을 따라가며 비교

```go
wrapped := fmt.Errorf("operation failed: %w", ErrKernelNotFound)
wrapped == ErrKernelNotFound       // false
errors.Is(wrapped, ErrKernelNotFound) // true
```

`%w` verb로 에러를 감싸면 `Unwrap()`이 자동 구현된다. 커스텀 에러 타입에서는 직접 `Unwrap()` 메서드를 구현한다.

참고: `internal/common/kernel.go:103-132`

## Go 인터페이스: 암묵적 구현

Go 인터페이스는 **암묵적(implicit)**으로 구현된다. `implements` 키워드가 없다.

```go
type KernelRuntime interface {
    Create(ctx context.Context, spec KernelSpec) (KernelID, error)
    Destroy(ctx context.Context, id KernelID) error
    Status(ctx context.Context, id KernelID) (KernelStatus, error)
}
```

어떤 타입이든 이 세 메서드를 가지면 자동으로 `KernelRuntime`을 만족한다:

```go
type LocalProcessRuntime struct { ... }

func (r *LocalProcessRuntime) Create(ctx context.Context, spec KernelSpec) (KernelID, error) { ... }
func (r *LocalProcessRuntime) Destroy(ctx context.Context, id KernelID) error { ... }
func (r *LocalProcessRuntime) Status(ctx context.Context, id KernelID) (KernelStatus, error) { ... }
// → 자동으로 KernelRuntime 인터페이스 충족
```

**컴파일 타임 검증 트릭:**

```go
var _ KernelRuntime = (*LocalProcessRuntime)(nil)
```

이 한 줄로 `LocalProcessRuntime`이 인터페이스를 구현하는지 컴파일 타임에 확인할 수 있다.

**Rust trait과의 차이:**
- Rust: `impl KernelRuntime for LocalProcessRuntime` — 명시적 선언 필요
- Go: 메서드 시그니처만 일치하면 자동 충족 — 디커플링에 유리하지만 실수로 인터페이스를 깨뜨릴 수 있음

참고: `internal/common/kernel.go:134-143`

## `context.Context` 전파

`KernelRuntime` 인터페이스의 모든 메서드가 첫 번째 인자로 `context.Context`를 받는다. 이는 Go의 관용적 패턴이다.

```go
Create(ctx context.Context, spec KernelSpec) (KernelID, error)
```

**Context가 전달하는 것:**
- **Cancellation**: 호출자가 취소하면 하위 작업도 중단
- **Timeout/Deadline**: `context.WithTimeout(ctx, 5*time.Second)`
- **Request-scoped values**: trace ID, 인증 정보 등

**규칙:**
- 첫 번째 파라미터로 전달 (변수명 `ctx`)
- struct 필드에 저장하지 않음
- `context.Background()`는 최상위 호출에서만 사용
- `nil` context를 전달하지 않음 — 불확실하면 `context.TODO()` 사용

## 커스텀 JSON 마샬링: `json.Marshaler` / `json.Unmarshaler`

Opaque struct 타입은 기본 JSON 직렬화가 동작하지 않으므로 `MarshalJSON`/`UnmarshalJSON`을 직접 구현해야 한다.

```go
// KernelID → JSON: UUID 문자열로 직렬화
func (id KernelID) MarshalJSON() ([]byte, error) {
    return json.Marshal(id.uuid.String())
}

// JSON → KernelID: UUID 검증 후 역직렬화
func (id *KernelID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    parsed, err := uuid.Parse(s)
    if err != nil {
        return ErrInvalidKernelID
    }
    id.uuid = parsed
    return nil
}
```

**핵심 포인트:**
- `MarshalJSON`은 값 리시버 — struct가 복사되어도 동작
- `UnmarshalJSON`은 포인터 리시버 — 내부 상태를 변경해야 하므로
- 역직렬화 시에도 UUID 검증이 적용되어, 잘못된 JSON 입력을 거부
- JSON 출력은 `"550e8400-e29b-41d4-a716-446655440000"` 형태의 단순 문자열

**struct tag 기반 자동 직렬화와의 차이:**
- struct tag(`json:"field"`)는 필드가 exported일 때 자동으로 동작
- unexported 필드는 `encoding/json`이 접근 불가하므로 커스텀 구현이 필수

참고: `internal/common/kernel.go:45-62`

## JSON 직렬화: struct tag

Go의 struct tag로 JSON 필드 매핑을 제어한다:

```go
type ApiResponse struct {
    Status         string `json:"status"`
    Reason         string `json:"reason"`
    NextActionHint string `json:"next_action_hint"`
}
```

- `json:"field_name"` — JSON 키 이름 지정 (Go는 PascalCase, JSON은 snake_case)
- `json:"code,omitempty"` — zero value이면 JSON 출력에서 생략
- `json:"-"` — 직렬화 대상에서 제외

`KernelStatus`에서 `omitempty` 활용:

```go
type KernelStatus struct {
    Type   KernelStatusType `json:"type"`
    Code   int              `json:"code,omitempty"`   // Running/Failed일 때 생략
    Reason string           `json:"reason,omitempty"` // Running/Exited일 때 생략
}
```

참고: `internal/common/kernel.go:44-48`, `internal/common/response.go:7-11`

## log/slog: 구조화된 로깅

Go 1.21+의 표준 라이브러리 `log/slog`를 사용하여 구조화된 로그를 출력한다.

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

slog.Info("manager started")
```

**slog의 장점 (fmt.Println / log.Println 대비):**
- 키-값 쌍의 구조화된 로그: `slog.Info("request", "method", "GET", "path", "/health")`
- 로그 레벨 지원 (Debug, Info, Warn, Error)
- Handler 교체 가능 — TextHandler(개발), JSONHandler(프로덕션)
- 프로젝트 규칙에서 `fmt.Println` / `log.Println` 사용 금지

참고: `cmd/manager/main.go:8-14`, `cmd/agent/main.go:8-14`
