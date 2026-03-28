# Go Programming

## os/exec — Child Process 관리

Go의 `os/exec` 패키지는 외부 프로세스를 실행하는 표준 방법이다.

### exec.Command vs exec.CommandContext

```go
// context 없이 — 취소 불가
cmd := exec.Command("sleep", "3600")

// context 포함 — context 취소 시 프로세스에 SIGKILL 전송
cmd := exec.CommandContext(ctx, "sleep", "3600")
```

처음에는 `exec.Command`를 사용했지만, 코드 리뷰에서 `exec.CommandContext`로 변경했다.

`CommandContext`를 써야 하는 이유:
- context 취소 시 자식 프로세스에 자동으로 SIGKILL이 전송된다
- `Create`의 caller가 timeout context를 넘길 수 있고, 이 경우 프로세스 생성이 context에 바인딩되어야 자연스럽다
- Go의 관례상 context를 받는 함수에서 `_`로 무시하면 context propagation 체인이 끊기므로 지양해야 한다

**교훈**: 함수 시그니처에 `context.Context`가 있으면 반드시 전파하라.
`_`로 무시하는 것은 context 체인을 의도적으로 끊는 것이므로, 그럴 만한 명확한 이유가 없다면 사용해야 한다.

### Start vs Run

- `cmd.Start()` — 프로세스를 시작하고 즉시 반환. 비동기.
- `cmd.Run()` — `Start()` + `Wait()`. 프로세스가 끝날 때까지 블로킹.

커널은 장시간 실행 프로세스이므로 `Start()`를 사용한다.

### Wait와 Zombie 방지

Unix에서 child process가 종료되면 parent가 `wait` syscall로 회수해야 한다.
회수하지 않으면 zombie process가 된다.

Go에서는 `cmd.Wait()`가 이 역할을 한다.
background goroutine에서 Wait를 호출하여 프로세스가 종료되는 즉시 회수한다:

```go
go func() {
    _ = cmd.Wait()
    close(entry.done)
}()
```

`Wait()`는 한 프로세스당 한 번만 호출 가능 — 두 번 호출하면 에러 반환.

## sync.Mutex — Map 보호

Go의 `map`은 concurrent read/write에 안전하지 않다.
여러 goroutine이 동시에 `Create`나 `Destroy`를 호출할 수 있으므로 `sync.Mutex`로 보호한다.

```go
lp.mu.Lock()
lp.processes[id] = entry
lp.mu.Unlock()
```

`defer lp.mu.Unlock()`을 쓸 수도 있지만, lock 범위를 최소화하기 위해 명시적 unlock을 사용했다.
특히 `Destroy`에서는 map 조작 후 `Kill()`이나 `<-entry.done` 같은 blocking 작업이 있으므로,
그 전에 lock을 풀어야 다른 goroutine이 블로킹되지 않는다.

### 상태 변경과 Lock 범위 — Destroy의 삭제 순서

처음에는 `Kill()` 호출 **전에** map에서 엔트리를 삭제했다:

```go
// 잘못된 순서
lp.mu.Lock()
entry, ok := lp.processes[id]
delete(lp.processes, id)   // Kill 전에 삭제
lp.mu.Unlock()

if err := entry.cmd.Process.Kill(); err != nil {
    // Kill 실패 → 에러 반환, 하지만 map에서는 이미 삭제됨
    // → 이 프로세스를 다시 찾을 방법이 없다 (orphan)
}
```

코드 리뷰에서 지적받아 Kill **성공 후에** 삭제하도록 변경했다:

```go
// 올바른 순서
lp.mu.Lock()
entry, ok := lp.processes[id]
lp.mu.Unlock()             // 삭제 없이 unlock

if err := entry.cmd.Process.Kill(); err != nil {
    select {
    case <-entry.done:     // 이미 종료된 경우 → 정상
        lp.mu.Lock()
        delete(lp.processes, id)   // 확인 후 삭제
        lp.mu.Unlock()
        return nil
    default:
        return err         // 진짜 실패 → map에 남아있으므로 재시도 가능
    }
}
<-entry.done
lp.mu.Lock()
delete(lp.processes, id)   // 성공 후 삭제
lp.mu.Unlock()
```

**교훈**: 부수 효과(map 삭제)는 작업 성공이 확인된 후에 수행하라.
실패 시 롤백이 필요 없도록 "확인 후 커밋" 순서를 따르는 것이 안전하다.
이는 데이터베이스의 commit 순서와 같은 원칙이다 — 작업이 완료되기 전에 상태를 변경하면 실패 시 불일치가 발생한다.

## Channel을 Signal로 사용하기

`done chan struct{}`는 데이터 전달이 아니라 "이벤트 발생"만 알린다.
`struct{}`는 메모리를 차지하지 않으며, `close(ch)`로 모든 수신자에게 동시에 신호를 보낸다.

```go
// 기다리기 (blocking)
<-entry.done

// 확인하기 (non-blocking)
select {
case <-entry.done:
    // 종료됨
default:
    // 아직 실행 중
}
```

이 패턴은 Go에서 completion signal로 널리 사용된다.

### 테스트에서 불필요한 time.Sleep 제거

처음 작성한 테스트에 이런 코드가 있었다:

```go
if err := lp.Destroy(ctx, id); err != nil {
    t.Fatalf("unexpected error on destroy: %v", err)
}
// Give the wait goroutine a moment to complete
time.Sleep(50 * time.Millisecond)
```

코드 리뷰에서 지적: `Destroy` 내부에서 이미 `<-entry.done`으로 reaper goroutine 완료를 대기하고 있으므로,
Destroy가 리턴한 시점에는 goroutine이 이미 끝난 상태다. 따라서 추가 Sleep은 불필요하다.

**교훈**: `time.Sleep`을 테스트에 넣기 전에, 테스트 대상 코드의 동기화 메커니즘을 먼저 확인하라.
이미 채널이나 WaitGroup으로 동기화가 보장되어 있다면 Sleep은 의미 없는 코드이며,
오히려 "이 코드가 제대로 동기화되지 않았다"는 잘못된 인상을 줄 수 있다.
`time.Sleep` 기반 동기화는 CI에서 flaky test의 주범이기도 하다.

## Compile-time Interface Verification

Go는 implicit interface satisfaction을 사용하므로, 구현이 인터페이스를 만족하는지
런타임이 아닌 컴파일 타임에 확인하는 관용구가 있다:

```go
var _ common.KernelRuntime = (*LocalProcess)(nil)
```

`nil` 포인터를 인터페이스 타입 변수에 대입한다.
메서드가 빠져있으면 컴파일 에러가 발생한다.
`_`는 변수를 사용하지 않음을 나타내고, 런타임에 영향을 주지 않는다.
