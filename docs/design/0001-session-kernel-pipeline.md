# 0001. 세션(커널) 생성 파이프라인 설계

- **작성일**: 2026-03-14
- **상태**: Draft (Epic 1만 활성, Epic 2·3은 future work)

---

## 1. 배경: 우리가 만들려는 것

Manager가 사용자로부터 "세션을 만들어줘"라는 요청을 받으면, Agent에 명령하여 **커널(Kernel)** 을 생성한다.

- **세션(Session)**: Manager가 관리하는 논리적 단위. ID, 상태, 메타데이터를 가짐.
- **커널(Kernel)**: Agent가 관리하는 실행 단위. 컨테이너 또는 프로세스의 wrapper.

```
사용자 ──REST──▶ Manager ──gRPC──▶ Agent
                                     │
                                     ▼
                              KernelRuntime (trait)
                             ┌───────┼───────────┐
                             │       │           │
                          Docker   K8s    LocalProcess
                                          (macOS native)
```

런타임은 여러 가지(Docker, K8s, macOS native process 등)를 지원해야 하므로, **추상화 계층이 필요**하다.

---

## 2. 핵심 설계 결정: 무엇을 추상화할 것인가?

세 가지 선택지를 검토했다.

### Option A: KernelRuntime 추상화 (채택)

Agent는 하나의 concrete struct. 내부에 교체 가능한 KernelRuntime trait을 가진다.

```rust
// Agent는 concrete — 모든 환경에서 동일한 코드
struct Agent {
    runtime: Box<dyn KernelRuntime>,  // 이것만 교체
}

// 커널 라이프사이클만 추상화
trait KernelRuntime {
    async fn create(&self, spec: KernelSpec) -> Result<KernelId>;
    async fn destroy(&self, id: KernelId) -> Result<()>;
    async fn status(&self, id: KernelId) -> Result<KernelStatus>;
}

// 구현체는 각각 독립 모듈
struct LocalProcessRuntime { /* ... */ }
struct DockerRuntime { /* ... */ }
struct K8sRuntime { /* ... */ }
```

**장점:**
- Agent의 공통 로직(gRPC 서버, heartbeat, 등록)이 **한 벌**만 존재 → 중복 없음
- 런타임 추가 시 trait만 구현하면 됨 → 기존 코드 수정 불필요
- 첫 구현이 단순 — LocalProcessRuntime 하나만 만들면 Agent가 동작함

**단점:**
- K8s처럼 Agent 자체의 배포·디스커버리 방식이 달라지는 경우를 이 trait만으로는 담지 못함
- 자원 보고(CPU/GPU/메모리) 방식이 런타임마다 다를 수 있는데, 이 축은 별도 추상화 필요

### Option B: Agent 자체를 추상화 (기각)

Agent를 trait으로 만들고, 환경마다 전체 구현체를 따로 둔다.

```rust
trait AgentBehavior {
    async fn create_kernel(&self, spec: KernelSpec) -> Result<KernelId>;
    async fn destroy_kernel(&self, id: KernelId) -> Result<()>;
    async fn status(&self, id: KernelId) -> Result<KernelStatus>;
    async fn register(&self) -> Result<()>;
    async fn heartbeat(&self) -> Result<()>;
    async fn report_resources(&self) -> Result<Resources>;
}

struct LocalAgent { /* ... */ }   // impl AgentBehavior
struct DockerAgent { /* ... */ }  // impl AgentBehavior
struct K8sAgent { /* ... */ }     // impl AgentBehavior
```

**장점:**
- 커널 라이프사이클뿐 아니라 등록, 자원 보고, 디스커버리까지 구현체마다 자유롭게 변경 가능
- 환경 간 차이가 극단적으로 클 때 유연함

**단점:**
- gRPC 서버, heartbeat 같은 공통 로직이 **모든 구현체에 중복**됨
  - default method로 완화 가능하지만, 상태(필드)에 접근해야 하므로 한계 있음
- 현재 런타임 간 차이는 커널 라이프사이클에 집중 → Agent 전체를 추상화하면 **과잉 설계**
- 구현체 수 × Agent 전체 기능 = 유지보수 비용 급증

### Option C: 계층 분리 (향후 확장 시)

관심사별로 trait을 분리한다. 지금은 만들지 않되, Option A에서 자연스럽게 확장 가능.

```rust
struct Agent {
    runtime: Box<dyn KernelRuntime>,      // 커널 라이프사이클
    resource: Box<dyn ResourceProvider>,   // 자원 보고 (CPU/GPU/메모리)
}
```

- A에서 시작해서, 자원 보고 등 새로운 축이 생기면 trait을 추가하는 방식
- Agent는 여전히 concrete, 변하는 축마다 trait 하나

### 최종 판단

| 판단 기준 | A (KernelRuntime) | B (Agent) |
|-----------|-------------------|-----------|
| 런타임 바뀌면 커널 관리만 달라지나? | **적합** | 과잉 |
| 등록/자원보고/디스커버리도 달라지나? | 부족 (C로 확장) | 적합 |
| 공통 로직 중복 최소화 | **유리** | 불리 |
| 첫 구현 단순성 | **유리** | 복잡 |
| 확장 가능성 | C로 자연 확장 | 이미 확장됨 |

**결론: Option A로 시작, 필요 시 Option C로 확장.**

