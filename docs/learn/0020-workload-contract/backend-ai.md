# Backend.AI Architecture

## Manager-Agent 분리와 workload contract

nano-backend.ai는 manager와 agent로 분리된다. 공개 실행 계약은 "로컬 Docker 런타임 API"가
아니라 "manager가 스케줄링하고 agent가 실행하는 workload"를 기술해야 한다. 그래서 contract는
Docker SDK 타입·container config를 포함하지 않고, agent가 준비·시작에 필요한 값(`Plan`)과
agent-side 참조(`Ref`)만 정의한다. 구체적 Docker 실행은 agent 내부 substrate의 몫이며 manager가
마주하는 계약이 아니다.

## Run / Spec / Plan 역할 구분

- **Spec**: 불변의 논리적 작업 정의(무엇을 학습할지). API/DB/사용자 관점의 입력.
- **Run**: Spec의 한 실행 인스턴스. 식별자와 lifecycle을 가진 실행 기록.
- **Plan**: manager가 Run/Spec + 배치 결정을 합쳐 agent에게 넘기는 실행 계약. agent가 컨테이너를
  준비/시작하는 데 필요한 값 묶음.

`Plan.Identifiers`는 이 Plan이 어떤 Run/Project/Spec에 속하는지 참조하며, Plan 자신의 id는 아니다
(agent-side 식별자는 `Ref.AgentWorkloadID`).

## Terminal observation은 ScheduleCoordinator의 몫 (Sokovan)

Sokovan 설계 리뷰에 따라, launch 경계는 작업을 trigger/materialize만 한다. 종료 관찰은 blocking
`Wait()`가 아니라 `ScheduleCoordinator`의 reconcile 경로가 담당한다. 그래서 launcher 포트는
`Prepare`/`Start`/`Cleanup`만 두고 `Wait`/`Inspect`/`StreamLogs`/`Remove`/`EnsureImage`는
의도적으로 제외한다. 종료 결과는 별도 관찰/finalization 데이터로 표현된다.

## Resources: 요청 개수 → device index 변환

Spec의 `ResourceOptions`는 CPU cores, memory, GPU **count**(논리적 개수)를 담는다. 반면 agent가
컨테이너를 만들려면 GPU **device index**가 필요하다. 따라서 provisioner가 `GPU.Count`를
agent-local `[]GPUIndex`로 펼친다. CPU/Memory는 `run.CPUOptions`/`run.MemoryOptions`를 그대로
재사용해 변환을 trivial copy로 만든다. 실행 이미지는 Spec에 없고 preset/provisioner가 resolve한다
(후속 #85).

## API 에러 envelope와 errordef 코드

모든 외부 API 응답은 `{ "status", "data" }` 또는 `{ "status", "error": { "code", ... } }` 구조를
쓰며, `error.code`는 안정적인 기계 판독 키다. `errordef`는 이 코드와 HTTP status를 한 곳에서
관리한다. workload prepare 실패는 `image_pull_failed`/`container_create_failed`(502) 코드로
표현되어, 호출자가 raw Docker stderr를 파싱하지 않고도 `error.code`로 분기할 수 있다.
