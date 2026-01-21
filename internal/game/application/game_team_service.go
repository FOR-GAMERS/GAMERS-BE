package application

import (
	contestPort "GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/game/application/dto"
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/exception"
)

type GameTeamService struct {
	gameTeamRepository port.GameTeamDatabasePort
	gameRepository     port.GameDatabasePort
	teamRepository     port.TeamDatabasePort
	contestRepository  contestPort.ContestDatabasePort
}

func NewGameTeamService(
	gameTeamRepository port.GameTeamDatabasePort,
	gameRepository port.GameDatabasePort,
	teamRepository port.TeamDatabasePort,
	contestRepository contestPort.ContestDatabasePort,
) *GameTeamService {
	return &GameTeamService{
		gameTeamRepository: gameTeamRepository,
		gameRepository:     gameRepository,
		teamRepository:     teamRepository,
		contestRepository:  contestRepository,
	}
}

// CreateGameTeam creates a new game-team relationship with grade validation
func (s *GameTeamService) CreateGameTeam(req *dto.CreateGameTeamRequest) (*domain.GameTeam, error) {
	// Validate game exists
	game, err := s.gameRepository.GetByID(req.GameID)
	if err != nil {
		return nil, err
	}

	// Validate team exists and belongs to the same contest as the game
	team, err := s.teamRepository.GetByID(req.TeamID)
	if err != nil {
		return nil, err
	}

	if team.ContestID != game.ContestID {
		return nil, exception.ErrTeamMemberNotFound
	}

	// Check if game team already exists
	_, err = s.gameTeamRepository.GetByGameAndTeam(req.GameID, req.TeamID)
	if err == nil {
		return nil, exception.ErrGameTeamAlreadyExists
	}

	// Get contest to validate maxTeamCount
	contest, err := s.contestRepository.GetContestById(game.ContestID)
	if err != nil {
		return nil, err
	}

	// Create game team
	var gameTeam *domain.GameTeam
	if req.Grade != nil {
		gameTeam = domain.NewGameTeamWithGrade(req.GameID, req.TeamID, *req.Grade)
	} else {
		gameTeam = domain.NewGameTeam(req.GameID, req.TeamID)
	}

	// Validate with maxTeamCount constraint
	if err := gameTeam.ValidateWithMaxTeamCount(contest.MaxTeamCount); err != nil {
		return nil, err
	}

	// Check for duplicate grade in the same game
	if req.Grade != nil {
		existing, err := s.gameTeamRepository.GetByGrade(req.GameID, *req.Grade)
		if err == nil && existing != nil {
			return nil, exception.ErrDuplicateGradeInGame
		}
	}

	return s.gameTeamRepository.Save(gameTeam)
}

// GetGameTeam returns a game team by ID
func (s *GameTeamService) GetGameTeam(gameTeamID int64) (*domain.GameTeam, error) {
	return s.gameTeamRepository.GetByID(gameTeamID)
}

// GetGameTeamsByGame returns all game teams for a game
func (s *GameTeamService) GetGameTeamsByGame(gameID int64) ([]*domain.GameTeam, error) {
	return s.gameTeamRepository.GetByGameID(gameID)
}

// GetGameTeamByTeam returns a game team for a specific team in a game
func (s *GameTeamService) GetGameTeamByTeam(gameID, teamID int64) (*domain.GameTeam, error) {
	return s.gameTeamRepository.GetByGameAndTeam(gameID, teamID)
}

// DeleteGameTeam deletes a game team by ID
func (s *GameTeamService) DeleteGameTeam(gameTeamID int64) error {
	return s.gameTeamRepository.Delete(gameTeamID)
}

// DeleteGameTeamsByGame deletes all game teams for a game
func (s *GameTeamService) DeleteGameTeamsByGame(gameID int64) error {
	return s.gameTeamRepository.DeleteByGameID(gameID)
}
