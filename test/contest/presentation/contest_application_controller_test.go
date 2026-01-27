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
	userDomain "GAMERS-BE/internal/user/domain"
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

// ==================== Application Service Mocks ====================

type MockContestDatabasePortForApp struct {
	mock.Mock
}

func (m *MockContestDatabasePortForApp) Save(contest *domain.Contest) (*domain.Contest, error) {
	args := m.Called(contest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contest), args.Error(1)
}

func (m *MockContestDatabasePortForApp) GetContestById(contestId int64) (*domain.Contest, error) {
	args := m.Called(contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contest), args.Error(1)
}

func (m *MockContestDatabasePortForApp) GetContests(offset, limit int, sortReq *commonDto.SortRequest, title *string) ([]domain.Contest, int64, error) {
	args := m.Called(offset, limit, sortReq, title)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.Contest), args.Get(1).(int64), args.Error(2)
}

func (m *MockContestDatabasePortForApp) DeleteContestById(contestId int64) error {
	args := m.Called(contestId)
	return args.Error(0)
}

func (m *MockContestDatabasePortForApp) UpdateContest(contest *domain.Contest) error {
	args := m.Called(contest)
	return args.Error(0)
}

type MockContestMemberDatabasePortForApp struct {
	mock.Mock
}

func (m *MockContestMemberDatabasePortForApp) Save(member *domain.ContestMember) error {
	args := m.Called(member)
	return args.Error(0)
}

func (m *MockContestMemberDatabasePortForApp) DeleteById(contestId, userId int64) error {
	args := m.Called(contestId, userId)
	return args.Error(0)
}

func (m *MockContestMemberDatabasePortForApp) GetByContestAndUser(contestId, userId int64) (*domain.ContestMember, error) {
	args := m.Called(contestId, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ContestMember), args.Error(1)
}

func (m *MockContestMemberDatabasePortForApp) GetMembersByContest(contestId int64) ([]*domain.ContestMember, error) {
	args := m.Called(contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContestMember), args.Error(1)
}

func (m *MockContestMemberDatabasePortForApp) SaveBatch(members []*domain.ContestMember) error {
	args := m.Called(members)
	return args.Error(0)
}

func (m *MockContestMemberDatabasePortForApp) GetMembersWithUserByContest(contestId int64, pagination *commonDto.PaginationRequest, sort *commonDto.SortRequest) ([]*port.ContestMemberWithUser, int64, error) {
	args := m.Called(contestId, pagination, sort)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*port.ContestMemberWithUser), args.Get(1).(int64), args.Error(2)
}

func (m *MockContestMemberDatabasePortForApp) GetContestsByUserId(userId int64, pagination *commonDto.PaginationRequest, sort *commonDto.SortRequest, status *domain.ContestStatus) ([]*port.ContestWithMembership, int64, error) {
	args := m.Called(userId, pagination, sort, status)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*port.ContestWithMembership), args.Get(1).(int64), args.Error(2)
}

func (m *MockContestMemberDatabasePortForApp) UpdateMemberType(contestId, userId int64, memberType domain.MemberType) error {
	args := m.Called(contestId, userId, memberType)
	return args.Error(0)
}

type MockContestApplicationRedisPortForApp struct {
	mock.Mock
}

func (m *MockContestApplicationRedisPortForApp) RequestParticipate(ctx context.Context, contestId int64, sender *port.SenderSnapshot, ttl time.Duration) error {
	args := m.Called(ctx, contestId, sender, ttl)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPortForApp) AcceptRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	args := m.Called(ctx, contestId, userId, processedBy)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPortForApp) RejectRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	args := m.Called(ctx, contestId, userId, processedBy)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPortForApp) CancelApplication(ctx context.Context, contestId, userId int64) error {
	args := m.Called(ctx, contestId, userId)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPortForApp) GetApplication(ctx context.Context, contestId, userId int64) (*port.ContestApplication, error) {
	args := m.Called(ctx, contestId, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ContestApplication), args.Error(1)
}

func (m *MockContestApplicationRedisPortForApp) GetPendingApplications(ctx context.Context, contestId int64) ([]*port.ContestApplication, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*port.ContestApplication), args.Error(1)
}

