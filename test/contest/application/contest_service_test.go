package application_test

import (
	"GAMERS-BE/internal/contest/application"
	"GAMERS-BE/internal/contest/application/dto"
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	gameApplication "GAMERS-BE/internal/game/application"
	gamePort "GAMERS-BE/internal/game/application/port"
	gameDomain "GAMERS-BE/internal/game/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/exception"
	oauth2Domain "GAMERS-BE/internal/oauth2/domain"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==================== Mock Definitions ====================

// MockContestDatabasePort mocks the ContestDatabasePort interface
type MockContestDatabasePort struct {
	mock.Mock
}

func (m *MockContestDatabasePort) Save(contest *domain.Contest) (*domain.Contest, error) {
	args := m.Called(contest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contest), args.Error(1)
}

func (m *MockContestDatabasePort) GetContestById(contestId int64) (*domain.Contest, error) {
	args := m.Called(contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contest), args.Error(1)
}

func (m *MockContestDatabasePort) GetContests(offset, limit int, sortReq *commonDto.SortRequest, title *string) ([]domain.Contest, int64, error) {
	args := m.Called(offset, limit, sortReq, title)
	return args.Get(0).([]domain.Contest), args.Get(1).(int64), args.Error(2)
}

func (m *MockContestDatabasePort) DeleteContestById(contestId int64) error {
	args := m.Called(contestId)
	return args.Error(0)
}

func (m *MockContestDatabasePort) UpdateContest(contest *domain.Contest) error {
	args := m.Called(contest)
	return args.Error(0)
}

// MockContestMemberDatabasePort mocks the ContestMemberDatabasePort interface
type MockContestMemberDatabasePort struct {
	mock.Mock
}

func (m *MockContestMemberDatabasePort) Save(member *domain.ContestMember) error {
	args := m.Called(member)
	return args.Error(0)
}

func (m *MockContestMemberDatabasePort) DeleteById(contestId, userId int64) error {
	args := m.Called(contestId, userId)
	return args.Error(0)
}

func (m *MockContestMemberDatabasePort) GetByContestAndUser(contestId, userId int64) (*domain.ContestMember, error) {
	args := m.Called(contestId, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ContestMember), args.Error(1)
}

func (m *MockContestMemberDatabasePort) GetMembersByContest(contestId int64) ([]*domain.ContestMember, error) {
	args := m.Called(contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContestMember), args.Error(1)
}

func (m *MockContestMemberDatabasePort) SaveBatch(members []*domain.ContestMember) error {
	args := m.Called(members)
	return args.Error(0)
}

func (m *MockContestMemberDatabasePort) GetMembersWithUserByContest(contestId int64, pagination *commonDto.PaginationRequest, sort *commonDto.SortRequest) ([]*port.ContestMemberWithUser, int64, error) {
	args := m.Called(contestId, pagination, sort)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*port.ContestMemberWithUser), args.Get(1).(int64), args.Error(2)
}

func (m *MockContestMemberDatabasePort) GetContestsByUserId(userId int64, pagination *commonDto.PaginationRequest, sort *commonDto.SortRequest, status *domain.ContestStatus) ([]*port.ContestWithMembership, int64, error) {
	args := m.Called(userId, pagination, sort, status)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*port.ContestWithMembership), args.Get(1).(int64), args.Error(2)
}

func (m *MockContestMemberDatabasePort) UpdateMemberType(contestId, userId int64, memberType domain.MemberType) error {
	args := m.Called(contestId, userId, memberType)
	return args.Error(0)
}

// MockContestApplicationRedisPort mocks the ContestApplicationRedisPort interface
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

// MockOAuth2DatabasePort mocks the OAuth2DatabasePort interface
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

// MockEventPublisherPort mocks the EventPublisherPort interface
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

// MockTournamentGeneratorPort mocks the TournamentGeneratorPort interface
type MockTournamentGeneratorPort struct {
	mock.Mock
}

func (m *MockTournamentGeneratorPort) GenerateTournamentBracket(contestID int64, maxTeamCount int, gameTeamType gameDomain.GameTeamType) ([]*gameDomain.Game, error) {
	args := m.Called(contestID, maxTeamCount, gameTeamType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*gameDomain.Game), args.Error(1)
}

func (m *MockTournamentGeneratorPort) ShuffleAndAllocateTeamsWithResult(contestID int64, gameTeamRepo gamePort.GameTeamDatabasePort) (*gameApplication.TeamAllocationResult, error) {
	args := m.Called(contestID, gameTeamRepo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gameApplication.TeamAllocationResult), args.Error(1)
}

// ==================== Test Fixtures ====================

