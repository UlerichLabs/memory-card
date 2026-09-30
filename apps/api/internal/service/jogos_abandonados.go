// Package service contem as regras de negocio da aplicacao.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrAbandonadoNomeObrigatorio    = errors.New("abandonados.nome_obrigatorio")
	ErrAbandonadoNomeMuitoLongo     = errors.New("abandonados.nome_muito_longo")
	ErrAbandonadoConsoleObrigatorio = errors.New("abandonados.console_obrigatorio")
	ErrAbandonadoConsoleMuitoLongo  = errors.New("abandonados.console_muito_longo")
	ErrAbandonadoTempoInvalido      = errors.New("abandonados.tempo_invalido")
	ErrAbandonadoMotivoMuitoLongo   = errors.New("abandonados.motivo_muito_longo")
	ErrAbandonadoDataInvalida       = errors.New("abandonados.data_invalida")
	ErrAbandonadoDataFutura         = errors.New("abandonados.data_futura")
	ErrAbandonadoPaginaInvalida     = errors.New("abandonados.pagina_invalida")
	ErrAbandonadoPorPaginaInvalida  = errors.New("abandonados.por_pagina_invalida")
	ErrAbandonadoOrdenarInvalido    = errors.New("abandonados.ordenar_invalido")
	ErrAbandonadoNaoEncontrado      = errors.New("abandonados.nao_encontrado")
	ErrAbandonadoIdInvalido         = errors.New("abandonados.id_invalido")
	ErrAbandonadoEntradaInvalida    = errors.New("abandonados.entrada_invalida")
)

type SalvarJogoAbandonadoInput struct {
	ID                  int32
	UsuarioID           int32
	Nome                string
	Console             string
	IgdbID              *int32
	IgdbCapaURL         *string
	TempoJogadoHoras    int
	TempoJogadoMinutos  int
	TempoJogadoSegundos int
	TempoJogado         *int32
	Motivo              *string
	AbandonadoEm        *time.Time
}

type ListarJogosAbandonadosInput struct {
	UsuarioID int32
	Busca     string
	Console   string
	Ordenar   string
	Pagina    *int
	PorPagina *int
}

type ResultadoListagemAbandonados struct {
	Jogos        []*repository.JogoAbandonado
	Total        int64
	TotalPaginas int
	Pagina       int
	PorPagina    int
}

type JogosAbandonadosRepository interface {
	Criar(ctx context.Context, params repository.CriarJogoAbandonadoParams) (*repository.JogoAbandonado, error)
	Atualizar(ctx context.Context, params repository.AtualizarJogoAbandonadoParams) (*repository.JogoAbandonado, error)
	Excluir(ctx context.Context, id int32, usuarioID int32) error
	Listar(ctx context.Context, params repository.ListarJogosAbandonadosParams) ([]*repository.JogoAbandonado, int64, error)
	ObterPorID(ctx context.Context, id int32, usuarioID int32) (*repository.JogoAbandonado, error)
	ObterConsoles(ctx context.Context, usuarioID int32) ([]string, error)
	ObterTotal(ctx context.Context, usuarioID int32) (int64, error)
}

type JogosAbandonadosService struct {
	repo JogosAbandonadosRepository
	now  func() time.Time
}

func NewJogosAbandonadosService(repo JogosAbandonadosRepository) *JogosAbandonadosService {
	return &JogosAbandonadosService{repo: repo, now: time.Now}
}

func (s *JogosAbandonadosService) SetNow(nowFn func() time.Time) {
	s.now = nowFn
}

