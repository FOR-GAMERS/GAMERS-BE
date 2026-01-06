package infra

import (
	"GAMERS-BE/internal/global/security/jwt/domain"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenStrategy struct {
	secretKey string
	duration  time.Duration
	issuer    string
}

func NewAccessTokenStrategy(config *domain.Token) *AccessTokenStrategy {
	return &AccessTokenStrategy{
		secretKey: config.SecretKey,
		duration:  config.AccessTokenDuration,
		issuer:    config.Issuer,
	}
}

func (s *AccessTokenStrategy) Generate(userID int64) (string, error) {
	now := time.Now()

	claims := &domain.Claims{
		UserID:    userID,
		TokenType: domain.TokenTypeAccess,
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

func (s *AccessTokenStrategy) GetTTL() *int64 {
	ttl := s.duration.Milliseconds()
	return &ttl
}

func (s *AccessTokenStrategy) Validate(tokenString string) (*domain.Claims, error) {
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

	if claims.TokenType != domain.TokenTypeAccess {
		return nil, errors.New("invalid token type: expected access token")
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

func (s *AccessTokenStrategy) GetTokenType() domain.TokenType {
	return domain.TokenTypeAccess
}
