# nano-backend.ai 제품 및 아키텍처 방향

> 상태: Draft
> 현재 범위: Phase 0 — 단일 노드 GPU 기반의 agent-native 모델 개발 기반

## 1. 비전

nano-backend.ai의 최종 목표는 단순한 파인튜닝 실행기가 아니다. AI agent가
모델 개발 목표를 받아 코드를 작성하고, notebook에서 탐색하고, 학습과 평가를
반복하고, 결과물을 비교해 다음 행동을 결정할 수 있는 작은 Backend.AI를 만드는
것이 목표다.

사용자는 Docker 명령이나 GPU 번호가 아니라 다음과 같은 의도를 표현해야 한다.

- 개발 환경을 열고 데이터와 모델을 탐색한다.
- 코드를 제출해 유한한 작업을 실행한다.
- 학습과 평가를 반복하며 가장 좋은 결과를 찾는다.
- 완성된 모델을 추론 환경으로 연다.
- 이전 실행의 입력과 결과를 바탕으로 실험을 재현한다.

플랫폼은 이 의도를 검증 가능한 실행 명세로 바꾸고, 자원을 할당하고, 실행
과정과 결과를 영속적으로 기록한다.

## 2. 현재 단계와 장기 방향

장기 비전은 넓지만 Phase 0은 의도적으로 작게 시작한다.

- 단일 머신, RTX 3090 GPU 2개
- session 하나당 GPU 1개
- preset으로 검증된 LoRA/SFT batch session
- SQLite 기반의 session ledger
- 로컬 파일시스템 기반 artifact 보존
- 최대 2개 session 동시 실행

Phase 0의 파인튜닝은 최종 제품 범위를 제한하는 것이 아니라, session 생성부터
artifact 보존까지 전체 경로를 검증하기 위한 첫 vertical slice다.

## 3. 도메인 모델

```text
Project
  └─ AgentTask
      └─ Experiment
          └─ Session
              └─ Kernel
```

### Project

관련 session, experiment, artifact와 asset을 묶는 namespace다. 장기적으로
사람과 agent가 동일한 모델 개발 문맥을 공유하는 경계가 된다.

### AgentTask

AI agent에게 위임한 장기 목표다. 예를 들어 “이 데이터셋으로 baseline보다 좋은
adapter를 개발하라”와 같은 요청을 나타낸다. 하나의 AgentTask는 여러
Experiment와 Session을 만들 수 있다.

AgentTask는 구체적인 lifecycle과 완료 조건이 정해질 때 도입한다. Phase 0에서는
용어와 책임 경계만 예약한다.

### Experiment

하나의 가설이나 비교 목적 아래 여러 개발 시도와 결과를 묶는다. 학습 session,
평가 session, metric, report, model artifact의 관계를 표현한다.

Experiment 역시 Phase 0의 영속 객체는 아니다. Session에 실험 orchestration
책임을 넣지 않기 위한 상위 경계로 정의한다.

### Session

사용자가 생성하고 관찰하는 compute lifecycle이다. 무엇을 실행할지, 어떤 자원이
필요한지, 어떤 asset과 mount를 사용할지를 immutable SessionSpec으로 표현한다.

Session type은 Backend.AI 용어를 따른다.

- `interactive`: Jupyter, shell, IDE 같은 개발 환경
- `batch`: 학습, 평가, 전처리, 제출 코드처럼 끝이 있는 실행
- `inference`: 모델 serving 환경
- `system`: 플랫폼 내부 작업

Phase 0의 파인튜닝은 `batch` session이다.

### Kernel

Agent가 실제로 생성하는 isolated compute unit이다. 일반적인 구현은 Docker
container지만 process, VM, Kubernetes Pod로도 구현할 수 있다.

Manager는 Session을 관리하고 Agent는 Kernel을 materialize한다. Docker는 Kernel의
한 구현일 뿐 public domain contract가 아니다.

## 4. 핵심 설계 원칙

### Agent-native interface

API는 사람이 읽기 좋은 문장보다 agent가 안정적으로 분기할 수 있는 구조를 먼저
제공한다.

