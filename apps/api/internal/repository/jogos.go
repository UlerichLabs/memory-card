package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type JogoZerado struct {
	ID            int32      `json:"id"`
	Numero        int        `json:"numero,omitempty"`
	UsuarioID     int32      `json:"usuario_id"`
	IgdbID        *int32     `json:"igdb_id,omitempty"`
	Nome          string     `json:"nome"`
	Console       string     `json:"console"`
	Genero        string     `json:"genero,omitempty"`
	Tipo          string     `json:"tipo,omitempty"`
	IniciadoEm    *time.Time `json:"iniciado_em,omitempty"`
	FinalizadoEm  time.Time  `json:"finalizado_em"`
	TempoJogado   int32      `json:"tempo_jogado"`
	Nota          int32      `json:"nota"`
	Dificuldade   string     `json:"dificuldade"`
	Review        string     `json:"review,omitempty"`
	Destaque      bool       `json:"destaque"`
	IgdbCapaURL   string     `json:"igdb_capa_url,omitempty"`
	IgdbDescricao string     `json:"igdb_descricao,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CriarJogoZeradoParams struct {
	UsuarioID     int32
	IgdbID        *int32
	Nome          string
	Console       string
	Genero        string
	Tipo          string
	IniciadoEm    *time.Time
	FinalizadoEm  time.Time
	TempoJogado   int32
	Nota          int32
	Dificuldade   string
	Review        string
	Destaque      bool
	IgdbCapaURL   string
	IgdbDescricao string
}

type AtualizarJogoZeradoParams struct {
	ID            int32
	UsuarioID     int32
	IgdbID        *int32
	Nome          string
	Console       string
	Genero        string
	Tipo          string
	IniciadoEm    *time.Time
	FinalizadoEm  time.Time
	TempoJogado   int32
	Nota          int32
	Dificuldade   string
	Review        string
	Destaque      bool
	IgdbCapaURL   string
	IgdbDescricao string
}

type ListarJogosZeradosParams struct {
	UsuarioID   int32
	Busca       string
	Console     string
	Genero      string
	Tipo        string
	NotaMin     *int32
	NotaMax     *int32
	Ano         *int32
	Dificuldade string
	Pagina      int
	PorPagina   int
}

type OpcoesFiltros struct {
	Consoles []string `json:"consoles"`
	Generos  []string `json:"generos"`
	Tipos    []string `json:"tipos"`
	Anos     []int    `json:"anos"`
}

type JogosRepository interface {
	Criar(ctx context.Context, params CriarJogoZeradoParams) (*JogoZerado, error)
	Atualizar(ctx context.Context, params AtualizarJogoZeradoParams) (*JogoZerado, error)
	Excluir(ctx context.Context, id int32, usuarioID int32) error
	Listar(ctx context.Context, params ListarJogosZeradosParams) ([]*JogoZerado, int64, error)
	ObterFiltros(ctx context.Context, usuarioID int32) (*OpcoesFiltros, error)
	ObterPorID(ctx context.Context, id int32, usuarioID int32) (*JogoZerado, error)
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

	var review pgtype.Text
	if params.Review != "" {
		review = pgtype.Text{String: params.Review, Valid: true}
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
		UsuarioID:     params.UsuarioID,
		IgdbID:        igdbID,
		Nome:          params.Nome,
		Console:       params.Console,
		Genero:        genero,
		Tipo:          tipo,
		IniciadoEm:    iniciadoEm,
		FinalizadoEm:  pgtype.Timestamp{Time: params.FinalizadoEm, Valid: true},
		TempoJogado:   params.TempoJogado,
		Nota:          params.Nota,
		Dificuldade:   db.Dificuldade(params.Dificuldade),
		Review:        review,
		Destaque:      pgtype.Bool{Bool: params.Destaque, Valid: true},
		IgdbCapaUrl:   capaURL,
		IgdbDescricao: descricao,
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

	var review pgtype.Text
	if params.Review != "" {
		review = pgtype.Text{String: params.Review, Valid: true}
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
		ID:            params.ID,
		UsuarioID:     params.UsuarioID,
		IgdbID:        igdbID,
		Nome:          params.Nome,
		Console:       params.Console,
		Genero:        genero,
		Tipo:          tipo,
		IniciadoEm:    iniciadoEm,
		FinalizadoEm:  pgtype.Timestamp{Time: params.FinalizadoEm, Valid: true},
		TempoJogado:   params.TempoJogado,
		Nota:          params.Nota,
		Dificuldade:   db.Dificuldade(params.Dificuldade),
		Review:        review,
		Destaque:      pgtype.Bool{Bool: params.Destaque, Valid: true},
		IgdbCapaUrl:   capaURL,
		IgdbDescricao: descricao,
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

func (r *SQLJogosRepository) Listar(ctx context.Context, params ListarJogosZeradosParams) ([]*JogoZerado, int64, error) {
	var busca pgtype.Text
	if params.Busca != "" {
		busca = pgtype.Text{String: params.Busca, Valid: true}
	}

	var console pgtype.Text
	if params.Console != "" {
		console = pgtype.Text{String: params.Console, Valid: true}
	}

	var genero pgtype.Text
	if params.Genero != "" {
		genero = pgtype.Text{String: params.Genero, Valid: true}
	}

	var tipo pgtype.Text
	if params.Tipo != "" {
		tipo = pgtype.Text{String: params.Tipo, Valid: true}
	}

	var notaMin pgtype.Int4
	if params.NotaMin != nil {
		notaMin = pgtype.Int4{Int32: *params.NotaMin, Valid: true}
	}

	var notaMax pgtype.Int4
	if params.NotaMax != nil {
		notaMax = pgtype.Int4{Int32: *params.NotaMax, Valid: true}
	}

	var ano pgtype.Int4
	if params.Ano != nil {
		ano = pgtype.Int4{Int32: *params.Ano, Valid: true}
	}

	var dificuldade pgtype.Text
	if params.Dificuldade != "" {
		dificuldade = pgtype.Text{String: params.Dificuldade, Valid: true}
	}

	total, err := r.queries.ContarJogosZerados(ctx, db.ContarJogosZeradosParams{
		UsuarioID:   params.UsuarioID,
		Busca:       busca,
		Console:     console,
		Genero:      genero,
		Tipo:        tipo,
		NotaMin:     notaMin,
		NotaMax:     notaMax,
		Ano:         ano,
		Dificuldade: dificuldade,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("contar jogos zerados: %w", err)
	}

	if total == 0 {
		return []*JogoZerado{}, 0, nil
	}

	offset := int32((params.Pagina - 1) * params.PorPagina)
	limite := int32(params.PorPagina)

	rows, err := r.queries.ListarJogosZerados(ctx, db.ListarJogosZeradosParams{
		UsuarioID:   params.UsuarioID,
		Busca:       busca,
		Console:     console,
		Genero:      genero,
		Tipo:        tipo,
		NotaMin:     notaMin,
		NotaMax:     notaMax,
		Ano:         ano,
		Dificuldade: dificuldade,
		OffsetVal:   offset,
		Limite:      limite,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("listar jogos zerados: %w", err)
	}

	jogos := make([]*JogoZerado, 0, len(rows))
	for _, row := range rows {
		jogo := mapearJogoZerado(row)
		jogo.IgdbDescricao = ""
		jogos = append(jogos, jogo)
	}

	return jogos, total, nil
}

func (r *SQLJogosRepository) ObterFiltros(ctx context.Context, usuarioID int32) (*OpcoesFiltros, error) {
	consoles, err := r.queries.ObterConsolesUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter consoles: %w", err)
	}
	if consoles == nil {
		consoles = []string{}
	}

	rawGeneros, err := r.queries.ObterGenerosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter generos: %w", err)
	}
	generosMap := make(map[string]struct{})
	for _, g := range rawGeneros {
		if g.Valid && g.String != "" {
			for _, part := range strings.Split(g.String, ",") {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					generosMap[trimmed] = struct{}{}
				}
			}
		}
	}
	generos := make([]string, 0, len(generosMap))
	for g := range generosMap {
		generos = append(generos, g)
	}
	sort.Strings(generos)

	rawTipos, err := r.queries.ObterTiposUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter tipos: %w", err)
	}
	tiposMap := make(map[string]struct{})
	for _, t := range rawTipos {
		if t.Valid && t.String != "" {
			trimmed := strings.TrimSpace(t.String)
			if trimmed != "" {
				tiposMap[trimmed] = struct{}{}
			}
		}
	}
	tipos := make([]string, 0, len(tiposMap))
	for t := range tiposMap {
		tipos = append(tipos, t)
	}
	sort.Strings(tipos)

	rawAnos, err := r.queries.ObterAnosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter anos: %w", err)
	}
	anos := make([]int, 0, len(rawAnos))
	for _, a := range rawAnos {
		anos = append(anos, int(a))
	}

	return &OpcoesFiltros{
		Consoles: consoles,
		Generos:  generos,
		Tipos:    tipos,
		Anos:     anos,
	}, nil
}

func (r *SQLJogosRepository) ObterPorID(ctx context.Context, id int32, usuarioID int32) (*JogoZerado, error) {
	row, err := r.queries.ObterDetalhesJogoZerado(ctx, db.ObterDetalhesJogoZeradoParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("obter detalhes jogo zerado: %w", err)
	}

	jogo := &JogoZerado{
		ID:          row.ID,
		Numero:      int(row.Numero),
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
	if row.Review.Valid {
		jogo.Review = row.Review.String
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
	return jogo, nil
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
	if row.Review.Valid {
		jogo.Review = row.Review.String
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
