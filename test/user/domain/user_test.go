package domain

import (
	domain2 "GAMERS-BE/internal/user/domain"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain2.NewInstance(tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if user.Email != tt.email {
					t.Errorf("Expected email %s, got %s", tt.email, user.Email)
				}
				if user.Password != tt.password {
					t.Errorf("Expected password %s, got %s", tt.password, user.Password)
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
			wantErr:  domain2.ErrInvalidEmail,
		},
		{
			name:     "Empty email",
			email:    "",
			password: "SecurePass123!",
			wantErr:  domain2.ErrInvalidEmail,
		},
		{
			name:     "Missing @ symbol",
			email:    "userexample.com",
			password: "SecurePass123!",
			wantErr:  domain2.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain2.NewInstance(tt.email, tt.password)
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
			wantErr:  domain2.ErrPasswordTooShort,
		},
		{
			name:     "Password too weak - only lowercase",
			email:    "test@example.com",
			password: "weakpassword",
			wantErr:  domain2.ErrPasswordTooWeak,
		},
		{
			name:     "Password too weak - only one type",
			email:    "test@example.com",
			password: "12345678",
			wantErr:  domain2.ErrPasswordTooWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain2.NewInstance(tt.email, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