- stable error code
- 명시적인 status와 result
- poll 가능한 long-running resource
- 재시도와 복구를 위한 next-action context
- 결정적인 spec normalization

### 선언적 입력

사용자는 container 생성 방법을 직접 서술하지 않는다. SessionSpecDraft와 preset,
option parameter로 의도를 제출한다. 플랫폼이 이를 검증하고 immutable
SessionSpec으로 확정한다.

### 재현성 우선

모든 terminal session은 성공 여부와 관계없이 입력, resolved configuration, log,
metric, report와 partial output을 보존해야 한다. 실패 역시 재현과 다음 판단에
필요한 결과다.

### infrastructure detail 격리

Scheduler와 service는 Docker SDK나 HTTP DTO를 알지 않는다. Manager와 Agent가
공유하는 Kernel contract는 transport와 runtime에 독립적이어야 한다.

### lifecycle과 outcome 분리

Session status는 현재 lifecycle 위치를, result는 종료 결과를 표현한다.

```text
pending → preparing → running → terminated
                                  └─ success | failure
```

실패한 session도 status는 `terminated`이며 result가 `failure`다. 구체 원인은
`failure_reason`으로 기록한다.

### 작은 시작, 확장 가능한 경계

Phase 0을 위해 범용 scheduler framework나 workflow engine을 만들지 않는다. 대신
Session, Kernel, Runtime, Artifact 경계를 명확하게 유지해 이후 기능이 기존
도메인 의미를 바꾸지 않고 확장되게 한다.

## 5. Spec과 Preset의 역할

### SessionSpecDraft

사용자 또는 agent가 제출한 아직 확정되지 않은 의도다. 일부 값은 생략될 수 있고
preset default가 적용되기 전일 수 있다.

### SessionSpec

검증과 default 적용을 마친 immutable 실행 정의다. 하나의 Session이 이 정의를
실행 시점의 snapshot으로 직접 소유한다. 재현할 때는 확정된 정의를 복사해 새로운
Session을 생성한다.

### Preset

Preset은 단순한 설정 묶음이 아니라 플랫폼과 runtime 사이의 검증된 behavioral
contract다. 어떤 입력을 허용하고, 어떤 default를 적용하며, 어떤 artifact를
생성해야 하는지를 정의한다.

Phase 0에서는 TrainerPreset이 중심이지만 장기적으로 resource policy와 output
policy를 독립적으로 조합할 수 있어야 한다.

Fine-tuning, notebook, evaluation, serving은 서로 다른 최상위 실행 객체가 아니다.
각 목적에 맞는 입력이 SessionSpec으로 resolve되는 구조를 지향한다.

## 6. Session에서 Kernel까지

```text
SessionSpecDraft
  → validation / preset resolution
  → immutable SessionSpec
  → 실행 정의를 포함한 Session persistence
  → scheduling and resource allocation
  → kernel.CreationSpec
  → Agent runtime
  → Kernel
  → observation and artifact reconciliation
```

- Manager는 session lifecycle과 scheduling 결정을 소유한다.
- Provisioner는 실행할 Agent와 GPU device를 선택한다.
- `kernel.CreationSpec`은 배치 결정까지 반영된 manager-agent contract다.
- Agent runtime은 이를 Docker container 또는 다른 substrate로 변환한다.
- 종료 관찰과 session result 확정은 Manager의 reconciliation 흐름으로 돌아온다.

Phase 0의 scheduling은 FIFO와 명시적 GPU allocation이면 충분하다. 공정성,
priority, multi-node placement는 실제 요구가 생긴 이후 추가한다.

## 7. Artifact와 재현성

Artifact는 부가 기능이 아니라 session ledger의 핵심 결과다.

최소한 다음 범주의 정보를 보존한다.

- 제출된 spec과 finalized spec
- runtime이 사용한 resolved configuration
- stdout과 stderr
- 구조화된 metric
- 사람이 읽을 수 있는 report
- 성공 결과 또는 partial output
- 실행 환경과 resource allocation 정보

