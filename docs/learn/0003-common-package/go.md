# Go Programming

## Struct와 메서드 — class 없는 OOP

Go에는 `class`가 없다. 대신 `struct`로 데이터를 정의하고, 별도의 **메서드**를 붙여서 class처럼 사용한다.

### class 기반 언어와의 비교

```python
# Python — class 안에 모든 것이 들어간다
class LocalProcess:
    def __init__(self):
        self.processes = {}

    def create(self, spec):
        ...

    def destroy(self, id):
        ...
```

```go
// Go — struct는 데이터만, 메서드는 밖에서 붙인다
type LocalProcess struct {
    mu        sync.Mutex
    processes map[common.KernelID]*processEntry
}

func (lp *LocalProcess) Create(_ context.Context, spec common.KernelSpec) (common.KernelID, error) {
    ...
}

func (lp *LocalProcess) Destroy(_ context.Context, id common.KernelID) error {
    ...
}
```

핵심 차이: Go에서 메서드는 struct 블록 안이 아니라 **밖에서** `func (receiver) MethodName(...)` 형태로 선언한다.

### 리시버 (Receiver) — self/this 대신

`func (lp *LocalProcess) Create(...)` 에서 `(lp *LocalProcess)` 부분이 **리시버**다.
Python의 `self`, Java의 `this`에 해당하지만, 이름을 자유롭게 지을 수 있다.
관례적으로 타입 이름의 첫 글자 소문자를 사용한다 (`LocalProcess` → `lp`).

### 값 리시버 vs 포인터 리시버

```go
// 값 리시버 — struct가 복사됨. 원본을 변경할 수 없다.
func (id KernelID) String() string {
    return id.uuid.String()
}

// 포인터 리시버 — 원본을 직접 조작. 상태 변경이 필요할 때.
func (lp *LocalProcess) Create(...) {
    lp.processes[id] = entry  // 원본 map에 추가됨
}
```

**언제 어떤 걸 쓰나:**
- 값 리시버: 읽기 전용 (String(), MarshalJSON() 등)
- 포인터 리시버: 내부 상태를 변경하거나 struct가 클 때 (복사 비용 회피)
- 한 타입의 메서드가 하나라도 포인터 리시버를 쓰면, 일관성을 위해 전부 포인터 리시버를 쓰는 것이 관례

### 생성자 — New 함수

Go에는 `constructor`가 없다. 대신 `New` 접두사 함수를 생성자로 사용하는 관례가 있다:

```go
func NewLocalProcess() *LocalProcess {
    return &LocalProcess{
        processes: make(map[common.KernelID]*processEntry),
    }
}
```

`&LocalProcess{...}`는 struct를 힙에 할당하고 포인터를 반환한다.
map처럼 초기화가 필요한 필드가 있을 때 생성자 함수가 유용하다.
(`make(map[...])` 없이 nil map에 쓰면 런타임 panic이 발생한다.)

### 상속 없음 — 대신 Composition

Go에는 상속이 없다. 코드 재사용은 **임베딩(embedding)**으로 한다:

```go
// 상속이 아니라 "포함"
type Server struct {
    LocalProcess          // embedded — LocalProcess의 메서드를 Server에서 직접 호출 가능
    addr         string
}

server := Server{}
server.Create(ctx, spec) // LocalProcess.Create가 호출됨
```

하지만 이 프로젝트에서는 임베딩 대신 **인터페이스**로 다형성을 구현한다.
`KernelRuntime` 인터페이스를 정의하고, `LocalProcess`가 이를 만족하는 구조다.

참고: `internal/agent/local_process.go`

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

## String Named Type을 사용한 열거형 패턴

Go에는 `enum` 키워드가 없다. 열거형을 표현하는 대표적인 방법이 두 가지 있다.

### 방법 1: `iota` (정수 기반)

```go
type Status int
const (
    StatusRunning Status = iota // 0
    StatusExited                // 1
)
```

- 간결하지만 JSON에 `0`, `1` 같은 정수가 출력됨
- 문자열 표현이 필요하면 `go generate` + `stringer` 도구 필요
- zero value(`0`)가 첫 상수와 겹쳐 미초기화 상태를 구분하기 어려움

