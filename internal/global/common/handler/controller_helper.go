package handler

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ControllerHelper struct {
}

func NewControllerHelper() *ControllerHelper {
	return &ControllerHelper{}
}

func (h *ControllerHelper) BindJSON(ctx *gin.Context, req interface{}) bool {
	if err := ctx.ShouldBindJSON(req); err != nil {
		response.JSON(ctx, response.BadRequest(err.Error()))
		return false
	}
	return true
}

func (h *ControllerHelper) RespondWithData(
	ctx *gin.Context,
	data interface{},
	err error,
	successStatus int,
	successMsg string,
) {
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	ctx.JSON(successStatus, response.Response{
		Status:  successStatus,
		Message: successMsg,
		Data:    data,
	})
}

// RespondCreated is a shorthand for 201 Created
func (h *ControllerHelper) RespondCreated(ctx *gin.Context, data interface{}, err error, msg string) {
	h.RespondWithData(ctx, data, err, http.StatusCreated, msg)
}

// RespondOK is a shorthand for 200 OK
func (h *ControllerHelper) RespondOK(ctx *gin.Context, data interface{}, err error, msg string) {
	h.RespondWithData(ctx, data, err, http.StatusOK, msg)
}

// RespondNoContent is for 204 No Content (delete operations)
func (h *ControllerHelper) RespondNoContent(ctx *gin.Context, err error) {
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (h *ControllerHelper) handleError(ctx *gin.Context, err error) {
	var businessErr *exception.BusinessError
	if errors.As(err, &businessErr) {
		ctx.JSON(businessErr.Status, businessErr)
		return
	}

	response.JSON(ctx, response.InternalServerError("something went wrong"))
}
