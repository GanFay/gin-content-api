package handlers

import (
	"blog/auth"
	"blog/models"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetAllPosts_Validation(t *testing.T) {
	_, _, p, id := setupTest(t, true)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	postsID, err := createBlogH(t, p, id, 12)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, postsID)

	testTable := []struct {
		name        string
		req         func(r *gin.Engine, h *Handler)
		reqTest     *http.Request
		wantBodyErr string
		wantCode    int
		auth        bool
	}{
		{
			name: "WithSearchTerm",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, `/posts?term=title`, nil),
			wantBodyErr: "",
			wantCode:    http.StatusOK,
			auth:        true,
		},
		{
			name: "InvalidLenLimit",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, "/posts?limit=1000", nil),
			wantBodyErr: "limit is too big",
			wantCode:    http.StatusBadRequest,
			auth:        true,
		},
		{
			name: "InvalidLimit",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, "/posts?limit=qwerty12345", nil),
			wantBodyErr: "limit must be int",
			wantCode:    http.StatusBadRequest,
			auth:        true,
		},
		{
			name: "InvalidOffset",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, "/posts?offset=qwerty12345", nil),
			wantBodyErr: "ERROR: invalid input syntax for type bigint: \"qwerty12345\" (SQLSTATE 22P02)",
			wantCode:    500,
			auth:        true,
		},
		{
			name: "Success_EmptyList",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, "/posts?offset=99999999", nil),
			wantBodyErr: "",
			wantCode:    http.StatusOK,
			auth:        true,
		},
		{
			name: "NoAuth",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, "/posts", nil),
			wantBodyErr: "missing authorization header",
			wantCode:    http.StatusUnauthorized,
			auth:        false,
		},
		{
			name: "ID_Invalid_ID",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts/:id`, h.AuthMiddleware(), h.GetByID)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%f", 9.5), nil),
			wantBodyErr: fmt.Sprintf(`invalid id: %f`, 9.5),
			wantCode:    http.StatusBadRequest,
			auth:        true,
		},
		{
			name: "ID_UserNotFound",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts/:id`, h.AuthMiddleware(), h.GetByID)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%d", 2147483647), nil),
			wantBodyErr: `no rows in result set`,
			wantCode:    http.StatusNotFound,
			auth:        true,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			h, r, _, _ := setupTest(t, false)
			testCase.req(r, h)
			w := httptest.NewRecorder()
			if testCase.auth {
				testCase.reqTest.Header.Set("Authorization", "Bearer "+jwt)
			}
			r.ServeHTTP(w, testCase.reqTest)
			if w.Code != testCase.wantCode {
				t.Fatal("want: ", testCase.wantCode, "got: ", w.Code, "body: ", w.Body.String())
			}
			if testCase.wantBodyErr != "" {
				resp := decodeJSON[map[string]string](t, w)
				if resp["error"] != testCase.wantBodyErr {
					t.Fatal("want: ", testCase.wantBodyErr, "got: ", resp["error"])
				}
			}
		})

	}
}

func TestGetAll_DefPagination(t *testing.T) {
	h, r, p, id := setupTest(t, true)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
	IDs, err := createBlogH(t, p, id, 12)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, IDs)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	req := httptest.NewRequest(http.MethodGet, `/posts`, nil)
	req.Header.Add("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("want: ", http.StatusOK, "got: ", w.Code, "body: ", w.Body.String())
	}
	resp := decodeJSON[map[string][]models.Post](t, w)
	if len(resp["posts"]) != 10 {
		t.Fatal("want: ", 10, ", got: ", len(resp["posts"]))
	}
}

func TestByID_Success(t *testing.T) {
	h, r, p, id := setupTest(t, true)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	r.GET(`/posts/:id`, h.AuthMiddleware(), h.GetByID)
	IDs, err := createBlogH(t, p, id, 1)
	defer deletePostsH(t, p, IDs)
	if err != nil {
		t.Fatal(err.Error())
	}
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%d", IDs[0]), nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("want: ", http.StatusOK, "got: ", w.Code, "body: ", w.Body.String())
	}
	post := decodeJSON[map[string]models.Post](t, w)
	if post["post"].ID != int64(IDs[0]) && post["post"].AuthorID != id {
		t.Fatal("wrong body")
	}
}
