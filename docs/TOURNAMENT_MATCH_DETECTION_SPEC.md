# Tournament Match Detection & Auto Result Recording

## 1. 개요

Tournament 브라켓이 생성되고 팀 배정이 완료된 이후, Staff가 각 Game(시합)의 시작 시간을 설정하면 해당 시간 이후 일정 시간(기본 2시간, 설정 가능) 이내에 Valorant API(govapi)를 통해 해당 Game에 참여하는 모든 멤버가 포함된 매치를 자동 감지하고, 감지된 매치의 결과를 Tournament Game에 자동 기입하는 기능.

---

## 2. 용어 정의

| 용어 | 설명 |
|------|------|
| **Contest** | 대회 전체를 의미하는 최상위 엔티티 (TOURNAMENT 타입) |
| **Game** | Tournament 브라켓 내 하나의 시합 (Round + MatchNumber로 식별) |
| **GameTeam** | Game에 참여하는 팀 (Grade로 순위 기록) |
| **Team / TeamMember** | 팀과 소속 멤버 |
| **Match** | Valorant 실제 인게임 매치 (VAPI에서 조회되는 단위) |
| **Staff** | Contest의 `MemberType=STAFF` 역할을 가진 멤버 |
| **Detection Window** | Game 시작 시간 이후 매치를 탐색하는 시간 범위 (기본 2시간) |

---

## 3. 전체 흐름

```
[Tournament 생성 & 팀 배정 완료]
            │
            ▼
[Staff가 Game 시작 시간 설정]  ← 이 기획의 시작점
            │
            ▼
[설정된 시작 시간 도달]
            │
            ▼
[Game 상태: PENDING → ACTIVE]
            │
            ▼
[Match Detection Scheduler 시작]
   - 주기적으로 VAPI 폴링
   - Detection Window 내 매치 탐색
            │
            ▼
[매치 감지 성공]
   - 양 팀 멤버 전원 포함 여부 검증
   - 커스텀 게임 여부 확인
            │
            ▼
[매치 결과 자동 기입]
   - GameTeam Grade 설정 (승리팀=1, 패배팀=2)
   - Game 상태: ACTIVE → FINISHED
   - 다음 라운드 Game에 승리팀 자동 배정
            │
            ▼
[결과 이벤트 발행]
   - RabbitMQ로 match.detected, game.finished 이벤트 발행
   - Discord 알림 (선택)
```

---

## 4. 기능 상세

### 4.1 Game 시작 시간 설정

#### 4.1.1 요구사항

- Staff 역할의 멤버만 Game의 시작 시간을 설정할 수 있다.
- Tournament 브라켓의 Game은 현재 `StartedAt`/`EndedAt`이 optional이므로, Staff가 이를 명시적으로 설정한다.
- 시작 시간은 현재 시간 이후로만 설정 가능하다.
- 같은 라운드 내 Game들의 시작 시간은 독립적으로 설정 가능하다.
- Detection Window(탐색 시간 범위)도 Game 단위로 설정 가능하다 (기본값: 2시간).

#### 4.1.2 API 설계

```
PUT /api/v1/contests/{contestId}/games/{gameId}/schedule
```

**Request Body:**
```json
{
  "scheduledStartTime": "2025-02-15T14:00:00+09:00",
  "detectionWindowMinutes": 120
}
```

**Response:**
```json
{
  "gameId": 1,
  "contestId": 1,
  "round": 1,
  "matchNumber": 1,
  "scheduledStartTime": "2025-02-15T14:00:00+09:00",
  "detectionWindowMinutes": 120,
  "gameStatus": "PENDING"
}
```

#### 4.1.3 Game 도메인 변경

기존 `Game` 엔티티에 다음 필드를 추가한다:

| 필드 | 타입 | 설명 |
|------|------|------|
| `ScheduledStartTime` | `*time.Time` | Staff가 설정한 시합 예정 시작 시간 |
| `DetectionWindowMinutes` | `int` | 매치 탐색 시간 범위 (분 단위, 기본 120) |
| `DetectedMatchID` | `*string` | VAPI에서 감지된 Valorant 매치 ID |
| `DetectionStatus` | `string` | 감지 상태: `NONE`, `DETECTING`, `DETECTED`, `FAILED`, `MANUAL` |

**DetectionStatus 상태 전이:**
```
NONE → DETECTING → DETECTED
                 → FAILED → MANUAL (Staff 수동 입력)
```

---

### 4.2 Game 자동 시작 (Scheduled Activation)

#### 4.2.1 요구사항

