package handlers

import (
	"blog/auth"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeletePost(t *testing.T) {
	h, r, p, id := setupTest(t, true)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	IDs, err := createBlogH(t, p, id, 2)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, IDs)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	r.DELETE(`/posts/:id`, h.AuthMiddleware(), h.DeletePost)

	testTable := []struct {
		name        string
		req         *http.Request
		wantBodyErr string
		wantCode    int
		auth        bool
	}{
		{
			name:        "Unauthorized",
			req:         httptest.NewRequest(http.MethodDelete, fmt.Sprintf(`/posts/%d`, IDs[0]), nil),
			wantBodyErr: "missing authorization header",
			wantCode:    http.StatusUnauthorized,
			auth:        false,
		},
		{
			name:        "InvalidID",
			req:         httptest.NewRequest(http.MethodDelete, `/posts/zxc`, nil),
			wantBodyErr: "invalid id: zxc",
			wantCode:    http.StatusBadRequest,
			auth:        true,
		},
		{
			name:        "PostNotFound",
			req:         httptest.NewRequest(http.MethodDelete, `/posts/2147483647`, nil),
			wantBodyErr: "no rows in result set",
			wantCode:    http.StatusNotFound,
			auth:        true,
		},
		{
			name:        "Success",
			req:         httptest.NewRequest(http.MethodDelete, fmt.Sprintf(`/posts/%d`, IDs[1]), nil),
			wantBodyErr: "",
			wantCode:    http.StatusNoContent,
			auth:        true,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.auth {
				testCase.req.Header.Set("Authorization", "Bearer "+jwt)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, testCase.req)
			if w.Code != testCase.wantCode {
				t.Fatal("want: ", testCase.wantCode, "got: ", w.Code, "body: ", w.Body.String())
			}

			switch testCase.wantBodyErr {
			case "":
			default:
				var resp map[string]string
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err.Error())
				}
				if resp["error"] != testCase.wantBodyErr {
					t.Fatal("want: ", testCase.wantBodyErr, "got: ", resp["error"])
				}
			}
		})
	}
}

func TestDeletePost_NoAuthor(t *testing.T) {
	h, r, p, id := setupTest(t, true)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	IDs, err := createBlogH(t, p, id, 1)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, IDs)
	HP, err := auth.HashPassword("test123")
	if err != nil {
		t.Fatal(err.Error())
	}
	id2, err := createTestUser(t, h, "test", "test", HP)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deleteTestUser(t, p, id2)
	jwt, err := auth.GenerateAccessJWT(id2)
	if err != nil {
		t.Fatal(err.Error())
	}

	r.DELETE(`/posts/:id`, h.AuthMiddleware(), h.DeletePost)
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf(`/posts/%d`, IDs[0]), nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatal("want: status 403, got: ", w.Code, ", body: ", w.Body.String())
	}
	resp := decodeJSON[map[string]string](t, w)
	if resp["error"] != "not permission" {
		t.Fatal("want: not permission, got: ", resp["error"])
	}
}
