package dto

import (
	"GAMERS-BE/internal/comment/application/port"
	"GAMERS-BE/internal/comment/domain"
	"GAMERS-BE/internal/global/utils"
	"time"
)

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,max=255"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,max=255"`
}

type AuthorResponse struct {
	UserID   int64   `json:"user_id"`
	Username string  `json:"username"`
	Tag      string  `json:"tag"`
	Avatar   *string `json:"avatar,omitempty"`
}

type CommentResponse struct {
	CommentID  int64          `json:"comment_id"`
	ContestID  int64          `json:"contest_id"`
	Content    string         `json:"content"`
	CreatedAt  time.Time      `json:"created_at"`
	ModifiedAt time.Time      `json:"modified_at"`
	Author     AuthorResponse `json:"author"`
}

func ToCommentResponse(comment *domain.Comment, author *AuthorResponse) *CommentResponse {
	return &CommentResponse{
		CommentID:  comment.CommentID,
		ContestID:  comment.ContestID,
		Content:    comment.Content,
		CreatedAt:  comment.CreatedAt,
		ModifiedAt: comment.ModifiedAt,
		Author:     *author,
	}
}

func ToCommentResponseFromWithUser(c *port.CommentWithUser) *CommentResponse {
	// Build Discord avatar URL if Discord account exists
	avatar := c.Avatar
	if c.DiscordId != nil && c.DiscordAvatar != nil {
		if url := utils.BuildDiscordAvatarURL(*c.DiscordId, *c.DiscordAvatar); url != "" {
			avatar = &url
		}
	}

	return &CommentResponse{
		CommentID:  c.CommentID,
		ContestID:  c.ContestID,
		Content:    c.Content,
		CreatedAt:  c.CreatedAt,
		ModifiedAt: c.ModifiedAt,
		Author: AuthorResponse{
			UserID:   c.UserID,
			Username: c.Username,
			Tag:      c.Tag,
			Avatar:   avatar,
		},
	}
}

func ToCommentResponses(comments []*port.CommentWithUser) []*CommentResponse {
	responses := make([]*CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = ToCommentResponseFromWithUser(comment)
	}
	return responses
}
