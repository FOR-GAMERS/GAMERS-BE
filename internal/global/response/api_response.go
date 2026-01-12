package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

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

func SendNoContent(ctx *gin.Context) {
	// 204 No Content는 본문이 없어야 함
	ctx.Status(http.StatusNoContent)
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

func Forbidden[T any](data T, message string) *ApiResponse[T] {
	return &ApiResponse[T]{
		Status:  http.StatusForbidden,
		Message: message,
		Data:    data,
	}
}

func JSON[T any](ctx *gin.Context, response *ApiResponse[T]) {
	ctx.JSON(response.Status, response)
}
