package exception

import "net/http"

var (
	ErrValorantNotLinked      = NewBadRequestError("Valorant account is not linked", "VAL001")
	ErrValorantApiError       = NewBusinessError(http.StatusBadGateway, "Failed to fetch data from Valorant API", "VAL002")
	ErrInvalidValorantRegion  = NewBadRequestError("Invalid region. Valid regions: ap, br, eu, kr, latam, na", "VAL003")
	ErrValorantAlreadyLinked  = NewBusinessError(http.StatusConflict, "Valorant account is already linked", "VAL004")
	ErrValorantPlayerNotFound = NewNotFoundError("Valorant player not found", "VAL005")
)
