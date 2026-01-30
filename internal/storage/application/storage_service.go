package application

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/storage/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/storage/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/storage/domain"
	"context"
	"fmt"
	"mime/multipart"

	"github.com/google/uuid"
)

type StorageService struct {
	storagePort port.StoragePort
}

func NewStorageService(storagePort port.StoragePort) *StorageService {
	return &StorageService{
		storagePort: storagePort,
	}
}

func (s *StorageService) UploadContestBanner(ctx context.Context, contestId int64, file *multipart.FileHeader) (*dto.UploadResponse, error) {
	if err := domain.ValidateFile(file, domain.UploadTypeContestBanner); err != nil {
		return nil, err
	}

	mimeType := file.Header.Get("Content-Type")
	ext := domain.GetExtensionFromMimeType(mimeType)
	key := generateKey(domain.UploadTypeContestBanner, contestId, ext)

	src, err := file.Open()
	if err != nil {
		return nil, exception.ErrStorageUploadFailed
	}
	defer src.Close()

	if err := s.storagePort.Upload(ctx, key, src, file.Size, mimeType); err != nil {
		return nil, exception.ErrStorageUploadFailed
	}

	url := s.storagePort.GetPublicURL(key)
	uploadedFile := domain.NewUploadedFile(key, url, mimeType, file.Size)

	return dto.ToUploadResponse(uploadedFile), nil
}

func (s *StorageService) UploadUserProfile(ctx context.Context, userId int64, file *multipart.FileHeader) (*dto.UploadResponse, error) {
	if err := domain.ValidateFile(file, domain.UploadTypeUserProfile); err != nil {
		return nil, err
	}

	mimeType := file.Header.Get("Content-Type")
	ext := domain.GetExtensionFromMimeType(mimeType)
	key := generateKey(domain.UploadTypeUserProfile, userId, ext)

	src, err := file.Open()
	if err != nil {
		return nil, exception.ErrStorageUploadFailed
	}
	defer src.Close()

	if err := s.storagePort.Upload(ctx, key, src, file.Size, mimeType); err != nil {
		return nil, exception.ErrStorageUploadFailed
	}

	url := s.storagePort.GetPublicURL(key)
	uploadedFile := domain.NewUploadedFile(key, url, mimeType, file.Size)

	return dto.ToUploadResponse(uploadedFile), nil
}

func (s *StorageService) UploadMainBanner(ctx context.Context, bannerId int64, file *multipart.FileHeader) (*dto.UploadResponse, error) {
	if err := domain.ValidateFile(file, domain.UploadTypeMainBanner); err != nil {
		return nil, err
	}

	mimeType := file.Header.Get("Content-Type")
	ext := domain.GetExtensionFromMimeType(mimeType)
	key := generateKey(domain.UploadTypeMainBanner, bannerId, ext)

	src, err := file.Open()
	if err != nil {
		return nil, exception.ErrStorageUploadFailed
	}
	defer src.Close()

	if err := s.storagePort.Upload(ctx, key, src, file.Size, mimeType); err != nil {
		return nil, exception.ErrStorageUploadFailed
	}

	url := s.storagePort.GetPublicURL(key)
	uploadedFile := domain.NewUploadedFile(key, url, mimeType, file.Size)

	return dto.ToUploadResponse(uploadedFile), nil
}

func (s *StorageService) DeleteFile(ctx context.Context, key string) error {
	if err := s.storagePort.Delete(ctx, key); err != nil {
		return exception.ErrStorageDeleteFailed
	}
	return nil
}

func generateKey(uploadType domain.UploadType, id int64, ext string) string {
	uniqueId := uuid.New().String()
	return fmt.Sprintf("%s/%d/%s%s", uploadType, id, uniqueId, ext)
}