- `ScheduledStartTime`이 설정된 Game은 해당 시간이 되면 자동으로 `PENDING → ACTIVE` 상태 전이.
- 자동 시작과 동시에 Match Detection이 시작된다.
- Scheduler가 주기적으로(1분 간격) 시작 시간이 도래한 Game을 확인한다.

#### 4.2.2 구현 방식

```
[Cron Scheduler - 1분 간격]
        │
        ▼
[ScheduledStartTime <= now && Status == PENDING 인 Game 조회]
        │
        ▼
[Game 상태 → ACTIVE, DetectionStatus → DETECTING]
        │
        ▼
[Match Detection Job 등록]
```

---

### 4.3 Match Detection (매치 감지)

#### 4.3.1 핵심 로직

1. **대상 Game 조회**: `GameStatus=ACTIVE` && `DetectionStatus=DETECTING`인 Game 목록 조회.
2. **팀 멤버 정보 조회**: Game에 참여하는 양 팀의 전체 멤버 Valorant 계정 정보(name, tag) 조회.
3. **VAPI 매치 히스토리 조회**: 팀 멤버 중 한 명을 기준으로 최근 매치 히스토리를 조회.
4. **매치 필터링**: `ScheduledStartTime` 이후 ~ `ScheduledStartTime + DetectionWindowMinutes` 이내의 매치만 대상.
5. **멤버 전원 포함 검증**: 매치 참가자 목록에 양 팀 멤버 전원이 포함되어 있는지 확인.
6. **결과 기입**: 조건을 만족하는 매치가 발견되면 결과를 자동 기입.

#### 4.3.2 VAPI 호출 전략

```
[Game의 Team A, Team B 멤버 목록 확보]
        │
        ▼
[Team A 리더의 최근 매치 조회]
   - GetMatchesByNameV3(region, name, tag)
   - 또는 GetMatchesByPUUIDv3(region, puuid)
        │
        ▼
[각 매치에 대해 시간 범위 필터링]
   - match.metadata.game_start >= ScheduledStartTime
   - match.metadata.game_start <= ScheduledStartTime + DetectionWindow
        │
        ▼
[시간 범위 내 매치의 참가자 목록 확인]
   - match.players.all_players 에서 name#tag 매칭
   - Team A 전원 + Team B 전원 포함 여부 검증
        │
        ▼
[조건 충족 매치 발견]
   - 해당 매치의 승패 결과 추출
   - DetectedMatchID에 매치 ID 저장
```

**폴링 주기:** 3분 간격으로 VAPI를 조회한다 (API Rate Limit 고려).

**재시도 정책:**
- Detection Window가 만료될 때까지 폴링 지속.
- Window 만료 후에도 매치가 감지되지 않으면 `DetectionStatus → FAILED`.
- VAPI 호출 실패 시 다음 폴링 주기에 재시도.

#### 4.3.3 매치 검증 조건

매치가 Tournament Game의 결과로 인정되기 위한 조건:

| 조건 | 설명 |
|------|------|
| 시간 범위 | `ScheduledStartTime` ~ `ScheduledStartTime + DetectionWindow` 이내 |
| 멤버 포함 | 양 팀의 **전체** 멤버가 매치 참가자에 포함 |
| 게임 모드 | Custom Game (커스텀 게임) 모드 |
| 매치 완료 | 매치가 정상 종료된 상태 (중도 포기 X) |

#### 4.3.4 엣지 케이스 처리

| 케이스 | 처리 방안 |
|--------|-----------|
| 멤버가 Valorant 계정을 연동하지 않음 | Game 시작 시간 설정 시 사전 검증, 미연동 멤버 존재 시 경고 |
| Detection Window 내 매치가 없음 | `FAILED` 상태 전환, Staff에게 수동 입력 유도 |
| 같은 시간대에 여러 매치 존재 | 전원 참여 매치만 필터링, 복수 시 가장 늦게 시작된 매치 선택 |
| VAPI 장애/Rate Limit | 재시도 큐에 등록, 지수 백오프 적용 |
| 팀원 일부만 참여한 매치 | 무시 (전원 참여만 인정) |
| 매치가 리메이크/닷지 | 매치 완료 상태가 아니면 무시 |

---

### 4.4 결과 자동 기입

#### 4.4.1 결과 기입 프로세스

