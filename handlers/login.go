package handlers

import (
	"blog/auth"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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

func (h *Handler) Login(c *gin.Context) {
	var req Login
	var user Users
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Username) < 4 || len(req.Username) > 32 {
		c.JSON(400, gin.H{"error": "username is too short or too long"})
		return
	} else if len(req.Password) < 5 || len(req.Password) > 128 {
		c.JSON(400, gin.H{"error": "password is too short or too long"})
		return
	}
	err = h.DB.QueryRow(c.Request.Context(), `SELECT * FROM users WHERE username=$1`, req.Username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "user does not exist"})
		return
	}
	if !auth.ComparePasswords(user.PasswordHash, req.Password) {
		c.JSON(401, gin.H{"error": "wrong password"})
		return
	}

	var AccessToken string
	AccessToken, err = auth.GenerateAccessJWT(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var RefreshToken string
	RefreshToken, err = auth.GenerateRefreshJWT(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refresh_token",
		RefreshToken,
		60*60*24*7,
		"/",
		"",
		false,
		true,
	)

	c.JSON(200, gin.H{
		"access_token": AccessToken,
	})
}
