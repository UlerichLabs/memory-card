package service

import (
	"context"
	"testing"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
)

type igdbMock struct {
	platformGamesCalls  int
	franchiseGamesCalls int
}

func (m *igdbMock) SearchGames(context.Context, string) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func (m *igdbMock) GameDetails(context.Context, int64) (*igdbclient.Game, error) {
	return nil, nil
}

func (m *igdbMock) Platforms(context.Context) ([]igdbclient.Platform, error) {
	return []igdbclient.Platform{}, nil
}

func (m *igdbMock) GamesByPlatform(context.Context, int64) ([]igdbclient.Game, error) {
	m.platformGamesCalls++
	return []igdbclient.Game{{ID: 7, Name: "Cached"}}, nil
}

func (m *igdbMock) AtualizarJogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return m.GamesByPlatform(ctx, id)
}

func (m *igdbMock) SearchFranchises(context.Context, string) ([]igdbclient.Franchise, error) {
	return []igdbclient.Franchise{}, nil
}

func (m *igdbMock) GamesByFranchise(context.Context, int64) ([]igdbclient.Game, error) {
	m.franchiseGamesCalls++
	return []igdbclient.Game{{ID: 8, Name: "Franchise"}}, nil
}

func (m *igdbMock) AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return m.GamesByFranchise(ctx, id)
}

type cacheMock struct {
	items map[string]any
}

func (m *cacheMock) Buscar(_ context.Context, key string, destino any) (bool, error) {
	item, found := m.items[key]
	if found {
		games := destino.(*[]igdbclient.Game)
		*games = item.([]igdbclient.Game)
	}
	return found, nil
}

func (m *cacheMock) Salvar(_ context.Context, key string, valor any) error {
	m.items[key] = valor
	return nil
}

func TestIGDBServiceSnapshotsPlatformCatalog(t *testing.T) {
	client := &igdbMock{}
	cache := &cacheMock{items: make(map[string]any)}
	service := NewIGDBService(client, cache)
	for range 2 {
		games, err := service.JogosDaPlataforma(context.Background(), 4)
		if err != nil || len(games) != 1 || games[0].Name != "Cached" {
			t.Fatalf("games=%v err=%v", games, err)
		}
	}
	if client.platformGamesCalls != 1 {
		t.Fatalf("IGDB calls=%d", client.platformGamesCalls)
	}
}

func TestIGDBServiceManualRefreshReplacesSnapshot(t *testing.T) {
	client := &igdbMock{}
	cache := &cacheMock{items: make(map[string]any)}
	service := NewIGDBService(client, cache)
	if _, err := service.JogosDaPlataforma(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AtualizarJogosDaPlataforma(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	if client.platformGamesCalls != 2 {
		t.Fatalf("IGDB calls=%d", client.platformGamesCalls)
	}
}

func TestIGDBServiceRejectsEmptySearch(t *testing.T) {
	service := NewIGDBService(&igdbMock{}, &cacheMock{items: make(map[string]any)})
	if _, err := service.BuscarJogos(context.Background(), "  "); err != ErrTermoIGDBVazio {
		t.Fatalf("error=%v", err)
	}
	if _, err := service.BuscarFranquias(context.Background(), ""); err != ErrTermoIGDBVazio {
		t.Fatalf("error=%v", err)
	}
}
