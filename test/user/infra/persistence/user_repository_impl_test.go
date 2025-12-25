package persistence

import (
	"GAMERS-BE/internal/user/domain"
	persistence2 "GAMERS-BE/internal/user/infra/persistence"
	"errors"
	"testing"
	"time"
)

func TestInMemoryUserRepository_Save(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()
	user := &domain.User{
		Email:     "test@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Save(user)
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	if user.Id == 0 {
		t.Error("Expected user ID to be assigned")
	}
}

func TestInMemoryUserRepository_Save_DuplicateEmail(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()

	user1 := &domain.User{
		Email:     "test@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user2 := &domain.User{
		Email:     "test@example.com",
		Password:  "password456",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Save(user1)
	if err != nil {
		return
	}

	err = repo.Save(user2)

	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestInMemoryUserRepository_FindById(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()

	original := &domain.User{
		Email:     "test@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Save(original)
	if err != nil {
		return
	}

	found, err := repo.FindById(original.Id)
	if err != nil {
		t.Errorf("FindById() error = %v", err)
	}

	if found.Id != original.Id {
		t.Errorf("Expected ID %d, got %d", original.Id, found.Id)
	}
	if found.Email != original.Email {
		t.Errorf("Expected email %s, got %s", original.Email, found.Email)
	}
}

func TestInMemoryUserRepository_FindById_NotFound(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()

	_, err := repo.FindById(999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_Update(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Save(user)
	if err != nil {
		return
	}

	user.Email = "updated@example.com"
	user.Password = "newpassword456"

	err = repo.Update(user)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	found, _ := repo.FindById(user.Id)
	if found.Email != "updated@example.com" {
		t.Errorf("Expected email updated@example.com, got %s", found.Email)
	}
}

func TestInMemoryUserRepository_Update_NotFound(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()

	user := &domain.User{
		Id:        999,
		Email:     "test@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Update(user)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_DeleteById(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()

	user := &domain.User{
		Email:     "test@example.com",
		Password:  "password123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Save(user)
	if err != nil {
		return
	}

	err = repo.DeleteById(user.Id)
	if err != nil {
		t.Errorf("DeleteById() error = %v", err)
	}

	_, err = repo.FindById(user.Id)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Error("Expected user to be deleted")
	}
}

func TestInMemoryUserRepository_DeleteById_NotFound(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()

	err := repo.DeleteById(999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserRepository_Concurrency(t *testing.T) {
	repo := persistence2.NewInMemoryUserRepository()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			user := &domain.User{
				Email:     "test" + string(rune(index)) + "@example.com",
				Password:  "password123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := repo.Save(user)
			if err != nil {
				return
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
