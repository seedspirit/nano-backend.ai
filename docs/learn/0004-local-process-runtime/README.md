# LocalProcess 런타임 — create/destroy

PR: #46
Date: 2026-03-21

## What was done

- `LocalProcess` struct 구현 — `KernelRuntime` 인터페이스의 첫 번째 concrete 구현체
- `Create`/`Destroy` 메서드로 child process 생성 및 종료 관리
- zombie process 방지를 위한 background reaper goroutine 패턴 적용

## Categories

- [Code Design](./code-design.md)
- [Go Programming](./go.md)
- [CS](./cs.md)

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| `sync.Mutex` + `map` for process tracking | 단순하고 충분한 동시성 보장; rwlock은 read 빈도가 높지 않아 불필요 | `sync.RWMutex`, `sync.Map` |
| Background goroutine에서 `cmd.Wait()` 호출 | zombie 방지 + `done` channel로 프로세스 종료 감지 가능 | 호출 시점에 Wait (cleanup 누락 위험) |
| Kill 실패 시 `done` channel 확인 후 에러 무시 | 이미 종료된 프로세스에 Kill 보내면 에러지만 의미 없음 | 모든 Kill 에러를 전파 (불필요한 에러 노출) |
| `Status`를 stub으로 둠 | 별도 Story(#44)에서 구현 예정; 인터페이스 만족만 보장 | Status 없이 Create/Destroy만 구현 (컴파일 불가) |

## Further study

- [ ] `os/exec.CommandContext`의 context 취소 시 동작 — SIGKILL vs graceful shutdown
- [ ] `sync.Map` vs `sync.Mutex` + `map` 성능 비교 (고빈도 concurrent access 시)
- [ ] Process group 관리 — child process가 spawn한 grandchild 정리 문제
