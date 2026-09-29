package handler

import (
	"bytes"
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
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type mockJogosService struct {
	criarFn         func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error)
	atualizarFn     func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error)
	excluirFn       func(ctx context.Context, id int32, usuarioID int32) error
	listarFn        func(ctx context.Context, params service.ListarJogosParams) (*service.ResultadoListagem, error)
	obterFiltrosFn  func(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error)
	obterDetalhesFn func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error)
}

func (m *mockJogosService) CriarJogoZerado(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
	if m.criarFn != nil {
		return m.criarFn(ctx, params)
	}
	return nil, nil
}

func (m *mockJogosService) AtualizarJogoZerado(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
	if m.atualizarFn != nil {
		return m.atualizarFn(ctx, params)
	}
	return nil, nil
}

func (m *mockJogosService) ExcluirJogoZerado(ctx context.Context, id int32, usuarioID int32) error {
	if m.excluirFn != nil {
		return m.excluirFn(ctx, id, usuarioID)
	}
	return nil
}

func (m *mockJogosService) ListarJogosZerados(ctx context.Context, params service.ListarJogosParams) (*service.ResultadoListagem, error) {
	if m.listarFn != nil {
		return m.listarFn(ctx, params)
	}
	return nil, nil
}

func (m *mockJogosService) ObterOpcoesFiltros(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error) {
	if m.obterFiltrosFn != nil {
		return m.obterFiltrosFn(ctx, usuarioID)
	}
	return nil, nil
}

func (m *mockJogosService) ObterDetalhesJogoZerado(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
	if m.obterDetalhesFn != nil {
		return m.obterDetalhesFn(ctx, id, usuarioID)
	}
	return nil, nil
}

