package handlers

import (
	"blog/auth"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var MuserID = 5

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	h, r, pool, id := setupTest(t)
	deleteTestUser(t, pool, id)
	defer pool.Close()
	r.GET("/ping", h.AuthMiddleware(), h.Ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
}
func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	h, r, pool, id := setupTest(t)
	defer pool.Close()
	deleteTestUser(t, pool, id)
	r.GET("/ping", h.AuthMiddleware(), h.Ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	BearerToken, err := auth.GenerateAccessJWT(MuserID)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", BearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
	var body map[string]string
	err = json.Unmarshal([]byte(w.Body.String()), &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["error"] != "invalid type header" {
		t.Fatal("got: ", body["error"], "want: invalid type header")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	h, r, pool, id := setupTest(t)
	deleteTestUser(t, pool, id)
	defer pool.Close()
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
	if body["error"] != "token is malformed: token contains an invalid number of segments" {
		t.Fatal("got: ", body["error"], "want: invalid token")
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	h, r, pool, id := setupTest(t)
	defer pool.Close()
	deleteTestUser(t, pool, id)
	r.GET("/ping", h.AuthMiddleware(), h.Ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	JWTFunc := func(userID int) (string, error) {
		claims := jwt.MapClaims{
			"user_id": userID,
			"exp":     time.Now().Unix(),
			"iat":     time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(auth.GetAccessSecret())
		if err != nil {
			return "", err
		}
		return signedToken, nil
	}

	token, err := JWTFunc(MuserID)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatal()
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	h, r, pool, id := setupTest(t)
	deleteTestUser(t, pool, id)
	defer pool.Close()
	r.GET("/ping", h.AuthMiddleware(), h.Ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	BearerToken, err := auth.GenerateAccessJWT(MuserID)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+BearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}

func TestAuthMiddleware_SetsUserIDInContext(t *testing.T) {
	h, r, pool, id := setupTest(t)
	defer pool.Close()
	deleteTestUser(t, pool, id)
	r.GET("/check", h.AuthMiddleware(), func(c *gin.Context) {
		value, exists := c.Get("user_id")
		if !exists {
			c.Status(http.StatusInternalServerError)
			return
		}

		id, ok := value.(int)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		if id != MuserID {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/check", nil)
	BearerToken, err := auth.GenerateAccessJWT(MuserID)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+BearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}
