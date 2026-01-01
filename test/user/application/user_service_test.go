package application_test

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/security/password"
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/domain"
	"errors"
	"testing"
	"time"
)

func setupUserService() (*application.UserService, *mockUserQueryPort, *mockUserCommandPort) {
	userQueryPort := newMockUserQueryPort()
	userCommandPort := newMockUserCommandPort(userQueryPort)
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(userQueryPort, userCommandPort, hasher)
	return service, userQueryPort, userCommandPort
}

func TestUserService_CreateUser(t *testing.T) {
	service, _, _ := setupUserService()

	req := dto.CreateUserRequest{
		Email:    "test@example.com",
		Password: "SecurePass123!",
		Username: "testuser",
		Tag:      "12345",
	}

	resp, err := service.CreateUser(req)
	if err != nil {
		t.Errorf("CreateUser() exception = %v", err)
	}

	if resp.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.Email)
	}
	if resp.Id == 0 {
		t.Error("Expected user ID to be assigned")
	}
}

func TestUserService_CreateUser_InvalidEmail(t *testing.T) {
	service, _, _ := setupUserService()

	req := dto.CreateUserRequest{
		Email:    "invalid-email",
		Password: "SecurePass123!",
		Username: "testuser",
		Tag:      "12345",
	}

	_, err := service.CreateUser(req)
	if err == nil {
		t.Error("Expected exception for invalid email")
	}
}

func TestUserService_CreateUser_InvalidPassword(t *testing.T) {
	service, _, _ := setupUserService()

	req := dto.CreateUserRequest{
		Email:    "test@example.com",
		Password: "weak",
		Username: "testuser",
		Tag:      "12345",
	}

	_, err := service.CreateUser(req)
	if err == nil {
		t.Error("Expected exception for weak password")
	}
}

func TestUserService_GetUserById(t *testing.T) {
	service, queryPort, commandPort := setupUserService()

	user := &domain.User{
		Email:      "test@example.com",
		Password:   "SecurePass123!",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err := commandPort.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	resp, err := service.GetUserById(user.Id)
	if err != nil {
		t.Errorf("GetUserById() exception = %v", err)
	}

	if resp.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, resp.Email)
	}

	_ = queryPort
}

func TestUserService_GetUserById_NotFound(t *testing.T) {
	service, _, _ := setupUserService()

	_, err := service.GetUserById(999)
	if !errors.Is(err, exception.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	service, queryPort, commandPort := setupUserService()

	user := &domain.User{
		Email:      "test@example.com",
		Password:   "SecurePass123!",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
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
		t.Errorf("UpdateUser() exception = %v", err)
	}

	if resp.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, resp.Email)
	}
	if resp.ModifiedAt.Before(user.CreatedAt) {
		t.Error("ModifiedAt should be after CreatedAt")
	}

	_ = queryPort
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	service, _, _ := setupUserService()

	updateReq := dto.UpdateUserRequest{
		Password: "NewPassword456@",
	}

	_, err := service.UpdateUser(999, updateReq)
	if !errors.Is(err, exception.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	service, queryPort, commandPort := setupUserService()

	user := &domain.User{
		Email:      "test@example.com",
		Password:   "SecurePass123!",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err := commandPort.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	err = service.DeleteUser(user.Id)
	if err != nil {
		t.Errorf("DeleteUser() exception = %v", err)
	}

	_, err = queryPort.FindById(user.Id)
	if !errors.Is(err, exception.ErrUserNotFound) {
		t.Error("Expected user to be deleted")
	}
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	service, _, _ := setupUserService()

	err := service.DeleteUser(999)
	if !errors.Is(err, exception.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
