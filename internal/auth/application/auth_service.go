package application

import (
	"GAMERS-BE/internal/auth/application/dto"
	"GAMERS-BE/internal/auth/application/port/command"
	"GAMERS-BE/internal/auth/application/port/query"
	"GAMERS-BE/internal/auth/domain"
	"GAMERS-BE/internal/auth/infra/jwt"
	"GAMERS-BE/internal/global/security/password"
	"context"
)

type AuthService struct {
	ctx                     context.Context
	authUserQueryPort       query.AuthUserQueryPort
	refreshTokenQueryPort   query.RefreshTokenQueryPort
	refreshTokenCommandPort command.RefreshTokenCommandPort
	tokenProvider           jwt.TokenProvider
	passwordHasher          password.Hasher
}

func NewAuthService(ctx context.Context, authUserQueryPort query.AuthUserQueryPort, refreshTokenCommandPort command.RefreshTokenCommandPort, refreshTokenQueryPort query.RefreshTokenQueryPort, tokenProvider jwt.TokenProvider, passwordHasher password.Hasher) *AuthService {
	return &AuthService{
		ctx:                     ctx,
		authUserQueryPort:       authUserQueryPort,
		refreshTokenCommandPort: refreshTokenCommandPort,
		refreshTokenQueryPort:   refreshTokenQueryPort,
		tokenProvider:           tokenProvider,
		passwordHasher:          passwordHasher,
	}
}

func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.authUserQueryPort.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	err = s.passwordHasher.ComparePassword(user.Password, req.Password)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenProvider.PublishToken(user.Id)
	if err != nil {
		return nil, err
	}

	refreshToken := domain.NewRefreshToken(token.RefreshToken, user.Id, token.RefreshTokenExp)

	ttl := s.tokenProvider.GetRefreshTokenDuration()
	err = s.refreshTokenCommandPort.Save(s.ctx, refreshToken, ttl)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (s *AuthService) Logout(req dto.LogoutRequest) error {
	err := s.refreshTokenCommandPort.Delete(s.ctx, req.RefreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Refresh(req dto.RefreshRequest) (*dto.TokenResponse, error) {
	refreshToken, err := s.refreshTokenQueryPort.FindByToken(s.ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	err = s.refreshTokenCommandPort.Delete(s.ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenProvider.PublishToken(refreshToken.UserID)
	if err != nil {
		return nil, err
	}

	newRefreshToken := domain.NewRefreshToken(token.RefreshToken, refreshToken.UserID, token.RefreshTokenExp)
	ttl := s.tokenProvider.GetRefreshTokenDuration()
	err = s.refreshTokenCommandPort.Save(s.ctx, newRefreshToken, ttl)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:     token.AccessToken,
		AccessTokenExp:  token.AccessTokenExp,
		RefreshToken:    token.RefreshToken,
		RefreshTokenExp: token.RefreshTokenExp,
	}, nil
}
