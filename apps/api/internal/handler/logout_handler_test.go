package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type revogacaoRepositoryMock struct {
	consultar func(context.Context, string) (bool, error)
}

func (mock revogacaoRepositoryMock) EstaRevogado(ctx context.Context, jti string) (bool, error) {
	return mock.consultar(ctx, jti)
}

func (mock revogacaoRepositoryMock) Revogar(context.Context, string, time.Time) error {
	return errors.New("revogacao inesperada")
}

type logoutServiceMock struct {
	logout func(context.Context, string, string) error
}

func (mock logoutServiceMock) Logout(ctx context.Context, subject, raw string) error {
	return mock.logout(ctx, subject, raw)
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	sign := func(expiry time.Time) string {
		t.Helper()
		raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{Tipo: "access", Idioma: "en", RegisteredClaims: jwt.RegisteredClaims{
			Subject: "42", Issuer: "memory-card", Audience: jwt.ClaimStrings{"access"}, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), ExpiresAt: jwt.NewNumericDate(expiry),
		}}).SignedString([]byte(secret))
		if err != nil {
			t.Fatal(err)
		}
		return "Bearer " + raw
	}
	valid := sign(time.Now().Add(time.Minute))
	for _, tc := range []struct {
		name, header, body string
		serviceErr         error
		status             int
		codigo, mensagem   string
		called             bool
	}{
		{"sucesso", valid, `{"refresh_token":"refresh"}`, nil, 204, "", "", true},
		{"sem sessao", "", `{"refresh_token":"refresh"}`, nil, 401, "auth.session.unauthorized", "Não autorizado. Faça login novamente.", false},
		{"access expirado", sign(time.Now().Add(-time.Minute)), `{}`, nil, 401, "auth.session.unauthorized", "Não autorizado. Faça login novamente.", false},
		{"json invalido", valid, `{`, nil, 401, "auth.session.expired", "Session expired. Please log in again.", false},
		{"refresh invalido", valid, `{"refresh_token":"refresh"}`, service.ErrSessaoExpirada, 401, "auth.session.expired", "Session expired. Please log in again.", true},
		{"outro usuario", valid, `{"refresh_token":"refresh"}`, service.ErrTokenInvalido, 401, "auth.session.unauthorized", "Unauthorized. Please log in again.", true},
		{"falha no banco", valid, `{"refresh_token":"refresh"}`, errors.New("detalhe privado"), 500, "server.internal_error", "Internal server error.", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			h := NewLogoutHandler(logoutServiceMock{logout: func(ctx context.Context, subject, raw string) error {
				called = true
				usuario, ok := middleware.UsuarioDoContexto(ctx)
				if !ok || subject != "42" || usuario.ID != subject || raw != "refresh" {
					t.Error("contexto ou argumentos incorretos")
				}
				return tc.serviceErr
			}})
			router := gin.New()
			middleware.GrupoPrivado(router, tokens).POST("/auth/logout", h.Logout)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(tc.body))
			req.Header.Set("Authorization", tc.header)
			req.Header.Set("Accept-Language", "pt-BR")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != tc.status || called != tc.called {
				t.Fatalf("status = %d, service chamado = %v", recorder.Code, called)
			}
			if tc.status == http.StatusNoContent {
				if recorder.Body.Len() != 0 {
					t.Fatal("204 com corpo")
				}
				return
			}
			var result struct {
				Error struct {
					Codigo   string `json:"codigo"`
					Mensagem string `json:"mensagem"`
				} `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Error.Codigo != tc.codigo || result.Error.Mensagem != tc.mensagem {
				t.Fatalf("erro = %+v", result.Error)
			}
		})
	}
}