```
[매치 감지 성공]
        │
        ▼
[매치 결과 파싱]
   - 승리팀/패배팀 판별
   - 라운드 스코어 추출
   - 개인별 K/D/A, 에이전트 정보 추출
        │
        ▼
[GameTeam Grade 설정]
   - 승리팀 GameTeam: Grade = 1
   - 패배팀 GameTeam: Grade = 2
        │
        ▼
[Game 상태 갱신]
   - GameStatus: ACTIVE → FINISHED
   - DetectionStatus: DETECTING → DETECTED
   - DetectedMatchID: 매치 ID 저장
        │
        ▼
[다음 라운드 자동 진행]
   - NextGameID가 설정된 경우
   - 승리팀을 다음 Game의 GameTeam으로 자동 등록
        │
        ▼
[이벤트 발행]
   - game.match.detected (매치 감지 알림)
   - game.finished (게임 종료 알림)
   - RabbitMQ routing key: game.<event_type>
```

#### 4.4.2 Match Result 엔티티 (신규)

Tournament Game에 연결된 매치의 상세 결과를 저장한다.

```go
type MatchResult struct {
    MatchResultID   uint      `gorm:"primaryKey"`
    GameID          uint      `gorm:"not null;index"`
    ValorantMatchID string    `gorm:"not null"`
    MapName         string
    RoundsPlayed    int
    WinnerTeamID    uint      `gorm:"not null"`
    LoserTeamID     uint      `gorm:"not null"`
    WinnerScore     int       // 승리팀 라운드 수
    LoserScore      int       // 패배팀 라운드 수
    GameStartedAt   time.Time // Valorant 매치 실제 시작 시간
    GameDuration    int       // 매치 진행 시간 (초)
    CreatedAt       time.Time
}
```

#### 4.4.3 Match Player Stat 엔티티 (신규)

매치 참가자 개인별 스탯을 저장한다.

```go
type MatchPlayerStat struct {
    MatchPlayerStatID uint   `gorm:"primaryKey"`
    MatchResultID     uint   `gorm:"not null;index"`
    UserID            uint   `gorm:"not null;index"`
    TeamID            uint   `gorm:"not null"`
    AgentName         string // 사용한 에이전트
    Kills             int
    Deaths            int
    Assists           int
    Score             int
    Headshots         int
    Bodyshots         int
    Legshots          int
}
```

---

### 4.5 Staff 수동 결과 입력 (Fallback)

#### 4.5.1 요구사항

- 자동 감지 실패(`DetectionStatus=FAILED`) 시 Staff가 수동으로 결과를 입력할 수 있다.
- 자동 감지 중에도 Staff가 수동으로 결과를 override할 수 있다.

#### 4.5.2 API 설계

```
POST /api/v1/contests/{contestId}/games/{gameId}/result
```

**Request Body:**
```json
{
  "winnerTeamId": 1,
  "winnerScore": 13,
  "loserScore": 8,
  "note": "수동 입력 사유 (선택)"
}
```

수동 입력 시 `DetectionStatus → MANUAL`로 변경된다.

---

## 5. 아키텍처 설계

### 5.1 신규 Port (인터페이스)

#### MatchDetectionPort

```go
type MatchDetectionPort interface {
    // 특정 플레이어의 최근 매치 히스토리 조회
    GetRecentMatches(region, name, tag string) ([]ValorantMatch, error)

    // 특정 매치의 상세 정보 조회
    GetMatchDetail(matchId string) (*ValorantMatchDetail, error)
}
```

#### MatchResultDatabasePort

```go
type MatchResultDatabasePort interface {
    Save(result *domain.MatchResult) error
    GetByGameID(gameID uint) (*domain.MatchResult, error)
    SavePlayerStats(stats []domain.MatchPlayerStat) error
    GetPlayerStatsByMatchResult(matchResultID uint) ([]domain.MatchPlayerStat, error)
}
```

#### GameSchedulerPort

```go
type GameSchedulerPort interface {
    // 시작 시간이 도래한 Game 목록 조회
    GetGamesReadyToStart() ([]domain.Game, error)

    // Detection 중인 Game 목록 조회
    GetGamesInDetection() ([]domain.Game, error)
}
```

### 5.2 신규 Service

#### MatchDetectionService

```go
type MatchDetectionService struct {
    matchDetectionPort  port.MatchDetectionPort
    gameDatabasePort    port.GameDatabasePort
    gameTeamDBPort      port.GameTeamDatabasePort
    teamDBPort          port.TeamDatabasePort
    matchResultDBPort   port.MatchResultDatabasePort
    eventPublisher      port.GameEventPublisherPort
}
```

**주요 메서드:**
- `DetectMatchForGame(gameID uint) error` - 단일 Game에 대한 매치 감지 실행
- `ProcessDetectedMatch(game, match) error` - 감지된 매치 결과 처리
- `ValidateMatchParticipants(match, teamAMembers, teamBMembers) bool` - 참가자 검증

#### GameSchedulerService