func (m *MockContestApplicationRedisPortForApp) GetAcceptedApplications(ctx context.Context, contestId int64) ([]int64, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockContestApplicationRedisPortForApp) GetRejectedApplications(ctx context.Context, contestId int64) ([]int64, error) {
	args := m.Called(ctx, contestId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockContestApplicationRedisPortForApp) HasApplied(ctx context.Context, contestId, userId int64) (bool, error) {
	args := m.Called(ctx, contestId, userId)
	return args.Bool(0), args.Error(1)
}

func (m *MockContestApplicationRedisPortForApp) ExtendTTL(ctx context.Context, contestId int64, newTTL time.Duration) error {
	args := m.Called(ctx, contestId, newTTL)
	return args.Error(0)
}

func (m *MockContestApplicationRedisPortForApp) ClearApplications(ctx context.Context, contestId int64) error {
	args := m.Called(ctx, contestId)
	return args.Error(0)
}

type MockEventPublisherPortForApp struct {
	mock.Mock
}

func (m *MockEventPublisherPortForApp) PublishContestApplicationEvent(ctx context.Context, event *port.ContestApplicationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisherPortForApp) PublishContestCreatedEvent(ctx context.Context, event *port.ContestCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisherPortForApp) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEventPublisherPortForApp) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockUserQueryPortForApp struct {
	mock.Mock
}

func (m *MockUserQueryPortForApp) FindById(id int64) (*userDomain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserQueryPortForApp) FindByEmail(email string) (*userDomain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserQueryPortForApp) FindByUsername(username string) (*userDomain.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserQueryPortForApp) FindAll(pagination *commonDto.PaginationRequest) ([]*userDomain.User, int64, error) {
	args := m.Called(pagination)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*userDomain.User), args.Get(1).(int64), args.Error(2)
}

type MockOAuth2DatabasePortForApp struct {
	mock.Mock
}

func (m *MockOAuth2DatabasePortForApp) FindDiscordAccountByDiscordId(discordId string) (*oauth2Domain.DiscordAccount, error) {
	args := m.Called(discordId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2Domain.DiscordAccount), args.Error(1)
}

func (m *MockOAuth2DatabasePortForApp) FindDiscordAccountByUserId(userId int64) (*oauth2Domain.DiscordAccount, error) {
	args := m.Called(userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2Domain.DiscordAccount), args.Error(1)
}

func (m *MockOAuth2DatabasePortForApp) CreateDiscordAccount(account *oauth2Domain.DiscordAccount) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockOAuth2DatabasePortForApp) UpdateDiscordAccount(account *oauth2Domain.DiscordAccount) error {
	args := m.Called(account)
	return args.Error(0)
}

// ==================== Test Setup ====================

func setupAppTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupAppMockService() (*application.ContestApplicationService, *MockContestDatabasePortForApp, *MockContestMemberDatabasePortForApp, *MockContestApplicationRedisPortForApp) {
	mockContestDB := new(MockContestDatabasePortForApp)
	mockMemberDB := new(MockContestMemberDatabasePortForApp)
	mockRedis := new(MockContestApplicationRedisPortForApp)
	mockEventPub := new(MockEventPublisherPortForApp)
	mockOAuth2 := new(MockOAuth2DatabasePortForApp)
	mockUserQuery := new(MockUserQueryPortForApp)

	service := application.NewContestApplicationService(
		mockRedis,      // applicationRepo
		mockContestDB,  // contestRepo
		mockMemberDB,   // memberRepo
		mockEventPub,   // eventPublisher
		mockOAuth2,     // oauth2Repository
		mockUserQuery,  // userQueryRepo
	)
	return service, mockContestDB, mockMemberDB, mockRedis
}

// ==================== Controller Tests ====================

