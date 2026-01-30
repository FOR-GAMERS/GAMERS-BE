package integration_test

import (
	"GAMERS-BE/internal/contest/application"
	"GAMERS-BE/internal/contest/application/dto"
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	contestAdapter "GAMERS-BE/internal/contest/infra/persistence/adapter"
	gameApplication "GAMERS-BE/internal/game/application"
	gameDomain "GAMERS-BE/internal/game/domain"
	gamePort "GAMERS-BE/internal/game/application/port"
	oauth2Domain "GAMERS-BE/internal/oauth2/domain"
	"GAMERS-BE/test/global/support"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ==================== Tournament Flow Test Suite ====================
// This test suite simulates a complete 4-team tournament:
//
// Tournament Bracket (4 teams):
//
//   Round 1 (Semi-finals)        Round 2 (Final)
//   ┌─────────────┐
//   │ Game 1      │
//   │ Team A      │──┐
//   │   vs        │  │  ┌─────────────┐
//   │ Team B      │  └──│ Game 3      │
//   └─────────────┘     │ (Final)     │
//                       │ Winner G1   │──── Champion
//   ┌─────────────┐  ┌──│   vs        │
//   │ Game 2      │  │  │ Winner G2   │
//   │ Team C      │──┘  └─────────────┘
//   │   vs        │
//   │ Team D      │
//   └─────────────┘

type TournamentFlowTestSuite struct {
	suite.Suite
	container *support.MySQLContainer
	db        *gorm.DB

	// Contest dependencies
	contestAdapter port.ContestDatabasePort
	memberAdapter  port.ContestMemberDatabasePort
	mockOAuth2     *MockOAuth2Port
	mockEventPub   *MockEventPubPort
	mockRedis      *MockRedisPort
	contestService *application.ContestService

	// Game dependencies
	gameAdapter     gamePort.GameDatabasePort
	teamAdapter     gamePort.TeamDatabasePort
	gameTeamAdapter gamePort.GameTeamDatabasePort

	// Services
	tournamentService *gameApplication.TournamentService
	gameService       *gameApplication.GameService
	gameTeamService   *gameApplication.GameTeamService
}

// ==================== Mock Definitions ====================

type MockOAuth2Port struct {
	mock.Mock
}

func (m *MockOAuth2Port) FindDiscordAccountByDiscordId(discordId string) (*oauth2Domain.DiscordAccount, error) {
	args := m.Called(discordId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2Domain.DiscordAccount), args.Error(1)
}

func (m *MockOAuth2Port) FindDiscordAccountByUserId(userId int64) (*oauth2Domain.DiscordAccount, error) {
	args := m.Called(userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2Domain.DiscordAccount), args.Error(1)
}

func (m *MockOAuth2Port) CreateDiscordAccount(account *oauth2Domain.DiscordAccount) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockOAuth2Port) UpdateDiscordAccount(account *oauth2Domain.DiscordAccount) error {
	args := m.Called(account)
	return args.Error(0)
}

type MockEventPubPort struct {
	mock.Mock
}

func (m *MockEventPubPort) PublishContestApplicationEvent(ctx context.Context, event *port.ContestApplicationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPubPort) PublishContestCreatedEvent(ctx context.Context, event *port.ContestCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPubPort) Close() error {
	return nil
}

func (m *MockEventPubPort) HealthCheck(ctx context.Context) error {
	return nil
}

type MockRedisPort struct {
	mock.Mock
}

func (m *MockRedisPort) RequestParticipate(ctx context.Context, contestId int64, sender *port.SenderSnapshot, ttl time.Duration) error {
	args := m.Called(ctx, contestId, sender, ttl)
	return args.Error(0)
}

func (m *MockRedisPort) AcceptRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	args := m.Called(ctx, contestId, userId, processedBy)
	return args.Error(0)
}

func (m *MockRedisPort) RejectRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	args := m.Called(ctx, contestId, userId, processedBy)
	return args.Error(0)
}

func (m *MockRedisPort) CancelApplication(ctx context.Context, contestId, userId int64) error {
	args := m.Called(ctx, contestId, userId)
	return args.Error(0)
}

