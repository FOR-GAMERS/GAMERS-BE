package application

import (
	"GAMERS-BE/internal/common/security/password"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/application/port/command"
	"GAMERS-BE/internal/user/application/port/port"
	"GAMERS-BE/internal/user/domain"
	"fmt"
	"math/rand"
	"time"
)

type UserService struct {
	userQueryPort      port.UserQueryPort
	userCommandPort    command.UserCommandPort
	profileCommandPort command.ProfileCommandPort
	passwordHasher     password.PasswordHasher
}

func NewUserService(userQueryPort port.UserQueryPort, userCommandPort command.UserCommandPort, profileCommandPort command.ProfileCommandPort, passwordHasher password.PasswordHasher) *UserService {
	return &UserService{
		userQueryPort:      userQueryPort,
		userCommandPort:    userCommandPort,
		profileCommandPort: profileCommandPort,
		passwordHasher:     passwordHasher,
	}
}

func (s *UserService) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	user, err := domain.NewUser(req.Email, req.Password, s.passwordHasher)
	if err != nil {
		return nil, err
	}

	if err := s.userCommandPort.Save(user); err != nil {
		return nil, err
	}

	tag := generateDefaultTag(user.Id)
	profile, err := domain.NewProfile(user.Id, "User", tag, "", "")
	if err != nil {
		return nil, err
	}

	if err := s.profileCommandPort.Save(profile); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func generateDefaultTag(userID int64) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := r.Intn(100000)
	return fmt.Sprintf("%04d", randomNum)
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
		Id:        user.Id,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
