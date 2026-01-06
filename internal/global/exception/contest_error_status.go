package exception

import (
	"net/http"
)

var (
	ErrContestNotFound         = NewBusinessError(http.StatusNotFound, "contest not found", "CT001")
	ErrInvalidStatusTransition = NewBusinessError(http.StatusConflict, "invalid contest status transition", "CT002")
	ErrContestMemberNotFound   = NewBusinessError(http.StatusNotFound, "contest member not found", "CT003")
	ErrUserNotExists           = NewBusinessError(http.StatusNotFound, "user or invitation not found", "CT004")
	ErrContestAlreadyExists    = NewBusinessError(http.StatusConflict, "contest already exists", "CT005")
	ErrContestNoChanges        = NewBadRequestError("contest has no changes", "CT006")
	ErrInvalidContestDates     = NewBadRequestError("contest start date must be before end date", "CT007")
	ErrInvalidContestType      = NewBadRequestError("invalid contest type", "CT008")
	ErrInvalidContestStatus    = NewBadRequestError("invalid contest status", "CT009")
	ErrInvalidMaxTeamCount     = NewBadRequestError("max team count must be positive", "CT010")
	ErrInvalidTotalPoint       = NewBadRequestError("total point must be non-negative", "CT011")
	ErrInvalidContestTitle     = NewBadRequestError("contest title is required", "CT012")
)
