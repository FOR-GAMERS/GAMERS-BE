package presentation_test

import (
	"GAMERS-BE/internal/common/response"
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/domain"
	"GAMERS-BE/internal/user/presentation"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupProfileRouter() (*gin.Engine, *presentation.ProfileController, *mockProfileQueryPort, *mockProfileCommandPort) {
	gin.SetMode(gin.TestMode)

	queryPort := newMockProfileQueryPort()
	commandPort := newMockProfileCommandPort(queryPort)
	service := application.NewProfileService(queryPort, commandPort)
	controller := presentation.NewProfileController(service)

	router := gin.Default()
	controller.RegisterRoutes(router)

	return router, controller, queryPort, commandPort
}

func TestProfileController_GetProfile(t *testing.T) {
	router, _, queryPort, commandPort := setupProfileRouter()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	req, _ := http.NewRequest("GET", "/api/profiles/"+strconv.FormatInt(profile.Id, 10), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var apiResp response.ApiResponse[dto.ProfileResponse]
	err = json.Unmarshal(w.Body.Bytes(), &apiResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if apiResp.Data.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", apiResp.Data.Username)
	}
	if apiResp.Message != "Profile retrieved successfully" {
		t.Errorf("Expected message 'Profile retrieved successfully', got %s", apiResp.Message)
	}

	_ = queryPort
}

func TestProfileController_GetProfile_NotFound(t *testing.T) {
	router, _, _, _ := setupProfileRouter()

	req, _ := http.NewRequest("GET", "/api/profiles/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestProfileController_GetProfile_InvalidID(t *testing.T) {
	router, _, _, _ := setupProfileRouter()

	req, _ := http.NewRequest("GET", "/api/profiles/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestProfileController_UpdateProfile(t *testing.T) {
	router, _, queryPort, commandPort := setupProfileRouter()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	updateReqBody := map[string]string{
		"username": "updateduser",
		"tag":      "5678",
		"bio":      "Updated bio",
		"avatar":   "new_avatar.jpg",
	}
	body, _ := json.Marshal(updateReqBody)

	req, _ := http.NewRequest("PATCH", "/api/profiles/"+strconv.FormatInt(profile.Id, 10), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var apiResp response.ApiResponse[dto.ProfileResponse]
	err = json.Unmarshal(w.Body.Bytes(), &apiResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if apiResp.Data.Username != "updateduser" {
		t.Errorf("Expected username updateduser, got %s", apiResp.Data.Username)
	}
	if apiResp.Data.Tag != "5678" {
		t.Errorf("Expected tag 5678, got %s", apiResp.Data.Tag)
	}
	if apiResp.Message != "Profile updated successfully" {
		t.Errorf("Expected message 'Profile updated successfully', got %s", apiResp.Message)
	}

	_ = queryPort
}

func TestProfileController_UpdateProfile_NotFound(t *testing.T) {
	router, _, _, _ := setupProfileRouter()

	updateReqBody := map[string]string{
		"username": "updateduser",
		"tag":      "5678",
		"bio":      "Updated bio",
		"avatar":   "new_avatar.jpg",
	}
	body, _ := json.Marshal(updateReqBody)

	req, _ := http.NewRequest("PATCH", "/api/profiles/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestProfileController_UpdateProfile_InvalidID(t *testing.T) {
	router, _, _, _ := setupProfileRouter()

	updateReqBody := map[string]string{
		"username": "updateduser",
		"tag":      "5678",
	}
	body, _ := json.Marshal(updateReqBody)

	req, _ := http.NewRequest("PATCH", "/api/profiles/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestProfileController_UpdateProfile_InvalidRequestBody(t *testing.T) {
	router, _, queryPort, commandPort := setupProfileRouter()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	req, _ := http.NewRequest("PATCH", "/api/profiles/"+strconv.FormatInt(profile.Id, 10), bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	_ = queryPort
}

func TestProfileController_UpdateProfile_InvalidUsername(t *testing.T) {
	router, _, queryPort, commandPort := setupProfileRouter()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	updateReqBody := map[string]string{
		"username": "",
		"tag":      "5678",
	}
	body, _ := json.Marshal(updateReqBody)

	req, _ := http.NewRequest("PATCH", "/api/profiles/"+strconv.FormatInt(profile.Id, 10), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	_ = queryPort
}

func TestProfileController_UpdateProfile_InvalidTag(t *testing.T) {
	router, _, queryPort, commandPort := setupProfileRouter()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	updateReqBody := map[string]string{
		"username": "updateduser",
		"tag":      "",
	}
	body, _ := json.Marshal(updateReqBody)

	req, _ := http.NewRequest("PATCH", "/api/profiles/"+strconv.FormatInt(profile.Id, 10), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	_ = queryPort
}

func TestProfileController_UpdateProfile_BioTooLong(t *testing.T) {
	router, _, queryPort, commandPort := setupProfileRouter()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	longBio := strings.Repeat("a", 256)
	updateReqBody := map[string]string{
		"username": "updateduser",
		"tag":      "5678",
		"bio":      longBio,
	}
	body, _ := json.Marshal(updateReqBody)

	req, _ := http.NewRequest("PATCH", "/api/profiles/"+strconv.FormatInt(profile.Id, 10), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	_ = queryPort
}

func TestProfileController_DeleteProfile(t *testing.T) {
	router, _, queryPort, commandPort := setupProfileRouter()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	req, _ := http.NewRequest("DELETE", "/api/profiles/"+strconv.FormatInt(profile.Id, 10), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	if w.Body.Len() != 0 {
		t.Errorf("Expected empty body for 204, got %s", w.Body.String())
	}

	getReq, _ := http.NewRequest("GET", "/api/profiles/"+strconv.FormatInt(profile.Id, 10), nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 after deletion, got %d", getW.Code)
	}

	_ = queryPort
}

func TestProfileController_DeleteProfile_NotFound(t *testing.T) {
	router, _, _, _ := setupProfileRouter()

	req, _ := http.NewRequest("DELETE", "/api/profiles/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestProfileController_DeleteProfile_InvalidID(t *testing.T) {
	router, _, _, _ := setupProfileRouter()

	req, _ := http.NewRequest("DELETE", "/api/profiles/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
