package domain_test

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/domain"
	gameDomain "github.com/FOR-GAMERS/GAMERS-BE/internal/game/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewContestInstance(t *testing.T) {
	t.Run("creates contest with valid parameters", func(t *testing.T) {
		gameType := gameDomain.GameTypeValorant
		guildId := "guild123"
		channelId := "channel456"
		thumbnail := "thumbnail.jpg"
		startedAt := time.Now().Add(1 * time.Hour)
		endedAt := time.Now().Add(48 * time.Hour)

		contest := domain.NewContestInstance(
			"Test Tournament",
			"A test tournament",
			8,
			100,
			domain.ContestTypeTournament,
			startedAt,
			endedAt,
			false,
			&gameType,
			nil,
			5,
			&guildId,
			&channelId,
			&thumbnail,
		)

		assert.NotNil(t, contest)
		assert.Equal(t, "Test Tournament", contest.Title)
		assert.Equal(t, "A test tournament", contest.Description)
		assert.Equal(t, 8, contest.MaxTeamCount)
		assert.Equal(t, 100, contest.TotalPoint)
		assert.Equal(t, domain.ContestTypeTournament, contest.ContestType)
		assert.Equal(t, domain.ContestStatusPending, contest.ContestStatus)
		assert.Equal(t, 5, contest.TotalTeamMember)
		assert.False(t, contest.AutoStart)
	})

	t.Run("creates contest with default pending status", func(t *testing.T) {
		contest := domain.NewContestInstance(
			"Test",
			"",
			4,
			50,
			domain.ContestTypeLeague,
			time.Now(),
			time.Now().Add(24*time.Hour),
			false,
			nil,
			nil,
			3,
			nil,
			nil,
			nil,
		)

		assert.Equal(t, domain.ContestStatusPending, contest.ContestStatus)
	})
}

func TestContest_TableName(t *testing.T) {
	contest := &domain.Contest{}
	assert.Equal(t, "contests", contest.TableName())
}

func TestContest_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus domain.ContestStatus
		targetStatus  domain.ContestStatus
		expected      bool
	}{
		// From PENDING
		{"PENDING to ACTIVE", domain.ContestStatusPending, domain.ContestStatusActive, true},
		{"PENDING to CANCELLED", domain.ContestStatusPending, domain.ContestStatusCancelled, true},
		{"PENDING to FINISHED", domain.ContestStatusPending, domain.ContestStatusFinished, false},
		{"PENDING to PENDING", domain.ContestStatusPending, domain.ContestStatusPending, false},

		// From ACTIVE
		{"ACTIVE to FINISHED", domain.ContestStatusActive, domain.ContestStatusFinished, true},
		{"ACTIVE to CANCELLED", domain.ContestStatusActive, domain.ContestStatusCancelled, true},
		{"ACTIVE to PENDING", domain.ContestStatusActive, domain.ContestStatusPending, false},
		{"ACTIVE to ACTIVE", domain.ContestStatusActive, domain.ContestStatusActive, false},

		// From FINISHED (terminal)
		{"FINISHED to PENDING", domain.ContestStatusFinished, domain.ContestStatusPending, false},
		{"FINISHED to ACTIVE", domain.ContestStatusFinished, domain.ContestStatusActive, false},
		{"FINISHED to CANCELLED", domain.ContestStatusFinished, domain.ContestStatusCancelled, false},

		// From CANCELLED (terminal)
		{"CANCELLED to PENDING", domain.ContestStatusCancelled, domain.ContestStatusPending, false},
		{"CANCELLED to ACTIVE", domain.ContestStatusCancelled, domain.ContestStatusActive, false},
		{"CANCELLED to FINISHED", domain.ContestStatusCancelled, domain.ContestStatusFinished, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contest := &domain.Contest{ContestStatus: tt.currentStatus}
			result := contest.CanTransitionTo(tt.targetStatus)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContest_TransitionTo(t *testing.T) {
	t.Run("successful transition from PENDING to ACTIVE", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusPending}
		err := contest.TransitionTo(domain.ContestStatusActive)

		assert.NoError(t, err)
		assert.Equal(t, domain.ContestStatusActive, contest.ContestStatus)
	})

	t.Run("successful transition from ACTIVE to FINISHED", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusActive}
		err := contest.TransitionTo(domain.ContestStatusFinished)

		assert.NoError(t, err)
		assert.Equal(t, domain.ContestStatusFinished, contest.ContestStatus)
	})

	t.Run("fails on invalid transition", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusFinished}
		err := contest.TransitionTo(domain.ContestStatusActive)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrInvalidStatusTransition, err)
		assert.Equal(t, domain.ContestStatusFinished, contest.ContestStatus)
	})
}

