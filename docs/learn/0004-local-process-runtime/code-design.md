# Code Design

## Strategy 패턴 — 인터페이스 구현체로서의 LocalProcess

`KernelRuntime` 인터페이스는 Strategy 패턴의 전형적인 활용이다.
Agent는 concrete struct로 유지하면서, 커널 라이프사이클 관리만 교체 가능하게 추상화한다.

```
Agent (concrete)
  └── KernelRuntime (interface)
        ├── LocalProcess    ← 이번 PR
        ├── DockerRuntime   ← 향후
        └── K8sRuntime      ← 향후
```

`LocalProcess`는 `KernelRuntime`의 3개 메서드를 모두 구현하여 Go의 implicit interface satisfaction을 만족한다.
컴파일 타임 검증은 `var _ common.KernelRuntime = (*LocalProcess)(nil)`로 보장.

**장점**: 새 런타임 추가 시 기존 Agent 코드 변경 없이 구현체만 교체하면 된다.

## Unexported 내부 타입 — processEntry

`processEntry`는 패키지 외부에 노출할 필요가 없는 내부 구현 상세이다.
unexported struct로 선언하여 `LocalProcess`의 구현이 외부 API에 영향 주지 않도록 캡슐화했다.

```go
type processEntry struct {
    cmd  *exec.Cmd
    done chan struct{}
}
```

`done` channel은 프로세스 종료 시점을 알리는 신호 역할을 한다.
이 설계 덕분에 `Destroy`에서 Kill 후 reaper goroutine 완료를 기다릴 수 있고,
향후 `Status`에서도 프로세스 종료 여부를 non-blocking으로 확인 가능하다.

## 에러 컨텍스트 래핑 패턴

모든 에러를 `KernelError`로 래핑하여 operation name과 kernel ID 컨텍스트를 유지한다.
sentinel error를 `%w`로 체이닝하면 호출자가 `errors.Is()`로 에러 유형을 판별할 수 있다.

```go
&common.KernelError{
    Op:  "create",
    Err: fmt.Errorf("empty command: %w", common.ErrKernelRuntime),
}
```

호출자 입장에서는:
- `errors.Is(err, common.ErrKernelRuntime)` → 에러 유형 판별
- `errors.As(err, &ke)` → 구체적 컨텍스트(Op, ID) 추출

## 상태 변경의 원자성 — "확인 후 커밋" 원칙

`Destroy`에서 map 삭제 시점이 중요한 설계 결정이었다.

**잘못된 설계**: Kill 전에 삭제 → Kill 실패 시 프로세스가 map에서 사라져 orphan이 됨
**올바른 설계**: Kill 성공 확인 후 삭제 → 실패 시 map에 남아있어 재시도 가능

이는 "side effect를 되돌릴 수 없다면, 성공이 확정된 후에 수행하라"는 원칙이다.
데이터베이스에서 트랜잭션을 먼저 commit하고 실패하면 롤백할 수 없는 것과 같다.

```
읽기 → 검증 → 실행 → 확인 → 상태 변경(커밋)
```

이 순서는 분산 시스템의 2-phase commit이나 compare-and-swap의 축소판이다.
단일 프로세스에서도 "실패 가능한 작업" 앞에 "되돌릴 수 없는 변경"을 두지 않는 것이 좋은 습관이다.
