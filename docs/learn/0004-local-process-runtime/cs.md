# CS

## Zombie Process

Unix 계열 OS에서 child process가 종료되면 exit status를 커널이 보관한다.
parent process가 `wait()` syscall을 호출하여 이 정보를 읽어가야 커널이 프로세스 테이블 엔트리를 제거한다.

Parent가 wait를 호출하지 않으면:
1. Child는 "zombie" 상태로 프로세스 테이블에 남음
2. PID와 exit status만 차지 (메모리/CPU 사용 없음)
3. 하지만 PID 고갈 가능 — 시스템 전체에 영향

Go의 `cmd.Wait()`가 내부적으로 `waitpid` syscall을 호출한다.
이번 구현에서는 background goroutine이 `Wait()`를 호출하여 프로세스 종료 즉시 회수한다.

## Mutual Exclusion (Mutex)

공유 자원(여기서는 프로세스 map)에 대한 동시 접근을 직렬화하는 동기화 기법.

Critical section을 최소화하는 것이 중요하다:
- Lock 범위가 넓으면 → 다른 goroutine이 불필요하게 대기
- Lock 범위가 좁으면 → 동시성 향상

이번 구현에서는 map 조작만 critical section에 포함하고,
프로세스 Kill이나 Wait 같은 I/O 작업은 lock 밖에서 수행한다.

## Signal-based Process Termination

Unix 프로세스 종료 시그널:
- `SIGTERM` (15) — graceful shutdown 요청. 프로세스가 cleanup 후 종료 가능.
- `SIGKILL` (9) — 즉시 종료. 프로세스가 무시할 수 없음.

`os.Process.Kill()`은 `SIGKILL`을 보낸다 — 가장 확실하지만 graceful하지 않다.
향후 graceful shutdown이 필요하면 `SIGTERM` → timeout → `SIGKILL` 패턴을 적용할 수 있다.
