package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type TokenResetSenha struct {
	ID        int32
	UsuarioID int32
	ExpiraEm  time.Time
	UsadoEm   *time.Time
}

type RefreshTokenAtivo struct {
	JTI       string
	UsuarioID int32
	ExpiraEm  time.Time
}

type SQLRecuperacaoSenhaRepository struct {
	queries *db.Queries
}

func NewRecuperacaoSenhaRepository(queries *db.Queries) *SQLRecuperacaoSenhaRepository {
	return &SQLRecuperacaoSenhaRepository{queries: queries}
}

func (repo *SQLRecuperacaoSenhaRepository) RegistrarSolicitacao(ctx context.Context, email string) (bool, error) {
	_, err := repo.queries.RegistrarSolicitacaoResetSenha(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("registrar solicitacao de reset: %w", err)
	}
	return true, nil
}

func (repo *SQLRecuperacaoSenhaRepository) CriarToken(ctx context.Context, usuarioID int32, tokenHash string, expiraEm time.Time) error {
	err := repo.queries.CriarTokenResetSenha(ctx, db.CriarTokenResetSenhaParams{
		UsuarioID: usuarioID,
		TokenHash: tokenHash,
		ExpiraEm:  pgtype.Timestamptz{Time: expiraEm, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("criar token de reset: %w", err)
	}
	return nil
}

func (repo *SQLRecuperacaoSenhaRepository) BuscarToken(ctx context.Context, tokenHash string) (*TokenResetSenha, error) {
	row, err := repo.queries.BuscarTokenResetSenha(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscar token de reset: %w", err)
	}
	var usadoEm *time.Time
	if row.UsadoEm.Valid {
		value := row.UsadoEm.Time
		usadoEm = &value
	}
	return &TokenResetSenha{ID: row.ID, UsuarioID: row.UsuarioID, ExpiraEm: row.ExpiraEm.Time, UsadoEm: usadoEm}, nil
}

func (repo *SQLRecuperacaoSenhaRepository) ConsumirETrocarSenha(ctx context.Context, tokenID int32, senhaHash string) (bool, error) {
	_, err := repo.queries.ConsumirTokenResetEAtualizarSenha(ctx, db.ConsumirTokenResetEAtualizarSenhaParams{ID: tokenID, SenhaHash: senhaHash})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("consumir token e atualizar senha: %w", err)
	}
	return true, nil
}

func (repo *SQLRecuperacaoSenhaRepository) RegistrarRefreshToken(ctx context.Context, jti string, usuarioID int32, expiraEm time.Time) error {
	err := repo.queries.RegistrarRefreshTokenAtivo(ctx, db.RegistrarRefreshTokenAtivoParams{
		Jti:       jti,
		UsuarioID: usuarioID,
		ExpiraEm:  pgtype.Timestamptz{Time: expiraEm, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("registrar refresh token ativo: %w", err)
	}
	return nil
}

func (repo *SQLRecuperacaoSenhaRepository) ListarRefreshTokensAtivos(ctx context.Context, usuarioID int32) ([]RefreshTokenAtivo, error) {
	rows, err := repo.queries.ListarRefreshTokensAtivosPorUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar refresh tokens ativos: %w", err)
	}
	result := make([]RefreshTokenAtivo, 0, len(rows))
	for _, row := range rows {
		result = append(result, RefreshTokenAtivo{JTI: row.Jti, UsuarioID: row.UsuarioID, ExpiraEm: row.ExpiraEm.Time})
	}
	return result, nil
}

func (repo *SQLRecuperacaoSenhaRepository) RemoverRefreshTokensAtivos(ctx context.Context, usuarioID int32) error {
	if err := repo.queries.RemoverRefreshTokensAtivosPorUsuario(ctx, usuarioID); err != nil {
		return fmt.Errorf("remover refresh tokens ativos: %w", err)
	}
	return nil
}
