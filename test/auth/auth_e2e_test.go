package auth

import (
	"GAMERS-BE/internal/auth/application"
	"GAMERS-BE/internal/auth/infra/jwt"
	authCommand "GAMERS-BE/internal/auth/infra/persistence/command"
	authQuery "GAMERS-BE/internal/auth/infra/persistence/query"
	"GAMERS-BE/internal/auth/presentation"
	"GAMERS-BE/internal/global/security/password"
	"GAMERS-BE/internal/user/domain"
	userQuery "GAMERS-BE/internal/user/infra/persistence/query"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testServer struct {
	router      *gin.Engine
	db          *gorm.DB
	redisClient *redis.Client
	ctx         context.Context
}

func setupE2EServer(t *testing.T) *testServer {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	// Setup SQLite in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&domain.User{})
	assert.NoError(t, err)

	// Setup Redis client for testing
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // Use a separate DB for testing
	})

	// Ping Redis to ensure it's available
	err = redisClient.Ping(ctx).Err()
	assert.NoError(t, err, "Redis must be running for E2E tests")

	// Clear all keys in test DB
	err = redisClient.FlushDB(ctx).Err()
	assert.NoError(t, err)

	// Create adapters
	authUserQueryAdapter := userQuery.NewAuthUserQueryAdapter(db)
	refreshTokenQueryAdapter := authQuery.NewRefreshTokenRedisQueryAdapter(redisClient)
	refreshTokenCommandAdapter := authCommand.NewRefreshTokenRedisCommandAdapter(redisClient)
	passwordHasher := password.NewBcryptPasswordHasher()

	// Setup JWT token provider
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-for-testing-only")
	os.Setenv("JWT_ACCESS_DURATION", "15m")
	os.Setenv("JWT_REFRESH_DURATION", "168h")

	jwtConfig := jwt.NewConfigFromEnv()
	tokenManager := jwt.NewTokenManager(jwtConfig)
	tokenProvider := jwt.NewTokenProvider(tokenManager)

	// Create service
	authService := application.NewAuthService(
		ctx,
		authUserQueryAdapter,
		refreshTokenCommandAdapter,
		refreshTokenQueryAdapter,
		tokenProvider,
		passwordHasher,
	)

	// Create controller
	authController := presentation.NewAuthController(authService)

	// Setup router
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authController.RegisterRoutes(router, nil)

	return &testServer{
		router:      router,
		db:          db,
		redisClient: redisClient,
		ctx:         ctx,
	}
}

func (s *testServer) cleanup() {
	s.redisClient.FlushDB(s.ctx).Err()
	s.redisClient.Close()
}

func createTestUser(t *testing.T, db *gorm.DB) *domain.User {
	hasher := password.NewBcryptPasswordHasher()
	user, err := domain.NewUser(
		"test@example.com",
		"TestPass123!",
		"testuser",
		"12345",
		"Test Bio",
		"",
	)
	assert.NoError(t, err)

	err = user.EncryptPassword(hasher)
	assert.NoError(t, err)

	result := db.Create(user)
	assert.NoError(t, result.Error)

	return user
}

func TestE2E_AuthHealthCheck(t *testing.T) {
	server := setupE2EServer(t)
	defer server.cleanup()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestE2E_AuthLoginLifecycle(t *testing.T) {
	server := setupE2EServer(t)
	defer server.cleanup()

	// Create a test user first
	_ = createTestUser(t, server.db)

	var accessToken, refreshToken string

	t.Run("Login Successfully", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "TestPass123!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Data.AccessToken)
		assert.NotEmpty(t, response.Data.RefreshToken)

		accessToken = response.Data.AccessToken
		refreshToken = response.Data.RefreshToken
	})

	t.Run("Logout Successfully", func(t *testing.T) {
		reqBody := map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Cannot Use Refresh Token After Logout", func(t *testing.T) {
		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestE2E_AuthRefreshToken(t *testing.T) {
	server := setupE2EServer(t)
	defer server.cleanup()

	// Create a test user first
	_ = createTestUser(t, server.db)

	var refreshToken string

	t.Run("Login to Get Tokens", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "TestPass123!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data struct {
				RefreshToken string `json:"refresh_token"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		refreshToken = response.Data.RefreshToken
	})

	t.Run("Refresh Token Successfully", func(t *testing.T) {
		// Wait a bit to ensure new token will have different timestamp
		time.Sleep(100 * time.Millisecond)

		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data struct {
				AccessToken     string `json:"access_token"`
				RefreshToken    string `json:"refresh_token"`
				AccessTokenExp  int64  `json:"access_token_exp"`
				RefreshTokenExp int64  `json:"refresh_token_exp"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Data.AccessToken)
		assert.NotEmpty(t, response.Data.RefreshToken)
		assert.NotEqual(t, refreshToken, response.Data.RefreshToken, "New refresh token should be different")
	})

	t.Run("Old Refresh Token Should Be Invalid After Refresh", func(t *testing.T) {
		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestE2E_AuthInvalidCredentials(t *testing.T) {
	server := setupE2EServer(t)
	defer server.cleanup()

	// Create a test user first
	_ = createTestUser(t, server.db)

	t.Run("Invalid Password", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "WrongPassword123!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Non-existent User", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "TestPass123!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Empty Request Body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestE2E_AuthInvalidRefreshToken(t *testing.T) {
	server := setupE2EServer(t)
	defer server.cleanup()

	t.Run("Invalid Refresh Token Format", func(t *testing.T) {
		reqBody := map[string]string{
			"refresh_token": "invalid-token-format",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Empty Refresh Token", func(t *testing.T) {
		reqBody := map[string]string{
			"refresh_token": "",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestE2E_AuthMultipleSessionsForSameUser(t *testing.T) {
	server := setupE2EServer(t)
	defer server.cleanup()

	// Create a test user first
	_ = createTestUser(t, server.db)

	t.Run("Multiple Login Sessions", func(t *testing.T) {
		var tokens []string

		// Login 3 times
		for i := 0; i < 3; i++ {
			reqBody := map[string]string{
				"email":    "test@example.com",
				"password": "TestPass123!",
			}
			body, _ := json.Marshal(reqBody)

			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response struct {
				Data struct {
					RefreshToken string `json:"refresh_token"`
				} `json:"data"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			tokens = append(tokens, response.Data.RefreshToken)
		}

		// All tokens should be different
		assert.NotEqual(t, tokens[0], tokens[1])
		assert.NotEqual(t, tokens[1], tokens[2])
		assert.NotEqual(t, tokens[0], tokens[2])

		// All tokens should be valid
		for i, token := range tokens {
			reqBody := map[string]string{
				"refresh_token": token,
			}
			body, _ := json.Marshal(reqBody)

			req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Token %d should be valid", i)
		}
	})
}
