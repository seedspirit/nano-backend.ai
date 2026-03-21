#!/usr/bin/env bash
# =============================================================================
# Batch issue creation script for Epic 2: Rust → Go 마이그레이션
#
# Prerequisites:
#   - gh CLI authenticated (gh auth login)
#   - Labels exist: epic, story, migration, go, infra, common, manager, agent, documentation
#
# Usage:
#   cd docs/issues && bash create-issues.sh
# =============================================================================

set -euo pipefail

REPO="seedspirit/nano-backend.ai"

echo "=== Step 0: Create labels (if missing) ==="
gh label create "migration" --color "D4C5F9" --description "Language/framework migration" --repo "$REPO" 2>/dev/null || echo "  label 'migration' already exists"
gh label create "go" --color "00ADD8" --description "Go language component" --repo "$REPO" 2>/dev/null || echo "  label 'go' already exists"

echo ""
echo "=== Step 1: Create Milestone ==="
MILESTONE_URL=$(gh api repos/$REPO/milestones --method POST \
  -f title="Epic 2: Rust → Go 마이그레이션" \
  -f description="Rust 코드베이스를 Go로 완전히 포팅" \
  --jq '.number' 2>/dev/null || echo "")

if [ -z "$MILESTONE_URL" ]; then
  echo "  Milestone may already exist. Fetching..."
  MILESTONE_NUM=$(gh api repos/$REPO/milestones --jq '.[] | select(.title | contains("Go")) | .number' | head -1)
else
  MILESTONE_NUM="$MILESTONE_URL"
fi
echo "  Milestone number: $MILESTONE_NUM"

echo ""
echo "=== Step 2: Create Epic Issue ==="
EPIC_NUM=$(gh issue create --repo "$REPO" \
  --title "Epic: Rust에서 Go로 언어 마이그레이션" \
  --label "epic,migration,go" \
  --milestone "Epic 2: Rust → Go 마이그레이션" \
  --body "$(cat <<'EPICEOF'
## Epic

### Goal
현재 Rust로 구현된 nano-backend.ai 코드베이스(common, manager, agent)를 Go로 완전히 포팅하여, 동일한 아키텍처와 API 계약을 Go 생태계 위에서 재현한다.

### Motivation
- Go는 빌드 속도, 배포 단순성, 동시성 모델에서 on-premise 환경에 적합
- 러닝 커브가 낮아 AI 에이전트(Claude)가 더 빠르고 정확하게 코드를 생성/수정 가능
- 기존 Backend.AI 생태계(Python)와의 연동 시 Go의 gRPC/HTTP 생태계가 성숙
- Rust의 생산성(컴파일 시간, 복잡한 라이프타임 관리)이 학습 프로젝트 진행 속도를 저해

### Context
현재 코드베이스 상태:
- **Cargo workspace**: `common`, `manager`, `agent` 3개 크레이트
- **common** (`crates/common/`): `ApiResponse`, `CommonError` 공유 타입 (82 LOC)
- **manager** (`crates/manager/`): axum 기반 HTTP 서버, `GET /health` 엔드포인트 (117 LOC)
- **agent** (`crates/agent/`): 스캐폴드만 존재, tracing 초기화 후 종료 (25 LOC)
- **CI** (`.github/workflows/ci.yml`): fmt → clippy → test 파이프라인
- **Skills** (`.claude/skills/`): Rust 기반 개발 가이드 및 자동화 스킬

### Stories

| # | Story | Summary | Component | Depends on |
|---|-------|---------|-----------|------------|
| S1 | Go 프로젝트 스캐폴드 및 CI 설정 | Go module 초기화, 디렉토리 구조, CI 파이프라인 구축 | infra | — |
| S2 | common 패키지 포팅 | ApiResponse, CommonError를 Go struct/error로 재구현 | common | S1 |
| S3 | manager 포팅 | Go net/http 기반 서버 + GET /health + 라우터 테스트 | manager | S2 |
| S4 | agent 스캐폴드 포팅 | Agent 바이너리 진입점 + 로깅 초기화 + 에러 타입 | agent | S1 |
| S5 | CLAUDE.md 및 문서 Go 전환 | 루트/서브 CLAUDE.md, README를 Go 컨벤션으로 갱신 | docs | S1 |
| S6 | Skills Go 전환 | rust-guide → go-guide, 기존 스킬 내 Rust 참조 변경 | infra | S5 |
| S7 | 설계 문서 Go 전환 | KernelRuntime 설계를 Go interface로 갱신 | docs | S2 |

### Dependency Graph
```
S1 ─┬→ S2 ─┬→ S3
    │      └→ S7
    ├→ S4
    └→ S5 → S6
```

### Design Decisions

| Decision | Why | Alternatives considered |
|----------|-----|------------------------|
| 표준 라이브러리 `net/http` + 경량 라우터 | Go 표준 라이브러리가 충분히 강력, 외부 의존성 최소화 | Gin, Echo, Chi |
| `log/slog` 구조적 로깅 | tracing 대응, Go 1.21+ 표준 | zerolog, zap |
| Go interface로 KernelRuntime 추상화 | Rust trait과 1:1 대응, 암묵적 인터페이스가 자연스러움 | 코드 생성 기반 |
| 단일 모듈 + internal/ 패키지 | crate 가시성 제어와 유사 | Go workspace |

### Out of Scope
- gRPC 통신 구현, PostgreSQL/Redis 연동, Docker/K8s 런타임, 기존 Rust 코드 삭제

### Success Criteria
- `go build ./...` 성공
- `go test ./...` 전체 통과 (기존 8개 테스트 시나리오 대응)
- `GET /health` 응답이 Rust 버전과 동일한 JSON 구조
- CI 파이프라인 전체 통과
- CLAUDE.md와 Skills가 Go 기준으로 갱신
EPICEOF
)" 2>&1 | grep -oP '#\d+' | head -1 | tr -d '#')

