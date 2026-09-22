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

type mockJogosService struct {
	criarFn     func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error)
	atualizarFn func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error)
	excluirFn   func(ctx context.Context, id int32, usuarioID int32) error
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
	if resp.Data.Nome != "Super Mario World 2" || resp.Data.Nota != 11 {
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

