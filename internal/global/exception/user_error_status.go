package exception

import (
	"net/http"
)

var (
	ErrUserNotFound      = NewNotFoundError("User not found", "US001")
	ErrUserAlreadyExists = NewBusinessError(http.StatusConflict, "user already exists", "US002")

	ErrUsernameEmpty       = NewBadRequestError("Username cannot be empty", "US003")
	ErrUsernameTooLong     = NewBadRequestError("Username must be 16 characters or less", "US004")
	ErrUsernameInvalidChar = NewBadRequestError("Username can only contain letters and numbers", "US005")

	ErrTagEmpty       = NewBadRequestError("Tag cannot be empty", "US006")
	ErrTagTooLong     = NewBadRequestError("Tag must be less than 6 characters", "US007")
	ErrTagInvalidChar = NewBadRequestError("Tag can only contain letters and numbers", "US008")

	ErrBioTooLong = NewBadRequestError("Bio is too long", "US009")

	ErrPasswordTooShort = NewBadRequestError("Password must be at least 8 characters", "US0010")
	ErrPasswordTooWeak  = NewBadRequestError("Password must contain at least 3 of: uppercase, lowercase, number, special character", "US0011")

	ErrInvalidEmail = NewBadRequestError("Invalid Email Format", "US0012")
)
