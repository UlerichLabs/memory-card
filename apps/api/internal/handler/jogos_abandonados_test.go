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

type mockAbandonadosService struct {
	criarFn         func(ctx context.Context, input service.SalvarJogoAbandonadoInput) (*repository.JogoAbandonado, error)
	atualizarFn     func(ctx context.Context, input service.SalvarJogoAbandonadoInput) (*repository.JogoAbandonado, error)
	excluirFn       func(ctx context.Context, id int32, usuarioID int32) error
	listarFn        func(ctx context.Context, p service.ListarJogosAbandonadosInput) (*service.ResultadoListagemAbandonados, error)
	obterPorIDFn    func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoAbandonado, error)
	obterConsolesFn func(ctx context.Context, usuarioID int32) ([]string, error)
	obterTotalFn    func(ctx context.Context, usuarioID int32) (int64, error)
}

func (m *mockAbandonadosService) CriarJogoAbandonado(
	ctx context.Context,
	input service.SalvarJogoAbandonadoInput,
) (*repository.JogoAbandonado, error) {
	if m.criarFn != nil {
		return m.criarFn(ctx, input)
	}
	return nil, nil
}

func (m *mockAbandonadosService) AtualizarJogoAbandonado(
	ctx context.Context,
	input service.SalvarJogoAbandonadoInput,
) (*repository.JogoAbandonado, error) {
	if m.atualizarFn != nil {
		return m.atualizarFn(ctx, input)
	}
	return nil, nil
}

func (m *mockAbandonadosService) ExcluirJogoAbandonado(ctx context.Context, id int32, usuarioID int32) error {
	if m.excluirFn != nil {
		return m.excluirFn(ctx, id, usuarioID)
	}
	return nil
}

func (m *mockAbandonadosService) ListarJogosAbandonados(
	ctx context.Context,
	params service.ListarJogosAbandonadosInput,
) (*service.ResultadoListagemAbandonados, error) {
	if m.listarFn != nil {
		return m.listarFn(ctx, params)
	}
	return nil, nil
}

func (m *mockAbandonadosService) ObterJogoAbandonado(
	ctx context.Context,
	id int32,
	usuarioID int32,
) (*repository.JogoAbandonado, error) {
	if m.obterPorIDFn != nil {
		return m.obterPorIDFn(ctx, id, usuarioID)
	}
	return nil, nil
}

func (m *mockAbandonadosService) ObterConsoles(ctx context.Context, usuarioID int32) ([]string, error) {
	if m.obterConsolesFn != nil {
		return m.obterConsolesFn(ctx, usuarioID)
	}
	return nil, nil
}

func (m *mockAbandonadosService) ObterTotal(ctx context.Context, usuarioID int32) (int64, error) {
	if m.obterTotalFn != nil {
		return m.obterTotalFn(ctx, usuarioID)
	}
	return 0, nil
}

func setupAbandonadosTestRouter(svc JogosAbandonadosServicer, secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	tokens, _ := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewJogosAbandonadosHandler(svc)
	privadas.GET("/jogos-abandonados", h.Listar)
	privadas.GET("/jogos-abandonados/filtros", h.ObterFiltros)
	privadas.GET("/jogos-abandonados/total", h.ObterTotal)
	privadas.GET("/jogos-abandonados/:id", h.ObterDetalhes)
	privadas.POST("/jogos-abandonados", h.Criar)
	privadas.PUT("/jogos-abandonados/:id", h.Atualizar)
	privadas.DELETE("/jogos-abandonados/:id", h.Excluir)
	return router
}

