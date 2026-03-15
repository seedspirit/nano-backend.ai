# Manager Crate — Agent Guidelines

HTTP API 서버이자 서비스 진입점 바이너리 크레이트.

## Role

- 외부 HTTP 요청을 수신하고 핸들러로 라우팅
- `common::ApiResponse`를 통해 구조화된 JSON 응답 반환
- 서버 부트스트랩 및 라이프사이클 관리

## Rules

- `main.rs`는 부트스트랩만 — 비즈니스 로직은 모듈로 분리
- 새 엔드포인트 추가 시: `app.rs`에 라우트 등록 + 전용 핸들러 모듈 생성
- 모든 핸들러는 성공/에러 양쪽 테스트 케이스 필수
- 로깅은 `tracing` 매크로만 사용 (`debug!`, `info!`, `warn!`, `error!`)
- 핸들러 에러는 `error.rs`가 아닌 각 핸들러 모듈 내부에서 정의
- `error.rs`는 서버 기동/운영 에러 전용 (`Bind`, `Serve`)
