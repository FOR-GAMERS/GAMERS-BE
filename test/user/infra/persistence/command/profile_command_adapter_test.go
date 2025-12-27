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

func TestProfileCommandAdapter_Save(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMysqlProfileCommandAdapter(db)

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Save(profile)
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	if profile.Id == 0 {
		t.Error("Expected profile ID to be assigned")
	}

	var savedProfile domain.Profile
	db.First(&savedProfile, profile.Id)

	if savedProfile.Username != "testuser" {
		t.Errorf("Expected username %s, got %s", "testuser", savedProfile.Username)
	}
}

func TestProfileCommandAdapter_Update(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMysqlProfileCommandAdapter(db)

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	profile.Username = "updateduser"
	profile.Tag = "5678"
	profile.Bio = "Updated bio"
	profile.Avatar = "new_avatar.jpg"

	err = adapter.Update(profile)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	var updatedProfile domain.Profile
	db.First(&updatedProfile, profile.Id)

	if updatedProfile.Username != "updateduser" {
		t.Errorf("Expected username %s, got %s", "updateduser", updatedProfile.Username)
	}
	if updatedProfile.Tag != "5678" {
		t.Errorf("Expected tag %s, got %s", "5678", updatedProfile.Tag)
	}
	if updatedProfile.Bio != "Updated bio" {
		t.Errorf("Expected bio %s, got %s", "Updated bio", updatedProfile.Bio)
	}
	if updatedProfile.Avatar != "new_avatar.jpg" {
		t.Errorf("Expected avatar %s, got %s", "new_avatar.jpg", updatedProfile.Avatar)
	}
}

func TestProfileCommandAdapter_Update_NotFound(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMysqlProfileCommandAdapter(db)

	profile := &domain.Profile{
		Id:        999,
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Update(profile)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileCommandAdapter_DeleteById(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMysqlProfileCommandAdapter(db)

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = adapter.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	err = adapter.DeleteById(profile.Id)
	if err != nil {
		t.Errorf("DeleteById() error = %v", err)
	}

	var count int64
	db.Model(&domain.Profile{}).Where("profile_id = ?", profile.Id).Count(&count)

	if count != 0 {
		t.Error("Expected profile to be deleted")
	}
}

func TestProfileCommandAdapter_DeleteById_NotFound(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	adapter := command.NewMysqlProfileCommandAdapter(db)

	err = adapter.DeleteById(999)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}
