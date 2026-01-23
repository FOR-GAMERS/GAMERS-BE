package exception

import "net/http"

var (
	ErrBannerNotFound      = NewNotFoundError("banner not found", "BN001")
	ErrBannerLimitExceeded = NewBusinessError(http.StatusConflict, "maximum banner limit (5) exceeded", "BN002")
	ErrAdminRequired       = NewBusinessError(http.StatusForbidden, "admin privileges required", "BN003")
	ErrInvalidBannerID     = NewBadRequestError("invalid banner id", "BN004")
	ErrInvalidDisplayOrder = NewBadRequestError("invalid display order", "BN005")
)
