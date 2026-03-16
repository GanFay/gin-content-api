package models

import "time"

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Users struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TestLoginResponse struct {
	UserId      int    `json:"userId" binding:"required"`
	AccessToken string `json:"access_token" binding:"required"`
}