### 방법 2: String named type (채택)

```go
type KernelStatusType string

const (
    StatusRunning KernelStatusType = "running"
    StatusExited  KernelStatusType = "exited"
    StatusFailed  KernelStatusType = "failed"
)
```

**장점:**
- JSON에 `"running"`, `"exited"` 같은 읽기 쉬운 문자열 출력 — 별도 마샬링 불필요
- 로그/디버깅 시 값 자체가 의미를 가짐
- zero value가 `""`(빈 문자열)이므로 유효한 상태와 명확히 구분됨
- `stringer` 같은 코드 생성 도구가 불필요

**트레이드오프:**
- 문자열 비교는 정수 비교보다 느리지만, 상태값 수가 적으므로 무시 가능
- 타입 캐스팅(`KernelStatusType("unknown")`)으로 잘못된 값 생성 가능 — 필요 시 검증 함수 추가

참고: `internal/common/kernel.go:69-76`

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

## context.Context — 취소·타임아웃·값 전달

Go에서 거의 모든 함수의 첫 번째 인자로 `ctx context.Context`가 등장한다.
`context.Context`는 **작업의 수명(lifecycle)**을 제어하는 객체다.

### 왜 필요한가?

서버가 요청을 처리하는 도중 클라이언트가 연결을 끊으면, 이미 의미 없는 작업을 계속할 이유가 없다.
context는 "이 작업을 더 이상 할 필요 없다"는 신호를 호출 체인 전체에 전파한다.

### 세 가지 핵심 기능

```go
// 1. 취소 (Cancellation)
ctx, cancel := context.WithCancel(parent)
cancel() // ctx.Done() 채널이 닫힘 → 하위 작업들이 중단 신호를 받음

// 2. 타임아웃 (Timeout)
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()
// 5초 후 자동으로 ctx.Done() 채널이 닫힘

// 3. 값 전달 (Value) — request ID, 인증 정보 등
ctx = context.WithValue(parent, "requestID", "abc-123")
val := ctx.Value("requestID") // "abc-123"
```

### context를 받는 함수에서의 사용 패턴

```go
func doWork(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err() // context.Canceled 또는 context.DeadlineExceeded
    case result := <-longOperation():
        return process(result)
    }
}
```

`ctx.Done()`은 채널을 반환한다. context가 취소되면 이 채널이 닫히므로,
`select`에서 취소 신호를 감지할 수 있다.

### context의 부모-자식 구조

context는 트리 형태로 구성된다. 부모가 취소되면 모든 자식도 취소된다.

```
Background (루트)
├── WithTimeout (API 요청 전체: 30초)
│   ├── WithCancel (DB 쿼리)
│   └── WithCancel (외부 API 호출)
```

부모의 30초 타임아웃이 만료되면 DB 쿼리와 외부 API 호출 모두 취소된다.

### context.Background vs context.TODO

```go
ctx := context.Background() // 루트 context. main이나 테스트에서 시작점으로 사용
ctx := context.TODO()       // "아직 어떤 context를 써야 할지 모르겠다"는 표시
```

둘 다 빈 context를 반환하지만 의도가 다르다.
`Background`는 의도적으로 루트를 만드는 것이고, `TODO`는 나중에 적절한 context로 교체하겠다는 표시다.

### 이 프로젝트에서의 사용

`KernelRuntime` 인터페이스가 모든 메서드에 `ctx`를 요구한다:

```go
type KernelRuntime interface {
    Create(ctx context.Context, spec KernelSpec) (KernelID, error)
    Destroy(ctx context.Context, id KernelID) error
    Status(ctx context.Context, id KernelID) (KernelStatus, error)
}
```

테스트에서는 `context.Background()`를 사용한다:

```go
ctx := context.Background()
id, err := runtime.Create(ctx, spec)
```

### 관용적 규칙

- 함수의 **첫 번째 인자**로 전달한다 (구조체 필드에 저장하지 않는다)
- `nil` context를 전달하지 않는다 — 모르겠으면 `context.TODO()`를 쓴다
- context는 **요청 범위(request-scoped)** 데이터에만 사용한다 — 함수 옵션 전달용이 아니다

참고: `internal/common/kernel.go:134-143`

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
