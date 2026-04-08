package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping
// @Summary      Health check
// @Description  Check if the server is running
// @Tags         System
// @Produce      json
// @Success      200  {object}  map[string]string "message: pong"
// @Router       /ping [get]
func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
