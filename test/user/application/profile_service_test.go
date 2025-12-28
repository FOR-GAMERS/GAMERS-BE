package application_test

import (
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/domain"
	"errors"
	"testing"
	"time"
)

func setupProfileService() (*application.ProfileService, *mockProfileQueryPort, *mockProfileCommandPort) {
	queryPort := newMockProfileQueryPort()
	commandPort := newMockProfileCommandPort(queryPort)
	service := application.NewProfileService(queryPort, commandPort)
	return service, queryPort, commandPort
}

func TestProfileService_CreateProfile(t *testing.T) {
	service, _, _ := setupProfileService()

	req := dto.CreateProfileRequest{
		UserId:   1,
		Username: "testuser",
		Tag:      "1234",
		Bio:      "Test bio",
	}

	resp, err := service.CreateProfile(req)
	if err != nil {
		t.Errorf("CreateProfile() error = %v", err)
	}

	if resp.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, resp.Username)
	}
	if resp.Tag != req.Tag {
		t.Errorf("Expected tag %s, got %s", req.Tag, resp.Tag)
	}
	if resp.Bio != req.Bio {
		t.Errorf("Expected bio %s, got %s", req.Bio, resp.Bio)
	}
	if resp.Id == 0 {
		t.Error("Expected profile ID to be assigned")
	}
}

func TestProfileService_CreateProfile_InvalidUsername(t *testing.T) {
	service, _, _ := setupProfileService()

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "Empty username",
			username: "",
			wantErr:  true,
		},
		{
			name:     "Username too long",
			username: "thisusernameiswaytoolong",
			wantErr:  true,
		},
		{
			name:     "Username with special characters",
			username: "user@name",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.CreateProfileRequest{
				UserId:   1,
				Username: tt.username,
				Tag:      "1234",
				Bio:      "Test bio",
			}

			_, err := service.CreateProfile(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfileService_CreateProfile_InvalidTag(t *testing.T) {
	service, _, _ := setupProfileService()

	tests := []struct {
		name    string
		tag     string
		wantErr bool
	}{
		{
			name:    "Empty tag",
			tag:     "",
			wantErr: true,
		},
		{
			name:    "Tag too long",
			tag:     "123456",
			wantErr: true,
		},
		{
			name:    "Tag with special characters",
			tag:     "12@3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.CreateProfileRequest{
				UserId:   1,
				Username: "testuser",
				Tag:      tt.tag,
				Bio:      "Test bio",
			}

			_, err := service.CreateProfile(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfileService_GetProfile(t *testing.T) {
	service, queryPort, commandPort := setupProfileService()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	resp, err := service.GetProfile(profile.Id)
	if err != nil {
		t.Errorf("GetProfile() error = %v", err)
	}

	if resp.Username != profile.Username {
		t.Errorf("Expected username %s, got %s", profile.Username, resp.Username)
	}
	if resp.Tag != profile.Tag {
		t.Errorf("Expected tag %s, got %s", profile.Tag, resp.Tag)
	}

	_ = queryPort
}

func TestProfileService_GetProfile_NotFound(t *testing.T) {
	service, _, _ := setupProfileService()

	_, err := service.GetProfile(999)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileService_UpdateProfile(t *testing.T) {
	service, queryPort, commandPort := setupProfileService()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	updateReq := dto.UpdateProfileRequest{
		Username: "updateduser",
		Tag:      "5678",
		Bio:      "Updated bio",
		Avatar:   "new_avatar.jpg",
	}

	resp, err := service.UpdateProfile(profile.Id, updateReq)
	if err != nil {
		t.Errorf("UpdateProfile() error = %v", err)
	}

	if resp.Username != updateReq.Username {
		t.Errorf("Expected username %s, got %s", updateReq.Username, resp.Username)
	}
	if resp.Tag != updateReq.Tag {
		t.Errorf("Expected tag %s, got %s", updateReq.Tag, resp.Tag)
	}
	if resp.Bio != updateReq.Bio {
		t.Errorf("Expected bio %s, got %s", updateReq.Bio, resp.Bio)
	}
	if resp.Avatar != updateReq.Avatar {
		t.Errorf("Expected avatar %s, got %s", updateReq.Avatar, resp.Avatar)
	}

	_ = queryPort
}

func TestProfileService_UpdateProfile_NotFound(t *testing.T) {
	service, _, _ := setupProfileService()

	updateReq := dto.UpdateProfileRequest{
		Username: "updateduser",
		Tag:      "5678",
		Bio:      "Updated bio",
		Avatar:   "new_avatar.jpg",
	}

	_, err := service.UpdateProfile(999, updateReq)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileService_UpdateProfile_InvalidData(t *testing.T) {
	service, queryPort, commandPort := setupProfileService()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	tests := []struct {
		name    string
		req     dto.UpdateProfileRequest
		wantErr bool
	}{
		{
			name: "Invalid username",
			req: dto.UpdateProfileRequest{
				Username: "",
				Tag:      "5678",
				Bio:      "Updated bio",
				Avatar:   "avatar.jpg",
			},
			wantErr: true,
		},
		{
			name: "Invalid tag",
			req: dto.UpdateProfileRequest{
				Username: "updateduser",
				Tag:      "",
				Bio:      "Updated bio",
				Avatar:   "avatar.jpg",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.UpdateProfile(profile.Id, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	_ = queryPort
}

func TestProfileService_DeleteProfile(t *testing.T) {
	service, queryPort, commandPort := setupProfileService()

	profile := &domain.Profile{
		UserId:    1,
		Username:  "testuser",
		Tag:       "1234",
		Bio:       "Test bio",
		Avatar:    "avatar.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := commandPort.Save(profile)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	err = service.DeleteProfile(profile.Id)
	if err != nil {
		t.Errorf("DeleteProfile() error = %v", err)
	}

	_, err = queryPort.FindById(profile.Id)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Error("Expected profile to be deleted")
	}
}

func TestProfileService_DeleteProfile_NotFound(t *testing.T) {
	service, _, _ := setupProfileService()

	err := service.DeleteProfile(999)
	if !errors.Is(err, domain.ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, got %v", err)
	}
}
