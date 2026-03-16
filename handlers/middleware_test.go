package handlers

import (
	"blog/auth"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{}
	r := gin.Default()

	r.GET("/ping", h.AuthMiddleware(), h.Ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{}
	r := gin.Default()

	r.GET("/ping", h.AuthMiddleware(), h.Ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var body map[string]string
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}

	err := json.Unmarshal([]byte(w.Body.String()), &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["error"] != "invalid token" {
		t.Fatal("got: ", body["error"], "want: invalid token")
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	userID := 1
	gin.SetMode(gin.TestMode)

	h := &Handler{}
	r := gin.Default()

	r.GET("/ping", h.AuthMiddleware(), h.Ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	BearerToken, err := auth.GenerateAccessJWT(userID)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+BearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
}
