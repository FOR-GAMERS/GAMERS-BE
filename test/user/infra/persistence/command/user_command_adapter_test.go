package command_test

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/user/domain"
	"GAMERS-BE/internal/user/infra/persistence/command"
	"GAMERS-BE/test/global/support"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCommandAdapter_Save(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := command.NewMySQLUserRepository(db)

	user := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword123",
		Username:   "testuser",
		Tag:        "12345",
		Bio:        "Test bio",
		Avatar:     "https://example.com/avatar.jpg",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = adapter.Save(user)
	require.NoError(t, err)
	assert.NotZero(t, user.Id)

	var savedUser domain.User
	db.First(&savedUser, user.Id)

	assert.Equal(t, "test@example.com", savedUser.Email)
	assert.Equal(t, "testuser", savedUser.Username)
	assert.Equal(t, "12345", savedUser.Tag)
}

func TestUserCommandAdapter_Save_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := command.NewMySQLUserRepository(db)

	user1 := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword123",
		Username:   "user1",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = adapter.Save(user1)
	require.NoError(t, err)

	user2 := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword456",
		Username:   "user2",
		Tag:        "67890",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = adapter.Save(user2)
	require.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrUserAlreadyExists))
}

func TestUserCommandAdapter_Save_DuplicateTag(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := command.NewMySQLUserRepository(db)

	user1 := &domain.User{
		Email:      "user1@example.com",
		Password:   "hashedPassword123",
		Username:   "user1",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = adapter.Save(user1)
	require.NoError(t, err)

	user2 := &domain.User{
		Email:      "user2@example.com",
		Password:   "hashedPassword456",
		Username:   "user2",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = adapter.Save(user2)
	require.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrUserAlreadyExists))
}

func TestUserCommandAdapter_Update(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := command.NewMySQLUserRepository(db)

	user := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword123",
		Username:   "testuser",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = adapter.Save(user)
	require.NoError(t, err)

	user.Password = "newHashedPassword456"

	err = adapter.Update(user)
	require.NoError(t, err)

	var updatedUser domain.User
	db.First(&updatedUser, user.Id)

	assert.Equal(t, "newHashedPassword456", updatedUser.Password)
}

func TestUserCommandAdapter_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := command.NewMySQLUserRepository(db)

	user := &domain.User{
		Id:         999,
		Email:      "test@example.com",
		Password:   "hashedPassword123",
		Username:   "testuser",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = adapter.Update(user)
	require.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrUserNotFound))
}

func TestUserCommandAdapter_DeleteById(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := command.NewMySQLUserRepository(db)

	user := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword123",
		Username:   "testuser",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = adapter.Save(user)
	require.NoError(t, err)

	err = adapter.DeleteById(user.Id)
	require.NoError(t, err)

	var count int64
	db.Model(&domain.User{}).Where("id = ?", user.Id).Count(&count)

	assert.Equal(t, int64(0), count)
}

func TestUserCommandAdapter_DeleteById_NotFound(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := command.NewMySQLUserRepository(db)

	err = adapter.DeleteById(999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrUserNotFound))
}
