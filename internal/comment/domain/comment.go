package domain

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"time"
)

const (
	MaxContentLength = 255
)

type Comment struct {
	CommentID  int64     `gorm:"column:comment_id;primaryKey;autoIncrement" json:"comment_id"`
	ContestID  int64     `gorm:"column:contest_id;type:bigint;not null" json:"contest_id"`
	UserID     int64     `gorm:"column:user_id;type:bigint;not null" json:"user_id"`
	Content    string    `gorm:"column:content;type:varchar(255);not null" json:"content"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;autoCreateTime" json:"created_at"`
	ModifiedAt time.Time `gorm:"column:modified_at;type:datetime;autoUpdateTime" json:"modified_at"`
}

func (c *Comment) TableName() string {
	return "contest_comments"
}

func NewComment(contestID, userID int64, content string) *Comment {
	return &Comment{
		ContestID: contestID,
		UserID:    userID,
		Content:   content,
	}
}

func (c *Comment) Validate() error {
	if c.Content == "" {
		return exception.ErrCommentContentEmpty
	}
	if len(c.Content) > MaxContentLength {
		return exception.ErrCommentContentTooLong
	}
	if c.ContestID <= 0 {
		return exception.ErrInvalidContestID
	}
	if c.UserID <= 0 {
		return exception.ErrInvalidUserID
	}
	return nil
}

func (c *Comment) UpdateContent(content string) error {
	if content == "" {
		return exception.ErrCommentContentEmpty
	}
	if len(content) > MaxContentLength {
		return exception.ErrCommentContentTooLong
	}
	c.Content = content
	return nil
}

func (c *Comment) IsOwner(userID int64) bool {
	return c.UserID == userID
}
