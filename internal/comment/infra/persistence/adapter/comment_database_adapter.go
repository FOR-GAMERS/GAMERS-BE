package adapter

import (
	"GAMERS-BE/internal/comment/application/port"
	"GAMERS-BE/internal/comment/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/exception"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type CommentDatabaseAdapter struct {
	db *gorm.DB
}

func NewCommentDatabaseAdapter(db *gorm.DB) *CommentDatabaseAdapter {
	return &CommentDatabaseAdapter{db: db}
}

func (a *CommentDatabaseAdapter) Save(comment *domain.Comment) (*domain.Comment, error) {
	if err := a.db.Create(comment).Error; err != nil {
		return nil, a.translateError(err)
	}
	return comment, nil
}

func (a *CommentDatabaseAdapter) GetByID(commentID int64) (*domain.Comment, error) {
	var comment domain.Comment
	result := a.db.Where("comment_id = ?", commentID).First(&comment)
	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}
	return &comment, nil
}

func (a *CommentDatabaseAdapter) GetByContestID(
	contestID int64,
	pagination *commonDto.PaginationRequest,
	sort *commonDto.SortRequest,
) ([]*port.CommentWithUser, int64, error) {
	var totalCount int64

	countResult := a.db.Model(&domain.Comment{}).
		Where("contest_id = ?", contestID).
		Count(&totalCount)
	if countResult.Error != nil {
		return nil, 0, a.translateError(countResult.Error)
	}

	orderClause := "cc.created_at DESC"
	if sort != nil {
		allowedSortFields := map[string]string{
			"created_at":  "cc.created_at",
			"modified_at": "cc.modified_at",
		}
		if field, ok := allowedSortFields[sort.SortBy]; ok {
			order := "DESC"
			if strings.ToUpper(sort.Order) == "ASC" {
				order = "ASC"
			}
			orderClause = field + " " + order
		}
	}

	var results []*port.CommentWithUser
	query := a.db.Table("contest_comments cc").
		Select("cc.comment_id, cc.contest_id, cc.user_id, cc.content, cc.created_at, cc.modified_at, u.username, u.tag, u.avatar, da.discord_id, da.discord_avatar").
		Joins("JOIN users u ON cc.user_id = u.id").
		Joins("LEFT JOIN discord_accounts da ON u.id = da.user_id").
		Where("cc.contest_id = ?", contestID).
		Order(orderClause)

	if pagination != nil {
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, a.translateError(err)
	}

	return results, totalCount, nil
}

func (a *CommentDatabaseAdapter) Update(comment *domain.Comment) error {
	result := a.db.Model(&domain.Comment{}).
		Where("comment_id = ?", comment.CommentID).
		Updates(map[string]interface{}{
			"content": comment.Content,
		})

	if result.Error != nil {
		return a.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrCommentNotFound
	}

	return nil
}

func (a *CommentDatabaseAdapter) Delete(commentID int64) error {
	result := a.db.Where("comment_id = ?", commentID).Delete(&domain.Comment{})
	if result.Error != nil {
		return a.translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return exception.ErrCommentNotFound
	}
	return nil
}

func (a *CommentDatabaseAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrCommentNotFound
	}

	if a.isForeignKeyError(err) {
		return exception.ErrContestNotFound
	}

	if a.isConnectionError(err) {
		return exception.ErrDBConnection
	}

	return err
}

func (a *CommentDatabaseAdapter) isForeignKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "foreign key constraint") ||
		strings.Contains(errMsg, "1452") ||
		strings.Contains(errMsg, "23503")
}

func (a *CommentDatabaseAdapter) isConnectionError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "refused")
}
