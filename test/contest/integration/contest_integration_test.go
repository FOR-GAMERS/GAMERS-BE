package integration_test

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/domain"
	contestAdapter "github.com/FOR-GAMERS/GAMERS-BE/internal/contest/infra/persistence/adapter"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	oauth2Domain "github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/test/global/support"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ==================== Mock for External Dependencies ====================

// MockOAuth2DatabasePort for integration tests
type MockOAuth2DatabasePort struct {
	mock.Mock
}

func (m *MockOAuth2DatabasePort) FindDiscordAccountByDiscordId(discordId string) (*oauth2Domain.DiscordAccount, error) {
	args := m.Called(discordId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2Domain.DiscordAccount), args.Error(1)
}

func (m *MockOAuth2DatabasePort) FindDiscordAccountByUserId(userId int64) (*oauth2Domain.DiscordAccount, error) {
	args := m.Called(userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2Domain.DiscordAccount), args.Error(1)
}

func (m *MockOAuth2DatabasePort) CreateDiscordAccount(account *oauth2Domain.DiscordAccount) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockOAuth2DatabasePort) UpdateDiscordAccount(account *oauth2Domain.DiscordAccount) error {
	args := m.Called(account)
	return args.Error(0)
}

// MockEventPublisherPort for integration tests
type MockEventPublisherPort struct {
	mock.Mock
}

func (m *MockEventPublisherPort) PublishContestApplicationEvent(ctx context.Context, event *port.ContestApplicationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisherPort) PublishContestCreatedEvent(ctx context.Context, event *port.ContestCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisherPort) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEventPublisherPort) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockContestApplicationRedisPort for integration tests
type MockContestApplicationRedisPort struct {
	mock.Mock
}

func (m *MockContestApplicationRedisPort) RequestParticipate(ctx context.Context, contestId int64, sender *port.SenderSnapshot, ttl time.Duration) error {
	args := m.Called(ctx, contestId, sender, ttl)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPort) AcceptRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	args := m.Called(ctx, contestId, userId, processedBy)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPort) RejectRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	args := m.Called(ctx, contestId, userId, processedBy)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPort) CancelApplication(ctx context.Context, contestId, userId int64) error {
	args := m.Called(ctx, contestId, userId)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPort) GetApplication(ctx context.Context, contestId, userId int64) (*port.ContestApplication, error) {
	args := m.Called(ctx, contestId, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ContestApplication), args.Error(1)
}

func (m *MockContestApplicationRedisPort) GetPendingApplications(ctx context.Context, contestId int64) ([]*port.ContestApplication, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*port.ContestApplication), args.Error(1)
}

func (m *MockContestApplicationRedisPort) GetAcceptedApplications(ctx context.Context, contestId int64) ([]int64, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockContestApplicationRedisPort) GetRejectedApplications(ctx context.Context, contestId int64) ([]int64, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockContestApplicationRedisPort) HasApplied(ctx context.Context, contestId, userId int64) (bool, error) {
	args := m.Called(ctx, contestId, userId)
	return args.Bool(0), args.Error(1)
}

func (m *MockContestApplicationRedisPort) ExtendTTL(ctx context.Context, contestId int64, newTTL time.Duration) error {
	args := m.Called(ctx, contestId, newTTL)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPort) ClearApplications(ctx context.Context, contestId int64) error {
	args := m.Called(ctx, contestId)
	return args.Error(0)
}

// ==================== Integration Test Suite ====================

type ContestIntegrationTestSuite struct {
	suite.Suite
	container        *support.MySQLContainer
	db               *gorm.DB
	contestAdapter   port.ContestDatabasePort
	memberAdapter    port.ContestMemberDatabasePort
	mockOAuth2       *MockOAuth2DatabasePort
	mockEventPub     *MockEventPublisherPort
	mockRedis        *MockContestApplicationRedisPort
	contestService   *application.ContestService
}

func (s *ContestIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()
	var err error

	s.container, err = support.SetupMySQLContainer(ctx)
	s.Require().NoError(err, "Failed to setup MySQL container")

	s.db = s.container.GetDB()

	// Auto-migrate schemas
	err = s.db.AutoMigrate(&domain.Contest{}, &domain.ContestMember{})
	s.Require().NoError(err, "Failed to migrate schemas")
}

