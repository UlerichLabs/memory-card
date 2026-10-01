// Package repository contem as implementacoes de persistencia.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type JogoEmAndamento struct {
	ID          int32      `json:"id"`
	UsuarioID   int32      `json:"-"`
	IgdbID      *int32     `json:"igdb_id"`
	IgdbCapaURL *string    `json:"igdb_capa_url"`
	Nome        string     `json:"nome"`
	IniciadoEm  time.Time  `json:"iniciado_em"`
	DeletedAt   *time.Time `json:"-"`
	CreatedAt   time.Time  `json:"-"`
	UpdatedAt   time.Time  `json:"-"`
}

type CriarJogoEmAndamentoParams struct {
	UsuarioID   int32
	IgdbID      *int32
	IgdbCapaURL *string
	Nome        string
	IniciadoEm  time.Time
}

type JogosEmAndamentoRepository interface {
	Criar(ctx context.Context, params CriarJogoEmAndamentoParams) (*JogoEmAndamento, error)
	Listar(ctx context.Context, usuarioID int32) ([]*JogoEmAndamento, error)
	Excluir(ctx context.Context, id int32, usuarioID int32) error
}

type SQLJogosEmAndamentoRepository struct {
	queries *db.Queries
}

func NewJogosEmAndamentoRepository(queries *db.Queries) *SQLJogosEmAndamentoRepository {
	return &SQLJogosEmAndamentoRepository{queries: queries}
}

func (r *SQLJogosEmAndamentoRepository) Criar(
	ctx context.Context,
	params CriarJogoEmAndamentoParams,
) (*JogoEmAndamento, error) {
	var igdbID pgtype.Int4
	if params.IgdbID != nil {
		igdbID = pgtype.Int4{Int32: *params.IgdbID, Valid: true}
	}
	var capaURL pgtype.Text
	if params.IgdbCapaURL != nil && *params.IgdbCapaURL != "" {
		capaURL = pgtype.Text{String: *params.IgdbCapaURL, Valid: true}
	}
	row, err := r.queries.CriarJogoEmAndamento(ctx, db.CriarJogoEmAndamentoParams{
		UsuarioID:   params.UsuarioID,
		IgdbID:      igdbID,
		IgdbCapaUrl: capaURL,
		Nome:        params.Nome,
		IniciadoEm:  pgtype.Timestamp{Time: params.IniciadoEm, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("criar jogo em andamento: %w", err)
	}
	return mapearJogoEmAndamento(row), nil
}

func (r *SQLJogosEmAndamentoRepository) Listar(ctx context.Context, usuarioID int32) ([]*JogoEmAndamento, error) {
	rows, err := r.queries.ListarJogosEmAndamento(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar jogos em andamento: %w", err)
	}
	result := make([]*JogoEmAndamento, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapearJogoEmAndamento(row))
	}
	return result, nil
}

func (r *SQLJogosEmAndamentoRepository) Excluir(ctx context.Context, id int32, usuarioID int32) error {
	rowsAffected, err := r.queries.ExcluirJogoEmAndamento(ctx, db.ExcluirJogoEmAndamentoParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		return fmt.Errorf("excluir jogo em andamento: %w", err)
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func mapearJogoEmAndamento(row db.JogosEmAndamento) *JogoEmAndamento {
	jogo := &JogoEmAndamento{
		ID:         row.ID,
		UsuarioID:  row.UsuarioID,
		Nome:       row.Nome,
		IniciadoEm: row.IniciadoEm.Time,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
	if row.IgdbID.Valid {
		jogo.IgdbID = &row.IgdbID.Int32
	}
	if row.IgdbCapaUrl.Valid && row.IgdbCapaUrl.String != "" {
		jogo.IgdbCapaURL = &row.IgdbCapaUrl.String
	}
	if row.DeletedAt.Valid {
		jogo.DeletedAt = &row.DeletedAt.Time
	}
	return jogo
}

var _ JogosEmAndamentoRepository = (*SQLJogosEmAndamentoRepository)(nil)
