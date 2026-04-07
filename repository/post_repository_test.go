package repository

import (
	"blog/db"
	"blog/models"
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func setupTestP(t *testing.T) (*PostRepository, context.Context) {
	t.Helper()
	ctx := context.Background()
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	dbURL := os.Getenv("DB_URL")
	p := db.MustConnect(dbURL)
	r := NewPostRepository(p)
	return r, ctx
}

func TestPostRepository_Create(t *testing.T) {
	r, ctx := setupTestP(t)
	defer r.db.Close()

	var (
		authorID = -1
		title    = "test"
		content  = "content"
		category = "category"
		tags     = []string{"te", "st"}
		id       int
	)

	err := r.Create(ctx, authorID, title, content, category, tags)
	if err != nil {
		t.Fatal(err)
	}
	err = r.db.QueryRow(ctx, `SELECT id FROM posts WHERE title=$1`, title).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	err = r.Delete(ctx, id, authorID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostRepository_GetAll(t *testing.T) {
	r, ctx := setupTestP(t)
	defer r.db.Close()

	var (
		authorID = -1
		title    = "test"
		content  = "content"
		category = "category"
		tags     = []string{"te", "st"}
		id       int
	)
	err := r.Create(ctx, authorID, title, content, category, tags)
	if err != nil {
		t.Fatal(err)
	}
	posts, err := r.GetAll(ctx, "", "10", "0")
	if err != nil {
		t.Fatal(err)
	}
	if posts[0].AuthorID != authorID && posts[0].Title != title {
		t.Fatal("invalid post")
	}
	err = r.db.QueryRow(ctx, `SELECT id FROM posts WHERE title=$1`, title).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	err = r.Delete(ctx, id, authorID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostRepository_GetByID(t *testing.T) {
	r, ctx := setupTestP(t)
	defer r.db.Close()

	var (
		authorID = -1
		title    = "test"
		content  = "content"
		category = "category"
		tags     = []string{"te", "st"}
		id       int
	)
	err := r.Create(ctx, authorID, title, content, category, tags)
	if err != nil {
		t.Fatal(err)
	}
	err = r.db.QueryRow(ctx, `SELECT id FROM posts WHERE title=$1`, title).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	post, err := r.GetByID(ctx, id)
	if post.ID != int64(id) && post.Title != title {
		t.Fatal("invalid post")
	}
	err = r.Delete(ctx, id, authorID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostRepository_Delete(t *testing.T) {
	r, ctx := setupTestP(t)
	defer r.db.Close()

	var (
		authorID = -1
		title    = "test"
		content  = "content"
		category = "category"
		tags     = []string{"te", "st"}
		id       int
	)
	err := r.Create(ctx, authorID, title, content, category, tags)
	if err != nil {
		t.Fatal(err)
	}
	err = r.db.QueryRow(ctx, `SELECT id FROM posts WHERE title=$1`, title).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	err = r.Delete(ctx, id, authorID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPostRepository_Update(t *testing.T) {
	r, ctx := setupTestP(t)
	defer r.db.Close()

	var (
		authorID = -1
		title    = "test"
		content  = "content"
		category = "category"
		tags     = []string{"te", "st"}
		id       int
		newBlog  = models.PostReq{Title: title + "2", Content: content + "2", Category: category + "2", Tags: tags}
	)
	err := r.Create(ctx, authorID, title, content, category, tags)
	if err != nil {
		t.Fatal(err)
	}
	err = r.db.QueryRow(ctx, `SELECT id FROM posts WHERE title=$1`, title).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	err = r.Update(ctx, id, authorID, newBlog)
	if err != nil {
		t.Fatal(err)
	}
	post, err := r.GetByID(ctx, id)
	if post.ID != int64(id) && post.Title != title {
		t.Fatal("invalid post")
	}
	if post.Title != newBlog.Title && post.Content != newBlog.Content && post.Category != newBlog.Category {
		t.Fatal("invalid post")
	}
	err = r.Delete(ctx, id, authorID)
	if err != nil {
		t.Fatal(err)
	}
}
