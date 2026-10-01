// Package repository contem as implementacoes de persistencia.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type ResumoDados struct {
	TotalJogos          int64
	TotalSegundos       int64
	SomaNotas           int64
	TotalAvaliados      int64
	JogosNoAnoAtual     int64
	PrimeiroZeramentoEm *time.Time
}

type EstatisticaAnoDados struct {
	Ano                 int
	TotalJogos          int64
	TotalSegundos       int64
	DestaqueID          *int32
	DestaqueNome        string
	DestaqueConsole     string
	DestaqueIgdbCapaURL string
	DestaqueNota        *int32
}

type RankingItemDados struct {
	Nome          string
	TotalJogos    int64
	TotalSegundos int64
}

type BreakdownTipoDados struct {
	Tipo       string
	TotalJogos int64
}

type RecordeJogoDados struct {
	ID                  int32
	Nome                string
	Console             string
	Ano                 int
	IgdbCapaURL         string
	TempoJogadoSegundos int32
}

type NotaDistribuicaoDados struct {
	Nota  int32
	Total int64
}

type DificuldadeDistribuicaoDados struct {
	Dificuldade string
	TotalJogos  int64
}

type DashboardRepository interface {
	ObterResumo(ctx context.Context, usuarioID int32, anoAtual int32) (*ResumoDados, error)
	ListarEstatisticasPorAno(ctx context.Context, usuarioID int32) ([]*EstatisticaAnoDados, error)
	ObterRankingPlataformas(ctx context.Context, usuarioID int32) ([]*RankingItemDados, error)
	ObterRankingGeneros(ctx context.Context, usuarioID int32) ([]*RankingItemDados, error)
	ObterBreakdownTipo(ctx context.Context, usuarioID int32, genero string) ([]*BreakdownTipoDados, error)
	ObterRecordeMaisLongo(ctx context.Context, usuarioID int32) (*RecordeJogoDados, error)
	ObterRecordeMaisCurto(ctx context.Context, usuarioID int32) (*RecordeJogoDados, error)
	ObterDistribuicaoNotas(ctx context.Context, usuarioID int32) ([]*NotaDistribuicaoDados, error)
	ObterDistribuicaoDificuldade(ctx context.Context, usuarioID int32) ([]*DificuldadeDistribuicaoDados, error)
}

type SQLDashboardRepository struct {
	queries *db.Queries
}

func NewDashboardRepository(queries *db.Queries) *SQLDashboardRepository {
	return &SQLDashboardRepository{queries: queries}
}

func (r *SQLDashboardRepository) ObterResumo(ctx context.Context, usuarioID int32, anoAtual int32) (*ResumoDados, error) {
	row, err := r.queries.ObterResumoDashboard(ctx, db.ObterResumoDashboardParams{
		AnoAtual:  anoAtual,
		UsuarioID: usuarioID,
	})
	if err != nil {
		return nil, fmt.Errorf("obter resumo dashboard: %w", err)
	}

	resumo := &ResumoDados{
		TotalJogos:      row.TotalJogos,
		TotalSegundos:   row.TotalSegundos,
		SomaNotas:       row.SomaNotas,
		TotalAvaliados:  row.TotalAvaliados,
		JogosNoAnoAtual: row.JogosNoAnoAtual,
	}

	if row.PrimeiroZeramentoEm.Valid {
		t := row.PrimeiroZeramentoEm.Time.UTC()
		resumo.PrimeiroZeramentoEm = &t
	}

	return resumo, nil
}