func TestContest_IsTerminalState(t *testing.T) {
	tests := []struct {
		name     string
		status   domain.ContestStatus
		expected bool
	}{
		{"PENDING is not terminal", domain.ContestStatusPending, false},
		{"ACTIVE is not terminal", domain.ContestStatusActive, false},
		{"FINISHED is terminal", domain.ContestStatusFinished, true},
		{"CANCELLED is terminal", domain.ContestStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contest := &domain.Contest{ContestStatus: tt.status}
			assert.Equal(t, tt.expected, contest.IsTerminalState())
		})
	}
}

func TestContest_CanStart(t *testing.T) {
	t.Run("can start when PENDING and past start time", func(t *testing.T) {
		contest := &domain.Contest{
			ContestStatus: domain.ContestStatusPending,
			StartedAt:     time.Now().Add(-1 * time.Hour),
		}
		assert.True(t, contest.CanStart())
	})

	t.Run("cannot start when PENDING but before start time", func(t *testing.T) {
		contest := &domain.Contest{
			ContestStatus: domain.ContestStatusPending,
			StartedAt:     time.Now().Add(1 * time.Hour),
		}
		assert.False(t, contest.CanStart())
	})

	t.Run("cannot start when already ACTIVE", func(t *testing.T) {
		contest := &domain.Contest{
			ContestStatus: domain.ContestStatusActive,
			StartedAt:     time.Now().Add(-1 * time.Hour),
		}
		assert.False(t, contest.CanStart())
	})

	t.Run("cannot start when FINISHED", func(t *testing.T) {
		contest := &domain.Contest{
			ContestStatus: domain.ContestStatusFinished,
			StartedAt:     time.Now().Add(-1 * time.Hour),
		}
		assert.False(t, contest.CanStart())
	})
}

func TestContest_CanStop(t *testing.T) {
	t.Run("can stop when ACTIVE", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusActive}
		assert.True(t, contest.CanStop())
	})

	t.Run("cannot stop when PENDING", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusPending}
		assert.False(t, contest.CanStop())
	})

	t.Run("cannot stop when already FINISHED", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusFinished}
		assert.False(t, contest.CanStop())
	})
}

func TestContest_IsBeforeStartTime(t *testing.T) {
	t.Run("returns true when before start time", func(t *testing.T) {
		contest := &domain.Contest{
			StartedAt: time.Now().Add(1 * time.Hour),
		}
		assert.True(t, contest.IsBeforeStartTime())
	})

	t.Run("returns false when after start time", func(t *testing.T) {
		contest := &domain.Contest{
			StartedAt: time.Now().Add(-1 * time.Hour),
		}
		assert.False(t, contest.IsBeforeStartTime())
	})
}

func TestContest_ValidateDates(t *testing.T) {
	t.Run("valid when start is before end", func(t *testing.T) {
		contest := &domain.Contest{
			StartedAt: time.Now(),
			EndedAt:   time.Now().Add(24 * time.Hour),
		}
		err := contest.ValidateDates()
		assert.NoError(t, err)
	})

	t.Run("invalid when start is after end", func(t *testing.T) {
		contest := &domain.Contest{
			StartedAt: time.Now().Add(24 * time.Hour),
			EndedAt:   time.Now(),
		}
		err := contest.ValidateDates()
		assert.Equal(t, exception.ErrInvalidContestDates, err)
	})

	t.Run("valid when dates are zero", func(t *testing.T) {
		contest := &domain.Contest{}
		err := contest.ValidateDates()
		assert.NoError(t, err)
	})
}

