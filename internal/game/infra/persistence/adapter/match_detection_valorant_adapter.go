package adapter

import (
	"GAMERS-BE/internal/game/application/port"
	"fmt"
	"log"
	"time"

	govapi "github.com/yldshv/go-valorant-api"
)

// MatchDetectionValorantAdapter implements MatchDetectionPort using the Henrik Dev Valorant API
type MatchDetectionValorantAdapter struct {
	client     *govapi.VAPI
	maxRetries int
}

func NewMatchDetectionValorantAdapter(apiKey string) *MatchDetectionValorantAdapter {
	var client *govapi.VAPI
	if apiKey != "" {
		client = govapi.New(govapi.WithKey(apiKey))
	} else {
		client = govapi.New()
	}
	return &MatchDetectionValorantAdapter{
		client:     client,
		maxRetries: 3,
	}
}

// GetRecentMatches fetches recent match history for a player via VAPI
func (a *MatchDetectionValorantAdapter) GetRecentMatches(region, name, tag string) ([]port.ValorantMatch, error) {
	var resp *govapi.GetMatchesByNameV3Response
	var err error

	// Retry with exponential backoff for rate limit handling
	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("[VAPI] Retry attempt %d for GetMatchesByName(%s#%s), waiting %v",
				attempt, name, tag, backoff)
			time.Sleep(backoff)
		}

		resp, err = a.client.GetMatchesByNameV3(govapi.GetMatchesByNameV3Params{
			Affinity: region,
			Name:     name,
			Tag:      tag,
		})
		if err == nil && resp != nil && resp.Status == 200 {
			break
		}

		// If rate limited (429), retry
		if resp != nil && resp.Status == 429 {
			log.Printf("[VAPI] Rate limited, will retry")
			continue
		}
	}

	if err != nil {
		return nil, fmt.Errorf("VAPI GetMatchesByName failed: %w", err)
	}
	if resp == nil || resp.Status != 200 {
		status := 0
		if resp != nil {
			status = resp.Status
		}
		return nil, fmt.Errorf("VAPI returned non-200 status: %d", status)
	}

	var matches []port.ValorantMatch
	for _, m := range resp.Data {
		gameStart := time.Unix(int64(m.Metadata.GameStart), 0).UTC()
		matches = append(matches, port.ValorantMatch{
			MatchID:    m.Metadata.Matchid,
			MapName:    m.Metadata.Map,
			GameMode:   m.Metadata.Mode,
			GameStart:  gameStart,
			GameLength: m.Metadata.GameLength,
		})
	}

	return matches, nil
}

// GetMatchDetail fetches full match details by match ID
func (a *MatchDetectionValorantAdapter) GetMatchDetail(matchID string) (*port.ValorantMatchDetail, error) {
	var resp *govapi.GetMatchResponse
	var err error

	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		resp, err = a.client.GetMatch(govapi.GetMatchParams{
			MatchId: matchID,
		})
		if err == nil && resp != nil && resp.Status == 200 {
			break
		}
		if resp != nil && resp.Status == 429 {
			continue
		}
	}

	if err != nil {
		return nil, fmt.Errorf("VAPI GetMatch failed: %w", err)
	}
	if resp == nil || resp.Status != 200 {
		status := 0
		if resp != nil {
			status = resp.Status
		}
		return nil, fmt.Errorf("VAPI match detail returned non-200: %d", status)
	}

	detail := &port.ValorantMatchDetail{
		MatchID:      resp.Data.Metadata.Matchid,
		MapName:      resp.Data.Metadata.Map,
		GameMode:     resp.Data.Metadata.Mode,
		GameStart:    time.Unix(int64(resp.Data.Metadata.GameStart), 0).UTC(),
		GameLength:   resp.Data.Metadata.GameLength,
		RoundsPlayed: resp.Data.Metadata.RoundsPlayed,
	}

	// Parse teams — govapi uses struct { Red, Blue } not an array
	detail.Teams = []port.ValorantTeamData{
		{
			TeamID:    "Red",
			HasWon:    resp.Data.Teams.Red.HasWon,
			RoundsWon: resp.Data.Teams.Red.RoundsWon,
		},
		{
			TeamID:    "Blue",
			HasWon:    resp.Data.Teams.Blue.HasWon,
			RoundsWon: resp.Data.Teams.Blue.RoundsWon,
		},
	}

	// Parse players
	for _, p := range resp.Data.Players.AllPlayers {
		detail.Players = append(detail.Players, port.ValorantPlayerData{
			PUUID:     p.Puuid,
			Name:      p.Name,
			Tag:       p.Tag,
			TeamID:    p.Team,
			Agent:     p.Character,
			Kills:     p.Stats.Kills,
			Deaths:    p.Stats.Deaths,
			Assists:   p.Stats.Assists,
			Score:     p.Stats.Score,
			Headshots: p.Stats.Headshots,
			Bodyshots: p.Stats.Bodyshots,
			Legshots:  p.Stats.Legshots,
		})
	}

	return detail, nil
}
