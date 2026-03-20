package handlers

import (
	"blog/auth"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
