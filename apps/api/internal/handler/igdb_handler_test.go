package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type igdbHandlerMock struct {
	searchErr error
}

func (m *igdbHandlerMock) BuscarJogos(context.Context, string) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, m.searchErr
}

func (m *igdbHandlerMock) BuscarJogo(context.Context, int64) (*igdbclient.Game, error) {
	return nil, nil
}

func (m *igdbHandlerMock) ListarPlataformas(context.Context) ([]igdbclient.Platform, error) {
	return []igdbclient.Platform{}, nil
}

func (m *igdbHandlerMock) JogosDaPlataforma(context.Context, int64) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func (m *igdbHandlerMock) AtualizarJogosDaPlataforma(context.Context, int64) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func (m *igdbHandlerMock) BuscarFranquias(context.Context, string) ([]igdbclient.Franchise, error) {
	return []igdbclient.Franchise{}, nil
}

func (m *igdbHandlerMock) JogosDaFranquia(context.Context, int64) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func (m *igdbHandlerMock) AtualizarJogosDaFranquia(context.Context, int64) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func TestIGDBHandlerSearchErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		err  error
		code int
		body string
	}{
		{name: "empty", err: service.ErrTermoIGDBVazio, code: http.StatusBadRequest, body: `"codigo":"igdb.search.empty"`},
		{name: "rate limit", err: igdbclient.ErrRateLimited, code: http.StatusTooManyRequests, body: `"codigo":"igdb.rate_limited"`},
		{name: "authentication", err: igdbclient.ErrAuthentication, code: http.StatusBadGateway, body: `"codigo":"igdb.unavailable"`},
		{name: "unavailable", err: igdbclient.ErrUnavailable, code: http.StatusBadGateway, body: `"codigo":"igdb.unavailable"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/igdb/jogos/busca", NewIGDBHandler(&igdbHandlerMock{searchErr: test.err}).BuscarJogos)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/igdb/jogos/busca?q=game", nil))
			if recorder.Code != test.code || !strings.Contains(recorder.Body.String(), test.body) {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestIGDBHandlerSearch_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	date := int64(794880000)
	expectedGames := []igdbclient.Game{
		{
			ID:               1802,
			Name:             "Chrono Trigger",
			FirstReleaseDate: &date,
			Platforms: []igdbclient.Platform{
				{ID: 19, Name: "Super Nintendo Entertainment System"},
				{ID: 7, Name: "PlayStation"},
			},
		},
	}

	mock := &igdbHandlerMockSuccess{games: expectedGames}
	router := gin.New()
	router.GET("/igdb/jogos/busca", NewIGDBHandler(mock).BuscarJogos)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/igdb/jogos/busca?q=chrono", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status esperado 200, obteve %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"name":"Chrono Trigger"`) || !strings.Contains(body, `"PlayStation"`) || !strings.Contains(body, `"Super Nintendo`) {
		t.Fatalf("corpo inesperado: %s", body)
	}
}

func TestIGDBHandlerSearch_TermoAusenteOuVazio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		url  string
	}{
		{"sem query param", "/igdb/jogos/busca"},
		{"query param vazio", "/igdb/jogos/busca?q="},
		{"query param com espacos", "/igdb/jogos/busca?q=%20%20"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/igdb/jogos/busca", NewIGDBHandler(&igdbHandlerMock{searchErr: service.ErrTermoIGDBVazio}).BuscarJogos)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.url, nil))

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status esperado 400, obteve %d", recorder.Code)
			}
			if !strings.Contains(recorder.Body.String(), `"codigo":"igdb.search.empty"`) {
				t.Fatalf("esperava código igdb.search.empty, obteve: %s", recorder.Body.String())
			}
		})
	}
}

func TestIGDBHandlerSearch_SemToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	tokens, err := service.NewAuthToken(uuid.NewString(), 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/igdb/jogos/busca", NewIGDBHandler(&igdbHandlerMock{}).BuscarJogos)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/igdb/jogos/busca?q=chrono", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status esperado 401, obteve %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"codigo":"auth.session.unauthorized"`) {
		t.Fatalf("esperava auth.session.unauthorized, obteve: %s", recorder.Body.String())
	}
}

type igdbHandlerMockSuccess struct {
	igdbHandlerMock
	games []igdbclient.Game
}

func (m *igdbHandlerMockSuccess) BuscarJogos(context.Context, string) ([]igdbclient.Game, error) {
	return m.games, nil
}