func TestContest_IsValidType(t *testing.T) {
	tests := []struct {
		name        string
		contestType domain.ContestType
		expected    bool
	}{
		{"TOURNAMENT is valid", domain.ContestTypeTournament, true},
		{"LEAGUE is valid", domain.ContestTypeLeague, true},
		{"CASUAL is valid", domain.ContestTypeCasual, true},
		{"empty string is invalid", domain.ContestType(""), false},
		{"unknown type is invalid", domain.ContestType("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contest := &domain.Contest{ContestType: tt.contestType}
			assert.Equal(t, tt.expected, contest.IsValidType())
		})
	}
}

func TestContest_IsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   domain.ContestStatus
		expected bool
	}{
		{"PENDING is valid", domain.ContestStatusPending, true},
		{"ACTIVE is valid", domain.ContestStatusActive, true},
		{"FINISHED is valid", domain.ContestStatusFinished, true},
		{"CANCELLED is valid", domain.ContestStatusCancelled, true},
		{"empty string is invalid", domain.ContestStatus(""), false},
		{"unknown status is invalid", domain.ContestStatus("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contest := &domain.Contest{ContestStatus: tt.status}
			assert.Equal(t, tt.expected, contest.IsValidStatus())
		})
	}
}

func TestContest_ValidateBusinessRules(t *testing.T) {
	t.Run("valid with correct values", func(t *testing.T) {
		guildId := "guild123"
		channelId := "channel456"
		contest := &domain.Contest{
			MaxTeamCount:         8,
			TotalPoint:           100,
			TotalTeamMember:      5,
			DiscordGuildId:       &guildId,
			DiscordTextChannelId: &channelId,
		}
		err := contest.ValidateBusinessRules()
		assert.NoError(t, err)
	})

	t.Run("invalid with negative MaxTeamCount", func(t *testing.T) {
		contest := &domain.Contest{
			MaxTeamCount:    -1,
			TotalPoint:      100,
			TotalTeamMember: 5,
		}
		err := contest.ValidateBusinessRules()
		assert.Equal(t, exception.ErrInvalidMaxTeamCount, err)
	})

	t.Run("invalid with negative TotalPoint", func(t *testing.T) {
		contest := &domain.Contest{
			MaxTeamCount:    8,
			TotalPoint:      -1,
			TotalTeamMember: 5,
		}
		err := contest.ValidateBusinessRules()
		assert.Equal(t, exception.ErrInvalidTotalPoint, err)
	})

	t.Run("invalid with zero TotalTeamMember", func(t *testing.T) {
		contest := &domain.Contest{
			MaxTeamCount:    8,
			TotalPoint:      100,
			TotalTeamMember: 0,
		}
		err := contest.ValidateBusinessRules()
		assert.Equal(t, exception.ErrInvalidTotalTeamMember, err)
	})
}

func TestContest_ValidateDiscordFields(t *testing.T) {
	t.Run("valid without Discord integration", func(t *testing.T) {
		contest := &domain.Contest{}
		err := contest.ValidateDiscordFields()
		assert.NoError(t, err)
	})

	t.Run("valid with both guild and channel", func(t *testing.T) {
		guildId := "guild123"
		channelId := "channel456"
		contest := &domain.Contest{
			DiscordGuildId:       &guildId,
			DiscordTextChannelId: &channelId,
		}
		err := contest.ValidateDiscordFields()
		assert.NoError(t, err)
	})

	t.Run("invalid with guild but no channel", func(t *testing.T) {
		guildId := "guild123"
		contest := &domain.Contest{
			DiscordGuildId: &guildId,
		}
		err := contest.ValidateDiscordFields()
		assert.Equal(t, exception.ErrDiscordTextChannelRequired, err)
	})

	t.Run("invalid with guild but empty channel", func(t *testing.T) {
		guildId := "guild123"
		emptyChannel := ""
		contest := &domain.Contest{
			DiscordGuildId:       &guildId,
			DiscordTextChannelId: &emptyChannel,
		}
		err := contest.ValidateDiscordFields()
		assert.Equal(t, exception.ErrDiscordTextChannelRequired, err)
	})
}

