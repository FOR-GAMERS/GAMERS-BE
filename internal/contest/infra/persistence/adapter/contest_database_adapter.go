package adapter

import (
	"GAMERS-BE/internal/contest/domain"
	"GAMERS-BE/internal/global/exception"
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

	result := c.db.First(&contest, contestId)

	if result.Error != nil {
		return nil, c.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, exception.ErrContestNotFound
	}

	return &contest, nil
}

func (c ContestDatabaseAdapter) GetContests() ([]domain.Contest, error) {
	//TODO implement me
	panic("implement me")
}

func (c ContestDatabaseAdapter) DeleteContestById(contestId int64) error {
	result := c.db.Delete(&domain.Contest{}, contestId)
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