echo "  Epic issue created: #$EPIC_NUM"

echo ""
echo "=== Step 3: Create Story Issues ==="

# S1
S1_NUM=$(gh issue create --repo "$REPO" \
  --title "S1: Go 프로젝트 스캐폴드 및 CI 설정" \
  --label "story,infra,go" \
  --milestone "Epic 2: Rust → Go 마이그레이션" \
  --body "$(cat <<EOF
## Story

**Epic**: #$EPIC_NUM
**Component**: infra

### Background
Rust Cargo workspace를 Go 프로젝트 구조로 전환하는 첫 번째 단계. 이후 모든 Story가 이 스캐폴드 위에 코드를 작성하므로 선행 필수.

### Goal
Go module을 초기화하고 manager/agent/common 패키지 디렉토리를 생성하며, CI 파이프라인을 Go 도구 체인으로 전환한다.

### Acceptance Criteria
- [ ] \`go build ./...\` 성공 (빈 main 패키지 포함)
- [ ] \`.github/workflows/ci.yml\`이 \`go fmt\`, \`go vet\`, \`staticcheck\`, \`go test ./...\` 실행
- [ ] 프로젝트 디렉토리: \`cmd/manager/\`, \`cmd/agent/\`, \`internal/common/\`, \`internal/manager/\`, \`internal/agent/\`

### Affected Code
- \`go.mod\`, \`cmd/manager/main.go\`, \`cmd/agent/main.go\`, \`internal/\` 패키지, \`.github/workflows/ci.yml\`

### Test Plan
- CI 파이프라인이 \`go build ./...\` 성공
- \`go vet ./...\` 경고 없음
- \`go fmt\` 검사 통과
EOF
)" 2>&1 | grep -oP '#\d+' | head -1 | tr -d '#')
echo "  S1 created: #$S1_NUM"

# S2
S2_NUM=$(gh issue create --repo "$REPO" \
  --title "S2: common 패키지 포팅 — ApiResponse 및 에러 타입" \
  --label "story,common,go" \
  --milestone "Epic 2: Rust → Go 마이그레이션" \
  --body "$(cat <<EOF
## Story

**Epic**: #$EPIC_NUM
**Component**: common
**blockedBy**: #$S1_NUM (S1)

### Background
Rust \`common\` 크레이트의 \`ApiResponse\`와 \`CommonError\`는 모든 API 응답의 계약이다. Manager와 Agent 포팅 전에 이 공유 타입이 Go로 먼저 전환되어야 한다.

### Goal
\`ApiResponse\` struct와 에러 타입을 Go로 재구현하고, JSON 직렬화/역직렬화 테스트를 포함한다.

### Acceptance Criteria
- [ ] \`internal/common/response.go\`에 \`ApiResponse\` struct + \`NewApiResponse\`, \`Ok\`, \`Error\` 생성 함수 + JSON 태그
- [ ] \`internal/common/error.go\`에 \`CommonError\` 타입 정의
- [ ] 기존 Rust 테스트 5개 시나리오를 Go 테스트로 대응

### Affected Code
- \`internal/common/response.go\`, \`internal/common/response_test.go\`, \`internal/common/error.go\`, \`internal/common/error_test.go\`

### Test Plan
- Unit test: \`NewApiResponse\` 모든 필드 설정
- Unit test: \`Ok\` → status="ok"
- Unit test: \`Error\` → status="error"
- Unit test: JSON 직렬화 검증
- Unit test: JSON 역직렬화 검증
EOF
)" 2>&1 | grep -oP '#\d+' | head -1 | tr -d '#')
echo "  S2 created: #$S2_NUM"

# S3
S3_NUM=$(gh issue create --repo "$REPO" \
  --title "S3: manager 포팅 — HTTP 서버 및 health 엔드포인트" \
  --label "story,manager,go" \
  --milestone "Epic 2: Rust → Go 마이그레이션" \
  --body "$(cat <<EOF
## Story

**Epic**: #$EPIC_NUM
**Component**: manager
**blockedBy**: #$S2_NUM (S2)

### Background
Rust manager는 axum 기반 HTTP 서버로 \`GET /health\` 엔드포인트를 제공한다. Go의 \`net/http\` 표준 라이브러리로 동일한 기능을 재구현한다.

### Goal
Go \`net/http\` 기반 HTTP 서버를 구축하고, \`GET /health\` 엔드포인트가 \`ApiResponse\` JSON을 반환하도록 한다.

### Acceptance Criteria
- [ ] \`cmd/manager/main.go\`에서 HTTP 서버가 \`127.0.0.1:8080\`에서 시작, \`slog\` 로거 초기화
- [ ] \`GET /health\` → 200 OK + 올바른 ApiResponse JSON
- [ ] 알 수 없는 경로 → 404 응답

### Affected Code
- \`cmd/manager/main.go\`, \`internal/manager/app.go\`, \`internal/manager/health.go\`, \`internal/manager/error.go\`, 테스트 파일

### Test Plan
- Integration test: \`GET /health\` → 200 + JSON 검증
- Integration test: \`GET /nonexistent\` → 404
- Unit test: health 핸들러 → status="healthy"
EOF
)" 2>&1 | grep -oP '#\d+' | head -1 | tr -d '#')
echo "  S3 created: #$S3_NUM"

# S4
S4_NUM=$(gh issue create --repo "$REPO" \
  --title "S4: agent 스캐폴드 포팅" \
  --label "story,agent,go" \
  --milestone "Epic 2: Rust → Go 마이그레이션" \
  --body "$(cat <<EOF
## Story

**Epic**: #$EPIC_NUM
**Component**: agent
**blockedBy**: #$S1_NUM (S1)

### Background
Rust agent는 현재 스캐폴드 상태로, tracing 초기화 후 로그를 출력하고 종료한다. 동일한 수준의 Go 스캐폴드를 구축한다.

### Goal
Agent 바이너리 진입점을 Go로 구현하고, 구조적 로깅을 초기화하며, 에러 타입을 정의한다.

### Acceptance Criteria
- [ ] \`cmd/agent/main.go\`에서 \`slog\` 로거 초기화 + "agent started" 로그 출력 후 정상 종료
- [ ] \`internal/agent/error.go\`에 \`AgentError\` 타입 정의
- [ ] \`go build ./cmd/agent\` 성공, \`go vet\` 경고 없음

### Affected Code
- \`cmd/agent/main.go\`, \`internal/agent/error.go\`

### Test Plan
- \`go build ./cmd/agent\` 성공 확인
- \`go vet ./internal/agent/...\` 경고 없음
EOF
)" 2>&1 | grep -oP '#\d+' | head -1 | tr -d '#')
echo "  S4 created: #$S4_NUM"

# S5
S5_NUM=$(gh issue create --repo "$REPO" \
  --title "S5: CLAUDE.md 및 문서 Go 전환" \
  --label "story,documentation,go" \
  --milestone "Epic 2: Rust → Go 마이그레이션" \
  --body "$(cat <<EOF
## Story

**Epic**: #$EPIC_NUM
**Component**: docs
**blockedBy**: #$S1_NUM (S1)

### Background
현재 CLAUDE.md와 README.md는 Rust 중심으로 작성되어 있다. AI 에이전트가 Go 코드를 올바르게 생성하려면 가이드라인이 Go 컨벤션을 반영해야 한다.

### Goal
CLAUDE.md(루트, common, manager)와 README.md를 Go 언어 및 도구 체인 기준으로 갱신한다.

### Acceptance Criteria
- [ ] 루트 CLAUDE.md: 언어를 Go로 변경, 도구/금지 사항을 Go 관용 표현으로 갱신
- [ ] README.md: Tech Stack을 Go 기준으로 갱신
- [ ] 서브 CLAUDE.md: 새 디렉토리 경로(\`internal/\`)와 Go 규칙 반영

### Affected Code
- \`CLAUDE.md\`, \`README.md\`, \`internal/common/CLAUDE.md\`, \`internal/manager/CLAUDE.md\`

### Test Plan
- "Rust", "cargo", "crate" 키워드가 Go 대응 용어로 교체 확인
- Tech Stack 섹션 갱신 확인
EOF
)" 2>&1 | grep -oP '#\d+' | head -1 | tr -d '#')
echo "  S5 created: #$S5_NUM"

# S6
S6_NUM=$(gh issue create --repo "$REPO" \
  --title "S6: Skills Go 전환 — go-guide 신규 + 기존 스킬 수정" \
  --label "story,infra,go" \
  --milestone "Epic 2: Rust → Go 마이그레이션" \
  --body "$(cat <<EOF
## Story

**Epic**: #$EPIC_NUM
**Component**: infra
**blockedBy**: #$S5_NUM (S5)

### Background
\`.claude/skills/\` 아래 스킬들이 Rust 도구와 패턴을 참조한다. Go 포팅 후 에이전트가 올바른 도구와 패턴을 사용하려면 스킬도 갱신되어야 한다.

### Goal
\`rust-guide\`를 \`go-guide\`로 대체하고, 기존 스킬 내 Rust 참조를 Go로 변경한다.

### Acceptance Criteria
- [ ] \`.claude/skills/go-guide/SKILL.md\` 신규 생성 — Go 코딩 컨벤션, 에러 처리, 테스트 패턴 포함
- [ ] \`tdd-guide\`, \`submit\`, \`autopilot\` 내 Rust 도구/코드 참조를 Go로 변경
- [ ] \`analyze\` 스킬의 크레이트 참조를 Go 패키지 경로로 변경

### Affected Code
- \`.claude/skills/go-guide/SKILL.md\` (신규), \`.claude/skills/rust-guide/\` (삭제/deprecated), \`tdd-guide\`, \`submit\`, \`autopilot\`, \`analyze\` SKILL.md (수정)

### Test Plan
- go-guide 파일 존재 및 주요 섹션 포함 확인
- tdd-guide에서 \`cargo\` 명령어 미참조 확인
EOF
)" 2>&1 | grep -oP '#\d+' | head -1 | tr -d '#')
echo "  S6 created: #$S6_NUM"

# S7
S7_NUM=$(gh issue create --repo "$REPO" \
  --title "S7: 설계 문서 Go 전환" \
  --label "story,documentation,go" \
  --milestone "Epic 2: Rust → Go 마이그레이션" \
  --body "$(cat <<EOF
## Story

**Epic**: #$EPIC_NUM
**Component**: docs
**blockedBy**: #$S2_NUM (S2)

### Background
\`docs/design/0001-session-kernel-pipeline.md\`는 KernelRuntime trait 설계를 Rust 코드로 설명한다. Go 포팅 후 설계 문서의 코드가 실제 구현과 괴리되면 혼란을 야기한다.

### Goal
설계 문서 내 Rust 코드 예시를 Go interface/struct 예시로 교체하고, Rust 특화 용어를 Go 대응 용어로 갱신한다.

### Acceptance Criteria
- [ ] 모든 Rust 코드 블록이 Go 코드 블록으로 교체
- [ ] \`KernelRuntime\` trait → interface, \`LocalProcessRuntime\` struct 예시 포함
- [ ] Option A/B/C 비교를 Go 관점으로 재평가

### Affected Code
- \`docs/design/0001-session-kernel-pipeline.md\`

### Test Plan
- 코드 블록이 Go 문법으로 유효한지 확인
- "trait", "crate", "async fn" 등 Rust 용어 미잔존 확인
EOF
)" 2>&1 | grep -oP '#\d+' | head -1 | tr -d '#')
echo "  S7 created: #$S7_NUM"

echo ""
echo "============================================="
echo "  All issues created successfully!"
echo "============================================="
echo ""
echo "Epic: #$EPIC_NUM"
echo "Stories:"
echo "  S1: #$S1_NUM (Go scaffold + CI)"
echo "  S2: #$S2_NUM (common package)"
echo "  S3: #$S3_NUM (manager porting)"
echo "  S4: #$S4_NUM (agent scaffold)"
echo "  S5: #$S5_NUM (docs conversion)"
echo "  S6: #$S6_NUM (skills conversion)"
echo "  S7: #$S7_NUM (design doc conversion)"
echo ""
echo "Dependency graph:"
echo "  S1 ─┬→ S2(#$S2_NUM) ─┬→ S3(#$S3_NUM)"
echo "       │                └→ S7(#$S7_NUM)"
echo "       ├→ S4(#$S4_NUM)"
echo "       └→ S5(#$S5_NUM) → S6(#$S6_NUM)"
