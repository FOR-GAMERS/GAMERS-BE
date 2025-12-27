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

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&domain.Profile{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestProfileQueryAdapter_FindById(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlProfileQueryAdapter(db)

	original := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db.Create(original)

	found, err := adapter.FindById(original.Id)
	if err != nil {
		t.Errorf("FindById() error = %v", err)
	}

	if found.Id != original.Id {
		t.Errorf("Expected ID %d, got %d", original.Id, found.Id)
	}
	if found.Username != original.Username {
		t.Errorf("Expected username %s, got %s", original.Username, found.Username)
	}
	if found.Tag != original.Tag {
		t.Errorf("Expected tag %s, got %s", original.Tag, found.Tag)
	}
	if found.Bio != original.Bio {
		t.Errorf("Expected bio %s, got %s", original.Bio, found.Bio)
	}
}

func TestProfileQueryAdapter_FindById_NotFound(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlProfileQueryAdapter(db)

	_, err = adapter.FindById(999)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileQueryAdapter_FindByUserId(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlProfileQueryAdapter(db)

	original := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db.Create(original)

	found, err := adapter.FindByUserId(1)
	if err != nil {
		t.Errorf("FindByUserId() error = %v", err)
	}

	if found.UserId != original.UserId {
		t.Errorf("Expected UserId %d, got %d", original.UserId, found.UserId)
	}
	if found.Username != original.Username {
		t.Errorf("Expected username %s, got %s", original.Username, found.Username)
	}
}

func TestProfileQueryAdapter_FindByUserId_NotFound(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlProfileQueryAdapter(db)

	_, err = adapter.FindByUserId(999)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileQueryAdapter_FindByUserId_MultipleProfiles(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := query.NewMysqlProfileQueryAdapter(db)

	profile1 := &domain.Profile{
		UserId:    1,
		Username:  "user1",
		Tag:       "1111",
		Bio:       "Bio 1",
		Avatar:    "avatar1.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	profile2 := &domain.Profile{
		UserId:    2,
		Username:  "user2",
		Tag:       "2222",
		Bio:       "Bio 2",
		Avatar:    "avatar2.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db.Create(profile1)
	db.Create(profile2)

	found1, err := adapter.FindByUserId(1)
	if err != nil {
		t.Errorf("FindByUserId(1) error = %v", err)
	}
	if found1.Username != "user1" {
		t.Errorf("Expected username user1, got %s", found1.Username)
	}

	found2, err := adapter.FindByUserId(2)
	if err != nil {
		t.Errorf("FindByUserId(2) error = %v", err)
	}
	if found2.Username != "user2" {
		t.Errorf("Expected username user2, got %s", found2.Username)
	}
}
