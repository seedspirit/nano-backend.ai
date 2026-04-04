# ImageRef 타입 및 Docker 클라이언트 인터페이스

PR: #60
Date: 2026-04-04

## What was done

- `ImageRef` opaque 값 객체를 common 패키지에 추가 (registry/repository/tag 파싱·검증)
- `KernelSpec`에 `Image *ImageRef` 필드 추가 (backward compatible)
- `ContainerClient` 인터페이스를 agent 패키지에 정의 (Docker SDK 추상화)

## Categories

- [Code Design](./code-design.md)
- [Go Programming](./go.md)

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| `ImageRef`를 opaque struct으로 구현 | `KernelID`와 동일한 패턴. unexported 필드로 불변성 보장, 팩토리 함수로만 생성 가능 | named string — 검증 누락 위험, 구조 정보 접근 불가 |
| `KernelSpec.Image`를 `*ImageRef` (포인터)로 선언 | nil일 때 JSON에서 omitempty로 필드 자체가 생략됨 → LocalProcess 호환 | 별도 DockerKernelSpec 타입 — KernelRuntime 인터페이스 시그니처 변경 필요 |
| `ContainerClient` 인터페이스로 Docker SDK 감싸기 | 단위 테스트 시 mock 주입 가능, 실제 Docker 데몬 불필요 | SDK 직접 사용 — 테스트마다 Docker 필요 |
| Docker SDK의 container 패키지 타입을 인터페이스에 직접 사용 | 불필요한 중간 타입 정의 방지, SDK 타입이 이미 잘 정의되어 있음 | 자체 타입 정의 — 변환 보일러플레이트 증가 |

## Further study

- [ ] Docker 이미지 참조의 전체 스펙 (digest, `@sha256:...` 형식) — 현재는 tag만 지원
- [ ] `github.com/distribution/reference` 라이브러리 — Docker 공식 이미지 참조 파서
- [ ] Docker SDK `client.NewClientWithOpts()` 패턴과 실제 ContainerClient 구현체 작성 방법
