package domain_test

import (
	"GAMERS-BE/internal/user/domain"
	"errors"
	"strings"
	"testing"
)

func TestNewProfile_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		userId   int64
		username string
		tag      string
		bio      string
		avatar   string
		wantErr  bool
	}{
		{
			name:     "Valid profile with all fields",
			userId:   1,
			username: "testuser",
			tag:      "1234",
			bio:      "This is a test bio",
			avatar:   "https://example.com/avatar.png",
			wantErr:  false,
		},
		{
			name:     "Valid profile with empty bio",
			userId:   2,
			username: "gamer123",
			tag:      "abc",
			bio:      "",
			avatar:   "",
			wantErr:  false,
		},
		{
			name:     "Valid profile with max username length",
			userId:   3,
			username: "sixteencharacter",
			tag:      "99999",
			bio:      "Short bio",
			avatar:   "avatar.jpg",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := domain.NewProfile(tt.userId, tt.username, tt.tag, tt.bio, tt.avatar)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if profile.UserId != tt.userId {
					t.Errorf("Expected userId %d, got %d", tt.userId, profile.UserId)
				}
				if profile.Username != tt.username {
					t.Errorf("Expected username %s, got %s", tt.username, profile.Username)
				}
				if profile.Tag != tt.tag {
					t.Errorf("Expected tag %s, got %s", tt.tag, profile.Tag)
				}
				if profile.Bio != tt.bio {
					t.Errorf("Expected bio %s, got %s", tt.bio, profile.Bio)
				}
				if profile.Avatar != tt.avatar {
					t.Errorf("Expected avatar %s, got %s", tt.avatar, profile.Avatar)
				}
			}
		})
	}
}

func TestNewProfile_InvalidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "Empty username",
			username: "",
			wantErr:  domain.ErrUsernameEmpty,
		},
		{
			name:     "Username too long",
			username: "thisusernameiswaytoolong",
			wantErr:  domain.ErrUsernameTooLong,
		},
		{
			name:     "Username with special characters",
			username: "user@name",
			wantErr:  domain.ErrUsernameInvalidChar,
		},
		{
			name:     "Username with spaces",
			username: "user name",
			wantErr:  domain.ErrUsernameInvalidChar,
		},
		{
			name:     "Username with underscore",
			username: "user_name",
			wantErr:  domain.ErrUsernameInvalidChar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewProfile(1, tt.username, "1234", "bio", "avatar")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNewProfile_InvalidTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantErr error
	}{
		{
			name:    "Empty tag",
			tag:     "",
			wantErr: domain.ErrTagEmpty,
		},
		{
			name:    "Tag too long",
			tag:     "123456",
			wantErr: domain.ErrTagTooLong,
		},
		{
			name:    "Tag with special characters",
			tag:     "12@3",
			wantErr: domain.ErrTagInvalidChar,
		},
		{
			name:    "Tag with spaces",
			tag:     "1 23",
			wantErr: domain.ErrTagInvalidChar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewProfile(1, "testuser", tt.tag, "bio", "avatar")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNewProfile_InvalidBio(t *testing.T) {
	longBio := strings.Repeat("a", 256)

	tests := []struct {
		name    string
		bio     string
		wantErr error
	}{
		{
			name:    "Bio too long (256 characters)",
			bio:     longBio,
			wantErr: domain.ErrBioTooLong,
		},
		{
			name:    "Bio too long (300 characters)",
			bio:     strings.Repeat("b", 300),
			wantErr: domain.ErrBioTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewProfile(1, "testuser", "1234", tt.bio, "avatar")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNewProfile_ValidBio(t *testing.T) {
	validBio := strings.Repeat("a", 255)

	profile, err := domain.NewProfile(1, "testuser", "1234", validBio, "avatar")
	if err != nil {
		t.Errorf("Expected no error for 255 character bio, got %v", err)
	}
	if profile.Bio != validBio {
		t.Errorf("Bio not set correctly")
	}
}

func TestUpdateProfile_ValidInput(t *testing.T) {
	profile, err := domain.NewProfile(1, "olduser", "1234", "old bio", "old.jpg")
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	newUsername := "newuser"
	newTag := "5678"
	newBio := "new bio"
	newAvatar := "new.jpg"

	updatedProfile, err := profile.UpdateProfile(newUsername, newTag, newBio, newAvatar)
	if err != nil {
		t.Errorf("UpdateProfile() error = %v, expected nil", err)
	}

	if updatedProfile.Username != newUsername {
		t.Errorf("Expected username %s, got %s", newUsername, updatedProfile.Username)
	}
	if updatedProfile.Tag != newTag {
		t.Errorf("Expected tag %s, got %s", newTag, updatedProfile.Tag)
	}
	if updatedProfile.Bio != newBio {
		t.Errorf("Expected bio %s, got %s", newBio, updatedProfile.Bio)
	}
	if updatedProfile.Avatar != newAvatar {
		t.Errorf("Expected avatar %s, got %s", newAvatar, updatedProfile.Avatar)
	}
	if updatedProfile.UserId != 1 {
		t.Errorf("UserId should not change, got %d", updatedProfile.UserId)
	}
}

func TestUpdateProfile_InvalidUsername(t *testing.T) {
	tests := []struct {
		name        string
		newUsername string
		wantErr     error
	}{
		{
			name:        "Empty username",
			newUsername: "",
			wantErr:     domain.ErrUsernameEmpty,
		},
		{
			name:        "Username too long",
			newUsername: "thisusernameiswaytoolong",
			wantErr:     domain.ErrUsernameTooLong,
		},
		{
			name:        "Username with special characters",
			newUsername: "user@name",
			wantErr:     domain.ErrUsernameInvalidChar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := domain.NewProfile(1, "olduser", "1234", "bio", "avatar")
			if err != nil {
				t.Fatalf("Failed to create profile: %v", err)
			}

			_, err = profile.UpdateProfile(tt.newUsername, "5678", "new bio", "new.jpg")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUpdateProfile_InvalidTag(t *testing.T) {
	tests := []struct {
		name    string
		newTag  string
		wantErr error
	}{
		{
			name:    "Empty tag",
			newTag:  "",
			wantErr: domain.ErrTagEmpty,
		},
		{
			name:    "Tag too long",
			newTag:  "123456",
			wantErr: domain.ErrTagTooLong,
		},
		{
			name:    "Tag with special characters",
			newTag:  "12@3",
			wantErr: domain.ErrTagInvalidChar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := domain.NewProfile(1, "olduser", "1234", "bio", "avatar")
			if err != nil {
				t.Fatalf("Failed to create profile: %v", err)
			}

			_, err = profile.UpdateProfile("newuser", tt.newTag, "new bio", "new.jpg")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUpdateProfile_InvalidBio(t *testing.T) {
	profile, err := domain.NewProfile(1, "olduser", "1234", "bio", "avatar")
	if err != nil {
		t.Fatalf("Failed to create profile: %v", err)
	}

	longBio := strings.Repeat("a", 256)
	_, err = profile.UpdateProfile("newuser", "5678", longBio, "new.jpg")
	if !errors.Is(err, domain.ErrBioTooLong) {
		t.Errorf("Expected error %v, got %v", domain.ErrBioTooLong, err)
	}
}
