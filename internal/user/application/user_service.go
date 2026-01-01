package application

import (
	"GAMERS-BE/internal/global/security/password"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/application/port/command"
	"GAMERS-BE/internal/user/application/port/port"
	"GAMERS-BE/internal/user/domain"
)

type UserService struct {
	userQueryPort   port.UserQueryPort
	userCommandPort command.UserCommandPort
	passwordHasher  password.Hasher
}

func NewUserService(userQueryPort port.UserQueryPort, userCommandPort command.UserCommandPort, passwordHasher password.Hasher) *UserService {
	return &UserService{
		userQueryPort:   userQueryPort,
		userCommandPort: userCommandPort,
		passwordHasher:  passwordHasher,
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

func (s *UserService) DeleteUser(id int64) error {
	return s.userCommandPort.DeleteById(id)
}

func toUserResponse(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		Id:         user.Id,
		Email:      user.Email,
		CreatedAt:  user.CreatedAt,
		ModifiedAt: user.ModifiedAt,
	}
}
