package dto

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/storage/domain"
	"time"
)

type UploadResponse struct {
	Key        string    `json:"key"`
	URL        string    `json:"url"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func ToUploadResponse(file *domain.UploadedFile) *UploadResponse {
	return &UploadResponse{
		Key:        file.Key,
		URL:        file.URL,
		Size:       file.Size,
		MimeType:   file.MimeType,
		UploadedAt: file.UploadedAt,
	}
}

type ContestBannerUploadRequest struct {
	ContestID int64 `form:"contest_id" binding:"required"`
}

type UserProfileUploadRequest struct {
}
