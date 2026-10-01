package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type mockJogosEmAndamentoService struct {
	criarFn   func(context.Context, service.SalvarJogoEmAndamentoInput) (*repository.JogoEmAndamento, error)
	listarFn  func(context.Context, int32) ([]*repository.JogoEmAndamento, error)
	excluirFn func(context.Context, int32, int32) error
}

func (m *mockJogosEmAndamentoService) Criar(ctx context.Context, input service.SalvarJogoEmAndamentoInput) (*repository.JogoEmAndamento, error) {
	return m.criarFn(ctx, input)
}
func (m *mockJogosEmAndamentoService) Listar(ctx context.Context, id int32) ([]*repository.JogoEmAndamento, error) {
	return m.listarFn(ctx, id)
}
func (m *mockJogosEmAndamentoService) Excluir(ctx context.Context, id int32, usuarioID int32) error {
	return m.excluirFn(ctx, id, usuarioID)
}

func setupJogandoRouter(svc JogosEmAndamentoServicer, secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	tokens, _ := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosEmAndamentoHandler(svc)
	privadas.GET("/jogando", h.Listar)
	privadas.POST("/jogando", h.Criar)
	privadas.DELETE("/jogando/:id", h.Excluir)
	return router
}

func tokenJogando(t *testing.T, secret string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo: "access", Idioma: "pt-BR",
		RegisteredClaims: jwt.RegisteredClaims{Subject: "42", Issuer: "memory-card", Audience: jwt.ClaimStrings{"access"}, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestJogosEmAndamentoHandler_CriarEListar(t *testing.T) {
	secret := uuid.NewString()
	var recebido service.SalvarJogoEmAndamentoInput
	svc := &mockJogosEmAndamentoService{
		criarFn: func(ctx context.Context, input service.SalvarJogoEmAndamentoInput) (*repository.JogoEmAndamento, error) {
			recebido = input
			return &repository.JogoEmAndamento{ID: 1, Nome: "Jogo"}, nil
		},
		listarFn: func(context.Context, int32) ([]*repository.JogoEmAndamento, error) {
			return []*repository.JogoEmAndamento{}, nil
		},
		excluirFn: func(context.Context, int32, int32) error { return nil },
	}
	router := setupJogandoRouter(svc, secret)
	token := tokenJogando(t, secret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/jogando", bytes.NewBufferString(`{"nome":" Jogo ","iniciado_em":"2026-09-30"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated || recebido.Nome != " Jogo " {
		t.Fatalf("resposta inesperada: %d %s", w.Code, w.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/api/v1/jogando", nil)
	get.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, get)
	if w.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", w.Code)
	}
	var response struct {
		Data []repository.JogoEmAndamento `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Data == nil {
		t.Fatalf("lista inesperada: %s", w.Body.String())
	}
}

func TestJogosEmAndamentoHandler_ValidacoesEAuth(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockJogosEmAndamentoService{
		criarFn: func(context.Context, service.SalvarJogoEmAndamentoInput) (*repository.JogoEmAndamento, error) {
			return nil, service.ErrJogandoIniciadoFuturo
		},
		listarFn: func(context.Context, int32) ([]*repository.JogoEmAndamento, error) {
			return []*repository.JogoEmAndamento{}, nil
		},
		excluirFn: func(context.Context, int32, int32) error { return nil },
	}
	router := setupJogandoRouter(svc, secret)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jogando", bytes.NewBufferString(`{"nome":"Jogo"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obteve %d", w.Code)
	}
}
