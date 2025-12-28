package application

import (
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/application/port/command"
	"GAMERS-BE/internal/user/application/port/port"
	"GAMERS-BE/internal/user/domain"
)

type ProfileService struct {
	profileQueryPort   port.ProfileQueryPort
	profileCommandPort command.ProfileCommandPort
}

func NewProfileService(profileQueryPort port.ProfileQueryPort, profileCommandPort command.ProfileCommandPort) *ProfileService {
	return &ProfileService{
		profileQueryPort:   profileQueryPort,
		profileCommandPort: profileCommandPort,
	}
}

func (s *ProfileService) CreateProfile(req dto.CreateProfileRequest) (*dto.ProfileResponse, error) {
	profile, err := domain.NewProfile(req.UserId, req.Username, req.Tag, req.Bio, "")
	if err != nil {
		return nil, err
	}

	if err := s.profileCommandPort.Save(profile); err != nil {
		return nil, err
	}

	return toProfileResponse(profile), nil
}

func (s *ProfileService) UpdateProfile(id int64, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	profile, err := s.profileQueryPort.FindById(id)
	if err != nil {
		return nil, err
	}

	newProfile, err := profile.UpdateProfile(req.Username, req.Tag, req.Bio, req.Avatar)

	if err != nil {
		return nil, err
	}

	if err := s.profileCommandPort.Update(newProfile); err != nil {
		return nil, err
	}

	return toProfileResponse(newProfile), nil
}

func (s *ProfileService) DeleteProfile(id int64) error {
	if err := s.profileCommandPort.DeleteById(id); err != nil {
		return err
	}
	return nil
}

func (s *ProfileService) GetProfile(id int64) (*dto.ProfileResponse, error) {
	profile, err := s.profileQueryPort.FindById(id)
	if err != nil {
		return nil, err
	}
	return toProfileResponse(profile), nil
}

func toProfileResponse(profile *domain.Profile) *dto.ProfileResponse {
	return &dto.ProfileResponse{
		Id:       profile.Id,
		Username: profile.Username,
		Tag:      profile.Tag,
		Bio:      profile.Bio,
		Avatar:   profile.Avatar,
	}
}
