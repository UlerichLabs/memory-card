package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
)

var ErrTermoIGDBVazio = errors.New("igdb.search.empty")

type IGDBClient interface {
	SearchGames(context.Context, string) ([]igdbclient.Game, error)
	GameDetails(context.Context, int64) (*igdbclient.Game, error)
	Platforms(context.Context) ([]igdbclient.Platform, error)
	GamesByPlatform(context.Context, int64) ([]igdbclient.Game, error)
	SearchFranchises(context.Context, string) ([]igdbclient.Franchise, error)
	GamesByFranchise(context.Context, int64) ([]igdbclient.Game, error)
}

type IGDBSnapshotRepository interface {
	Buscar(context.Context, string, any) (bool, error)
	Salvar(context.Context, string, any) error
}

type IGDBService struct {
	client IGDBClient
	cache  IGDBSnapshotRepository
}

func NewIGDBService(client IGDBClient, cache IGDBSnapshotRepository) *IGDBService {
	return &IGDBService{client: client, cache: cache}
}

func (svc *IGDBService) BuscarJogos(ctx context.Context, termo string) ([]igdbclient.Game, error) {
	if strings.TrimSpace(termo) == "" {
		return nil, ErrTermoIGDBVazio
	}
	return svc.client.SearchGames(ctx, termo)
}

func (svc *IGDBService) BuscarJogo(ctx context.Context, id int64) (*igdbclient.Game, error) {
	return svc.client.GameDetails(ctx, id)
}

func (svc *IGDBService) ListarPlataformas(ctx context.Context) ([]igdbclient.Platform, error) {
	return svc.client.Platforms(ctx)
}

func (svc *IGDBService) BuscarFranquias(ctx context.Context, termo string) ([]igdbclient.Franchise, error) {
	if strings.TrimSpace(termo) == "" {
		return nil, ErrTermoIGDBVazio
	}
	return svc.client.SearchFranchises(ctx, termo)
}

func (svc *IGDBService) JogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return svc.jogosSnapshot(ctx, fmt.Sprintf("platform:%d", id), func() ([]igdbclient.Game, error) {
		return svc.client.GamesByPlatform(ctx, id)
	})
}

func (svc *IGDBService) JogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return svc.jogosSnapshot(ctx, fmt.Sprintf("franchise:%d", id), func() ([]igdbclient.Game, error) {
		return svc.client.GamesByFranchise(ctx, id)
	})
}

func (svc *IGDBService) AtualizarJogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return svc.atualizarSnapshot(ctx, fmt.Sprintf("platform:%d", id), func() ([]igdbclient.Game, error) {
		return svc.client.GamesByPlatform(ctx, id)
	})
}

func (svc *IGDBService) AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return svc.atualizarSnapshot(ctx, fmt.Sprintf("franchise:%d", id), func() ([]igdbclient.Game, error) {
		return svc.client.GamesByFranchise(ctx, id)
	})
}

func (svc *IGDBService) jogosSnapshot(ctx context.Context, key string, buscar func() ([]igdbclient.Game, error)) ([]igdbclient.Game, error) {
	var games []igdbclient.Game
	found, err := svc.cache.Buscar(ctx, key, &games)
	if err != nil {
		return nil, err
	}
	if found {
		return games, nil
	}
	return svc.atualizarSnapshot(ctx, key, buscar)
}

func (svc *IGDBService) atualizarSnapshot(ctx context.Context, key string, buscar func() ([]igdbclient.Game, error)) ([]igdbclient.Game, error) {
	games, err := buscar()
	if err != nil {
		return nil, err
	}
	if games == nil {
		games = []igdbclient.Game{}
	}
	if err := svc.cache.Salvar(ctx, key, games); err != nil {
		return nil, err
	}
	return games, nil
}
