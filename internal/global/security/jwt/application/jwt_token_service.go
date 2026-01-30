package application

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt/domain"
	"errors"
)

type TokenService struct {
	strategies map[domain.TokenType]domain.TokenStrategy
}

func NewTokenService() *TokenService {
	return &TokenService{
		strategies: make(map[domain.TokenType]domain.TokenStrategy),
	}
}

func (t *TokenService) RegisterStrategy(strategy domain.TokenStrategy) {
	t.strategies[strategy.GetTokenType()] = strategy
}

func (t *TokenService) Generate(tokenType domain.TokenType, userID int64) (string, error) {
	strategy, exists := t.strategies[tokenType]
	if !exists {
		return "", errors.New("token strategy not found")
	}

	return strategy.Generate(userID)
}

func (t *TokenService) GetTTL(tokenType domain.TokenType) *int64 {
	strategy, _ := t.strategies[tokenType]

	ttl := strategy.GetTTL()
	return ttl
}

func (t *TokenService) Validate(tokenType domain.TokenType, tokenString string) (*domain.Claims, error) {
	strategy, exists := t.strategies[tokenType]
	if !exists {
		return nil, errors.New("token strategy not found")
	}

	return strategy.Validate(tokenString)
}

func (t *TokenService) GenerateTokenPair(userID int64) (*dto.TokenResponse, error) {
	accessToken, err := t.Generate(domain.TokenTypeAccess, userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := t.Generate(domain.TokenTypeRefresh, userID)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (t *TokenService) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := t.Validate(domain.TokenTypeRefresh, refreshToken)
	if err != nil {
		return "", err
	}

	return t.Generate(domain.TokenTypeAccess, claims.UserID)
}