func (s *JogosAbandonadosService) validarESintetizar(
	input SalvarJogoAbandonadoInput,
) (string, string, int32, *string, *string, time.Time, error) {
	nome := strings.TrimSpace(input.Nome)
	if nome == "" {
		return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoNomeObrigatorio
	}
	if len([]rune(nome)) > 200 {
		return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoNomeMuitoLongo
	}

	console := strings.TrimSpace(input.Console)
	if console == "" {
		return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoConsoleObrigatorio
	}
	if len([]rune(console)) > 100 {
		return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoConsoleMuitoLongo
	}

	if input.TempoJogadoHoras < 0 ||
		input.TempoJogadoMinutos < 0 || input.TempoJogadoMinutos > 59 ||
		input.TempoJogadoSegundos < 0 || input.TempoJogadoSegundos > 59 {
		return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoTempoInvalido
	}

	if input.TempoJogado != nil && *input.TempoJogado < 0 {
		return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoTempoInvalido
	}

	var tempoTotal int32
	if input.TempoJogadoHoras > 0 || input.TempoJogadoMinutos > 0 || input.TempoJogadoSegundos > 0 {
		tempoTotal = int32(input.TempoJogadoHoras*3600 + input.TempoJogadoMinutos*60 + input.TempoJogadoSegundos)
	} else if input.TempoJogado != nil {
		tempoTotal = *input.TempoJogado
	}

	if tempoTotal < 0 {
		return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoTempoInvalido
	}

	var motivoFinal *string
	if input.Motivo != nil {
		trimmedMotivo := strings.TrimSpace(*input.Motivo)
		if len([]rune(trimmedMotivo)) > 500 {
			return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoMotivoMuitoLongo
		}
		if trimmedMotivo != "" {
			motivoFinal = &trimmedMotivo
		}
	}

	var capaFinal *string
	if input.IgdbCapaURL != nil {
		trimmedCapa := strings.TrimSpace(*input.IgdbCapaURL)
		if trimmedCapa != "" {
			capaFinal = &trimmedCapa
		}
	}

	agora := s.now()
	var abandonadoEm time.Time
	if input.AbandonadoEm == nil {
		abandonadoEm = agora
	} else {
		if input.AbandonadoEm.IsZero() {
			return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoDataInvalida
		}
		abandonadoEm = *input.AbandonadoEm
	}

	if abandonadoEm.After(agora) {
		return "", "", 0, nil, nil, time.Time{}, ErrAbandonadoDataFutura
	}

	return nome, console, tempoTotal, motivoFinal, capaFinal, abandonadoEm, nil
}

func (s *JogosAbandonadosService) CriarJogoAbandonado(
	ctx context.Context,
	input SalvarJogoAbandonadoInput,
) (*repository.JogoAbandonado, error) {
	nome, console, tempoTotal, motivoFinal, capaFinal, abandonadoEm, err := s.validarESintetizar(input)
	if err != nil {
		return nil, err
	}

	jogo, err := s.repo.Criar(ctx, repository.CriarJogoAbandonadoParams{
		UsuarioID:    input.UsuarioID,
		IgdbID:       input.IgdbID,
		IgdbCapaURL:  capaFinal,
		Nome:         nome,
		Console:      console,
		TempoJogado:  tempoTotal,
		Motivo:       motivoFinal,
		AbandonadoEm: abandonadoEm,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22001" {
				return nil, erroCampoMuitoLongoAbandonados(pgErr.ColumnName)
			}
			if pgErr.Code == pgerrcode.CheckViolation || pgErr.Code == "23514" {
				return nil, erroCheckConstraintAbandonados(pgErr.ConstraintName)
			}
		}
		return nil, fmt.Errorf("salvar jogo abandonado: %w", err)
	}

	return jogo, nil
}

func (s *JogosAbandonadosService) AtualizarJogoAbandonado(
	ctx context.Context,
	input SalvarJogoAbandonadoInput,
) (*repository.JogoAbandonado, error) {
	nome, console, tempoTotal, motivoFinal, capaFinal, abandonadoEm, err := s.validarESintetizar(input)
	if err != nil {
		return nil, err
	}

	jogo, err := s.repo.Atualizar(ctx, repository.AtualizarJogoAbandonadoParams{
		ID:           input.ID,
		UsuarioID:    input.UsuarioID,
		IgdbID:       input.IgdbID,
		IgdbCapaURL:  capaFinal,
		Nome:         nome,
		Console:      console,
		TempoJogado:  tempoTotal,
		Motivo:       motivoFinal,
		AbandonadoEm: abandonadoEm,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAbandonadoNaoEncontrado
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22001" {
				return nil, erroCampoMuitoLongoAbandonados(pgErr.ColumnName)
			}
			if pgErr.Code == pgerrcode.CheckViolation || pgErr.Code == "23514" {
				return nil, erroCheckConstraintAbandonados(pgErr.ConstraintName)
			}
		}
		return nil, fmt.Errorf("atualizar jogo abandonado: %w", err)
	}

	return jogo, nil
}

