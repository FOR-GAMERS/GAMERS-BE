package infra

import (
	"log"
	"time"

	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/valorant/application/port"

	govapi "github.com/yldshv/go-valorant-api"
)

type ValorantApiClient struct {
	client     *govapi.VAPI
	maxRetries int
}

func NewValorantApiClient(apiKey string) *ValorantApiClient {
	var client *govapi.VAPI
	if apiKey != "" {
		client = govapi.New(govapi.WithKey(apiKey))
	} else {
		client = govapi.New()
	}
	return &ValorantApiClient{
		client:     client,
		maxRetries: 3,
	}
}

func (c *ValorantApiClient) GetMMRByName(region, name, tag string) (*port.ValorantMMRData, error) {
	var mmrResp *govapi.GetMMRByNameV2Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("[ValorantAPI] Retry attempt %d for GetMMRByName(%s#%s), waiting %v",
				attempt, name, tag, backoff)
			time.Sleep(backoff)
		}

		mmrResp, err = c.client.GetMMRByNameV2(govapi.GetMMRByNameV2Params{
			Affinity: region,
			Name:     name,
			Tag:      tag,
		})
		if err == nil && mmrResp != nil && mmrResp.Status == 200 {
			break
		}

		if mmrResp != nil && mmrResp.Status == 404 {
			log.Printf("[ValorantAPI] Player not found: %s#%s in region %s", name, tag, region)
			return nil, exception.ErrValorantPlayerNotFound
		}

		if mmrResp != nil && mmrResp.Status == 429 {
			log.Printf("[ValorantAPI] Rate limited on GetMMRByName(%s#%s), attempt %d/%d",
				name, tag, attempt, c.maxRetries)
			continue
		}

		if err != nil {
			log.Printf("[ValorantAPI] Error on GetMMRByName(%s#%s): %v (attempt %d/%d)",
				name, tag, err, attempt, c.maxRetries)
		} else if mmrResp != nil {
			log.Printf("[ValorantAPI] Non-200 status %d on GetMMRByName(%s#%s) (attempt %d/%d)",
				mmrResp.Status, name, tag, attempt, c.maxRetries)
		}
	}

	if err != nil {
		log.Printf("[ValorantAPI] All retries exhausted for GetMMRByName(%s#%s), last error: %v", name, tag, err)
		return nil, exception.ErrValorantApiError
	}
	if mmrResp == nil || mmrResp.Status != 200 {
		status := 0
		if mmrResp != nil {
			status = mmrResp.Status
		}
		if status == 429 {
			log.Printf("[ValorantAPI] Rate limit exhausted for GetMMRByName(%s#%s) after %d retries",
				name, tag, c.maxRetries)
			return nil, exception.ErrValorantApiRateLimit
		}
		log.Printf("[ValorantAPI] Failed GetMMRByName(%s#%s), final status: %d", name, tag, status)
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
	if err != nil {
		log.Printf("[ValorantAPI] Failed to fetch MMR history for %s#%s: %v, using current rank as peak", name, tag, err)
		return currentTier, currentTierPatched
	}
	if historyResp.Status != 200 {
		log.Printf("[ValorantAPI] MMR history returned status %d for %s#%s, using current rank as peak", historyResp.Status, name, tag)
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
