// Package handler contem os handlers HTTP da API.
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
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type mockListasService struct {
	criarListaFn            func(ctx context.Context, input service.CriarListaInput) (*service.ListaDetalhada, error)
	obterListaFn            func(ctx context.Context, id int64, usuarioID int32) (*service.ListaDetalhada, error)
	listarListasFn          func(ctx context.Context, usuarioID int32) ([]*service.ListaResumo, error)
	atualizarListaFn        func(ctx context.Context, input service.AtualizarListaInput) (*service.ListaDetalhada, error)
	excluirListaFn          func(ctx context.Context, id int64, usuarioID int32) error
	adicionarItemFn         func(ctx context.Context, input service.AdicionarItemInput) (*service.ListaItemDetalhe, error)
	adicionarItensLoteFn    func(ctx context.Context, input service.AdicionarItensLoteInput) (*service.AdicionarItensLoteResultado, error)
	excluirItemFn           func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error
	reordenarItensFn        func(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*service.ListaItemDetalhe, error)
	associarJogoZeradoFn    func(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*service.ListaItemDetalhe, error)
	desassociarJogoZeradoFn func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error)
	restaurarItemFn         func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error)
	sincronizarFranquiaFn   func(ctx context.Context, listaID int64, usuarioID int32) (*service.SincronizarResultado, error)
	previaDesafioFranquiaFn func(ctx context.Context, franquiaID int64, usuarioID int32) (*service.PreviaDesafioResultado, error)
}

func (m *mockListasService) CriarLista(ctx context.Context, input service.CriarListaInput) (*service.ListaDetalhada, error) {
	if m.criarListaFn != nil {
		return m.criarListaFn(ctx, input)
	}
	return &service.ListaDetalhada{}, nil
}

func (m *mockListasService) ObterLista(ctx context.Context, id int64, usuarioID int32) (*service.ListaDetalhada, error) {
	if m.obterListaFn != nil {
		return m.obterListaFn(ctx, id, usuarioID)
	}
	return &service.ListaDetalhada{}, nil
}

func (m *mockListasService) ListarListas(ctx context.Context, usuarioID int32) ([]*service.ListaResumo, error) {
	if m.listarListasFn != nil {
		return m.listarListasFn(ctx, usuarioID)
	}
	return []*service.ListaResumo{}, nil
}

func (m *mockListasService) AtualizarLista(ctx context.Context, input service.AtualizarListaInput) (*service.ListaDetalhada, error) {
	if m.atualizarListaFn != nil {
		return m.atualizarListaFn(ctx, input)
	}
	return &service.ListaDetalhada{}, nil
}

func (m *mockListasService) ExcluirLista(ctx context.Context, id int64, usuarioID int32) error {
	if m.excluirListaFn != nil {
		return m.excluirListaFn(ctx, id, usuarioID)
	}
	return nil
}

func (m *mockListasService) AdicionarItem(ctx context.Context, input service.AdicionarItemInput) (*service.ListaItemDetalhe, error) {
	if m.adicionarItemFn != nil {
		return m.adicionarItemFn(ctx, input)
	}
	return &service.ListaItemDetalhe{}, nil
}

func (m *mockListasService) AdicionarItensLote(ctx context.Context, input service.AdicionarItensLoteInput) (*service.AdicionarItensLoteResultado, error) {
	if m.adicionarItensLoteFn != nil {
		return m.adicionarItensLoteFn(ctx, input)
	}
	return &service.AdicionarItensLoteResultado{}, nil
}

func (m *mockListasService) ExcluirItem(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error {
	if m.excluirItemFn != nil {
		return m.excluirItemFn(ctx, itemID, listaID, usuarioID)
	}
	return nil
}

func (m *mockListasService) ReordenarItens(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*service.ListaItemDetalhe, error) {
	if m.reordenarItensFn != nil {
		return m.reordenarItensFn(ctx, listaID, usuarioID, itemIDs)
	}
	return []*service.ListaItemDetalhe{}, nil
}

func (m *mockListasService) AssociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*service.ListaItemDetalhe, error) {
	if m.associarJogoZeradoFn != nil {
		return m.associarJogoZeradoFn(ctx, itemID, listaID, usuarioID, jogoZeradoID)
	}
	return &service.ListaItemDetalhe{}, nil
}

