//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/handler"
	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/UlerichLabs/memory-card/apps/api/internal/testutil"
)

type snapshotIGDBMock struct {
	calls atomic.Int32
}

func (m *snapshotIGDBMock) SearchGames(context.Context, string) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func (m *snapshotIGDBMock) GameDetails(context.Context, int64) (*igdbclient.Game, error) {
	return nil, nil
}

func (m *snapshotIGDBMock) Platforms(context.Context) ([]igdbclient.Platform, error) {
	return []igdbclient.Platform{}, nil
}

func (m *snapshotIGDBMock) GamesByPlatform(context.Context, int64) ([]igdbclient.Game, error) {
	m.calls.Add(1)
	return []igdbclient.Game{{ID: 123, Name: "Snapshot Game"}}, nil
}

func (m *snapshotIGDBMock) AtualizarJogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return m.GamesByPlatform(ctx, id)
}

func (m *snapshotIGDBMock) SearchFranchises(context.Context, string) ([]igdbclient.Franchise, error) {
	return []igdbclient.Franchise{}, nil
}

func (m *snapshotIGDBMock) GamesByFranchise(context.Context, int64) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func (m *snapshotIGDBMock) AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return m.GamesByFranchise(ctx, id)
}

func TestE2E_IGDBSnapshotReadsDatabaseAfterCreation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pg := testutil.SetupPostgres(t)
	mock := &snapshotIGDBMock{}
	svc := service.NewIGDBService(mock, repository.NewIGDBCacheRepository(db.New(pg.Pool)))
	handler := handler.NewIGDBHandler(svc)
	router := gin.New()
	router.GET("/api/v1/igdb/plataformas/:id/jogos", handler.JogosDaPlataforma)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	for range 2 {
		response, err := http.Get(server.URL + "/api/v1/igdb/plataformas/4/jogos")
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("status=%d", response.StatusCode)
		}
	}
	if mock.calls.Load() != 1 {
		t.Fatalf("IGDB calls=%d", mock.calls.Load())
	}
	var count int
	if err := pg.Pool.QueryRow(context.Background(), "SELECT count(*) FROM igdb_cache WHERE chave = $1", "platform:4").Scan(&count); err != nil || count != 1 {
		t.Fatalf("snapshot count=%d err=%v", count, err)
	}
}