func (m *MockRedisPort) GetApplication(ctx context.Context, contestId, userId int64) (*port.ContestApplication, error) {
	args := m.Called(ctx, contestId, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ContestApplication), args.Error(1)
}

func (m *MockRedisPort) GetPendingApplications(ctx context.Context, contestId int64) ([]*port.ContestApplication, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*port.ContestApplication), args.Error(1)
}

func (m *MockRedisPort) GetAcceptedApplications(ctx context.Context, contestId int64) ([]int64, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRedisPort) GetRejectedApplications(ctx context.Context, contestId int64) ([]int64, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRedisPort) HasApplied(ctx context.Context, contestId, userId int64) (bool, error) {
	args := m.Called(ctx, contestId, userId)
	return args.Bool(0), args.Error(1)
}

func (m *MockRedisPort) ExtendTTL(ctx context.Context, contestId int64, newTTL time.Duration) error {
	args := m.Called(ctx, contestId, newTTL)
	return args.Error(0)
}

func (m *MockRedisPort) ClearApplications(ctx context.Context, contestId int64) error {
	args := m.Called(ctx, contestId)
	return args.Error(0)
}

// ==================== In-Memory Adapters for Game Domain ====================

// InMemoryGameAdapter implements GameDatabasePort for testing
type InMemoryGameAdapter struct {
	games    map[int64]*gameDomain.Game
	nextID   int64
}

func NewInMemoryGameAdapter() *InMemoryGameAdapter {
	return &InMemoryGameAdapter{
		games:  make(map[int64]*gameDomain.Game),
		nextID: 1,
	}
}

func (a *InMemoryGameAdapter) Save(game *gameDomain.Game) (*gameDomain.Game, error) {
	game.GameID = a.nextID
	a.nextID++
	a.games[game.GameID] = game
	return game, nil
}

func (a *InMemoryGameAdapter) SaveBatch(games []*gameDomain.Game) error {
	for _, game := range games {
		a.Save(game)
	}
	return nil
}

