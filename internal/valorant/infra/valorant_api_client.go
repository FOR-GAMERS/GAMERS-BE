package infra

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/valorant/application/port"

	govapi "github.com/yldshv/go-valorant-api"
)

type ValorantApiClient struct {
	client *govapi.VAPI
}

func NewValorantApiClient(apiKey string) *ValorantApiClient {
	var client *govapi.VAPI
	if apiKey != "" {
		client = govapi.New(govapi.WithKey(apiKey))
	} else {
		client = govapi.New()
	}
	return &ValorantApiClient{
		client: client,
	}
}

func (c *ValorantApiClient) GetMMRByName(region, name, tag string) (*port.ValorantMMRData, error) {
	// Get current MMR
	mmrResp, err := c.client.GetMMRByNameV2(govapi.GetMMRByNameV2Params{
		Affinity: region,
		Name:     name,
		Tag:      tag,
	})
	if err != nil {
		return nil, exception.ErrValorantApiError
	}

	if mmrResp.Status != 200 {
		if mmrResp.Status == 404 {
			return nil, exception.ErrValorantPlayerNotFound
		}
		return nil, exception.ErrValorantApiError
	}

	currentTier := mmrResp.Data.CurrentData.Currenttier
	currentTierPatched := mmrResp.Data.CurrentData.CurrenttierPatched
	rankingInTier := mmrResp.Data.CurrentData.RankingInTier
	elo := mmrResp.Data.CurrentData.Elo

	// Get peak rank from lifetime MMR history
	peakTier, peakTierPatched := c.findPeakRank(region, name, tag, currentTier, currentTierPatched)

	return &port.ValorantMMRData{
		CurrentTier:        currentTier,
		CurrentTierPatched: currentTierPatched,
		RankingInTier:      rankingInTier,
		Elo:                elo,
		PeakTier:           peakTier,
		PeakTierPatched:    peakTierPatched,
	}, nil
}

func (c *ValorantApiClient) findPeakRank(region, name, tag string, currentTier int, currentTierPatched string) (int, string) {
	// Try to get lifetime MMR history for peak rank
	historyResp, err := c.client.GetLifetimeMMRHistoryByName(govapi.GetLifetimeMMRHistoryByNameParams{
		Affinity: region,
		Name:     name,
		Tag:      tag,
		Page:     "1",
		Size:     "100",
	})
	if err != nil || historyResp.Status != 200 {
		// If we can't get history, use current rank as peak
		return currentTier, currentTierPatched
	}

	peakTier := currentTier
	peakTierPatched := currentTierPatched

	// Find the highest tier across all history
	for _, match := range historyResp.Data {
		if match.Tier.ID > peakTier {
			peakTier = match.Tier.ID
			peakTierPatched = match.Tier.Name
		}
	}

	return peakTier, peakTierPatched
}