func (m *mockListasService) DesassociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error) {
	if m.desassociarJogoZeradoFn != nil {
		return m.desassociarJogoZeradoFn(ctx, itemID, listaID, usuarioID)
	}
	return &service.ListaItemDetalhe{}, nil
}

func (m *mockListasService) RestaurarItem(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error) {
	if m.restaurarItemFn != nil {
		return m.restaurarItemFn(ctx, itemID, listaID, usuarioID)
	}
	return &service.ListaItemDetalhe{}, nil
}

func (m *mockListasService) PreviaDesafioFranquia(ctx context.Context, franquiaID int64, usuarioID int32) (*service.PreviaDesafioResultado, error) {
	if m.previaDesafioFranquiaFn != nil {
		return m.previaDesafioFranquiaFn(ctx, franquiaID, usuarioID)
	}
	return &service.PreviaDesafioResultado{}, nil
}

func (m *mockListasService) SincronizarFranquia(ctx context.Context, listaID int64, usuarioID int32) (*service.SincronizarResultado, error) {
	if m.sincronizarFranquiaFn != nil {
		return m.sincronizarFranquiaFn(ctx, listaID, usuarioID)
	}
	return &service.SincronizarResultado{}, nil
}

func setupListasTestRouter(t *testing.T, svc ListasServicer) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	handler := NewListasHandler(svc)

	privadas.GET("/listas", handler.ListarListas)
	privadas.POST("/listas", handler.CriarLista)
	privadas.GET("/listas/:id", handler.ObterLista)
	privadas.PUT("/listas/:id", handler.AtualizarLista)
	privadas.DELETE("/listas/:id", handler.ExcluirLista)
	privadas.POST("/listas/:id/itens", handler.AdicionarItem)
	privadas.POST("/listas/:id/itens/lote", handler.AdicionarItensLote)
	privadas.DELETE("/listas/:id/itens/:itemId", handler.ExcluirItem)
	privadas.POST("/listas/:id/itens/:itemId/restaurar", handler.RestaurarItem)
	privadas.PUT("/listas/:id/ordem", handler.ReordenarItens)
	privadas.PUT("/listas/:id/itens/:itemId/zeramento", handler.AssociarJogoZerado)
	privadas.DELETE("/listas/:id/itens/:itemId/zeramento", handler.DesassociarJogoZerado)
	privadas.POST("/listas/:id/sincronizar", handler.SincronizarFranquia)
	privadas.GET("/franquias/:igdbId/previa-desafio", handler.PreviaDesafioFranquia)

	token := generateTestToken(t, secret, "42")
	return router, token
}

func generateTestToken(t *testing.T, secret, subject string) string {
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

func TestListasHandler_Autenticacao_401(t *testing.T) {
	router, _ := setupListasTestRouter(t, &mockListasService{})

	endpoints := []struct {
		method string
		url    string
	}{
		{"GET", "/api/v1/listas"},
		{"POST", "/api/v1/listas"},
		{"GET", "/api/v1/listas/1"},
		{"PUT", "/api/v1/listas/1"},
		{"DELETE", "/api/v1/listas/1"},
		{"POST", "/api/v1/listas/1/itens"},
		{"DELETE", "/api/v1/listas/1/itens/1"},
		{"POST", "/api/v1/listas/1/itens/1/restaurar"},
		{"PUT", "/api/v1/listas/1/ordem"},
		{"PUT", "/api/v1/listas/1/itens/1/zeramento"},
		{"DELETE", "/api/v1/listas/1/itens/1/zeramento"},
		{"POST", "/api/v1/listas/1/sincronizar"},
		{"GET", "/api/v1/franquias/1/previa-desafio"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.url, func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, ep.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("esperava 401, obteve %d", w.Code)
			}
		})
	}
}

func TestListasHandler_IDsInvalidos_400(t *testing.T) {
	router, token := setupListasTestRouter(t, &mockListasService{})

	endpoints := []struct {
		method string
		url    string
	}{
		{"GET", "/api/v1/listas/abc"},
		{"GET", "/api/v1/listas/0"},
		{"GET", "/api/v1/listas/-5"},
		{"PUT", "/api/v1/listas/abc"},
		{"DELETE", "/api/v1/listas/abc"},
		{"POST", "/api/v1/listas/abc/itens"},
		{"DELETE", "/api/v1/listas/1/itens/abc"},
		{"PUT", "/api/v1/listas/abc/ordem"},
		{"PUT", "/api/v1/listas/1/itens/abc/zeramento"},
		{"DELETE", "/api/v1/listas/abc/itens/1/zeramento"},
		{"POST", "/api/v1/listas/abc/sincronizar"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.url, func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, ep.url, bytes.NewBufferString("{}"))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("esperava 400 para id invalido, obteve %d", w.Code)
			}
		})
	}
}

