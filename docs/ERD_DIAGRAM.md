```mermaid
erDiagram
    USER ||--o{ CUSTOM_MATCH : hosts
    USER ||--o{ PARTICIPANT : "participates as"
    CUSTOM_MATCH ||--o{ PARTICIPANT : contains
    CUSTOM_MATCH ||--|| APPLICATION_PERIOD : has
    CUSTOM_MATCH ||--o| MATCH_SCHEDULE : has

    USER {
        string id PK
        string username
        string avatar
    }

    CUSTOM_MATCH {
        string id PK
        string title
        string description
        string gameMode
        number players
        number maxPlayers
        string hostId FK
        enum status "waiting | in-progress | completed"
        string prize
        array rules
        datetime createdAt
        datetime updatedAt
    }

    PARTICIPANT {
        string id PK
        string userId FK
        string matchId FK
        string username
        string avatar
        datetime joinedAt
        enum status "pending | approved | rejected"
    }

    APPLICATION_PERIOD {
        string matchId FK
        datetime start
        datetime end
    }

    MATCH_SCHEDULE {
        string matchId FK
        datetime start
        datetime end
    }

    TOURNAMENT {
        string id PK
        string title
        enum status "upcoming | ongoing | completed"
        datetime startDate
        number participants
        string prize
    }

    MATCH_HISTORY {
        string id PK
        string title
        string result
        string score
        datetime date
    }

엔티티 상세 설명

관계 설명

1. USER ↔ CUSTOM_MATCH (1:N)
  - 한 명의 사용자는 여러 개의 커스텀 매치를 주최할 수 있습니다.
2. USER ↔ PARTICIPANT (1:N)
  - 한 명의 사용자는 여러 매치에 참가자로 참여할 수 있습니다.
3. CUSTOM_MATCH ↔ PARTICIPANT (1:N)
  - 하나의 커스텀 매치는 여러 참가자를 가질 수 있습니다.
  - 참가자는 approved 또는 pending 상태로 관리됩니다.
4. CUSTOM_MATCH ↔ APPLICATION_PERIOD (1:1)
  - 각 커스텀 매치는 하나의 신청 기간을 가집니다.
5. CUSTOM_MATCH ↔ MATCH_SCHEDULE (1:0..1)
  - 각 커스텀 매치는 하나의 매치 일정을 가질 수 있습니다 (optional).

데이터 흐름

1. 매치 생성: USER가 CUSTOM_MATCH를 생성 (host)
2. 참가 신청: 다른 USER가 PARTICIPANT로 참가 신청 (status: pending)
3. 승인/거부: 호스트가 PARTICIPANT의 status를 approved/rejected로 변경
4. 매치 진행: CUSTOM_MATCH의 status가 waiting → in-progress → completed로 변경

이 ERD는 현재 코드베이스에 있는 타입 정의를 기반으로 작성되었습니다. 주요 엔티티는 다음과 같습니다:

- **User**: 사용자 정보
- **CustomMatch**: 커스텀 매치 (src/shared/types/customMatch.ts:17)
- **Participant**: 참가자 정보 (src/shared/types/customMatch.ts:8)
- **Tournament**: 토너먼트 정보 (src/entities/tournament/ui/TournamentCard.tsx:3)
- **MatchHistory**: 매치 기록 (src/entities/match/ui/MatchCard.tsx:3)
```

## USER

사용자 기본 정보를 저장하는 엔티티

- id: 사용자 고유 식별자 (PK)
- username: 사용자명
- avatar: 프로필 이미지 URL (optional)

## CUSTOM_MATCH

사용자가 생성한 커스텀 게임 매치 정보

- id: 매치 고유 식별자 (PK)
- title: 매치 제목
- description: 매치 설명
- gameMode: 게임 모드 (예: "5v5 ランクマッチ")
- players: 현재 참가자 수
- maxPlayers: 최대 참가자 수
- hostId: 주최자 사용자 ID (FK → USER)
- status: 매치 상태
    - waiting: 대기 중 (참가자 모집)
    - in-progress: 진행 중
    - completed: 완료
- prize: 상품 정보 (optional)
- rules: 매치 규칙 배열 (optional)
- createdAt: 생성 일시
- updatedAt: 수정 일시

## PARTICIPANT (참가자)

매치에 참가하는 사용자 정보

- id: 참가자 고유 식별자 (PK)
- userId: 사용자 ID (FK → USER)
- matchId: 매치 ID (FK → CUSTOM_MATCH)
- username: 사용자명 (비정규화)
- avatar: 프로필 이미지 (비정규화, optional)
- joinedAt: 참가 신청 일시
- status: 참가 상태
- pending: 승인 대기
- approved: 승인됨
- rejected: 거부됨

## APPLICATION_PERIOD

매치 참가 신청 기간 정보

- matchId: 매치 ID (FK → CUSTOM_MATCH)
- start: 신청 시작 일시
- end: 신청 종료 일시

## MATCH_SCHEDULE

실제 매치 진행 일정

- matchId: 매치 ID (FK → CUSTOM_MATCH)
- start: 매치 시작 일시
- end: 매치 종료 일시 (optional)

## TOURNAMENT

공식 토너먼트 정보

- id: 토너먼트 고유 식별자 (PK)
- title: 토너먼트 제목
- status: 토너먼트 상태
    - upcoming: 예정
    - ongoing: 진행 중
    - completed: 완료
- startDate: 시작 일시
- participants: 참가 팀 수
- prize: 상금 정보

## MATCH_HISTORY

사용자의 과거 매치 전적

- id: 기록 고유 식별자 (PK)
- title: 매치 제목
- result: 결과 ("勝利", "敗北", "引き分け")
- score: 점수 (예: "13-7")
- date: 매치 날짜