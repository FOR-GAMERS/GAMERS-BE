package application_test

import (
	"GAMERS-BE/internal/common/security/password"
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/domain"
	"errors"
	"testing"
	"time"
)

func setupUserService() (*application.UserService, *mockUserQueryPort, *mockUserCommandPort, *mockProfileCommandPort) {
	userQueryPort := newMockUserQueryPort()
	userCommandPort := newMockUserCommandPort(userQueryPort)
	profileQueryPort := newMockProfileQueryPort()
	profileCommandPort := newMockProfileCommandPort(profileQueryPort)
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(userQueryPort, userCommandPort, profileCommandPort, hasher)
	return service, userQueryPort, userCommandPort, profileCommandPort
}

func TestUserService_CreateUser(t *testing.T) {
	service, _, _, _ := setupUserService()

	req := dto.CreateUserRequest{
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}

	resp, err := service.CreateUser(req)
	if err != nil {
		t.Errorf("CreateUser() error = %v", err)
	}

	if resp.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.Email)
	}
	if resp.Id == 0 {
		t.Error("Expected user ID to be assigned")
	}
}

func TestUserService_CreateUser_InvalidEmail(t *testing.T) {
	service, _, _, _ := setupUserService()

	req := dto.CreateUserRequest{
		Email:    "invalid-email",
		Password: "SecurePass123!",
	}

	_, err := service.CreateUser(req)
	if err == nil {
		t.Error("Expected error for invalid email")
	}
}

func TestUserService_CreateUser_InvalidPassword(t *testing.T) {
	service, _, _, _ := setupUserService()

	req := dto.CreateUserRequest{
		Email:    "test@example.com",
		Password: "weak",
	}

	_, err := service.CreateUser(req)
	if err == nil {
		t.Error("Expected error for weak password")
	}
}

func TestUserService_GetUserById(t *testing.T) {
	service, queryPort, commandPort, _ := setupUserService()

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "SecurePass123!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	resp, err := service.GetUserById(user.Id)
	if err != nil {
		t.Errorf("GetUserById() error = %v", err)
	}

	if resp.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, resp.Email)
	}

	_ = queryPort
}

func TestUserService_GetUserById_NotFound(t *testing.T) {
	service, _, _, _ := setupUserService()

	_, err := service.GetUserById(999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	service, queryPort, commandPort, _ := setupUserService()

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "SecurePass123!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	updateReq := dto.UpdateUserRequest{
		Password: "NewPassword456@",
	}

	resp, err := service.UpdateUser(user.Id, updateReq)
	if err != nil {
		t.Errorf("UpdateUser() error = %v", err)
	}

	if resp.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, resp.Email)
	}
	if resp.UpdatedAt.Before(user.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}

	_ = queryPort
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	service, _, _, _ := setupUserService()

	updateReq := dto.UpdateUserRequest{
		Password: "NewPassword456@",
	}

	_, err := service.UpdateUser(999, updateReq)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	service, queryPort, commandPort, _ := setupUserService()

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "SecurePass123!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	err = service.DeleteUser(user.Id)
	if err != nil {
		t.Errorf("DeleteUser() error = %v", err)
	}

	_, err = queryPort.FindById(user.Id)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Error("Expected user to be deleted")
	}
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	service, _, _, _ := setupUserService()

	err := service.DeleteUser(999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
