package user

import (
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/infra/persistence"
	"GAMERS-BE/internal/user/presentation"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupE2EServer() *gin.Engine {
	gin.SetMode(gin.TestMode)

	userRepository := persistence.NewInMemoryUserRepository()
	userService := application.NewUserService(userRepository)
	userController := presentation.NewUserController(userService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userController.RegisterRoutes(router)

	return router
}

func TestE2E_UserLifecycle(t *testing.T) {
	router := setupE2EServer()

	t.Run("Health Check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Create User", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "user@example.com",
			"password": "SecurePass123!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", w.Code)
		}

		var response struct {
			Data struct {
				Email string `json:"email"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}

		if response.Data.Email != "user@example.com" {
			t.Errorf("Expected email user@example.com, got %v", response.Data.Email)
		}
	})

	t.Run("Get Created User", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/users/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response struct {
			Data struct {
				Email string `json:"email"`
			} `json:"data"`
		}

		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}

		if response.Data.Email != "user@example.com" {
			t.Errorf("Expected email user@example.com, got %v", response.Data.Email)
		}
	})

	t.Run("Update User", func(t *testing.T) {
		reqBody := map[string]string{
			"password": "NewPassword456@",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PATCH", "/api/users/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response struct {
			Data struct {
				Email string `json:"email"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}

		if response.Data.Email != "user@example.com" {
			t.Errorf("Expected email user@example.com (unchanged), got %v", response.Data.Email)
		}
	})

	t.Run("Verify Update", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/users/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var response struct {
			Data struct {
				Email string `json:"email"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			return
		}

		if response.Data.Email != "user@example.com" {
			t.Errorf("Expected email user@example.com (unchanged), got %v", response.Data.Email)
		}
	})

	t.Run("Delete User", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/users/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}
	})

	t.Run("Verify Deletion", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/users/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestE2E_DuplicateEmailScenario(t *testing.T) {
	router := setupE2EServer()

	t.Run("Create First User", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "duplicate@example.com",
			"password": "SecurePass123!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", w.Code)
		}
	})

	t.Run("Attempt Duplicate Email", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "duplicate@example.com",
			"password": "DifferentPass456!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", w.Code)
		}
	})
}

func TestE2E_InvalidInputScenarios(t *testing.T) {
	router := setupE2EServer()

	t.Run("Invalid Email Format", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "not-an-email",
			"password": "SecurePass123!",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Weak Password", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "test@example.com",
			"password": "weak",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Empty Request Body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestE2E_MultipleUsersScenario(t *testing.T) {
	router := setupE2EServer()

	users := []map[string]string{
		{"email": "user1@example.com", "password": "Password1!"},
		{"email": "user2@example.com", "password": "Password2!"},
		{"email": "user3@example.com", "password": "Password3!"},
	}

	t.Run("Create Multiple Users", func(t *testing.T) {
		for _, user := range users {
			body, _ := json.Marshal(user)

			req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("Failed to create user %s: status %d", user["email"], w.Code)
			}
		}
	})

	t.Run("Verify All Users Created", func(t *testing.T) {
		for i := 1; i <= len(users); i++ {
			req, _ := http.NewRequest("GET", "/api/users/"+string(rune(i+'0')), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Failed to get user %d: status %d", i, w.Code)
			}
		}
	})
}