ArtifactIndex는 파일 위치, 크기, checksum과 같은 metadata를 제공한다. 저장소
구현은 Phase 0에서 로컬 파일시스템이지만, session domain이 특정 storage backend에
의존해서는 안 된다.

## 8. API 방향

사용자-facing API의 중심 resource는 Session이다.

- session spec draft 제출
- session과 finalized spec 조회
- project별 session 탐색
- log와 artifact polling
- 장기적으로 terminate/cancel 요청

Manager-Agent API의 중심 resource는 Kernel이다.

- kernel prepare/create
- kernel start
- kernel status observation
- kernel cleanup/destroy

외부 API와 내부 Agent API는 동일한 DTO를 공유하지 않는다. Domain type은
transport-independent하게 유지한다.

## 9. 오류와 복구 방향

오류는 raw runtime message가 아니라 stable failure taxonomy로 노출한다.

예시:

- image pull failure
- model 또는 dataset staging failure
- resource exhaustion / OOM
- timeout
- trainer failure
- artifact reconciliation failure

모든 failure는 log와 partial artifact를 최대한 보존해야 한다. Agent는
`failure_reason`과 session context를 사용해 재시도, spec 수정, 사용자 보고 중
다음 행동을 선택할 수 있어야 한다.

Idempotency는 네트워크 재시도가 중복 session을 생성하지 않게 한다. 동일한 key와
동일한 normalized spec은 기존 session을 반환하고, 다른 spec은 conflict로
처리한다.

## 10. 단계별 확장

### Phase 0 — Fine-tuning vertical slice

- preset-backed batch session
- 단일 노드 GPU allocation
- session ledger와 기본 조회 API
- Kernel contract와 Docker runtime 방향 확정
- log, metric, model artifact 보존

### Phase 1 — 실행 경로 완성

- Manager-Agent Kernel API
- Docker runtime materialization
- asset staging과 cache
- terminal observation과 reconciliation
- 실패 taxonomy의 실제 runtime mapping

### Phase 2 — 개발 환경과 제어

- interactive notebook/shell session
- submitted-code batch session
- cancel/terminate와 timeout
- orphan kernel cleanup
- richer log and artifact APIs

### Phase 3 — Agent-driven model development

- Experiment lifecycle
- AgentTask와 완료 조건
- training/evaluation comparison loop
- lineage와 model promotion
- agent가 관찰 결과를 바탕으로 다음 session을 계획하는 흐름

### 이후 확장

- inference session과 model serving
- multi-node 또는 multi-GPU session
- resource group과 richer scheduling policy
- remote artifact storage
- multi-tenant quota와 policy

## 11. MVP에서 하지 않는 것

- 범용 workflow engine
- 분산 scheduler framework
- Kubernetes-native orchestration
- multi-tenant quota 및 billing
- 복잡한 priority와 bin-packing
- 실시간 dashboard
- 외부 experiment tracking SaaS에 대한 필수 의존

이 항목들은 영구적인 non-goal이 아니라 Phase 0에서 핵심 경계를 검증하는 데
필요하지 않은 기능이다.

## 12. 성공 기준

Phase 0은 다음 질문에 모두 “예”라고 답할 수 있을 때 성공한 것으로 본다.

- Agent가 Docker 세부사항 없이 training session을 제출할 수 있는가?
- 잘못된 입력이 GPU를 점유하기 전에 검증되는가?
- session 상태와 실패 원인을 API만으로 이해할 수 있는가?
- 모든 terminal session에서 재현에 필요한 입력과 artifact를 찾을 수 있는가?
- Scheduler가 동시에 두 GPU를 안전하게 할당하고 회수할 수 있는가?
- Docker 이외의 runtime이나 training 이외의 session type을 추가해도 핵심 domain을
  다시 정의할 필요가 없는가?

구체적인 타입과 패키지 배치는 코드와 `docs/design/` 문서를 source of truth로
삼는다. 이 문서는 제품과 아키텍처가 향해야 할 방향을 정의한다.
