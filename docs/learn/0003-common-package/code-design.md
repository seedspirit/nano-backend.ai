# Code Design

## Strategy 패턴: KernelRuntime 인터페이스

`KernelRuntime`은 전형적인 **Strategy 패턴**이다. Agent가 커널 라이프사이클 관리라는 행위를 인터페이스로 추상화하고, 실제 구현(LocalProcess, Docker, K8s)은 런타임에 교체할 수 있다.

```go
// 추상화 — 변하지 않는 계약
type KernelRuntime interface {
    Create(ctx context.Context, spec KernelSpec) (KernelID, error)
    Destroy(ctx context.Context, id KernelID) error
    Status(ctx context.Context, id KernelID) (KernelStatus, error)
}
```

설계 문서(`docs/design/0001-session-kernel-pipeline.md`)에서 세 가지 선택지를 검토했다:

| Option | 내용 | 판단 |
|--------|------|------|
| A. KernelRuntime만 추상화 | Agent는 concrete, 런타임만 교체 | **채택** — 공통 로직 중복 없음 |
| B. Agent 자체를 추상화 | 환경마다 Agent 전체 구현 | 기각 — gRPC/heartbeat 등 공통 로직 중복 |
| C. 관심사별 trait 분리 | Runtime + ResourceProvider 등 | A에서 자연 확장 가능, 현시점 불필요 |

핵심 원칙: **변하는 축(런타임 종류)만 추상화하고, 변하지 않는 것(Agent 공통 로직)은 concrete로 유지.**

참고: `internal/common/kernel.go:134-143`

## Opaque Type 패턴: unexported struct로 불변성 강제

`KernelID`를 unexported 필드를 가진 struct로 정의하여, 반드시 생성자를 통해서만 유효한 값을 만들 수 있도록 강제한다.

```go
// unexported 필드 → 외부 패키지에서 직접 생성 불가
type KernelID struct {
    uuid uuid.UUID
}

// 유일한 생성 경로 ① — 새 UUID 발급
func NewKernelID() KernelID {
    return KernelID{uuid: uuid.New()}
}

// 유일한 생성 경로 ② — 문자열 파싱 (UUID 포맷 검증)
func ParseKernelID(s string) (KernelID, error) {
    id, err := uuid.Parse(s)
    if err != nil {
        return KernelID{}, ErrInvalidKernelID
    }
    return KernelID{uuid: id}, nil
}
```

**named string(`type KernelID string`)과의 차이:**

| 비교 | `type KernelID string` | `struct { uuid uuid.UUID }` |
|------|----------------------|---------------------------|
| 임의 값 대입 | `KernelID("anything")` 가능 | 컴파일 에러 — 외부에서 struct 리터럴 불가 |
| 포맷 검증 | 없음 — 컨벤션으로만 강제 | `ParseKernelID`에서 UUID 검증 |
| zero value | `""` (빈 문자열) | `KernelID{}` — `IsZero()`로 명시적 확인 |
| JSON 직렬화 | 자동 (string) | `MarshalJSON`/`UnmarshalJSON` 직접 구현 필요 |

**트레이드오프:**
- 같은 패키지 내에서는 `KernelID{uuid: ...}` 직접 생성이 가능 — Go는 패키지 단위 접근 제어이므로 모듈 내부에서는 우회 가능
- JSON 커스텀 마샬링 코드가 추가로 필요하지만, 역직렬화 시에도 UUID 검증이 적용되는 이점이 있음

참고: `internal/common/kernel.go:16-62`

## 에러 설계: Sentinel + Wrapping 조합

Go 에러 처리의 두 가지 도구를 조합했다:

**1. Sentinel error — 에러 분류용**

```go
var (
    ErrKernelNotFound      = errors.New("kernel not found")
    ErrKernelAlreadyExists = errors.New("kernel already exists")
    ErrKernelRuntime       = errors.New("kernel runtime error")
)
```

호출자는 `errors.Is(err, ErrKernelNotFound)`로 에러 종류를 판별한다.

**2. KernelError — context 부가용**

```go
type KernelError struct {
    Op  string   // 실패한 연산 (create, destroy, status)
    ID  KernelID // 대상 커널
    Err error    // 근본 원인 (sentinel 또는 다른 에러)
}

func (e *KernelError) Unwrap() error { return e.Err }
```

`Unwrap()`을 구현하여 `errors.Is()`가 체인을 타고 sentinel까지 도달한다.

```go
err := &KernelError{Op: "status", ID: id, Err: ErrKernelNotFound}
errors.Is(err, ErrKernelNotFound) // true
err.Error() // "kernel status abc-123: kernel not found"
```

이 패턴은 Rob Pike의 "Errors are values" 철학과 일치: 에러에 구조를 부여하되, 표준 `error` 인터페이스를 준수한다.

참고: `internal/common/kernel.go:103-132`

## 팩토리 함수 패턴: 상태 생성자

`KernelStatus`는 Go에 sum type이 없기 때문에 struct + 팩토리 함수로 구현했다.

```go
func Running() KernelStatus  { return KernelStatus{Type: StatusRunning} }
func Exited(code int) KernelStatus { return KernelStatus{Type: StatusExited, Code: code} }
func Failed(reason string) KernelStatus { return KernelStatus{Type: StatusFailed, Reason: reason} }
```

**장점:**
- 호출부에서 `Running()`, `Exited(0)` 처럼 의미가 명확
- 각 상태에 필요한 필드만 설정 — `Running()`은 Code/Reason 불필요
- JSON 직렬화 시 `omitempty`로 불필요한 필드 생략

**Rust와의 차이:**
Rust의 `enum KernelStatus { Running, Exited(i32), Failed(String) }`은 각 variant가 다른 데이터를 가질 수 있다. Go에서는 string named type + struct + 팩토리 함수로 이를 근사한다. `KernelStatusType`을 string으로 정의하여 JSON에서도 `"running"`, `"exited"` 등 사람이 읽을 수 있는 값으로 직렬화된다.
