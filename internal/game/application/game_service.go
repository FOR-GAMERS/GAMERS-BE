package application

import (
	"GAMERS-BE/internal/game/application/dto"
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/exception"
)

type GameService struct {
	gameRepository port.GameDatabasePort
	teamRepository port.TeamDatabasePort
}

func NewGameService(
	gameRepository port.GameDatabasePort,
	teamRepository port.TeamDatabasePort,
) *GameService {
	return &GameService{
		gameRepository: gameRepository,
		teamRepository: teamRepository,
	}
}

// CreateGame creates a new game
func (s *GameService) CreateGame(req *dto.CreateGameRequest) (*domain.Game, error) {
	game := domain.NewGame(
		req.ContestID,
		req.GameTeamType,
		req.StartedAt,
		req.EndedAt,
	)

	if err := game.Validate(); err != nil {
		return nil, err
	}

	savedGame, err := s.gameRepository.Save(game)
	if err != nil {
		return nil, err
	}

	return savedGame, nil
}

// GetGame returns a game by ID
func (s *GameService) GetGame(gameID int64) (*domain.Game, error) {
	return s.gameRepository.GetByID(gameID)
}

// GetGamesByContest returns all games for a contest
func (s *GameService) GetGamesByContest(contestID int64) ([]*domain.Game, error) {
	return s.gameRepository.GetByContestID(contestID)
}

// UpdateGame updates a game (only in PENDING status)
func (s *GameService) UpdateGame(gameID int64, req *dto.UpdateGameRequest) (*domain.Game, error) {
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	if !req.HasChanges() {
		return game, nil
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	req.ApplyTo(game)

	if err := game.Validate(); err != nil {
		return nil, err
	}

	if err := s.gameRepository.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

// StartGame transitions the game to ACTIVE status
func (s *GameService) StartGame(gameID int64) (*domain.Game, error) {
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	if err := game.TransitionTo(domain.GameStatusActive); err != nil {
		return nil, err
	}

	if err := s.gameRepository.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

// FinishGame transitions the game to FINISHED status
func (s *GameService) FinishGame(gameID int64) (*domain.Game, error) {
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	if err := game.TransitionTo(domain.GameStatusFinished); err != nil {
		return nil, err
	}

	if err := s.gameRepository.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

// CancelGame transitions the game to CANCELLED status
func (s *GameService) CancelGame(gameID, userID int64) (*domain.Game, error) {
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	// Check if user is the leader
	member, err := s.teamRepository.GetByGameAndUser(gameID, userID)
	if err != nil {
		return nil, exception.ErrNotTeamMember
	}

	if !member.IsLeader() {
		return nil, exception.ErrNoPermissionToDelete
	}

	if err := game.TransitionTo(domain.GameStatusCancelled); err != nil {
		return nil, err
	}

	if err := s.gameRepository.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

// DeleteGame deletes a game (Leader only, only in PENDING status)
func (s *GameService) DeleteGame(gameID, userID int64) error {
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return err
	}

	if !game.IsPending() {
		return exception.ErrGameNotPending
	}

	// Check if user is the leader
	member, err := s.teamRepository.GetByGameAndUser(gameID, userID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	if !member.IsLeader() {
		return exception.ErrNoPermissionToDelete
	}

	// Delete all team members first
	if err := s.teamRepository.DeleteAllByGameID(gameID); err != nil {
		return err
	}

	// Delete the game
	return s.gameRepository.Delete(gameID)
}
