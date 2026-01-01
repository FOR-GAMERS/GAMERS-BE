package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	config *Config
}

func NewTokenManager(config *Config) *TokenManager {
	return &TokenManager{config: config}
}

func (tm *TokenManager) GenerateAccessToken(userId int64) (string, int64, error) {
	expiresAt := generateExpiration(tm.config.AccessTokenDuration)

	claims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    tm.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tm.config.SecretKey))
	if err != nil {
		return "", time.Time{}.Unix(), err
	}

	return tokenString, expiresAt.Unix(), nil
}

func (tm *TokenManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(tm.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (tm *TokenManager) GenerateRefreshToken(userId int64) (string, int64, error) {
	expiresAt := generateExpiration(tm.config.RefreshTokenDuration)

	claims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    tm.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tm.config.RefreshSecretKey))
	if err != nil {
		return "", time.Time{}.Unix(), err
	}

	return tokenString, expiresAt.Unix(), nil
}

func (tm *TokenManager) GetRefreshTokenDuration() time.Duration {
	return tm.config.RefreshTokenDuration
}

func generateExpiration(ttl time.Duration) time.Time {
	return time.Now().Add(ttl)
}
