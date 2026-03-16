package handlers

import (
	"blog/auth"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMe(t *testing.T) {
	userID := 1
	DBUrl := "postgres://app1:app@localhost:5432/db?sslmode=disable"
	jwt, err := auth.GenerateAccessJWT(userID)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := pgxpool.New(context.Background(), DBUrl)
	if err != nil {
		log.Fatal("Failed to connected DB: ", err)
	}
	defer pool.Close()
	gin.SetMode(gin.TestMode)

	h := &Handler{DB: pool}
	r := gin.Default()

	r.GET("/users/me", h.AuthMiddleware(), h.Me)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)

	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}

}
