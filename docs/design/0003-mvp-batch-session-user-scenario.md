# MVP 사용자 시나리오: 파인튜닝 Batch Session 실행

> 기준: `main` at `e9f943e`  
> 목적: MVP를 사용자 행동 중심으로 정의하고, 필요한 하위 기능과 현재 구현 상태를 추적한다.

## 완료 판정 규칙

- `[x]`: 현재 코드에서 사용자 흐름에 연결되어 동작하며 테스트가 있다.
- `[ ]`: 미구현이거나, 타입·schema·실험 구현만 있어 사용자 흐름에서는 아직 동작하지 않는다.
- 부분 구현은 `[ ]`로 유지하고 현재 존재하는 기반을 괄호 안에 기록한다.

## 대표 사용자 시나리오

> 사용자는 Project에 Axolotl 또는 Unsloth preset 기반 LoRA 파인튜닝
> `batch` Session을 제출한다. 플랫폼은 입력을 검증하고, 사용 가능한 GPU를
> 할당해 Agent의 Docker runtime에서 학습을 실행한다. 사용자는 Session 상태와
> 로그를 polling하고, 종료 후 metric, report, adapter artifact를 확인한다.
> 실패한 경우에도 machine-readable failure reason과 partial artifact를 확인할 수
> 있다.

```text
Project 준비
  → SessionSpecDraft 제출
  → 입력 및 preset 검증
  → immutable SessionSpec + pending Session 저장
  → pending Session scheduling
  → Agent/GPU allocation
  → kernel.CreationSpec 생성
  → Agent에서 Kernel 준비 및 시작
  → 학습 상태와 로그 관찰
  → 종료 결과 reconcile
  → artifact 조회
```

## 1. 사용자가 Project를 준비한다

사용자는 Session과 artifact를 묶을 Project를 API만으로 준비할 수 있어야 한다.

- [x] Project domain type이 있다.
- [x] SQLite에 `projects` table이 있다.
- [x] Session 제출 전에 Project 존재 여부를 확인한다.
- [x] Project 생성 API가 있다.
- [x] Project 단건 조회 API가 있다.
- [x] Project name 중복과 유효성 오류가 stable error code로 반환된다.
- [x] 신규 설치 후 API만으로 첫 Project를 준비할 수 있다.

사용자는 `POST /v1/projects`로 첫 Project를 만들고 `GET /v1/projects/{id}`로
확인한 뒤 Session 제출을 시작할 수 있다.

## 2. 사용자가 파인튜닝 Session을 제출한다

사용자는 Docker 세부사항 없이 trainer preset과 학습 option을 제출해야 한다.

- [x] `POST /v1/sessions` endpoint가 있다.
- [x] request DTO를 `SessionSpecDraft` domain data로 변환한다.
- [x] `batch`, `interactive`, `inference`, `system` Session type이 정의되어 있다.
- [x] Axolotl LoRA SFT preset이 정의되어 있다.
- [x] Unsloth LoRA SFT preset이 정의되어 있다.
- [x] preset에는 image, entrypoint, environment, default parameter가 있다.
- [x] preset ref를 static registry에서 조회한다.
- [ ] Phase 0 제출에서 `type=batch`만 허용한다.
- [ ] trainer preset이 반드시 지정되도록 검증한다.
- [ ] API request에 `idempotency_key`를 받을 수 있다.

## 3. 플랫폼이 실행 전에 입력을 검증한다

잘못된 요청은 queue나 GPU capacity를 사용하기 전에 거부되어야 한다.

- [x] SpecBuilder가 validator 호출 지점을 제공한다.
- [x] TrainerPreset에 parameter type과 min/max policy가 정의되어 있다.
- [x] transport binding 실패를 structured error envelope로 반환한다.
- [ ] 실제 rule-based validator가 애플리케이션에 연결되어 있다.
  - 현재 `validator.Noop`이 모든 candidate를 허용한다.
