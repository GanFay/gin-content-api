package handlers

import (
	"blog/auth"
	"blog/models"
	"blog/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type errorResponse struct {
	Error string `json:"error"`
}

// TODO: optimize setupTest - don't create user for tests that don't need it

func setupTest(t *testing.T, addUsr bool) (*Handler, *gin.Engine, *pgxpool.Pool, int) {
	t.Helper()
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	dbURL := os.Getenv("DB_URL")
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	postRep := repository.NewPostRepository(pool)
	userRep := repository.NewUserRepository(pool)
	h := NewHandler(postRep, userRep)
	r := gin.Default()

	if addUsr {
		pass, err := auth.HashPassword("testout123")
		if err != nil {
			t.Fatal(err.Error())
		}

		username := "test_logout"
		email := "test_logout@gmail.com"

		id, err := createTestUser(t, h, username, email, pass)
		if err != nil {
			t.Fatal(err)
		}
		return h, r, pool, id
	}

	return h, r, pool, -1
}

func performJSONRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func decodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	err := json.Unmarshal(w.Body.Bytes(), &v)
	if err != nil {
		t.Fatalf("failed to decode response body: %v; body: %s", err, w.Body.String())
	}

	return v
}

func createTestUser(t *testing.T, h *Handler, username, email, passwordHash string) (int, error) {
	t.Helper()

	err := h.Users.Add(context.Background(), username, email, passwordHash)
	if err != nil {
		t.Fatal(err)
	}
	user, err := h.Users.GetByUserName(context.Background(), username)
	return user.ID, err
}

