package port

// ValorantMMRData represents MMR data fetched from Valorant API
type ValorantMMRData struct {
	CurrentTier        int
	CurrentTierPatched string
	RankingInTier      int
	Elo                int
	PeakTier           int
	PeakTierPatched    string
}

// ValorantApiPort defines the interface for interacting with Valorant API
type ValorantApiPort interface {
	// GetMMRByName fetches MMR data for a player by their Riot ID
	GetMMRByName(region, name, tag string) (*ValorantMMRData, error)
}
