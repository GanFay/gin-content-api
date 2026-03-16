package handlers

import (
	"blog/auth"
	"blog/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTest(t *testing.T) (*Handler, *gin.Engine, *pgxpool.Pool) {
	t.Helper()

	DBUrl := "postgres://app1:app@localhost:5432/db?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), DBUrl)
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	h := &Handler{DB: pool}
	r := gin.Default()

	return h, r, pool
}

func createTestUser(t *testing.T, pool *pgxpool.Pool, username, email, passwordHash string) int {
	t.Helper()

	var id int
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (username, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		username, email, passwordHash,
	).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func deleteTestUser(t *testing.T, pool *pgxpool.Pool, id int) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegister_Valid(t *testing.T) {
	h, r, pool := setupTest(t)
	defer pool.Close()

	r.POST("/auth/register", h.Register)
	body := `{
	"username": "test_reg",
	"email": "testreg@test.com",
	"password": "testreg123"
}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatal(w.Code)
	}

	var userID int
	err := h.DB.QueryRow(context.Background(), "SELECT id FROM users WHERE username=$1", "test_reg").Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}
	defer deleteTestUser(t, pool, userID)
}

func TestRegister_Invalid(t *testing.T) {
	h, r, pool := setupTest(t)
	defer pool.Close()

	r.POST("/auth/register", h.Register)
	body := `{
	"username": "",
	"email": "",
	"password": ""
}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}

func TestLogin_Valid(t *testing.T) {
	h, r, pool := setupTest(t)
	defer pool.Close()

	username := "test_log"
	password := "test123log"
	email := "testlog@test.com"
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	userId := createTestUser(t, pool, username, email, passwordHash)
	defer deleteTestUser(t, pool, userId)
	r.POST("/auth/login", h.Login)
	body := `{
	"username": "test_log",
	"password": "test123log"
}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var bodyResult models.TestLoginResponse

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	err = json.Unmarshal(w.Body.Bytes(), &bodyResult)
	if err != nil {
		t.Fatal(err)
	}

	userIDAccessJWT, err := auth.ParseJWTAccess(bodyResult.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if userIDAccessJWT != userId {
		t.Fatal("wrong access token")
	}

	cookies := w.Header()["Set-Cookie"]
	if len(cookies) == 0 {
		t.Fatal("no cookies")
	}
	cookie := cookies[0]
	cookieSplit := strings.Split(cookie, ";")
	cookieFinish := strings.Split(cookieSplit[0], "=")
	userIDRefreshJWT, err := auth.ParseJWTRefresh(cookieFinish[1])
	if err != nil {
		t.Fatal(err)
	}
	if userIDRefreshJWT != userId {
		t.Fatal("wrong refresh token")
	}
}

func TestLogin_Invalid(t *testing.T) {
	h, r, pool := setupTest(t)
	defer pool.Close()

	r.POST("/auth/login", h.Login)
	body := `{
	"username": "",
	"password": ""
}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}

}
