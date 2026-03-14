## 개요
Backend.AI의 nano 버전을 지금부터 같이 만들어나가고자 합니다. 아래 내용을 통해 프로젝트를 init하고, 프로젝트의 root 폴더에 들어갈 README.md, CLAUDE.md, AGENTS.md를 만들고자 합니다.

0-1. 제작 목적
- Backend.AI의 핵심 구조를 작은 크기로 다시 설계하며 학습한다
- 실제 제품 설계에서 필요한 design decision 훈련을 한다
- AI agent와의 협업을 전제로 한 코드베이스 운영 방식을 실험한다
- 설치, 업그레이드, 마이그레이션이 쉬운 온프레미스형 소프트웨어를 지향한다.

0-2. 사용자 계층
- AI Agent가 일등 사용자임을 가정하여 만듭니다. CLI, SDK를 가장 먼저 제공하고, json 등의 응답을 해줄 때 AI Agent가 받아 적절하게 tool calling을 하는 방식으로 응답을 설계하고 싶습니다
    - 응답은 구조적이어야 한다
    - machine-readable field를 우선한다
    - 비정형 문자열보다 명시적 status / reason / next_action_hint를 선호한다.
    - long-running 작업은 polling 가능한 job/session model로 제공한다.
- 사람 유저 향을 대상으로 한다고 했을 때 /v1/chat/completions 등을 통해 대화를 통해 애플리케이션을 제어하도록 하면 좋을 것 같습니다

1. 우선 최소한의 기능이 동작하는 것을 목표로 합니다. 기능 고도화 등은 이후에 진행합니다

Nano Backend.AI 컴포넌트
- Manager: control plane
    - 외부 API 제공
    - Job/Agent 메타데이터 관리
    - 스케줄링 트리거
    - 상태 전이 검증
    - 결과 조회 API 제공
    - 설치/운영 관점의 system-of-record 역할

- Agent: execution plane
    - Manager에 등록
    - heartbeat 송신
    - job 수신 및 실행
    - 실행 결과 및 상태 보고
    - 로컬 executor 추상화 소유

- Database(Postgres): durable state, source of truth
    - durable metadata 저장
    - 재시작 후 상태 복원 근거 제공
    - Lock 제공

- Redis: ephemeral coordination
    - ephemeral coordination
    - event bus
    - queue-like signal transport
    - cache / lease / lock 보조

- Reverse Proxy(Web Server & App Proxy 겸용): ingress boundary
    - 외부 진입점
    - TLS 종료
    - static/API reverse proxy
    - future app proxy 확장 가능성 보유

1-2. 비목표
초기 버전에서는 의도적으로 제외하는 기능은 아래와 같다
- 멀티테넌시
- 인증/인가
- GPU 스케줄링
- 이미지 빌드 파이프라인
- 고급 reverse proxy 라우팅
- 분산 DB

2. Backend.AI nano 또한 온프레미스 혹은 vm 위에 설치되어 동작하게 됩니다. 이때 가장 중요한 것은 설치 용이성, 마이그레이션 용이성입니다. 처음부터 설치와 버전 업그레이드, 다운그레이드 등을 신경 써서 구조를 설계하여 진행하면 좋을 것 같습니다

3. Tech Spec
- Language: rust
- Manager/ Agent runtime: tokio
- API: 
    - 외부: HTTP + JSON REST
    - 내부: gRPC
- Redis Client: Valkey Glide
- DB: Postgres

3-1. Agentic Coding
- 각 디렉토리에 CLAUDE.md와 이것이 심볼릭 링크로 연결된 AGENTS.md를 둡니다.
- 세부 설계의 경우 docs/design 하위에 넣습니다
- 각 md 파일에는 디렉토리에 대한 설명, 역할, 제약 등을 적어 agent의 역할과 기능 수행을 제어합니다
    - 각 md에 들어갈 내용
        - 디렉토리 목적
        - 포함 가능한 코드의 범위
        - 금지 사항
        - 의존성 규칙
        - 테스트 원칙
        - 변경 시 함께 수정해야 할 파일
- 이때 md 파일은 최대한 간결하게 유지하게 만들어 한 번에 너무 많은 md 파일 정보가 context에 쌓이지 않도록 만듭니다. policy와 role만 쓰고 구현 상세를 장황하게 반복하지 않는다
- 루트 문서는 전체 원칙, 하위 문서는 지역 규칙만 다룬다.