func generateAbandonadosToken(t *testing.T, secret, subject, lang string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: lang,
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

func TestJogosAbandonadosHandler_Criar_Sucesso(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{
		criarFn: func(ctx context.Context, input service.SalvarJogoAbandonadoInput) (*repository.JogoAbandonado, error) {
			if input.UsuarioID != 42 {
				t.Fatalf("esperava usuarioID 42, obteve %d", input.UsuarioID)
			}
			return &repository.JogoAbandonado{
				ID:          1,
				UsuarioID:   input.UsuarioID,
				Nome:        input.Nome,
				Console:     input.Console,
				TempoJogado: 3600,
			}, nil
		},
	}
	router := setupAbandonadosTestRouter(svc, secret)
	token := generateAbandonadosToken(t, secret, "42", "pt-BR")

	body := []byte(`{"nome":"Demon's Souls","console":"PS3","tempo_jogado_horas":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperava status 201, obteve %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data repository.JogoAbandonado `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.ID != 1 || resp.Data.Nome != "Demon's Souls" {
		t.Fatalf("resposta inesperada: %+v", resp.Data)
	}
}

func TestJogosAbandonadosHandler_Atualizar_Sucesso(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{
		atualizarFn: func(ctx context.Context, input service.SalvarJogoAbandonadoInput) (*repository.JogoAbandonado, error) {
			if input.ID != 10 || input.UsuarioID != 42 {
				t.Fatalf("esperava ID 10 e UsuarioID 42, obteve ID=%d, User=%d", input.ID, input.UsuarioID)
			}
			return &repository.JogoAbandonado{
				ID:        10,
				UsuarioID: input.UsuarioID,
				Nome:      input.Nome,
				Console:   input.Console,
			}, nil
		},
	}
	router := setupAbandonadosTestRouter(svc, secret)
	token := generateAbandonadosToken(t, secret, "42", "pt-BR")

	body := []byte(`{"nome":"Demon's Souls Remake","console":"PS5"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/jogos-abandonados/10", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava status 200, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosAbandonadosHandler_Excluir_Sucesso(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			if id != 10 || usuarioID != 42 {
				t.Fatalf("esperava ID 10 e UsuarioID 42, obteve ID=%d, User=%d", id, usuarioID)
			}
			return nil
		},
	}
	router := setupAbandonadosTestRouter(svc, secret)
	token := generateAbandonadosToken(t, secret, "42", "pt-BR")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/jogos-abandonados/10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("esperava status 204, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestJogosAbandonadosHandler_NaoAutorizado_401(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{}
	router := setupAbandonadosTestRouter(svc, secret)

	endpoints := []struct {
		method string
		url    string
	}{
		{http.MethodPost, "/api/v1/jogos-abandonados"},
		{http.MethodGet, "/api/v1/jogos-abandonados"},
		{http.MethodGet, "/api/v1/jogos-abandonados/filtros"},
		{http.MethodGet, "/api/v1/jogos-abandonados/total"},
		{http.MethodGet, "/api/v1/jogos-abandonados/1"},
		{http.MethodPut, "/api/v1/jogos-abandonados/1"},
		{http.MethodDelete, "/api/v1/jogos-abandonados/1"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+"_"+ep.url, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("esperava status 401, obteve %d", w.Code)
			}
		})
	}
}

func TestJogosAbandonadosHandler_NaoEncontrado_404(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{
		obterPorIDFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoAbandonado, error) {
			return nil, service.ErrAbandonadoNaoEncontrado
		},
		atualizarFn: func(ctx context.Context, input service.SalvarJogoAbandonadoInput) (*repository.JogoAbandonado, error) {
			return nil, service.ErrAbandonadoNaoEncontrado
		},
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			return service.ErrAbandonadoNaoEncontrado
		},
	}
	router := setupAbandonadosTestRouter(svc, secret)
	token := generateAbandonadosToken(t, secret, "42", "pt-BR")

	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados/99", nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusNotFound {
		t.Fatalf("esperava 404 em GET, obteve %d", wGet.Code)
	}

	reqPut := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/jogos-abandonados/99",
		bytes.NewReader([]byte(`{"nome":"Jogo","console":"PS5"}`)),
	)
	reqPut.Header.Set("Authorization", "Bearer "+token)
	reqPut.Header.Set("Content-Type", "application/json")
	wPut := httptest.NewRecorder()
	router.ServeHTTP(wPut, reqPut)
	if wPut.Code != http.StatusNotFound {
		t.Fatalf("esperava 404 em PUT, obteve %d", wPut.Code)
	}

	reqDel := httptest.NewRequest(http.MethodDelete, "/api/v1/jogos-abandonados/99", nil)
	reqDel.Header.Set("Authorization", "Bearer "+token)
	wDel := httptest.NewRecorder()
	router.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusNotFound {
		t.Fatalf("esperava 404 em DELETE, obteve %d", wDel.Code)
	}
}

func TestJogosAbandonadosHandler_Validacoes_400(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{
		criarFn: func(ctx context.Context, input service.SalvarJogoAbandonadoInput) (*repository.JogoAbandonado, error) {
			if input.Nome == "NomeLongo" {
				return nil, service.ErrAbandonadoNomeMuitoLongo
			}
			if input.Console == "ConsoleLongo" {
				return nil, service.ErrAbandonadoConsoleMuitoLongo
			}
			if input.TempoJogadoHoras < 0 {
				return nil, service.ErrAbandonadoTempoInvalido
			}
			if input.Motivo != nil && *input.Motivo == "MotivoLongo" {
				return nil, service.ErrAbandonadoMotivoMuitoLongo
			}
			return nil, service.ErrAbandonadoNomeObrigatorio
		},
		listarFn: func(ctx context.Context, p service.ListarJogosAbandonadosInput) (*service.ResultadoListagemAbandonados, error) {
			if p.Ordenar == "invalido" {
				return nil, service.ErrAbandonadoOrdenarInvalido
			}
			return nil, nil
		},
	}
	router := setupAbandonadosTestRouter(svc, secret)
	token := generateAbandonadosToken(t, secret, "42", "pt-BR")

	tests := []struct {
		name       string
		method     string
		url        string
		body       string
		esperaCode string
	}{
		{
			name:       "json_malformado",
			method:     http.MethodPost,
			url:        "/api/v1/jogos-abandonados",
			body:       `{invalido`,
			esperaCode: "abandonados.entrada_invalida",
		},
		{
			name:       "data_invalida_no_parse",
			method:     http.MethodPost,
			url:        "/api/v1/jogos-abandonados",
			body:       `{"nome":"Jogo","console":"NES","abandonado_em":"data-errada"}`,
			esperaCode: "abandonados.data_invalida",
		},
		{
			name:       "id_invalido_string",
			method:     http.MethodGet,
			url:        "/api/v1/jogos-abandonados/abc",
			body:       "",
			esperaCode: "abandonados.id_invalido",
		},
		{
			name:       "id_invalido_zero",
			method:     http.MethodDelete,
			url:        "/api/v1/jogos-abandonados/0",
			body:       "",
			esperaCode: "abandonados.id_invalido",
		},
		{
			name:       "pagina_invalida_string",
			method:     http.MethodGet,
			url:        "/api/v1/jogos-abandonados?pagina=abc",
			body:       "",
			esperaCode: "abandonados.pagina_invalida",
		},
		{
			name:       "por_pagina_invalida_string",
			method:     http.MethodGet,
			url:        "/api/v1/jogos-abandonados?por_pagina=xyz",
			body:       "",
			esperaCode: "abandonados.por_pagina_invalida",
		},
		{
			name:       "ordenar_invalido",
			method:     http.MethodGet,
			url:        "/api/v1/jogos-abandonados?ordenar=invalido",
			body:       "",
			esperaCode: "abandonados.ordenar_invalido",
		},
		{
			name:       "nome_muito_longo",
			method:     http.MethodPost,
			url:        "/api/v1/jogos-abandonados",
			body:       `{"nome":"NomeLongo","console":"NES"}`,
			esperaCode: "abandonados.nome_muito_longo",
		},
		{
			name:       "console_muito_longo",
			method:     http.MethodPost,
			url:        "/api/v1/jogos-abandonados",
			body:       `{"nome":"Jogo","console":"ConsoleLongo"}`,
			esperaCode: "abandonados.console_muito_longo",
		},
		{
			name:       "motivo_muito_longo",
			method:     http.MethodPost,
			url:        "/api/v1/jogos-abandonados",
			body:       `{"nome":"Jogo","console":"NES","motivo":"MotivoLongo"}`,
			esperaCode: "abandonados.motivo_muito_longo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("esperava status 400, obteve %d: %s", w.Code, w.Body.String())
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
			if resp.Error.Codigo != tc.esperaCode {
				t.Fatalf("esperava codigo %s, obteve %s", tc.esperaCode, resp.Error.Codigo)
			}
			if resp.Error.Mensagem == "" {
				t.Fatal("mensagem de erro nao pode ser vazia")
			}
		})
	}
}

func TestJogosAbandonadosHandler_I18n_PtBr_E_En(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{
		criarFn: func(ctx context.Context, input service.SalvarJogoAbandonadoInput) (*repository.JogoAbandonado, error) {
			return nil, service.ErrAbandonadoNomeObrigatorio
		},
	}
	router := setupAbandonadosTestRouter(svc, secret)

	reqPt := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/jogos-abandonados",
		bytes.NewReader([]byte(`{"nome":""}`)),
	)
	reqPt.Header.Set("Authorization", "Bearer "+generateAbandonadosToken(t, secret, "1", "pt-BR"))
	reqPt.Header.Set("Accept-Language", "pt-BR")
	reqPt.Header.Set("Content-Type", "application/json")
	wPt := httptest.NewRecorder()
	router.ServeHTTP(wPt, reqPt)

	var respPt struct {
		Error struct {
			Codigo   string `json:"codigo"`
			Mensagem string `json:"mensagem"`
		} `json:"error"`
	}
	_ = json.Unmarshal(wPt.Body.Bytes(), &respPt)
	if respPt.Error.Mensagem != "O nome do jogo é obrigatório." {
		t.Fatalf("mensagem pt-BR incorreta: %s", respPt.Error.Mensagem)
	}

	reqEn := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/jogos-abandonados",
		bytes.NewReader([]byte(`{"nome":""}`)),
	)
	reqEn.Header.Set("Authorization", "Bearer "+generateAbandonadosToken(t, secret, "1", "en"))
	reqEn.Header.Set("Accept-Language", "en")
	reqEn.Header.Set("Content-Type", "application/json")
	wEn := httptest.NewRecorder()
	router.ServeHTTP(wEn, reqEn)

	var respEn struct {
		Error struct {
			Codigo   string `json:"codigo"`
			Mensagem string `json:"mensagem"`
		} `json:"error"`
	}
	_ = json.Unmarshal(wEn.Body.Bytes(), &respEn)
	if respEn.Error.Mensagem != "Game name is required." {
		t.Fatalf("mensagem en incorreta: %s", respEn.Error.Mensagem)
	}
}

func TestJogosAbandonadosHandler_ListarFiltrosTotal_Sucesso(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{
		listarFn: func(ctx context.Context, p service.ListarJogosAbandonadosInput) (*service.ResultadoListagemAbandonados, error) {
			return &service.ResultadoListagemAbandonados{
				Jogos:        []*repository.JogoAbandonado{{ID: 1, Nome: "Jogo 1"}},
				Total:        1,
				TotalPaginas: 1,
				Pagina:       1,
				PorPagina:    24,
			}, nil
		},
		obterConsolesFn: func(ctx context.Context, usuarioID int32) ([]string, error) {
			return []string{"Mega Drive", "SNES"}, nil
		},
		obterTotalFn: func(ctx context.Context, usuarioID int32) (int64, error) {
			return 7, nil
		},
	}
	router := setupAbandonadosTestRouter(svc, secret)
	token := generateAbandonadosToken(t, secret, "42", "pt-BR")

	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados", nil)
	reqList.Header.Set("Authorization", "Bearer "+token)
	wList := httptest.NewRecorder()
	router.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Fatalf("esperava 200 na listagem, obteve %d", wList.Code)
	}

	reqFiltros := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados/filtros", nil)
	reqFiltros.Header.Set("Authorization", "Bearer "+token)
	wFiltros := httptest.NewRecorder()
	router.ServeHTTP(wFiltros, reqFiltros)
	if wFiltros.Code != http.StatusOK {
		t.Fatalf("esperava 200 em filtros, obteve %d", wFiltros.Code)
	}
	var respFiltros struct {
		Data struct {
			Consoles []string `json:"consoles"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wFiltros.Body.Bytes(), &respFiltros)
	if len(respFiltros.Data.Consoles) != 2 {
		t.Fatalf("consoles inesperados: %+v", respFiltros.Data.Consoles)
	}

	reqTotal := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados/total", nil)
	reqTotal.Header.Set("Authorization", "Bearer "+token)
	wTotal := httptest.NewRecorder()
	router.ServeHTTP(wTotal, reqTotal)
	if wTotal.Code != http.StatusOK {
		t.Fatalf("esperava 200 em total, obteve %d", wTotal.Code)
	}
	var respTotal struct {
		Data struct {
			Total int64 `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wTotal.Body.Bytes(), &respTotal)
	if respTotal.Data.Total != 7 {
		t.Fatalf("total inesperado: %d", respTotal.Data.Total)
	}
}

func TestJogosAbandonadosHandler_Detalhes_Sucesso(t *testing.T) {
	secret := uuid.NewString()
	svc := &mockAbandonadosService{
		obterPorIDFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoAbandonado, error) {
			return &repository.JogoAbandonado{
				ID:        id,
				UsuarioID: usuarioID,
				Nome:      "Chrono Cross",
			}, nil
		},
	}
	router := setupAbandonadosTestRouter(svc, secret)
	token := generateAbandonadosToken(t, secret, "42", "pt-BR")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados/3", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("esperava 200 em detalhes, obteve %d", w.Code)
	}
	var resp struct {
		Data repository.JogoAbandonado `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.ID != 3 || resp.Data.Nome != "Chrono Cross" {
		t.Fatalf("dados inesperados: %+v", resp.Data)
	}
}
