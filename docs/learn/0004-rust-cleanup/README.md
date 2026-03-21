# Rust 아티팩트 제거 — Go 마이그레이션 정리

PR: #pending
Date: 2026-03-21

## What was done

- Rust 소스코드(`crates/`), 빌드 설정(`Cargo.toml`, `Cargo.lock`), 빌드 출력(`target/`) 삭제
- `.gitignore`, `.coderabbit.yaml`, `README.md`, `CLAUDE.md`에서 Rust 참조를 Go로 교체

## Categories

- [Code Design](./code-design.md)

## Key decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| Rust 코드를 완전 삭제 | Go 마이그레이션 완료 후 혼란과 594MB 불필요한 용량 제거 | 별도 브랜치에 보관 — git history로 충분히 추적 가능하므로 불필요 |
| `.coderabbit.yaml`을 Go 경로로 재작성 | 코드리뷰 봇이 Go 규칙을 적용하도록 | 파일 삭제 후 재생성 — 기존 설정(language, chat 등) 유지 위해 수정 선택 |

## Further study

- [ ] `.coderabbit.yaml` path_instructions에 `cmd/**/*.go` 규칙 추가 고려
- [ ] golangci-lint 로컬 설치 및 `.golangci.yml` 설정 파일 작성
