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

// Team-related Redis keys (Contest-based)

// GetTeamKey returns the key for a team's data in a contest
func GetTeamKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:team", contestId)
}

// GetTeamMemberKey returns the key for a specific member in a team
func GetTeamMemberKey(contestId, userId int64) string {
	return fmt.Sprintf("contest:%d:team:member:%d", contestId, userId)
}

// GetTeamMembersKey returns the key for the set of all team members
func GetTeamMembersKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:team:members", contestId)
}

// GetTeamInvitesKey returns the key for pending team invites
func GetTeamInvitesKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:team:invites", contestId)
}

// GetTeamInviteKey returns the key for a specific invite
func GetTeamInviteKey(contestId, userId int64) string {
	return fmt.Sprintf("contest:%d:team:invite:%d", contestId, userId)
}

// GetUserTeamsKey returns the key for tracking user's team memberships (by contestID)
func GetUserTeamsKey(userId int64) string {
	return fmt.Sprintf("user:%d:contest-teams", userId)
}

// GetContestTeamPatternKey returns the pattern for all team-related keys of a contest
func GetContestTeamPatternKey(contestId int64) string {
	return fmt.Sprintf("contest:%d:team*", contestId)
}

// Discord token related Redis keys

// GetDiscordTokenKey returns the key for a user's Discord OAuth2 token
func GetDiscordTokenKey(userId int64) string {
	return fmt.Sprintf("discord:token:%d", userId)
}

// Discord bot cache related Redis keys

// GetDiscordBotGuildsKey returns the key for caching the bot's guild list
func GetDiscordBotGuildsKey() string {
	return "discord:cache:bot:guilds"
}

// GetDiscordGuildChannelsKey returns the key for caching a guild's channel list
func GetDiscordGuildChannelsKey(guildID string) string {
	return fmt.Sprintf("discord:cache:guild:%s:channels", guildID)
}
