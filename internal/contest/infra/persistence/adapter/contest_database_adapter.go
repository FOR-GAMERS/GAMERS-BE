package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"errors"
	"strings"

	"gorm.io/gorm"
)

func NewContestDatabaseAdapter(db *gorm.DB) *ContestDatabaseAdapter {
	return &ContestDatabaseAdapter{
		db: db,
	}
}

type ContestDatabaseAdapter struct {
	db *gorm.DB
}

func (c ContestDatabaseAdapter) Save(contest *domain.Contest) (*domain.Contest, error) {
	err := c.db.Save(contest).Error

	if err != nil {
		return nil, c.translateError(err)
	}

	return contest, nil
}

func (c ContestDatabaseAdapter) GetContestById(contestId int64) (*domain.Contest, error) {
	var contest domain.Contest

	result := c.db.Where("contest_id = ?", contestId).First(&contest)

	if result.Error != nil {
		return nil, c.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, exception.ErrContestNotFound
	}

	return &contest, nil
}

func (c ContestDatabaseAdapter) GetContests(offset, limit int, sortReq *dto.SortRequest, title *string) ([]domain.Contest, int64, error) {
	var contests []domain.Contest
	var totalCount int64

	query := c.db.Model(&domain.Contest{})

	// Apply title search filter if provided
	if title != nil && *title != "" {
		query = query.Where("title LIKE ?", "%"+*title+"%")
	}

	// Get total count
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, c.translateError(err)
	}

	// Get paginated data with dynamic sorting
	orderClause := "created_at DESC" // default
	if sortReq != nil {
		orderClause = sortReq.GetOrderClause()
	}

	if err := query.Order(orderClause).
		Offset(offset).
		Limit(limit).
		Find(&contests).Error; err != nil {
		return nil, 0, c.translateError(err)
	}

	return contests, totalCount, nil
}

func (c ContestDatabaseAdapter) DeleteContestById(contestId int64) error {
	result := c.db.Where("contest_id = ?", contestId).Delete(&domain.Contest{})
	if result.Error != nil {
		return c.translateError(result.Error)
	}

	return nil
}

func (c ContestDatabaseAdapter) UpdateContest(contest *domain.Contest) error {
	return c.db.Save(contest).Error
}

func (c ContestDatabaseAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrContestNotFound
	}

	if isDuplicateKeyError(err) {
		return exception.ErrContestAlreadyExists
	}

	if isConnectionError(err) {
		return exception.ErrDBConnection
	}

	return err
}

func isDuplicateKeyError(err error) bool {
	errMsg := err.Error()

	if strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "1062") {
		return true
	}
	if strings.Contains(errMsg, "duplicate key value") ||
		strings.Contains(errMsg, "23505") {
		return true
	}
	return false
}

func isConnectionError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "refused")
}
