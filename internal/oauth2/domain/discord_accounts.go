package domain

type DiscordAccount struct {
	DiscordId       string `gorm:"primaryKey;column:discord_id;" json:"discord_id"`
	UserId          int64  `gorm:"column:user_id;unique;" json:"user_id"`
	DiscordAvatar   string `gorm:"column:discord_avatar;type:varchar(255)" json:"discord_avatar"`
	DiscordVerified bool   `gorm:"column:discord_verified;type:boolean" json:"discord_verified"`
}
