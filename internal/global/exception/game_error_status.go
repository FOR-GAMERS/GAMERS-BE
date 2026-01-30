package exception

import (
	"net/http"
)

var (
	// Game errors
	ErrGameNotFound                = NewBusinessError(http.StatusNotFound, "game not found", "GM001")
	ErrInvalidGameStatusTransition = NewBusinessError(http.StatusConflict, "invalid game status transition", "GM002")
	ErrInvalidGameStatus           = NewBadRequestError("invalid game status", "GM003")
	ErrInvalidGameTeamType         = NewBadRequestError("invalid game team type", "GM004")
	ErrInvalidGameDates            = NewBadRequestError("game start date must be before end date", "GM005")
	ErrGameDurationExceeded        = NewBadRequestError("game duration cannot exceed 2 hours", "GM006")
	ErrGameStartTimeRequired       = NewBadRequestError("game start time is required", "GM007")
	ErrGameEndTimeRequired         = NewBadRequestError("game end time is required", "GM008")
	ErrGameNotPending              = NewBadRequestError("game is not in pending status", "GM009")
	ErrCannotInviteToGame          = NewBadRequestError("cannot invite to game when game is not pending", "GM010")
	ErrInvalidGameID               = NewBadRequestError("invalid game id", "GM011")
	ErrInvalidTournamentRound      = NewBadRequestError("invalid tournament round", "GM012")
	ErrInvalidMatchNumber          = NewBadRequestError("invalid match number", "GM013")
	ErrMaxTeamCountNotPowerOfTwo   = NewBadRequestError("max team count must be a power of 2 for tournament brackets", "GM014")
	ErrTournamentGamesAlreadyExist = NewBusinessError(http.StatusConflict, "tournament games already exist for this contest", "GM015")
	ErrNoTeamsToAllocate           = NewBadRequestError("no teams registered for allocation", "GM016")
	ErrNoGamesToAllocate           = NewBadRequestError("no first round games found for allocation", "GM017")
	ErrNotEnoughTeams              = NewBadRequestError("not enough teams registered for tournament", "GM018")

	// Team errors
	ErrTeamNotFound            = NewBusinessError(http.StatusNotFound, "team not found", "TM001")
	ErrTeamMemberNotFound      = NewBusinessError(http.StatusNotFound, "team member not found", "TM002")
	ErrInvalidTeamMemberType   = NewBadRequestError("invalid team member type", "TM003")
	ErrTeamMemberAlreadyExists = NewBusinessError(http.StatusConflict, "user is already a member of this team", "TM004")
	ErrNotTeamMember           = NewBusinessError(http.StatusForbidden, "you are not a member of this team", "TM005")
	ErrNoPermissionToKick      = NewBusinessError(http.StatusForbidden, "only leader can kick members", "TM006")
	ErrNoPermissionToDelete    = NewBusinessError(http.StatusForbidden, "only leader can delete the team", "TM007")
	ErrCannotKickLeader        = NewBadRequestError("cannot kick the leader", "TM008")
	ErrCannotLeaveAsLeader     = NewBadRequestError("leader cannot leave the team, transfer leadership or delete the team", "TM009")
	ErrTeamIsFull              = NewBadRequestError("team has reached maximum member limit", "TM010")
	ErrNoPermissionToInvite    = NewBusinessError(http.StatusForbidden, "you do not have permission to invite", "TM011")
	ErrTeamInviteNotFound      = NewBusinessError(http.StatusNotFound, "team invite not found", "TM012")
	ErrTeamInviteNotPending    = NewBadRequestError("team invite is not pending", "TM013")
	ErrTeamAlreadyFinalized    = NewBadRequestError("team is already finalized", "TM014")
	ErrTeamNotReady            = NewBadRequestError("team has not reached maximum members", "TM015")
	ErrInvalidTeamName         = NewBadRequestError("team name is required", "TM016")
	ErrTeamNameTooLong         = NewBadRequestError("team name cannot exceed 50 characters", "TM017")
	ErrTeamNameAlreadyExists   = NewBusinessError(http.StatusConflict, "team name already exists in this contest", "TM018")

	// ScoreTable errors
	ErrScoreTableNotFound = NewBusinessError(http.StatusNotFound, "score table not found", "ST001")

	// Match Detection errors
	ErrInvalidDetectionStatusTransition = NewBusinessError(http.StatusConflict, "invalid detection status transition", "MD001")
	ErrScheduledTimeInPast              = NewBadRequestError("scheduled start time must be in the future", "MD002")
	ErrGameNotActive                    = NewBadRequestError("game is not in active status", "MD003")
	ErrDetectionNotFailed               = NewBadRequestError("manual result can only be set when detection has failed or is in progress", "MD004")
	ErrMatchResultNotFound              = NewBusinessError(http.StatusNotFound, "match result not found", "MD005")
	ErrMatchResultAlreadyExists         = NewBusinessError(http.StatusConflict, "match result already exists for this game", "MD006")
	ErrWinnerTeamNotInGame              = NewBadRequestError("winner team is not participating in this game", "MD007")
	ErrMissingValorantAccount           = NewBadRequestError("some team members have not linked their Valorant account", "MD008")
	ErrDetectionWindowExpired           = NewBadRequestError("detection window has expired", "MD009")
	ErrSchedulerLockFailed              = NewBusinessError(http.StatusConflict, "scheduler is already running on another instance", "MD010")

	// GameTeam errors
	ErrGameTeamNotFound         = NewBusinessError(http.StatusNotFound, "game team not found", "GT001")
	ErrGameTeamAlreadyExists    = NewBusinessError(http.StatusConflict, "team already exists in this game", "GT002")
	ErrInvalidTeamID            = NewBadRequestError("invalid team id", "GT003")
	ErrInvalidGradeMin          = NewBadRequestError("grade must be at least 1", "GT004")
	ErrGradeExceedsMaxTeamCount = NewBadRequestError("grade exceeds maximum team count", "GT005")
	ErrDuplicateGradeInGame     = NewBusinessError(http.StatusConflict, "duplicate grade in the same game", "GT006")
)
