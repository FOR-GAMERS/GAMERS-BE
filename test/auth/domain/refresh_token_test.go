package domain_test

import (
	"GAMERS-BE/internal/auth/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRefreshToken_Success(t *testing.T) {
	token := "test-refresh-token"
	userID := int64(123)
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	refreshToken := domain.NewRefreshToken(token, userID, expiresAt)

	assert.NotNil(t, refreshToken)
	assert.Equal(t, token, refreshToken.Token)
	assert.Equal(t, userID, refreshToken.UserID)
	assert.Equal(t, expiresAt, refreshToken.ExpiresAt)
}

func TestNewRefreshToken_WithDifferentValues(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		userID    int64
		expiresAt int64
	}{
		{
			name:      "Standard refresh token",
			token:     "refresh-token-1",
			userID:    1,
			expiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
		},
		{
			name:      "Long token",
			token:     "very-long-refresh-token-with-many-characters-12345",
			userID:    999,
			expiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
		},
		{
			name:      "Short expiration",
			token:     "short-exp-token",
			userID:    42,
			expiresAt: time.Now().Add(1 * time.Hour).Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken := domain.NewRefreshToken(tt.token, tt.userID, tt.expiresAt)

			assert.NotNil(t, refreshToken)
			assert.Equal(t, tt.token, refreshToken.Token)
			assert.Equal(t, tt.userID, refreshToken.UserID)
			assert.Equal(t, tt.expiresAt, refreshToken.ExpiresAt)
		})
	}
}
