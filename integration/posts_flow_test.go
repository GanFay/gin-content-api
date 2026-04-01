package integration

import (
	"blog/auth"
	"blog/models"
	"blog/router"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func deletePost(t *testing.T, p *pgxpool.Pool, post string) {
	t.Helper()
	_, err := p.Exec(context.Background(), "DELETE FROM posts WHERE title = $1", post)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func fullCreateUser(t *testing.T, p *pgxpool.Pool, username string) (accJWT string, refJWT string) {
	t.Helper()
	HP, err := auth.HashPassword("maksPassWord123")
	if err != nil {
		t.Fatal(err.Error())
	}
	randNum := rand.Int()
	id, err := createTestUser(t, p, username, fmt.Sprintf("maks%d@maks.com", randNum), HP)
	if err != nil {
		t.Fatal(err.Error())
	}
	accessJwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	refreshJwt, err := auth.GenerateRefreshJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	return accessJwt, refreshJwt
}

func TestPostsFlow_CreateGetUpdateDelete(t *testing.T) {
	// 0. Prepare
	h, p := setupTest(t)
	defer p.Close()
	r := router.SetupRouter(h)
	deleteTestUser(t, p, "maks")
	jwt, _ := fullCreateUser(t, p, "maks")
	defer deleteTestUser(t, p, "maks")
	deletePost(t, p, "test_flow_1")
	deletePost(t, p, "test_flow_2")

	body := `{
	"title": "test_flow_1",
	"content": "test1",
	"category": "test1",
	"tags": ["test1", "ng"]
}`
	updBody := `{
	"title": "test_flow_2",
	"content": "test2",
	"category": "test2",
	"tags": ["test2", "g"]
}`

	// 1. Create
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatal("want: ", http.StatusCreated, "got: ", w.Code, ", body: ", w.Body.String())
	}
	resp := decodeJSON[map[string]string](t, w)
	if resp["message"] != "post created successfully" {
		t.Fatal("want: post created successfully, get: ", resp["message"])
	}
	// 2. Get
	req = httptest.NewRequest(http.MethodGet, "/posts?term=test_flow_1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("want: ", http.StatusOK, "got: ", w.Code, ", body: ", w.Body.String())
	}
	respGet := decodeJSON[map[string][]models.Post](t, w)
	if respGet["posts"][0].Title != "test_flow_1" {
		t.Fatal("want: test_flow_1, get: ", respGet["posts"][0].Title)
	}
	postID := respGet["posts"][0].ID
	// 3. Update
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf(`/posts/%d`, postID), strings.NewReader(updBody))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("want: ", http.StatusOK, "got: ", w.Code, ", body: ", w.Body.String())
	}
	respUpd := decodeJSON[map[string]string](t, w)
	if respUpd["message"] != "successfully updated post" {
		t.Fatal("want: successfully updated post, get: ", respUpd["message"])
	}

	// 4. Delete
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf(`/posts/%d`, postID), nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatal("want: ", http.StatusNoContent, "got: ", w.Code, ", body: ", w.Body.String())
	}
}

func TestPostsFlow_NonOwnerCannotUpdate(t *testing.T) {
	// 0. prepare
	h, p := setupTest(t)
	defer p.Close()
	r := router.SetupRouter(h)
	defer deletePost(t, p, "test_flow_3")
	defer deleteTestUser(t, p, "maks1")
	defer deleteTestUser(t, p, "maks2")
	jwt1User, _ := fullCreateUser(t, p, "maks1")
	jwt2User, _ := fullCreateUser(t, p, "maks2")
	bodyCreate := `{
	"title": "test_flow_3",
	"content": "test3",
	"category": "test3",
	"tags": ["test3", "ng"]
}`
	bodyUpd := `{
	"title": "test_flow_4",
	"content": "test4",
	"category": "test4",
	"tags": ["test4", "me"]
}`
	// 1. Create
	req := httptest.NewRequest(http.MethodPost, `/posts`, strings.NewReader(bodyCreate))
	req.Header.Set("Authorization", "Bearer "+jwt1User)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatal("want: ", http.StatusCreated, "got: ", w.Code, ", body: ", w.Body.String())
	}
	respCreate := decodeJSON[map[string]string](t, w)
	if respCreate["message"] != "post created successfully" {
		t.Fatal("want: post created successfully, get: ", respCreate["message"])
	}
	// 2. Get
	req = httptest.NewRequest(http.MethodGet, "/posts?term=test_flow_3", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("want: ", http.StatusOK, "got: ", w.Code, ", body: ", w.Body.String())
	}
	respGet := decodeJSON[map[string][]models.Post](t, w)
	if respGet["posts"][0].Title != "test_flow_3" {
		t.Fatal("want: test_flow_3, get: ", respGet["posts"][0].Title)
	}

	// 3. Update
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf(`/posts/%d`, respGet["posts"][0].ID), strings.NewReader(bodyUpd))
	req.Header.Set("Authorization", "Bearer "+jwt2User)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatal("want: ", http.StatusOK, "got: ", w.Code, ", body: ", w.Body.String())
	}
	respUpd := decodeJSON[map[string]string](t, w)
	if respUpd["message"] != "not permission" {
		t.Fatal("want: not permission, get: ", respUpd["message"])
	}
}git