```go
type GameSchedulerService struct {
    gameDBPort           port.GameDatabasePort
    matchDetectionSvc    *MatchDetectionService
}
```

**주요 메서드:**
- `RunScheduledActivation()` - 시작 시간 도래한 Game 활성화 (Cron)
- `RunMatchDetection()` - Detection 중인 Game들의 매치 감지 실행 (Cron)

### 5.3 Scheduler 구조

```
[Application Startup]
        │
        ▼
[Cron Scheduler 등록]
   ├── @every 1m : GameSchedulerService.RunScheduledActivation()
   └── @every 3m : GameSchedulerService.RunMatchDetection()
        │
        ▼
[각 Cron Job은 독립적으로 실행]
   - 동시 실행 방지를 위한 분산 락 (Redis) 적용
   - 다중 인스턴스 환경 대응
```

### 5.4 디렉토리 구조 (신규/변경)

```
internal/game/
├── domain/
│   ├── game.go                    # [변경] ScheduledStartTime 등 필드 추가
│   ├── match_result.go            # [신규] MatchResult 엔티티
│   └── match_player_stat.go       # [신규] MatchPlayerStat 엔티티
├── application/
│   ├── game_service.go            # [변경] ScheduleGame, ManualResult 메서드 추가
│   ├── match_detection_service.go # [신규] 매치 감지 서비스
│   ├── game_scheduler_service.go  # [신규] 스케줄러 서비스
│   ├── port/
│   │   ├── match_detection_port.go      # [신규]
│   │   ├── match_result_database_port.go # [신규]
│   │   └── game_event_publisher_port.go  # [신규]
│   └── dto/
│       ├── schedule_game_request.go      # [신규]
│       └── match_result_response.go      # [신규]
├── infra/
│   └── persistence/
│       └── adapter/
│           ├── match_detection_valorant_adapter.go    # [신규] VAPI 연동
│           ├── match_result_database_adapter.go       # [신규]
│           └── game_event_publisher_rabbitmq_adapter.go # [신규]
└── presentation/
    └── game_controller.go         # [변경] schedule, result 엔드포인트 추가

internal/valorant/
├── application/
│   └── port/
│       └── valorant_api_port.go   # [변경] Match 관련 메서드 추가
└── infra/
    └── valorant_api_client.go     # [변경] GetMatchesByName, GetMatch 메서드 추가
```

---

## 6. 데이터베이스 변경

### 6.1 games 테이블 변경

```sql
ALTER TABLE games
    ADD COLUMN scheduled_start_time DATETIME NULL,
    ADD COLUMN detection_window_minutes INT NOT NULL DEFAULT 120,
    ADD COLUMN detected_match_id VARCHAR(255) NULL,
    ADD COLUMN detection_status VARCHAR(20) NOT NULL DEFAULT 'NONE';

CREATE INDEX idx_games_detection ON games(game_status, detection_status);
CREATE INDEX idx_games_scheduled ON games(scheduled_start_time, game_status);
```

### 6.2 match_results 테이블 (신규)

```sql
CREATE TABLE match_results (
    match_result_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    game_id         BIGINT UNSIGNED NOT NULL,
    valorant_match_id VARCHAR(255) NOT NULL,
    map_name        VARCHAR(50),
    rounds_played   INT NOT NULL,
    winner_team_id  BIGINT UNSIGNED NOT NULL,
    loser_team_id   BIGINT UNSIGNED NOT NULL,
    winner_score    INT NOT NULL,
    loser_score     INT NOT NULL,
    game_started_at DATETIME NOT NULL,
    game_duration   INT NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE INDEX idx_match_results_game (game_id),
    INDEX idx_match_results_valorant (valorant_match_id),
    CONSTRAINT fk_match_results_game FOREIGN KEY (game_id) REFERENCES games(game_id),
    CONSTRAINT fk_match_results_winner FOREIGN KEY (winner_team_id) REFERENCES teams(team_id),
    CONSTRAINT fk_match_results_loser FOREIGN KEY (loser_team_id) REFERENCES teams(team_id)
);
```

### 6.3 match_player_stats 테이블 (신규)

