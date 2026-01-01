package domain

type RefreshToken struct {
	Token     string
	UserID    int64
	ExpiresAt int64
}

func NewRefreshToken(token string, userID int64, expiresAt int64) *RefreshToken {
	return &RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
}
