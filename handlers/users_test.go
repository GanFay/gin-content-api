package handlers

import (
	"blog/auth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMe_Success(t *testing.T) {
	h, r, pool, id := setupTest(t)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err)
	}
	r.GET("/users/me", h.AuthMiddleware(), h.Me)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)

	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
}

func TestMe_Unauthorized(t *testing.T) {
	h, r, pool, id := setupTest(t)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err)
	}
	r.GET("/users/me", h.AuthMiddleware(), h.Me)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)

	req.Header.Set("Authorization", "Bearer "+jwt+"awd")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
}

func TestMe_UserNotFound(t *testing.T) {
	h, r, pool, id := setupTest(t)
	defer pool.Close()
	deleteTestUser(t, pool, id)
	jwt, err := auth.GenerateAccessJWT(-1)
	if err != nil {
		t.Fatal(err)
	}
	r.GET("/users/me", h.AuthMiddleware(), h.Me)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
}
