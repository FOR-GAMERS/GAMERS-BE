package domain

import (
	"GAMERS-BE/internal/global/exception"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

type UploadType string

const (
	UploadTypeContestBanner UploadType = "contest-banners"
	UploadTypeUserProfile   UploadType = "user-profiles"
)

const (
	MaxContestBannerSize = 5 * 1024 * 1024  // 5MB
	MaxUserProfileSize   = 2 * 1024 * 1024  // 2MB
)

var AllowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var MimeToExtension = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type UploadedFile struct {
	Key        string    `json:"key"`
	URL        string    `json:"url"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func NewUploadedFile(key, url, mimeType string, size int64) *UploadedFile {
	return &UploadedFile{
		Key:        key,
		URL:        url,
		Size:       size,
		MimeType:   mimeType,
		UploadedAt: time.Now(),
	}
}

func ValidateFile(file *multipart.FileHeader, uploadType UploadType) error {
	if file == nil {
		return exception.ErrFileRequired
	}

	maxSize := getMaxSize(uploadType)
	if file.Size > maxSize {
		return exception.ErrFileTooLarge
	}

	mimeType := file.Header.Get("Content-Type")
	if !IsAllowedMimeType(mimeType) {
		return exception.ErrInvalidFileType
	}

	return nil
}

func getMaxSize(uploadType UploadType) int64 {
	switch uploadType {
	case UploadTypeContestBanner:
		return MaxContestBannerSize
	case UploadTypeUserProfile:
		return MaxUserProfileSize
	default:
		return MaxUserProfileSize
	}
}

func IsAllowedMimeType(mimeType string) bool {
	return AllowedMimeTypes[mimeType]
}

func GetExtensionFromMimeType(mimeType string) string {
	if ext, ok := MimeToExtension[mimeType]; ok {
		return ext
	}
	return ".jpg"
}

func GetMimeTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
