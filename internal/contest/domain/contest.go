package domain

import (
	"GAMERS-BE/internal/global/exception"
	"time"
)

type ContestType string

const (
	ContestTypeTournament ContestType = "TOURNAMENT"
	ContestTypeLeague     ContestType = "LEAGUE"
	ContestTypeCasual     ContestType = "CASUAL"
)

type ContestStatus string

const (
	ContestStatusPending   ContestStatus = "PENDING"
	ContestStatusActive    ContestStatus = "ACTIVE"
	ContestStatusFinished  ContestStatus = "FINISHED"
	ContestStatusCancelled ContestStatus = "CANCELLED"
)

type Contest struct {
	ContestID     int64         `gorm:"column:contest_id;primaryKey;autoIncrement" json:"contest_id"`
	Title         string        `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description   string        `gorm:"column:description;type:varchar(255)" json:"description,omitempty"`
	MaxTeamCount  int           `gorm:"column:max_team_count;type:int" json:"max_team_count,omitempty"`
	TotalPoint    int           `gorm:"column:total_point;type:int;default:100" json:"total_point"`
	ContestType   ContestType   `gorm:"column:contest_type;type:varchar(16);not null" json:"contest_type"`
	ContestStatus ContestStatus `gorm:"column:contest_status;type:varchar(16);not null" json:"contest_status"`
	StartedAt     time.Time     `gorm:"column:started_at;type:datetime" json:"started_at,omitempty"`
	EndedAt       time.Time     `gorm:"column:ended_at;type:datetime" json:"ended_at,omitempty"`

	AutoStart bool `gorm:"column:auto_start;type:boolean;default:false" json:"auto_start"`

	DiscordGuildId       *string `gorm:"column:discord_guild_id;type:varchar(255)" json:"discord_guild_id,omitempty"`
	DiscordTextChannelId *string `gorm:"column:discord_text_channel_id;type:varchar(255)" json:"discord_text_channel_id,omitempty"`

	CreatedAt  time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	ModifiedAt time.Time `gorm:"column:modified_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"modified_at"`
}

func NewContestInstance(
	title, description string,
	maxTeamCount, totalPoint int,
	contestType ContestType,
	startedAt, endedAt time.Time,
	autoStart bool,
	discordGuildId, discordTextChannelId *string,
) *Contest {
	return &Contest{
		Title:                title,
		Description:          description,
		MaxTeamCount:         maxTeamCount,
		TotalPoint:           totalPoint,
		ContestType:          contestType,
		ContestStatus:        ContestStatusPending,
		StartedAt:            startedAt,
		EndedAt:              endedAt,
		AutoStart:            autoStart,
		DiscordGuildId:       discordGuildId,
		DiscordTextChannelId: discordTextChannelId,
	}
}

func (c *Contest) TableName() string {
	return "contests"
}

var allowedTransitions = map[ContestStatus][]ContestStatus{
	ContestStatusPending: {
		ContestStatusActive,
		ContestStatusCancelled,
	},
	ContestStatusActive: {
		ContestStatusFinished,
		ContestStatusCancelled,
	},
	ContestStatusFinished:  {},
	ContestStatusCancelled: {},
}

func (c *Contest) CanTransitionTo(targetStatus ContestStatus) bool {
	allowedTargets, exists := allowedTransitions[c.ContestStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedTargets {
		if allowed == targetStatus {
			return true
		}
	}
	return false
}

func (c *Contest) TransitionTo(targetStatus ContestStatus) error {
	if !c.CanTransitionTo(targetStatus) {
		return exception.ErrInvalidStatusTransition
	}

	c.ContestStatus = targetStatus
	return nil
}

func (c *Contest) IsTerminalState() bool {
	return c.ContestStatus == ContestStatusFinished || c.ContestStatus == ContestStatusCancelled
}

func (c *Contest) CanStart() bool {
	return c.ContestStatus == ContestStatusPending && time.Now().After(c.StartedAt)
}

func (c *Contest) IsBeforeStartTime() bool {
	return time.Now().Before(c.StartedAt)
}

// ValidateDates checks if the contest dates are valid
func (c *Contest) ValidateDates() error {
	if !c.StartedAt.IsZero() && !c.EndedAt.IsZero() {
		if !c.StartedAt.Before(c.EndedAt) {
			return exception.ErrInvalidContestDates
		}
	}
	return nil
}

func (c *Contest) IsValidType() bool {
	switch c.ContestType {
	case ContestTypeTournament, ContestTypeLeague, ContestTypeCasual:
		return true
	default:
		return false
	}
}

func (c *Contest) IsValidStatus() bool {
	switch c.ContestStatus {
	case ContestStatusPending, ContestStatusActive, ContestStatusFinished, ContestStatusCancelled:
		return true
	default:
		return false
	}
}

func (c *Contest) ValidateBusinessRules() error {
	if c.MaxTeamCount < 0 {
		return exception.ErrInvalidMaxTeamCount
	}

	if c.TotalPoint < 0 {
		return exception.ErrInvalidTotalPoint
	}

	if err := c.ValidateDiscordFields(); err != nil {
		return err
	}

	return nil
}

// ValidateDiscordFields checks if Discord fields are valid
// If guild_id is provided, text_channel_id must also be provided
func (c *Contest) ValidateDiscordFields() error {
	if c.DiscordGuildId != nil && *c.DiscordGuildId != "" {
		if c.DiscordTextChannelId == nil || *c.DiscordTextChannelId == "" {
			return exception.ErrDiscordTextChannelRequired
		}
	}
	return nil
}

// HasDiscordIntegration checks if the contest has Discord integration configured
func (c *Contest) HasDiscordIntegration() bool {
	return c.DiscordGuildId != nil && *c.DiscordGuildId != ""
}

func (c *Contest) Validate() error {
	if c.Title == "" {
		return exception.ErrInvalidContestTitle
	}

	if !c.IsValidType() {
		return exception.ErrInvalidContestType
	}

	if !c.IsValidStatus() {
		return exception.ErrInvalidContestStatus
	}

	if err := c.ValidateDates(); err != nil {
		return err
	}

	if err := c.ValidateBusinessRules(); err != nil {
		return err
	}

	return nil
}
