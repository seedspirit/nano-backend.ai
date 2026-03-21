# Go Programming

## Non-blocking Channel Select

Go의 `select`문에 `default` case를 추가하면 non-blocking 동작이 된다.
이 패턴으로 channel의 상태를 즉시 확인할 수 있다.

```go
select {
case <-entry.done:
    // channel이 닫혀있음 → 프로세스 종료
    return common.Exited(entry.cmd.ProcessState.ExitCode()), nil
default:
    // channel이 아직 열려있음 → 프로세스 실행 중
    return common.Running(), nil
}
```

### Blocking vs Non-blocking

| 패턴 | 동작 |
|------|------|
| `<-ch` | channel에서 값을 받을 때까지 블로킹 |
| `select { case <-ch: ... default: ... }` | 즉시 확인 후 반환 (non-blocking) |

`done chan struct{}`가 `close()`되면 `<-done`은 즉시 zero value를 반환한다.
이 특성 덕분에 여러 goroutine이 동시에 종료를 감지할 수 있다.

## cmd.ProcessState와 ExitCode

`exec.Cmd`의 `Wait()` 호출이 완료되면 `cmd.ProcessState` 필드가 채워진다.
이 필드는 `Wait()` 전에는 `nil`이다.

```go
// Wait() 완료 후에만 접근 가능
exitCode := entry.cmd.ProcessState.ExitCode()
```

### ExitCode() 반환값

| 값 | 의미 |
|----|------|
| `0` | 정상 종료 |
| `1-255` | 비정상 종료 (프로그램이 반환한 코드) |
| `-1` | signal에 의해 종료되었거나 아직 종료되지 않음 |

현재 구현에서는 `done` channel이 닫힌 후에만 `ExitCode()`를 호출하므로
`ProcessState`가 `nil`인 경우는 발생하지 않는다.
reaper goroutine이 `Wait()`를 호출한 후 `close(done)`을 실행하기 때문이다.

## t.Cleanup을 활용한 테스트 리소스 정리

`t.Cleanup()`은 테스트 함수가 완료된 후(패닉 포함) 실행되는 정리 함수를 등록한다.
`defer`와 비슷하지만 테스트 프레임워크에 의해 관리되어 서브테스트에서도 올바르게 동작한다.

```go
func TestStatusRunningProcess(t *testing.T) {
    id, _ := lp.Create(ctx, spec)
    t.Cleanup(func() { _ = lp.Destroy(ctx, id) })
    // 테스트가 실패하더라도 프로세스가 정리됨
}
```

기존 Create/Destroy 테스트에서는 `_ = lp.Destroy(ctx, id)`를 함수 끝에 직접 호출했지만,
Status 테스트에서는 `t.Cleanup`을 사용하여 테스트 실패 시에도 정리가 보장되도록 했다.