func erroCampoMuitoLongoAbandonados(campo string) error {
	switch campo {
	case "nome":
		return ErrAbandonadoNomeMuitoLongo
	case "console":
		return ErrAbandonadoConsoleMuitoLongo
	case "motivo":
		return ErrAbandonadoMotivoMuitoLongo
	default:
		return ErrAbandonadoMotivoMuitoLongo
	}
}

func erroCheckConstraintAbandonados(constraint string) error {
	switch {
	case strings.Contains(constraint, "tempo_jogado"):
		return ErrAbandonadoTempoInvalido
	case strings.Contains(constraint, "motivo"):
		return ErrAbandonadoMotivoMuitoLongo
	default:
		return ErrAbandonadoTempoInvalido
	}
}

func (s *JogosAbandonadosService) ExcluirJogoAbandonado(
	ctx context.Context,
	id int32,
	usuarioID int32,
) error {
	err := s.repo.Excluir(ctx, id, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAbandonadoNaoEncontrado
		}
		return fmt.Errorf("excluir jogo abandonado: %w", err)
	}
	return nil
}

func (s *JogosAbandonadosService) ObterJogoAbandonado(
	ctx context.Context,
	id int32,
	usuarioID int32,
) (*repository.JogoAbandonado, error) {
	jogo, err := s.repo.ObterPorID(ctx, id, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAbandonadoNaoEncontrado
		}
		return nil, fmt.Errorf("obter jogo abandonado: %w", err)
	}
	return jogo, nil
}

func (s *JogosAbandonadosService) ListarJogosAbandonados(
	ctx context.Context,
	params ListarJogosAbandonadosInput,
) (*ResultadoListagemAbandonados, error) {
	pagina := 1
	if params.Pagina != nil {
		if *params.Pagina < 1 {
			return nil, ErrAbandonadoPaginaInvalida
		}
		pagina = *params.Pagina
	}

	porPagina := 24
	if params.PorPagina != nil {
		if *params.PorPagina < 1 || *params.PorPagina > 100 {
			return nil, ErrAbandonadoPorPaginaInvalida
		}
		porPagina = *params.PorPagina
	}

	ordenar := strings.TrimSpace(params.Ordenar)
	if ordenar == "" {
		ordenar = "recentes"
	}
	if ordenar != "recentes" && ordenar != "antigos" && ordenar != "nome" && ordenar != "tempo" {
		return nil, ErrAbandonadoOrdenarInvalido
	}

	termoBusca := strings.TrimSpace(params.Busca)
	if termoBusca != "" {
		termoBusca = escaparLike(termoBusca)
	}

	termoConsole := strings.TrimSpace(params.Console)

	jogos, total, err := s.repo.Listar(ctx, repository.ListarJogosAbandonadosParams{
		UsuarioID: params.UsuarioID,
		Busca:     termoBusca,
		Console:   termoConsole,
		Ordenar:   ordenar,
		Pagina:    pagina,
		PorPagina: porPagina,
	})
	if err != nil {
		return nil, fmt.Errorf("listar jogos abandonados: %w", err)
	}

	var totalPaginas int
	if total > 0 {
		totalPaginas = int((total + int64(porPagina) - 1) / int64(porPagina))
	}

	return &ResultadoListagemAbandonados{
		Jogos:        jogos,
		Total:        total,
		TotalPaginas: totalPaginas,
		Pagina:       pagina,
		PorPagina:    porPagina,
	}, nil
}

func (s *JogosAbandonadosService) ObterConsoles(ctx context.Context, usuarioID int32) ([]string, error) {
	consoles, err := s.repo.ObterConsoles(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter consoles abandonados: %w", err)
	}
	if consoles == nil {
		return []string{}, nil
	}
	return consoles, nil
}

func (s *JogosAbandonadosService) ObterTotal(ctx context.Context, usuarioID int32) (int64, error) {
	total, err := s.repo.ObterTotal(ctx, usuarioID)
	if err != nil {
		return 0, fmt.Errorf("obter total abandonados: %w", err)
	}
	return total, nil
}
