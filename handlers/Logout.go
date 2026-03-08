package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Logout(c *gin.Context) {

	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.JSON(200, gin.H{
		"message": "logged out",
	})
}
