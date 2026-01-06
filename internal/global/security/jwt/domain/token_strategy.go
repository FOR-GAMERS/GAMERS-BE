package domain

import (
	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "ACCESS"
	TokenTypeRefresh TokenType = "REFRESH"
)

type TokenStrategy interface {
	Generate(userID int64) (string, error)
	GetTTL() *int64
	Validate(tokenString string) (*Claims, error)
	GetTokenType() TokenType
}

type Claims struct {
	UserID    int64     `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}
