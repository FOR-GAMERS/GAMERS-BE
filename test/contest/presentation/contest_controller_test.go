package presentation_test

import (
	"GAMERS-BE/internal/contest/application"
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	"GAMERS-BE/internal/contest/presentation"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/exception"
	oauth2Domain "GAMERS-BE/internal/oauth2/domain"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==================== Mocks ====================

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
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
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
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*port.ContestMemberWithUser), args.Get(1).(int64), args.Error(2)
}

func (m *MockContestMemberDatabasePort) GetContestsByUserId(userId int64, pagination *commonDto.PaginationRequest, sort *commonDto.SortRequest, status *domain.ContestStatus) ([]*port.ContestWithMembership, int64, error) {
	args := m.Called(userId, pagination, sort, status)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*port.ContestWithMembership), args.Get(1).(int64), args.Error(2)
}

func (m *MockContestMemberDatabasePort) UpdateMemberType(contestId, userId int64, memberType domain.MemberType) error {
	args := m.Called(contestId, userId, memberType)
	return args.Error(0)
}

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

// ==================== Test Helper Functions ====================

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupMockService() (*application.ContestService, *MockContestDatabasePort, *MockContestMemberDatabasePort) {
	mockContestDB := new(MockContestDatabasePort)
	mockMemberDB := new(MockContestMemberDatabasePort)
	mockRedis := new(MockContestApplicationRedisPort)
	mockOAuth2 := new(MockOAuth2DatabasePort)
	mockEventPub := new(MockEventPublisherPort)

	service := application.NewContestService(mockContestDB, mockMemberDB, mockRedis, mockOAuth2, mockEventPub)
	return service, mockContestDB, mockMemberDB
}

// ==================== Controller Tests ====================

func TestContestController_GetContestById_Success(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, mockContestDB, _ := setupMockService()
	helper := handler.NewControllerHelper()

	contest := &domain.Contest{
		ContestID:     1,
		Title:         "Test Tournament",
		Description:   "Test Description",
		ContestType:   domain.ContestTypeTournament,
		ContestStatus: domain.ContestStatusPending,
	}

	mockContestDB.On("GetContestById", int64(1)).Return(contest, nil)

	router.GET("/api/contests/:id", func(c *gin.Context) {
		presentation.HandleGetContestById(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/1", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(200), response["status"])
}

func TestContestController_GetContestById_NotFound(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, mockContestDB, _ := setupMockService()
	helper := handler.NewControllerHelper()

	mockContestDB.On("GetContestById", int64(999)).Return(nil, exception.ErrContestNotFound)

	router.GET("/api/contests/:id", func(c *gin.Context) {
		presentation.HandleGetContestById(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/999", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, exception.ErrContestNotFound.Status, w.Code)
}

func TestContestController_GetContestById_InvalidId(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, _, _ := setupMockService()
	helper := handler.NewControllerHelper()

	router.GET("/api/contests/:id", func(c *gin.Context) {
		presentation.HandleGetContestById(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/invalid", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestController_GetAllContests_Success(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, mockContestDB, _ := setupMockService()
	helper := handler.NewControllerHelper()

	contests := []domain.Contest{
		{ContestID: 1, Title: "Contest 1", ContestType: domain.ContestTypeTournament, ContestStatus: domain.ContestStatusPending},
		{ContestID: 2, Title: "Contest 2", ContestType: domain.ContestTypeLeague, ContestStatus: domain.ContestStatusActive},
	}

	mockContestDB.On("GetContests", 0, 10, mock.Anything, (*string)(nil)).Return(contests, int64(2), nil)

	router.GET("/api/contests", func(c *gin.Context) {
		presentation.HandleGetAllContests(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests?page=1&page_size=10", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(200), response["status"])
}

func TestContestController_GetAllContests_WithTitleSearch(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, mockContestDB, _ := setupMockService()
	helper := handler.NewControllerHelper()

	contests := []domain.Contest{
		{ContestID: 1, Title: "Valorant Tournament", ContestType: domain.ContestTypeTournament, ContestStatus: domain.ContestStatusPending},
	}

	mockContestDB.On("GetContests", 0, 10, mock.Anything, mock.MatchedBy(func(title *string) bool {
		return title != nil && *title == "Valorant"
	})).Return(contests, int64(1), nil)

	router.GET("/api/contests", func(c *gin.Context) {
		presentation.HandleGetAllContests(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests?title=Valorant", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContestController_DeleteContest_Success(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, mockContestDB, _ := setupMockService()
	helper := handler.NewControllerHelper()

	mockContestDB.On("DeleteContestById", int64(1)).Return(nil)

	router.DELETE("/api/contests/:id", func(c *gin.Context) {
		presentation.HandleDeleteContest(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/contests/1", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestContestController_UpdateContest_Success(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, mockContestDB, _ := setupMockService()
	helper := handler.NewControllerHelper()

	existingContest := &domain.Contest{
		ContestID:       1,
		Title:           "Original Title",
		Description:     "Original Description",
		ContestType:     domain.ContestTypeTournament,
		ContestStatus:   domain.ContestStatusPending,
		MaxTeamCount:    8,
		TotalPoint:      100,
		TotalTeamMember: 5,
	}

	mockContestDB.On("GetContestById", int64(1)).Return(existingContest, nil)
	mockContestDB.On("UpdateContest", mock.Anything).Return(nil)

	router.PATCH("/api/contests/:id", func(c *gin.Context) {
		presentation.HandleUpdateContest(c, service, helper)
	})

	updateReq := map[string]interface{}{
		"title": "Updated Title",
	}
	body, _ := json.Marshal(updateReq)

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/contests/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContestController_StartContest_NotAuthenticated(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, _, _ := setupMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/start", func(c *gin.Context) {
		// No userId in context
		presentation.HandleStartContest(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/1/start", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContestController_StartContest_InvalidId(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, _, _ := setupMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/start", func(c *gin.Context) {
		c.Set("userId", int64(1))
		presentation.HandleStartContest(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/invalid/start", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestController_StopContest_NotAuthenticated(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, _, _ := setupMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/stop", func(c *gin.Context) {
		presentation.HandleStopContest(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/1/stop", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContestController_StopContest_InvalidId(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, _, _ := setupMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/stop", func(c *gin.Context) {
		c.Set("userId", int64(1))
		presentation.HandleStopContest(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/invalid/stop", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestController_UpdateContest_InvalidId(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, _, _ := setupMockService()
	helper := handler.NewControllerHelper()

	router.PATCH("/api/contests/:id", func(c *gin.Context) {
		presentation.HandleUpdateContest(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/contests/invalid", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestController_DeleteContest_InvalidId(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, _, _ := setupMockService()
	helper := handler.NewControllerHelper()

	router.DELETE("/api/contests/:id", func(c *gin.Context) {
		presentation.HandleDeleteContest(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/contests/invalid", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestController_GetAllContests_Empty(t *testing.T) {
	// Given
	router := setupTestRouter()
	service, mockContestDB, _ := setupMockService()
	helper := handler.NewControllerHelper()

	mockContestDB.On("GetContests", 0, 10, mock.Anything, (*string)(nil)).Return([]domain.Contest{}, int64(0), nil)

	router.GET("/api/contests", func(c *gin.Context) {
		presentation.HandleGetAllContests(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)
}
