# Common 패키지: 커널 타입 및 KernelRuntime 인터페이스

PR: #32
Date: 2026-03-21

## What was done

- Go 프로젝트 구조 초기화 (go.mod, cmd/, internal/, Makefile, .golangci.yml)
- `internal/common` 패키지에 커널 도메인 타입 정의 (KernelID, KernelSpec, KernelStatus, KernelError)
- `KernelRuntime` 인터페이스 정의 — Strategy 패턴으로 런타임 교체 가능한 구조
- `ApiResponse` 표준 응답 구조체 및 팩토리 함수

## Categories

- [Code Design](./code-design.md) — 타입 설계, Strategy 패턴, 에러 래핑
- [Go Programming](./go.md) — named type, iota, sentinel error, 인터페이스

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| `KernelID`를 unexported `uuid.UUID` 필드를 가진 struct로 정의 | 컴파일 타임에 임의 문자열 대입 차단 — `NewKernelID()` 또는 `ParseKernelID()` 생성자만 허용하여 UUID 포맷 강제 | `type KernelID string` named type — 캐스팅으로 우회 가능, UUID 검증 불가 |
| UUID v4 기반 ID 생성 (`google/uuid`) | 분산 환경에서 충돌 없는 고유 식별자 보장 | 순차 정수 — 단일 노드에서만 유효, 분산 환경에서 충돌 |
| `KernelStatus`를 iota enum + struct로 구현 | Go에는 sum type이 없으므로, iota + 관련 필드 struct가 관용적 패턴 | interface 기반 다형성 — Status 같은 단순 값에는 과잉 |
| Sentinel error + wrapping `KernelError` 조합 | `errors.Is()`로 분류 + `KernelError`로 context(Op, ID) 부가 | 단일 에러 타입에 code 필드 — 패턴 매칭이 불편 |
| `KernelRuntime`을 `internal/common`에 배치 | Manager도 이 타입을 참조해야 하므로 공유 패키지에 위치 | agent 패키지 내부 — Manager에서 import 시 순환 의존 발생 |

## Further study

- [ ] `uuid.New()` 내부 구현 — V4 UUID의 엔트로피 소스와 충돌 확률 (Birthday problem)
- [ ] Go의 `errors.Is()` vs `errors.As()` 체이닝 동작 심화 학습
- [ ] `context.Context` 전파 패턴 — KernelRuntime 메서드에서 timeout/cancellation 활용법
- [ ] iota enum의 한계와 대안: `go generate` + `stringer`, 또는 코드 생성 도구
- [ ] Backend.AI 원본 코드의 `KernelId` 구현 비교 (`src/ai/backend/common/types.py`)
