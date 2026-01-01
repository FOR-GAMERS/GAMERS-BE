package dto

import "time"

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
	Tag      string `json:"tag" binding:"required"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
}

type UpdateUserRequest struct {
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	Id         int64     `json:"user_id"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}
