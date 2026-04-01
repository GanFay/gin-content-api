package router_test

import (
	"blog/handlers"
	"blog/router"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTest(t *testing.T) (*handlers.Handler, *pgxpool.Pool) {
	t.Helper()

	dbURL := "postgres://app1:app@localhost:5432/db?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	h := &handlers.Handler{DB: pool}

	return h, pool
}

func TestSetupRouter_HasPingRoute(t *testing.T) {
	h, p := setupTest(t)
	defer p.Close()
	r := router.SetupRouter(h)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestSetupRouter_PublicAuthRoutesAccessible(t *testing.T) {
	h, p := setupTest(t)
	defer p.Close()
	r := router.SetupRouter(h)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "register route exists",
			method: http.MethodPost,
			path:   "/auth/register",
			body:   `{}`,
		},
		{
			name:   "login route exists",
			method: http.MethodPost,
			path:   "/auth/login",
			body:   `{}`,
		},
		{
			name:   "refresh route exists",
			method: http.MethodGet,
			path:   "/auth/refresh",
			body:   "",
		},
		{
			name:   "logout route exists",
			method: http.MethodPost,
			path:   "/auth/logout",
			body:   "",
		},
		{
			name:   "get posts exists",
			method: http.MethodGet,
			path:   "/posts",
			body:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Fatalf("expected route %s %s to exist, got 404", tt.method, tt.path)
			}
		})
	}
}

func TestSetupRouter_ProtectedRoutesRequireAuth(t *testing.T) {
	h, p := setupTest(t)
	defer p.Close()
	r := router.SetupRouter(h)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"create post", http.MethodPost, "/posts"},
		{"update post", http.MethodPut, "/posts/1"},
		{"delete post", http.MethodDelete, "/posts/1"},
		{"get me", http.MethodGet, "/users/me"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d for %s %s, got %d. body: %s",
					http.StatusUnauthorized, tt.method, tt.path, w.Code, w.Body.String())
			}
		})
	}
}

func TestSetupRouter_NotFound(t *testing.T) {
	h, p := setupTest(t)
	defer p.Close()
	r := router.SetupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/no-such-route", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d. body: %s",
			http.StatusNotFound, w.Code, w.Body.String())
	}
}
