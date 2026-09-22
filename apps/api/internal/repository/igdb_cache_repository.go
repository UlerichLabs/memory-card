package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/jackc/pgx/v5"
)

type IGDBCacheRepository struct {
	queries *db.Queries
}

func NewIGDBCacheRepository(queries *db.Queries) *IGDBCacheRepository {
	return &IGDBCacheRepository{queries: queries}
}

func (repo *IGDBCacheRepository) Buscar(ctx context.Context, chave string, destino any) (bool, error) {
	payload, err := repo.queries.BuscarSnapshotIGDB(ctx, chave)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("buscar snapshot IGDB: %w", err)
	}
	if err := json.Unmarshal(payload, destino); err != nil {
		return false, fmt.Errorf("decodificar snapshot IGDB: %w", err)
	}
	return true, nil
}

func (repo *IGDBCacheRepository) Salvar(ctx context.Context, chave string, valor any) error {
	payload, err := json.Marshal(valor)
	if err != nil {
		return fmt.Errorf("serializar snapshot IGDB: %w", err)
	}
	if err := repo.queries.SalvarSnapshotIGDB(ctx, db.SalvarSnapshotIGDBParams{Chave: chave, Payload: payload}); err != nil {
		return fmt.Errorf("salvar snapshot IGDB: %w", err)
	}
	return nil
}