- [ ] Session name과 Project ID 필수 조건을 검증한다.
- [ ] base model이 비어 있지 않은지 검증한다.
- [ ] dataset이 하나 이상인지 검증한다.
- [ ] dataset split과 asset reference 형식을 검증한다.
- [ ] GPU count가 정확히 1인지 검증한다.
- [ ] memory와 timeout 범위를 검증한다.
- [ ] 허용되지 않은 training parameter key를 거부한다.
- [ ] parameter value type과 min/max 범위를 검증한다.
- [ ] validation failure가 field별 stable detail을 제공한다.

## 4. 플랫폼이 immutable SessionSpec을 만든다

검증된 draft는 재현 가능한 하나의 finalized spec으로 확정되어야 한다.

- [x] preset default를 SessionSpec에 적용한다.
- [x] 사용자 training parameter로 preset default를 override한다.
- [x] model, dataset, resource option을 finalized spec에 반영한다.
- [x] 사용된 preset ref를 provenance로 보존한다.
- [x] 입력 map과 slice를 복사해 외부 mutation을 막는다.
- [x] deterministic canonical JSON helper가 있다.
- [ ] asset reference의 canonical normalization이 있다.
- [ ] canonical spec이 idempotency 비교에 사용된다.
- [ ] finalized spec에서 trainer runtime을 Kernel 실행 정보로 resolve하는 경로가 있다.

## 5. 플랫폼이 pending Session을 안전하게 기록한다

Session 제출이 성공하면 finalized spec과 pending lifecycle이 하나의 Session
record 및 그 하위 데이터로 같은 transaction 안에서 기록되어야 한다.

- [x] Session이 finalized spec을 실행 시점 snapshot으로 직접 소유한다.
- [x] Session은 `status=pending`으로 생성된다.
- [x] Session은 `result=undefined`로 생성된다.
- [x] Session type과 Project/Spec 관계를 저장한다.
- [x] dataset, training parameter, preset ref를 정규화된 table에 저장한다.
- [x] Project의 최근 Session 목록을 조회할 수 있다.
- [x] Session ID로 finalized spec을 조회할 수 있다.
- [ ] Session 단건 전체 representation을 조회할 수 있다.
- [ ] idempotency key를 Session에 저장한다.
- [ ] 같은 key와 같은 spec이면 기존 Session을 반환한다.
- [ ] 같은 key와 다른 spec이면 conflict를 반환한다.
- [ ] 동시 제출 시 idempotency race를 안전하게 처리한다.

현재 사용자 흐름은 여기까지 도달한다. 생성된 Session을 `pending` 이후로 진행시키는
application flow는 아직 없다.

## 6. Scheduler가 pending Session을 선택한다

GPU가 비어 있으면 가장 오래 기다린 Session을 안전하게 claim해야 한다.

- [ ] `SchedulerCoordinator`가 실행된다.
- [ ] pending Session을 FIFO 순서로 조회한다.
- [ ] 하나의 Session을 원자적으로 claim한다.
- [ ] 여러 scheduler tick이 같은 Session을 중복 claim하지 않는다.
- [ ] GPU가 없으면 Session을 pending에 유지한다.
- [ ] Manager 재시작 후 pending/preparing/running Session을 복구한다.
- [ ] scheduler wake-up 누락을 periodic reconcile로 복구한다.

## 7. Provisioner가 Agent와 GPU를 할당한다

Phase 0에서는 단일 Agent의 GPU 0 또는 1 중 하나를 선택하면 된다.

- [x] `agent.ID` value type이 있다.
- [x] `kernel.GPUIndex`와 `kernel.Allocation` type이 있다.
- [x] Allocation 생성자가 Agent ID를 검증하고 GPU slice를 복사한다.
- [ ] 실행 가능한 Agent registry 또는 고정 Agent 설정이 있다.
- [ ] GPU 0/1의 사용 가능 상태를 계산한다.
- [ ] 첫 번째 free GPU를 선택한다.
- [ ] assigned Agent ID와 GPU index를 Session에 저장한다.
- [ ] `preparing` 또는 `running` Session의 GPU 중복 할당을 방지한다.
- [ ] terminal 전환 시 GPU capacity를 release한다.

