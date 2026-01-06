package exception

import "net/http"

var (
	ErrDBConnection = NewBusinessError(http.StatusInternalServerError, "database connection error", "GL001")
)
