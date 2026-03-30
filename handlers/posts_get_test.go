package handlers

import (
	"blog/auth"
	"blog/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetAllPosts_Validation(t *testing.T) {
	_, _, p, id := setupTest(t)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	jwt, err := auth.GenerateAccessJWT(id)
	strID := strconv.Itoa(id)
	postsID, err := createBlogH(t, p, strID, 12)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, postsID)

	testTable := []struct {
		name        string
		req         func(r *gin.Engine, h *Handler)
		reqTest     *http.Request
		wantLen     int
		wantBodyErr string
		wantCode    int
		auth        bool
	}{
		{
			name: "DefaultPagination",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, "/posts", nil),
			wantLen:     10,
			wantBodyErr: "",
			wantCode:    http.StatusOK,
			auth:        true,
		},
		{
			name: "WithSearchTerm",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts`, h.AuthMiddleware(), h.GetPosts)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, `/posts?term=title`, nil),
			wantLen:     -1,
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
			wantLen:     -1,
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
			wantLen:     -1,
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
			wantLen:     -1,
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
			wantLen:     0,
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
			wantLen:     -1,
			wantBodyErr: "missing authorization header",
			wantCode:    http.StatusUnauthorized,
			auth:        false,
		},
		{
			name: "ID_Success",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts/:id`, h.AuthMiddleware(), h.GetByID)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%d", postsID[0]), nil),
			wantLen:     -1,
			wantBodyErr: "",
			wantCode:    http.StatusOK,
			auth:        true,
		},
		{
			name: "ID_Invalid_ID",
			req: func(r *gin.Engine, h *Handler) {
				r.GET(`/posts/:id`, h.AuthMiddleware(), h.GetByID)
			},
			reqTest:     httptest.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%f", 9.5), nil),
			wantLen:     -1,
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
			wantLen:     -1,
			wantBodyErr: fmt.Sprintf(`user not found: %d`, 2147483647),
			wantCode:    http.StatusNotFound,
			auth:        true,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			h, r, _, id2 := setupTest(t)
			deleteTestUser(t, p, id2)
			testCase.req(r, h)
			w := httptest.NewRecorder()
			if testCase.auth {
				testCase.reqTest.Header.Set("Authorization", "Bearer "+jwt)
			}
			r.ServeHTTP(w, testCase.reqTest)
			if testCase.wantBodyErr == "" && testCase.wantLen != 0 {
				var posts map[string][]models.Post
				err = json.Unmarshal(w.Body.Bytes(), &posts)
				if len(posts["posts"]) != testCase.wantLen && testCase.wantLen != -1 {
					t.Fatal("wrong pagination len")

				}
			} else {
				var posts map[string]string
				err = json.Unmarshal(w.Body.Bytes(), &posts)
				if posts["error"] != testCase.wantBodyErr {

					t.Fatal("wrong error body", w.Code, posts, testCase.wantBodyErr)
				}
			}
			if err != nil {
				t.Fatal(err.Error())
			}
			if w.Code != testCase.wantCode {
				t.Fatal("want: ", testCase.wantCode, ", got: ", w.Code)
			}

		})

	}
}
