package application

import (
	"GAMERS-BE/internal/auth/application/dto"
	"GAMERS-BE/internal/auth/application/port"
	"GAMERS-BE/internal/auth/application/port/query"
	"GAMERS-BE/internal/auth/domain"
	"GAMERS-BE/internal/global/security/jwt/application"
	jwtToken "GAMERS-BE/internal/global/security/jwt/domain"
	"GAMERS-BE/internal/global/security/password"
)

type AuthService struct {
	authUserQueryPort     query.AuthUserQueryPort
	refreshTokenCachePort port.RefreshTokenCachePort
	tokenService          application.TokenService
	passwordHasher        password.Hasher
}

func NewAuthService(authUserQueryPort query.AuthUserQueryPort, refreshTokenCachePort port.RefreshTokenCachePort, tokenService application.TokenService, passwordHasher password.Hasher) *AuthService {
	return &AuthService{
		authUserQueryPort:     authUserQueryPort,
		refreshTokenCachePort: refreshTokenCachePort,
		tokenService:          tokenService,
		passwordHasher:        passwordHasher,
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

	token, err := s.tokenService.GenerateTokenPair(user.Id)
	if err != nil {
		return nil, err
	}

	ttl := s.tokenService.GetTTL(jwtToken.TokenTypeRefresh)

	refreshToken := domain.NewRefreshToken(token.RefreshToken, user.Id, *ttl)

	err = s.refreshTokenCachePort.Save(refreshToken, ttl)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (s *AuthService) Logout(req dto.LogoutRequest) error {
	err := s.refreshTokenCachePort.Delete(&req.RefreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Refresh(req dto.RefreshRequest) (*dto.RefreshResponse, error) {
	refreshToken, err := s.refreshTokenCachePort.FindByToken(&req.RefreshToken)
	if err != nil {
		return nil, err
	}

	err = s.refreshTokenCachePort.Delete(&req.RefreshToken)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenService.GenerateTokenPair(refreshToken.UserID)
	if err != nil {
		return nil, err
	}

	ttl := s.tokenService.GetTTL(jwtToken.TokenTypeRefresh)

	newRefreshToken := domain.NewRefreshToken(token.RefreshToken, refreshToken.UserID, *ttl)

	err = s.refreshTokenCachePort.Save(newRefreshToken, ttl)

	if err != nil {
		return nil, err
	}

	return &dto.RefreshResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}
