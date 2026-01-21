package application

import (
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/exception"
	"math"
	"math/rand"
	"time"
)

// TournamentService handles tournament bracket generation and management
type TournamentService struct {
	gameRepository port.GameDatabasePort
	teamRepository port.TeamDatabasePort
}

// NewTournamentService creates a new tournament service
func NewTournamentService(
	gameRepository port.GameDatabasePort,
	teamRepository port.TeamDatabasePort,
) *TournamentService {
	return &TournamentService{
		gameRepository: gameRepository,
		teamRepository: teamRepository,
	}
}

// GenerateTournamentBracket creates all games needed for a tournament bracket
// For a tournament with N teams, we need N-1 games
// Returns the created games in round order
func (s *TournamentService) GenerateTournamentBracket(
	contestID int64,
	maxTeamCount int,
	gameTeamType domain.GameTeamType,
) ([]*domain.Game, error) {
	// Validate maxTeamCount is a power of 2
	if !isPowerOfTwo(maxTeamCount) {
		return nil, exception.ErrMaxTeamCountNotPowerOfTwo
	}

	// Check if games already exist for this contest
	existingGames, err := s.gameRepository.GetByContestID(contestID)
	if err != nil {
		return nil, err
	}
	if len(existingGames) > 0 {
		return nil, exception.ErrTournamentGamesAlreadyExist
	}

	numRounds := int(math.Log2(float64(maxTeamCount)))

	games := make([]*domain.Game, 0, maxTeamCount-1)
	bracketPosition := 0

	for round := 1; round <= numRounds; round++ {
		matchesInRound := maxTeamCount / int(math.Pow(2, float64(round)))

		for match := 1; match <= matchesInRound; match++ {
			bracketPosition++
			game := domain.NewTournamentGame(
				contestID,
				gameTeamType,
				round,
				match,
				bracketPosition,
			)

			savedGame, err := s.gameRepository.Save(game)
			if err != nil {
				return nil, err
			}
			games = append(games, savedGame)
		}
	}

	if err := s.linkTournamentGames(games, numRounds, maxTeamCount); err != nil {
		return nil, err
	}

	return games, nil
}

