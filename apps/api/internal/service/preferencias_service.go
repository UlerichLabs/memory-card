// Package service contem as regras de negocio da aplicacao.
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrPreferenciasIdiomaInvalido           = errors.New("preferencias.idioma_invalido")
	ErrPreferenciasFormatoDataInvalido      = errors.New("preferencias.formato_data_invalido")
	ErrPreferenciasFusoHorarioInvalido      = errors.New("preferencias.fuso_horario_invalido")
	ErrPreferenciasVisualBibliotecaInvalido = errors.New("preferencias.visual_biblioteca_invalido")
	ErrPreferenciasItensPorPaginaInvalido   = errors.New("preferencias.itens_por_pagina_invalido")
	ErrPreferenciasModoTemaInvalido         = errors.New("preferencias.modo_tema_invalido")
	ErrPreferenciasEstiloTemaInvalido       = errors.New("preferencias.estilo_tema_invalido")
	ErrPreferenciasEntradaInvalida          = errors.New("preferencias.entrada_invalida")
)

const (
	DefaultFormatoData      = "dmy"
	DefaultFusoHorario      = "America/Sao_Paulo"
	DefaultVisualBiblioteca = "grade"
	DefaultItensPorPagina   = 24
	DefaultReduzirAnimacoes = false
	DefaultModoTema         = "escuro"
	DefaultEstiloTema       = "padrao"
)

var (
	IdiomasPermitidos = map[string]struct{}{
		"pt-BR": {},
		"en":    {},
	}
	FormatosDataPermitidos = map[string]struct{}{
		"dmy": {},
		"mdy": {},
		"ymd": {},
	}
	VisuaisBibliotecaPermitidos = map[string]struct{}{
		"grade": {},
		"lista": {},
	}
	ItensPorPaginaPermitidos = map[int]struct{}{
		24:  {},
		50:  {},
		100: {},
	}
	ModosTemaPermitidos = map[string]struct{}{
		"claro":      {},
		"escuro":     {},
		"automatico": {},
	}
	EstilosTemaPermitidos = map[string]struct{}{
		"padrao":      {},
		"playstation": {},
		"nintendo":    {},
		"xbox":        {},
		"steam":       {},
	}
)

type AtualizarPreferenciasInput struct {
	Idioma           *string `json:"idioma"`
	FormatoData      *string `json:"formato_data"`
	FusoHorario      *string `json:"fuso_horario"`
	VisualBiblioteca *string `json:"visual_biblioteca"`
	ItensPorPagina   *int    `json:"itens_por_pagina"`
	ReduzirAnimacoes *bool   `json:"reduzir_animacoes"`
	ModoTema         *string `json:"modo_tema"`
	EstiloTema       *string `json:"estilo_tema"`
}

type PreferenciasRepository interface {
	BuscarPorUsuarioID(ctx context.Context, usuarioID int32) (*repository.PreferenciasUsuario, error)
	AtualizarPreferencias(ctx context.Context, params repository.AtualizarPreferenciasParams) (*repository.PreferenciasUsuario, error)
}

type PreferenciasService struct {
	repo PreferenciasRepository
}

func NewPreferenciasService(repo PreferenciasRepository) *PreferenciasService {
	return &PreferenciasService{repo: repo}
}

func (svc *PreferenciasService) ObterPreferencias(ctx context.Context, subject string) (*repository.PreferenciasUsuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("obter preferencias cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}
	pref, err := svc.repo.BuscarPorUsuarioID(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("consultar preferencias: %w", err)
	}
	if pref == nil {
		return nil, ErrTokenInvalido
	}
	return pref, nil
}

func (svc *PreferenciasService) AtualizarPreferencias(
	ctx context.Context,
	subject string,
	input AtualizarPreferenciasInput,
) (*repository.PreferenciasUsuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("atualizar preferencias cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}

	if input.Idioma == nil ||
		input.FormatoData == nil ||
		input.FusoHorario == nil ||
		input.VisualBiblioteca == nil ||
		input.ItensPorPagina == nil ||
		input.ReduzirAnimacoes == nil ||
		input.ModoTema == nil ||
		input.EstiloTema == nil {
		return nil, ErrPreferenciasEntradaInvalida
	}

	if _, ok := IdiomasPermitidos[*input.Idioma]; !ok {
		return nil, ErrPreferenciasIdiomaInvalido
	}

	if _, ok := FormatosDataPermitidos[*input.FormatoData]; !ok {
		return nil, ErrPreferenciasFormatoDataInvalido
	}

	fuso := *input.FusoHorario
	if strings.TrimSpace(fuso) == "" || strings.EqualFold(fuso, "Local") {
		return nil, ErrPreferenciasFusoHorarioInvalido
	}
	if _, err := time.LoadLocation(fuso); err != nil {
		return nil, ErrPreferenciasFusoHorarioInvalido
	}

	if _, ok := VisuaisBibliotecaPermitidos[*input.VisualBiblioteca]; !ok {
		return nil, ErrPreferenciasVisualBibliotecaInvalido
	}

	if _, ok := ItensPorPaginaPermitidos[*input.ItensPorPagina]; !ok {
		return nil, ErrPreferenciasItensPorPaginaInvalido
	}

	if _, ok := ModosTemaPermitidos[*input.ModoTema]; !ok {
		return nil, ErrPreferenciasModoTemaInvalido
	}

	if _, ok := EstilosTemaPermitidos[*input.EstiloTema]; !ok {
		return nil, ErrPreferenciasEstiloTemaInvalido
	}

	pref, err := svc.repo.AtualizarPreferencias(ctx, repository.AtualizarPreferenciasParams{
		UsuarioID:        int32(id),
		Idioma:           *input.Idioma,
		FormatoData:      *input.FormatoData,
		FusoHorario:      fuso,
		VisualBiblioteca: *input.VisualBiblioteca,
		ItensPorPagina:   *input.ItensPorPagina,
		ReduzirAnimacoes: *input.ReduzirAnimacoes,
		ModoTema:         *input.ModoTema,
		EstiloTema:       *input.EstiloTema,
	})
	if err != nil {
		return nil, fmt.Errorf("atualizar preferencias no repositorio: %w", err)
	}
	if pref == nil {
		return nil, ErrTokenInvalido
	}

	return pref, nil
}
