package domain_test

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/security/password"
	"GAMERS-BE/internal/user/domain"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		username string
		tag      string
		bio      string
		avatar   string
		wantErr  bool
	}{
		{
			name:     "Valid user with all fields",
			email:    "test@example.com",
			password: "SecurePass123!",
			username: "testuser",
			tag:      "12345",
			bio:      "This is my bio",
			avatar:   "https://example.com/avatar.jpg",
			wantErr:  false,
		},
		{
			name:     "Valid user with minimal fields",
			email:    "user@test.com",
			password: "Password1@",
			username: "user",
			tag:      "abc12",
			bio:      "",
			avatar:   "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.email, tt.password, tt.username, tt.tag, tt.bio, tt.avatar)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.email, user.Email)
			assert.Equal(t, tt.username, user.Username)
			assert.Equal(t, tt.tag, user.Tag)
			assert.Equal(t, tt.bio, user.Bio)
			assert.Equal(t, tt.avatar, user.Avatar)
		})
	}
}

func TestNewUser_InvalidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		username string
		tag      string
		wantErr  error
	}{
		{
			name:     "Invalid email format",
			email:    "invalid-email",
			password: "SecurePass123!",
			username: "testuser",
			tag:      "12345",
			wantErr:  exception.ErrInvalidEmail,
		},
		{
			name:     "Empty email",
			email:    "",
			password: "SecurePass123!",
			username: "testuser",
			tag:      "12345",
			wantErr:  exception.ErrInvalidEmail,
		},
		{
			name:     "Missing @ symbol",
			email:    "userexample.com",
			password: "SecurePass123!",
			username: "testuser",
			tag:      "12345",
			wantErr:  exception.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser(tt.email, tt.password, tt.username, tt.tag, "", "")
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func TestNewUser_InvalidPassword(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		username string
		tag      string
		wantErr  error
	}{
		{
			name:     "Password too short",
			email:    "test@example.com",
			password: "Short1!",
			username: "testuser",
			tag:      "12345",
			wantErr:  exception.ErrPasswordTooShort,
		},
		{
			name:     "Password too weak - only lowercase",
			email:    "test@example.com",
			password: "weakpassword",
			username: "testuser",
			tag:      "12345",
			wantErr:  exception.ErrPasswordTooWeak,
		},
		{
			name:     "Password too weak - only numbers",
			email:    "test@example.com",
			password: "12345678",
			username: "testuser",
			tag:      "12345",
			wantErr:  exception.ErrPasswordTooWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser(tt.email, tt.password, tt.username, tt.tag, "", "")
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func TestNewUser_InvalidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "Empty username",
			username: "",
			wantErr:  exception.ErrUsernameEmpty,
		},
		{
			name:     "Username too long",
			username: "thisusernameiswaytoolong",
			wantErr:  exception.ErrUsernameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser("test@example.com", "SecurePass123!", tt.username, "12345", "", "")
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func TestNewUser_InvalidTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantErr error
	}{
		{
			name:    "Empty tag",
			tag:     "",
			wantErr: exception.ErrTagEmpty,
		},
		{
			name:    "Tag too long",
			tag:     "123456",
			wantErr: exception.ErrTagTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser("test@example.com", "SecurePass123!", "testuser", tt.tag, "", "")
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func TestNewUser_InvalidBio(t *testing.T) {
	longBio := make([]byte, 257)
	for i := range longBio {
		longBio[i] = 'a'
	}

	_, err := domain.NewUser("test@example.com", "SecurePass123!", "testuser", "12345", string(longBio), "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrBioTooLong))
}

func TestEncryptPassword_Success(t *testing.T) {
	hasher := password.NewBcryptPasswordHasher()

	user, err := domain.NewUser("test@example.com", "SecurePass123!", "testuser", "12345", "", "")
	require.NoError(t, err)

	originalPassword := user.Password
	err = user.EncryptPassword(hasher)
	require.NoError(t, err)

	assert.NotEqual(t, originalPassword, user.Password)
	err = hasher.ComparePassword(user.Password, originalPassword)
	assert.NoError(t, err)
}

func TestUpdateUser_ValidPassword(t *testing.T) {
	hasher := password.NewBcryptPasswordHasher()

	user, err := domain.NewUser("test@example.com", "OldPassword123!", "testuser", "12345", "", "")
	require.NoError(t, err)

	err = user.EncryptPassword(hasher)
	require.NoError(t, err)

	newPassword := "NewPassword456@"
	updatedUser, err := user.UpdateUser(newPassword, hasher)
	require.NoError(t, err)

	assert.NotNil(t, updatedUser)
	err = hasher.ComparePassword(updatedUser.Password, newPassword)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", updatedUser.Email)
}

func TestUpdateUser_InvalidPassword(t *testing.T) {
	tests := []struct {
		name        string
		newPassword string
		wantErr     error
	}{
		{
			name:        "Password too short",
			newPassword: "Short1!",
			wantErr:     exception.ErrPasswordTooShort,
		},
		{
			name:        "Password too weak",
			newPassword: "weakpassword",
			wantErr:     exception.ErrPasswordTooWeak,
		},
		{
			name:        "Password only numbers",
			newPassword: "12345678",
			wantErr:     exception.ErrPasswordTooWeak,
		},
	}

	hasher := password.NewBcryptPasswordHasher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser("test@example.com", "OldPassword123!", "testuser", "12345", "", "")
			require.NoError(t, err)

			_, err = user.UpdateUser(tt.newPassword, hasher)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}
