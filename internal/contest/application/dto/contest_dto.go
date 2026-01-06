package dto

import (
	"GAMERS-BE/internal/contest/domain"
	"errors"
	"time"
)

type CreateContestRequest struct {
	Title        string             `json:"title" binding:"required"`
	Description  string             `json:"description,omitempty"`
	MaxTeamCount int                `json:"max_team_count,omitempty"`
	TotalPoint   int                `json:"total_point,omitempty"`
	ContestType  domain.ContestType `json:"contest_type" binding:"required"`
	StartedAt    time.Time          `json:"started_at,omitempty"`
	EndedAt      time.Time          `json:"ended_at,omitempty"`
	AutoStart    bool               `json:"auto_start,omitempty"`
}

type UpdateContestRequest struct {
	Title         *string               `json:"title,omitempty"`
	Description   *string               `json:"description,omitempty"`
	MaxTeamCount  *int                  `json:"max_team_count,omitempty"`
	TotalPoint    *int                  `json:"total_point,omitempty"`
	ContestType   *domain.ContestType   `json:"contest_type,omitempty"`
	ContestStatus *domain.ContestStatus `json:"contest_status,omitempty"`
	StartedAt     *time.Time            `json:"started_at,omitempty"`
	EndedAt       *time.Time            `json:"ended_at,omitempty"`
	AutoStart     *bool                 `json:"auto_start,omitempty"`
}

type ContestResponse struct {
	ContestID     int64                `json:"contest_id"`
	Title         string               `json:"title"`
	Description   string               `json:"description,omitempty"`
	MaxTeamCount  int                  `json:"max_team_count,omitempty"`
	TotalPoint    int                  `json:"total_point"`
	ContestType   domain.ContestType   `json:"contest_type"`
	ContestStatus domain.ContestStatus `json:"contest_status"`
	StartedAt     time.Time            `json:"started_at,omitempty"`
	EndedAt       time.Time            `json:"ended_at,omitempty"`
	AutoStart     bool                 `json:"auto_start,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	ModifiedAt    time.Time            `json:"modified_at"`
}

func (req *UpdateContestRequest) ApplyTo(contest *domain.Contest) {
	if req.Title != nil {
		contest.Title = *req.Title
	}
	if req.Description != nil {
		contest.Description = *req.Description
	}
	if req.MaxTeamCount != nil {
		contest.MaxTeamCount = *req.MaxTeamCount
	}
	if req.TotalPoint != nil {
		contest.TotalPoint = *req.TotalPoint
	}
	if req.ContestType != nil {
		contest.ContestType = *req.ContestType
	}
	if req.ContestStatus != nil {
		contest.ContestStatus = *req.ContestStatus
	}
	if req.StartedAt != nil {
		contest.StartedAt = *req.StartedAt
	}
	if req.EndedAt != nil {
		contest.EndedAt = *req.EndedAt
	}
	if req.AutoStart != nil {
		contest.AutoStart = *req.AutoStart
	}
}

func (req *UpdateContestRequest) HasChanges() bool {
	return req.Title != nil ||
		req.Description != nil ||
		req.MaxTeamCount != nil ||
		req.TotalPoint != nil ||
		req.ContestType != nil ||
		req.ContestStatus != nil ||
		req.StartedAt != nil ||
		req.EndedAt != nil ||
		req.AutoStart != nil
}

func (req *UpdateContestRequest) Validate() error {
	if req.StartedAt != nil && req.EndedAt != nil {
		if req.EndedAt.Before(*req.StartedAt) {
			return errors.New("end time must be after start time")
		}
	}

	if req.MaxTeamCount != nil && *req.MaxTeamCount <= 0 {
		return errors.New("max team count must be positive")
	}

	if req.TotalPoint != nil && *req.TotalPoint < 0 {
		return errors.New("total point must be non-negative")
	}

	return nil
}
