package exception

import "net/http"

var (
	ErrRedisCannotSave        = NewBadRequestError("failed to marshal refresh token", "RE001")
	ErrRedisCannotSet         = NewBadRequestError("failed to add token to user set", "RE002")
	ErrRedisCannotDelete      = NewBadRequestError("failed to delete token from user set", "RE003")
	ErrRedisCannotSetDuration = NewBadRequestError("failed to set expiration on user tokens", "RE004")
	ErrRedisCannotGetToken    = NewBadRequestError("failed to get token from user set", "RE005")
	ErrRedisCannotFindToken   = NewBusinessError(http.StatusNotFound, "token not found", "RE006")
	ErrTokenExpired           = NewBusinessError(http.StatusUnauthorized, "token expired", "OA006")
	ErrRedisSetError          = NewBadRequestError("failed to set value in redis", "RE007")
	ErrRedisGetError          = NewBadRequestError("failed to get value from redis", "RE008")
	ErrRedisDeleteError       = NewBadRequestError("failed to delete value from redis", "RE009")
)