```sql
CREATE TABLE match_player_stats (
    match_player_stat_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    match_result_id      BIGINT UNSIGNED NOT NULL,
    user_id              BIGINT UNSIGNED NOT NULL,
    team_id              BIGINT UNSIGNED NOT NULL,
    agent_name           VARCHAR(50),
    kills                INT NOT NULL DEFAULT 0,
    deaths               INT NOT NULL DEFAULT 0,
    assists              INT NOT NULL DEFAULT 0,
    score                INT NOT NULL DEFAULT 0,
    headshots            INT NOT NULL DEFAULT 0,
    bodyshots            INT NOT NULL DEFAULT 0,
    legshots             INT NOT NULL DEFAULT 0,

    INDEX idx_match_player_stats_result (match_result_id),
    INDEX idx_match_player_stats_user (user_id),
    CONSTRAINT fk_match_player_stats_result FOREIGN KEY (match_result_id) REFERENCES match_results(match_result_id),
    CONSTRAINT fk_match_player_stats_user FOREIGN KEY (user_id) REFERENCES users(user_id),
    CONSTRAINT fk_match_player_stats_team FOREIGN KEY (team_id) REFERENCES teams(team_id)
);
```

---

## 7. 이벤트 설계

### 7.1 신규 이벤트

| Routing Key | 이벤트 | 발행 시점 |
|-------------|--------|-----------|
| `game.scheduled` | GameScheduledEvent | Staff가 Game 시작 시간 설정 시 |
| `game.activated` | GameActivatedEvent | 예정 시간 도달로 Game 활성화 시 |
| `game.match.detecting` | MatchDetectingEvent | 매치 감지 시작 시 |
| `game.match.detected` | MatchDetectedEvent | 매치 감지 성공 시 |
| `game.match.failed` | MatchDetectionFailedEvent | Detection Window 만료 시 |
| `game.finished` | GameFinishedEvent | Game 결과 확정 시 |
| `game.result.manual` | ManualResultEvent | Staff 수동 결과 입력 시 |

### 7.2 이벤트 페이로드 예시

```json
// game.match.detected
{
  "eventId": "uuid",
  "eventType": "game.match.detected",
  "timestamp": "2025-02-15T14:45:00Z",
  "contestId": 1,
  "gameId": 5,
  "round": 1,
  "matchNumber": 2,
  "valorantMatchId": "abc-123-def",
  "winnerTeamId": 3,
  "winnerTeamName": "Team Alpha",
  "loserTeamId": 7,
  "loserTeamName": "Team Beta",
  "score": "13-8",
  "mapName": "Ascent"
}
```

---

## 8. API 엔드포인트 요약

| Method | Endpoint | 설명 | 권한 |
|--------|----------|------|------|
| `PUT` | `/api/v1/contests/{contestId}/games/{gameId}/schedule` | Game 시작 시간 설정 | Staff |
| `GET` | `/api/v1/contests/{contestId}/games/{gameId}/detection-status` | 매치 감지 상태 조회 | Staff, Normal |
| `POST` | `/api/v1/contests/{contestId}/games/{gameId}/result` | 수동 결과 입력 | Staff |
| `GET` | `/api/v1/contests/{contestId}/games/{gameId}/result` | 매치 결과 조회 | Staff, Normal |
| `GET` | `/api/v1/contests/{contestId}/games/{gameId}/result/stats` | 개인별 스탯 조회 | Staff, Normal |
| `POST` | `/api/v1/contests/{contestId}/games/{gameId}/detect` | 수동 매치 감지 트리거 | Staff |

---

## 9. 사전 조건 및 의존성

### 9.1 Valorant 계정 연동 필수

매치 감지를 위해 모든 팀 멤버는 Valorant 계정(name + tag)이 연동되어 있어야 한다.
Game 시작 시간 설정 시 미연동 멤버가 존재하면 경고를 반환한다.

### 9.2 VAPI Rate Limit 고려

- Henrik Dev Valorant API는 요청 제한이 있다.
- 동시에 다수의 Game이 Detection 중일 경우 요청 큐잉이 필요하다.
- API Key 설정 시 더 높은 Rate Limit을 확보할 수 있다.

### 9.3 시간대 처리

- 모든 시간은 UTC로 저장한다.
- VAPI 매치의 `game_start` 타임스탬프와 비교 시 UTC 기준으로 통일한다.
- 클라이언트에게 응답 시 요청 헤더의 timezone을 참고하여 변환한다.

---

## 10. 확장 고려사항

| 항목 | 설명 |
|------|------|
| LOL 지원 | `GameType`에 따라 MatchDetectionPort 구현체를 분기 (Strategy 패턴) |
| Best of 3/5 | 단일 매치가 아닌 시리즈 매치 감지로 확장 가능 |
| Discord 알림 | 매치 감지/결과 확정 시 Discord 채널에 자동 알림 발송 |
| 포인트 자동 계산 | 매치 결과 기반 ContestMember의 Point 자동 갱신 |
| 통계 대시보드 | MatchPlayerStat 기반 대회 전체 통계 집계 |
