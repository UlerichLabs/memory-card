package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type SQLTokenRevogadoRepository struct {
	queries *db.Queries
}

func NewTokenRevogadoRepository(queries *db.Queries) *SQLTokenRevogadoRepository {
	return &SQLTokenRevogadoRepository{queries: queries}
}

func (repo *SQLTokenRevogadoRepository) Revogar(ctx context.Context, jti string, expiraEm time.Time) error {
	if err := repo.queries.RevogarToken(ctx, db.RevogarTokenParams{Jti: jti, ExpiraEm: pgtype.Timestamptz{Time: expiraEm, Valid: true}}); err != nil {
		return fmt.Errorf("inserir revogacao: %w", err)
	}
	return nil
}

func (repo *SQLTokenRevogadoRepository) EstaRevogado(ctx context.Context, jti string) (bool, error) {
	if err := repo.queries.LimparTokensRevogadosExpirados(ctx); err != nil {
		return false, fmt.Errorf("limpar revogacoes expiradas: %w", err)
	}
	revogado, err := repo.queries.TokenEstaRevogado(ctx, jti)
	if err != nil {
		return false, fmt.Errorf("consultar revogacao: %w", err)
	}
	return revogado, nil
}
