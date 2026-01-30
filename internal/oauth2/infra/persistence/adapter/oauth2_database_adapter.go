package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/domain"
	"errors"

	"gorm.io/gorm"
)

type OAuth2DatabaseAdapter struct {
	db *gorm.DB
}

func NewOAuth2DatabaseAdapter(db *gorm.DB) *OAuth2DatabaseAdapter {
	return &OAuth2DatabaseAdapter{db: db}
}

func (a *OAuth2DatabaseAdapter) FindDiscordAccountByDiscordId(discordId string) (*domain.DiscordAccount, error) {
	var account domain.DiscordAccount

	result := a.db.Where("discord_id = ?", discordId).First(&account)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrDiscordUserCannotFound
		}
		return nil, result.Error
	}

	return &account, nil
}

func (a *OAuth2DatabaseAdapter) FindDiscordAccountByUserId(userId int64) (*domain.DiscordAccount, error) {
	var account domain.DiscordAccount

	result := a.db.Where("user_id = ?", userId).First(&account)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrDiscordUserCannotFound
		}
		return nil, result.Error
	}

	return &account, nil
}

func (a *OAuth2DatabaseAdapter) CreateDiscordAccount(account *domain.DiscordAccount) error {
	result := a.db.Create(account)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (a *OAuth2DatabaseAdapter) UpdateDiscordAccount(account *domain.DiscordAccount) error {
	result := a.db.Model(&domain.DiscordAccount{}).
		Where("discord_id = ?", account.DiscordId).
		Updates(map[string]interface{}{
			"discord_avatar":   account.DiscordAvatar,
			"discord_verified": account.DiscordVerified,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}
