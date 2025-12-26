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

type mockUserRepository struct {
	users  map[int64]*domain.User
	nextID int64
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:  make(map[int64]*domain.User),
		nextID: 1,
	}
}

func (m *mockUserRepository) Save(user *domain.User) error {
	if user.Id == 0 {
		user.Id = m.nextID
		m.nextID++
	}
	m.users[user.Id] = user
	return nil
}

func (m *mockUserRepository) FindById(id int64) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepository) Update(user *domain.User) error {
	if _, exists := m.users[user.Id]; !exists {
		return domain.ErrUserNotFound
	}
	m.users[user.Id] = user
	return nil
}

func (m *mockUserRepository) DeleteById(id int64) error {
	if _, exists := m.users[id]; !exists {
		return domain.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func TestUserService_CreateUser(t *testing.T) {
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

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
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

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
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

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
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "SecurePass123!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Save(user)
	if err != nil {
		return
	}

	resp, err := service.GetUserById(user.Id)
	if err != nil {
		t.Errorf("GetUserById() error = %v", err)
	}

	if resp.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, resp.Email)
	}
}

func TestUserService_GetUserById_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

	_, err := service.GetUserById(999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "SecurePass123!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Save(user)
	if err != nil {
		return
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
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

	updateReq := dto.UpdateUserRequest{
		Password: "NewPassword456@",
	}

	_, err := service.UpdateUser(999, updateReq)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "SecurePass123!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Save(user)
	if err != nil {
		return
	}

	err = service.DeleteUser(user.Id)
	if err != nil {
		t.Errorf("DeleteUser() error = %v", err)
	}

	_, err = repo.FindById(user.Id)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Error("Expected user to be deleted")
	}
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := password.NewBcryptPasswordHasher()
	service := application.NewUserService(repo, hasher)

	err := service.DeleteUser(999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
