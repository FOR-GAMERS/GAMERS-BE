package port

import (
	"GAMERS-BE/internal/comment/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"time"
)

type CommentDatabasePort interface {
	Save(comment *domain.Comment) (*domain.Comment, error)
	GetByID(commentID int64) (*domain.Comment, error)
	GetByContestID(contestID int64, pagination *commonDto.PaginationRequest, sort *commonDto.SortRequest) ([]*CommentWithUser, int64, error)
	Update(comment *domain.Comment) error
	Delete(commentID int64) error
}

type CommentWithUser struct {
	CommentID  int64
	ContestID  int64
	UserID     int64
	Content    string
	CreatedAt  time.Time
	ModifiedAt time.Time
	Username   string
	Tag        string
	Avatar     *string
}