func generateTestAccessToken(t *testing.T, secret, subject string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: "pt-BR",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    "memory-card",
			Audience:  jwt.ClaimStrings{"access"},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestJogosHandler_CriarJogo_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			if params.UsuarioID != 42 {
				t.Fatalf("esperava usuarioID 42, obteve %d", params.UsuarioID)
			}
			if params.TempoJogado != 45000 {
				t.Fatalf("esperava 45000s, obteve %d", params.TempoJogado)
			}
			return &repository.JogoZerado{
				ID:          1,
				UsuarioID:   params.UsuarioID,
				Nome:        params.Nome,
				Console:     params.Console,
				TempoJogado: params.TempoJogado,
				Nota:        params.Nota,
				Dificuldade: params.Dificuldade,
				Review:      params.Review,
			}, nil
		},
	}
	h := NewJogosHandler(svc)
	privadas.POST("/jogos", h.CriarJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	body := map[string]any{
		"nome":                  "Super Mario World",
		"console":               "SNES",
		"finalizado_em":         "2026-05-10T14:30:00Z",
		"tempo_jogado_horas":    12,
		"tempo_jogado_minutos":  30,
		"tempo_jogado_segundos": 0,
		"nota":                  10,
		"dificuldade":           "A",
		"review":                "Excelente campanha.",
		"destaque":              true,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperava status 201, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosHandler_CriarJogo_ConflitoDestaque(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, service.ErrDestaqueAnoConflito
		},
	}
	h := NewJogosHandler(svc)
	privadas.POST("/jogos", h.CriarJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	body := map[string]any{
		"nome":          "Zelda",
		"console":       "NES",
		"finalizado_em": "2026-05-10T14:30:00Z",
		"tempo_jogado":  3600,
		"nota":          10,
		"dificuldade":   "AA",
		"destaque":      true,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("esperava status 409, obteve %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Error struct {
			Codigo   string `json:"codigo"`
			Mensagem string `json:"mensagem"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error.Codigo != "jogos.destaque_ano_conflito" {
		t.Fatalf("codigo inesperado: %s", resp.Error.Codigo)
	}
	if resp.Error.Mensagem != "já existe um destaque para este ano" {
		t.Fatalf("mensagem inesperada: %s", resp.Error.Mensagem)
	}
}

func TestJogosHandler_CriarJogo_Validacao(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, service.ErrNotaInvalida
		},
	}
	h := NewJogosHandler(svc)
	privadas.POST("/jogos", h.CriarJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	body := map[string]any{
		"nome":          "Zelda",
		"console":       "NES",
		"finalizado_em": "2026-05-10T14:30:00Z",
		"nota":          15,
		"dificuldade":   "A",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperava status 400, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosHandler_AtualizarJogo_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			if params.ID != 10 || params.UsuarioID != 42 {
				t.Fatalf("esperava id 10 e usuarioID 42, obteve id=%d, usuarioID=%d", params.ID, params.UsuarioID)
			}
			return &repository.JogoZerado{
				ID:          params.ID,
				UsuarioID:   params.UsuarioID,
				Nome:        params.Nome,
				Console:     params.Console,
				TempoJogado: params.TempoJogado,
				Nota:        params.Nota,
				Dificuldade: params.Dificuldade,
				Review:      params.Review,
			}, nil
		},
	}
	h := NewJogosHandler(svc)
	privadas.PUT("/jogos/:id", h.AtualizarJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	body := map[string]any{
		"nome":                  "Super Mario World 2",
		"console":               "SNES",
		"finalizado_em":         "2026-06-10T14:30:00Z",
		"tempo_jogado_horas":    15,
		"tempo_jogado_minutos":  0,
		"tempo_jogado_segundos": 0,
		"nota":                  11,
		"dificuldade":           "AAA",
		"review":                "Review atualizada.",
		"destaque":              false,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/jogos/10", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava status 200, obteve %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data repository.JogoZerado `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Nome != "Super Mario World 2" || resp.Data.Nota != 11 || resp.Data.Review != "Review atualizada." {
		t.Fatalf("resposta inesperada: %+v", resp.Data)
	}
}

func TestJogosHandler_AtualizarJogo_NaoEncontrado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, service.ErrJogoNaoEncontrado
		},
	}
	h := NewJogosHandler(svc)
	privadas.PUT("/jogos/:id", h.AtualizarJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	body := map[string]any{
		"nome":          "Inexistente",
		"console":       "SNES",
		"finalizado_em": "2026-06-10T14:30:00Z",
		"nota":          8,
		"dificuldade":   "B",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/jogos/999", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("esperava status 404, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosHandler_AtualizarJogo_ConflitoDestaque(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, service.ErrDestaqueAnoConflito
		},
	}
	h := NewJogosHandler(svc)
	privadas.PUT("/jogos/:id", h.AtualizarJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	body := map[string]any{
		"nome":          "Zelda",
		"console":       "NES",
		"finalizado_em": "2026-05-10T14:30:00Z",
		"nota":          10,
		"dificuldade":   "AA",
		"destaque":      true,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/jogos/1", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("esperava status 409, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosHandler_AtualizarJogo_IDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{}
	h := NewJogosHandler(svc)
	privadas.PUT("/jogos/:id", h.AtualizarJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	req := httptest.NewRequest(http.MethodPut, "/api/v1/jogos/abc", bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperava status 400, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosHandler_ExcluirJogo_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			if id != 1 || usuarioID != 42 {
				t.Fatalf("esperava id=1 e usuarioID=42, obteve id=%d, usuarioID=%d", id, usuarioID)
			}
			return nil
		},
	}
	h := NewJogosHandler(svc)
	privadas.DELETE("/jogos/:id", h.ExcluirJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/jogos/1", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("esperava status 204, obteve %d: %s", w.Code, w.Body.String())
	}
	if w.Body.Len() != 0 {
		t.Fatalf("esperava body vazio para 204, obteve: %s", w.Body.String())
	}
}

func TestJogosHandler_ExcluirJogo_NaoEncontrado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			return service.ErrJogoNaoEncontrado
		},
	}
	h := NewJogosHandler(svc)
	privadas.DELETE("/jogos/:id", h.ExcluirJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/jogos/999", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("esperava status 404, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosHandler_ExcluirJogo_IDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	svc := &mockJogosService{}
	h := NewJogosHandler(svc)
	privadas.DELETE("/jogos/:id", h.ExcluirJogo)

	accessToken := generateTestAccessToken(t, secret, "42")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/jogos/0", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperava status 400, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosHandler_CamposMuitoLongos(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		codigo string
	}{
		{"nome", service.ErrNomeMuitoLongo, "jogos.nome_muito_longo"},
		{"console", service.ErrConsoleMuitoLongo, "jogos.console_muito_longo"},
		{"genero", service.ErrGeneroMuitoLongo, "jogos.genero_muito_longo"},
		{"tipo", service.ErrTipoMuitoLongo, "jogos.tipo_muito_longo"},
		{"review", service.ErrReviewMuitoLongo, "jogos.review_muito_longo"},
	}
	for _, tc := range tests {
		t.Run("criar_"+tc.name, func(t *testing.T) {
			secret := uuid.NewString()
			tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
			if err != nil {
				t.Fatal(err)
			}
			router := gin.New()
			privadas := middleware.GrupoPrivado(router, tokens)
			svc := &mockJogosService{criarFn: func(context.Context, repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
				return nil, tc.err
			}}
			privadas.POST("/jogos", NewJogosHandler(svc).CriarJogo)
			body := map[string]any{"nome": "Jogo", "console": "Console", "finalizado_em": "2026-05-10", "tempo_jogado": 1, "nota": 10, "dificuldade": "A"}
			requestBody, _ := json.Marshal(body)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(requestBody))
			request.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"codigo":"`+tc.codigo+`"`) {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
		t.Run("atualizar_"+tc.name, func(t *testing.T) {
			secret := uuid.NewString()
			tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
			if err != nil {
				t.Fatal(err)
			}
			router := gin.New()
			privadas := middleware.GrupoPrivado(router, tokens)
			svc := &mockJogosService{atualizarFn: func(context.Context, repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
				return nil, tc.err
			}}
			privadas.PUT("/jogos/:id", NewJogosHandler(svc).AtualizarJogo)
			body := map[string]any{"nome": "Jogo", "console": "Console", "finalizado_em": "2026-05-10", "tempo_jogado": 1, "nota": 10, "dificuldade": "A"}
			requestBody, _ := json.Marshal(body)
			request := httptest.NewRequest(http.MethodPut, "/api/v1/jogos/1", bytes.NewReader(requestBody))
			request.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"codigo":"`+tc.codigo+`"`) {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

type jogoTestPayload struct {
	Nome                string  `json:"nome"`
	Console             string  `json:"console"`
	Genero              string  `json:"genero,omitempty"`
	Tipo                string  `json:"tipo,omitempty"`
	IniciadoEm          *string `json:"iniciado_em,omitempty"`
	FinalizadoEm        string  `json:"finalizado_em"`
	TempoJogadoHoras    int     `json:"tempo_jogado_horas,omitempty"`
	TempoJogadoMinutos  int     `json:"tempo_jogado_minutos,omitempty"`
	TempoJogadoSegundos int     `json:"tempo_jogado_segundos,omitempty"`
	TempoJogado         *int32  `json:"tempo_jogado,omitempty"`
	Nota                int32   `json:"nota"`
	Dificuldade         string  `json:"dificuldade"`
	Review              string  `json:"review,omitempty"`
	Destaque            bool    `json:"destaque,omitempty"`
}

type jogoErrorResponse struct {
	Error struct {
		Codigo   string `json:"codigo"`
		Mensagem string `json:"mensagem"`
	} `json:"error"`
}

func TestJogosHandler_NaoAutorizado_SemToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	svc := &mockJogosService{}
	h := NewJogosHandler(svc)
	privadas.POST("/jogos", h.CriarJogo)
	privadas.PUT("/jogos/:id", h.AtualizarJogo)
	privadas.DELETE("/jogos/:id", h.ExcluirJogo)
	privadas.GET("/jogos", h.ListarJogos)
	privadas.GET("/jogos/filtros", h.ObterFiltros)

	endpoints := []struct {
		name   string
		method string
		url    string
		body   string
	}{
		{"criar_sem_token", http.MethodPost, "/api/v1/jogos", `{"nome":"Jogo"}`},
		{"atualizar_sem_token", http.MethodPut, "/api/v1/jogos/1", `{"nome":"Jogo"}`},
		{"excluir_sem_token", http.MethodDelete, "/api/v1/jogos/1", ""},
		{"listar_sem_token", http.MethodGet, "/api/v1/jogos", ""},
		{"filtros_sem_token", http.MethodGet, "/api/v1/jogos/filtros", ""},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.url, strings.NewReader(ep.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("esperava status 401, obteve %d", w.Code)
			}
			var resp jogoErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error.Codigo != "auth.session.unauthorized" {
				t.Fatalf("esperava codigo auth.session.unauthorized, obteve %s", resp.Error.Codigo)
			}
		})
	}
}

