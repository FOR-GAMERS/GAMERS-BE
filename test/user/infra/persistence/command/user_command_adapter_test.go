package command_test

import (
	"GAMERS-BE/internal/user/domain"
	"GAMERS-BE/internal/user/infra/persistence/command"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&domain.User{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestUserCommandAdapter_Save(t *testing.T) {
	db, err := setupUserTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMySQLUserRepository(db)

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "hashedPassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Save(user)
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	if user.Id == 0 {
		t.Error("Expected user ID to be assigned")
	}

	var savedUser domain.User
	db.First(&savedUser, user.Id)

	if savedUser.Email != "test@example.com" {
		t.Errorf("Expected email %s, got %s", "test@example.com", savedUser.Email)
	}
}

func TestUserCommandAdapter_Save_DuplicateEmail(t *testing.T) {
	db, err := setupUserTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMySQLUserRepository(db)

	user1 := &domain.User{
		Email:     "test@example.com",
		Password:  "hashedPassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Save(user1)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	user2 := &domain.User{
		Email:     "test@example.com",
		Password:  "hashedPassword456",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Save(user2)
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestUserCommandAdapter_Update(t *testing.T) {
	db, err := setupUserTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMySQLUserRepository(db)

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "hashedPassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	user.Password = "newHashedPassword456"

	err = adapter.Update(user)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	var updatedUser domain.User
	db.First(&updatedUser, user.Id)

	if updatedUser.Password != "newHashedPassword456" {
		t.Errorf("Expected password %s, got %s", "newHashedPassword456", updatedUser.Password)
	}
}

func TestUserCommandAdapter_Update_NotFound(t *testing.T) {
	db, err := setupUserTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMySQLUserRepository(db)

	user := &domain.User{
		Id:        999,
		Email:     "test@example.com",
		Password:  "hashedPassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Update(user)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUserCommandAdapter_DeleteById(t *testing.T) {
	db, err := setupUserTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMySQLUserRepository(db)

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "hashedPassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Save(user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	err = adapter.DeleteById(user.Id)
	if err != nil {
		t.Errorf("DeleteById() error = %v", err)
	}

	var count int64
	db.Model(&domain.User{}).Where("id = ?", user.Id).Count(&count)

	if count != 0 {
		t.Error("Expected user to be deleted")
	}
}

func TestUserCommandAdapter_DeleteById_NotFound(t *testing.T) {
	db, err := setupUserTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMySQLUserRepository(db)

	err = adapter.DeleteById(999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}
