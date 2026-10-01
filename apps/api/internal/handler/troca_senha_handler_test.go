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

type trocaSenhaHandlerMock struct {
	trocarFn func(context.Context, string, string, string) error
}

func (mock trocaSenhaHandlerMock) Trocar(ctx context.Context, subject, senhaAtual, novaSenha string) error {
	return mock.trocarFn(ctx, subject, senhaAtual, novaSenha)
}

func TestTrocaSenhaHandler_Cenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{Tipo: "access", Idioma: "pt-BR", RegisteredClaims: jwt.RegisteredClaims{
		Subject: "42", Issuer: "memory-card", Audience: jwt.ClaimStrings{"access"}, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body string
		serviceErr error
		status     int
		codigo     string
		mensagem   string
	}{
		{"sucesso", `{"senha_atual":"SenhaAtual@123","nova_senha":"SenhaNova@123"}`, nil, http.StatusOK, "", "Senha alterada com sucesso."},
		{"senha atual incorreta", `{"senha_atual":"errada","nova_senha":"SenhaNova@123"}`, service.ErrSenhaAtualIncorreta, http.StatusBadRequest, "auth.password_change.current_password_invalid", "Senha atual incorreta."},
		{"senha fraca", `{"senha_atual":"SenhaAtual@123","nova_senha":"fraca"}`, service.ErrSenhaFraca, http.StatusBadRequest, "auth.password_change.weak_password", "A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial."},
		{"senha igual", `{"senha_atual":"SenhaAtual@123","nova_senha":"SenhaAtual@123"}`, service.ErrNovaSenhaIgualAtual, http.StatusBadRequest, "auth.password_change.same_password", "A nova senha deve ser diferente da atual."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewTrocaSenhaHandler(trocaSenhaHandlerMock{trocarFn: func(ctx context.Context, subject, atual, nova string) error {
				if subject != "42" || atual == "" || nova == "" {
					t.Fatal("argumentos incorretos")
				}
				return tc.serviceErr
			}})
			router := gin.New()
			middleware.GrupoPrivado(router, tokens).POST("/auth/trocar-senha", handler.TrocarSenha)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/trocar-senha", strings.NewReader(tc.body))
			request.Header.Set("Authorization", "Bearer "+raw)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status = %d", recorder.Code)
			}
			var response struct {
				Data struct {
					Mensagem string `json:"mensagem"`
				} `json:"data"`
				Error struct{ Codigo, Mensagem string } `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if tc.status == http.StatusOK && response.Data.Mensagem != tc.mensagem {
				t.Fatal(response.Data.Mensagem)
			}
			if tc.status != http.StatusOK && (response.Error.Codigo != tc.codigo || response.Error.Mensagem != tc.mensagem) {
				t.Fatalf("erro = %+v", response.Error)
			}
		})
	}
}

func TestTrocaSenhaHandler_SemAutenticacao(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewTrocaSenhaHandler(trocaSenhaHandlerMock{trocarFn: func(context.Context, string, string, string) error { return errors.New("nao deveria chamar") }})
	router := gin.New()
	tokens, err := service.NewAuthToken(uuid.NewString(), time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	middleware.GrupoPrivado(router, tokens).POST("/auth/trocar-senha", handler.TrocarSenha)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/auth/trocar-senha", strings.NewReader(`{}`)))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", recorder.Code)
	}
}
