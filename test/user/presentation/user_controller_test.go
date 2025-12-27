package presentation_test

import (
	"GAMERS-BE/internal/common/response"
	"GAMERS-BE/internal/common/security/password"
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/presentation"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter() (*gin.Engine, *presentation.UserController) {
	gin.SetMode(gin.TestMode)

	userQueryPort := newMockUserQueryPort()
	userCommandPort := newMockUserCommandPort(userQueryPort)
	profileQueryPort := newMockProfileQueryPort()
	profileCommandPort := newMockProfileCommandPort(profileQueryPort)
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(userQueryPort, userCommandPort, profileCommandPort, hasher)
	controller := presentation.NewUserController(service)

	router := gin.Default()
	controller.RegisterRoutes(router)

	return router, controller
}

func TestUserController_CreateUser(t *testing.T) {
	router, _ := setupRouter()

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "SecurePass123!",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var apiResp response.ApiResponse[dto.UserResponse]

	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	if err != nil {
		return
	}

	if apiResp.Status != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", apiResp.Status)
	}

	if apiResp.Data.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", apiResp.Data.Email)
	}

	if apiResp.Message != "User created successfully" {
		t.Errorf("Expected message 'User created successfully', got %v", apiResp.Message)
	}
}

func TestUserController_CreateUser_InvalidEmail(t *testing.T) {
	router, _ := setupRouter()

	reqBody := map[string]string{
		"email":    "invalid-email",
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
}

func TestUserController_CreateUser_WeakPassword(t *testing.T) {
	router, _ := setupRouter()

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
}

func TestUserController_GetUser(t *testing.T) {
	router, _ := setupRouter()

	createReqBody := map[string]string{
		"email":    "test@example.com",
		"password": "SecurePass123!",
	}
	body, _ := json.Marshal(createReqBody)

	createReq, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var createResp response.ApiResponse[dto.UserResponse]
	err := json.Unmarshal(createW.Body.Bytes(), &createResp)
	if err != nil {
		return
	}
	userId := createResp.Data.Id

	getReq, _ := http.NewRequest("GET", "/api/users/"+strconv.FormatInt(userId, 10), nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", getW.Code)
	}

	var getResp response.ApiResponse[dto.UserResponse]

	err = json.Unmarshal(getW.Body.Bytes(), &getResp)
	if err != nil {
		return
	}

	if getResp.Data.Id != userId {
		t.Errorf("Expected user_id %v, got %v", userId, getResp.Data.Id)
	}
}

func TestUserController_GetUser_NotFound(t *testing.T) {
	router, _ := setupRouter()

	req, _ := http.NewRequest("GET", "/api/users/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUserController_UpdateUser(t *testing.T) {
	router, _ := setupRouter()

	createReqBody := map[string]string{
		"email":    "test@example.com",
		"password": "SecurePass123!",
	}
	body, _ := json.Marshal(createReqBody)

	createReq, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	updateReqBody := map[string]string{
		"password": "NewPassword456@",
	}
	updateBody, _ := json.Marshal(updateReqBody)

	updateReq, _ := http.NewRequest("PATCH", "/api/users/1", bytes.NewBuffer(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	if updateW.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", updateW.Code)
	}

	var apiResp response.ApiResponse[dto.UserResponse]

	err := json.Unmarshal(updateW.Body.Bytes(), &apiResp)
	if err != nil {
		return
	}

	if apiResp.Data.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", apiResp.Data.Email)
	}
}

func TestUserController_UpdateUser_NotFound(t *testing.T) {
	router, _ := setupRouter()

	updateReqBody := map[string]string{
		"password": "NewPassword456@",
	}
	body, _ := json.Marshal(updateReqBody)

	req, _ := http.NewRequest("PATCH", "/api/users/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUserController_DeleteUser(t *testing.T) {
	router, _ := setupRouter()

	createReqBody := map[string]string{
		"email":    "test@example.com",
		"password": "SecurePass123!",
	}
	body, _ := json.Marshal(createReqBody)

	createReq, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	deleteReq, _ := http.NewRequest("DELETE", "/api/users/1", nil)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)

	if deleteW.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", deleteW.Code)
	}

	if deleteW.Body.Len() != 0 {
		t.Errorf("Expected empty body for 204, got %s", deleteW.Body.String())
	}

	getReq, _ := http.NewRequest("GET", "/api/users/1", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 after deletion, got %d", getW.Code)
	}
}

func TestUserController_DeleteUser_NotFound(t *testing.T) {
	router, _ := setupRouter()

	req, _ := http.NewRequest("DELETE", "/api/users/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}
