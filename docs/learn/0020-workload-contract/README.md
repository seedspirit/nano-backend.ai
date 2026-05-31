# Manager-Agent Workload Contract (#67)

PR: pending
Date: 2026-05-31

## What was done

- `internal/common/workload`에 manager-agent 실행 계약의 최소 단위 정의:
  `Plan`(+ `Identifiers`/`Execution`/`Resources`/`Assignment`/`IO`), opaque value type
  `ImageRef`·`AgentPath`, 그리고 agent-side 참조 `Ref`.
- `internal/common/agent`에 UUID 기반 `agent.ID` 신설 (kernel.ID opaque 패턴).
- `errordef`를 `internal/manager`에서 `internal/common`으로 승격하고 prepare 에러 코드
  (`image_pull_failed`, `container_create_failed`)를 추가.

## Categories

- [Code Design](./code-design.md)
- [Go Programming](./go.md)
- [Backend.AI Architecture](./backend-ai.md)

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| `Launcher` 인터페이스를 workload에 두지 않음 | consumer-owned interface 원칙 — 포트는 소비자(스케줄러)가 자기 필요에 맞게 정의 | workload에 포트 정의 → 의존 방향이 거꾸로, 미사용 메서드 노출 위험 |
| `Plan`을 5개 value 그룹으로 분할 | 한 struct에 평면 필드 몰림 방지, 관심사 분리 | flat struct → 필드 폭증, 의미 모호 |
| `ImageRef`/`agent.ID`를 opaque type으로 | 파싱·검증을 한 곳에 모으고 타입 안전 확보 | raw `string`/`uuid.UUID` → 검증 산재, 혼동 |
| GPU를 `[]GPUIndex`로 | Spec의 `GPU.Count`를 provisioner가 device index 여러 개로 변환 | 단일 index → count>1 표현 불가 |
| `errordef`를 common으로 승격 | common 패키지가 manager를 import할 수 없으므로 공통 에러를 공유하려면 이동 필요 | workload 로컬 에러 → 에러 코드 산재 |

## Further study

- [ ] Sokovan `ScheduleCoordinator` reconcile 모델과 terminal observation 분리 설계 읽기
- [ ] `HTTPWorkloadLauncher`(REST) 구현 스토리에서 `Plan` 직렬화 라운드트립 검증 방식
- [ ] 후속 #85: Spec/preset에서 실행 이미지 소스를 정의하고 `ImageRef`로 resolve하는 경로
- [ ] `internal/common/kernel`(process-kernel experiment)과 workload contract의 경계 비교
