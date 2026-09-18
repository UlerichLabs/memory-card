package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type perfilServiceMock struct {
	perfil func(context.Context, string) (*repository.Usuario, error)
}

func (mock perfilServiceMock) Perfil(ctx context.Context, subject string) (*repository.Usuario, error) {
	return mock.perfil(ctx, subject)
}

func TestMeHandler_Cenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{Tipo: "access", Idioma: "en", RegisteredClaims: jwt.RegisteredClaims{
		Subject: "42", Issuer: "memory-card", Audience: jwt.ClaimStrings{"access"}, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	usuario := &repository.Usuario{ID: 42, Nome: "Lucas", Email: "lucas@example.com", Idioma: "en"}
	for _, tc := range []struct {
		name             string
		err              error
		status           int
		codigo, mensagem string
		semContexto      bool
	}{
		{"sucesso", nil, http.StatusOK, "", "", false},
		{"usuario removido", fmt.Errorf("perfil: %w", service.ErrTokenInvalido), http.StatusUnauthorized, "auth.session.unauthorized", "Unauthorized. Please log in again.", false},
		{"erro interno", errors.New("detalhe privado do banco"), http.StatusInternalServerError, "server.internal_error", "Internal server error.", false},
		{"sem contexto", service.ErrTokenInvalido, http.StatusUnauthorized, "auth.session.unauthorized", "Não autorizado. Faça login novamente.", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			h := NewMeHandler(perfilServiceMock{perfil: func(ctx context.Context, subject string) (*repository.Usuario, error) {
				called = true
				autenticado, ok := middleware.UsuarioDoContexto(ctx)
				if tc.semContexto {
					if ok || subject != "" {
						t.Error("contexto inesperado")
					}
				} else if !ok || autenticado.ID != subject || subject != "42" {
					t.Error("identidade nao foi propagada ao service")
				}
				return usuario, tc.err
			}})
			router := gin.New()
			if tc.semContexto {
				router.GET("/api/v1/me", h.Me)
			} else {
				middleware.GrupoPrivado(router, tokens).GET("/me", h.Me)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
			req.Header.Set("Authorization", "Bearer "+raw)
			req.Header.Set("Accept-Language", "pt-BR")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if !called || recorder.Code != tc.status {
				t.Fatalf("status = %d", recorder.Code)
			}
			var response struct {
				Data  repository.Usuario `json:"data"`
				Error struct {
					Codigo   string `json:"codigo"`
					Mensagem string `json:"mensagem"`
				} `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if tc.status == http.StatusOK && response.Data != *usuario {
				t.Fatal("perfil incorreto")
			}
			if response.Error.Codigo != tc.codigo || response.Error.Mensagem != tc.mensagem {
				t.Fatalf("erro = %+v", response.Error)
			}
			if strings.Contains(recorder.Body.String(), "detalhe privado") || strings.Contains(recorder.Body.String(), "senha") {
				t.Fatal("resposta expoe dados internos")
			}
		})
	}
}
