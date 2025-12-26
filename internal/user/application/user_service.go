package application

import (
	"GAMERS-BE/internal/common/security/password"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/domain"
	"time"
)

type UserService struct {
	userRepository domain.UserRepository
	passwordHasher password.PasswordHasher
}

func NewUserService(userRepository domain.UserRepository, passwordHasher password.PasswordHasher) *UserService {
	return &UserService{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
	}
}

func (s *UserService) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	user, err := domain.NewInstance(req.Email, req.Password, s.passwordHasher)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := s.userRepository.Save(user); err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

func (s *UserService) GetUserById(id int64) (*dto.UserResponse, error) {
	user, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

func (s *UserService) UpdateUser(id int64, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, err
	}

	updatedUser, err := user.UpdateUser(req.Password, s.passwordHasher)

	if err != nil {
		return nil, err
	}

	updatedUser.UpdatedAt = time.Now()

	if err := s.userRepository.Update(updatedUser); err != nil {
		return nil, err
	}

	return s.toUserResponse(updatedUser), nil
}

func (s *UserService) DeleteUser(id int64) error {
	return s.userRepository.DeleteById(id)
}

func (s *UserService) toUserResponse(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		Id:        user.Id,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