func TestJogosHandler_ValidacoesEntrada_400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	iniciadoInvalido := "data-invalida"
	tempoNegativo := int32(-1)

	tests := []struct {
		name           string
		rawJSON        string
		payload        *jogoTestPayload
		serviceErr     error
		expectedCodigo string
	}{
		{
			name:           "json_malformado",
			rawJSON:        `{"nome": "incompleto`,
			expectedCodigo: "jogos.invalid_input",
		},
		{
			name: "finalizado_em_obrigatorio_vazio",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "   ",
			},
			expectedCodigo: "jogos.finalizado_em_required",
		},
		{
			name: "finalizado_em_formato_invalido",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "invalida",
			},
			expectedCodigo: "jogos.finalizado_em_invalid",
		},
		{
			name: "iniciado_em_formato_invalido",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "2026-05-10",
				IniciadoEm:   &iniciadoInvalido,
			},
			expectedCodigo: "jogos.iniciado_em_invalid",
		},
		{
			name: "tempo_jogado_horas_negativo",
			payload: &jogoTestPayload{
				Nome:             "Jogo",
				Console:          "Console",
				FinalizadoEm:     "2026-05-10",
				TempoJogadoHoras: -2,
			},
			expectedCodigo: "jogos.tempo_jogado_invalid",
		},
		{
			name: "tempo_jogado_minutos_maior_59",
			payload: &jogoTestPayload{
				Nome:               "Jogo",
				Console:            "Console",
				FinalizadoEm:       "2026-05-10",
				TempoJogadoMinutos: 60,
			},
			expectedCodigo: "jogos.tempo_jogado_invalid",
		},
		{
			name: "tempo_jogado_segundos_maior_59",
			payload: &jogoTestPayload{
				Nome:                "Jogo",
				Console:             "Console",
				FinalizadoEm:        "2026-05-10",
				TempoJogadoSegundos: 60,
			},
			expectedCodigo: "jogos.tempo_jogado_invalid",
		},
		{
			name: "tempo_jogado_direto_negativo",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "2026-05-10",
				TempoJogado:  &tempoNegativo,
			},
			expectedCodigo: "jogos.tempo_jogado_invalid",
		},
		{
			name: "service_nome_obrigatorio",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "2026-05-10",
			},
			serviceErr:     service.ErrNomeObrigatorio,
			expectedCodigo: "jogos.nome_required",
		},
		{
			name: "service_console_obrigatorio",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "2026-05-10",
			},
			serviceErr:     service.ErrConsoleObrigatorio,
			expectedCodigo: "jogos.console_required",
		},
		{
			name: "service_finalizado_em_obrigatorio",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "2026-05-10",
			},
			serviceErr:     service.ErrFinalizadoEmObrigatorio,
			expectedCodigo: "jogos.finalizado_em_required",
		},
		{
			name: "service_tempo_jogado_invalido",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "2026-05-10",
			},
			serviceErr:     service.ErrTempoJogadoInvalido,
			expectedCodigo: "jogos.tempo_jogado_invalid",
		},
		{
			name: "service_dificuldade_invalida",
			payload: &jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "2026-05-10",
			},
			serviceErr:     service.ErrDificuldadeInvalida,
			expectedCodigo: "jogos.dificuldade_invalid",
		},
	}

	for _, tc := range tests {
		t.Run("criar_"+tc.name, func(t *testing.T) {
			router := gin.New()
			privadas := middleware.GrupoPrivado(router, tokens)
			svc := &mockJogosService{
				criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
					if tc.serviceErr != nil {
						return nil, tc.serviceErr
					}
					return &repository.JogoZerado{}, nil
				},
			}
			h := NewJogosHandler(svc)
			privadas.POST("/jogos", h.CriarJogo)

			var payloadBytes []byte
			if tc.rawJSON != "" {
				payloadBytes = []byte(tc.rawJSON)
			} else {
				payloadBytes, _ = json.Marshal(tc.payload)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(payloadBytes))
			req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("esperava status 400, obteve %d: %s", w.Code, w.Body.String())
			}
			var resp jogoErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error.Codigo != tc.expectedCodigo {
				t.Fatalf("esperava codigo %s, obteve %s", tc.expectedCodigo, resp.Error.Codigo)
			}
		})

		t.Run("atualizar_"+tc.name, func(t *testing.T) {
			router := gin.New()
			privadas := middleware.GrupoPrivado(router, tokens)
			svc := &mockJogosService{
				atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
					if tc.serviceErr != nil {
						return nil, tc.serviceErr
					}
					return &repository.JogoZerado{}, nil
				},
			}
			h := NewJogosHandler(svc)
			privadas.PUT("/jogos/:id", h.AtualizarJogo)

			var payloadBytes []byte
			if tc.rawJSON != "" {
				payloadBytes = []byte(tc.rawJSON)
			} else {
				payloadBytes, _ = json.Marshal(tc.payload)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/jogos/1", bytes.NewReader(payloadBytes))
			req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("esperava status 400, obteve %d: %s", w.Code, w.Body.String())
			}
			var resp jogoErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error.Codigo != tc.expectedCodigo {
				t.Fatalf("esperava codigo %s, obteve %s", tc.expectedCodigo, resp.Error.Codigo)
			}
		})
	}
}

