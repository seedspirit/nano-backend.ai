# Code Design

## 언어 마이그레이션 시 정리 범위

프로젝트의 언어를 전환할 때 소스코드 삭제만으로는 부족하다. 정리해야 할 범위:

1. **소스코드 및 빌드 설정**: `crates/`, `Cargo.toml`, `Cargo.lock`
2. **빌드 출력물**: `target/` (git에 추적되지 않더라도 로컬에 존재)
3. **CI/CD 설정**: 이전 언어의 빌드/테스트 스텝 (이미 PR #37에서 완료)
4. **코드리뷰 도구 설정**: `.coderabbit.yaml`의 경로 패턴과 언어별 규칙
5. **프로젝트 문서**: `README.md`의 Tech Stack, Project Layout 섹션
6. **개발 가이드**: `CLAUDE.md`의 포매터, 린터, 테스트 컨벤션
7. **VCS 설정**: `.gitignore`의 언어별 무시 패턴

## 설정 파일의 경로 패턴 일관성

코드리뷰 봇(`coderabbit`), 린터, CI 등 여러 도구가 경로 패턴으로 규칙을 적용한다.
언어 전환 시 이 패턴들이 실제 프로젝트 구조와 일치하는지 반드시 확인해야 한다.

- Rust: `crates/**/*.rs` → Go: `internal/**/*.go`
- Rust: `!target/**` (빌드 출력 제외) → Go에서는 불필요 (`go build`는 캐시를 `GOPATH`에 저장)

## Git History를 활용한 코드 보존

삭제된 파일은 git history에 영구 보존되므로, 참조용 브랜치를 따로 만들 필요 없다.
`git log --all --full-history -- crates/` 명령으로 언제든 과거 코드를 확인할 수 있다.
