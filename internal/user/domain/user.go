package domain

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/security/password"
	"GAMERS-BE/internal/global/utils"
	"strings"
	"time"
	"unicode"
)

type User struct {
	Id         int64     `gorm:"primaryKey;column:id;autoIncrement" json:"user_id"`
	Email      string    `gorm:"uniqueIndex;column:email;type:varchar(255);not null" json:"email"`
	Password   string    `gorm:"column:password;type:varchar(255);not null" json:"-"`
	Username   string    `gorm:"column:username;type:varchar(16);not_null" json:"username"`
	Tag        string    `gorm:"uniqueIndex;column:tag;type:varchar(6);not_null" json:"tag"`
	Bio        string    `gorm:"column:bio;type:varchar(256);" json:"bio"`
	Avatar     string    `gorm:"column:avatar;type:text;" json:"avatar"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"created_at"`
	ModifiedAt time.Time `gorm:"column:modified_at;type:datetime;not null;autoUpdateTime" json:"modified_at"`
}

func (u *User) TableName() string {
	return "users"
}

func NewUser(email, password, username, tag, bio, avatar string) (*User, error) {
	if err := isValidateEmail(email); err != nil {
		return nil, err
	}
	if err := isValidatePassword(password); err != nil {
		return nil, err
	}
	if err := isValidateUsername(username); err != nil {
		return nil, err
	}
	if err := isValidateTag(tag); err != nil {
		return nil, err
	}
	if err := isValidateBio(bio); err != nil {
		return nil, err
	}

	return &User{
		Email:    email,
		Password: password,
		Username: username,
		Tag:      tag,
		Bio:      bio,
		Avatar:   avatar,
	}, nil
}

func (u *User) EncryptPassword(hasher password.Hasher) error {
	hashPassword, err := hasher.HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashPassword
	return nil
}

func (u *User) UpdateUser(password string, hasher password.Hasher) (*User, error) {
	if err := isValidatePassword(password); err != nil {
		return nil, err
	}

	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u.Password = hashedPassword

	return u, nil
}

func isValidateEmail(email string) error {
	if email == "" {
		return exception.ErrInvalidEmail
	}

	email = strings.TrimSpace(email)

	if len(email) > 256 {
		return exception.ErrInvalidEmail
	}

	if !utils.IsMatchingEmail(email) {
		return exception.ErrInvalidEmail
	}

	return nil
}

func isValidatePassword(password string) error {
	if len(password) < 8 {
		return exception.ErrPasswordTooShort
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

	if complexity < 3 {
		return exception.ErrPasswordTooWeak
	}

	return nil
}

func isValidateUsername(username string) error {
	if username == "" {
		return exception.ErrUsernameEmpty
	}

	if !utils.IsMatchingUsername(username) {
		return exception.ErrUsernameInvalidChar
	}

	if len(username) > 16 {
		return exception.ErrUsernameTooLong
	}

	return nil
}

func isValidateTag(tag string) error {
	if tag == "" {
		return exception.ErrTagEmpty
	}

	if !utils.IsMatchingTag(tag) {
		return exception.ErrTagInvalidChar
	}

	if len(tag) > 5 {
		return exception.ErrTagTooLong
	}

	return nil
}

func isValidateBio(bio string) error {
	if len(bio) > 256 {
		return exception.ErrBioTooLong
	}

	return nil
}