func TestJogosHandler_NaoEncontrado_Cenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	scenarios := []struct {
		name string
		id   string
	}{
		{"inexistente", "999"},
		{"outro_usuario", "50"},
		{"ja_excluido", "100"},
	}

	for _, sc := range scenarios {
		t.Run("atualizar_"+sc.name, func(t *testing.T) {
			router := gin.New()
			privadas := middleware.GrupoPrivado(router, tokens)
			svc := &mockJogosService{
				atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
					return nil, service.ErrJogoNaoEncontrado
				},
			}
			h := NewJogosHandler(svc)
			privadas.PUT("/jogos/:id", h.AtualizarJogo)

			payload := jogoTestPayload{
				Nome:         "Jogo",
				Console:      "Console",
				FinalizadoEm: "2026-05-10",
				Nota:         10,
				Dificuldade:  "A",
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/jogos/"+sc.id, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Fatalf("esperava status 404, obteve %d", w.Code)
			}
			var resp jogoErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error.Codigo != "jogos.not_found" {
				t.Fatalf("esperava codigo jogos.not_found, obteve %s", resp.Error.Codigo)
			}
		})

		t.Run("excluir_"+sc.name, func(t *testing.T) {
			router := gin.New()
			privadas := middleware.GrupoPrivado(router, tokens)
			svc := &mockJogosService{
				excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
					return service.ErrJogoNaoEncontrado
				},
			}
			h := NewJogosHandler(svc)
			privadas.DELETE("/jogos/:id", h.ExcluirJogo)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/jogos/"+sc.id, nil)
			req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Fatalf("esperava status 404, obteve %d", w.Code)
			}
			var resp jogoErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error.Codigo != "jogos.not_found" {
				t.Fatalf("esperava codigo jogos.not_found, obteve %s", resp.Error.Codigo)
			}
		})
	}
}