func TestContestApplicationController_GetPendingApplications_Success(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, mockContestDB, _, mockRedis := setupAppMockService()
	helper := handler.NewControllerHelper()

	contest := &domain.Contest{
		ContestID:     1,
		Title:         "Test Contest",
		ContestStatus: domain.ContestStatusPending,
	}

	applications := []*port.ContestApplication{
		{
			ContestID: 1,
			UserID:    2,
			Sender: &port.SenderSnapshot{
				UserID:   2,
				Username: "testuser",
				Tag:      "1234",
			},
			Status:      port.ApplicationStatusPending,
			RequestedAt: time.Now(),
		},
	}

	mockContestDB.On("GetContestById", int64(1)).Return(contest, nil)
	mockRedis.On("GetPendingApplications", mock.Anything, int64(1)).Return(applications, nil)

	router.GET("/api/contests/:id/applications", func(c *gin.Context) {
		presentation.HandleGetPendingApplications(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/1/applications", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(200), response["status"])
}

func TestContestApplicationController_GetPendingApplications_InvalidId(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.GET("/api/contests/:id/applications", func(c *gin.Context) {
		presentation.HandleGetPendingApplications(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/invalid/applications", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestApplicationController_GetMyApplication_NotAuthenticated(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.GET("/api/contests/:id/applications/me", func(c *gin.Context) {
		// No userId in context
		presentation.HandleGetMyApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/1/applications/me", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContestApplicationController_GetMyApplication_InvalidId(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.GET("/api/contests/:id/applications/me", func(c *gin.Context) {
		c.Set("userId", int64(1))
		presentation.HandleGetMyApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/invalid/applications/me", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestApplicationController_AcceptApplication_NotAuthenticated(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/applications/:userId/accept", func(c *gin.Context) {
		presentation.HandleAcceptApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/1/applications/2/accept", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContestApplicationController_AcceptApplication_InvalidContestId(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/applications/:userId/accept", func(c *gin.Context) {
		c.Set("userId", int64(1))
		presentation.HandleAcceptApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/invalid/applications/2/accept", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestApplicationController_AcceptApplication_InvalidUserId(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/applications/:userId/accept", func(c *gin.Context) {
		c.Set("userId", int64(1))
		presentation.HandleAcceptApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/1/applications/invalid/accept", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestApplicationController_RejectApplication_NotAuthenticated(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/applications/:userId/reject", func(c *gin.Context) {
		presentation.HandleRejectApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/1/applications/2/reject", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContestApplicationController_RejectApplication_InvalidContestId(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.POST("/api/contests/:id/applications/:userId/reject", func(c *gin.Context) {
		c.Set("userId", int64(1))
		presentation.HandleRejectApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/contests/invalid/applications/2/reject", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestApplicationController_CancelApplication_NotAuthenticated(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.DELETE("/api/contests/:id/applications/cancel", func(c *gin.Context) {
		presentation.HandleCancelApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/contests/1/applications/cancel", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContestApplicationController_CancelApplication_InvalidId(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.DELETE("/api/contests/:id/applications/cancel", func(c *gin.Context) {
		c.Set("userId", int64(1))
		presentation.HandleCancelApplication(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/contests/invalid/applications/cancel", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestApplicationController_GetMyContestStatus_NotAuthenticated(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.GET("/api/contests/:id/status/me", func(c *gin.Context) {
		presentation.HandleGetMyContestStatus(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/1/status/me", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestContestApplicationController_GetMyContestStatus_InvalidId(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.GET("/api/contests/:id/status/me", func(c *gin.Context) {
		c.Set("userId", int64(1))
		presentation.HandleGetMyContestStatus(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/invalid/status/me", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestApplicationController_GetContestMembers_Success(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, mockContestDB, mockMemberDB, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	contest := &domain.Contest{
		ContestID:     1,
		Title:         "Test Contest",
		ContestStatus: domain.ContestStatusActive,
	}

	members := []*port.ContestMemberWithUser{
		{
			UserID:     1,
			ContestID:  1,
			MemberType: domain.MemberTypeStaff,
			LeaderType: domain.LeaderTypeLeader,
			Point:      100,
			Username:   "leader",
			Tag:        "1111",
		},
		{
			UserID:     2,
			ContestID:  1,
			MemberType: domain.MemberTypeNormal,
			LeaderType: domain.LeaderTypeMember,
			Point:      50,
			Username:   "member",
			Tag:        "2222",
		},
	}

	mockContestDB.On("GetContestById", int64(1)).Return(contest, nil)
	mockMemberDB.On("GetMembersWithUserByContest", int64(1), mock.Anything, mock.Anything).Return(members, int64(2), nil)

	router.GET("/api/contests/:id/members", func(c *gin.Context) {
		presentation.HandleGetContestMembers(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/1/members", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContestApplicationController_GetContestMembers_InvalidId(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, _, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	router.GET("/api/contests/:id/members", func(c *gin.Context) {
		presentation.HandleGetContestMembers(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/invalid/members", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContestApplicationController_GetContestMembers_NotFound(t *testing.T) {
	// Given
	router := setupAppTestRouter()
	service, mockContestDB, _, _ := setupAppMockService()
	helper := handler.NewControllerHelper()

	mockContestDB.On("GetContestById", int64(999)).Return(nil, exception.ErrContestNotFound)

	router.GET("/api/contests/:id/members", func(c *gin.Context) {
		presentation.HandleGetContestMembers(c, service, helper)
	})

	// When
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/contests/999/members", nil)
	router.ServeHTTP(w, req)

	// Then
	assert.Equal(t, exception.ErrContestNotFound.Status, w.Code)
}
