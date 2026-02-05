package application

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/application/port/query"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt/application"
	jwtToken "github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/password"
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

	token, err := s.tokenService.GenerateTokenPair(user.Id, string(user.Role))
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

	claims, err := s.tokenService.Validate(jwtToken.TokenTypeRefresh, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	err = s.refreshTokenCachePort.Delete(&req.RefreshToken)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenService.GenerateTokenPair(refreshToken.UserID, claims.Role)
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
