package domain

type GameType string

const (
	GameTypeValorant GameType = "VALORANT"
	GameTypeLOL      GameType = "LOL"
)

func (g GameType) IsValid() bool {
	switch g {
	case GameTypeValorant, GameTypeLOL:
		return true
	default:
		return false
	}
}
