package exception

import "net/http"

var (
	ErrDBConnection = NewBusinessError(http.StatusInternalServerError, "config connection error", "GL001")
)