func createValidContestRequest() *dto.CreateContestRequest {
	return &dto.CreateContestRequest{
		Title:           "Test Tournament",
		Description:     "A test tournament for unit testing",
		MaxTeamCount:    4,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		StartedAt:       time.Now().Add(-1 * time.Hour), // Started 1 hour ago (can start)
		EndedAt:         time.Now().Add(48 * time.Hour),
		AutoStart:       false,
		TotalTeamMember: 5,
	}
}

func createDiscordAccount(userId int64) *oauth2Domain.DiscordAccount {
	return &oauth2Domain.DiscordAccount{
		DiscordId:       "123456789",
		UserId:          userId,
		DiscordAvatar:   "avatar_hash",
		DiscordVerified: true,
	}
}

func createSavedContest(req *dto.CreateContestRequest, contestID int64) *domain.Contest {
	return &domain.Contest{
		ContestID:       contestID,
		Title:           req.Title,
		Description:     req.Description,
		MaxTeamCount:    req.MaxTeamCount,
		TotalPoint:      req.TotalPoint,
		ContestType:     req.ContestType,
		ContestStatus:   domain.ContestStatusPending,
		StartedAt:       req.StartedAt,
		EndedAt:         req.EndedAt,
		AutoStart:       req.AutoStart,
		TotalTeamMember: req.TotalTeamMember,
	}
}

// ==================== Unit Tests ====================

func TestContestService_SaveContest_Success(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	userID := int64(1)
	req := createValidContestRequest()
	discordAccount := createDiscordAccount(userID)
	savedContest := createSavedContest(req, 1)

	// Mock Discord account lookup
	mockOAuth2DB.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	// Mock contest save
	mockContestDB.On("Save", mock.AnythingOfType("*domain.Contest")).Return(savedContest, nil)

	// Mock member save (leader)
	mockMemberDB.On("Save", mock.AnythingOfType("*domain.ContestMember")).Return(nil)

	// When
	result, linkRequired, err := service.SaveContest(req, userID)

	// Then
	assert.NoError(t, err)
	assert.Nil(t, linkRequired)
	assert.NotNil(t, result)
	assert.Equal(t, req.Title, result.Title)
	assert.Equal(t, domain.ContestStatusPending, result.ContestStatus)

	mockOAuth2DB.AssertExpectations(t)
	mockContestDB.AssertExpectations(t)
	mockMemberDB.AssertExpectations(t)
}

func TestContestService_SaveContest_WithTournamentBracketGeneration(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)
	mockTournamentGen := new(MockTournamentGeneratorPort)

	service := application.NewContestServiceFull(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
		nil, // discord validator
		mockTournamentGen,
		nil, // teamDBPort
		nil, // gameTeamDBPort
	)

	userID := int64(1)
	req := createValidContestRequest()
	req.ContestType = domain.ContestTypeTournament
	req.MaxTeamCount = 4

	discordAccount := createDiscordAccount(userID)
	savedContest := createSavedContest(req, 1)

	mockOAuth2DB.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)
	mockContestDB.On("Save", mock.AnythingOfType("*domain.Contest")).Return(savedContest, nil)
	mockMemberDB.On("Save", mock.AnythingOfType("*domain.ContestMember")).Return(nil)

	// Mock tournament bracket generation
	games := []*gameDomain.Game{
		{GameID: 1, ContestID: 1, Round: intPtr(1), MatchNumber: intPtr(1)},
		{GameID: 2, ContestID: 1, Round: intPtr(1), MatchNumber: intPtr(2)},
		{GameID: 3, ContestID: 1, Round: intPtr(2), MatchNumber: intPtr(1)},
	}
	mockTournamentGen.On("GenerateTournamentBracket", int64(1), 4, gameDomain.GameTeamTypeHurupa).Return(games, nil)

	// When
	result, _, err := service.SaveContest(req, userID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.ContestTypeTournament, result.ContestType)

	mockTournamentGen.AssertExpectations(t)
}

func TestContestService_SaveContest_FailWithoutDiscordLink(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	userID := int64(1)
	req := createValidContestRequest()

	// Mock: Discord account not found
	mockOAuth2DB.On("FindDiscordAccountByUserId", userID).Return(nil, exception.ErrDiscordUserCannotFound)

	// When
	result, linkRequired, err := service.SaveContest(req, userID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, exception.ErrDiscordLinkRequired, err)
	assert.Nil(t, result)
	assert.NotNil(t, linkRequired)
	assert.Contains(t, linkRequired.Message, "Discord")

	mockOAuth2DB.AssertExpectations(t)
}

func TestContestService_SaveContest_FailWithInvalidTitle(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	userID := int64(1)
	req := createValidContestRequest()
	req.Title = "" // Invalid: empty title

	discordAccount := createDiscordAccount(userID)
	mockOAuth2DB.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	// When
	result, _, err := service.SaveContest(req, userID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInvalidContestTitle, err)
	assert.Nil(t, result)
}

