package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

import "github.com/gin-gonic/gin"

func TestCORS_Requisicoes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, origin, method string
		status               int
		called               bool
	}{
		{"preflight", "http://localhost:5173", http.MethodOptions, 204, false},
		{"post autorizado", "http://localhost:5173", http.MethodPost, 401, true},
		{"segunda origem", "https://app.example.com", http.MethodPost, 401, true},
		{"origem proibida", "http://localhost:5174", http.MethodOptions, 403, false},
		{"post proibido", "https://evil.example.com", http.MethodPost, 403, false},
		{"sem origem", "", http.MethodPost, 401, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			router := gin.New()
			router.Use(CORS([]string{"http://localhost:5173", "https://app.example.com"}))
			router.POST("/api/v1/auth/register", func(c *gin.Context) { called = true; c.Status(http.StatusUnauthorized) })
			req := httptest.NewRequest(tc.method, "/api/v1/auth/register", nil)
			req.Header.Set("Origin", tc.origin)
			if tc.method == http.MethodOptions {
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "content-type,authorization,accept-language")
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != tc.status || called != tc.called {
				t.Fatalf("status %d, handler chamado %v", recorder.Code, called)
			}
			wantOrigin := tc.origin
			if tc.status == http.StatusForbidden {
				wantOrigin = ""
			}
			if recorder.Header().Get("Access-Control-Allow-Origin") != wantOrigin {
				t.Fatal("origem incorreta")
			}
			if recorder.Header().Get("Access-Control-Allow-Credentials") != "" {
				t.Fatal("cookies nao devem ser habilitados")
			}
			if !strings.Contains(strings.Join(recorder.Header().Values("Vary"), ","), "Origin") {
				t.Fatal("Vary ausente")
			}
			if tc.status == http.StatusNoContent {
				if recorder.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, PATCH, DELETE, OPTIONS" || recorder.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization, Accept-Language" {
					t.Fatal("preflight incompleto")
				}
				if recorder.Body.Len() != 0 {
					t.Fatal("preflight deve ter corpo vazio")
				}
			}
		})
	}
}
