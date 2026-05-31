# Go Programming

## Opaque value type + 커스텀 JSON marshaler

`ImageRef`, `AgentPath`, `agent.ID`는 내부 필드를 unexported로 두고 생성자만 노출한다.
필드가 unexported이면 `encoding/json`이 기본적으로 직렬화할 수 없으므로, `json.Marshaler`/
`json.Unmarshaler`를 직접 구현해야 한다.

```go
type ImageRef struct {
	registry, repository, tag string // unexported → 기본 JSON 직렬화 안 됨
}

func (r ImageRef) MarshalJSON() ([]byte, error) {
	s, err := encoding.MarshalJSON(r.String())
	return []byte(s), err
}

func (r *ImageRef) UnmarshalJSON(data []byte) error {
	var s string
	if err := encoding.UnmarshalJSON(string(data), &s); err != nil {
		return err
	}
	parsed, err := ParseImageRef(s)
	...
}
```

핵심: `MarshalJSON`은 값 리시버(직렬화는 상태 변경 없음), `UnmarshalJSON`은 포인터 리시버
(수신자를 채워야 함). 내부 직렬화는 공용 헬퍼 `internal/common/encoding`을 재사용한다.
`kernel.ID`도 같은 패턴을 따른다.

## comparable struct → zero 비교로 IsZero 대체

`ImageRef`(string 3개), `AgentPath`(string 1개), `agent.ID`(`uuid.UUID` = `[16]byte`)는 모두
비교 가능(comparable) 타입이다. 따라서 `IsZero()` 메서드를 따로 만들지 않고 zero 리터럴과
직접 비교한다.

```go
if args.Execution.Image == (ImageRef{}) { ... }
if args.Assignment.AgentID == (agent.ID{}) { ... }
```

unexported 필드가 있어도 다른 패키지에서 `workload.ImageRef{}` 같은 zero 리터럴 생성과 비교는
가능하다(필드 접근이 아니므로). slice/map을 포함한 타입(`Execution`, `Resources`)은 비교
불가라 핵심 식별 필드로 판단한다.

## slices.Clone / maps.Clone (표준 라이브러리)

방어적 복사를 직접 구현(`cloneStrings` 등)하는 대신 Go 1.21+ 표준 `slices.Clone`,
`maps.Clone`을 쓴다.

```go
Entrypoint: slices.Clone(args.Entrypoint),
Env:        maps.Clone(args.Env),
GPUs:       slices.Clone(gpus),
```

둘 다 nil 입력에 nil을 반환해 안전하다. `Mount`처럼 포인터를 품지 않은 struct 슬라이스는
`slices.Clone`의 얕은 복사가 곧 완전한 독립 복사가 된다(요소가 값 복사됨).

## golangci-lint: gocritic hugeParam, revive stutter

이 코드베이스는 두 린트를 에러로 다룬다.

- **hugeParam (gocritic)**: 80바이트 초과 값을 파라미터로 받으면 경고. `ExecutionArgs`(112B),
  `Plan`(288B) 등은 포인터로 전달한다 — 기존 `*spec.Spec` 관용과 일치.
  ```go
  func NewExecution(args *ExecutionArgs) (Execution, error)
  func NewPlan(args *PlanArgs) (Plan, error)
  ```
- **stutter (revive exported)**: 패키지명이 타입명 접두어와 겹치면 경고. 패키지 `workload`에서
  `WorkloadPlan`은 `workload.WorkloadPlan`으로 중복되므로 `Plan`으로, 패키지 `agent`에서는
  `AgentID` 대신 `ID`로 둔다 → 호출부 `workload.Plan`, `agent.ID`. (정확히 같은 단어인
  `run.Run`은 예외적으로 통과한다.)
