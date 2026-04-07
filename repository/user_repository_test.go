package repository

import (
	"blog/auth"
	"blog/db"
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func setupTestU(t *testing.T) (*UserRepository, context.Context) {
	t.Helper()
	ctx := context.Background()
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	dbURL := os.Getenv("DB_URL")
	p := db.MustConnect(dbURL)
	r := NewUserRepository(p)
	return r, ctx
}

func TestUserRepository_Add_Get(t *testing.T) {
	r, ctx := setupTestU(t)
	defer r.db.Close()
	var (
		username = "test2"
		email    = "test1234@test.com"
	)
	hPass, err := auth.HashPassword("password")
	if err != nil {
		t.Fatal(err)
	}
	err = r.Add(ctx, username, email, hPass)
	if err != nil {
		t.Fatal(err)
	}
	usr1, err := r.GetByUserName(ctx, username)
	if err != nil {
		t.Fatal(err)
	}
	if usr1.Username != username {
		t.Errorf("Username mismatch: got %q, want %q", usr1.Username, username)
	}
	usr2, err := r.GetByID(ctx, usr1.ID)
	if usr2.Username != username {
		t.Errorf("Username mismatch: got %q, want %q", usr2.Username, username)
	}
	_, err = r.db.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
	if err != nil {
		t.Fatal(err)
	}
}
