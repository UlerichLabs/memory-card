package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockHealthChecker struct {
	checkDatabaseFn func(ctx context.Context) error
}

func (m *mockHealthChecker) CheckDatabase(ctx context.Context) error {
	if m.checkDatabaseFn != nil {
		return m.checkDatabaseFn(ctx)
	}
	return nil
}

func TestHealthHandler_Check(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		serviceErr error
		wantCode   int
		wantBody   string
	}{
		{
			name:       "banco saudavel retorna 200 ok",
			serviceErr: nil,
			wantCode:   http.StatusOK,
			wantBody:   `{"status":"ok"}`,
		},
		{
			name:       "banco indisponivel retorna 503 unavailable",
			serviceErr: errors.New("database unavailable"),
			wantCode:   http.StatusServiceUnavailable,
			wantBody:   `{"status":"unavailable"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			checker := &mockHealthChecker{
				checkDatabaseFn: func(ctx context.Context) error {
					return tc.serviceErr
				},
			}
			h := NewHealthHandler(checker)

			router := gin.New()
			router.GET("/api/v1/health", h.Check)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			req.Header.Set("Origin", "http://localhost:5173")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tc.wantCode {
				t.Fatalf("got status %d, want %d", recorder.Code, tc.wantCode)
			}
			if recorder.Body.String() != tc.wantBody {
				t.Fatalf("got body %s, want %s", recorder.Body.String(), tc.wantBody)
			}
			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
			}
			if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "" {
				t.Errorf("Access-Control-Allow-Credentials = %q, want empty", got)
			}
		})
	}
}

func TestHealth_CrossOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &mockHealthChecker{
		checkDatabaseFn: func(ctx context.Context) error { return nil },
	}

	router := gin.New()
	router.GET("/api/v1/health", Health(checker))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("Health() got status %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("health response must be readable across origins")
	}
	if recorder.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatal("public health endpoint must not enable cross-origin credentials")
	}
}