현 시점에서 런타임 간 차이는 커널 라이프사이클(create/destroy/status)에 집중되어 있다.
Agent의 gRPC 서버, heartbeat, 등록 로직은 어떤 런타임을 쓰든 동일하다.
따라서 Agent를 통째로 추상화하는 것은 과잉이고, KernelRuntime만 trait으로 빼는 것이 맞다.
나중에 자원 보고 방식이 런타임마다 달라지면, `ResourceProvider` trait을 추가하면 된다(Option C).

---

## 3. Epic 구조

전체 파이프라인은 3개 Epic으로 나뉜다. **현재는 Epic 1만 활성.**

| Epic | Goal | 상태 |
|------|------|------|
| **Epic 1: Agent Crate 기반 구축** | Agent crate 생성 + KernelRuntime 추상화 + 첫 구현체 | **활성** |
| Epic 2: Manager ↔ Agent gRPC 통신 | 내부 통신 채널 구축 | Future work |
| Epic 3: 세션 관리 REST API | 외부 REST API | Future work |

Epic 2, 3의 상세 내용은 `docs/future-work.md` (gitignore 대상) 참조.

---

## 4. Epic 1: Agent Crate 기반 구축 — 상세

### Goal

Agent binary crate를 생성하고, KernelRuntime trait 위에 LocalProcess 구현체를 만들어 커널을 로컬에서 생성/종료/상태 조회할 수 있게 한다.

### Story 의존 그래프

```
S1 → S2 ─┬→ S3
          └→ S4
```

- S1(crate scaffold)이 선행되어야 Agent 코드를 놓을 곳이 생김
- S2(trait 정의)가 확정되면 S3, S4는 **병렬 진행 가능**

### Stories

#### S1. Agent crate scaffold

| 항목 | 내용 |
|------|------|
| **Goal** | Agent binary crate를 workspace에 추가하고 기본 구조를 잡는다 |
| **Component** | agent |
| **AC** | 1) `cargo build -p agent` 성공 2) `main()`에서 tracing 초기화 후 로그 출력 3) clippy/fmt 통과 |
| **Notes** | Manager의 구조(main.rs, error.rs, app.rs)를 참고. 아직 gRPC 서버 없이 로그만 출력하고 종료해도 됨. |

#### S2. Kernel 공통 타입 및 KernelRuntime trait 정의

| 항목 | 내용 |
|------|------|
| **Goal** | 커널 관련 공통 타입과 KernelRuntime trait을 common crate에 정의한다 |
| **Component** | common |
| **AC** | 1) `KernelId`, `KernelSpec`, `KernelStatus` 타입 정의 (Serialize/Deserialize) 2) `KernelRuntime` trait에 `create`, `destroy`, `status` async 메서드 존재 3) 타입 생성 및 직렬화 단위 테스트 |
| **Notes** | `KernelSpec`은 최소한 실행할 명령어(command)를 포함. `KernelStatus`는 enum으로 `Running`, `Exited`, `Failed` 등. async trait은 Rust 1.75+에서 네이티브 지원되므로 `async-trait` crate 불필요할 수 있음. |

#### S3. LocalProcess 런타임 — create/destroy

| 항목 | 내용 |
|------|------|
| **Goal** | KernelRuntime trait의 LocalProcess 구현체를 만들어 child process를 생성/종료한다 |
| **Component** | agent |
| **AC** | 1) `create()` 호출 시 child process 기동, `KernelId` 반환 2) `destroy()` 호출 시 프로세스 종료 3) 존재하지 않는 커널 destroy 시 에러 반환 |
| **Notes** | `tokio::process::Command` 사용. 프로세스 핸들은 `HashMap<KernelId, Child>`로 관리. 테스트에서는 `sleep` 또는 `cat` 같은 장시간 실행 명령어 사용. |
| **의존** | S1 (crate 존재), S2 (trait 정의) |

#### S4. LocalProcess 런타임 — status 조회

| 항목 | 내용 |
|------|------|
| **Goal** | LocalProcess에 `status()` 메서드를 구현하여 프로세스 생존 여부를 확인한다 |
| **Component** | agent |
| **AC** | 1) 실행 중 커널 → `KernelStatus::Running` 2) 종료된 커널 → `KernelStatus::Exited` 3) 존재하지 않는 커널 ID → 에러 반환 |
| **Notes** | `Child::try_wait()`로 프로세스 상태 확인. Exited 시 exit code도 KernelStatus에 포함하면 좋음. |
| **의존** | S1, S2. S3과는 **병렬 가능** (trait만 있으면 독립 구현 가능) |

### 병렬 가능 구간

| 시점 | 병렬 가능 Story |
|------|----------------|
| S2 완료 후 | **S3, S4 동시 진행** |

---

## 5. 미결 사항

| 질문 | 현재 판단 | 재검토 시점 |
|------|-----------|------------|
| 첫 커널에서 실행할 대상은? | `sleep` 또는 `cat` (테스트용) | Epic 2에서 실제 워크로드 붙일 때 |
| gRPC vs HTTP 내부 통신? | README 기준 gRPC (tonic) | Epic 2 시작 시 |
| KernelRuntime trait을 common에? agent에? | common (Manager도 타입 참조 필요) | Epic 2에서 gRPC proto와 정합 확인 시 |
| async trait 방식? | Rust 네이티브 async fn in trait 시도, 안 되면 async-trait crate | S2 구현 시 |
| 자원 보고 추상화 (Option C) 필요? | 현재 불필요 | K8s/Docker 런타임 추가 시 |
