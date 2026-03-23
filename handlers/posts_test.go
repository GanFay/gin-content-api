package handlers

import (
	"blog/auth"
	"blog/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createBlogH(t *testing.T, pool *pgxpool.Pool, authorID string, n int) ([]int, error) {
	t.Helper()

	var postsID []int
	for j := 1; j <= n; j++ {
		var postID int

		err := pool.QueryRow(context.Background(), `

		INSERT INTO posts (author_id, title, content, category)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
			authorID,
			fmt.Sprintf(`title%d`, j),
			fmt.Sprintf(`content%d`, j),
			fmt.Sprintf(`category%d`, j),
		).Scan(&postID)

		postsID = append(postsID, postID)

		if err != nil {
			t.Log("err in createpost")
			return postsID, err

		}
	}

	return postsID, nil
}

func deletePostsH(t *testing.T, pool *pgxpool.Pool, IDs []int) {
	t.Helper()
	for _, i := range IDs {
		_, err := pool.Exec(context.Background(), `DELETE FROM posts WHERE id = $1`, i)
		if err != nil {
			t.Fatal(err)
		}

	}
}

func TestCreateBlog_Validation(t *testing.T) {
	h, r, pool, id := setupTest(t)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}

	r.POST("/posts", h.AuthMiddleware(), h.CreateBlog)

	testTable := []struct {
		testName string
		body     string
		expected string
		code     int
		auth     bool
	}{
		{
			testName: "Unauthorized",
			body: `{
					"title": "Test1234",
					"content": "test1",
					"category": "test2",
					"tags": ["test"]
				}`,
			expected: "missing authorization header",
			code:     401,
			auth:     false,
		},
		{
			testName: "InvalidJSON",
			body: `{
				"awda": awda
				"title": "Test1",
				"content": "test1",
				"category": "test1",
				"tags": ["test1"]
				}`,
			expected: "JSON can't unmarshal body",
			code:     400,
			auth:     true,
		},
		{
			testName: "InvalidTitle",
			body: `{
				"title": "1",
				"content": "test2",
				"category": "test2",
				"tags": ["test2"]
				}`,
			expected: "Incorrect title. It must be between 3 and 50 characters long.",
			code:     400,
			auth:     true,
		},
		{
			testName: "InvalidContent",
			body: `{
				"title": "Test3",
				"content": "",
				"category": "test3",
				"tags": ["test3"]
				}`,
			expected: "Incorrect content. It must be between 3 and 500 characters long.",
			code:     400,
			auth:     true,
		},
	}

	for _, testCase := range testTable {
		body := strings.NewReader(testCase.body)
		req := httptest.NewRequest(http.MethodPost, "/posts", body)
		req.Header.Set("Content-Type", "application/json")
		if testCase.auth {
			req.Header.Set("Authorization", "Bearer "+jwt)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err.Error())
		}
		if w.Code != testCase.code {
			t.Fatal("test: ", testCase.testName, ", want: ", testCase.code, ", got: ", w.Code)
		}
		if resp["error"] != testCase.expected {
			t.Fatal("test: ", testCase.testName, ", want: ", testCase.expected, ", got: ", resp["error"])
		}
	}

}

func TestCreateBlog_Success(t *testing.T) {
	h, r, pool, id := setupTest(t)
	defer pool.Close()
	defer deleteTestUser(t, pool, id)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	r.POST("/posts", h.AuthMiddleware(), h.CreateBlog)

	body := `{
		"title": "Test",
		"content": "test",
		"category": "test",
		"tags": ["test"]	
			}`

	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]string
	if w.Code != http.StatusCreated {
		t.Fatal("want: ", http.StatusCreated, ", got: ", w.Code)
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err.Error())
	}
	want := "post created successfully"
	if resp["message"] != want {
		t.Fatal("want: ", want, ", got: ", resp["message"])
	}
	_, err = h.DB.Exec(context.Background(), `DELETE FROM posts WHERE title=$1`, "Test")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestGetAllPosts_DefaultPagination(t *testing.T) {
	h, r, p, id := setupTest(t)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	testTable := []struct {
		name     string
		req      gin.IRoutes
		reqtest  *http.Request
		expected string
		code     int
	}{
		{
			name:    "DefaultPagination",
			req:     r.GET(`/posts`, h.AuthMiddleware(), h.GetAllPosts),
			reqtest: httptest.NewRequest(http.MethodGet, "/posts", nil),
		},
	}

	strID := strconv.Itoa(id)
	postsID, err := createBlogH(t, p, strID, 12)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, postsID)

	r.GET(`/posts`, h.AuthMiddleware(), h.GetAllPosts)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var posts map[string][]models.Post
	err = json.Unmarshal(w.Body.Bytes(), &posts)
	if err != nil {
		t.Fatal(err.Error())
	}

	if w.Code != http.StatusOK {
		t.Fatal("want: ", http.StatusOK, ", got: ", w.Code)
	}

	if len(posts["posts"]) != 10 {
		t.Fatal("wrong pagination len")
	}
}