func (r *SQLDashboardRepository) ListarEstatisticasPorAno(ctx context.Context, usuarioID int32) ([]*EstatisticaAnoDados, error) {
	rows, err := r.queries.ListarEstatisticasPorAno(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar estatisticas por ano: %w", err)
	}

	items := make([]*EstatisticaAnoDados, 0, len(rows))
	for _, row := range rows {
		item := &EstatisticaAnoDados{
			Ano:           int(row.Ano),
			TotalJogos:    row.TotalJogos,
			TotalSegundos: row.TotalSegundos,
		}
		if row.DestaqueID.Valid {
			id := row.DestaqueID.Int32
			item.DestaqueID = &id
			item.DestaqueNome = row.DestaqueNome.String
			item.DestaqueConsole = row.DestaqueConsole.String
			item.DestaqueIgdbCapaURL = row.DestaqueIgdbCapaUrl.String
			nota := row.DestaqueNota.Int32
			item.DestaqueNota = &nota
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *SQLDashboardRepository) ObterRankingPlataformas(ctx context.Context, usuarioID int32) ([]*RankingItemDados, error) {
	rows, err := r.queries.ObterRankingPlataformas(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter ranking plataformas: %w", err)
	}

	items := make([]*RankingItemDados, 0, len(rows))
	for _, row := range rows {
		items = append(items, &RankingItemDados{
			Nome:          row.Console,
			TotalJogos:    row.TotalJogos,
			TotalSegundos: row.TotalSegundos,
		})
	}

	return items, nil
}

func (r *SQLDashboardRepository) ObterRankingGeneros(ctx context.Context, usuarioID int32) ([]*RankingItemDados, error) {
	rows, err := r.queries.ObterRankingGeneros(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter ranking generos: %w", err)
	}

	items := make([]*RankingItemDados, 0, len(rows))
	for _, row := range rows {
		items = append(items, &RankingItemDados{
			Nome:          row.Genero.String,
			TotalJogos:    row.TotalJogos,
			TotalSegundos: row.TotalSegundos,
		})
	}

	return items, nil
}

func (r *SQLDashboardRepository) ObterBreakdownTipo(ctx context.Context, usuarioID int32, genero string) ([]*BreakdownTipoDados, error) {
	rows, err := r.queries.ObterBreakdownTipo(ctx, db.ObterBreakdownTipoParams{
		UsuarioID: usuarioID,
		Genero:    genero,
	})
	if err != nil {
		return nil, fmt.Errorf("obter breakdown tipo: %w", err)
	}

	items := make([]*BreakdownTipoDados, 0, len(rows))
	for _, row := range rows {
		items = append(items, &BreakdownTipoDados{
			Tipo:       row.Tipo.String,
			TotalJogos: row.TotalJogos,
		})
	}

	return items, nil
}

func (r *SQLDashboardRepository) ObterRecordeMaisLongo(ctx context.Context, usuarioID int32) (*RecordeJogoDados, error) {
	row, err := r.queries.ObterRecordeMaisLongo(ctx, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("obter recorde mais longo: %w", err)
	}

	return &RecordeJogoDados{
		ID:                  row.ID,
		Nome:                row.Nome,
		Console:             row.Console,
		Ano:                 int(row.Ano),
		IgdbCapaURL:         row.IgdbCapaUrl,
		TempoJogadoSegundos: row.TempoJogadoSegundos,
	}, nil
}

func (r *SQLDashboardRepository) ObterRecordeMaisCurto(ctx context.Context, usuarioID int32) (*RecordeJogoDados, error) {
	row, err := r.queries.ObterRecordeMaisCurto(ctx, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("obter recorde mais curto: %w", err)
	}

	return &RecordeJogoDados{
		ID:                  row.ID,
		Nome:                row.Nome,
		Console:             row.Console,
		Ano:                 int(row.Ano),
		IgdbCapaURL:         row.IgdbCapaUrl,
		TempoJogadoSegundos: row.TempoJogadoSegundos,
	}, nil
}

func (r *SQLDashboardRepository) ObterDistribuicaoNotas(ctx context.Context, usuarioID int32) ([]*NotaDistribuicaoDados, error) {
	rows, err := r.queries.ObterDistribuicaoNotas(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter distribuicao notas: %w", err)
	}

	items := make([]*NotaDistribuicaoDados, 0, len(rows))
	for _, row := range rows {
		items = append(items, &NotaDistribuicaoDados{
			Nota:  row.Nota,
			Total: row.Total,
		})
	}

	return items, nil
}

func (r *SQLDashboardRepository) ObterDistribuicaoDificuldade(ctx context.Context, usuarioID int32) ([]*DificuldadeDistribuicaoDados, error) {
	rows, err := r.queries.ObterDistribuicaoDificuldade(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter distribuicao dificuldade: %w", err)
	}

	items := make([]*DificuldadeDistribuicaoDados, 0, len(rows))
	for _, row := range rows {
		items = append(items, &DificuldadeDistribuicaoDados{
			Dificuldade: string(row.Dificuldade),
			TotalJogos:  row.TotalJogos,
		})
	}

	return items, nil
}
