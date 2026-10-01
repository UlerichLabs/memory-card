// Package repository contem as implementacoes de persistencia.
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

type JogoAbandonado struct {
	ID           int32      `json:"id"`
	UsuarioID    int32      `json:"-"`
	IgdbID       *int32     `json:"igdb_id,omitempty"`
	IgdbCapaURL  *string    `json:"igdb_capa_url,omitempty"`
	Nome         string     `json:"nome"`
	Console      string     `json:"console"`
	TempoJogado  int32      `json:"tempo_jogado"`
	Motivo       *string    `json:"motivo,omitempty"`
	AbandonadoEm time.Time  `json:"abandonado_em"`
	IniciadoEm   *time.Time `json:"iniciado_em"`
	DeletedAt    *time.Time `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CriarJogoAbandonadoParams struct {
	UsuarioID    int32
	IgdbID       *int32
	IgdbCapaURL  *string
	Nome         string
	Console      string
	TempoJogado  int32
	Motivo       *string
	AbandonadoEm time.Time
	IniciadoEm   *time.Time
}

type AtualizarJogoAbandonadoParams struct {
	ID           int32
	UsuarioID    int32
	IgdbID       *int32
	IgdbCapaURL  *string
	Nome         string
	Console      string
	TempoJogado  int32
	Motivo       *string
	AbandonadoEm time.Time
	IniciadoEm   *time.Time
}

type ListarJogosAbandonadosParams struct {
	UsuarioID int32
	Busca     string
	Console   string
	Ordenar   string
	Pagina    int
	PorPagina int
}

type JogosAbandonadosRepository interface {
	Criar(ctx context.Context, params CriarJogoAbandonadoParams) (*JogoAbandonado, error)
	Atualizar(ctx context.Context, params AtualizarJogoAbandonadoParams) (*JogoAbandonado, error)
	Excluir(ctx context.Context, id int32, usuarioID int32) error
	Listar(ctx context.Context, params ListarJogosAbandonadosParams) ([]*JogoAbandonado, int64, error)
	ObterPorID(ctx context.Context, id int32, usuarioID int32) (*JogoAbandonado, error)
	ObterConsoles(ctx context.Context, usuarioID int32) ([]string, error)
	ObterTotal(ctx context.Context, usuarioID int32) (int64, error)
}

type SQLJogosAbandonadosRepository struct {
	queries *db.Queries
}

func NewJogosAbandonadosRepository(queries *db.Queries) *SQLJogosAbandonadosRepository {
	return &SQLJogosAbandonadosRepository{queries: queries}
}

func (r *SQLJogosAbandonadosRepository) Criar(
	ctx context.Context,
	params CriarJogoAbandonadoParams,
) (*JogoAbandonado, error) {
	var igdbID pgtype.Int4
	if params.IgdbID != nil {
		igdbID = pgtype.Int4{Int32: *params.IgdbID, Valid: true}
	}

	var capaURL pgtype.Text
	if params.IgdbCapaURL != nil && *params.IgdbCapaURL != "" {
		capaURL = pgtype.Text{String: *params.IgdbCapaURL, Valid: true}
	}

	var motivo pgtype.Text
	if params.Motivo != nil && *params.Motivo != "" {
		motivo = pgtype.Text{String: *params.Motivo, Valid: true}
	}

	var iniciadoEm pgtype.Timestamp
	if params.IniciadoEm != nil {
		iniciadoEm = pgtype.Timestamp{Time: *params.IniciadoEm, Valid: true}
	}

	row, err := r.queries.CriarJogoAbandonado(ctx, db.CriarJogoAbandonadoParams{
		UsuarioID:    params.UsuarioID,
		IgdbID:       igdbID,
		IgdbCapaUrl:  capaURL,
		Nome:         params.Nome,
		Console:      params.Console,
		TempoJogado:  params.TempoJogado,
		Motivo:       motivo,
		AbandonadoEm: pgtype.Timestamp{Time: params.AbandonadoEm, Valid: true},
		IniciadoEm:   iniciadoEm,
	})
	if err != nil {
		return nil, fmt.Errorf("criar jogo abandonado: %w", err)
	}

	return mapearJogoAbandonado(row), nil
}

func (r *SQLJogosAbandonadosRepository) Atualizar(
	ctx context.Context,
	params AtualizarJogoAbandonadoParams,
) (*JogoAbandonado, error) {
	var igdbID pgtype.Int4
	if params.IgdbID != nil {
		igdbID = pgtype.Int4{Int32: *params.IgdbID, Valid: true}
	}

	var capaURL pgtype.Text
	if params.IgdbCapaURL != nil && *params.IgdbCapaURL != "" {
		capaURL = pgtype.Text{String: *params.IgdbCapaURL, Valid: true}
	}

	var motivo pgtype.Text
	if params.Motivo != nil && *params.Motivo != "" {
		motivo = pgtype.Text{String: *params.Motivo, Valid: true}
	}

	var iniciadoEm pgtype.Timestamp
	if params.IniciadoEm != nil {
		iniciadoEm = pgtype.Timestamp{Time: *params.IniciadoEm, Valid: true}
	}

	row, err := r.queries.AtualizarJogoAbandonado(ctx, db.AtualizarJogoAbandonadoParams{
		ID:           params.ID,
		UsuarioID:    params.UsuarioID,
		IgdbID:       igdbID,
		IgdbCapaUrl:  capaURL,
		Nome:         params.Nome,
		Console:      params.Console,
		TempoJogado:  params.TempoJogado,
		Motivo:       motivo,
		AbandonadoEm: pgtype.Timestamp{Time: params.AbandonadoEm, Valid: true},
		IniciadoEm:   iniciadoEm,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("atualizar jogo abandonado: %w", err)
	}

	return mapearJogoAbandonado(row), nil
}

func (r *SQLJogosAbandonadosRepository) Excluir(ctx context.Context, id int32, usuarioID int32) error {
	rowsAffected, err := r.queries.ExcluirJogoAbandonado(ctx, db.ExcluirJogoAbandonadoParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		return fmt.Errorf("excluir jogo abandonado: %w", err)
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *SQLJogosAbandonadosRepository) Listar(
	ctx context.Context,
	params ListarJogosAbandonadosParams,
) ([]*JogoAbandonado, int64, error) {
	var busca pgtype.Text
	if params.Busca != "" {
		busca = pgtype.Text{String: params.Busca, Valid: true}
	}

	var console pgtype.Text
	if params.Console != "" {
		console = pgtype.Text{String: params.Console, Valid: true}
	}

	total, err := r.queries.ContarJogosAbandonados(ctx, db.ContarJogosAbandonadosParams{
		UsuarioID: params.UsuarioID,
		Busca:     busca,
		Console:   console,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("contar jogos abandonados: %w", err)
	}

	if total == 0 {
		return []*JogoAbandonado{}, 0, nil
	}

	offset := int32((params.Pagina - 1) * params.PorPagina)
	limite := int32(params.PorPagina)

	var rows []db.JogosAbandonado
	switch params.Ordenar {
	case "antigos":
		rows, err = r.queries.ListarJogosAbandonadosAntigos(ctx, db.ListarJogosAbandonadosAntigosParams{
			UsuarioID: params.UsuarioID,
			Busca:     busca,
			Console:   console,
			OffsetVal: offset,
			Limite:    limite,
		})
	case "nome":
		rows, err = r.queries.ListarJogosAbandonadosNome(ctx, db.ListarJogosAbandonadosNomeParams{
			UsuarioID: params.UsuarioID,
			Busca:     busca,
			Console:   console,
			OffsetVal: offset,
			Limite:    limite,
		})
	case "tempo":
		rows, err = r.queries.ListarJogosAbandonadosTempo(ctx, db.ListarJogosAbandonadosTempoParams{
			UsuarioID: params.UsuarioID,
			Busca:     busca,
			Console:   console,
			OffsetVal: offset,
			Limite:    limite,
		})
	default:
		rows, err = r.queries.ListarJogosAbandonadosRecentes(ctx, db.ListarJogosAbandonadosRecentesParams{
			UsuarioID: params.UsuarioID,
			Busca:     busca,
			Console:   console,
			OffsetVal: offset,
			Limite:    limite,
		})
	}
	if err != nil {
		return nil, 0, fmt.Errorf("listar jogos abandonados: %w", err)
	}

	jogos := make([]*JogoAbandonado, 0, len(rows))
	for _, row := range rows {
		jogos = append(jogos, mapearJogoAbandonado(row))
	}

	return jogos, total, nil
}

func (r *SQLJogosAbandonadosRepository) ObterPorID(
	ctx context.Context,
	id int32,
	usuarioID int32,
) (*JogoAbandonado, error) {
	row, err := r.queries.BuscarJogoAbandonadoPorID(ctx, db.BuscarJogoAbandonadoPorIDParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("buscar jogo abandonado por id: %w", err)
	}

	return mapearJogoAbandonado(row), nil
}

func (r *SQLJogosAbandonadosRepository) ObterConsoles(ctx context.Context, usuarioID int32) ([]string, error) {
	consoles, err := r.queries.ObterConsolesAbandonadosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter consoles abandonados: %w", err)
	}
	if consoles == nil {
		return []string{}, nil
	}
	return consoles, nil
}

func (r *SQLJogosAbandonadosRepository) ObterTotal(ctx context.Context, usuarioID int32) (int64, error) {
	total, err := r.queries.ObterTotalAbandonadosUsuario(ctx, usuarioID)
	if err != nil {
		return 0, fmt.Errorf("obter total abandonados: %w", err)
	}
	return total, nil
}

func mapearJogoAbandonado(row db.JogosAbandonado) *JogoAbandonado {
	jogo := &JogoAbandonado{
		ID:           row.ID,
		UsuarioID:    row.UsuarioID,
		Nome:         row.Nome,
		Console:      row.Console,
		TempoJogado:  row.TempoJogado,
		AbandonadoEm: row.AbandonadoEm.Time,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
	if row.IgdbID.Valid {
		jogo.IgdbID = &row.IgdbID.Int32
	}
	if row.IgdbCapaUrl.Valid && row.IgdbCapaUrl.String != "" {
		jogo.IgdbCapaURL = &row.IgdbCapaUrl.String
	}
	if row.Motivo.Valid && row.Motivo.String != "" {
		jogo.Motivo = &row.Motivo.String
	}
	if row.DeletedAt.Valid {
		jogo.DeletedAt = &row.DeletedAt.Time
	}
	if row.IniciadoEm.Valid {
		jogo.IniciadoEm = &row.IniciadoEm.Time
	}
	return jogo
}