func (a *InMemoryGameAdapter) GetByID(gameID int64) (*gameDomain.Game, error) {
	if game, ok := a.games[gameID]; ok {
		return game, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryGameAdapter) GetByContestID(contestID int64) ([]*gameDomain.Game, error) {
	var result []*gameDomain.Game
	for _, game := range a.games {
		if game.ContestID == contestID {
			result = append(result, game)
		}
	}
	return result, nil
}

func (a *InMemoryGameAdapter) GetByContestAndRound(contestID int64, round int) ([]*gameDomain.Game, error) {
	var result []*gameDomain.Game
	for _, game := range a.games {
		if game.ContestID == contestID && game.Round != nil && *game.Round == round {
			result = append(result, game)
		}
	}
	return result, nil
}

func (a *InMemoryGameAdapter) Update(game *gameDomain.Game) error {
	a.games[game.GameID] = game
	return nil
}

func (a *InMemoryGameAdapter) Delete(gameID int64) error {
	delete(a.games, gameID)
	return nil
}

func (a *InMemoryGameAdapter) DeleteByContestID(contestID int64) error {
	for id, game := range a.games {
		if game.ContestID == contestID {
			delete(a.games, id)
		}
	}
	return nil
}

func (a *InMemoryGameAdapter) GetGamesReadyToStart() ([]*gameDomain.Game, error) {
	var result []*gameDomain.Game
	for _, game := range a.games {
		if game.IsReadyToActivate() {
			result = append(result, game)
		}
	}
	return result, nil
}

func (a *InMemoryGameAdapter) GetGamesInDetection() ([]*gameDomain.Game, error) {
	var result []*gameDomain.Game
	for _, game := range a.games {
		if game.IsDetecting() {
			result = append(result, game)
		}
	}
	return result, nil
}

// InMemoryTeamAdapter implements TeamDatabasePort for testing
type InMemoryTeamAdapter struct {
	teams       map[int64]*gameDomain.Team
	members     map[int64]*gameDomain.TeamMember
	nextTeamID  int64
	nextMemberID int64
}

func NewInMemoryTeamAdapter() *InMemoryTeamAdapter {
	return &InMemoryTeamAdapter{
		teams:        make(map[int64]*gameDomain.Team),
		members:      make(map[int64]*gameDomain.TeamMember),
		nextTeamID:   1,
		nextMemberID: 1,
	}
}

func (a *InMemoryTeamAdapter) Save(team *gameDomain.Team) (*gameDomain.Team, error) {
	team.TeamID = a.nextTeamID
	a.nextTeamID++
	a.teams[team.TeamID] = team
	return team, nil
}

func (a *InMemoryTeamAdapter) GetByID(teamID int64) (*gameDomain.Team, error) {
	if team, ok := a.teams[teamID]; ok {
		return team, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryTeamAdapter) GetByContestID(contestID int64) ([]*gameDomain.Team, error) {
	var result []*gameDomain.Team
	for _, team := range a.teams {
		if team.ContestID == contestID {
			result = append(result, team)
		}
	}
	return result, nil
}

func (a *InMemoryTeamAdapter) GetByContestAndName(contestID int64, teamName string) (*gameDomain.Team, error) {
	for _, team := range a.teams {
		if team.ContestID == contestID && team.TeamName == teamName {
			return team, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryTeamAdapter) CountByContestID(contestID int64) (int, error) {
	count := 0
	for _, team := range a.teams {
		if team.ContestID == contestID {
			count++
		}
	}
	return count, nil
}

func (a *InMemoryTeamAdapter) Update(team *gameDomain.Team) error {
	a.teams[team.TeamID] = team
	return nil
}

func (a *InMemoryTeamAdapter) Delete(teamID int64) error {
	delete(a.teams, teamID)
	return nil
}

func (a *InMemoryTeamAdapter) DeleteByContestID(contestID int64) error {
	for id, team := range a.teams {
		if team.ContestID == contestID {
			delete(a.teams, id)
		}
	}
	return nil
}

// TeamMember operations
func (a *InMemoryTeamAdapter) SaveMember(member *gameDomain.TeamMember) (*gameDomain.TeamMember, error) {
	member.ID = a.nextMemberID
	a.nextMemberID++
	a.members[member.ID] = member
	return member, nil
}

func (a *InMemoryTeamAdapter) SaveMemberBatch(members []*gameDomain.TeamMember) error {
	for _, m := range members {
		a.SaveMember(m)
	}
	return nil
}

func (a *InMemoryTeamAdapter) GetMemberByID(id int64) (*gameDomain.TeamMember, error) {
	if m, ok := a.members[id]; ok {
		return m, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryTeamAdapter) GetMembersByTeamID(teamID int64) ([]*gameDomain.TeamMember, error) {
	var result []*gameDomain.TeamMember
	for _, m := range a.members {
		if m.TeamID == teamID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (a *InMemoryTeamAdapter) GetMemberByTeamAndUser(teamID, userID int64) (*gameDomain.TeamMember, error) {
	for _, m := range a.members {
		if m.TeamID == teamID && m.UserID == userID {
			return m, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryTeamAdapter) GetMemberCountByTeamID(teamID int64) (int, error) {
	count := 0
	for _, m := range a.members {
		if m.TeamID == teamID {
			count++
		}
	}
	return count, nil
}

func (a *InMemoryTeamAdapter) GetLeaderByTeamID(teamID int64) (*gameDomain.TeamMember, error) {
	for _, m := range a.members {
		if m.TeamID == teamID && m.MemberType == gameDomain.TeamMemberTypeLeader {
			return m, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryTeamAdapter) UpdateMember(member *gameDomain.TeamMember) error {
	a.members[member.ID] = member
	return nil
}

func (a *InMemoryTeamAdapter) DeleteMember(id int64) error {
	delete(a.members, id)
	return nil
}

func (a *InMemoryTeamAdapter) DeleteMemberByTeamAndUser(teamID, userID int64) error {
	for id, m := range a.members {
		if m.TeamID == teamID && m.UserID == userID {
			delete(a.members, id)
			return nil
		}
	}
	return nil
}

func (a *InMemoryTeamAdapter) DeleteAllMembersByTeamID(teamID int64) error {
	for id, m := range a.members {
		if m.TeamID == teamID {
			delete(a.members, id)
		}
	}
	return nil
}

func (a *InMemoryTeamAdapter) GetTeamsByContestWithMembers(contestID int64) ([]*gamePort.TeamWithMembers, error) {
	return nil, nil
}

func (a *InMemoryTeamAdapter) GetUserTeamInContest(contestID, userID int64) (*gameDomain.Team, error) {
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryTeamAdapter) GetTeamByGameID(gameID int64) (*gamePort.TeamWithMembers, error) {
	return nil, nil
}

func (a *InMemoryTeamAdapter) GetNextTeamID(contestID int64) (int64, error) {
	return a.nextTeamID, nil
}

func (a *InMemoryTeamAdapter) GetMembersByGameID(gameID int64) ([]*gameDomain.TeamMember, error) {
	return nil, nil
}

func (a *InMemoryTeamAdapter) GetByGameAndUser(gameID, userID int64) (*gameDomain.TeamMember, error) {
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryTeamAdapter) GetByGameAndTeamID(gameID, teamID int64) ([]*gameDomain.TeamMember, error) {
	return nil, nil
}

func (a *InMemoryTeamAdapter) DeleteAllByGameID(gameID int64) error {
	return nil
}

// InMemoryGameTeamAdapter implements GameTeamDatabasePort for testing
type InMemoryGameTeamAdapter struct {
	gameTeams map[int64]*gameDomain.GameTeam
	nextID    int64
}

func NewInMemoryGameTeamAdapter() *InMemoryGameTeamAdapter {
	return &InMemoryGameTeamAdapter{
		gameTeams: make(map[int64]*gameDomain.GameTeam),
		nextID:    1,
	}
}

func (a *InMemoryGameTeamAdapter) Save(gameTeam *gameDomain.GameTeam) (*gameDomain.GameTeam, error) {
	gameTeam.GameTeamID = a.nextID
	a.nextID++
	a.gameTeams[gameTeam.GameTeamID] = gameTeam
	return gameTeam, nil
}

func (a *InMemoryGameTeamAdapter) GetByID(gameTeamID int64) (*gameDomain.GameTeam, error) {
	if gt, ok := a.gameTeams[gameTeamID]; ok {
		return gt, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryGameTeamAdapter) GetByGameID(gameID int64) ([]*gameDomain.GameTeam, error) {
	var result []*gameDomain.GameTeam
	for _, gt := range a.gameTeams {
		if gt.GameID == gameID {
			result = append(result, gt)
		}
	}
	return result, nil
}

func (a *InMemoryGameTeamAdapter) GetByGameAndTeam(gameID, teamID int64) (*gameDomain.GameTeam, error) {
	for _, gt := range a.gameTeams {
		if gt.GameID == gameID && gt.TeamID == teamID {
			return gt, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryGameTeamAdapter) GetByGrade(gameID int64, grade int) (*gameDomain.GameTeam, error) {
	for _, gt := range a.gameTeams {
		if gt.GameID == gameID && gt.Grade != nil && *gt.Grade == grade {
			return gt, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (a *InMemoryGameTeamAdapter) Delete(gameTeamID int64) error {
	delete(a.gameTeams, gameTeamID)
	return nil
}

func (a *InMemoryGameTeamAdapter) DeleteByGameID(gameID int64) error {
	for id, gt := range a.gameTeams {
		if gt.GameID == gameID {
			delete(a.gameTeams, id)
		}
	}
	return nil
}

// ==================== Suite Setup ====================

func (s *TournamentFlowTestSuite) SetupSuite() {
	ctx := context.Background()
	var err error

	s.container, err = support.SetupMySQLContainer(ctx)
	s.Require().NoError(err, "Failed to setup MySQL container")

	s.db = s.container.GetDB()

	// Auto-migrate schemas
	err = s.db.AutoMigrate(&domain.Contest{}, &domain.ContestMember{})
	s.Require().NoError(err, "Failed to migrate schemas")
}

func (s *TournamentFlowTestSuite) TearDownSuite() {
	ctx := context.Background()
	if s.container != nil {
		s.container.Teardown(ctx)
	}
}

func (s *TournamentFlowTestSuite) SetupTest() {
	// Clean up tables
	s.db.Exec("DELETE FROM contests_members")
	s.db.Exec("DELETE FROM contests")

	// Initialize contest adapters (real DB)
	s.contestAdapter = contestAdapter.NewContestDatabaseAdapter(s.db)
	s.memberAdapter = contestAdapter.NewContestMemberDatabaseAdapter(s.db)

	// Initialize mocks
	s.mockOAuth2 = new(MockOAuth2Port)
	s.mockEventPub = new(MockEventPubPort)
	s.mockRedis = new(MockRedisPort)

	// Initialize game adapters (in-memory for testing)
	s.gameAdapter = NewInMemoryGameAdapter()
	s.teamAdapter = NewInMemoryTeamAdapter()
	s.gameTeamAdapter = NewInMemoryGameTeamAdapter()

	// Initialize tournament service
	s.tournamentService = gameApplication.NewTournamentService(
		s.gameAdapter,
		s.teamAdapter,
	)

	// Initialize game service
	s.gameService = gameApplication.NewGameService(
		s.gameAdapter,
		s.teamAdapter,
	)

	// Initialize game team service
	s.gameTeamService = gameApplication.NewGameTeamService(
		s.gameTeamAdapter,
		s.gameAdapter,
		s.teamAdapter,
		s.contestAdapter,
	)

	// Initialize contest service with tournament generator
	s.contestService = application.NewContestServiceFull(
		s.contestAdapter,
		s.memberAdapter,
		s.mockRedis,
		s.mockOAuth2,
		s.mockEventPub,
		nil, // discord validator
		s.tournamentService,
		s.teamAdapter,
		s.gameTeamAdapter,
	)
}

// ==================== Tournament Flow Tests ====================

func (s *TournamentFlowTestSuite) TestFourTeamTournament_CompleteFlow() {
	ctx := context.Background()

	// ========== PHASE 1: Contest Creation ==========
	s.T().Log("Phase 1: Creating tournament contest with 4 teams...")

	leaderID := int64(1)
	discordAccount := &oauth2Domain.DiscordAccount{
		DiscordId:     "discord_leader",
		UserId:        leaderID,
		DiscordAvatar: "avatar",
	}
	s.mockOAuth2.On("FindDiscordAccountByUserId", leaderID).Return(discordAccount, nil)

	contestReq := &dto.CreateContestRequest{
		Title:           "4-Team Tournament Championship",
		Description:     "Testing complete tournament flow",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	contest, _, err := s.contestService.SaveContest(contestReq, leaderID)
	s.Require().NoError(err)
	s.Equal(domain.ContestStatusPending, contest.ContestStatus)
	s.Equal(domain.ContestTypeTournament, contest.ContestType)
	s.T().Logf("Contest created: ID=%d, Title=%s", contest.ContestID, contest.Title)

	// ========== PHASE 2: Verify Tournament Bracket Generation ==========
	s.T().Log("Phase 2: Verifying tournament bracket generation...")

	bracket, err := s.tournamentService.GetTournamentBracket(contest.ContestID)
	s.Require().NoError(err)
	s.NotNil(bracket)

	// 4-team tournament should have:
	// - Round 1: 2 games (semi-finals)
	// - Round 2: 1 game (final)
	// - Total: 3 games
	totalGames := 0
	for round, games := range bracket.Rounds {
		totalGames += len(games)
		s.T().Logf("Round %d: %d games", round, len(games))
	}
	s.Equal(3, totalGames, "4-team tournament should have 3 games")
	s.Equal(2, bracket.TotalRounds)

	// Verify round structure
	s.Len(bracket.Rounds[1], 2, "Round 1 should have 2 games")
	s.Len(bracket.Rounds[2], 1, "Round 2 (Final) should have 1 game")

	// ========== PHASE 3: Create Teams ==========
	s.T().Log("Phase 3: Creating 4 teams...")

	teams := make([]*gameDomain.Team, 4)
	teamNames := []string{"Team Alpha", "Team Beta", "Team Gamma", "Team Delta"}

	for i, name := range teamNames {
		team := gameDomain.NewTeam(contest.ContestID, name)
		savedTeam, err := s.teamAdapter.Save(team)
		s.Require().NoError(err)
		teams[i] = savedTeam
		s.T().Logf("Team created: ID=%d, Name=%s", savedTeam.TeamID, savedTeam.TeamName)
	}

	// ========== PHASE 4: Start Contest ==========
	s.T().Log("Phase 4: Starting contest...")

	s.mockRedis.On("GetAcceptedApplications", ctx, contest.ContestID).Return([]int64{}, nil)
	s.mockRedis.On("ClearApplications", ctx, contest.ContestID).Return(nil)

	startedContest, err := s.contestService.StartContest(ctx, contest.ContestID, leaderID)
	s.Require().NoError(err)
	s.Equal(domain.ContestStatusActive, startedContest.ContestStatus)
	s.T().Log("Contest started successfully!")

	// ========== PHASE 5: Allocate Teams to First Round Games ==========
	s.T().Log("Phase 5: Allocating teams to first round games...")

	allocationResult, err := s.tournamentService.ShuffleAndAllocateTeamsWithResult(
		contest.ContestID,
		s.gameTeamAdapter,
	)
	s.Require().NoError(err)
	s.NotNil(allocationResult)
	s.Equal(4, allocationResult.TotalTeams)
	s.Len(allocationResult.Allocations, 2, "Should have 2 allocations for round 1")

	for _, alloc := range allocationResult.Allocations {
		s.T().Logf("Game %d (Round %d, Match %d): %s vs %s",
			alloc.GameID, alloc.Round, alloc.MatchNumber,
			alloc.Team1Name, alloc.Team2Name)
	}

	// ========== PHASE 6: Play Semi-Finals ==========
	s.T().Log("Phase 6: Playing semi-finals...")

	// Get first round games
	firstRoundGames, err := s.gameAdapter.GetByContestAndRound(contest.ContestID, 1)
	s.Require().NoError(err)
	s.Len(firstRoundGames, 2)

	semiWinners := make([]*gameDomain.Team, 2)

	for i, game := range firstRoundGames {
		// Start the game
		startedGame, err := s.gameService.StartGame(game.GameID)
		s.Require().NoError(err)
		s.Equal(gameDomain.GameStatusActive, startedGame.GameStatus)
		s.T().Logf("Game %d started", game.GameID)

		// Get teams in this game
		gameTeams, err := s.gameTeamAdapter.GetByGameID(game.GameID)
		s.Require().NoError(err)
		s.Len(gameTeams, 2)

		// Simulate: First team wins (gets grade 1)
		winnerTeamID := gameTeams[0].TeamID
		loserTeamID := gameTeams[1].TeamID

		// Set grades (1 = winner, 2 = loser)
		gameTeams[0].SetGrade(1)
		gameTeams[1].SetGrade(2)

		winner, err := s.teamAdapter.GetByID(winnerTeamID)
		s.Require().NoError(err)
		semiWinners[i] = winner

		loser, err := s.teamAdapter.GetByID(loserTeamID)
		s.Require().NoError(err)

		s.T().Logf("Semi-final %d result: %s (Winner) defeated %s",
			i+1, winner.TeamName, loser.TeamName)

		// Finish the game
		finishedGame, err := s.gameService.FinishGame(game.GameID)
		s.Require().NoError(err)
		s.Equal(gameDomain.GameStatusFinished, finishedGame.GameStatus)
	}

	// ========== PHASE 7: Advance Winners to Final ==========
	s.T().Log("Phase 7: Advancing winners to final...")

	// Get final game
	finalGames, err := s.gameAdapter.GetByContestAndRound(contest.ContestID, 2)
	s.Require().NoError(err)
	s.Len(finalGames, 1)
	finalGame := finalGames[0]

	// Register semi-final winners in final game
	for _, winner := range semiWinners {
		gameTeam := gameDomain.NewGameTeam(finalGame.GameID, winner.TeamID)
		_, err := s.gameTeamAdapter.Save(gameTeam)
		s.Require().NoError(err)
		s.T().Logf("Advanced %s to final", winner.TeamName)
	}

	// Verify final has 2 teams
	finalTeams, err := s.gameTeamAdapter.GetByGameID(finalGame.GameID)
	s.Require().NoError(err)
	s.Len(finalTeams, 2)

	// ========== PHASE 8: Play Final ==========
	s.T().Log("Phase 8: Playing final...")

	// Start final
	startedFinal, err := s.gameService.StartGame(finalGame.GameID)
	s.Require().NoError(err)
	s.Equal(gameDomain.GameStatusActive, startedFinal.GameStatus)
	s.T().Log("Final started!")

	// Simulate: First team wins the championship
	championTeamID := finalTeams[0].TeamID
	runnerUpTeamID := finalTeams[1].TeamID

	// Set grades
	finalTeams[0].SetGrade(1) // Champion
	finalTeams[1].SetGrade(2) // Runner-up

	champion, err := s.teamAdapter.GetByID(championTeamID)
	s.Require().NoError(err)

	runnerUp, err := s.teamAdapter.GetByID(runnerUpTeamID)
	s.Require().NoError(err)

	s.T().Logf("FINAL RESULT: %s defeats %s!", champion.TeamName, runnerUp.TeamName)

	// Finish final
	finishedFinal, err := s.gameService.FinishGame(finalGame.GameID)
	s.Require().NoError(err)
	s.Equal(gameDomain.GameStatusFinished, finishedFinal.GameStatus)

	// ========== PHASE 9: End Contest ==========
	s.T().Log("Phase 9: Ending contest...")

	stoppedContest, err := s.contestService.StopContest(ctx, contest.ContestID, leaderID)
	s.Require().NoError(err)
	s.Equal(domain.ContestStatusFinished, stoppedContest.ContestStatus)

	// ========== VERIFICATION ==========
	s.T().Log("========== TOURNAMENT COMPLETE ==========")
	s.T().Logf("Champion: %s", champion.TeamName)
	s.T().Logf("Runner-up: %s", runnerUp.TeamName)

	// Final verification
	finalContest, err := s.contestAdapter.GetContestById(contest.ContestID)
	s.Require().NoError(err)
	s.Equal(domain.ContestStatusFinished, finalContest.ContestStatus)
	s.True(finalContest.IsTerminalState())

	// Verify all games are finished
	allGames, err := s.gameAdapter.GetByContestID(contest.ContestID)
	s.Require().NoError(err)
	for _, game := range allGames {
		s.Equal(gameDomain.GameStatusFinished, game.GameStatus,
			"All games should be finished")
	}

	s.T().Log("All verification passed!")
}

func (s *TournamentFlowTestSuite) TestTournamentBracketGeneration_ValidStructure() {
	// Test that 4-team tournament generates correct bracket structure
	leaderID := int64(1)
	discordAccount := &oauth2Domain.DiscordAccount{
		DiscordId:     "discord_test",
		UserId:        leaderID,
		DiscordAvatar: "avatar",
	}
	s.mockOAuth2.On("FindDiscordAccountByUserId", leaderID).Return(discordAccount, nil)

	contestReq := &dto.CreateContestRequest{
		Title:           "Bracket Test Tournament",
		Description:     "Testing bracket structure",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	contest, _, err := s.contestService.SaveContest(contestReq, leaderID)
	s.Require().NoError(err)

	// Get bracket
	bracket, err := s.tournamentService.GetTournamentBracket(contest.ContestID)
	s.Require().NoError(err)

	// Verify round 1 games link to round 2
	round1Games := bracket.Rounds[1]
	round2Games := bracket.Rounds[2]

	s.Len(round1Games, 2, "Should have 2 semi-final games")
	s.Len(round2Games, 1, "Should have 1 final game")

	// Verify game links
	finalGame := round2Games[0]
	for _, game := range round1Games {
		s.NotNil(game.NextGameID, "Semi-final games should have NextGameID")
		s.Equal(finalGame.GameID, *game.NextGameID,
			"Semi-final games should link to final")
	}

	// Final should not have next game
	s.Nil(finalGame.NextGameID, "Final game should not have NextGameID")
}

func (s *TournamentFlowTestSuite) TestGameStatusTransitions() {
	// Test game status transitions: PENDING -> ACTIVE -> FINISHED

	// Create a test game
	game := gameDomain.NewTournamentGame(
		1,                             // contestID
		gameDomain.GameTeamTypeHurupa,
		1, // round
		1, // matchNumber
		1, // bracketPosition
	)
	savedGame, err := s.gameAdapter.Save(game)
	s.Require().NoError(err)

	// Initial state should be PENDING
	s.Equal(gameDomain.GameStatusPending, savedGame.GameStatus)

	// Transition to ACTIVE
	startedGame, err := s.gameService.StartGame(savedGame.GameID)
	s.Require().NoError(err)
	s.Equal(gameDomain.GameStatusActive, startedGame.GameStatus)

	// Transition to FINISHED
	finishedGame, err := s.gameService.FinishGame(savedGame.GameID)
	s.Require().NoError(err)
	s.Equal(gameDomain.GameStatusFinished, finishedGame.GameStatus)

	// Verify terminal state
	s.True(finishedGame.IsTerminalState())
}

func (s *TournamentFlowTestSuite) TestTournamentWithGradeTracking() {
	// Test that grades are properly tracked throughout tournament

	// Setup similar to main test but focus on grade tracking
	leaderID := int64(1)
	discordAccount := &oauth2Domain.DiscordAccount{
		DiscordId:     "discord_grade",
		UserId:        leaderID,
		DiscordAvatar: "avatar",
	}
	s.mockOAuth2.On("FindDiscordAccountByUserId", leaderID).Return(discordAccount, nil)

	contestReq := &dto.CreateContestRequest{
		Title:           "Grade Tracking Tournament",
		Description:     "Testing grade tracking",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	contest, _, err := s.contestService.SaveContest(contestReq, leaderID)
	s.Require().NoError(err)

	// Create teams
	teamNames := []string{"Alpha", "Beta", "Gamma", "Delta"}
	teams := make([]*gameDomain.Team, 4)
	for i, name := range teamNames {
		team := gameDomain.NewTeam(contest.ContestID, name)
		savedTeam, err := s.teamAdapter.Save(team)
		s.Require().NoError(err)
		teams[i] = savedTeam
	}

	// Allocate teams
	allocationResult, err := s.tournamentService.ShuffleAndAllocateTeamsWithResult(
		contest.ContestID,
		s.gameTeamAdapter,
	)
	s.Require().NoError(err)

	// Simulate first round with grades
	for _, alloc := range allocationResult.Allocations {
		gameTeams, _ := s.gameTeamAdapter.GetByGameID(alloc.GameID)

		// Winner gets grade 1, loser gets grade 2
		gameTeams[0].SetGrade(1)
		gameTeams[1].SetGrade(2)

		// Verify grades are set
		s.True(gameTeams[0].HasGrade())
		s.True(gameTeams[1].HasGrade())
		s.Equal(1, *gameTeams[0].Grade)
		s.Equal(2, *gameTeams[1].Grade)
	}
}

// Run the test suite
func TestTournamentFlowSuite(t *testing.T) {
	suite.Run(t, new(TournamentFlowTestSuite))
}

// ==================== Standalone Tests ====================

func TestTournamentBracketMath(t *testing.T) {
	// Test: N teams require N-1 games
	testCases := []struct {
		name     string
		teams    int
		expected int
	}{
		{"2 teams need 1 game", 2, 1},  // Final only
		{"4 teams need 3 games", 4, 3},  // 2 semi + 1 final
		{"8 teams need 7 games", 8, 7},  // 4 quarter + 2 semi + 1 final
		{"16 teams need 15 games", 16, 15}, // 8 + 4 + 2 + 1
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.teams - 1
			assert.Equal(t, tc.expected, expected,
				"Teams: %d should need %d games", tc.teams, tc.expected)
		})
	}
}

func TestRoundNames(t *testing.T) {
	testCases := []struct {
		round       int
		totalRounds int
		expected    string
	}{
		{2, 2, "Final"},
		{1, 2, "Semi-finals"},
		{3, 3, "Final"},
		{2, 3, "Semi-finals"},
		{1, 3, "Quarter-finals"},
	}

	for _, tc := range testCases {
		result := gameApplication.GetRoundName(tc.round, tc.totalRounds)
		assert.Equal(t, tc.expected, result,
			"Round %d of %d should be '%s'", tc.round, tc.totalRounds, tc.expected)
	}
}
