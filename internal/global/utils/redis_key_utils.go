package utils

import "fmt"

func GetApplicationKey(contestId, userId int64) string {
	return fmt.Sprintf("contest:%d:application:%d", contestId, userId)
}

func GetPendingKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:applications:pending", contestId)
}

func GetAcceptedKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:applications:accepted", contestId)
}

func GetRejectedKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:applications:rejected", contestId)
}

func GetUserApplicationsKey(userId int64) string {
	return fmt.Sprintf("user:%d:applications", userId)
}

func GetContestPatternKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:application*", contestId)
}

// Team-related Redis keys

// GetTeamKey returns the key for a team's data in a game
func GetTeamKey(gameId int64) string {
	return fmt.Sprintf("game:%d:team", gameId)
}

// GetTeamMemberKey returns the key for a specific member in a team
func GetTeamMemberKey(gameId, userId int64) string {
	return fmt.Sprintf("game:%d:team:member:%d", gameId, userId)
}

// GetTeamMembersKey returns the key for the set of all team members
func GetTeamMembersKey(gameId int64) string {
	return fmt.Sprintf("game:%d:team:members", gameId)
}

// GetTeamInvitesKey returns the key for pending team invites
func GetTeamInvitesKey(gameId int64) string {
	return fmt.Sprintf("game:%d:team:invites", gameId)
}

// GetTeamInviteKey returns the key for a specific invite
func GetTeamInviteKey(gameId, userId int64) string {
	return fmt.Sprintf("game:%d:team:invite:%d", gameId, userId)
}

// GetUserTeamsKey returns the key for tracking user's team memberships
func GetUserTeamsKey(userId int64) string {
	return fmt.Sprintf("user:%d:teams", userId)
}

// GetGameTeamPatternKey returns the pattern for all team-related keys of a game
func GetGameTeamPatternKey(gameId int64) string {
	return fmt.Sprintf("game:%d:team*", gameId)
}

// Discord token related Redis keys

// GetDiscordTokenKey returns the key for a user's Discord OAuth2 token
func GetDiscordTokenKey(userId int64) string {
	return fmt.Sprintf("discord:token:%d", userId)
}
