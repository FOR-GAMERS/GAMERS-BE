package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiResponse[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

func Success[T any](data T, message string) *ApiResponse[T] {
	return &ApiResponse[T]{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	}
}

func Created[T any](data T, message string) *ApiResponse[T] {
	return &ApiResponse[T]{
		Status:  http.StatusCreated,
		Message: message,
		Data:    data,
	}
}

func NoContent(message string) *ApiResponse[any] {
	return &ApiResponse[any]{
		Status:  http.StatusNoContent,
		Message: message,
		Data:    nil,
	}
}

func Error(status int, message string) *ApiResponse[any] {
	return &ApiResponse[any]{
		Status:  status,
		Message: message,
		Data:    nil,
	}
}

func BadRequest(message string) *ApiResponse[any] {
	return Error(http.StatusBadRequest, message)
}

func NotFound(message string) *ApiResponse[any] {
	return Error(http.StatusNotFound, message)
}

func Conflict(message string) *ApiResponse[any] {
	return Error(http.StatusConflict, message)
}

func InternalServerError(message string) *ApiResponse[any] {
	return Error(http.StatusInternalServerError, message)
}

func JSON[T any](ctx *gin.Context, response *ApiResponse[T]) {
	ctx.JSON(response.Status, response)
}
