# Go Programming

## Docker 이미지 참조 파싱 규칙

Docker 이미지 참조는 `[registry/]repository[:tag]` 형식을 따른다. Go에서 이를 파싱할 때 고려해야 할 규칙:

1. **단일 이름** (`nginx`): Docker Hub의 공식 이미지 → `docker.io/library/nginx:latest`
2. **사용자/저장소** (`myuser/myapp`): Docker Hub 사용자 저장소 → `docker.io/myuser/myapp:latest`
3. **커스텀 레지스트리** (`registry.example.com/app`): 첫 세그먼트에 `.` 또는 `:`가 포함되면 레지스트리로 인식

`strings.ContainsAny(first, ".:")` 패턴으로 레지스트리 여부를 판별했다. 이는 Docker CLI가 사용하는 휴리스틱과 동일하다 — 도메인 이름은 반드시 `.`을 포함하고, `localhost:5000` 같은 경우는 `:`를 포함하기 때문.

```go
func splitRegistryRepo(repo string) (registry, repository string) {
    parts := strings.SplitN(repo, "/", 2)
    if len(parts) == 1 {
        return "docker.io", "library/" + repo
    }
    if strings.ContainsAny(parts[0], ".:") {
        return parts[0], parts[1]
    }
    return "docker.io", repo
}
```

## `omitempty`와 포인터 타입의 조합

Go의 `encoding/json`에서 `omitempty` 태그는 "zero value이면 필드를 생략"한다. 하지만 struct 타입의 zero value는 모든 필드가 zero인 struct이지, nil이 아니다. 따라서:

- `Image ImageRef` + `omitempty` → 빈 struct가 `""` 등으로 직렬화될 수 있음
- `Image *ImageRef` + `omitempty` → nil이면 필드 자체가 JSON에서 사라짐

이 차이가 backward compatibility에 중요하다. 기존에 `{"command":["sleep","10"]}` 형태로 직렬화되던 `KernelSpec`이 `image` 필드 없이 그대로 유지되어야 하기 때문.

```go
type KernelSpec struct {
    Command []string  `json:"command"`
    Image   *ImageRef `json:"image,omitempty"`  // nil → JSON에서 완전히 생략
}
```

## Docker SDK의 `+incompatible` 모듈

`go get github.com/docker/docker`를 실행하면 `v28.5.2+incompatible`처럼 `+incompatible` 태그가 붙는다. 이는 Docker SDK가 아직 Go 모듈 체계(`go.mod`)를 완전히 채택하지 않았기 때문이다.

결과적으로 transitive dependency를 직접 추가해야 하는 경우가 있다:
- `github.com/docker/go-units`
- `github.com/moby/docker-image-spec`
- `github.com/opencontainers/image-spec`

이는 `+incompatible` 모듈의 알려진 한계이며, Docker SDK를 사용하는 Go 프로젝트에서 일반적이다.
