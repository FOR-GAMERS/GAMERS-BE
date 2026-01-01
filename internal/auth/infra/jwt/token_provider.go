package jwt

import (
	"GAMERS-BE/internal/auth/application/dto"
	"time"
)

type TokenProvider struct {
	tokenManager *TokenManager
}

func NewTokenProvider(TokenManager *TokenManager) TokenProvider {
	return TokenProvider{
		tokenManager: TokenManager,
	}
}

func (tp *TokenProvider) PublishToken(id int64) (*dto.TokenResponse, error) {
	accessToken, accessTokenExp, err := tp.publishAccessToken(id)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshTokenExp, err := tp.publishRefreshToken(id)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		AccessTokenExp:  accessTokenExp,
		RefreshTokenExp: refreshTokenExp,
	}, nil
}

func (tp *TokenProvider) publishAccessToken(id int64) (string, int64, error) {
	accessToken, expiresAt, err := tp.tokenManager.GenerateAccessToken(id)
	if err != nil {
		return "", time.Time{}.UnixMilli(), err
	}

	return accessToken, expiresAt, nil
}

func (tp *TokenProvider) publishRefreshToken(id int64) (string, int64, error) {
	refreshToken, refreshTokenExpire, err := tp.tokenManager.GenerateRefreshToken(id)
	if err != nil {
		return "", time.Time{}.UnixMilli(), err
	}

	return refreshToken, refreshTokenExpire, nil
}

func (tp *TokenProvider) GetRefreshTokenDuration() time.Duration {
	return tp.tokenManager.GetRefreshTokenDuration()
}
