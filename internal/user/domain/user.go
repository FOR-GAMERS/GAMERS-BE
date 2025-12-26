package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type User struct {
	Id        int64     `json:"user_id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrEmailCannotChange = errors.New("email cannot be changed")
)

func NewInstance(email, password string) (*User, error) {
	if err := isValidateEmail(email); err != nil {
		return nil, err
	}

	if err := isValidatePassword(password); err != nil {
		return nil, err
	}

	return &User{
		Email:    email,
		Password: password,
	}, nil
}

func (u *User) UpdateUser(password string) (*User, error) {
	if err := isValidatePassword(password); err != nil {
		return nil, err
	}

	u.Password = password

	return u, nil
}

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooWeak  = errors.New("password must contain at least 2 of: uppercase, lowercase, number, special character")
	ErrInvalidEmail     = errors.New("invalid email format")
)

func isValidateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	email = strings.TrimSpace(email)

	if len(email) > 254 {
		return ErrInvalidEmail
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

func isValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	complexity := 0
	if hasUpper {
		complexity++
	}
	if hasLower {
		complexity++
	}
	if hasNumber {
		complexity++
	}
	if hasSpecial {
		complexity++
	}

	if complexity < 2 {
		return ErrPasswordTooWeak
	}

	return nil
}
