package exception

import "net/http"

type BusinessError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *BusinessError) Error() string {
	return e.Message
}

func NewBusinessError(status int, message, code string) *BusinessError {
	return &BusinessError{
		Status:  status,
		Message: message,
		Code:    code,
	}
}

func NewNotFoundError(message, code string) *BusinessError {
	return &BusinessError{
		Status:  http.StatusNotFound,
		Message: message,
		Code:    code,
	}
}

func NewBadRequestError(message, code string) *BusinessError {
	return &BusinessError{
		Status:  http.StatusBadRequest,
		Message: message,
		Code:    code,
	}
}