## 8. 플랫폼이 kernel.CreationSpec을 만든다

SessionSpec과 allocation을 Agent가 실행할 수 있는 완전히 bound된 계약으로
변환해야 한다.

- [x] `kernel.CreationSpec` contract가 있다.
- [x] Session/Project/Spec identifier group이 있다.
- [x] image, entrypoint, command, environment, timeout contract가 있다.
- [x] resource slot과 physical GPU allocation을 구분한다.
- [x] mount와 Agent-visible path contract가 있다.
- [x] image reference와 Agent path value type이 JSON round-trip을 지원한다.
- [ ] SessionSpec과 TrainerPreset에서 execution 정보를 resolve한다.
- [ ] Agent/GPU allocation을 CreationSpec에 결합한다.
- [ ] workspace, cache, config, log, artifact path layout을 결정한다.
- [ ] model과 dataset mount를 생성한다.
- [ ] trainer가 사용할 resolved config를 materialize한다.
- [ ] CreationSpec 생성 실패를 Session failure로 기록한다.

## 9. Manager가 Agent에 Kernel 실행을 요청한다

Manager와 Agent 사이에는 transport-independent Kernel port와 첫 HTTP adapter가
필요하다.

- [ ] Scheduler가 소비하는 `KernelLauncher` interface가 있다.
- [ ] Manager-side HTTP Kernel client가 있다.
- [ ] Agent HTTP server가 실행된다.
- [ ] Kernel prepare/create endpoint가 있다.
- [ ] Kernel start endpoint가 있다.
- [ ] Kernel status endpoint가 있다.
- [ ] Kernel cleanup/destroy endpoint가 있다.
- [ ] `kernel.CreationSpec` request/response round-trip test가 있다.
- [ ] Manager의 timeout과 retry policy가 있다.
- [ ] `kernel.ID`를 Session에 저장한다.

현재 `cmd/agent`는 시작 로그만 출력하고 종료한다.

## 10. Agent가 model과 dataset을 준비한다

Kernel이 시작되기 전에 모든 asset과 runtime configuration이 사용 가능해야 한다.

- [ ] Hugging Face model reference를 resolve한다.
- [ ] Hugging Face dataset reference를 resolve한다.
- [ ] local model/dataset path를 검증한다.
- [ ] host `HF_HOME` cache layout이 있다.
- [ ] cache hit와 miss를 구분한다.
- [ ] model과 dataset을 read-only mount한다.
- [ ] partial download와 staging failure를 정리한다.
- [ ] model download failure를 stable failure reason으로 변환한다.
- [ ] dataset staging failure를 stable failure reason으로 변환한다.

## 11. Agent가 Docker Kernel을 실행한다

Agent runtime은 CreationSpec을 Docker container로 materialize해야 한다.

- [x] Local process create/destroy/status 실험 구현이 있다.
- [x] Local process exit code를 관찰할 수 있다.
- [ ] LocalProcess runtime이 Agent server에 연결되어 있다.
- [ ] Docker runtime 구현이 있다.
- [ ] trainer image를 pull하거나 존재 여부를 확인한다.
- [ ] image, entrypoint, command, environment를 container에 적용한다.
- [ ] bind mount와 working directory를 적용한다.
- [ ] 정확히 하나의 GPU device를 container에 노출한다.
- [ ] timeout을 적용한다.
- [ ] stdout/stderr를 지정된 경로에 보존한다.
- [ ] container exit code, OOM, timeout을 관찰한다.
- [ ] terminal 이후 container를 best-effort로 cleanup한다.

## 12. Manager가 Session lifecycle을 reconcile한다

Kernel 관찰 결과를 Session status, result, failure reason으로 확정해야 한다.

