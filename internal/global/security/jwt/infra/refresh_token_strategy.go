package infra

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt/domain"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RefreshTokenStrategy struct {
	secretKey string
	duration  time.Duration
	issuer    string
}

func NewRefreshTokenStrategy(config *domain.Token) *RefreshTokenStrategy {
	return &RefreshTokenStrategy{
		secretKey: config.RefreshSecretKey,
		duration:  config.RefreshTokenDuration,
		issuer:    config.Issuer,
	}
}

func (s *RefreshTokenStrategy) Generate(userID int64) (string, error) {
	now := time.Now()

	claims := &domain.Claims{
		UserID:    userID,
		TokenType: domain.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *RefreshTokenStrategy) GetTTL() *int64 {
	ttl := s.duration.Milliseconds()
	return &ttl
}

func (s *RefreshTokenStrategy) Validate(tokenString string) (*domain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*domain.Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Validate token type
	if claims.TokenType != domain.TokenTypeRefresh {
		return nil, errors.New("invalid token type: expected refresh token")
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

func (s *RefreshTokenStrategy) GetTokenType() domain.TokenType {
	return domain.TokenTypeRefresh
}
