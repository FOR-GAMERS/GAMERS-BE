package query_test

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/user/domain"
	"GAMERS-BE/internal/user/infra/persistence/query"
	"GAMERS-BE/test/global/support"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserQueryAdapter_FindById(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := query.NewMysqlUserRepository(db)

	user := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword123",
		Username:   "testuser",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	result := db.Create(user)
	require.NoError(t, result.Error)

	foundUser, err := adapter.FindById(user.Id)
	require.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, user.Id, foundUser.Id)
	assert.Equal(t, "test@example.com", foundUser.Email)
}

func TestUserQueryAdapter_FindById_NotFound(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := query.NewMysqlUserRepository(db)

	foundUser, err := adapter.FindById(999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrUserNotFound))
	assert.Nil(t, foundUser)
}

func TestUserQueryAdapter_FindByEmail(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := query.NewMysqlUserRepository(db)

	user := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword123",
		Username:   "testuser",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	result := db.Create(user)
	require.NoError(t, result.Error)

	foundUser, err := adapter.FindByEmail("test@example.com")
	require.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, "test@example.com", foundUser.Email)
	assert.Equal(t, user.Id, foundUser.Id)
}

func TestUserQueryAdapter_FindByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := query.NewMysqlUserRepository(db)

	foundUser, err := adapter.FindByEmail("nonexistent@example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrUserNotFound))
	assert.Nil(t, foundUser)
}

func TestUserQueryAdapter_FindByEmail_MultipleSameEmailNotPossible(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	user1 := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword123",
		Username:   "user1",
		Tag:        "12345",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	result := db.Create(user1)
	require.NoError(t, result.Error)

	user2 := &domain.User{
		Email:      "test@example.com",
		Password:   "hashedPassword456",
		Username:   "user2",
		Tag:        "67890",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	result = db.Create(user2)
	assert.Error(t, result.Error)
}