func TestListasHandler_ListarListas_200(t *testing.T) {
	svc := &mockListasService{
		listarListasFn: func(ctx context.Context, usuarioID int32) ([]*service.ListaResumo, error) {
			return []*service.ListaResumo{
				{ID: 1, Tipo: "fila", Nome: "Fila A"},
			}, nil
		},
	}
	router, token := setupListasTestRouter(t, svc)

	req, _ := http.NewRequest("GET", "/api/v1/listas", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", w.Code)
	}

	var res struct {
		Data  []service.ListaResumo `json:"data"`
		Dados []service.ListaResumo `json:"dados"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("falha ao deserializar: %v", err)
	}
	if len(res.Data) != 1 || res.Data[0].Nome != "Fila A" {
		t.Fatalf("resposta inesperada: %+v", res)
	}
}

func TestListasHandler_CriarLista_201_400_404(t *testing.T) {
	t.Run("sucesso 201", func(t *testing.T) {
		svc := &mockListasService{
			criarListaFn: func(ctx context.Context, input service.CriarListaInput) (*service.ListaDetalhada, error) {
				return &service.ListaDetalhada{
					ListaResumo: service.ListaResumo{ID: 10, Tipo: input.Tipo, Nome: input.Nome},
				}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"tipo":"fila","nome":"Minha Fila"}`
		req, _ := http.NewRequest("POST", "/api/v1/listas", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("esperava 201, obteve %d", w.Code)
		}
	})

	t.Run("400 validacao nome", func(t *testing.T) {
		svc := &mockListasService{
			criarListaFn: func(ctx context.Context, input service.CriarListaInput) (*service.ListaDetalhada, error) {
				return nil, service.ErrListaNomeObrigatorio
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"tipo":"fila","nome":""}`
		req, _ := http.NewRequest("POST", "/api/v1/listas", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", w.Code)
		}
	})

	t.Run("404 franquia inexistente", func(t *testing.T) {
		svc := &mockListasService{
			criarListaFn: func(ctx context.Context, input service.CriarListaInput) (*service.ListaDetalhada, error) {
				return nil, service.ErrListaFranquiaNaoEncontrada
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"tipo":"desafio","nome":"Zelda","origem":{"tipo":"franquia","igdb_id":999,"nome":"Zelda"},"itens":[{"igdb_id":1,"nome":"Zelda"}]}`
		req, _ := http.NewRequest("POST", "/api/v1/listas", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("esperava 404, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_ObterLista_200_404(t *testing.T) {
	t.Run("sucesso 200", func(t *testing.T) {
		svc := &mockListasService{
			obterListaFn: func(ctx context.Context, id int64, usuarioID int32) (*service.ListaDetalhada, error) {
				return &service.ListaDetalhada{
					ListaResumo: service.ListaResumo{ID: id, Tipo: "fila", Nome: "Fila Detalhe"},
				}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		req, _ := http.NewRequest("GET", "/api/v1/listas/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", w.Code)
		}
	})

	t.Run("404 nao encontrada", func(t *testing.T) {
		svc := &mockListasService{
			obterListaFn: func(ctx context.Context, id int64, usuarioID int32) (*service.ListaDetalhada, error) {
				return nil, service.ErrListaNaoEncontrada
			},
		}
		router, token := setupListasTestRouter(t, svc)

		req, _ := http.NewRequest("GET", "/api/v1/listas/999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("esperava 404, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_AtualizarLista_200_400_404(t *testing.T) {
	t.Run("sucesso 200", func(t *testing.T) {
		svc := &mockListasService{
			atualizarListaFn: func(ctx context.Context, input service.AtualizarListaInput) (*service.ListaDetalhada, error) {
				return &service.ListaDetalhada{
					ListaResumo: service.ListaResumo{ID: input.ID, Nome: *input.Nome},
				}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"nome":"Novo Nome"}`
		req, _ := http.NewRequest("PUT", "/api/v1/listas/1", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", w.Code)
		}
	})

	t.Run("400 campo imutavel", func(t *testing.T) {
		svc := &mockListasService{
			atualizarListaFn: func(ctx context.Context, input service.AtualizarListaInput) (*service.ListaDetalhada, error) {
				return nil, service.ErrListaCampoImutavel
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"tipo":"desafio"}`
		req, _ := http.NewRequest("PUT", "/api/v1/listas/1", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_ExcluirLista_204_404(t *testing.T) {
	t.Run("sucesso 204", func(t *testing.T) {
		svc := &mockListasService{
			excluirListaFn: func(ctx context.Context, id int64, usuarioID int32) error {
				return nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		req, _ := http.NewRequest("DELETE", "/api/v1/listas/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("esperava 204, obteve %d", w.Code)
		}
	})

	t.Run("404 nao encontrada", func(t *testing.T) {
		svc := &mockListasService{
			excluirListaFn: func(ctx context.Context, id int64, usuarioID int32) error {
				return service.ErrListaNaoEncontrada
			},
		}
		router, token := setupListasTestRouter(t, svc)

		req, _ := http.NewRequest("DELETE", "/api/v1/listas/999", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("esperava 404, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_Itens_201_400_404_409_204(t *testing.T) {
	t.Run("adicionar item 201", func(t *testing.T) {
		svc := &mockListasService{
			adicionarItemFn: func(ctx context.Context, input service.AdicionarItemInput) (*service.ListaItemDetalhe, error) {
				return &service.ListaItemDetalhe{ID: 5, Nome: input.Nome, Posicao: 1}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"nome":"Chrono Trigger"}`
		req, _ := http.NewRequest("POST", "/api/v1/listas/1/itens", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("esperava 201, obteve %d", w.Code)
		}
	})

	t.Run("adicionar item 409 duplicado", func(t *testing.T) {
		svc := &mockListasService{
			adicionarItemFn: func(ctx context.Context, input service.AdicionarItemInput) (*service.ListaItemDetalhe, error) {
				return nil, service.ErrListaItemDuplicado
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"nome":"Chrono Trigger","igdb_id":100}`
		req, _ := http.NewRequest("POST", "/api/v1/listas/1/itens", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("esperava 409, obteve %d", w.Code)
		}
	})

	t.Run("adicionar item 400 contagem", func(t *testing.T) {
		svc := &mockListasService{
			adicionarItemFn: func(ctx context.Context, input service.AdicionarItemInput) (*service.ListaItemDetalhe, error) {
				return nil, service.ErrListaItensNaoPermitidos
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"nome":"Chrono Trigger"}`
		req, _ := http.NewRequest("POST", "/api/v1/listas/1/itens", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", w.Code)
		}
	})

	t.Run("excluir item 204", func(t *testing.T) {
		svc := &mockListasService{
			excluirItemFn: func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error {
				return nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		req, _ := http.NewRequest("DELETE", "/api/v1/listas/1/itens/5", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("esperava 204, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_AdicionarItensLote_201_Erros(t *testing.T) {
	t.Run("adiciona e conta repetidos", func(t *testing.T) {
		svc := &mockListasService{
			adicionarItensLoteFn: func(ctx context.Context, input service.AdicionarItensLoteInput) (*service.AdicionarItensLoteResultado, error) {
				if len(input.Itens) != 2 {
					t.Fatalf("esperava 2 itens, obteve %d", len(input.Itens))
				}
				return &service.AdicionarItensLoteResultado{Adicionados: 2, JaExistentes: 1}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/listas/1/itens/lote", bytes.NewBufferString(
			`{"itens":[{"igdb_id":1,"nome":"Um"},{"igdb_id":2,"nome":"Dois"}]}`,
		))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"adicionados":2`) {
			t.Fatalf("resposta inesperada: %d %s", response.Code, response.Body.String())
		}
	})

	t.Run("erro de validação", func(t *testing.T) {
		svc := &mockListasService{
			adicionarItensLoteFn: func(context.Context, service.AdicionarItensLoteInput) (*service.AdicionarItensLoteResultado, error) {
				return nil, service.ErrListaRegraInvalida
			},
		}
		router, token := setupListasTestRouter(t, svc)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/listas/1/itens/lote", bytes.NewBufferString(
			`{"itens":[{"igdb_id":0,"nome":"Inválido"}]}`,
		))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", response.Code)
		}
	})
}

func TestListasHandler_Reordenar_200_400(t *testing.T) {
	t.Run("reordenar 200", func(t *testing.T) {
		svc := &mockListasService{
			reordenarItensFn: func(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*service.ListaItemDetalhe, error) {
				return []*service.ListaItemDetalhe{
					{ID: itemIDs[0], Posicao: 1},
					{ID: itemIDs[1], Posicao: 2},
				}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"item_ids":[2,1]}`
		req, _ := http.NewRequest("PUT", "/api/v1/listas/1/ordem", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", w.Code)
		}
	})

	t.Run("reordenar 400", func(t *testing.T) {
		svc := &mockListasService{
			reordenarItensFn: func(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*service.ListaItemDetalhe, error) {
				return nil, service.ErrListaOrdemInvalida
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"item_ids":[2,2]}`
		req, _ := http.NewRequest("PUT", "/api/v1/listas/1/ordem", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_Zeramento_200_404(t *testing.T) {
	t.Run("associar zeramento 200", func(t *testing.T) {
		svc := &mockListasService{
			associarJogoZeradoFn: func(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*service.ListaItemDetalhe, error) {
				return &service.ListaItemDetalhe{ID: itemID, Zerado: true}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"jogo_zerado_id":15}`
		req, _ := http.NewRequest("PUT", "/api/v1/listas/1/itens/2/zeramento", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", w.Code)
		}
	})

	t.Run("associar zeramento 404 jogo nao encontrado", func(t *testing.T) {
		svc := &mockListasService{
			associarJogoZeradoFn: func(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*service.ListaItemDetalhe, error) {
				return nil, service.ErrJogoNaoEncontrado
			},
		}
		router, token := setupListasTestRouter(t, svc)

		body := `{"jogo_zerado_id":999}`
		req, _ := http.NewRequest("PUT", "/api/v1/listas/1/itens/2/zeramento", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("esperava 404, obteve %d", w.Code)
		}

		var errResp struct {
			Error struct {
				Codigo   string `json:"codigo"`
				Mensagem string `json:"mensagem"`
			} `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("falha ao deserializar erro: %v", err)
		}
		if errResp.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo 'jogos.not_found', obteve '%s'", errResp.Error.Codigo)
		}
	})

	t.Run("desassociar zeramento 200", func(t *testing.T) {
		svc := &mockListasService{
			desassociarJogoZeradoFn: func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error) {
				return &service.ListaItemDetalhe{ID: itemID, Zerado: false}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		req, _ := http.NewRequest("DELETE", "/api/v1/listas/1/itens/2/zeramento", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_Sincronizar_200_400(t *testing.T) {
	t.Run("sincronizar 200", func(t *testing.T) {
		svc := &mockListasService{
			sincronizarFranquiaFn: func(ctx context.Context, listaID int64, usuarioID int32) (*service.SincronizarResultado, error) {
				return &service.SincronizarResultado{
					ListaDetalhada: service.ListaDetalhada{
						ListaResumo: service.ListaResumo{ID: listaID, Nome: "Zelda"},
					},
					Adicionados: 3,
				}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)

		req, _ := http.NewRequest("POST", "/api/v1/listas/1/sincronizar", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", w.Code)
		}
	})

	t.Run("sincronizar 400 nao permitida", func(t *testing.T) {
		svc := &mockListasService{
			sincronizarFranquiaFn: func(ctx context.Context, listaID int64, usuarioID int32) (*service.SincronizarResultado, error) {
				return nil, service.ErrListaSincronizacaoNaoPermitida
			},
		}
		router, token := setupListasTestRouter(t, svc)

		req, _ := http.NewRequest("POST", "/api/v1/listas/1/sincronizar", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_PreviaDesafioFranquia(t *testing.T) {
	t.Run("id invalido retorna 400", func(t *testing.T) {
		router, token := setupListasTestRouter(t, &mockListasService{})
		req, _ := http.NewRequest("GET", "/api/v1/franquias/abc/previa-desafio", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", w.Code)
		}
	})

	t.Run("franquia nao encontrada retorna 404", func(t *testing.T) {
		svc := &mockListasService{
			previaDesafioFranquiaFn: func(ctx context.Context, franquiaID int64, usuarioID int32) (*service.PreviaDesafioResultado, error) {
				return nil, service.ErrListaFranquiaNaoEncontrada
			},
		}
		router, token := setupListasTestRouter(t, svc)
		req, _ := http.NewRequest("GET", "/api/v1/franquias/999/previa-desafio", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("esperava 404, obteve %d", w.Code)
		}
	})

	t.Run("franquia sem jogos retorna 400", func(t *testing.T) {
		svc := &mockListasService{
			previaDesafioFranquiaFn: func(ctx context.Context, franquiaID int64, usuarioID int32) (*service.PreviaDesafioResultado, error) {
				return nil, service.ErrListaFranquiaSemJogos
			},
		}
		router, token := setupListasTestRouter(t, svc)
		req, _ := http.NewRequest("GET", "/api/v1/franquias/10/previa-desafio", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", w.Code)
		}
	})

	t.Run("sucesso 200", func(t *testing.T) {
		svc := &mockListasService{
			previaDesafioFranquiaFn: func(ctx context.Context, franquiaID int64, usuarioID int32) (*service.PreviaDesafioResultado, error) {
				return &service.PreviaDesafioResultado{
					Franquia:       service.PreviaFranquiaInfo{IgdbID: franquiaID, Nome: "Zelda"},
					Total:          1,
					TotalSugeridos: 1,
					Jogos: []*service.PreviaJogoItem{
						{IgdbID: 1025, Nome: "The Legend of Zelda", Tipo: "principal", Sugerido: true},
					},
				}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)
		req, _ := http.NewRequest("GET", "/api/v1/franquias/596/previa-desafio", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", w.Code)
		}
		var body struct {
			Data service.PreviaDesafioResultado `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Data.TotalSugeridos != 1 || len(body.Data.Jogos) != 1 || body.Data.Jogos[0].Tipo != "principal" || !body.Data.Jogos[0].Sugerido {
			t.Fatalf("resposta inesperada: %+v", body.Data)
		}
	})
}

func TestListasHandler_RestaurarItem(t *testing.T) {
	t.Run("restauracao nao permitida retorna 400", func(t *testing.T) {
		svc := &mockListasService{
			restaurarItemFn: func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error) {
				return nil, service.ErrListaRestauracaoNaoPermitida
			},
		}
		router, token := setupListasTestRouter(t, svc)
		req, _ := http.NewRequest("POST", "/api/v1/listas/1/itens/2/restaurar", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400, obteve %d", w.Code)
		}
	})

	t.Run("item nao encontrado retorna 404", func(t *testing.T) {
		svc := &mockListasService{
			restaurarItemFn: func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error) {
				return nil, service.ErrListaItemNaoEncontrado
			},
		}
		router, token := setupListasTestRouter(t, svc)
		req, _ := http.NewRequest("POST", "/api/v1/listas/1/itens/99/restaurar", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("esperava 404, obteve %d", w.Code)
		}
	})

	t.Run("sucesso 200", func(t *testing.T) {
		svc := &mockListasService{
			restaurarItemFn: func(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error) {
				return &service.ListaItemDetalhe{ID: itemID, Nome: "Item Restaurado", Ignorado: false}, nil
			},
		}
		router, token := setupListasTestRouter(t, svc)
		req, _ := http.NewRequest("POST", "/api/v1/listas/1/itens/2/restaurar", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", w.Code)
		}
	})
}

func TestListasHandler_ListarListas_SemDadosDuplicado(t *testing.T) {
	svc := &mockListasService{
		listarListasFn: func(ctx context.Context, usuarioID int32) ([]*service.ListaResumo, error) {
			return []*service.ListaResumo{{ID: 1, Nome: "Minha Lista"}}, nil
		},
	}
	router, token := setupListasTestRouter(t, svc)
	req, _ := http.NewRequest("GET", "/api/v1/listas", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", w.Code)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("erro ao decodificar json: %v", err)
	}
	if _, ok := raw["data"]; !ok {
		t.Fatalf("esperava chave data na resposta")
	}
	if _, ok := raw["dados"]; ok {
		t.Fatalf("chave dados nao deve existir na resposta")
	}
}

func TestListasHandler_CriarLista_FranquiaSemJogos_400(t *testing.T) {
	svc := &mockListasService{
		criarListaFn: func(ctx context.Context, input service.CriarListaInput) (*service.ListaDetalhada, error) {
			return nil, service.ErrListaFranquiaSemJogos
		},
	}
	router, token := setupListasTestRouter(t, svc)
	body := `{"tipo":"desafio","nome":"Zelda","regra":{"tipo":"franquia","igdb_id":596,"igdb_ids_ignorados":[1,2]}}`
	req, _ := http.NewRequest("POST", "/api/v1/listas", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", w.Code)
	}
}
