package handler

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestJogosEmAndamentoHandler_NaoAutorizado(t *testing.T) {
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
	for _, endpoint := range []struct {
		method string
		url    string
	}{
		{http.MethodGet, "/api/v1/jogando"},
		{http.MethodPost, "/api/v1/jogando"},
		{http.MethodDelete, "/api/v1/jogando/1"},
	} {
		t.Run(endpoint.method, func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.url, bytes.NewBufferString(`{"nome":"Jogo"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("esperava 401, obteve %d", w.Code)
			}
		})
	}
}

func TestJogosEmAndamentoHandler_Validacoes(t *testing.T) {
	secret := uuid.NewString()
	token := tokenJogando(t, secret)
	testes := []struct {
		nome   string
		body   string
		erro   error
		codigo string
	}{
		{"nome ausente", `{"iniciado_em":"2026-09-01"}`, service.ErrJogandoNomeObrigatorio, "jogando.nome_obrigatorio"},
		{"data ausente", `{"nome":"Jogo"}`, service.ErrJogandoIniciadoObrigatorio, "jogando.iniciado_em_obrigatorio"},
		{"data inválida", `{"nome":"Jogo","iniciado_em":"invalida"}`, service.ErrJogandoIniciadoInvalido, "jogando.iniciado_em_invalido"},
		{"data futura", `{"nome":"Jogo","iniciado_em":"2026-10-02"}`, service.ErrJogandoIniciadoFuturo, "jogando.iniciado_em_futuro"},
		{"body inválido", `{invalido`, service.ErrJogandoEntradaInvalida, "jogando.entrada_invalida"},
	}
	for _, teste := range testes {
		t.Run(teste.nome, func(t *testing.T) {
			svc := &mockJogosEmAndamentoService{
				criarFn: func(context.Context, service.SalvarJogoEmAndamentoInput) (*repository.JogoEmAndamento, error) {
					return nil, teste.erro
				},
				listarFn: func(context.Context, int32) ([]*repository.JogoEmAndamento, error) {
					return []*repository.JogoEmAndamento{}, nil
				},
				excluirFn: func(context.Context, int32, int32) error { return nil },
			}
			router := setupJogandoRouter(svc, secret)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/jogando", bytes.NewBufferString(teste.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assertJogandoErrorCode(t, w, http.StatusBadRequest, teste.codigo)
		})
	}
}

func TestJogosEmAndamentoHandler_Excluir(t *testing.T) {
	secret := uuid.NewString()
	token := tokenJogando(t, secret)
	casos := []struct {
		nome   string
		id     string
		erro   error
		status int
	}{
		{"sucesso", "1", nil, http.StatusNoContent},
		{"id não numérico", "abc", nil, http.StatusBadRequest},
		{"id zero", "0", nil, http.StatusBadRequest},
		{"não encontrado", "2", service.ErrJogandoNaoEncontrado, http.StatusNotFound},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			svc := &mockJogosEmAndamentoService{
				criarFn: func(context.Context, service.SalvarJogoEmAndamentoInput) (*repository.JogoEmAndamento, error) {
					return nil, nil
				},
				listarFn: func(context.Context, int32) ([]*repository.JogoEmAndamento, error) {
					return []*repository.JogoEmAndamento{}, nil
				},
				excluirFn: func(context.Context, int32, int32) error { return caso.erro },
			}
			router := setupJogandoRouter(svc, secret)
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/jogando/"+caso.id, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != caso.status {
				t.Fatalf("esperava status %d, obteve %d: %s", caso.status, w.Code, w.Body.String())
			}
		})
	}
}

func TestJogosEmAndamentoHandler_ListaVaziaEIngles(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockJogosEmAndamentoService{
		criarFn: func(context.Context, service.SalvarJogoEmAndamentoInput) (*repository.JogoEmAndamento, error) {
			return nil, service.ErrJogandoNomeObrigatorio
		},
		listarFn: func(context.Context, int32) ([]*repository.JogoEmAndamento, error) {
			return []*repository.JogoEmAndamento{}, nil
		},
		excluirFn: func(context.Context, int32, int32) error { return nil },
	}
	router := setupJogandoRouter(svc, secret)
	token := tokenJogando(t, secret)
	get := httptest.NewRequest(http.MethodGet, "/api/v1/jogando", nil)
	get.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, get)
	if strings.TrimSpace(w.Body.String()) != `{"data":[]}` {
		t.Fatalf("esperava data vazio, obteve %s", w.Body.String())
	}
	post := httptest.NewRequest(http.MethodPost, "/api/v1/jogando", bytes.NewBufferString(`{}`))
	post.Header.Set("Authorization", "Bearer "+token)
	post.Header.Set("Accept-Language", "en")
	post.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, post)
	if !strings.Contains(w.Body.String(), "Start date is required.") {
		t.Fatalf("esperava mensagem em inglês, obteve %s", w.Body.String())
	}
}

func assertJogandoErrorCode(t *testing.T, response *httptest.ResponseRecorder, status int, codigo string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("esperava status %d, obteve %d: %s", status, response.Code, response.Body.String())
	}
	var body struct {
		Error struct {
			Codigo string `json:"codigo"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Codigo != codigo {
		t.Fatalf("esperava código %s, obteve %s", codigo, body.Error.Codigo)
	}
}
