package handlers

import (
	"blog/auth"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestDeletePost(t *testing.T) {
	h, r, p, id := setupTest(t)
	defer deleteTestUser(t, p, id)
	defer p.Close()
	IDs, err := createBlogH(t, p, strconv.Itoa(id), 1)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer deletePostsH(t, p, IDs)
	jwt, err := auth.GenerateAccessJWT(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	r.DELETE(`/posts/:id`, h.AuthMiddleware(), h.DeleteBlog)

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
			wantBodyErr: "",
			wantCode:    401,
			auth:        false,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.req.Header.Set("Authorization", "Bearer "+jwt)
		})
	}

}