func (s *ContestIntegrationTestSuite) TearDownSuite() {
	ctx := context.Background()
	if s.container != nil {
		s.container.Teardown(ctx)
	}
}

func (s *ContestIntegrationTestSuite) SetupTest() {
	// Clean up tables before each test
	s.db.Exec("DELETE FROM contests_members")
	s.db.Exec("DELETE FROM contests")

	// Initialize adapters
	s.contestAdapter = contestAdapter.NewContestDatabaseAdapter(s.db)
	s.memberAdapter = contestAdapter.NewContestMemberDatabaseAdapter(s.db)

	// Initialize mocks
	s.mockOAuth2 = new(MockOAuth2DatabasePort)
	s.mockEventPub = new(MockEventPublisherPort)
	s.mockRedis = new(MockContestApplicationRedisPort)

	// Initialize service with real DB adapters and mocked external dependencies
	s.contestService = application.NewContestService(
		s.contestAdapter,
		s.memberAdapter,
		s.mockRedis,
		s.mockOAuth2,
		s.mockEventPub,
	)
}

func (s *ContestIntegrationTestSuite) createDiscordAccount(userId int64) *oauth2Domain.DiscordAccount {
	return &oauth2Domain.DiscordAccount{
		DiscordId:       "discord_123456",
		UserId:          userId,
		DiscordAvatar:   "avatar_hash",
		DiscordVerified: true,
	}
}

// ==================== Integration Tests ====================

