# Common Crate — Agent Guidelines

Workspace 전체에서 공유하는 타입과 계약(contract)을 정의하는 라이브러리 크레이트.

## Role

- 모든 크레이트가 의존하는 공유 타입 제공 (`ApiResponse`, `CommonError`)
- API 응답 형식의 단일 진실 공급원 (single source of truth)
- 런타임 로직 없음 — 순수 데이터 구조와 trait만 포함

## Rules

- `ApiResponse` 변경은 모든 소비자에게 영향 — 변경 전 하위 크레이트 테스트 확인 필수
- 다른 workspace 크레이트에 의존 금지 (의존 그래프의 최하위 계층)
- 새 공개 타입 추가 시 `lib.rs`에 `pub use` re-export 필수
- 에러 타입은 인프라 수준만 — 도메인 에러는 각 크레이트에서 정의