func TestJogosHandler_ErroInterno_500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	unexpectedErr := errors.New("falha interna")

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	svc := &mockJogosService{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, unexpectedErr
		},
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, unexpectedErr
		},
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			return unexpectedErr
		},
		listarFn: func(ctx context.Context, params service.ListarJogosParams) (*service.ResultadoListagem, error) {
			return nil, unexpectedErr
		},
		obterFiltrosFn: func(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error) {
			return nil, unexpectedErr
		},
	}
	h := NewJogosHandler(svc)
	privadas.POST("/jogos", h.CriarJogo)
	privadas.PUT("/jogos/:id", h.AtualizarJogo)
	privadas.DELETE("/jogos/:id", h.ExcluirJogo)
	privadas.GET("/jogos", h.ListarJogos)
	privadas.GET("/jogos/filtros", h.ObterFiltros)

	payload := jogoTestPayload{
		Nome:         "Jogo",
		Console:      "Console",
		FinalizadoEm: "2026-05-10",
		Nota:         10,
		Dificuldade:  "A",
	}
	body, _ := json.Marshal(payload)

	t.Run("criar_500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("esperava status 500, obteve %d", w.Code)
		}
		var resp jogoErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Error.Codigo != "server.internal_error" {
			t.Fatalf("esperava codigo server.internal_error, obteve %s", resp.Error.Codigo)
		}
	})

	t.Run("atualizar_500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/jogos/1", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("esperava status 500, obteve %d", w.Code)
		}
		var resp jogoErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Error.Codigo != "server.internal_error" {
			t.Fatalf("esperava codigo server.internal_error, obteve %s", resp.Error.Codigo)
		}
	})

	t.Run("excluir_500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/jogos/1", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("esperava status 500, obteve %d", w.Code)
		}
		var resp jogoErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Error.Codigo != "server.internal_error" {
			t.Fatalf("esperava codigo server.internal_error, obteve %s", resp.Error.Codigo)
		}
	})

	t.Run("listar_500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("esperava status 500, obteve %d", w.Code)
		}
		var resp jogoErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Error.Codigo != "server.internal_error" {
			t.Fatalf("esperava codigo server.internal_error, obteve %s", resp.Error.Codigo)
		}
	})

	t.Run("filtros_500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/filtros", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("esperava status 500, obteve %d", w.Code)
		}
		var resp jogoErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Error.Codigo != "server.internal_error" {
			t.Fatalf("esperava codigo server.internal_error, obteve %s", resp.Error.Codigo)
		}
	})
}

