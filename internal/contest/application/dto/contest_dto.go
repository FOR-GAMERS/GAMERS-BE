package dto

import (
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	gameDomain "GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/utils"
	"errors"
	"time"
)

type CreateContestRequest struct {
	Title                string               `json:"title" binding:"required"`
	Description          string               `json:"description,omitempty"`
	MaxTeamCount         int                  `json:"max_team_count,omitempty"`
	TotalPoint           int                  `json:"total_point,omitempty"`
	ContestType          domain.ContestType   `json:"contest_type" binding:"required"`
	StartedAt            time.Time            `json:"started_at,omitempty"`
	EndedAt              time.Time            `json:"ended_at,omitempty"`
	AutoStart            bool                 `json:"auto_start,omitempty"`
	GameType             *gameDomain.GameType `json:"game_type,omitempty"`
	GamePointTableId     *int64               `json:"game_point_table_id,omitempty"`
	TotalTeamMember      int                  `json:"total_team_member,omitempty"`
	DiscordGuildId       *string              `json:"discord_guild_id,omitempty"`
	DiscordTextChannelId *string              `json:"discord_text_channel_id,omitempty"`
	Thumbnail            *string              `json:"thumbnail,omitempty"`
}

type UpdateContestRequest struct {
	Title                *string               `json:"title,omitempty"`
	Description          *string               `json:"description,omitempty"`
	MaxTeamCount         *int                  `json:"max_team_count,omitempty"`
	TotalPoint           *int                  `json:"total_point,omitempty"`
	ContestType          *domain.ContestType   `json:"contest_type,omitempty"`
	ContestStatus        *domain.ContestStatus `json:"contest_status,omitempty"`
	StartedAt            *time.Time            `json:"started_at,omitempty"`
	EndedAt              *time.Time            `json:"ended_at,omitempty"`
	AutoStart            *bool                 `json:"auto_start,omitempty"`
	GameType             *gameDomain.GameType  `json:"game_type,omitempty"`
	GamePointTableId     *int64                `json:"game_point_table_id,omitempty"`
	TotalTeamMember      *int                  `json:"total_team_member,omitempty"`
	DiscordGuildId       *string               `json:"discord_guild_id,omitempty"`
	DiscordTextChannelId *string               `json:"discord_text_channel_id,omitempty"`
	Thumbnail            *string               `json:"thumbnail,omitempty"`
}

