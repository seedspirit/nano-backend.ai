# Learning Notes

PR 단위로 자동 생성되는 학습 기록.

`/submit` 스킬이 PR 제출 시 이 디렉토리에 학습 문서를 추가합니다.

## 디렉토리 규칙

- 구조: `NNNN-<slug>/` (예: `0002-job-scheduler/`)
- 각 디렉토리 안에 카테고리별 MD 파일 생성:
  - `README.md` — PR 요약, 카테고리 목차, 핵심 결정, Further study
  - `code-design.md` — 코드 설계 (디자인 패턴, SOLID, 모듈 구조 등)
  - `cs.md` — CS 개념 (자료구조, 알고리즘, 네트워킹, 동시성 등)
  - `rust.md` — Rust 프로그래밍 (문법, 소유권, 트레이트, async 등)
  - `backend-ai.md` — Backend.AI 아키텍처 (도메인 모델, 세션 라이프사이클 등)
- 해당 PR에서 배운 것이 없는 카테고리는 파일을 생성하지 않음
- 한국어로 작성
- 구현 내용보다 **배운 것**에 집중
- "Further study" 체크리스트는 구체적으로