func TestJogosHandler_ListarJogos_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	var capturado service.ListarJogosParams
	svc := &mockJogosService{
		listarFn: func(ctx context.Context, params service.ListarJogosParams) (*service.ResultadoListagem, error) {
			capturado = params
			return &service.ResultadoListagem{
				Jogos: []*repository.JogoZerado{
					{
						ID:          1,
						Nome:        "Super Mario World",
						Console:     "SNES",
						TempoJogado: 3600,
						Nota:        10,
						Dificuldade: "A",
					},
				},
				Total:        1,
				TotalPaginas: 1,
				Pagina:       *params.Pagina,
				PorPagina:    *params.PorPagina,
			}, nil
		},
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosHandler(svc)
	privadas.GET("/jogos", h.ListarJogos)

	url := "/api/v1/jogos?pagina=2&por_pagina=10&busca=mario&console=SNES&genero=Platform&tipo=Campanha&nota_min=8&nota_max=10&ano=2024&dificuldade=A"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava status 200, obteve %d: %s", w.Code, w.Body.String())
	}

	if capturado.UsuarioID != 42 {
		t.Fatalf("esperava usuarioID 42, obteve %d", capturado.UsuarioID)
	}
	if capturado.Pagina == nil || *capturado.Pagina != 2 {
		t.Fatalf("esperava pagina 2, obteve %v", capturado.Pagina)
	}
	if capturado.PorPagina == nil || *capturado.PorPagina != 10 {
		t.Fatalf("esperava por_pagina 10, obteve %v", capturado.PorPagina)
	}
	if capturado.Busca != "mario" || capturado.Console != "SNES" || capturado.Genero != "Platform" || capturado.Tipo != "Campanha" || capturado.Dificuldade != "A" {
		t.Fatalf("filtros de texto incorretos: %+v", capturado)
	}
	if capturado.NotaMin == nil || *capturado.NotaMin != 8 || capturado.NotaMax == nil || *capturado.NotaMax != 10 || capturado.Ano == nil || *capturado.Ano != 2024 {
		t.Fatalf("filtros numericos incorretos: %+v", capturado)
	}

	var resp struct {
		Data []map[string]any `json:"data"`
		Meta struct {
			Pagina       int   `json:"pagina"`
			PorPagina    int   `json:"por_pagina"`
			Total        int64 `json:"total"`
			TotalPaginas int   `json:"total_paginas"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Data) != 1 || resp.Meta.Total != 1 || resp.Meta.Pagina != 2 || resp.Meta.PorPagina != 10 || resp.Meta.TotalPaginas != 1 {
		t.Fatalf("resposta inesperada: %+v", resp)
	}
}

func TestJogosHandler_ListarJogos_Validacoes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	testes := []struct {
		nome           string
		query          string
		serviceErr     error
		esperadoCodigo string
	}{
		{"pagina nao inteira", "?pagina=abc", nil, "jogos.pagina_invalida"},
		{"pagina zero via service", "?pagina=0", service.ErrPaginaInvalida, "jogos.pagina_invalida"},
		{"por_pagina nao inteira", "?por_pagina=xyz", nil, "jogos.por_pagina_invalido"},
		{"por_pagina zero via service", "?por_pagina=0", service.ErrPorPaginaInvalido, "jogos.por_pagina_invalido"},
		{"nota_min nao inteira", "?nota_min=abc", nil, "jogos.nota_filtro_invalida"},
		{"nota_min invalida via service", "?nota_min=12", service.ErrNotaFiltroInvalida, "jogos.nota_filtro_invalida"},
		{"nota_max nao inteira", "?nota_max=abc", nil, "jogos.nota_filtro_invalida"},
		{"nota_max invalida via service", "?nota_max=0", service.ErrNotaFiltroInvalida, "jogos.nota_filtro_invalida"},
		{"faixa notas invalida", "?nota_min=10&nota_max=5", service.ErrNotaFaixaInvalida, "jogos.nota_faixa_invalida"},
		{"ano nao inteiro", "?ano=abc", nil, "jogos.ano_invalido"},
		{"ano invalido via service", "?ano=1969", service.ErrAnoInvalido, "jogos.ano_invalido"},
		{"dificuldade invalida via service", "?dificuldade=INVALIDA", service.ErrDificuldadeInvalida, "jogos.dificuldade_invalida"},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			svc := &mockJogosService{
				listarFn: func(ctx context.Context, params service.ListarJogosParams) (*service.ResultadoListagem, error) {
					if tt.serviceErr != nil {
						return nil, tt.serviceErr
					}
					return &service.ResultadoListagem{}, nil
				},
			}
			router := gin.New()
			privadas := middleware.GrupoPrivado(router, tokens)
			h := NewJogosHandler(svc)
			privadas.GET("/jogos", h.ListarJogos)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos"+tt.query, nil)
			req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("esperava status 400, obteve %d: %s", w.Code, w.Body.String())
			}
			var resp jogoErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error.Codigo != tt.esperadoCodigo {
				t.Fatalf("esperava codigo %s, obteve %s", tt.esperadoCodigo, resp.Error.Codigo)
			}
		})
	}
}

func TestJogosHandler_ObterFiltros_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	svc := &mockJogosService{
		obterFiltrosFn: func(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error) {
			if usuarioID != 42 {
				t.Fatalf("esperava usuarioID 42, obteve %d", usuarioID)
			}
			return &repository.OpcoesFiltros{
				Consoles: []string{"Game Boy Color", "PlayStation 5"},
				Generos:  []string{"Action", "Adventure"},
				Tipos:    []string{"Campanha", "DLC"},
				Anos:     []int{2026, 2025},
			}, nil
		},
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosHandler(svc)
	privadas.GET("/jogos/filtros", h.ObterFiltros)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/filtros", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava status 200, obteve %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Consoles []string `json:"consoles"`
			Generos  []string `json:"generos"`
			Tipos    []string `json:"tipos"`
			Anos     []int    `json:"anos"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Data.Consoles) != 2 || len(resp.Data.Generos) != 2 || len(resp.Data.Tipos) != 2 || len(resp.Data.Anos) != 2 {
		t.Fatalf("opcoes incorretas: %+v", resp.Data)
	}
}

