package domain_test

import (
	"GAMERS-BE/internal/common/security/password"
	"GAMERS-BE/internal/user/domain"
	"errors"
	"testing"
)

func TestNewInstance_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid email and password",
			email:    "test@example.com",
			password: "SecurePass123!",
			wantErr:  false,
		},
		{
			name:     "Valid with number and special char",
			email:    "user@test.com",
			password: "Password1@",
			wantErr:  false,
		},
	}

	hasher := password.NewBcryptPasswordHasher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.email, tt.password, hasher)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if user.Email != tt.email {
					t.Errorf("Expected email %s, got %s", tt.email, user.Email)
				}
				if err := hasher.ComparePassword(user.Password, tt.password); err != nil {
					t.Errorf("Password hash verification failed: %v", err)
				}
			}
		})
	}
}

func TestNewInstance_InvalidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "Invalid email format",
			email:    "invalid-email",
			password: "SecurePass123!",
			wantErr:  domain.ErrInvalidEmail,
		},
		{
			name:     "Empty email",
			email:    "",
			password: "SecurePass123!",
			wantErr:  domain.ErrInvalidEmail,
		},
		{
			name:     "Missing @ symbol",
			email:    "userexample.com",
			password: "SecurePass123!",
			wantErr:  domain.ErrInvalidEmail,
		},
	}

	hasher := password.NewBcryptPasswordHasher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser(tt.email, tt.password, hasher)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNewInstance_InvalidPassword(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "Password too short",
			email:    "test@example.com",
			password: "Short1!",
			wantErr:  domain.ErrPasswordTooShort,
		},
		{
			name:     "Password too weak - only lowercase",
			email:    "test@example.com",
			password: "weakpassword",
			wantErr:  domain.ErrPasswordTooWeak,
		},
		{
			name:     "Password too weak - only one type",
			email:    "test@example.com",
			password: "12345678",
			wantErr:  domain.ErrPasswordTooWeak,
		},
	}

	hasher := password.NewBcryptPasswordHasher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewUser(tt.email, tt.password, hasher)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUpdateUser_ValidPassword(t *testing.T) {
	hasher := password.NewBcryptPasswordHasher()

	user, err := domain.NewUser("test@example.com", "OldPassword123!", hasher)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	newPassword := "NewPassword456@"
	updatedUser, err := user.UpdateUser(newPassword, hasher)
	if err != nil {
		t.Errorf("UpdateUser() error = %v, expected nil", err)
	}

	if err := hasher.ComparePassword(updatedUser.Password, newPassword); err != nil {
		t.Errorf("Password hash verification failed: %v", err)
	}

	if updatedUser.Email != "test@example.com" {
		t.Errorf("Email should not change, got %s", updatedUser.Email)
	}
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
			wantErr:     domain.ErrPasswordTooShort,
		},
		{
			name:        "Password too weak",
			newPassword: "weakpassword",
			wantErr:     domain.ErrPasswordTooWeak,
		},
		{
			name:        "Password only numbers",
			newPassword: "12345678",
			wantErr:     domain.ErrPasswordTooWeak,
		},
	}

	hasher := password.NewBcryptPasswordHasher()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser("test@example.com", "OldPassword123!", hasher)
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}

			_, err = user.UpdateUser(tt.newPassword, hasher)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
