# Backend.AI Architecture

## 세션과 커널의 관계

Backend.AI에서 **세션(Session)**과 **커널(Kernel)**은 서로 다른 계층의 개념이다:

```
사용자 ──REST──▶ Manager ──gRPC──▶ Agent
                  │                  │
              세션 관리           커널 관리
          (논리적 단위)       (실행 단위)
```

- **세션**: Manager가 관리하는 논리적 단위. ID, 상태, 메타데이터, 접근 권한 등을 가짐
- **커널**: Agent가 관리하는 실행 단위. 실제 프로세스/컨테이너의 wrapper

사용자가 "세션을 만들어줘"라고 요청하면:
1. Manager가 세션 레코드를 생성하고 적절한 Agent를 선택
2. Agent에게 커널 생성을 명령 (gRPC)
3. Agent가 `KernelRuntime.create()`를 호출하여 실제 프로세스/컨테이너를 기동
4. 커널 ID를 Manager에 반환 → 세션에 연결

## KernelRuntime 추상화의 의미

설계 문서(`docs/design/0001-session-kernel-pipeline.md`)에서 Option A를 채택한 핵심 이유:

```
Agent (concrete — 모든 환경에서 동일)
  └── KernelRuntime (trait — 교체 가능)
        ├── LocalProcessRuntime  ← macOS에서 개발/테스트용
        ├── DockerRuntime        ← 실 서비스 운영
        └── K8sRuntime           ← 클라우드 환경
```

- Agent의 gRPC 서버, heartbeat, 등록 로직은 **어떤 런타임을 쓰든 동일**
- 변하는 부분(커널 라이프사이클)만 trait으로 분리 → OCP(Open-Closed Principle) 적용
- 첫 구현은 `LocalProcessRuntime` — `tokio::process::Command`로 자식 프로세스를 관리

## 커널 상태 머신

`KernelStatus` enum이 표현하는 커널의 생명주기:

```
  create()
    │
    ▼
 Running ──destroy()──▶ Exited { code }
    │
    │ (비정상 종료)
    ▼
 Failed { reason }
```

- `Running`: 프로세스가 실행 중
- `Exited { code: i32 }`: 정상/비정상 종료. `code == 0`이면 정상
- `Failed { reason: String }`: 시스템 수준 실패 (프로세스 시작 불가, 리소스 부족 등)

이 상태 모델은 단순화된 버전이다. 실제 Backend.AI에서는 더 세분화된 상태(`Preparing`, `Pulling`, `Running`, `Terminating`, `Terminated`, `Error`, `Cancelled` 등)를 가진다.

참고: `docs/design/0001-session-kernel-pipeline.md` Section 2
