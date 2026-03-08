package handlers

import (
	"blog/auth"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Refresh(c *gin.Context) {

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(401, gin.H{"message": "no refreshToken"})
		return
	}

	userID, err := auth.ParseJWT(refreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid refresh"})
		return
	}

	access, _ := auth.GenerateAccessJWT(userID)

	c.JSON(200, gin.H{
		"access_token": access,
	})

}
