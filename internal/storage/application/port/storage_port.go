package port

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/storage/domain"
	"context"
	"io"
)

type StoragePort interface {
	Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	GetPublicURL(key string) string
}

type ContestRepositoryPort interface {
	UpdateBannerKey(contestId int64, bannerKey string) error
	GetContestOwnerId(contestId int64) (int64, error)
}

type UserRepositoryPort interface {
	UpdateProfileKey(userId int64, profileKey string) error
}

type UploadResult struct {
	File *domain.UploadedFile
}
