package application

import (
	authDomain "GAMERS-BE/internal/auth/domain"
	authPort "GAMERS-BE/internal/auth/application/port"
	discordDto "GAMERS-BE/internal/discord/application/dto"
	discordPort "GAMERS-BE/internal/discord/application/port"
	"GAMERS-BE/internal/global/exception"
	jwtApplication "GAMERS-BE/internal/global/security/jwt/application"
	jwtDomain "GAMERS-BE/internal/global/security/jwt/domain"
	"GAMERS-BE/internal/global/utils"
	"GAMERS-BE/internal/oauth2/application/dto"
	"GAMERS-BE/internal/oauth2/application/port"
	"GAMERS-BE/internal/oauth2/domain"
	"GAMERS-BE/internal/oauth2/infra/discord"
	"GAMERS-BE/internal/oauth2/infra/state"
	userDomain "GAMERS-BE/internal/user/domain"
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
)

type DiscordService struct {
	ctx                   context.Context
	config                *oauth2.Config
	discordClient         *discord.Client
	stateManager          *state.Manager
	oauth2UserPort        port.OAuth2UserPort
	oauth2DatabasePort    port.OAuth2DatabasePort
	discordTokenPort      discordPort.DiscordTokenPort
	refreshTokenCachePort authPort.RefreshTokenCachePort
	tokenService          jwtApplication.TokenService
}

func NewOAuth2Service(
	ctx context.Context,
	config *oauth2.Config,
	discordClient *discord.Client,
	stateManager *state.Manager,
	oauth2UserPort port.OAuth2UserPort,
	oauth2DatabasePort port.OAuth2DatabasePort,
	discordTokenPort discordPort.DiscordTokenPort,
	refreshTokenCachePort authPort.RefreshTokenCachePort,
	tokenService jwtApplication.TokenService,
) *DiscordService {
	return &DiscordService{
		ctx:                   ctx,
		config:                config,
		discordClient:         discordClient,
		stateManager:          stateManager,
		oauth2UserPort:        oauth2UserPort,
		oauth2DatabasePort:    oauth2DatabasePort,
		discordTokenPort:      discordTokenPort,
		refreshTokenCachePort: refreshTokenCachePort,
		tokenService:          tokenService,
	}
}

func (s *DiscordService) GetDiscordLoginURL() (string, error) {
	randomState, err := s.stateManager.GenerateState()
	if err != nil {
		return "", err
	}
	return s.config.AuthCodeURL(randomState, oauth2.AccessTypeOnline), nil
}

func (s *DiscordService) HandleDiscordCallback(req *dto.DiscordCallbackRequest) (*dto.OAuth2LoginResponse, error) {
	token, err := s.config.Exchange(s.ctx, req.Code)
	if err != nil {
		return nil, exception.ErrDiscordTokenExchange
	}

	userInfo, err := s.discordClient.GetUserInfo(token)
	if err != nil {
		return nil, exception.ErrDiscordCannotGetUserInfo
	}

	discordAccount, err := s.oauth2DatabasePort.FindDiscordAccountByDiscordId(userInfo.Id)
	if err != nil {
		if !errors.Is(err, exception.ErrDiscordUserCannotFound) {
			return nil, err
		}
	}

	var userId int64
	isNewUser := false

	if discordAccount == nil {
		user, err := s.createNewUser(userInfo)
		if err != nil {
			return nil, err
		}

		discordAccount = &domain.DiscordAccount{
			DiscordId:       userInfo.Id,
			UserId:          user.Id,
			DiscordAvatar:   userInfo.Avatar,
			DiscordVerified: userInfo.Verified,
		}

		if err := s.oauth2DatabasePort.CreateDiscordAccount(discordAccount); err != nil {
			return nil, errors.New("failed to create discord account: " + err.Error())
		}

		userId = user.Id
		isNewUser = true
	} else {
		discordAccount.DiscordAvatar = userInfo.Avatar
		discordAccount.DiscordVerified = userInfo.Verified

		if err := s.oauth2DatabasePort.UpdateDiscordAccount(discordAccount); err != nil {
			return nil, errors.New("failed to update discord account: " + err.Error())
		}

		userId = discordAccount.UserId
	}

	// Save Discord OAuth2 token to Redis for future API calls
	discordToken := &discordDto.DiscordToken{
		UserID:       userId,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry.Unix(),
	}
	if err := s.discordTokenPort.SaveToken(discordToken); err != nil {
		// Log the error but don't fail the login
		// The user can still use the app, just guild features may be limited
		fmt.Printf("Warning: failed to save Discord token to Redis: %v\n", err)
	}

	jwtToken, err := s.tokenService.GenerateTokenPair(userId)
	if err != nil {
		return nil, errors.New("failed to generate JWT token: " + err.Error())
	}

	// Save refresh token to Redis
	ttl := s.tokenService.GetTTL(jwtDomain.TokenTypeRefresh)
	refreshToken := authDomain.NewRefreshToken(jwtToken.RefreshToken, userId, *ttl)
	if err := s.refreshTokenCachePort.Save(refreshToken, ttl); err != nil {
		return nil, errors.New("failed to save refresh token: " + err.Error())
	}

	return &dto.OAuth2LoginResponse{
		AccessToken:  jwtToken.AccessToken,
		RefreshToken: jwtToken.RefreshToken,
		IsNewUser:    isNewUser,
	}, nil
}

func (s *DiscordService) createNewUser(userInfo *dto.DiscordUserInfo) (*userDomain.User, error) {
	const retriesCnt = 10

	user := &userDomain.User{
		Email:    fmt.Sprintf("%s@discord.oauth", userInfo.Id),
		Password: utils.GenerateSecurePassword(),
		Username: userInfo.Username,
		Bio:      "",
		Avatar:   userInfo.Avatar,
	}

	var err error

	for i := 0; i < retriesCnt; i++ {
		user.Tag, err = utils.GenerateRandomTag()

		if err == nil {
			if err := s.oauth2UserPort.SaveRandomUser(user); err != nil {
				return nil, errors.New("failed to create user: " + err.Error())
			}
			return user, err
		}
	}

	return nil, err
}
