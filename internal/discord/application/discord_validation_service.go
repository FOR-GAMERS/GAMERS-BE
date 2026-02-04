package application

import (
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	oauth2Port "github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/application/port"
)

// DiscordValidationService handles Discord validation logic for contests
type DiscordValidationService struct {
	botClient        port.DiscordBotPort
	userClient       port.DiscordUserPort
	discordTokenPort port.DiscordTokenPort
	oauth2DBPort     oauth2Port.OAuth2DatabasePort
}

// NewDiscordValidationService creates a new Discord validation service
func NewDiscordValidationService(
	botClient port.DiscordBotPort,
	userClient port.DiscordUserPort,
	discordTokenPort port.DiscordTokenPort,
	oauth2DBPort oauth2Port.OAuth2DatabasePort,
) *DiscordValidationService {
	return &DiscordValidationService{
		botClient:        botClient,
		userClient:       userClient,
		discordTokenPort: discordTokenPort,
		oauth2DBPort:     oauth2DBPort,
	}
}

// ValidateGuildForContest validates that:
// 1. The bot is in the specified guild
// 2. The user (by Discord ID) is in the specified guild
// 3. The channel exists in the guild and is a text channel
func (s *DiscordValidationService) ValidateGuildForContest(guildID, channelID, userDiscordID string) error {
	// Check if bot is in the guild
	botInGuild, err := s.botClient.IsBotInGuild(guildID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !botInGuild {
		return exception.ErrBotNotInGuild
	}

	// Check if user is in the guild
	userInGuild, err := s.botClient.IsUserInGuild(guildID, userDiscordID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !userInGuild {
		return exception.ErrUserNotInGuild
	}

	// Check if channel exists and is a text channel
	validChannel, err := s.botClient.ValidateGuildChannel(guildID, channelID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !validChannel {
		return exception.ErrChannelNotInGuild
	}

	return nil
}

// ValidateBotInGuild checks if the bot is in the specified guild
func (s *DiscordValidationService) ValidateBotInGuild(guildID string) error {
	botInGuild, err := s.botClient.IsBotInGuild(guildID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !botInGuild {
		return exception.ErrBotNotInGuild
	}
	return nil
}

// ValidateUserInGuild checks if the user is in the specified guild
func (s *DiscordValidationService) ValidateUserInGuild(guildID, userDiscordID string) error {
	userInGuild, err := s.botClient.IsUserInGuild(guildID, userDiscordID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !userInGuild {
		return exception.ErrUserNotInGuild
	}
	return nil
}

// GetGuildTextChannels returns all text channels in a guild
func (s *DiscordValidationService) GetGuildTextChannels(guildID string) ([]dto.DiscordChannel, error) {
	// First verify bot is in the guild
	if err := s.ValidateBotInGuild(guildID); err != nil {
		return nil, err
	}

	channels, err := s.botClient.GetGuildTextChannels(guildID)
	if err != nil {
		return nil, exception.ErrDiscordAPIError
	}

	return channels, nil
}

// GetBotGuilds returns all guilds the bot is a member of
func (s *DiscordValidationService) GetBotGuilds() ([]dto.DiscordGuild, error) {
	guilds, err := s.botClient.GetBotGuilds()
	if err != nil {
		return nil, exception.ErrDiscordAPIError
	}
	return guilds, nil
}

// GetAvailableGuilds returns guilds where both the GAMERS bot and the user are members.
// This is used to determine which guilds a user can create contests in.
// User guild and bot guild fetches run in parallel for better latency.
func (s *DiscordValidationService) GetAvailableGuilds(userID int64) ([]dto.DiscordGuild, error) {
	// Get user's Discord token from Redis
	discordToken, err := s.discordTokenPort.GetToken(userID)
	if err != nil {
		return nil, exception.ErrDiscordTokenNotFound
	}

	// Check if token is expired
	if time.Now().Unix() > discordToken.ExpiresAt {
		return nil, exception.ErrDiscordTokenExpired
	}

	var userGuilds []dto.DiscordGuild
	var botGuilds []dto.DiscordGuild

	// Fetch user guilds and bot guilds in parallel
	g := new(errgroup.Group)

	g.Go(func() error {
		var err error
		userGuilds, err = s.userClient.GetUserGuilds(discordToken.AccessToken)
		return err
	})

	g.Go(func() error {
		var err error
		botGuilds, err = s.botClient.GetBotGuilds()
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, exception.ErrDiscordAPIError
	}

	// Create a map of bot guilds for efficient lookup
	botGuildMap := make(map[string]bool)
	for _, guild := range botGuilds {
		botGuildMap[guild.ID] = true
	}

	// Find intersection: guilds where both user and bot are members
	availableGuilds := make([]dto.DiscordGuild, 0)
	for _, guild := range userGuilds {
		if botGuildMap[guild.ID] {
			availableGuilds = append(availableGuilds, guild)
		}
	}

	return availableGuilds, nil
}

// GetAvailableGuildTextChannels returns text channels for a guild where both the bot and user are members.
// Bot-in-guild validation and user-in-guild check run in parallel for better latency.
func (s *DiscordValidationService) GetAvailableGuildTextChannels(guildID string, userID int64) ([]dto.DiscordChannel, error) {
	// Get user's Discord token from Redis
	discordToken, err := s.discordTokenPort.GetToken(userID)
	if err != nil {
		return nil, exception.ErrDiscordTokenNotFound
	}

	// Check if token is expired
	if time.Now().Unix() > discordToken.ExpiresAt {
		return nil, exception.ErrDiscordTokenExpired
	}

	// Validate bot-in-guild and user-in-guild in parallel
	var userInGuild bool

	g := new(errgroup.Group)

	g.Go(func() error {
		return s.ValidateBotInGuild(guildID)
	})

	g.Go(func() error {
		var err error
		userInGuild, err = s.userClient.IsUserInGuild(discordToken.AccessToken, guildID)
		if err != nil {
			return err
		}
		if !userInGuild {
			return exception.ErrUserNotInGuild
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		// Return the error directly if it's already a known business error
		if err == exception.ErrBotNotInGuild || err == exception.ErrUserNotInGuild {
			return nil, err
		}
		return nil, exception.ErrDiscordAPIError
	}

	// GetGuildTextChannels benefits from Redis caching on subsequent calls
	channels, err := s.botClient.GetGuildTextChannels(guildID)
	if err != nil {
		return nil, exception.ErrDiscordAPIError
	}

	return channels, nil
}