// linkTournamentGames links each game to the next game (where winner advances)
func (s *TournamentService) linkTournamentGames(games []*domain.Game, numRounds, maxTeamCount int) error {
	// Create a map of (round, match) -> game for easy lookup
	gameMap := make(map[string]*domain.Game)
	for _, game := range games {
		key := getGameKey(game.GetRound(), game.GetMatchNumber())
		gameMap[key] = game
	}

	// Link each game to its next game
	for round := 1; round < numRounds; round++ {
		matchesInRound := maxTeamCount / int(math.Pow(2, float64(round)))

		for match := 1; match <= matchesInRound; match++ {
			currentKey := getGameKey(round, match)
			currentGame := gameMap[currentKey]

			// The winner of matches 1&2 goes to next round match 1
			// The winner of matches 3&4 goes to next round match 2
			// etc.
			nextMatch := (match + 1) / 2
			nextKey := getGameKey(round+1, nextMatch)
			nextGame := gameMap[nextKey]

			if nextGame != nil {
				currentGame.SetNextGame(nextGame.GameID)
				if err := s.gameRepository.Update(currentGame); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// GetTournamentBracket retrieves all games for a contest organized by round
func (s *TournamentService) GetTournamentBracket(contestID int64) (*TournamentBracket, error) {
	games, err := s.gameRepository.GetByContestID(contestID)
	if err != nil {
		return nil, err
	}

	bracket := &TournamentBracket{
		ContestID: contestID,
		Rounds:    make(map[int][]*domain.Game),
	}

	for _, game := range games {
		if game.IsTournamentGame() {
			round := game.GetRound()
			bracket.Rounds[round] = append(bracket.Rounds[round], game)
		}
	}

	// Calculate total rounds
	if len(bracket.Rounds) > 0 {
		for round := range bracket.Rounds {
			if round > bracket.TotalRounds {
				bracket.TotalRounds = round
			}
		}
	}

	return bracket, nil
}

// TournamentBracket represents the tournament bracket structure
type TournamentBracket struct {
	ContestID   int64
	TotalRounds int
	Rounds      map[int][]*domain.Game // round number -> games in that round
}

// GetRoundName returns a human-readable name for the round
func GetRoundName(round, totalRounds int) string {
	roundsFromEnd := totalRounds - round

	switch roundsFromEnd {
	case 0:
		return "Final"
	case 1:
		return "Semi-finals"
	case 2:
		return "Quarter-finals"
	default:
		// Calculate the round number (e.g., Round of 16, Round of 32)
		teamsInRound := int(math.Pow(2, float64(roundsFromEnd+1)))
		return "Round of " + intToString(teamsInRound)
	}
}

// ShuffleAndAllocateTeams shuffles all registered teams and assigns them to first round games
// This should be called when the contest starts or when team recruitment is complete
func (s *TournamentService) ShuffleAndAllocateTeams(contestID int64) error {
	// Get all teams for the contest
	teams, err := s.teamRepository.GetByContestID(contestID)
	if err != nil {
		return err
	}

	if len(teams) == 0 {
		return exception.ErrNoTeamsToAllocate
	}

	// Get first round games
	firstRoundGames, err := s.gameRepository.GetByContestAndRound(contestID, 1)
	if err != nil {
		return err
	}

	if len(firstRoundGames) == 0 {
		return exception.ErrNoGamesToAllocate
	}

	// Shuffle teams using Fisher-Yates algorithm
	shuffledTeams := shuffleTeams(teams)

	// Assign teams to first round games
	// Each game needs 2 teams
	teamIndex := 0
	for _, game := range firstRoundGames {
		if teamIndex+1 >= len(shuffledTeams) {
			break // Not enough teams for this game
		}

		// Assign two teams to this game
		team1 := shuffledTeams[teamIndex]
		team2 := shuffledTeams[teamIndex+1]

		// Create game_team entries
		gameTeam1 := domain.NewGameTeam(game.GameID, team1.TeamID)
		gameTeam2 := domain.NewGameTeam(game.GameID, team2.TeamID)

		// Note: We would need a GameTeamDatabasePort to save these
		// For now, we'll return the allocation info
		_ = gameTeam1
		_ = gameTeam2

		teamIndex += 2
	}

	return nil
}

// ShuffleAndAllocateTeamsWithResult shuffles teams and returns the allocation result
func (s *TournamentService) ShuffleAndAllocateTeamsWithResult(contestID int64, gameTeamRepo port.GameTeamDatabasePort) (*TeamAllocationResult, error) {
	// Get all teams for the contest
	teams, err := s.teamRepository.GetByContestID(contestID)
	if err != nil {
		return nil, err
	}

	if len(teams) == 0 {
		return nil, exception.ErrNoTeamsToAllocate
	}

	// Get first round games
	firstRoundGames, err := s.gameRepository.GetByContestAndRound(contestID, 1)
	if err != nil {
		return nil, err
	}

	if len(firstRoundGames) == 0 {
		return nil, exception.ErrNoGamesToAllocate
	}

	// Validate team count matches expected
	expectedTeams := len(firstRoundGames) * 2
	if len(teams) < expectedTeams {
		return nil, exception.ErrNotEnoughTeams
	}

	// Shuffle teams using Fisher-Yates algorithm
	shuffledTeams := shuffleTeams(teams)

	result := &TeamAllocationResult{
		ContestID:   contestID,
		TotalTeams:  len(teams),
		Allocations: make([]GameAllocation, 0, len(firstRoundGames)),
	}

	// Assign teams to first round games
	teamIndex := 0
	for _, game := range firstRoundGames {
		team1 := shuffledTeams[teamIndex]
		team2 := shuffledTeams[teamIndex+1]

		// Create and save game_team entries
		gameTeam1 := domain.NewGameTeam(game.GameID, team1.TeamID)
		gameTeam2 := domain.NewGameTeam(game.GameID, team2.TeamID)

		if _, err := gameTeamRepo.Save(gameTeam1); err != nil {
			return nil, err
		}
		if _, err := gameTeamRepo.Save(gameTeam2); err != nil {
			return nil, err
		}

		result.Allocations = append(result.Allocations, GameAllocation{
			GameID:      game.GameID,
			Round:       game.GetRound(),
			MatchNumber: game.GetMatchNumber(),
			Team1ID:     team1.TeamID,
			Team1Name:   team1.TeamName,
			Team2ID:     team2.TeamID,
			Team2Name:   team2.TeamName,
		})

		teamIndex += 2
	}

	return result, nil
}

// TeamAllocationResult represents the result of team allocation
type TeamAllocationResult struct {
	ContestID   int64
	TotalTeams  int
	Allocations []GameAllocation
}

// GameAllocation represents a single game allocation
type GameAllocation struct {
	GameID      int64
	Round       int
	MatchNumber int
	Team1ID     int64
	Team1Name   string
	Team2ID     int64
	Team2Name   string
}

// shuffleTeams shuffles the teams using Fisher-Yates algorithm
func shuffleTeams(teams []*domain.Team) []*domain.Team {
	result := make([]*domain.Team, len(teams))
	copy(result, teams)

	// Seed the random number generator with current time
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Fisher-Yates shuffle
	for i := len(result) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// Helper functions

func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

func getGameKey(round, match int) string {
	return intToString(round) + "-" + intToString(match)
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	result := ""
	for n > 0 {
		digit := n % 10
		result = string(rune('0'+digit)) + result
		n /= 10
	}
	return result
}
