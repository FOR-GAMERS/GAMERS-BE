package application

import (
	"GAMERS-BE/internal/global/security/password"
	"GAMERS-BE/internal/global/utils"
	"GAMERS-BE/internal/oauth2/application/port"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/application/port/command"
	userPort "GAMERS-BE/internal/user/application/port/port"
	"GAMERS-BE/internal/user/domain"
)

type UserService struct {
	userQueryPort    userPort.UserQueryPort
	userCommandPort  command.UserCommandPort
	passwordHasher   password.Hasher
	oauth2DbPort     port.OAuth2DatabasePort
}

func NewUserService(
	userQueryPort userPort.UserQueryPort,
	userCommandPort command.UserCommandPort,
	passwordHasher password.Hasher,
	oauth2DbPort port.OAuth2DatabasePort,
) *UserService {
	return &UserService{
		userQueryPort:   userQueryPort,
		userCommandPort: userCommandPort,
		passwordHasher:  passwordHasher,
		oauth2DbPort:    oauth2DbPort,
	}
}

func (s *UserService) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	user, err := domain.NewUser(req.Email, req.Password, req.Username, req.Tag, req.Bio, req.Avatar)
	if err != nil {
		return nil, err
	}

	err = user.EncryptPassword(s.passwordHasher)
	if err != nil {
		return nil, err
	}

	if err := s.userCommandPort.Save(user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func (s *UserService) GetUserById(id int64) (*dto.UserResponse, error) {
	user, err := s.userQueryPort.FindById(id)
	if err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func (s *UserService) UpdateUser(id int64, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userQueryPort.FindById(id)
	if err != nil {
		return nil, err
	}

	updatedUser, err := user.UpdateUser(req.Password, s.passwordHasher)

	if err != nil {
		return nil, err
	}

	if err := s.userCommandPort.Update(updatedUser); err != nil {
		return nil, err
	}

	return toUserResponse(updatedUser), nil
}

func (s *UserService) UpdateUserInfo(id int64, req dto.UpdateUserInfoRequest) (*dto.MyUserResponse, error) {
	user, err := s.userQueryPort.FindById(id)
	if err != nil {
		return nil, err
	}

	if err := user.UpdateUserInfo(req.Username, req.Tag, req.Bio, req.Avatar); err != nil {
		return nil, err
	}

	if err := s.userCommandPort.UpdateUserInfo(user); err != nil {
		return nil, err
	}

	avatarURL := user.Avatar
	if s.oauth2DbPort != nil {
		discordAccount, err := s.oauth2DbPort.FindDiscordAccountByUserId(id)
		if err == nil && discordAccount != nil {
			if url := utils.BuildDiscordAvatarURL(discordAccount.DiscordId, discordAccount.DiscordAvatar); url != "" {
				avatarURL = url
			}
		}
	}

	return toMyUserResponseWithAvatar(user, avatarURL), nil
}

func (s *UserService) DeleteUser(id int64) error {
	return s.userCommandPort.DeleteById(id)
}

func (s *UserService) GetMyInfo(id int64) (*dto.MyUserResponse, error) {
	user, err := s.userQueryPort.FindById(id)
	if err != nil {
		return nil, err
	}

	// Try to build Discord avatar URL if user has Discord account
	avatarURL := user.Avatar
	if s.oauth2DbPort != nil {
		discordAccount, err := s.oauth2DbPort.FindDiscordAccountByUserId(id)
		if err == nil && discordAccount != nil {
			if url := utils.BuildDiscordAvatarURL(discordAccount.DiscordId, discordAccount.DiscordAvatar); url != "" {
				avatarURL = url
			}
		}
	}

	return toMyUserResponseWithAvatar(user, avatarURL), nil
}

func toUserResponse(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		Id:         user.Id,
		Email:      user.Email,
		CreatedAt:  user.CreatedAt,
		ModifiedAt: user.ModifiedAt,
	}
}

func toMyUserResponse(user *domain.User) *dto.MyUserResponse {
	return toMyUserResponseWithAvatar(user, user.Avatar)
}

func toMyUserResponseWithAvatar(user *domain.User, avatarURL string) *dto.MyUserResponse {
	return &dto.MyUserResponse{
		Id:                 user.Id,
		Email:              user.Email,
		Username:           user.Username,
		Tag:                user.Tag,
		Bio:                user.Bio,
		Avatar:             avatarURL,
		ProfileKey:         user.ProfileKey,
		CreatedAt:          user.CreatedAt,
		ModifiedAt:         user.ModifiedAt,
		RiotName:           user.RiotName,
		RiotTag:            user.RiotTag,
		Region:             user.Region,
		CurrentTier:        user.CurrentTier,
		CurrentTierPatched: user.CurrentTierPatched,
		Elo:                user.Elo,
		RankingInTier:      user.RankingInTier,
		PeakTier:           user.PeakTier,
		PeakTierPatched:    user.PeakTierPatched,
		ValorantUpdatedAt:  user.ValorantUpdatedAt,
	}
}
