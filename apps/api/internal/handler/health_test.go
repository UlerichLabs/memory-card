package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type healthPinger struct{ err error }

func (p healthPinger) Ping(context.Context) error { return p.err }

func TestHealthCrossOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name string
		err  error
		code int
		body string
	}{
		{"healthy", nil, http.StatusOK, `{"status":"ok"}`},
		{"unavailable", errors.New("database unavailable"), http.StatusServiceUnavailable, `{"status":"unavailable"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/api/v1/health", Health(healthPinger{tc.err}))
			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			req.Header.Set("Origin", "http://localhost:5173")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != tc.code || response.Body.String() != tc.body {
				t.Fatalf("got %d %s; want %d %s", response.Code, response.Body, tc.code, tc.body)
			}
			if response.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Fatal("health response must be readable across origins, including failures")
			}
			if response.Header().Get("Access-Control-Allow-Credentials") != "" {
				t.Fatal("public health endpoint must not enable cross-origin credentials")
			}
		})
	}
}
