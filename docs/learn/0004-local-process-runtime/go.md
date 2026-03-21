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

이번 구현에서는 `exec.Command`를 사용하고 `Destroy`에서 직접 `Kill()`을 호출한다.
`CommandContext`를 쓰면 context 취소 시 자동으로 프로세스를 종료하지만,
현재 `Create`의 context는 "생성 작업" 자체의 취소를 의미하지 "프로세스 수명"을 의미하지 않기 때문이다.

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

## Compile-time Interface Verification

Go는 implicit interface satisfaction을 사용하므로, 구현이 인터페이스를 만족하는지
런타임이 아닌 컴파일 타임에 확인하는 관용구가 있다:

```go
var _ common.KernelRuntime = (*LocalProcess)(nil)
```

`nil` 포인터를 인터페이스 타입 변수에 대입한다.
메서드가 빠져있으면 컴파일 에러가 발생한다.
`_`는 변수를 사용하지 않음을 나타내고, 런타임에 영향을 주지 않는다.