func TestContestService_SaveContest_FailWithInvalidDates(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	userID := int64(1)
	req := createValidContestRequest()
	req.StartedAt = time.Now().Add(48 * time.Hour)
	req.EndedAt = time.Now().Add(24 * time.Hour) // EndedAt before StartedAt

	discordAccount := createDiscordAccount(userID)
	mockOAuth2DB.On("FindDiscordAccountByUserId", userID).Return(discordAccount, nil)

	// When
	result, _, err := service.SaveContest(req, userID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInvalidContestDates, err)
	assert.Nil(t, result)
}

func TestContestService_StartContest_Success(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	ctx := context.Background()
	contestID := int64(1)
	userID := int64(1)

	contest := &domain.Contest{
		ContestID:     contestID,
		Title:         "Test Contest",
		ContestStatus: domain.ContestStatusPending,
		StartedAt:     time.Now().Add(-1 * time.Hour), // Started in past
		EndedAt:       time.Now().Add(48 * time.Hour),
	}

	leader := &domain.ContestMember{
		UserID:     userID,
		ContestID:  contestID,
		MemberType: domain.MemberTypeStaff,
		LeaderType: domain.LeaderTypeLeader,
	}

	acceptedUserIDs := []int64{2, 3, 4}

	mockContestDB.On("GetContestById", contestID).Return(contest, nil)
	mockMemberDB.On("GetByContestAndUser", contestID, userID).Return(leader, nil)
	mockRedis.On("GetAcceptedApplications", ctx, contestID).Return(acceptedUserIDs, nil)
	mockMemberDB.On("SaveBatch", mock.AnythingOfType("[]*domain.ContestMember")).Return(nil)
	mockContestDB.On("UpdateContest", mock.AnythingOfType("*domain.Contest")).Return(nil)
	mockRedis.On("ClearApplications", ctx, contestID).Return(nil)

	// When
	result, err := service.StartContest(ctx, contestID, userID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.ContestStatusActive, result.ContestStatus)

	mockContestDB.AssertExpectations(t)
	mockMemberDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestContestService_StartContest_FailNotLeader(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	ctx := context.Background()
	contestID := int64(1)
	userID := int64(2) // Not the leader

	contest := &domain.Contest{
		ContestID:     contestID,
		Title:         "Test Contest",
		ContestStatus: domain.ContestStatusPending,
		StartedAt:     time.Now().Add(-1 * time.Hour),
		EndedAt:       time.Now().Add(48 * time.Hour),
	}

	// This user is a normal member, not leader
	member := &domain.ContestMember{
		UserID:     userID,
		ContestID:  contestID,
		MemberType: domain.MemberTypeNormal,
		LeaderType: domain.LeaderTypeMember,
	}

	mockContestDB.On("GetContestById", contestID).Return(contest, nil)
	mockMemberDB.On("GetByContestAndUser", contestID, userID).Return(member, nil)

	// When
	result, err := service.StartContest(ctx, contestID, userID)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestContestService_StartContest_FailNotPending(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	ctx := context.Background()
	contestID := int64(1)
	userID := int64(1)

	// Contest is already ACTIVE
	contest := &domain.Contest{
		ContestID:     contestID,
		Title:         "Test Contest",
		ContestStatus: domain.ContestStatusActive,
		StartedAt:     time.Now().Add(-1 * time.Hour),
		EndedAt:       time.Now().Add(48 * time.Hour),
	}

	leader := &domain.ContestMember{
		UserID:     userID,
		ContestID:  contestID,
		MemberType: domain.MemberTypeStaff,
		LeaderType: domain.LeaderTypeLeader,
	}

	mockContestDB.On("GetContestById", contestID).Return(contest, nil)
	mockMemberDB.On("GetByContestAndUser", contestID, userID).Return(leader, nil)

	// When
	result, err := service.StartContest(ctx, contestID, userID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, exception.ErrContestNotPending, err)
	assert.Nil(t, result)
}

func TestContestService_StartContest_FailBeforeStartTime(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	ctx := context.Background()
	contestID := int64(1)
	userID := int64(1)

	// Contest start time is in the future
	contest := &domain.Contest{
		ContestID:     contestID,
		Title:         "Test Contest",
		ContestStatus: domain.ContestStatusPending,
		StartedAt:     time.Now().Add(24 * time.Hour), // Future
		EndedAt:       time.Now().Add(48 * time.Hour),
	}

	leader := &domain.ContestMember{
		UserID:     userID,
		ContestID:  contestID,
		MemberType: domain.MemberTypeStaff,
		LeaderType: domain.LeaderTypeLeader,
	}

	mockContestDB.On("GetContestById", contestID).Return(contest, nil)
	mockMemberDB.On("GetByContestAndUser", contestID, userID).Return(leader, nil)

	// When
	result, err := service.StartContest(ctx, contestID, userID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, exception.ErrContestCannotStart, err)
	assert.Nil(t, result)
}

func TestContestService_StopContest_Success(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	ctx := context.Background()
	contestID := int64(1)
	userID := int64(1)

	contest := &domain.Contest{
		ContestID:     contestID,
		Title:         "Test Contest",
		ContestStatus: domain.ContestStatusActive,
		StartedAt:     time.Now().Add(-24 * time.Hour),
		EndedAt:       time.Now().Add(24 * time.Hour),
	}

	leader := &domain.ContestMember{
		UserID:     userID,
		ContestID:  contestID,
		MemberType: domain.MemberTypeStaff,
		LeaderType: domain.LeaderTypeLeader,
	}

	mockContestDB.On("GetContestById", contestID).Return(contest, nil)
	mockMemberDB.On("GetByContestAndUser", contestID, userID).Return(leader, nil)
	mockContestDB.On("UpdateContest", mock.AnythingOfType("*domain.Contest")).Return(nil)

	// When
	result, err := service.StopContest(ctx, contestID, userID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.ContestStatusFinished, result.ContestStatus)

	mockContestDB.AssertExpectations(t)
	mockMemberDB.AssertExpectations(t)
}

func TestContestService_StopContest_FailNotActive(t *testing.T) {
	// Given
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2DB := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(
		mockContestDB,
		mockMemberDB,
		mockRedis,
		mockOAuth2DB,
		mockEventPub,
	)

	ctx := context.Background()
	contestID := int64(1)
	userID := int64(1)

	// Contest is still PENDING
	contest := &domain.Contest{
		ContestID:     contestID,
		Title:         "Test Contest",
		ContestStatus: domain.ContestStatusPending,
		StartedAt:     time.Now().Add(-24 * time.Hour),
		EndedAt:       time.Now().Add(24 * time.Hour),
	}

	leader := &domain.ContestMember{
		UserID:     userID,
		ContestID:  contestID,
		MemberType: domain.MemberTypeStaff,
		LeaderType: domain.LeaderTypeLeader,
	}

	mockContestDB.On("GetContestById", contestID).Return(contest, nil)
	mockMemberDB.On("GetByContestAndUser", contestID, userID).Return(leader, nil)

	// When
	result, err := service.StopContest(ctx, contestID, userID)

	// Then
	assert.Error(t, err)
	assert.Equal(t, exception.ErrContestNotActive, err)
	assert.Nil(t, result)
}

// ==================== Domain Tests ====================

func TestContest_StatusTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus domain.ContestStatus
		targetStatus  domain.ContestStatus
		shouldSucceed bool
	}{
		{"PENDING to ACTIVE", domain.ContestStatusPending, domain.ContestStatusActive, true},
		{"PENDING to CANCELLED", domain.ContestStatusPending, domain.ContestStatusCancelled, true},
		{"PENDING to FINISHED", domain.ContestStatusPending, domain.ContestStatusFinished, false},
		{"ACTIVE to FINISHED", domain.ContestStatusActive, domain.ContestStatusFinished, true},
		{"ACTIVE to CANCELLED", domain.ContestStatusActive, domain.ContestStatusCancelled, true},
		{"ACTIVE to PENDING", domain.ContestStatusActive, domain.ContestStatusPending, false},
		{"FINISHED to any", domain.ContestStatusFinished, domain.ContestStatusActive, false},
		{"CANCELLED to any", domain.ContestStatusCancelled, domain.ContestStatusActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contest := &domain.Contest{
				ContestID:     1,
				Title:         "Test",
				ContestStatus: tt.currentStatus,
			}

			err := contest.TransitionTo(tt.targetStatus)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Equal(t, tt.targetStatus, contest.ContestStatus)
			} else {
				assert.Error(t, err)
				assert.Equal(t, exception.ErrInvalidStatusTransition, err)
			}
		})
	}
}

func TestContestMember_IsLeader(t *testing.T) {
	tests := []struct {
		name       string
		leaderType domain.LeaderType
		expected   bool
	}{
		{"Leader returns true", domain.LeaderTypeLeader, true},
		{"Member returns false", domain.LeaderTypeMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := &domain.ContestMember{
				UserID:     1,
				ContestID:  1,
				MemberType: domain.MemberTypeStaff,
				LeaderType: tt.leaderType,
			}

			assert.Equal(t, tt.expected, member.IsLeader())
		})
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