- [x] Session lifecycle state machine이 있다.
- [x] lifecycle status와 terminal result가 분리되어 있다.
- [x] failure transition에 failure reason을 강제한다.
- [x] invalid transition을 거부하는 unit test가 있다.
- [ ] repository에 lifecycle transition을 저장하는 메서드가 있다.
- [ ] `pending → preparing`을 저장한다.
- [ ] Kernel 시작 후 `preparing → running`을 저장한다.
- [ ] exit 0과 required artifact 성공을 `result=success`로 변환한다.
- [ ] non-zero exit, OOM, timeout을 `result=failure`로 변환한다.
- [ ] `started_at`과 `finished_at`을 저장한다.
- [ ] terminal transition과 GPU release를 일관되게 처리한다.
- [ ] cleanup 실패가 Session result를 잘못 덮어쓰지 않는다.

## 13. 플랫폼이 log와 artifact를 보존한다

성공과 실패 모두 사용자가 나중에 검사할 수 있어야 한다.

- [x] `ArtifactIndex`와 `ArtifactFile` domain type이 있다.
- [x] artifact metadata table이 있다.
- [x] artifact index missing error code가 있다.
- [ ] Session별 artifact directory layout을 생성한다.
- [ ] submitted/finalized spec을 artifact로 기록한다.
- [ ] resolved trainer config를 기록한다.
- [ ] stdout과 stderr를 기록한다.
- [ ] metrics와 report를 수집한다.
- [ ] adapter output을 수집한다.
- [ ] 파일 size와 SHA-256 checksum을 계산한다.
- [ ] ArtifactIndex를 DB에 저장한다.
- [ ] required output 누락을 failure로 처리한다.
- [ ] 실패한 실행의 partial artifact를 보존한다.

## 14. 사용자가 완료까지 관찰하고 결과를 확인한다

사용자는 Session ID 하나로 상태, 실패 원인, log, artifact를 조회할 수 있어야 한다.

- [x] Session 제출 응답에 Session summary가 포함된다.
- [x] Project별 최근 Session 목록이 있다.
- [x] Session의 finalized spec을 조회할 수 있다.
- [ ] Session 단건 조회 endpoint가 있다.
- [ ] status, result, failure reason을 polling할 수 있다.
- [ ] allocation과 Kernel 정보를 조회할 수 있다.
- [ ] cursor 기반 stdout/stderr 조회 API가 있다.
- [ ] artifact 목록 API가 있다.
- [ ] artifact download API가 있다.
- [ ] terminal Session 응답에서 log와 artifact로 이동할 수 있다.
- [ ] 성공과 실패 대표 시나리오의 end-to-end test가 있다.

## 현재 도달 지점

```text
Project 생성                                       [x]
  → SessionSpecDraft 제출                         [x]
  → static preset 조회                           [x]
  → validator 호출                               [x]
     └─ 실제 규칙 검증                           [ ]
  → immutable SessionSpec 생성                   [x]
  → pending Session + Spec SQLite 저장           [x]
  → scheduling                                   [ ]
  → Agent/GPU allocation                         [ ]
  → Kernel 실행                                  [ ]
  → reconcile / log / artifact                   [ ]
```

현재 실제 사용자 흐름은 `pending Session 저장`에서 멈춘다.

## MVP Critical Path

아래 순서는 앞 단계가 뒷 단계를 unblock한다.

1. Phase 0 SessionSpec validator
2. submission idempotency
3. scheduling persistence와 GPU allocation
4. `kernel.CreationSpec` builder
5. Manager-Agent Kernel API
6. Docker runtime과 asset staging
7. SchedulerCoordinator reconciliation
8. log와 artifact 수집·조회
9. Session polling API와 end-to-end test

MVP는 이 문서의 모든 미래 확장 항목이 아니라, 대표 사용자 시나리오의 체크박스가
처음부터 끝까지 연결되어 성공·실패 두 경로로 검증될 때 완료된 것으로 본다.
