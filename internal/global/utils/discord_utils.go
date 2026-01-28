package utils

import "fmt"

const (
	DiscordCDNBaseURL = "https://cdn.discordapp.com"
)

// BuildDiscordAvatarURL constructs a full Discord CDN avatar URL
// Returns empty string if discordId or avatarHash is empty
func BuildDiscordAvatarURL(discordId, avatarHash string) string {
	if discordId == "" || avatarHash == "" {
		return ""
	}

	// Animated avatars start with "a_"
	extension := "png"
	if len(avatarHash) > 2 && avatarHash[:2] == "a_" {
		extension = "gif"
	}

	return fmt.Sprintf("%s/avatars/%s/%s.%s", DiscordCDNBaseURL, discordId, avatarHash, extension)
}