func (s *ContestIntegrationTestSuite) TestCreateContest_Success() {
	// Given
	userID := int64(1)
	req := &dto.CreateContestRequest{
		Title:           "Integration Test Tournament",
		Description:     "Testing contest creation with real database",
		MaxTeamCount:    8,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		AutoStart:       false,
		TotalTeamMember: 5,
	}

	discordAccount := s.createDiscordAccount(userID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	// When
	contest, linkRequired, err := s.contestService.SaveContest(req, userID)

	// Then
	s.NoError(err)
	s.Nil(linkRequired)
	s.NotNil(contest)
	s.NotZero(contest.ContestID)
	s.Equal(req.Title, contest.Title)
	s.Equal(domain.ContestStatusPending, contest.ContestStatus)

	// Verify contest is persisted in database
	savedContest, err := s.contestAdapter.GetContestById(contest.ContestID)
	s.NoError(err)
	s.Equal(contest.Title, savedContest.Title)

	// Verify leader member is created
	member, err := s.memberAdapter.GetByContestAndUser(contest.ContestID, userID)
	s.NoError(err)
	s.True(member.IsLeader())
	s.Equal(domain.MemberTypeStaff, member.MemberType)
}

func (s *ContestIntegrationTestSuite) TestCreateContest_LeaderMemberCreated() {
	// Given
	userID := int64(1)
	req := &dto.CreateContestRequest{
		Title:           "Leader Test Tournament",
		Description:     "Testing leader member creation",
		MaxTeamCount:    4,
		TotalPoint:      50,
		ContestType:     domain.ContestTypeLeague,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(24 * time.Hour),
		TotalTeamMember: 3,
	}

	discordAccount := s.createDiscordAccount(userID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	// When
	contest, _, err := s.contestService.SaveContest(req, userID)

	// Then
	s.NoError(err)

	// Verify leader member
	member, err := s.memberAdapter.GetByContestAndUser(contest.ContestID, userID)
	s.NoError(err)
	s.Equal(userID, member.UserID)
	s.Equal(contest.ContestID, member.ContestID)
	s.Equal(domain.LeaderTypeLeader, member.LeaderType)
	s.Equal(domain.MemberTypeStaff, member.MemberType)
	s.Equal(0, member.Point)
}

func (s *ContestIntegrationTestSuite) TestStartContest_Success() {
	// Given: Create a contest first
	userID := int64(1)
	req := &dto.CreateContestRequest{
		Title:           "Start Test Tournament",
		Description:     "Testing contest start",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour), // Already past start time
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	discordAccount := s.createDiscordAccount(userID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	contest, _, err := s.contestService.SaveContest(req, userID)
	s.NoError(err)

	ctx := context.Background()

	// Mock: 3 accepted applications
	acceptedUserIDs := []int64{2, 3, 4}
	s.mockRedis.On("GetAcceptedApplications", ctx, contest.ContestID).Return(acceptedUserIDs, nil)
	s.mockRedis.On("ClearApplications", ctx, contest.ContestID).Return(nil)

	// When
	startedContest, err := s.contestService.StartContest(ctx, contest.ContestID, userID)

	// Then
	s.NoError(err)
	s.NotNil(startedContest)
	s.Equal(domain.ContestStatusActive, startedContest.ContestStatus)

	// Verify status is persisted
	savedContest, err := s.contestAdapter.GetContestById(contest.ContestID)
	s.NoError(err)
	s.Equal(domain.ContestStatusActive, savedContest.ContestStatus)

	// Verify accepted users are now members
	for _, acceptedUserID := range acceptedUserIDs {
		member, err := s.memberAdapter.GetByContestAndUser(contest.ContestID, acceptedUserID)
		s.NoError(err)
		s.Equal(domain.MemberTypeNormal, member.MemberType)
		s.Equal(domain.LeaderTypeMember, member.LeaderType)
	}
}

func (s *ContestIntegrationTestSuite) TestStopContest_Success() {
	// Given: Create and start a contest
	userID := int64(1)
	req := &dto.CreateContestRequest{
		Title:           "Stop Test Tournament",
		Description:     "Testing contest stop",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-2 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	discordAccount := s.createDiscordAccount(userID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	contest, _, err := s.contestService.SaveContest(req, userID)
	s.NoError(err)

	ctx := context.Background()

	// Start the contest first
	s.mockRedis.On("GetAcceptedApplications", ctx, contest.ContestID).Return([]int64{}, nil)
	s.mockRedis.On("ClearApplications", ctx, contest.ContestID).Return(nil)

	startedContest, err := s.contestService.StartContest(ctx, contest.ContestID, userID)
	s.NoError(err)
	s.Equal(domain.ContestStatusActive, startedContest.ContestStatus)

	// When: Stop the contest
	stoppedContest, err := s.contestService.StopContest(ctx, contest.ContestID, userID)

	// Then
	s.NoError(err)
	s.NotNil(stoppedContest)
	s.Equal(domain.ContestStatusFinished, stoppedContest.ContestStatus)

	// Verify status is persisted
	savedContest, err := s.contestAdapter.GetContestById(contest.ContestID)
	s.NoError(err)
	s.Equal(domain.ContestStatusFinished, savedContest.ContestStatus)
}

func (s *ContestIntegrationTestSuite) TestFullContestLifecycle() {
	// This test covers the complete lifecycle:
	// 1. Create contest
	// 2. Applications are accepted (mocked via Redis)
	// 3. Start contest
	// 4. Stop contest

	ctx := context.Background()
	leaderID := int64(1)

	// Step 1: Create contest
	req := &dto.CreateContestRequest{
		Title:           "Full Lifecycle Tournament",
		Description:     "Complete lifecycle test",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	discordAccount := s.createDiscordAccount(leaderID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", leaderID).Return(discordAccount, nil)

	contest, _, err := s.contestService.SaveContest(req, leaderID)
	s.NoError(err)
	s.Equal(domain.ContestStatusPending, contest.ContestStatus)

	// Step 2: Simulate accepted applications
	participantIDs := []int64{2, 3, 4, 5}
	s.mockRedis.On("GetAcceptedApplications", ctx, contest.ContestID).Return(participantIDs, nil)
	s.mockRedis.On("ClearApplications", ctx, contest.ContestID).Return(nil)

	// Step 3: Start contest
	startedContest, err := s.contestService.StartContest(ctx, contest.ContestID, leaderID)
	s.NoError(err)
	s.Equal(domain.ContestStatusActive, startedContest.ContestStatus)

	// Verify all members exist
	members, err := s.memberAdapter.GetMembersByContest(contest.ContestID)
	s.NoError(err)
	s.Len(members, 5) // 1 leader + 4 participants

	// Step 4: Stop contest
	stoppedContest, err := s.contestService.StopContest(ctx, contest.ContestID, leaderID)
	s.NoError(err)
	s.Equal(domain.ContestStatusFinished, stoppedContest.ContestStatus)

	// Verify final state
	finalContest, err := s.contestAdapter.GetContestById(contest.ContestID)
	s.NoError(err)
	s.Equal(domain.ContestStatusFinished, finalContest.ContestStatus)
	s.True(finalContest.IsTerminalState())
}

func (s *ContestIntegrationTestSuite) TestStartContest_FailNonLeader() {
	// Given
	leaderID := int64(1)
	nonLeaderID := int64(99)

	req := &dto.CreateContestRequest{
		Title:           "Non-Leader Test",
		Description:     "Testing non-leader cannot start",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	discordAccount := s.createDiscordAccount(leaderID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", leaderID).Return(discordAccount, nil)

	contest, _, err := s.contestService.SaveContest(req, leaderID)
	s.NoError(err)

	ctx := context.Background()

	// When: Non-leader tries to start
	_, err = s.contestService.StartContest(ctx, contest.ContestID, nonLeaderID)

	// Then: Should fail with access error
	s.Error(err)
	s.Equal(exception.ErrInvalidAccess, err)
}

func (s *ContestIntegrationTestSuite) TestGetContestById_Success() {
	// Given
	userID := int64(1)
	req := &dto.CreateContestRequest{
		Title:           "Get By ID Test",
		Description:     "Testing get contest by ID",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeCasual,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	discordAccount := s.createDiscordAccount(userID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	created, _, err := s.contestService.SaveContest(req, userID)
	s.NoError(err)

	// When
	contest, err := s.contestService.GetContestById(created.ContestID)

	// Then
	s.NoError(err)
	s.NotNil(contest)
	s.Equal(created.ContestID, contest.ContestID)
	s.Equal(req.Title, contest.Title)
	s.Equal(domain.ContestTypeCasual, contest.ContestType)
}

func (s *ContestIntegrationTestSuite) TestGetContestById_NotFound() {
	// When
	contest, err := s.contestService.GetContestById(99999)

	// Then
	s.Error(err)
	s.Nil(contest)
}

func (s *ContestIntegrationTestSuite) TestUpdateContest_Success() {
	// Given
	userID := int64(1)
	req := &dto.CreateContestRequest{
		Title:           "Original Title",
		Description:     "Original Description",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	discordAccount := s.createDiscordAccount(userID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	contest, _, err := s.contestService.SaveContest(req, userID)
	s.NoError(err)

	// When
	newTitle := "Updated Title"
	newMaxTeamCount := 8
	updateReq := &dto.UpdateContestRequest{
		Title:        &newTitle,
		MaxTeamCount: &newMaxTeamCount,
	}

	updated, err := s.contestService.UpdateContest(contest.ContestID, updateReq)

	// Then
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(newTitle, updated.Title)
	s.Equal(newMaxTeamCount, updated.MaxTeamCount)

	// Verify persisted
	saved, err := s.contestAdapter.GetContestById(contest.ContestID)
	s.NoError(err)
	s.Equal(newTitle, saved.Title)
	s.Equal(newMaxTeamCount, saved.MaxTeamCount)
}

func (s *ContestIntegrationTestSuite) TestDeleteContest_Success() {
	// Given
	userID := int64(1)
	req := &dto.CreateContestRequest{
		Title:           "Delete Test",
		Description:     "Will be deleted",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeCasual,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}

	discordAccount := s.createDiscordAccount(userID)
	s.mockOAuth2.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	contest, _, err := s.contestService.SaveContest(req, userID)
	s.NoError(err)

	// When
	err = s.contestService.DeleteContestById(contest.ContestID)

	// Then
	s.NoError(err)

	// Verify deleted
	_, err = s.contestAdapter.GetContestById(contest.ContestID)
	s.Error(err)
}

// Run the test suite
func TestContestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ContestIntegrationTestSuite))
}
