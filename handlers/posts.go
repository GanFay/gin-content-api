package handlers

import (
	"blog/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreatePost GoDoc
// @Summary Create a new blog post
// @Description Creates a new blog post for the authenticated user
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body models.PostReq true "Blog data"
// @Success 201 {object} map[string]string "Post created successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to create blog"
// @Router /posts [post]
func (h *Handler) CreatePost(c *gin.Context) {
	ctx := c.Request.Context()
	var newBlog models.PostReq
	err := c.ShouldBindJSON(&newBlog)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON can't unmarshal body"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found, unauthorized"})
		return
	}
	if len(newBlog.Title) < 3 || len(newBlog.Title) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect title. It must be between 3 and 50 characters long."})
		return
	}
	if len(newBlog.Content) < 3 || len(newBlog.Content) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect content. It must be between 3 and 500 characters long."})
		return
	}

	err = h.Posts.Create(ctx, userID.(int), newBlog.Title, newBlog.Content, newBlog.Category, newBlog.Tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "post created successfully"})
}

// GetPosts GoDoc
// @Summary Get all posts
// @Description Returns a list of posts with optional search by term and pagination
// @Tags posts
// @Accept json
// @Produce json
// @Param term query string false "Search term for title, content, or category"
// @Param limit query int false "Number of posts to return" default(10)
// @Param offset query int false "Number of posts to skip" default(0)
// @Success 200 {object} map[string]interface{} "List of posts"
// @Failure 400 {object} map[string]string "Query error"
// @Failure 500 {object} map[string]string "Server error"
// @Router /posts [get]
func (h *Handler) GetPosts(c *gin.Context) {
	ctx := c.Request.Context()
	term := c.Query("term")

	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be int"})
		return
	}
	if intLimit > 99 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit is too big"})
		return
	}
	posts, err := h.Posts.GetAll(ctx, term, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// GetByID GoDoc
// @Summary Get post by ID
// @Description Returns a single post by its ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{} "Post found"
// @Failure 400 {object} map[string]string "Invalid post ID"
// @Router /posts/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + idStr})
		return
	}
	post, err := h.Posts.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"post": post})
}

// DeletePost GoDoc
// @Summary Delete blog post
// @Description Deletes a post if the authenticated user is its author
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 204 {object} map[string]string "Post deleted successfully"
// @Failure 400 {object} map[string]string "Invalid request or database error"
// @Failure 401 {object} map[string]string "Unauthorized or no permission"
// @Failure 404 {object} map[string]string "Post not found"
// @Router /posts/{id} [delete]
func (h *Handler) DeletePost(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found"})
		return
	}
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + id})
		return
	}
	err = h.Posts.Delete(ctx, idInt, userID.(int))
	if err != nil {
		switch err.Error() {
		case "not permission":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		case "no rows in result set":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		case "post not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.Status(204)
}

// UpdatePost GoDoc
// @Summary Update blog post
// @Description Updates a post if the authenticated user is its author
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param input body models.PostReq true "Updated blog data"
// @Success 200 {object} map[string]string "Post updated successfully"
// @Failure 400 {object} map[string]string "Invalid input or update failed"
// @Failure 401 {object} map[string]string "Unauthorized or no permission"
// @Router /posts/{id} [put]
func (h *Handler) UpdatePost(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id " + idStr})
		return
	}
	var newBlog models.PostReq
	err = c.ShouldBindJSON(&newBlog)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.Posts.Update(ctx, id, userID.(int), newBlog)
	if err != nil {
		switch err.Error() {
		case "not permission":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		case "post not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		case "no rows in result set":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "successfully updated post"})
}
