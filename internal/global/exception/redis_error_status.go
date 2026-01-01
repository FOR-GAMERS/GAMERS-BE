package exception

var (
	ErrRedisCannotSave        = NewBadRequestError("failed to marshal refresh token", "RE001")
	ErrRedisCannotSet         = NewBadRequestError("failed to add token to user set", "RE002")
	ErrRedisCannotDelete      = NewBadRequestError("failed to delete token from user set", "RE003")
	ErrRedisCannotSetDuration = NewBadRequestError("failed to set expiration on user tokens", "RE004")
	ErrRedisCannotGetToken    = NewBadRequestError("failed to get token from user set", "RE005")
)
