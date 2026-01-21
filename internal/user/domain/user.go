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
	Username   string    `gorm:"uniqueIndex:idx_username_tag;column:username;type:varchar(16);not_null" json:"username"`
	Tag        string    `gorm:"uniqueIndex:idx_username_tag;column:tag;type:varchar(6);not_null" json:"tag"`
	Bio        string    `gorm:"column:bio;type:varchar(256);" json:"bio"`
	Avatar     string    `gorm:"column:avatar;type:text;" json:"avatar"`
	ProfileKey *string   `gorm:"column:profile_key;type:varchar(512)" json:"profile_key,omitempty"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime" json:"created_at"`
	ModifiedAt time.Time `gorm:"column:modified_at;type:datetime;not null;autoUpdateTime" json:"modified_at"`

	// Valorant fields
	RiotName           *string    `gorm:"column:riot_name;type:varchar(32)" json:"riot_name,omitempty"`
	RiotTag            *string    `gorm:"column:riot_tag;type:varchar(8)" json:"riot_tag,omitempty"`
	Region             *string    `gorm:"column:region;type:varchar(10)" json:"region,omitempty"`
	CurrentTier        *int       `gorm:"column:current_tier" json:"current_tier,omitempty"`
	CurrentTierPatched *string    `gorm:"column:current_tier_patched;type:varchar(32)" json:"current_tier_patched,omitempty"`
	Elo                *int       `gorm:"column:elo" json:"elo,omitempty"`
	RankingInTier      *int       `gorm:"column:ranking_in_tier" json:"ranking_in_tier,omitempty"`
	PeakTier           *int       `gorm:"column:peak_tier" json:"peak_tier,omitempty"`
	PeakTierPatched    *string    `gorm:"column:peak_tier_patched;type:varchar(32)" json:"peak_tier_patched,omitempty"`
	ValorantUpdatedAt  *time.Time `gorm:"column:valorant_updated_at" json:"valorant_updated_at,omitempty"`
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

// UpdateValorantInfo updates all Valorant-related fields
func (u *User) UpdateValorantInfo(riotName, riotTag, region string, currentTier int, currentTierPatched string, elo, rankingInTier int, peakTier int, peakTierPatched string) {
	u.RiotName = &riotName
	u.RiotTag = &riotTag
	u.Region = &region
	u.CurrentTier = &currentTier
	u.CurrentTierPatched = &currentTierPatched
	u.Elo = &elo
	u.RankingInTier = &rankingInTier
	u.PeakTier = &peakTier
	u.PeakTierPatched = &peakTierPatched
	now := time.Now()
	u.ValorantUpdatedAt = &now
}

// HasValorantLinked checks if the user has a Valorant account linked
func (u *User) HasValorantLinked() bool {
	return u.RiotName != nil && u.RiotTag != nil && *u.RiotName != "" && *u.RiotTag != ""
}

// ClearValorantInfo removes all Valorant-related information
func (u *User) ClearValorantInfo() {
	u.RiotName = nil
	u.RiotTag = nil
	u.Region = nil
	u.CurrentTier = nil
	u.CurrentTierPatched = nil
	u.Elo = nil
	u.RankingInTier = nil
	u.PeakTier = nil
	u.PeakTierPatched = nil
	u.ValorantUpdatedAt = nil
}

// IsValorantRefreshNeeded checks if 24 hours have passed since the last Valorant update
func (u *User) IsValorantRefreshNeeded() bool {
	if u.ValorantUpdatedAt == nil {
		return true
	}
	return time.Since(*u.ValorantUpdatedAt) > 24*time.Hour
}

// GetCurrentTierName extracts the base tier name from CurrentTierPatched (e.g., "Diamond 1" -> "Diamond")
func (u *User) GetCurrentTierName() string {
	if u.CurrentTierPatched == nil || *u.CurrentTierPatched == "" {
		return ""
	}
	return extractBaseTierName(*u.CurrentTierPatched)
}

// GetPeakTierName extracts the base tier name from PeakTierPatched
func (u *User) GetPeakTierName() string {
	if u.PeakTierPatched == nil || *u.PeakTierPatched == "" {
		return ""
	}
	return extractBaseTierName(*u.PeakTierPatched)
}

// extractBaseTierName extracts the tier name without the number (e.g., "Diamond 1" -> "Diamond")
func extractBaseTierName(tierPatched string) string {
	parts := strings.Split(tierPatched, " ")
	if len(parts) > 0 {
		return parts[0]
	}
	return tierPatched
}
