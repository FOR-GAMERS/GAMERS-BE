package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type Profile struct {
	Id        int64     `gorm:"primary_key;column:profile_id;autoIncrement" json:"profileId"`
	UserId    int64     `gorm:"uniqueIndex;column:user_id;not_null" json:"userId"`
	Username  string    `gorm:"column:username;type:varchar(16);not_null" json:"username"`
	Tag       string    `gorm:"uniqueIndex;column:tag;type:varchar(6);not_null" json:"tag"`
	Bio       string    `gorm:"column:bio;type:varchar(256);" json:"bio"`
	Avatar    string    `gorm:"column:profile_url;type:text;" json:"avatar"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime" json:"updated_at"`
}

func (p *Profile) TableName() string {
	return "profiles"
}

var (
	ErrProfileNotFound      = errors.New("profile not found")
	ErrProfileAlreadyExists = errors.New("profile already exists")

	ErrUsernameEmpty       = errors.New("username cannot be empty")
	ErrUsernameTooLong     = errors.New("username must be 16 characters or less")
	ErrUsernameInvalidChar = errors.New("username can only contain letters and numbers")

	ErrTagEmpty       = errors.New("tag cannot be empty")
	ErrTagTooLong     = errors.New("tag must be less than 6 characters")
	ErrTagInvalidChar = errors.New("tag can only contain letters and numbers")

	ErrBioTooLong = errors.New("bio is too long")
)

func NewProfile(userId int64, username, tag, bio, avatar string) (*Profile, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	if err := validateTag(tag); err != nil {
		return nil, err
	}

	if err := validateBio(bio); err != nil {
		return nil, err
	}

	return &Profile{
		UserId:   userId,
		Username: username,
		Tag:      tag,
		Bio:      bio,
		Avatar:   avatar,
	}, nil
}

func (p *Profile) UpdateProfile(username, tag, bio, avatar string) (*Profile, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}

	if err := validateTag(tag); err != nil {
		return nil, err
	}

	if err := validateBio(bio); err != nil {
		return nil, err
	}

	p.Username = username
	p.Tag = tag
	p.Bio = bio
	p.Avatar = avatar

	return p, nil
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
var tagRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return ErrUsernameEmpty
	}

	if len(username) > 16 {
		return ErrUsernameTooLong
	}

	if !usernameRegex.MatchString(username) {
		return ErrUsernameInvalidChar
	}

	return nil
}

func validateTag(tag string) error {
	tag = strings.TrimSpace(tag)

	if tag == "" {
		return ErrTagEmpty
	}

	if len(tag) >= 6 {
		return ErrTagTooLong
	}

	if !tagRegex.MatchString(tag) {
		return ErrTagInvalidChar
	}

	return nil
}

func validateBio(bio string) error {
	if len(bio) > 256 {
		return ErrBioTooLong
	}

	return nil
}
