package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
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
