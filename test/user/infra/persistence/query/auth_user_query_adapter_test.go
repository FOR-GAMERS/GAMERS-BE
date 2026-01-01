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

func TestAuthUserQueryAdapter_FindByEmail(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := query.NewAuthUserQueryAdapter(db)

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
	assert.Equal(t, "testuser", foundUser.Username)
}

func TestAuthUserQueryAdapter_FindByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := query.NewAuthUserQueryAdapter(db)

	foundUser, err := adapter.FindByEmail("nonexistent@example.com")
	require.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrUserNotFound))
	assert.Nil(t, foundUser)
}

func TestAuthUserQueryAdapter_FindByEmail_WithPassword(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := query.NewAuthUserQueryAdapter(db)

	user := &domain.User{
		Email:      "auth@example.com",
		Password:   "$2a$10$hashedPasswordValue",
		Username:   "authuser",
		Tag:        "99999",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	result := db.Create(user)
	require.NoError(t, result.Error)

	foundUser, err := adapter.FindByEmail("auth@example.com")
	require.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, "$2a$10$hashedPasswordValue", foundUser.Password)
}

func TestAuthUserQueryAdapter_FindByEmail_CaseSensitive(t *testing.T) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	require.NoError(t, err)
	defer container.Teardown(ctx)

	db := container.GetDB()
	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	adapter := query.NewAuthUserQueryAdapter(db)

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
}