func TestJogosHandler_ObterDetalhesJogo_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	igdbID := int32(1029)
	iniciadoEm := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	finalizadoEm := time.Date(2024, 5, 15, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 5, 15, 12, 0, 0, 0, time.UTC)

	svc := &mockJogosService{
		obterDetalhesFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
			if id != 20 {
				t.Fatalf("esperava id 20, obteve %d", id)
			}
			if usuarioID != 42 {
				t.Fatalf("esperava usuarioID 42, obteve %d", usuarioID)
			}
			return &repository.JogoZerado{
				ID:            20,
				Numero:        1,
				UsuarioID:     42,
				IgdbID:        &igdbID,
				Nome:          "The Legend of Zelda: Ocarina of Time",
				Console:       "Nintendo 64",
				Genero:        "Adventure, Role-playing (RPG)",
				Tipo:          "Campanha",
				IniciadoEm:    &iniciadoEm,
				FinalizadoEm:  finalizadoEm,
				TempoJogado:   128700,
				Nota:          11,
				Dificuldade:   "A",
				Review:        "Masterpiece.",
				Destaque:      true,
				IgdbCapaURL:   "//images.igdb.com/igdb/image/upload/t_thumb/co3nnx.jpg",
				IgdbDescricao: "A young boy named Link...",
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			}, nil
		},
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosHandler(svc)
	privadas.GET("/jogos/:id", h.ObterDetalhesJogo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/20", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava status 200, obteve %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			ID            int32  `json:"id"`
			Numero        int    `json:"numero"`
			IgdbID        *int32 `json:"igdb_id"`
			Nome          string `json:"nome"`
			Console       string `json:"console"`
			Genero        string `json:"genero"`
			Tipo          string `json:"tipo"`
			TempoJogado   int32  `json:"tempo_jogado"`
			Nota          int32  `json:"nota"`
			Dificuldade   string `json:"dificuldade"`
			Review        string `json:"review"`
			Destaque      bool   `json:"destaque"`
			IgdbCapaURL   string `json:"igdb_capa_url"`
			IgdbDescricao string `json:"igdb_descricao"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp.Data.ID != 20 || resp.Data.Numero != 1 || resp.Data.Nome != "The Legend of Zelda: Ocarina of Time" {
		t.Fatalf("dados incorretos: %+v", resp.Data)
	}
	if resp.Data.IgdbDescricao != "A young boy named Link..." {
		t.Fatalf("igdb_descricao incorreto: %q", resp.Data.IgdbDescricao)
	}
}

func TestJogosHandler_ObterDetalhesJogo_IDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	svc := &mockJogosService{}
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosHandler(svc)
	privadas.GET("/jogos/:id", h.ObterDetalhesJogo)

	casos := []string{"abc", "0", "-1"}
	for _, idParam := range casos {
		t.Run("id="+idParam, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/"+idParam, nil)
			req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("esperava status 400, obteve %d", w.Code)
			}

			var resp struct {
				Error struct {
					Codigo string `json:"codigo"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error.Codigo != "jogos.invalid_id" {
				t.Fatalf("esperava codigo jogos.invalid_id, obteve %s", resp.Error.Codigo)
			}
		})
	}
}

