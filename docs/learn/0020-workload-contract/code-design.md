# Code Design

## Consumer-owned interface — Launcher를 contract 패키지에 두지 않는다

처음엔 `WorkloadLauncher` 포트를 `internal/common/workload`에 정의했지만, Go의 "인터페이스는
소비자가 정의한다" 원칙에 따라 제거했다. 포트는 그것을 호출하는 쪽(추후 스케줄러)이 자기에게
필요한 메서드만으로 정의하는 게 맞다.

- contract 패키지(`workload`)는 **데이터 타입**(`Plan`, `Ref`, value types)만 제공한다.
- `Launcher`(Prepare/Start/Cleanup)는 스케줄러 패키지에서 정의하고, `HTTPWorkloadLauncher`가
  암묵적으로 만족시킨다.

이렇게 하면 의존 방향이 안정적으로 한 방향(소비자 → 데이터)으로 유지되고, 데이터 패키지가
미사용 동작을 강제로 노출하지 않는다.

## Value object를 관심사별로 그룹핑

`Plan`은 평면 필드를 한 struct에 몰지 않고 5개의 작은 value 그룹으로 나눈다.

```go
type Plan struct {
	Identifiers Identifiers // 소속 Run/Project/Spec
	Execution   Execution   // 무엇을: image, cmd, env, timeout
	Resources   Resources   // 얼마나: cpu, memory, gpu devices
	Assignment  Assignment  // 어디에: agent
	IO          IOBindings  // 입출력: mounts, paths
}
```

각 그룹은 단독으로 이해·테스트 가능하고, 자기 생성자에서 검증·복사 책임을 진다. `Identifiers`는
복수형으로 두어 "워크로드 자체 id"가 아니라 "여러 소속 식별자 묶음"임을 드러낸다.

## Opaque type으로 파싱·검증을 한 곳에 모으기

`ImageRef`는 raw `string`이 아니라 opaque 타입이고, 유일한 생성자 `ParseImageRef`가
`[registry/]repository[:tag]` 규칙과 기본 tag(`latest`)를 한 곳에서 처리한다. REST DTO와
agent-side Docker backend가 같은 타입을 공유하므로 파싱 규칙이 산재하지 않는다. `AgentPath`도
같은 이유로, manager-local 경로와 혼동되지 않도록 별도 타입으로 둔다.

## 불변성: 생성자에서 외부 입력 복사

`Plan`은 immutable 계약이므로, 생성자가 외부에서 받은 slice/map을 복사해 저장한다. 호출자가
나중에 원본을 변경해도 저장된 값은 영향받지 않는다. 테스트는 이 보장 자체를 검증한다(생성 후
입력을 mutate → 저장값 불변 확인).

## errordef의 의존 방향과 공통화

`errordef`는 원래 `internal/manager`에 있었는데, common 패키지는 manager를 import할 수 없다
(공유 패키지 제약). 에러를 errordef로 통일하려면 패키지를 `internal/common`으로 승격해야 한다.
`errordef`는 `common/dto/response`에만 의존하므로 이동이 안전했다. 이후 `ImageRef` 파싱 실패,
`Plan` 필수 필드 누락 등은 모두 `errordef.InvalidInput`으로, prepare 실패는 신규 코드
(`image_pull_failed`/`container_create_failed`)로 표현된다.
