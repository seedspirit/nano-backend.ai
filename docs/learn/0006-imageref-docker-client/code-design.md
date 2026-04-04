# Code Design

## Opaque Value Object 패턴 재활용

`KernelID`에서 확립한 opaque type 패턴을 `ImageRef`에 그대로 적용했다.

핵심 구조:
1. **unexported 필드** — 외부에서 직접 생성 불가
2. **팩토리 함수** (`ParseImageRef`) — 검증을 통과한 값만 생성
3. **값 타입 시맨틱** — 포인터가 아닌 struct로 비교 가능 (`==`)
4. **JSON 직렬화** — canonical 문자열 ��태로 직렬화/역직렬화

```go
type ImageRef struct {
    registry   string  // unexported
    repository string
    tag        string
}
```

`KernelID`는 UUID 하나를 감쌌지만, `ImageRef`는 세 개의 필드(registry, repository, tag)를 가진다. 그러나 패턴의 본질은 동일하다: **생성 시점에 검증하고, 이후에는 유효함을 보장**한다.

## 포인터 필드를 활용한 Optional 표현

Go에는 `Option<T>` 같은 타입이 없다. `KernelSpec.Image`를 optional로 만들기 위해 포인터를 사용했다:

```go
type KernelSpec struct {
    Command []string  `json:"command"`
    Image   *ImageRef `json:"image,omitempty"`
}
```

`*ImageRef`의 장점:
- `nil`은 "이미지 없음" (LocalProcess용)
- `omitempty`와 조합하면 nil일 때 JSON 필드가 아예 생략됨 → 기존 API와 완전히 호환
- 값이 있을 때는 `*ImageRef`를 통해 모든 메서드 접근 가능

대안이었던 별도 `DockerKernelSpec` 타입은 `KernelRuntime.Create(spec KernelSpec)` 시그니처를 변경해야 하므로 기각.

## 인터페이스 경계 설계 — ContainerClient

`ContainerClient`는 Docker SDK의 거대한 `client.Client`에서 컨테이너 라이프사이클에 필요한 5개 메서드�� 추출한 인터페이스다.

```go
type ContainerClient interface {
    CreateContainer(ctx, config, name) (CreateResponse, error)
    StartContainer(ctx, containerID) error
    StopContainer(ctx, containerID, timeout) error
    RemoveContainer(ctx, containerID) error
    InspectContainer(ctx, containerID) (InspectResponse, error)
}
```

이 설계의 핵심 원리:
- **Interface Segregation**: 필요한 메서드만 정의 (Docker SDK에는 수십 개의 메서드가 있음)
- **테스트 용이성**: mock 구현체가 5개 메서드만 구현하면 됨
- **의존 역전**: `DockerRuntime`은 SDK가 아닌 이 인터페이스에 의존

SDK 타입(`container.Config`, `container.InspectResponse`)을 인터페이스 시그니처에 직접 사용한 이유는, 별도 중간 타입을 만들면 변환 로직만 늘어나고 실질적 이점이 없기 때문이다.
