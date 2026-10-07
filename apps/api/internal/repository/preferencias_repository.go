// Package repository contem as implementacoes de repositorio.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type PreferenciasUsuario struct {
	Idioma           string `json:"idioma"`
	FormatoData      string `json:"formato_data"`
	FusoHorario      string `json:"fuso_horario"`
	VisualBiblioteca string `json:"visual_biblioteca"`
	ItensPorPagina   int    `json:"itens_por_pagina"`
	ReduzirAnimacoes bool   `json:"reduzir_animacoes"`
	ModoTema         string `json:"modo_tema"`
	EstiloTema       string `json:"estilo_tema"`
}

type AtualizarPreferenciasParams struct {
	UsuarioID        int32
	Idioma           string
	FormatoData      string
	FusoHorario      string
	VisualBiblioteca string
	ItensPorPagina   int
	ReduzirAnimacoes bool
	ModoTema         string
	EstiloTema       string
}

type PreferenciasRepository interface {
	BuscarPorUsuarioID(ctx context.Context, usuarioID int32) (*PreferenciasUsuario, error)
	AtualizarPreferencias(ctx context.Context, params AtualizarPreferenciasParams) (*PreferenciasUsuario, error)
}

type SQLPreferenciasRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewPreferenciasRepository(queries *db.Queries, pool ...*pgxpool.Pool) *SQLPreferenciasRepository {
	var p *pgxpool.Pool
	if len(pool) > 0 {
		p = pool[0]
	}
	return &SQLPreferenciasRepository{queries: queries, pool: p}
}

func (r *SQLPreferenciasRepository) BuscarPorUsuarioID(ctx context.Context, usuarioID int32) (*PreferenciasUsuario, error) {
	row, err := r.queries.BuscarPreferenciasPorUsuarioID(ctx, usuarioID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscar preferencias por usuario id: %w", err)
	}

	formatoData := "dmy"
	if row.FormatoData.Valid && row.FormatoData.String != "" {
		formatoData = row.FormatoData.String
	}

	fusoHorario := "America/Sao_Paulo"
	if row.FusoHorario.Valid && row.FusoHorario.String != "" {
		fusoHorario = row.FusoHorario.String
	}

	visualBiblioteca := "grade"
	if row.VisualBiblioteca.Valid && row.VisualBiblioteca.String != "" {
		visualBiblioteca = row.VisualBiblioteca.String
	}

	itensPorPagina := 24
	if row.ItensPorPagina.Valid && row.ItensPorPagina.Int32 > 0 {
		itensPorPagina = int(row.ItensPorPagina.Int32)
	}

	reduzirAnimacoes := false
	if row.ReduzirAnimacoes.Valid {
		reduzirAnimacoes = row.ReduzirAnimacoes.Bool
	}

	modoTema := "escuro"
	if row.ModoTema.Valid && row.ModoTema.String != "" {
		modoTema = row.ModoTema.String
	}

	estiloTema := "padrao"
	if row.EstiloTema.Valid && row.EstiloTema.String != "" {
		estiloTema = row.EstiloTema.String
	}

	return &PreferenciasUsuario{
		Idioma:           row.Idioma,
		FormatoData:      formatoData,
		FusoHorario:      fusoHorario,
		VisualBiblioteca: visualBiblioteca,
		ItensPorPagina:   itensPorPagina,
		ReduzirAnimacoes: reduzirAnimacoes,
		ModoTema:         modoTema,
		EstiloTema:       estiloTema,
	}, nil
}

func (r *SQLPreferenciasRepository) AtualizarPreferencias(ctx context.Context, params AtualizarPreferenciasParams) (*PreferenciasUsuario, error) {
	var (
		qtx = r.queries
		tx  pgx.Tx
		err error
	)

	if r.pool != nil {
		tx, err = r.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("iniciar transacao: %w", err)
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback(ctx)
			}
		}()
		qtx = r.queries.WithTx(tx)
	}

	rows, err := qtx.AtualizarIdiomaUsuario(ctx, db.AtualizarIdiomaUsuarioParams{
		ID:     params.UsuarioID,
		Idioma: params.Idioma,
	})
	if err != nil {
		return nil, fmt.Errorf("atualizar idioma usuario: %w", err)
	}
	if rows == 0 {
		return nil, nil
	}

	prefRow, err := qtx.UpsertPreferenciasUsuario(ctx, db.UpsertPreferenciasUsuarioParams{
		UsuarioID:        params.UsuarioID,
		FormatoData:      params.FormatoData,
		FusoHorario:      params.FusoHorario,
		VisualBiblioteca: params.VisualBiblioteca,
		ItensPorPagina:   int32(params.ItensPorPagina),
		ReduzirAnimacoes: params.ReduzirAnimacoes,
		ModoTema:         params.ModoTema,
		EstiloTema:       params.EstiloTema,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert preferencias usuario: %w", err)
	}

	if tx != nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit transacao: %w", err)
		}
		tx = nil
	}

	return &PreferenciasUsuario{
		Idioma:           params.Idioma,
		FormatoData:      prefRow.FormatoData,
		FusoHorario:      prefRow.FusoHorario,
		VisualBiblioteca: prefRow.VisualBiblioteca,
		ItensPorPagina:   int(prefRow.ItensPorPagina),
		ReduzirAnimacoes: prefRow.ReduzirAnimacoes,
		ModoTema:         prefRow.ModoTema,
		EstiloTema:       prefRow.EstiloTema,
	}, nil
}
