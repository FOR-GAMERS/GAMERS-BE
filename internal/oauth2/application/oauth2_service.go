package application

import (
	"GAMERS-BE/internal/auth/infra/jwt"
	"GAMERS-BE/internal/oauth2/application/dto"
	"GAMERS-BE/internal/oauth2/application/port"
	"GAMERS-BE/internal/oauth2/domain"
	"GAMERS-BE/internal/oauth2/infra/discord"
	userCommand "GAMERS-BE/internal/user/application/port/command"
	userDomain "GAMERS-BE/internal/user/domain"
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
)

type OAuth2Service struct {
	ctx                context.Context
	config             *oauth2.Config
	discordClient      *discord.DiscordClient
	oauth2DatabasePort port.OAuth2DatabasePort
	userCommandPort    userCommand.UserCommandPort
	tokenProvider      jwt.TokenProvider
}

func NewOAuth2Service(
	ctx context.Context,
	config *oauth2.Config,
	discordClient *discord.DiscordClient,
	oauth2DatabasePort port.OAuth2DatabasePort,
	userCommandPort userCommand.UserCommandPort,
	tokenProvider jwt.TokenProvider,
) *OAuth2Service {
	return &OAuth2Service{
		ctx:                ctx,
		config:             config,
		discordClient:      discordClient,
		oauth2DatabasePort: oauth2DatabasePort,
		userCommandPort:    userCommandPort,
		tokenProvider:      tokenProvider,
	}
}

func (s *OAuth2Service) GetDiscordLoginURL() string {
	return s.config.AuthCodeURL("state", oauth2.AccessTypeOnline)
}

func (s *OAuth2Service) HandleDiscordCallback(req *dto.DiscordCallbackRequest) (*dto.OAuth2LoginResponse, error) {
	token, err := s.config.Exchange(s.ctx, req.Code)
	if err != nil {
		return nil, errors.New("failed to exchange code for token: " + err.Error())
	}

	userInfo, err := s.discordClient.GetUserInfo(token)
	if err != nil {
		return nil, errors.New("failed to get user info from Discord: " + err.Error())
	}

	discordAccount, err := s.oauth2DatabasePort.FindDiscordAccountByDiscordId(userInfo.Id)
	if err != nil {
		return nil, err
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

	jwtToken, err := s.tokenProvider.PublishToken(userId)
	if err != nil {
		return nil, errors.New("failed to generate JWT token: " + err.Error())
	}

	return &dto.OAuth2LoginResponse{
		AccessToken:  jwtToken.AccessToken,
		RefreshToken: jwtToken.RefreshToken,
		IsNewUser:    isNewUser,
	}, nil
}

func (s *OAuth2Service) createNewUser(userInfo *dto.DiscordUserInfo) (*userDomain.User, error) {
	user := &userDomain.User{
		Email:    fmt.Sprintf("%s@discord.oauth", userInfo.Id),
		Password: "",
		Username: userInfo.Username,
		Tag:      "00000",
		Bio:      "",
		Avatar:   userInfo.Avatar,
	}

	if err := s.userCommandPort.Save(user); err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}

	return user, nil
}
