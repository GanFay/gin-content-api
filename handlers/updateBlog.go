package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateBlog(c *gin.Context) {
	idstr := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found"})
		return
	}
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + idstr})
	}
	var post Post
	err = h.DB.QueryRow(c.Request.Context(), `SELECT * FROM posts WHERE id=$1`, id).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	atoi, err := strconv.Atoi(post.AuthorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author id: " + post.AuthorID})
		return
	}

	if atoi != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "not permission"})
		return
	}

	var newBlog Blog
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
