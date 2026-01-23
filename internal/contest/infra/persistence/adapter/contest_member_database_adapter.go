package adapter

import (
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
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

func (c ContestMemberDatabaseAdapter) GetMembersWithUserByContest(
	contestId int64,
	pagination *commonDto.PaginationRequest,
	sort *commonDto.SortRequest,
) ([]*port.ContestMemberWithUser, int64, error) {
	var totalCount int64

	// Count total members
	countResult := c.db.Model(&domain.ContestMember{}).
		Where("contest_id = ?", contestId).
		Count(&totalCount)
	if countResult.Error != nil {
		return nil, 0, c.translateError(countResult.Error)
	}

	// Build order clause with validation
	orderClause := "cm.point DESC" // Default order
	if sort != nil {
		allowedSortFields := map[string]string{
			"point":    "cm.point",
			"username": "u.username",
		}
		if field, ok := allowedSortFields[sort.SortBy]; ok {
			order := "DESC"
			if strings.ToUpper(sort.Order) == "ASC" {
				order = "ASC"
			}
			orderClause = field + " " + order
		}
	}

	// Query with JOIN
	var results []*port.ContestMemberWithUser
	query := c.db.Table("contests_members cm").
		Select("cm.user_id, cm.contest_id, cm.member_type, cm.leader_type, cm.point, u.username, u.tag, u.avatar").
		Joins("JOIN users u ON cm.user_id = u.id").
		Where("cm.contest_id = ?", contestId).
		Order(orderClause)

	// Apply pagination
	if pagination != nil {
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, c.translateError(err)
	}

	return results, totalCount, nil
}

func (c ContestMemberDatabaseAdapter) UpdateMemberType(contestId, userId int64, memberType domain.MemberType) error {
	result := c.db.Model(&domain.ContestMember{}).
		Where("contest_id = ? AND user_id = ?", contestId, userId).
		Update("member_type", memberType)

	if result.Error != nil {
		return c.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrContestMemberNotFound
	}

	return nil
}

func (c ContestMemberDatabaseAdapter) GetContestsByUserId(
	userId int64,
	pagination *commonDto.PaginationRequest,
	sort *commonDto.SortRequest,
	status *domain.ContestStatus,
) ([]*port.ContestWithMembership, int64, error) {
	var totalCount int64

	// Build base count query
	countQuery := c.db.Table("contests_members cm").
		Joins("JOIN contests c ON cm.contest_id = c.contest_id").
		Where("cm.user_id = ?", userId)

	// Apply status filter to count query
	if status != nil {
		countQuery = countQuery.Where("c.contest_status = ?", *status)
	}

	countResult := countQuery.Count(&totalCount)
	if countResult.Error != nil {
		return nil, 0, c.translateError(countResult.Error)
	}

	// Build order clause with validation
	orderClause := "c.created_at DESC" // Default order
	if sort != nil {
		allowedSortFields := map[string]string{
			"created_at":     "c.created_at",
			"started_at":     "c.started_at",
			"ended_at":       "c.ended_at",
			"point":          "cm.point",
			"contest_status": "c.contest_status",
		}
		if field, ok := allowedSortFields[sort.SortBy]; ok {
			order := "DESC"
			if strings.ToUpper(sort.Order) == "ASC" {
				order = "ASC"
			}
			orderClause = field + " " + order
		}
	}

	// Query with JOIN to get contest info with membership
	var results []*port.ContestWithMembership
	query := c.db.Table("contests_members cm").
		Select(`
			c.contest_id, c.title, c.description, c.max_team_count, c.total_point,
			c.contest_type, c.contest_status, c.started_at, c.ended_at, c.auto_start,
			c.game_type, c.game_point_table_id, c.total_team_member,
			c.discord_guild_id, c.discord_text_channel_id, c.thumbnail,
			c.created_at, c.modified_at,
			cm.member_type, cm.leader_type, cm.point
		`).
		Joins("JOIN contests c ON cm.contest_id = c.contest_id").
		Where("cm.user_id = ?", userId).
		Order(orderClause)

	// Apply status filter
	if status != nil {
		query = query.Where("c.contest_status = ?", *status)
	}

	// Apply pagination
	if pagination != nil {
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, c.translateError(err)
	}

	return results, totalCount, nil
}