func TestContest_HasDiscordIntegration(t *testing.T) {
	t.Run("returns true when guild is set", func(t *testing.T) {
		guildId := "guild123"
		contest := &domain.Contest{DiscordGuildId: &guildId}
		assert.True(t, contest.HasDiscordIntegration())
	})

	t.Run("returns false when guild is nil", func(t *testing.T) {
		contest := &domain.Contest{}
		assert.False(t, contest.HasDiscordIntegration())
	})

	t.Run("returns false when guild is empty", func(t *testing.T) {
		emptyGuild := ""
		contest := &domain.Contest{DiscordGuildId: &emptyGuild}
		assert.False(t, contest.HasDiscordIntegration())
	})
}

func TestContest_IsActive(t *testing.T) {
	t.Run("returns true when ACTIVE", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusActive}
		assert.True(t, contest.IsActive())
	})

	t.Run("returns false when not ACTIVE", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusPending}
		assert.False(t, contest.IsActive())
	})
}

func TestContest_IsPending(t *testing.T) {
	t.Run("returns true when PENDING", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusPending}
		assert.True(t, contest.IsPending())
	})

	t.Run("returns false when not PENDING", func(t *testing.T) {
		contest := &domain.Contest{ContestStatus: domain.ContestStatusActive}
		assert.False(t, contest.IsPending())
	})
}

func TestContest_Validate(t *testing.T) {
	t.Run("valid contest passes validation", func(t *testing.T) {
		contest := &domain.Contest{
			Title:           "Valid Tournament",
			ContestType:     domain.ContestTypeTournament,
			ContestStatus:   domain.ContestStatusPending,
			MaxTeamCount:    8,
			TotalPoint:      100,
			TotalTeamMember: 5,
			StartedAt:       time.Now(),
			EndedAt:         time.Now().Add(24 * time.Hour),
		}
		err := contest.Validate()
		assert.NoError(t, err)
	})

	t.Run("fails with empty title", func(t *testing.T) {
		contest := &domain.Contest{
			Title:         "",
			ContestType:   domain.ContestTypeTournament,
			ContestStatus: domain.ContestStatusPending,
		}
		err := contest.Validate()
		assert.Equal(t, exception.ErrInvalidContestTitle, err)
	})

	t.Run("fails with invalid type", func(t *testing.T) {
		contest := &domain.Contest{
			Title:         "Test",
			ContestType:   domain.ContestType("INVALID"),
			ContestStatus: domain.ContestStatusPending,
		}
		err := contest.Validate()
		assert.Equal(t, exception.ErrInvalidContestType, err)
	})

	t.Run("fails with invalid status", func(t *testing.T) {
		contest := &domain.Contest{
			Title:         "Test",
			ContestType:   domain.ContestTypeTournament,
			ContestStatus: domain.ContestStatus("INVALID"),
		}
		err := contest.Validate()
		assert.Equal(t, exception.ErrInvalidContestStatus, err)
	})

	t.Run("fails with invalid dates", func(t *testing.T) {
		contest := &domain.Contest{
			Title:           "Test",
			ContestType:     domain.ContestTypeTournament,
			ContestStatus:   domain.ContestStatusPending,
			MaxTeamCount:    8,
			TotalPoint:      100,
			TotalTeamMember: 5,
			StartedAt:       time.Now().Add(24 * time.Hour),
			EndedAt:         time.Now(),
		}
		err := contest.Validate()
		assert.Equal(t, exception.ErrInvalidContestDates, err)
	})

	t.Run("fails with invalid business rules", func(t *testing.T) {
		contest := &domain.Contest{
			Title:           "Test",
			ContestType:     domain.ContestTypeTournament,
			ContestStatus:   domain.ContestStatusPending,
			MaxTeamCount:    -1,
			TotalPoint:      100,
			TotalTeamMember: 5,
		}
		err := contest.Validate()
		assert.Equal(t, exception.ErrInvalidMaxTeamCount, err)
	})
}
