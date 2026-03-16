package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{}
	r := gin.Default()
	var body map[string]string
	r.GET("/ping", h.Ping)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}

	err := json.Unmarshal([]byte(w.Body.String()), &body)
	if err != nil {
		t.Fatal(err)
	}
	if body["message"] != "pong" {
		t.Fatal("got: ", body["message"], "want: pong")
	}
}
