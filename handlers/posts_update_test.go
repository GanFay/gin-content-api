package handlers

import (
	"blog/auth"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestUpdatePosts_Valid(t *testing.T) {
	h, r, p, id := setupTest(t)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	IDs, err := createBlogH(t, p, strconv.Itoa(id), 3)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, IDs)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	r.PUT(`/posts/:id`, h.AuthMiddleware(), h.UpdatePost)
	TestTable := []struct {
		name        string
		body        string
		reqTest     func(body string) *http.Request
		wantBodyErr string
		wantCode    int
		auth        bool
	}{
		{
			name: "Unauthorized",
			body: `{
						"title": "test",
						"content": "test",
						"category": "test",
						"tags": ["test1", "test2"]
					}`,
			reqTest: func(body string) *http.Request {
				return httptest.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%d", IDs[0]), strings.NewReader(body))
			},
			wantBodyErr: "missing authorization header",
			wantCode:    http.StatusUnauthorized,
			auth:        false,
		},
		{
			name: "InvalidId",
			body: `{
						"title": "test",
						"content": "test",
						"category": "test",
						"tags": ["test1", "test2"]
					}`,
			reqTest: func(body string) *http.Request {
				return httptest.NewRequest(http.MethodPut, "/posts/qwe", strings.NewReader(body))
			},
			wantBodyErr: "invalid id qwe",
			wantCode:    http.StatusBadRequest,
			auth:        true,
		},
		{
			name: "Invalid_Json",
			body: `{
						"title": 123,
						"content": 123,
						"category": 123,
						"tags": "qwerty"
					}`,
			reqTest: func(body string) *http.Request {
				return httptest.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%d", IDs[1]), strings.NewReader(body))
			},
			wantBodyErr: "json: cannot unmarshal number into Go struct",
			wantCode:    http.StatusBadRequest,
			auth:        true,
		},
		{
			name: "PostNotFound",
			body: `{
						"title": "test",
						"content": "test",
						"category": "test",
						"tags": ["test1", "test2"]
					}`,
			reqTest: func(body string) *http.Request {
				return httptest.NewRequest(http.MethodPut, "/posts/2147483647", strings.NewReader(body))
			},
			wantBodyErr: "no rows in result set",
			wantCode:    http.StatusBadRequest,
			auth:        true,
		},
		{
			name: "Success",
			body: `{
						"title": "test",
						"content": "test",
						"category": "test",
						"tags": ["test1", "test2"]
					}`,
			reqTest: func(body string) *http.Request {
				return httptest.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%d", IDs[2]), strings.NewReader(body))
			},
			wantBodyErr: "",
			wantCode:    http.StatusOK,
			auth:        true,
		},
	}
	for _, testCase := range TestTable {
		t.Run(testCase.name, func(t *testing.T) {
			req := testCase.reqTest(testCase.body)
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
			switch testCase.wantBodyErr {
			case "":
				if resp["message"] != "successfully updated blog!" {
					t.Fatal("want: successfully updated, got: ", resp["message"])
				}
			default:
				if !strings.Contains(resp["error"], testCase.wantBodyErr) {
					t.Fatal("want: \"", testCase.wantBodyErr, "\", got: ", resp["error"])
				}
				if w.Code != testCase.wantCode {
					t.Fatal("want: ", testCase.wantCode, ", got: ", w.Code)
				}
			}
		})
	}
}

func TestUpdateBlog_NotOwner(t *testing.T) {
	h, r, p, id := setupTest(t)
	defer p.Close()
	defer deleteTestUser(t, p, id)
	IDs, err := createBlogH(t, p, strconv.Itoa(id), 1)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, IDs)
	HP, err := auth.HashPassword("user123")
	if err != nil {
		t.Fatal(err.Error())
	}
	id2, err := createTestUser(t, p, "user", "user", HP)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deleteTestUser(t, p, id2)
	jwt, err := auth.GenerateAccessJWT(id2)

	body := `{
						"title": "test",
						"content": "test",
						"category": "test",
						"tags": ["test1", "test2"]
					}`

	r.PUT(`/posts/:id`, h.AuthMiddleware(), h.UpdatePost)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%d", IDs[0]), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err.Error())
	}
	if w.Code != http.StatusForbidden {
		t.Fatal("got: ", w.Code, ", want: ", http.StatusForbidden)
	}
	if resp["message"] != "not permission" {
		t.Fatal("got: ", resp["message"], ", want: 'not permission'")
	}
}
