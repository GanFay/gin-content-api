package handlers

import (
	"blog/auth"
	"blog/models"
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
)

// Login godoc
// @Summary Login user
// @Description Authenticates user by username and password, returns access token and sets refresh token in HttpOnly cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.Login true "Login credentials"
// @Success 200 {object} map[string]string "Access token returned successfully"
// @Failure 400 {object} map[string]string "Invalid request body or validation error"
// @Failure 401 {object} map[string]string "Wrong password"
// @Failure 500 {object} map[string]string "User does not exist or internal server error"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req models.Login

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Username) < 4 || len(req.Username) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is too short or too long"})
		return
	} else if len(req.Password) < 5 || len(req.Password) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is too short or too long"})
		return
	}
	user, err := h.Users.GetByUserName(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !auth.ComparePasswords(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "password wrong"})
		return
	}

	var AccessToken string
	AccessToken, err = auth.GenerateAccessJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var RefreshToken string
	RefreshToken, err = auth.GenerateRefreshJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	c.JSON(http.StatusOK, gin.H{
		"userId":       user.ID,
		"access_token": AccessToken,
	})
}

// Register
// @Summary      Register a new user
// @Description  Create a new user account with username, email, and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input  body      models.RegisterRequest  true  "User registration data"
// @Success      201    {object}  map[string]string "message: user created successfully"
// @Failure      400    {object}  map[string]string "error: invalid input"
// @Failure      409    {object}  map[string]string "error: user already exists"
// @Failure      500    {object}  map[string]string "error: internal server error"
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	var req models.RegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Username) < 4 || len(req.Username) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be between 4 and 32 characters"})
		return
	}

	if len(req.Password) < 5 || len(req.Password) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be between 5 and 128 characters"})
		return
	}

	if len(req.Email) < 6 || len(req.Email) > 256 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email must be between 6 and 256 characters"})
		return
	}

	_, err = mail.ParseAddress(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.Users.Add(ctx, req.Username, req.Email, hashPassword)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "register successfully"})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Generates a new access token using refresh token from HttpOnly cookie
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "New access token generated"
// @Failure 401 {object} map[string]string "Missing or invalid refresh token"
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	userID, err := auth.ParseJWTRefresh(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	access, _ := auth.GenerateAccessJWT(userID)

	c.JSON(http.StatusOK, gin.H{
		"access_token": access,
	})

}

// Logout godoc
// @Summary Logout user
// @Description Clears refresh token cookie and logs user out
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "User logged out successfully"
// @Router /auth/logout [post]v
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
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out",
	})
}
