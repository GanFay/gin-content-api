package repository

import (
	"blog/models"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, authorId int, title string, content string, category string, tags []string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO posts (author_id, title, content, category, tags) VALUES ($1, $2, $3, $4, $5)`, authorId, title, content, category, tags)
	return err
}

func (r *PostRepository) GetAll(ctx context.Context, term string, limit string, offset string) ([]models.Post, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if term != "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, author_id, title, content, category, tags, created_at, updated_at
			FROM posts
			WHERE
				title ILIKE '%' || $1 || '%'
				OR content ILIKE '%' || $1 || '%'
				OR category ILIKE '%' || $1 || '%'
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3;
		`, term, limit, offset)
	} else {
		rows, err = r.db.Query(ctx, `SELECT posts.* FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2;`, limit, offset)
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err = rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.Category, &p.Tags, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return posts, err
		}
		posts = append(posts, p)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}
	return posts, err
}

func (r *PostRepository) GetByID(ctx context.Context, id int) (models.Post, error) {
	var post models.Post
	err := r.db.QueryRow(ctx, `SELECT * FROM posts WHERE id=$1`, id).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	return post, err
}

func (r *PostRepository) Delete(ctx context.Context, id int, userID int) error {
	var post models.Post
	err := r.db.QueryRow(ctx, `SELECT * FROM posts WHERE id=$1`, id).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}

	if post.AuthorID != userID {
		err = errors.New("not permission")
		return err
	}

	cmdTag, err := r.db.Exec(ctx, `DELETE FROM posts WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		err = errors.New("post not found")
		return err
	}
	return err
}

func (r *PostRepository) Update(ctx context.Context, id int, userID int, newBlog models.PostReq) error {
	var post models.Post
	err := r.db.QueryRow(ctx, `SELECT * FROM posts WHERE id=$1`, id).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Category, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}

	if post.AuthorID != userID {
		err = errors.New("not permission")
		return err
	}

	timeNow := time.Now()

	cmdTag, err := r.db.Exec(ctx, `
	UPDATE posts SET title=$1, content=$2, category=$3, tags=$4, updated_at=$6 WHERE id=$5`, newBlog.Title, newBlog.Content, newBlog.Category, newBlog.Tags, id, timeNow)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		err = errors.New("post not found")
		return err
	}
	return err
}
