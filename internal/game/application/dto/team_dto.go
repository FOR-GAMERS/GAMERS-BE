package dto

import (
	"GAMERS-BE/internal/game/application/port"
	gameDomain "GAMERS-BE/internal/game/domain"
)

type CreateTeamRequest struct {
	TeamName *string `json:"team_name"`
}

type InviteMemberRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

type KickMemberRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

type TransferLeadershipRequest struct {
	NewLeaderUserID int64 `json:"new_leader_user_id" binding:"required"`
}

// TeamMemberResponse represents a team member in a game
type TeamMemberResponse struct {
	TeamID     int64  `json:"team_id"`
	GameID     int64  `json:"game_id"`
	UserID     int64  `json:"user_id"`
	MemberType string `json:"member_type"`
	Username   string `json:"username,omitempty"`
	Tag        string `json:"tag,omitempty"`
}

type TeamResponse struct {
	GameID      int64                 `json:"game_id"`
	TeamID      int64                 `json:"team_id"`
	TeamName    *string               `json:"team_name,omitempty"`
	MaxMembers  int                   `json:"max_members"`
	MemberCount int                   `json:"member_count"`
	IsFinalized bool                  `json:"is_finalized"`
	Members     []*TeamMemberResponse `json:"members"`
}

func ToTeamMemberResponse(member *gameDomain.TeamMember, gameID int64) *TeamMemberResponse {
	return &TeamMemberResponse{
		TeamID:     member.TeamID,
		GameID:     gameID,
		UserID:     member.UserID,
		MemberType: string(member.MemberType),
	}
}

func ToTeamMemberResponses(members []*gameDomain.TeamMember, gameID int64) []*TeamMemberResponse {
	responses := make([]*TeamMemberResponse, len(members))
	for i, member := range members {
		responses[i] = ToTeamMemberResponse(member, gameID)
	}
	return responses
}

func ToTeamResponse(game *gameDomain.Game, team *gameDomain.Team, members []*gameDomain.TeamMember) *TeamResponse {
	var teamID int64
	var teamName *string
	if team != nil {
		teamID = team.TeamID
		teamName = &team.TeamName
	}
	return &TeamResponse{
		GameID:      game.GameID,
		TeamID:      teamID,
		TeamName:    teamName,
		MaxMembers:  game.GameTeamType.GetMaxTeamMembers(),
		MemberCount: len(members),
		IsFinalized: true,
		Members:     ToTeamMemberResponses(members, game.GameID),
	}
}

// ToCachedTeamResponse converts cached team data to TeamResponse
func ToCachedTeamResponse(team *port.CachedTeam, members []*port.CachedTeamMember) *TeamResponse {
	memberResponses := make([]*TeamMemberResponse, len(members))
	for i, m := range members {
		memberType := string(gameDomain.TeamMemberTypeMember)
		if m.MemberType == port.TeamMemberTypeLeader {
			memberType = string(gameDomain.TeamMemberTypeLeader)
		}
		memberResponses[i] = &TeamMemberResponse{
			TeamID:     m.TeamID,
			GameID:     m.GameID,
			UserID:     m.UserID,
			MemberType: memberType,
			Username:   m.Username,
			Tag:        m.Tag,
		}
	}

	return &TeamResponse{
		GameID:      team.GameID,
		TeamID:      team.TeamID,
		TeamName:    team.TeamName,
		MaxMembers:  team.MaxMembers,
		MemberCount: team.CurrentCount,
		IsFinalized: team.IsFinalized,
		Members:     memberResponses,
	}
}

// TeamInviteResponse represents a team invite in response
type TeamInviteResponse struct {
	GameID      int64  `json:"game_id"`
	InviterID   int64  `json:"inviter_id"`
	InviteeID   int64  `json:"invitee_id"`
	Status      string `json:"status"`
	InviterName string `json:"inviter_name,omitempty"`
	InviteeName string `json:"invitee_name,omitempty"`
}

func ToTeamInviteResponse(invite *port.TeamInvite) *TeamInviteResponse {
	return &TeamInviteResponse{
		GameID:      invite.GameID,
		InviterID:   invite.InviterID,
		InviteeID:   invite.InviteeID,
		Status:      string(invite.Status),
		InviterName: invite.InviterName,
		InviteeName: invite.InviteeName,
	}
}

// CachedMemberResponse converts cached member to response
type CachedMemberResponse struct {
	UserID     int64  `json:"user_id"`
	GameID     int64  `json:"game_id"`
	TeamID     int64  `json:"team_id"`
	MemberType string `json:"member_type"`
	Username   string `json:"username,omitempty"`
	Tag        string `json:"tag,omitempty"`
	DiscordID  string `json:"discord_id,omitempty"`
}

func ToCachedMemberResponse(member *port.CachedTeamMember) *CachedMemberResponse {
	return &CachedMemberResponse{
		UserID:     member.UserID,
		GameID:     member.GameID,
		TeamID:     member.TeamID,
		MemberType: string(member.MemberType),
		Username:   member.Username,
		Tag:        member.Tag,
		DiscordID:  member.DiscordID,
	}
}

func ToCachedMemberResponses(members []*port.CachedTeamMember) []*CachedMemberResponse {
	responses := make([]*CachedMemberResponse, len(members))
	for i, m := range members {
		responses[i] = ToCachedMemberResponse(m)
	}
	return responses
}
