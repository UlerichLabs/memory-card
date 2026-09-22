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

type JogoZerado struct {
	ID                int32      `json:"id"`
	UsuarioID         int32      `json:"usuario_id"`
	IgdbID            *int32     `json:"igdb_id,omitempty"`
	Nome              string     `json:"nome"`
	Console           string     `json:"console"`
	Genero            string     `json:"genero,omitempty"`
	Tipo              string     `json:"tipo,omitempty"`
	IniciadoEm        *time.Time `json:"iniciado_em,omitempty"`
	FinalizadoEm      time.Time  `json:"finalizado_em"`
	TempoJogado       int32      `json:"tempo_jogado"`
	Nota              int32      `json:"nota"`
	Dificuldade       string     `json:"dificuldade"`
	CondicaoZeramento string     `json:"condicao_zeramento,omitempty"`
	Destaque          bool       `json:"destaque"`
	IgdbCapaURL       string     `json:"igdb_capa_url,omitempty"`
	IgdbDescricao     string     `json:"igdb_descricao,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CriarJogoZeradoParams struct {
	UsuarioID         int32
	IgdbID            *int32
	Nome              string
	Console           string
	Genero            string
	Tipo              string
	IniciadoEm        *time.Time
	FinalizadoEm      time.Time
	TempoJogado       int32
	Nota              int32
	Dificuldade       string
	CondicaoZeramento string
	Destaque          bool
	IgdbCapaURL       string
	IgdbDescricao     string
}

type AtualizarJogoZeradoParams struct {
	ID                int32
	UsuarioID         int32
	IgdbID            *int32
	Nome              string
	Console           string
	Genero            string
	Tipo              string
	IniciadoEm        *time.Time
	FinalizadoEm      time.Time
	TempoJogado       int32
	Nota              int32
	Dificuldade       string
	CondicaoZeramento string
	Destaque          bool
	IgdbCapaURL       string
	IgdbDescricao     string
}

type JogosRepository interface {
	Criar(ctx context.Context, params CriarJogoZeradoParams) (*JogoZerado, error)
	Atualizar(ctx context.Context, params AtualizarJogoZeradoParams) (*JogoZerado, error)
	Excluir(ctx context.Context, id int32, usuarioID int32) error
}

type SQLJogosRepository struct {
	queries *db.Queries
}

func NewJogosRepository(queries *db.Queries) *SQLJogosRepository {
	return &SQLJogosRepository{queries: queries}
}

func (r *SQLJogosRepository) Criar(ctx context.Context, params CriarJogoZeradoParams) (*JogoZerado, error) {
	var igdbID pgtype.Int4
	if params.IgdbID != nil {
		igdbID = pgtype.Int4{Int32: *params.IgdbID, Valid: true}
	}

	var genero pgtype.Text
	if params.Genero != "" {
		genero = pgtype.Text{String: params.Genero, Valid: true}
	}

	var tipo pgtype.Text
	if params.Tipo != "" {
		tipo = pgtype.Text{String: params.Tipo, Valid: true}
	}

	var iniciadoEm pgtype.Timestamp
	if params.IniciadoEm != nil {
		iniciadoEm = pgtype.Timestamp{Time: *params.IniciadoEm, Valid: true}
	}

	var condicao pgtype.Text
	if params.CondicaoZeramento != "" {
		condicao = pgtype.Text{String: params.CondicaoZeramento, Valid: true}
	}

	var capaURL pgtype.Text
	if params.IgdbCapaURL != "" {
		capaURL = pgtype.Text{String: params.IgdbCapaURL, Valid: true}
	}

	var descricao pgtype.Text
	if params.IgdbDescricao != "" {
		descricao = pgtype.Text{String: params.IgdbDescricao, Valid: true}
	}

	row, err := r.queries.CriarJogoZerado(ctx, db.CriarJogoZeradoParams{
		UsuarioID:         params.UsuarioID,
		IgdbID:            igdbID,
		Nome:              params.Nome,
		Console:           params.Console,
		Genero:            genero,
		Tipo:              tipo,
		IniciadoEm:        iniciadoEm,
		FinalizadoEm:      pgtype.Timestamp{Time: params.FinalizadoEm, Valid: true},
		TempoJogado:       params.TempoJogado,
		Nota:              params.Nota,
		Dificuldade:       db.Dificuldade(params.Dificuldade),
		CondicaoZeramento: condicao,
		Destaque:          pgtype.Bool{Bool: params.Destaque, Valid: true},
		IgdbCapaUrl:       capaURL,
		IgdbDescricao:     descricao,
	})
	if err != nil {
		return nil, fmt.Errorf("criar jogo zerado: %w", err)
	}

	return mapearJogoZerado(row), nil
}

func (r *SQLJogosRepository) Atualizar(ctx context.Context, params AtualizarJogoZeradoParams) (*JogoZerado, error) {
	var igdbID pgtype.Int4
	if params.IgdbID != nil {
		igdbID = pgtype.Int4{Int32: *params.IgdbID, Valid: true}
	}

	var genero pgtype.Text
	if params.Genero != "" {
		genero = pgtype.Text{String: params.Genero, Valid: true}
	}

	var tipo pgtype.Text
	if params.Tipo != "" {
		tipo = pgtype.Text{String: params.Tipo, Valid: true}
	}

	var iniciadoEm pgtype.Timestamp
	if params.IniciadoEm != nil {
		iniciadoEm = pgtype.Timestamp{Time: *params.IniciadoEm, Valid: true}
	}

	var condicao pgtype.Text
	if params.CondicaoZeramento != "" {
		condicao = pgtype.Text{String: params.CondicaoZeramento, Valid: true}
	}

	var capaURL pgtype.Text
	if params.IgdbCapaURL != "" {
		capaURL = pgtype.Text{String: params.IgdbCapaURL, Valid: true}
	}

	var descricao pgtype.Text
	if params.IgdbDescricao != "" {
		descricao = pgtype.Text{String: params.IgdbDescricao, Valid: true}
	}

	row, err := r.queries.AtualizarJogoZerado(ctx, db.AtualizarJogoZeradoParams{
		ID:                params.ID,
		UsuarioID:         params.UsuarioID,
		IgdbID:            igdbID,
		Nome:              params.Nome,
		Console:           params.Console,
		Genero:            genero,
		Tipo:              tipo,
		IniciadoEm:        iniciadoEm,
		FinalizadoEm:      pgtype.Timestamp{Time: params.FinalizadoEm, Valid: true},
		TempoJogado:       params.TempoJogado,
		Nota:              params.Nota,
		Dificuldade:       db.Dificuldade(params.Dificuldade),
		CondicaoZeramento: condicao,
		Destaque:          pgtype.Bool{Bool: params.Destaque, Valid: true},
		IgdbCapaUrl:       capaURL,
		IgdbDescricao:     descricao,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("atualizar jogo zerado: %w", err)
	}

	return mapearJogoZerado(row), nil
}

func (r *SQLJogosRepository) Excluir(ctx context.Context, id int32, usuarioID int32) error {
	rowsAffected, err := r.queries.ExcluirJogoZerado(ctx, db.ExcluirJogoZeradoParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		return fmt.Errorf("excluir jogo zerado: %w", err)
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}


func mapearJogoZerado(row db.JogosZerado) *JogoZerado {
	jogo := &JogoZerado{
		ID:          row.ID,
		UsuarioID:   row.UsuarioID,
		Nome:        row.Nome,
		Console:     row.Console,
		TempoJogado: row.TempoJogado,
		Nota:        row.Nota,
		Dificuldade: string(row.Dificuldade),
	}
	if row.IgdbID.Valid {
		jogo.IgdbID = &row.IgdbID.Int32
	}
	if row.Genero.Valid {
		jogo.Genero = row.Genero.String
	}
	if row.Tipo.Valid {
		jogo.Tipo = row.Tipo.String
	}
	if row.IniciadoEm.Valid {
		jogo.IniciadoEm = &row.IniciadoEm.Time
	}
	if row.FinalizadoEm.Valid {
		jogo.FinalizadoEm = row.FinalizadoEm.Time
	}
	if row.CondicaoZeramento.Valid {
		jogo.CondicaoZeramento = row.CondicaoZeramento.String
	}
	if row.Destaque.Valid {
		jogo.Destaque = row.Destaque.Bool
	}
	if row.IgdbCapaUrl.Valid {
		jogo.IgdbCapaURL = row.IgdbCapaUrl.String
	}
	if row.IgdbDescricao.Valid {
		jogo.IgdbDescricao = row.IgdbDescricao.String
	}
	if row.CreatedAt.Valid {
		jogo.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		jogo.UpdatedAt = row.UpdatedAt.Time
	}
	return jogo
}
