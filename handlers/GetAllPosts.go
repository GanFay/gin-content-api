package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Post struct {
	ID        int64     `json:"id"`
	AuthorID  string    `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *Handler) GetAllPosts(c *gin.Context) {
	term := c.Query("term")

	var (
		rows pgx.Rows
		err  error
	)

	if term != "" {
		query := `
			SELECT id, author_id, title, content, category, tags, created_at, updated_at
			FROM posts
			WHERE
				title ILIKE '%' || $1 || '%'
				OR content ILIKE '%' || $1 || '%'
				OR category ILIKE '%' || $1 || '%'
			ORDER BY id DESC;
		`
		rows, err = h.DB.Query(c.Request.Context(), query, term)
	} else {
		rows, err = h.DB.Query(c.Request.Context(), `SELECT posts.* FROM posts ORDER BY id DESC;`)
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
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
