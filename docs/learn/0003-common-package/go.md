# Go Programming

## Named Type (Defined Type)

Go에서 `type X underlying`은 새로운 타입을 정의한다. underlying type의 연산은 사용 가능하지만, 타입 시스템에서 별개로 취급된다.

```go
type KernelID string

var id KernelID = "abc"
var s string = id          // 컴파일 에러
var s string = string(id)  // 명시적 변환 필요
```

Named type에 메서드를 정의할 수 있다:

```go
func (id KernelID) String() string { return string(id) }
```

`String()` 메서드는 `fmt.Stringer` 인터페이스를 만족시켜 `fmt.Printf("%s", id)` 등에서 자동 호출된다.

참고: `internal/common/kernel.go:14-24`

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

참고: `internal/common/kernel.go:32-41`

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

참고: `internal/common/kernel.go:67-92`

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

참고: `internal/common/kernel.go:96-103`

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