func TestJogosHandler_ObterDetalhesJogo_SemToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	svc := &mockJogosService{}
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosHandler(svc)
	privadas.GET("/jogos/:id", h.ObterDetalhesJogo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperava status 401, obteve %d", w.Code)
	}

	var resp struct {
		Error struct {
			Codigo string `json:"codigo"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error.Codigo != "auth.session.unauthorized" {
		t.Fatalf("esperava auth.session.unauthorized, obteve %s", resp.Error.Codigo)
	}
}

func TestJogosHandler_ObterDetalhesJogo_NaoEncontrado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	svc := &mockJogosService{
		obterDetalhesFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
			return nil, service.ErrJogoNaoEncontrado
		},
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosHandler(svc)
	privadas.GET("/jogos/:id", h.ObterDetalhesJogo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/999999", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestAccessToken(t, secret, "42"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("esperava status 404, obteve %d", w.Code)
	}

	var resp struct {
		Error struct {
			Codigo string `json:"codigo"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error.Codigo != "jogos.not_found" {
		t.Fatalf("esperava codigo jogos.not_found, obteve %s", resp.Error.Codigo)
	}
}

func TestJogosHandler_RotasSemConflito(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	svc := &mockJogosService{
		obterFiltrosFn: func(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error) {
			return &repository.OpcoesFiltros{Consoles: []string{"N64"}}, nil
		},
		obterDetalhesFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
			return &repository.JogoZerado{ID: id, Numero: 1, Nome: "Zelda"}, nil
		},
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosHandler(svc)

	privadas.GET("/jogos/filtros", h.ObterFiltros)
	privadas.GET("/jogos/igdb/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"igdb_id": c.Param("id")}})
	})
	privadas.GET("/jogos/:id", h.ObterDetalhesJogo)

	token := generateTestAccessToken(t, secret, "42")

	reqFiltros := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/filtros", nil)
	reqFiltros.Header.Set("Authorization", "Bearer "+token)
	wFiltros := httptest.NewRecorder()
	router.ServeHTTP(wFiltros, reqFiltros)
	if wFiltros.Code != http.StatusOK {
		t.Fatalf("filtros esperava status 200, obteve %d", wFiltros.Code)
	}

	reqIgdb := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/igdb/1029", nil)
	reqIgdb.Header.Set("Authorization", "Bearer "+token)
	wIgdb := httptest.NewRecorder()
	router.ServeHTTP(wIgdb, reqIgdb)
	if wIgdb.Code != http.StatusOK {
		t.Fatalf("igdb esperava status 200, obteve %d", wIgdb.Code)
	}

	reqJogo := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/42", nil)
	reqJogo.Header.Set("Authorization", "Bearer "+token)
	wJogo := httptest.NewRecorder()
	router.ServeHTTP(wJogo, reqJogo)
	if wJogo.Code != http.StatusOK {
		t.Fatalf("jogo esperava status 200, obteve %d", wJogo.Code)
	}
}