func deleteTestUser(t *testing.T, pool *pgxpool.Pool, id int) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegister_Validation(t *testing.T) {
	h, r, pool, _ := setupTest(t, false)
	defer pool.Close()

	r.POST("/auth/register", h.Register)

	tests := []struct {
		name          string
		body          string
		wantStatus    int
		wantContains  []string
		notEmptyError bool
	}{
		{
			name: "invalid json fields",
			body: `{
				"user2name": "test_reg",
				"emai2l": "testreg@test.com",
				"passw2ord": "testreg123"
			}`,
			wantStatus: http.StatusBadRequest,
			wantContains: []string{
				"RegisterRequest.Username",
				"RegisterRequest.Email",
				"RegisterRequest.Password",
			},
			notEmptyError: true,
		},
		{
			name: "empty username",
			body: `{
				"username": "",
				"email": "testreg@test.com",
				"password": "testreg123"
			}`,
			wantStatus: http.StatusBadRequest,
			wantContains: []string{
				"RegisterRequest.Username",
			},
			notEmptyError: true,
		},
		{
			name: "empty email",
			body: `{
				"username": "test_reg",
				"email": "",
				"password": "testreg123"
			}`,
			wantStatus: http.StatusBadRequest,
			wantContains: []string{
				"RegisterRequest.Email",
			},
			notEmptyError: true,
		},
		{
			name: "empty password",
			body: `{
				"username": "test_reg",
				"email": "testreg@test.com",
				"password": ""
			}`,
			wantStatus: http.StatusBadRequest,
			wantContains: []string{
				"RegisterRequest.Password",
			},
			notEmptyError: true,
		},
		{
			name: "invalid email",
			body: `{
				"username": "test_reg",
				"email": "testregtest12312com",
				"password": "testreg123"
			}`,
			wantStatus: http.StatusBadRequest,
			wantContains: []string{
				"mail: missing '@' or angle-addr",
			},
			notEmptyError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performJSONRequest(r, http.MethodPost, "/auth/register", tt.body)

			if w.Code != tt.wantStatus {
				t.Fatalf("wantLen status %d, got %d, body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			resp := decodeJSON[errorResponse](t, w)

			if tt.notEmptyError && resp.Error == "" {
				t.Fatal("wantLen non-empty error")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(resp.Error, want) {
					t.Fatalf("wantLen error to contain %q, got %q", want, resp.Error)
				}
			}
		})
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	h, r, pool, _ := setupTest(t, false)
	defer pool.Close()

	username := "test_reg_exists"
	email := "test_exists@test.com"
	password := "test123"

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	id, err := createTestUser(t, h, username, email, passwordHash)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deleteTestUser(t, pool, id)

	r.POST("/auth/register", h.Register)

	body := fmt.Sprintf(`{
		"username": "%s",
		"email": "%s",
		"password": "%s"
	}`, username, email, password)

	w := performJSONRequest(r, http.MethodPost, "/auth/register", body)

	if w.Code != http.StatusConflict {
		t.Fatalf("wantLen status %d, got %d, body: %s", http.StatusConflict, w.Code, w.Body.String())
	}

	resp := decodeJSON[errorResponse](t, w)

	if !strings.Contains(resp.Error, "SQLSTATE 23505") {
		t.Fatalf("wantLen duplicate key error, got: %q", resp.Error)
	}
}

func TestRegister_Success(t *testing.T) {
	h, r, pool, _ := setupTest(t, false)
	defer pool.Close()

	r.POST("/auth/register", h.Register)

	username := "test_reg_success"
	email := "testregsuccess@test.com"
	password := "testreg123"

	body := fmt.Sprintf(`{
		"username": "%s",
		"email": "%s",
		"password": "%s"
	}`, username, email, password)

	w := performJSONRequest(r, http.MethodPost, "/auth/register", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("wantLen status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	user, err := h.Users.GetByUserName(context.Background(), username)
	if err != nil {
		t.Fatal(err)
	}

	defer deleteTestUser(t, pool, user.ID)
}

func TestLogin_Validation(t *testing.T) {
	h, r, pool, id := setupTest(t, true)
	defer pool.Close()

	defer deleteTestUser(t, pool, id)

	r.POST("/auth/login", h.Login)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name: "invalid json fields",
			body: `{
				"usern2ame": "test_log",
				"passw2ord": "test123log"
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "username is too short or too long",
		},
		{
			name: "empty username",
			body: `{
				"username": "",
				"password": "test123"
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "username is too short or too long",
		},
		{
			name: "empty password",
			body: `{
				"username": "test_log",
				"password": ""
			}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "password is too short or too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := performJSONRequest(r, http.MethodPost, "/auth/login", tt.body)

			if w.Code != tt.wantStatus {
				t.Fatalf("wantLen status %d, got %d, body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			resp := decodeJSON[errorResponse](t, w)

			if resp.Error != tt.wantError {
				t.Fatalf("wantLen error %q, got %q", tt.wantError, resp.Error)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	h, r, pool, id := setupTest(t, true)
	defer pool.Close()

	defer deleteTestUser(t, pool, id)

	r.POST("/auth/login", h.Login)

	username := "test_logout"
	password := "testout123"
	body := fmt.Sprintf(`{
		"username": "%s",
		"password": "%s"
	}`, username, password)

	w := performJSONRequest(r, http.MethodPost, "/auth/login", body)
	t.Log(id)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	resp := decodeJSON[models.LoginResponse](t, w)

	userIDAccessJWT, err := auth.ParseJWTAccess(resp.AccessToken)
	if err != nil {
		t.Fatal(err)
	}

	if userIDAccessJWT != id {
		t.Fatalf("wrong access token user id: want %d, got %d", id, userIDAccessJWT)
	}

	cookies := w.Header()["Set-Cookie"]
	if len(cookies) == 0 {
		t.Fatal("no cookies in response")
	}

	cookie := cookies[0]
	cookieParts := strings.Split(cookie, ";")
	tokenPart := strings.SplitN(cookieParts[0], "=", 2)
	if len(tokenPart) != 2 {
		t.Fatalf("invalid cookie format: %s", cookie)
	}

	userIDRefreshJWT, err := auth.ParseJWTRefresh(tokenPart[1])
	if err != nil {
		t.Fatal(err)
	}
	if userIDRefreshJWT != id {
		t.Fatalf("wrong refresh token user id: wantLen %d, got %d", id, userIDRefreshJWT)
	}
}

func TestLogin_Invalid(t *testing.T) {
	h, r, pool, _ := setupTest(t, false)
	defer pool.Close()

	r.POST("/auth/login", h.Login)

	body := `{
		"username": "",
		"password": ""
	}`

	w := performJSONRequest(r, http.MethodPost, "/auth/login", body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("wantLen status %d, got %d, body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestLogout_Success(t *testing.T) {
	h, r, pool, id := setupTest(t, true)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)

	r.POST("/auth/logout", h.Logout)

	refreshToken, err := auth.GenerateRefreshJWT(id)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatal("want:", http.StatusOK, "got:", w.Code, "body:", w.Body.String())
	}

	respCookies := w.Result().Cookies()
	if len(respCookies) == 0 {
		t.Fatal("expected Set-Cookie in response, got none")
	}

	var refreshCookie *http.Cookie
	for _, c := range respCookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}

	if refreshCookie == nil {
		t.Fatal("refresh_token cookie not found in response")
	}

	if refreshCookie.Value != "" {
		t.Fatal("expected empty refresh_token after logout, got:", refreshCookie.Value)
	}

	// Обычно logout ещё помечает cookie как удалённую
	if refreshCookie.MaxAge != -1 {
		t.Log("warning: expected MaxAge = -1, got:", refreshCookie.MaxAge)
	}
}

func TestRefresh_NoCookie(t *testing.T) {
	h, r, pool, id := setupTest(t, true)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)

	r.POST("/auth/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatal("want status 401, got", w.Code, "body:", w.Body.String())
	}

	resp := decodeJSON[map[string]string](t, w)
	if resp["message"] != "http: named cookie not present" {
		t.Fatal("want cookie error, got", resp["message"])
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	h, r, pool, id := setupTest(t, true)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)
	r.POST("/auth/refresh", h.Refresh)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "invalid",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatal("wantLen status 401, got", w.Code, w.Body.String())
	}
	resp := decodeJSON[map[string]string](t, w)
	if !strings.Contains(resp["error"], "token is malformed") {
		t.Fatal("want token is malformed, got", resp["message"])
	}
}

func TestRefresh_ExpiredToken(t *testing.T) {
	h, r, pool, id := setupTest(t, true)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)
	jwtFunc := func() (string, error) {
		claims := jwt.MapClaims{
			"user_id": id,
			"exp":     time.Now().Add(0).Unix(),
			"iat":     time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(auth.GetRefreshSecret())
		if err != nil {
			return "", err
		}
		return signedToken, nil
	}
	r.POST("/auth/refresh", h.Refresh)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	RefJWT, err := jwtFunc()
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: RefJWT,
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatal("wantLen status 401, got", w.Code, w.Body.String())
	}
	resp := decodeJSON[map[string]string](t, w)

	if !strings.Contains(resp["error"], "token is expired") {
		t.Fatal("want: token is expired, got: ", resp["error"])
	}
}

func TestRefresh_UserID_NotFound(t *testing.T) {
	h, r, pool, _ := setupTest(t, false)
	defer pool.Close()
	jwtFunc := func() (string, error) {
		claims := jwt.MapClaims{
			"exp": time.Now().Add(15 * time.Minute).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(auth.GetRefreshSecret())
		if err != nil {
			return "", err
		}
		return signedToken, nil
	}
	r.POST("auth/refresh", h.Refresh)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	refJWT, err := jwtFunc()
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refJWT,
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := decodeJSON[map[string]string](t, w)
	if w.Code != http.StatusUnauthorized {
		t.Fatal("wantLen status 401, got", w.Code, w.Body.String())
	}
	if resp["error"] != "user_id not found" {
		t.Fatal("want: user_id not found. got: " + resp["error"])
	}
}

func TestRefresh_Success(t *testing.T) {
	h, r, pool, id := setupTest(t, true)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)
	refJWT, err := auth.GenerateRefreshJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	r.POST("/auth/refresh", h.Refresh)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refJWT,
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := decodeJSON[map[string]string](t, w)
	if w.Code != http.StatusOK {
		t.Fatal("wantLen status 200, got", w.Code, w.Body.String())
	}
	if resp["access_token"] == "" {
		t.Fatal("want: access_token")
	}
}
