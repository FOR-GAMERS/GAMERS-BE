package application

import (
	"GAMERS-BE/internal/common/security/password"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/application/port/command"
	"GAMERS-BE/internal/user/application/port/port"
	"GAMERS-BE/internal/user/domain"
	"fmt"
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

	// Auto-create profile with default values
	// Use user ID to generate a unique tag (format: "0001", "0002", etc.)
	// This ensures uniqueness while staying within the 5-character limit
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

// generateDefaultTag creates a unique tag based on user ID
// Returns a zero-padded tag up to 4 digits (e.g., "0001", "0042", "1234")
// For IDs >= 10000, returns the ID as-is (up to 99999 max for 5-char limit)
func generateDefaultTag(userID int64) string {
	if userID < 10000 {
		return fmt.Sprintf("%04d", userID)
	}
	return fmt.Sprintf("%d", userID)
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
