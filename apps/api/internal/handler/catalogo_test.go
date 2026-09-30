package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

type catalogoHandlerServiceMock struct{}

func (catalogoHandlerServiceMock) Jogos(_ context.Context, _ int32, filtro service.CatalogoFiltro) (*service.CatalogoResultado, error) {
	if filtro.Origem == "xyz" {
		return nil, service.ErrCatalogoParametroInvalido
	}
	return &service.CatalogoResultado{Itens: []*service.CatalogoItem{}, Meta: service.CatalogoMeta{Pagina: 1, PorPagina: 60, Total: 0}}, nil
}
func (catalogoHandlerServiceMock) Generos(context.Context) ([]service.GeneroCatalogo, error) {
	return []service.GeneroCatalogo{{ID: 12, Nome: "RPG"}}, nil
}

func TestCatalogoHandler_ValidaAutenticacaoEParametros(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := strings.Repeat("a", 32)
	tokens, err := service.NewAuthToken(secret, time.Hour, 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	handler := NewCatalogoHandler(catalogoHandlerServiceMock{})
	privadas.GET("/catalogo/jogos", handler.Jogos)
	privadas.GET("/igdb/generos", handler.Generos)
	token := generateTestToken(t, secret, "42")

	request := httptest.NewRequest(http.MethodGet, "/api/v1/catalogo/jogos?origem=xyz&id=1", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obteve %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/catalogo/jogos?origem=xyz&id=1", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/igdb/generos", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() == "" {
		t.Fatalf("resposta de gêneros inesperada: %d %s", response.Code, response.Body.String())
	}
}
