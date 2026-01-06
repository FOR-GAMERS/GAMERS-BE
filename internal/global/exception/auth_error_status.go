package exception

import (
	"net/http"
)

var (
	ErrInvalidAccessToken  = NewBusinessError(http.StatusUnauthorized, "Invalid access token", "AU001")
	ErrInvalidCredentials  = NewBusinessError(http.StatusUnauthorized, "Invalid credentials", "AU002")
	ErrInvalidRefreshToken = NewBusinessError(http.StatusUnauthorized, "Invalid refresh token", "AU003")

	ErrPasswordMismatch = NewBadRequestError("Password Mismatch", "AU004")
	ErrEmailMismatch    = NewNotFoundError("No user", "AU005")

	ErrUnauthorized        = NewBusinessError(http.StatusUnauthorized, "Unauthorized access", "AU006")
	ErrInternalServerError = NewBusinessError(http.StatusInternalServerError, "Internal server error", "AU007")
)
