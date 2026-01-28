package port

import "GAMERS-BE/internal/oauth2/domain"

type OAuth2DatabasePort interface {
	FindDiscordAccountByDiscordId(discordId string) (*domain.DiscordAccount, error)
	FindDiscordAccountByUserId(userId int64) (*domain.DiscordAccount, error)
	CreateDiscordAccount(account *domain.DiscordAccount) error
	UpdateDiscordAccount(account *domain.DiscordAccount) error
}
