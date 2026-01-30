package application

import (
	contestPort "GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/game/application/dto"
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/domain"
	"log"
	"sort"
)

// TournamentResultService handles retrieving tournament results
type TournamentResultService struct {
	gameDBPort      port.GameDatabasePort
	gameTeamDBPort  port.GameTeamDatabasePort
	teamDBPort      port.TeamDatabasePort
	matchResultPort port.MatchResultDatabasePort
	contestDBPort   contestPort.ContestDatabasePort
}

func NewTournamentResultService(
	gameDBPort port.GameDatabasePort,
	gameTeamDBPort port.GameTeamDatabasePort,
	teamDBPort port.TeamDatabasePort,
	matchResultPort port.MatchResultDatabasePort,
	contestDBPort contestPort.ContestDatabasePort,
) *TournamentResultService {
	return &TournamentResultService{
		gameDBPort:      gameDBPort,
		gameTeamDBPort:  gameTeamDBPort,
		teamDBPort:      teamDBPort,
		matchResultPort: matchResultPort,
		contestDBPort:   contestDBPort,
	}
}

// SetContestDBPort sets the contest database port (to resolve circular dependency)
func (s *TournamentResultService) SetContestDBPort(port contestPort.ContestDatabasePort) {
	s.contestDBPort = port
}

// GetContestResult returns the full tournament bracket and results for a contest
func (s *TournamentResultService) GetContestResult(contestID int64) (*dto.ContestResultResponse, error) {
	contest, err := s.contestDBPort.GetContestById(contestID)
	if err != nil {
		return nil, err
	}

	games, err := s.gameDBPort.GetByContestID(contestID)
	if err != nil {
		return nil, err
	}

	// Build team name lookup
	teams, err := s.teamDBPort.GetByContestID(contestID)
	if err != nil {
		return nil, err
	}
	teamNameMap := make(map[int64]string, len(teams))
	for _, t := range teams {
		teamNameMap[t.TeamID] = t.TeamName
	}

	// Group games by round and find total rounds
	roundGames := make(map[int][]*domain.Game)
	totalRounds := 0
	for _, g := range games {
		if !g.IsTournamentGame() {
			continue
		}
		round := g.GetRound()
		roundGames[round] = append(roundGames[round], g)
		if round > totalRounds {
			totalRounds = round
		}
	}

	// Build rounds
	rounds := make([]dto.RoundResult, 0, totalRounds)
	for round := 1; round <= totalRounds; round++ {
		gamesInRound := roundGames[round]
		sort.Slice(gamesInRound, func(i, j int) bool {
			return gamesInRound[i].GetMatchNumber() < gamesInRound[j].GetMatchNumber()
		})

		gameResults := make([]dto.GameResult, 0, len(gamesInRound))
		for _, g := range gamesInRound {
			gr := s.buildGameResult(g, teamNameMap)
			gameResults = append(gameResults, gr)
		}

		rounds = append(rounds, dto.RoundResult{
			Round:     round,
			RoundName: GetRoundName(round, totalRounds),
			Games:     gameResults,
		})
	}

	// Determine champion from the final round
	var champion *dto.TeamSummary
	if totalRounds > 0 {
		finalGames := roundGames[totalRounds]
		if len(finalGames) == 1 && finalGames[0].GameStatus == domain.GameStatusFinished {
			result, err := s.matchResultPort.GetByGameID(finalGames[0].GameID)
			if err == nil && result != nil {
				champion = &dto.TeamSummary{
					TeamID:   result.WinnerTeamID,
					TeamName: teamNameMap[result.WinnerTeamID],
				}
			}
		}
	}

	return &dto.ContestResultResponse{
		ContestID:     contest.ContestID,
		Title:         contest.Title,
		ContestStatus: string(contest.ContestStatus),
		TotalRounds:   totalRounds,
		Champion:      champion,
		Rounds:        rounds,
	}, nil
}

// buildGameResult constructs a GameResult for a single game
func (s *TournamentResultService) buildGameResult(game *domain.Game, teamNameMap map[int64]string) dto.GameResult {
	gr := dto.GameResult{
		GameID:          game.GameID,
		MatchNumber:     game.GetMatchNumber(),
		GameStatus:      string(game.GameStatus),
		DetectionStatus: string(game.DetectionStatus),
		Teams:           make([]dto.GameTeamResult, 0),
	}

	// Get game teams
	gameTeams, err := s.gameTeamDBPort.GetByGameID(game.GameID)
	if err != nil {
		log.Printf("[TournamentResult] Failed to get game teams for game %d: %v", game.GameID, err)
	} else {
		for _, gt := range gameTeams {
			gr.Teams = append(gr.Teams, dto.GameTeamResult{
				TeamID:   gt.TeamID,
				TeamName: teamNameMap[gt.TeamID],
				Grade:    gt.Grade,
			})
		}
	}

	// Get match result for finished games
	if game.GameStatus == domain.GameStatusFinished {
		result, err := s.matchResultPort.GetByGameID(game.GameID)
		if err == nil && result != nil {
			gr.MatchResult = &dto.MatchResultSummary{
				WinnerTeamID: result.WinnerTeamID,
				LoserTeamID:  result.LoserTeamID,
				WinnerScore:  result.WinnerScore,
				LoserScore:   result.LoserScore,
				MapName:      result.MapName,
			}
		}
	}

	return gr
}
