package exception

import (
	"net/http"
)

var (
	ErrStateNotFound = NewBusinessError(http.StatusNotFound, "invalid state: not found", "ST001")
	ErrStateExpired  = NewBusinessError(http.StatusUnauthorized, "invalid state: expired", "ST002")
)
