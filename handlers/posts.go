package handlers

import (
	"blog/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// CreateBlog GoDoc
// @Summary Create a new blog post
// @Description Creates a new blog post for the authenticated user
// @Tags posts
// @Accept JSON
// @Produce JSON
// @Security BearerAuth
// @Param input body models.Blog true "Blog data"
// @Success 201 {object} map[string]string "Post created successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to create blog"
// @Router /posts [post]
func (h *Handler) CreateBlog(c *gin.Context) {
	var newBlog models.Blog
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
	userIdStr := strconv.Itoa(userID.(int))
	if len(newBlog.Title) < 3 || len(newBlog.Title) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect title. It must be between 3 and 50 characters long."})
		return
	}
	if len(newBlog.Content) < 3 || len(newBlog.Content) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect content. It must be between 3 and 500 characters long."})
		return
	}

	_, err = h.DB.Exec(c.Request.Context(), `
		INSERT INTO posts (author_id, title, content, category, tags)
		VALUES ($1, $2, $3, $4, $5)
	`, userIdStr, newBlog.Title, newBlog.Content, newBlog.Category, newBlog.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write in DB"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "post created successfully"})
}

// GetAllPosts GoDoc
// @Summary Get all posts
// @Description Returns a list of posts with optional search by term and pagination
// @Tags posts
// @Accept JSON
// @Produce JSON
// @Param term query string false "Search term for title, content, or category"
// @Param limit query int false "Number of posts to return" default(10)
// @Param offset query int false "Number of posts to skip" default(0)
// @Success 200 {object} map[string]interface{} "List of posts"
// @Failure 400 {object} map[string]string "Query error"
// @Failure 500 {object} map[string]string "Server error"
// @Router /posts [get]
func (h *Handler) GetAllPosts(c *gin.Context) {
	term := c.Query("term")

	var (
		rows pgx.Rows
		err  error
	)

	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	if term != "" {
		query := `
			SELECT id, author_id, title, content, category, tags, created_at, updated_at
			FROM posts
			WHERE
				title ILIKE '%' || $1 || '%'
				OR content ILIKE '%' || $1 || '%'
				OR category ILIKE '%' || $1 || '%'
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4;
		`
		rows, err = h.DB.Query(c.Request.Context(), query, term, limit, offset)
	} else {
		rows, err = h.DB.Query(c.Request.Context(), `SELECT posts.* FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2;`, limit, offset)
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.Category, &p.Tags, &p.CreatedAt, &p.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// GetPoID GoDoc
// @Summary Get post by ID
// @Description Returns a single post by its ID
// @Tags posts
// @Accept JSON
// @Produce JSON
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{} "Post found"
// @Failure 400 {object} map[string]string "Invalid post ID"
// @Router /posts/{id} [get]
func (h *Handler) GetPoID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + idStr})
	}
	var post models.Post
	err = h.DB.QueryRow(c.Request.Context(), `SELECT * FROM posts WHERE id=$1`, id).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": "invalid id: " + idStr})
	}
	c.JSON(http.StatusOK, gin.H{"post": post})
}

// DeleteBlog GoDoc
// @Summary Delete blog post
// @Description Deletes a post if the authenticated user is its author
// @Tags posts
// @Accept JSON
// @Produce JSON
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 204 {object} map[string]string "Post deleted successfully"
// @Failure 400 {object} map[string]string "Invalid request or database error"
// @Failure 401 {object} map[string]string "Unauthorized or no permission"
// @Failure 404 {object} map[string]string "Post not found"
// @Router /posts/{id} [delete]
func (h *Handler) DeleteBlog(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found"})
		return
	}
	var post models.Post
	id := c.Param("id")

	err := h.DB.QueryRow(c.Request.Context(), `SELECT * FROM posts WHERE id=$1`, id).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	authorID, err := strconv.Atoi(post.AuthorID)
	if err != nil {
		return
	}
	if authorID != userID.(int) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "not permission"})
		return
	}

	cmdTag, err := h.DB.Exec(c.Request.Context(), `DELETE FROM posts WHERE id=$1`, id)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
		return
	}

	c.JSON(204, gin.H{"message": "deleted post successfully"})
}

// UpdateBlog GoDoc
// @Summary Update blog post
// @Description Updates a post if the authenticated user is its author
// @Tags posts
// @Accept JSON
// @Produce JSON
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param input body models.Blog true "Updated blog data"
// @Success 200 {object} map[string]string "Post updated successfully"
// @Failure 400 {object} map[string]string "Invalid input or update failed"
// @Failure 401 {object} map[string]string "Unauthorized or no permission"
// @Router /posts/{id} [put]
func (h *Handler) UpdateBlog(c *gin.Context) {
	idStr := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + idStr})
	}
	var post models.Post
	err = h.DB.QueryRow(c.Request.Context(), `SELECT * FROM posts WHERE id=$1`, id).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	AtoI, err := strconv.Atoi(post.AuthorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author id: " + post.AuthorID})
		return
	}

	if AtoI != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "not permission"})
		return
	}

	var newBlog models.Blog
	err = c.ShouldBindJSON(&newBlog)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	timeNow := time.Now()

	cmdTag, err := h.DB.Exec(c.Request.Context(), `
	UPDATE posts SET title=$1, content=$2, category=$3, tags=$4, updated_at=$6 WHERE id=$5`, newBlog.Title, newBlog.Content, newBlog.Category, newBlog.Tags, id, timeNow)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to update post: " + err.Error()})
		return
	}
	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "successfully updated blog!"})
}
