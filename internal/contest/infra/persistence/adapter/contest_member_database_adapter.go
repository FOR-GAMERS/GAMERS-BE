package adapter

import (
	"GAMERS-BE/internal/contest/domain"
	"GAMERS-BE/internal/global/exception"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type ContestMemberDatabaseAdapter struct {
	db *gorm.DB
}

func (c ContestMemberDatabaseAdapter) Save(member *domain.ContestMember) error {
	if err := member.Validate(); err != nil {
		return err
	}

	err := c.db.Save(member).Error
	if err != nil {
		return c.translateError(err)
	}

	return nil
}

func (c ContestMemberDatabaseAdapter) DeleteById(contestId, userId int64) error {
	result := c.db.Where("contest_id = ? AND user_id = ?", contestId, userId).Delete(&domain.ContestMember{})
	if result.Error != nil {
		return c.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrContestMemberNotFound
	}

	return nil
}

func (c ContestMemberDatabaseAdapter) GetByContestAndUser(contestId, userId int64) (*domain.ContestMember, error) {
	var member domain.ContestMember
	result := c.db.Where("contest_id = ? AND user_id = ?", contestId, userId).First(&member)

	if result.Error != nil {
		return nil, c.translateError(result.Error)
	}

	return &member, nil
}

func (c ContestMemberDatabaseAdapter) GetMembersByContest(contestId int64) ([]*domain.ContestMember, error) {
	var members []*domain.ContestMember
	result := c.db.Where("contest_id = ?", contestId).Find(&members)

	if result.Error != nil {
		return nil, c.translateError(result.Error)
	}

	return members, nil
}

func (c ContestMemberDatabaseAdapter) SaveBatch(members []*domain.ContestMember) error {
	if len(members) == 0 {
		return nil
	}

	// Transaction으로 일괄 저장
	err := c.db.Transaction(func(tx *gorm.DB) error {
		for _, member := range members {
			if err := member.Validate(); err != nil {
				return err
			}

			if err := tx.Save(member).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return c.translateError(err)
	}

	return nil
}

func NewContestMemberDatabaseAdapter(db *gorm.DB) *ContestMemberDatabaseAdapter {
	return &ContestMemberDatabaseAdapter{db: db}
}

func (c ContestMemberDatabaseAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrContestMemberNotFound
	}

	if isDuplicateKeyError(err) {
		return exception.ErrAlreadyContestMemberExists
	}

	if c.isForeignKeyError(err) {
		return exception.ErrContestNotFound
	}

	if isConnectionError(err) {
		return exception.ErrDBConnection
	}

	return err
}

func (c ContestMemberDatabaseAdapter) isForeignKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "foreign key constraint") || strings.Contains(errMsg, "1452") ||
		strings.Contains(errMsg, "23503")
}
