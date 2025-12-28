package query_test

import (
	"GAMERS-BE/internal/user/domain"
	"GAMERS-BE/internal/user/infra/persistence/query"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserQueryTestDB() (*gorm.DB, error) {
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

func TestUserQueryAdapter_FindById(t *testing.T) {
	db, err := setupUserQueryTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlUserRepository(db)

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "hashedPassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create test user: %v", result.Error)
	}

	foundUser, err := adapter.FindById(user.Id)
	if err != nil {
		t.Errorf("FindById() error = %v", err)
	}

	if foundUser == nil {
		t.Fatal("Expected user to be found")
	}

	if foundUser.Id != user.Id {
		t.Errorf("Expected user ID %d, got %d", user.Id, foundUser.Id)
	}

	if foundUser.Email != "test@example.com" {
		t.Errorf("Expected email %s, got %s", "test@example.com", foundUser.Email)
	}
}

func TestUserQueryAdapter_FindById_NotFound(t *testing.T) {
	db, err := setupUserQueryTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlUserRepository(db)

	foundUser, err := adapter.FindById(999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if foundUser != nil {
		t.Error("Expected user to be nil")
	}
}

func TestUserQueryAdapter_FindByEmail(t *testing.T) {
	db, err := setupUserQueryTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlUserRepository(db)

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "hashedPassword123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create test user: %v", result.Error)
	}

	foundUser, err := adapter.FindByEmail("test@example.com")
	if err != nil {
		t.Errorf("FindByEmail() error = %v", err)
	}

	if foundUser == nil {
		t.Fatal("Expected user to be found")
	}

	if foundUser.Email != "test@example.com" {
		t.Errorf("Expected email %s, got %s", "test@example.com", foundUser.Email)
	}

	if foundUser.Id != user.Id {
		t.Errorf("Expected user ID %d, got %d", user.Id, foundUser.Id)
	}
}

func TestUserQueryAdapter_FindByEmail_NotFound(t *testing.T) {
	db, err := setupUserQueryTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlUserRepository(db)

	foundUser, err := adapter.FindByEmail("nonexistent@example.com")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if foundUser != nil {
		t.Error("Expected user to be nil")
	}
}
