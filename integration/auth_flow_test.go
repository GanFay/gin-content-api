package integration

import (
	"blog/handlers"
	"blog/models"
	"blog/router"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTest(t *testing.T) (*handlers.Handler, *pgxpool.Pool) {
	t.Helper()

	dbURL := "postgres://app1:app@localhost:5432/db?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatal(err)
	}
	h := &handlers.Handler{DB: pool}

	return h, pool
}
func deleteTestUser(t *testing.T, pool *pgxpool.Pool, username string) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `DELETE FROM users WHERE users.username = $1`, username)
	if err != nil {
		t.Fatal(err)
	}
}
func decodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var resp T
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err.Error())
	}
	return resp
}
func createTestUser(t *testing.T, pool *pgxpool.Pool, username, email, passwordHash string) (int, error) {
	t.Helper()

	var id int
	err := pool.QueryRow(
		context.Background(),
		`INSERT INTO users (username, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		username, email, passwordHash,
	).Scan(&id)

	return id, err
}

func TestAuthFlow_RegisterLoginMe(t *testing.T) {
	// 0.Prepare
	h, p := setupTest(t)
	defer p.Close()

	deleteTestUser(t, p, "maks")
	defer deleteTestUser(t, p, "maks")

	r := router.SetupRouter(h)

	// 1. Register
	body := `{
		"username": "maks",
		"email": "maks@maks.com",
		"password": "maksPassWord123"
	}`

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatal("want:", http.StatusCreated, "got:", w.Code, "body:", w.Body.String())
	}

	respRegister := decodeJSON[map[string]string](t, w)
	if respRegister["message"] != "register successfully" {
		t.Fatal("want: register successfully, got:", respRegister["message"])
	}

	// 2. Login
	body = `{
		"username": "maks",
		"password": "maksPassWord123"
	}`

	req = httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatal("want:", http.StatusOK, "got:", w.Code, "body:", w.Body.String())
	}

	respLogin := decodeJSON[models.LoginResponse](t, w)
	if respLogin.AccessToken == "" {
		t.Fatal("access token is empty")
	}

	// 3. Me
	req = httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+respLogin.AccessToken)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatal("want:", http.StatusOK, "got:", w.Code, "body:", w.Body.String())
	}

	respMe := decodeJSON[models.MeResponse](t, w)
	if respMe.Username != "maks" {
		t.Fatal("want: maks, got:", respMe.Username)
	}
	if respMe.Email != "maks@maks.com" {
		t.Fatal("want: maks@maks.com, got:", respMe.Email)
	}
}

func TestAuthFlow_LoginRefreshMe(t *testing.T) {
	// 0.Prepare
	h, p := setupTest(t)
	defer p.Close()

	deleteTestUser(t, p, "maks")
	fullCreateUser(t, p, "maks")
	defer deleteTestUser(t, p, "maks")

	r := router.SetupRouter(h)

	// 1.Login
	body := `{
			"username": "maks",
			"password": "maksPassWord123"
}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	rLog := decodeJSON[models.LoginResponse](t, w)
	if rLog.AccessToken == "" {
		t.Fatal("access token is empty")
	}

	var refreshCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}
	if refreshCookie == nil {
		t.Fatal("refresh_token cookie not found")
	}

	// 2. Refresh
	req = httptest.NewRequest(http.MethodGet, "/auth/refresh", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshCookie.Value)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("want:", http.StatusOK, "got:", w.Code, "body:", w.Body.String())
	}
	rRef := decodeJSON[map[string]string](t, w)
	if rRef["access_token"] == "" {
		t.Fatal("access token is empty")
	}

	// 3. Me

	req = httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+rRef["access_token"])
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("want:", http.StatusOK, "got:", w.Code, "body:", w.Body.String())
	}
	rMe := decodeJSON[models.MeResponse](t, w)
	if rMe.Username != "maks" {
		t.Fatal("want: maks, got:", rMe.Username)
	}

}

func TestAuthFlow_LogoutThenRefreshFails(t *testing.T) {
	// 0.Prepare
	h, p := setupTest(t)
	defer p.Close()
	deleteTestUser(t, p, "maks")
	defer deleteTestUser(t, p, "maks")
	_, refreshCookie := fullCreateUser(t, p, "maks")
	r := router.SetupRouter(h)

	// 1.Logout
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Cookie", "refresh_token="+refreshCookie)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("want:", http.StatusOK, "got:", w.Code, "body:", w.Body.String())
	}
	resp := decodeJSON[map[string]string](t, w)
	if resp["message"] != "logged out" {
		t.Fatal("want: logged out, got: ", resp)
	}
	cookies := w.Result().Cookies()
	// 2.Refresh
	req = httptest.NewRequest(http.MethodGet, "/auth/refresh", nil)
	var refreshCookieResp *http.Cookie

	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookieResp = c
			break
		}
	}
	if refreshCookieResp == nil {
		t.Fatal("refresh_token cookie not found")
	}
	req.Header.Set("Cookie", "refresh_token="+refreshCookieResp.Value)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatal("want:", http.StatusOK, "got:", w.Code, "body:", w.Body.String())
	}
	resp = decodeJSON[map[string]string](t, w)
	if !strings.Contains(resp["error"], "token is malformed") {
		t.Fatal("expected malformed token error, got:", resp["error"])
	}
}