type ContestResponse struct {
	ContestID            int64                `json:"contest_id"`
	Title                string               `json:"title"`
	Description          string               `json:"description,omitempty"`
	MaxTeamCount         int                  `json:"max_team_count,omitempty"`
	TotalPoint           int                  `json:"total_point"`
	ContestType          domain.ContestType   `json:"contest_type"`
	ContestStatus        domain.ContestStatus `json:"contest_status"`
	StartedAt            time.Time            `json:"started_at,omitempty"`
	EndedAt              time.Time            `json:"ended_at,omitempty"`
	AutoStart            bool                 `json:"auto_start,omitempty"`
	GameType             *gameDomain.GameType `json:"game_type,omitempty"`
	GamePointTableId     *int64               `json:"game_point_table_id,omitempty"`
	TotalTeamMember      int                  `json:"total_team_member"`
	DiscordGuildId       *string              `json:"discord_guild_id,omitempty"`
	DiscordTextChannelId *string              `json:"discord_text_channel_id,omitempty"`
	Thumbnail            *string              `json:"thumbnail,omitempty"`
	CreatedAt            time.Time            `json:"created_at"`
	ModifiedAt           time.Time            `json:"modified_at"`
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
	if req.GameType != nil {
		contest.GameType = req.GameType
	}
	if req.GamePointTableId != nil {
		contest.GamePointTableId = req.GamePointTableId
	}
	if req.TotalTeamMember != nil {
		contest.TotalTeamMember = *req.TotalTeamMember
	}
	if req.DiscordGuildId != nil {
		contest.DiscordGuildId = req.DiscordGuildId
	}
	if req.DiscordTextChannelId != nil {
		contest.DiscordTextChannelId = req.DiscordTextChannelId
	}
	if req.Thumbnail != nil {
		contest.Thumbnail = req.Thumbnail
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
		req.AutoStart != nil ||
		req.GameType != nil ||
		req.GamePointTableId != nil ||
		req.TotalTeamMember != nil ||
		req.DiscordGuildId != nil ||
		req.DiscordTextChannelId != nil ||
		req.Thumbnail != nil
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

	if req.TotalTeamMember != nil && *req.TotalTeamMember < 1 {
		return errors.New("total team member must be at least 1")
	}

	if req.GameType != nil && !req.GameType.IsValid() {
		return errors.New("invalid game type")
	}

	return nil
}

// ContestMemberResponse represents a contest member with user information
type ContestMemberResponse struct {
	UserID             int64             `json:"user_id"`
	ContestID          int64             `json:"contest_id"`
	MemberType         domain.MemberType `json:"member_type"`
	LeaderType         domain.LeaderType `json:"leader_type"`
	Point              int               `json:"point"`
	Username           string            `json:"username"`
	Tag                string            `json:"tag"`
	Avatar             string            `json:"avatar"`
	CurrentTier        *int              `json:"current_tier,omitempty"`
	CurrentTierPatched *string           `json:"current_tier_patched,omitempty"`
	PeakTier           *int              `json:"peak_tier,omitempty"`
	PeakTierPatched    *string           `json:"peak_tier_patched,omitempty"`
}

// ToContestMemberResponse converts port.ContestMemberWithUser to ContestMemberResponse
func ToContestMemberResponse(member *port.ContestMemberWithUser) *ContestMemberResponse {
	// Build Discord avatar URL if Discord account exists
	avatar := member.Avatar
	if member.DiscordId != nil && member.DiscordAvatar != nil {
		if url := utils.BuildDiscordAvatarURL(*member.DiscordId, *member.DiscordAvatar); url != "" {
			avatar = url
		}
	}

	return &ContestMemberResponse{
		UserID:             member.UserID,
		ContestID:          member.ContestID,
		MemberType:         member.MemberType,
		LeaderType:         member.LeaderType,
		Point:              member.Point,
		Username:           member.Username,
		Tag:                member.Tag,
		Avatar:             avatar,
		CurrentTier:        member.CurrentTier,
		CurrentTierPatched: member.CurrentTierPatched,
		PeakTier:           member.PeakTier,
		PeakTierPatched:    member.PeakTierPatched,
	}
}

// ToContestMemberResponses converts a slice of port.ContestMemberWithUser to ContestMemberResponse slice
func ToContestMemberResponses(members []*port.ContestMemberWithUser) []*ContestMemberResponse {
	responses := make([]*ContestMemberResponse, len(members))
	for i, member := range members {
		responses[i] = ToContestMemberResponse(member)
	}
	return responses
}

// MyContestResponse represents a contest the user has joined with membership info
type MyContestResponse struct {
	ContestID            int64                 `json:"contest_id"`
	Title                string                `json:"title"`
	Description          string                `json:"description,omitempty"`
	MaxTeamCount         int                   `json:"max_team_count,omitempty"`
	TotalPoint           int                   `json:"total_point"`
	ContestType          domain.ContestType    `json:"contest_type"`
	ContestStatus        domain.ContestStatus  `json:"contest_status"`
	StartedAt            time.Time             `json:"started_at,omitempty"`
	EndedAt              time.Time             `json:"ended_at,omitempty"`
	AutoStart            bool                  `json:"auto_start,omitempty"`
	GameType             *gameDomain.GameType  `json:"game_type,omitempty"`
	GamePointTableId     *int64                `json:"game_point_table_id,omitempty"`
	TotalTeamMember      int                   `json:"total_team_member"`
	DiscordGuildId       *string               `json:"discord_guild_id,omitempty"`
	DiscordTextChannelId *string               `json:"discord_text_channel_id,omitempty"`
	Thumbnail            *string               `json:"thumbnail,omitempty"`
	CreatedAt            time.Time             `json:"created_at"`
	ModifiedAt           time.Time             `json:"modified_at"`
	MemberType           domain.MemberType     `json:"member_type"`
	LeaderType           domain.LeaderType     `json:"leader_type"`
	Point                int                   `json:"point"`
}

// ToMyContestResponse converts port.ContestWithMembership to MyContestResponse
func ToMyContestResponse(c *port.ContestWithMembership) *MyContestResponse {
	return &MyContestResponse{
		ContestID:            c.ContestID,
		Title:                c.Title,
		Description:          c.Description,
		MaxTeamCount:         c.MaxTeamCount,
		TotalPoint:           c.TotalPoint,
		ContestType:          c.ContestType,
		ContestStatus:        c.ContestStatus,
		StartedAt:            c.StartedAt,
		EndedAt:              c.EndedAt,
		AutoStart:            c.AutoStart,
		GameType:             c.GameType,
		GamePointTableId:     c.GamePointTableId,
		TotalTeamMember:      c.TotalTeamMember,
		DiscordGuildId:       c.DiscordGuildId,
		DiscordTextChannelId: c.DiscordTextChannelId,
		Thumbnail:            c.Thumbnail,
		CreatedAt:            c.CreatedAt,
		ModifiedAt:           c.ModifiedAt,
		MemberType:           c.MemberType,
		LeaderType:           c.LeaderType,
		Point:                c.Point,
	}
}

// ToMyContestResponses converts a slice of port.ContestWithMembership to MyContestResponse slice
func ToMyContestResponses(contests []*port.ContestWithMembership) []*MyContestResponse {
	responses := make([]*MyContestResponse, len(contests))
	for i, contest := range contests {
		responses[i] = ToMyContestResponse(contest)
	}
	return responses
}

// ChangeMemberRoleRequest represents the request to change a member's role
type ChangeMemberRoleRequest struct {
	MemberType domain.MemberType `json:"member_type" binding:"required"`
}

// ChangeMemberRoleResponse represents the response after changing a member's role
type ChangeMemberRoleResponse struct {
	UserID     int64             `json:"user_id"`
	ContestID  int64             `json:"contest_id"`
	MemberType domain.MemberType `json:"member_type"`
	LeaderType domain.LeaderType `json:"leader_type"`
}
