package exception

import (
	"net/http"
)

var (
	ErrStorageUploadFailed    = NewBusinessError(http.StatusInternalServerError, "failed to upload file to storage", "ST001")
	ErrStorageDeleteFailed    = NewBusinessError(http.StatusInternalServerError, "failed to delete file from storage", "ST002")
	ErrInvalidFileType        = NewBadRequestError("invalid file type. allowed: jpeg, png, webp", "ST003")
	ErrFileTooLarge           = NewBadRequestError("file size exceeds maximum limit", "ST004")
	ErrFileRequired           = NewBadRequestError("file is required", "ST005")
	ErrInvalidContestForBanner = NewBadRequestError("invalid contest id for banner upload", "ST006")
	ErrStorageInitFailed      = NewBusinessError(http.StatusInternalServerError, "failed to initialize storage client", "ST007")
)
