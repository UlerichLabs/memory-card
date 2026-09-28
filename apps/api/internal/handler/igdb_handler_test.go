package handler

import (
	"context"
	"errors"
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
	detail    *igdbclient.Game
	detailErr error
}

func (m *igdbHandlerMock) BuscarJogos(context.Context, string) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, m.searchErr
}

func (m *igdbHandlerMock) BuscarJogo(context.Context, int64) (*igdbclient.Game, error) {
	return m.detail, m.detailErr
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
		{name: "too short", err: service.ErrTermoIGDBCurto, code: http.StatusBadRequest, body: `"codigo":"igdb.search.too_short"`},
		{name: "rate limit", err: service.ErrIGDBRateLimit, code: http.StatusTooManyRequests, body: `"codigo":"igdb.rate_limited"`},
		{name: "unavailable", err: service.ErrIGDBIndisponivel, code: http.StatusServiceUnavailable, body: `"codigo":"igdb.unavailable"`},
		{name: "query invalid", err: service.ErrIGDBQueryInvalida, code: http.StatusBadGateway, body: `"codigo":"igdb.query_invalid"`},
		{name: "canceled", err: context.Canceled, code: 499, body: ""},
		{name: "deadline exceeded", err: context.DeadlineExceeded, code: http.StatusServiceUnavailable, body: `"codigo":"igdb.unavailable"`},
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
			if errors.Is(test.err, context.Canceled) && recorder.Body.Len() != 0 {
				t.Fatalf("cancelamento deveria não ter corpo: %s", recorder.Body.String())
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
			Cover:            &igdbclient.Image{ID: 1, ImageID: "co1uid", URL: "//images.igdb.com/igdb/image/upload/t_thumb/co1uid.jpg"},
			FirstReleaseDate: &date,
			Summary:          "RPG",
			GameType:         igdbclient.GameTypeMainGame,
			TotalRatingCount: func() *int { value := 10; return &value }(),
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
	if !strings.Contains(body, `"name":"Chrono Trigger"`) || !strings.Contains(body, `"PlayStation"`) || !strings.Contains(body, `"Super Nintendo`) || !strings.Contains(body, `"cover":{"url":"//images.igdb.com`) {
		t.Fatalf("corpo inesperado: %s", body)
	}
	for _, field := range []string{"game_type", "version_parent", "total_rating_count", "parent_game", "image_id"} {
		if strings.Contains(body, field) {
			t.Fatalf("campo interno %q exposto: %s", field, body)
		}
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

func TestIGDBHandlerDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	date := int64(794880000)
	game := &igdbclient.Game{ID: 1802, Name: "Chrono Trigger", FirstReleaseDate: &date, Platforms: []igdbclient.Platform{{ID: 19, Name: "Super Nintendo"}}}
	router := gin.New()
	router.GET("/jogos/igdb/:id", NewIGDBHandler(&igdbHandlerMock{detail: game}).BuscarJogo)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/jogos/igdb/1802", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"Super Nintendo"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, field := range []string{"game_type", "version_parent", "total_rating_count", "parent_game", "image_id"} {
		if strings.Contains(recorder.Body.String(), field) {
			t.Fatalf("campo interno %q exposto: %s", field, recorder.Body.String())
		}
	}
}

func TestIGDBHandlerDetailErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		path string
		mock *igdbHandlerMock
		code int
	}{
		{name: "invalid id", path: "/jogos/igdb/nope", mock: &igdbHandlerMock{}, code: http.StatusBadRequest},
		{name: "not found", path: "/jogos/igdb/999", mock: &igdbHandlerMock{detailErr: service.ErrJogoIGDBNaoEncontrado}, code: http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/jogos/igdb/:id", NewIGDBHandler(tc.mock).BuscarJogo)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != tc.code {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestIGDBHandlerDetailSemToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	tokens, err := service.NewAuthToken(uuid.NewString(), 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/jogos/igdb/:id", NewIGDBHandler(&igdbHandlerMock{}).BuscarJogo)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/jogos/igdb/1802", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status esperado 401, obteve %d", recorder.Code)
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
