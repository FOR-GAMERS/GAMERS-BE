package exception

import "net/http"

var (
	ErrCommentNotFound         = NewNotFoundError("comment not found", "CM001")
	ErrCommentContentEmpty     = NewBadRequestError("comment content is required", "CM002")
	ErrCommentContentTooLong   = NewBadRequestError("comment content exceeds 255 characters", "CM003")
	ErrCommentPermissionDenied = NewBusinessError(http.StatusForbidden, "you can only modify your own comments", "CM004")
)
