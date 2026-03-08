package handlers

import (
	"github.com/gin-gonic/gin"
)

type info struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Me godoc
// @Summary Get current user
// @Description Returns information about the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} info "User information"
// @Failure 400 {object} map[string]string "Database error"
// @Router /users/me [get]
func (h *Handler) Me(c *gin.Context) {
	var req info
	userID := c.GetInt("user_id")

	err := h.DB.QueryRow(c.Request.Context(), `SELECT id, username, email FROM users WHERE id=$1`, userID).Scan(&req.Id, &req.Username, &req.Email)
	if err != nil {
		c.JSON(400, gin.H{
			"message": err,
		})
		return
	}
	c.JSON(200, gin.H{"id": req.Id, "username": req.Username, "email": req.Email})

}
