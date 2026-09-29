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
	ErrNotaInvalida            = errors.New("jogos.nota_invalida")
	ErrDificuldadeInvalida     = errors.New("jogos.dificuldade_invalida")
	ErrReviewMuitoLongo        = errors.New("jogos.review_muito_longo")
	ErrNomeObrigatorio         = errors.New("jogos.nome_obrigatorio")
	ErrConsoleObrigatorio      = errors.New("jogos.console_obrigatorio")
	ErrFinalizadoEmObrigatorio = errors.New("jogos.finalizado_em_obrigatorio")
	ErrTempoJogadoInvalido     = errors.New("jogos.tempo_jogado_invalido")
	ErrDestaqueAnoConflito     = errors.New("jogos.destaque_ano_conflito")
	ErrJogoNaoEncontrado       = errors.New("jogos.nao_encontrado")
	ErrNomeMuitoLongo          = errors.New("jogos.nome_muito_longo")
	ErrConsoleMuitoLongo       = errors.New("jogos.console_muito_longo")
	ErrGeneroMuitoLongo        = errors.New("jogos.genero_muito_longo")
	ErrTipoMuitoLongo          = errors.New("jogos.tipo_muito_longo")
	ErrPaginaInvalida          = errors.New("jogos.pagina_invalida")
	ErrPorPaginaInvalido       = errors.New("jogos.por_pagina_invalido")
	ErrNotaFiltroInvalida      = errors.New("jogos.nota_filtro_invalida")
	ErrNotaFaixaInvalida       = errors.New("jogos.nota_faixa_invalida")
	ErrAnoInvalido             = errors.New("jogos.ano_invalido")
)

type JogosRepository interface {
	Criar(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error)
	Atualizar(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error)
	Excluir(ctx context.Context, id int32, usuarioID int32) error
	Listar(ctx context.Context, params repository.ListarJogosZeradosParams) ([]*repository.JogoZerado, int64, error)
	ObterFiltros(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error)
	ObterPorID(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error)
}

type JogosService struct {
	repo JogosRepository
	now  func() time.Time
}

func NewJogosService(repo JogosRepository) *JogosService {
	return &JogosService{repo: repo, now: time.Now}
}

func validarJogoZerado(nome, console, genero, tipo, review string, finalizadoEm time.Time, tempoJogado, nota int32, dificuldade string) error {
	if strings.TrimSpace(nome) == "" {
		return ErrNomeObrigatorio
	}
	if strings.TrimSpace(console) == "" {
		return ErrConsoleObrigatorio
	}
	if len([]rune(nome)) > 200 {
		return ErrNomeMuitoLongo
	}
	if len([]rune(console)) > 100 {
		return ErrConsoleMuitoLongo
	}
	if len([]rune(genero)) > 150 {
		return ErrGeneroMuitoLongo
	}
	if len([]rune(tipo)) > 50 {
		return ErrTipoMuitoLongo
	}
	if finalizadoEm.IsZero() {
		return ErrFinalizadoEmObrigatorio
	}
	if tempoJogado < 0 {
		return ErrTempoJogadoInvalido
	}
	if nota < 1 || nota > 11 {
		return ErrNotaInvalida
	}

	switch dificuldade {
	case "C", "B", "A", "AA", "AAA":
	default:
		return ErrDificuldadeInvalida
	}

	if len([]rune(review)) > 5000 {
		return ErrReviewMuitoLongo
	}

	return nil
}

func (s *JogosService) CriarJogoZerado(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
	if err := validarJogoZerado(params.Nome, params.Console, params.Genero, params.Tipo, params.Review, params.FinalizadoEm, params.TempoJogado, params.Nota, params.Dificuldade); err != nil {
		return nil, err
	}

	jogo, err := s.repo.Criar(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation || pgErr.Code == "23505" {
				return nil, ErrDestaqueAnoConflito
			}
			if pgErr.Code == "22001" {
				return nil, erroCampoMuitoLongo(pgErr.ColumnName)
			}
		}
		return nil, fmt.Errorf("salvar jogo zerado: %w", err)
	}

	return jogo, nil
}

func (s *JogosService) AtualizarJogoZerado(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
	if err := validarJogoZerado(params.Nome, params.Console, params.Genero, params.Tipo, params.Review, params.FinalizadoEm, params.TempoJogado, params.Nota, params.Dificuldade); err != nil {
		return nil, err
	}

	jogo, err := s.repo.Atualizar(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJogoNaoEncontrado
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation || pgErr.Code == "23505" {
				return nil, ErrDestaqueAnoConflito
			}
			if pgErr.Code == "22001" {
				return nil, erroCampoMuitoLongo(pgErr.ColumnName)
			}
		}
		return nil, fmt.Errorf("atualizar jogo zerado: %w", err)
	}

	return jogo, nil
}

func erroCampoMuitoLongo(campo string) error {
	switch campo {
	case "nome":
		return ErrNomeMuitoLongo
	case "console":
		return ErrConsoleMuitoLongo
	case "genero":
		return ErrGeneroMuitoLongo
	case "tipo":
		return ErrTipoMuitoLongo
	case "review":
		return ErrReviewMuitoLongo
	default:
		return ErrReviewMuitoLongo
	}
}

func (s *JogosService) ExcluirJogoZerado(ctx context.Context, id int32, usuarioID int32) error {
	err := s.repo.Excluir(ctx, id, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrJogoNaoEncontrado
		}
		return fmt.Errorf("excluir jogo zerado: %w", err)
	}
	return nil
}

type ListarJogosParams struct {
	UsuarioID   int32
	Busca       string
	Console     string
	Genero      string
	Tipo        string
	NotaMin     *int
	NotaMax     *int
	Ano         *int
	Dificuldade string
	Pagina      *int
	PorPagina   *int
}

type ResultadoListagem struct {
	Jogos        []*repository.JogoZerado
	Total        int64
	TotalPaginas int
	Pagina       int
	PorPagina    int
}

func (s *JogosService) ListarJogosZerados(ctx context.Context, params ListarJogosParams) (*ResultadoListagem, error) {
	pagina := 1
	if params.Pagina != nil {
		if *params.Pagina < 1 {
			return nil, ErrPaginaInvalida
		}
		pagina = *params.Pagina
	}

	porPagina := 24
	if params.PorPagina != nil {
		if *params.PorPagina < 1 || *params.PorPagina > 100 {
			return nil, ErrPorPaginaInvalido
		}
		porPagina = *params.PorPagina
	}

	var notaMinPtr *int32
	if params.NotaMin != nil {
		if *params.NotaMin < 1 || *params.NotaMin > 11 {
			return nil, ErrNotaFiltroInvalida
		}
		v := int32(*params.NotaMin)
		notaMinPtr = &v
	}

	var notaMaxPtr *int32
	if params.NotaMax != nil {
		if *params.NotaMax < 1 || *params.NotaMax > 11 {
			return nil, ErrNotaFiltroInvalida
		}
		v := int32(*params.NotaMax)
		notaMaxPtr = &v
	}

	if notaMinPtr != nil && notaMaxPtr != nil && *notaMinPtr > *notaMaxPtr {
		return nil, ErrNotaFaixaInvalida
	}

	var anoPtr *int32
	if params.Ano != nil {
		anoLimite := s.now().Year() + 1
		if *params.Ano < 1970 || *params.Ano > anoLimite {
			return nil, ErrAnoInvalido
		}
		v := int32(*params.Ano)
		anoPtr = &v
	}

	if params.Dificuldade != "" {
		switch params.Dificuldade {
		case "C", "B", "A", "AA", "AAA":
		default:
			return nil, ErrDificuldadeInvalida
		}
	}

	termoBusca := strings.TrimSpace(params.Busca)
	if termoBusca != "" {
		termoBusca = escaparLike(termoBusca)
	}

	termoGenero := strings.TrimSpace(params.Genero)
	if termoGenero != "" {
		termoGenero = escaparLike(termoGenero)
	}

	termoConsole := strings.TrimSpace(params.Console)
	termoTipo := strings.TrimSpace(params.Tipo)

	jogos, total, err := s.repo.Listar(ctx, repository.ListarJogosZeradosParams{
		UsuarioID:   params.UsuarioID,
		Busca:       termoBusca,
		Console:     termoConsole,
		Genero:      termoGenero,
		Tipo:        termoTipo,
		NotaMin:     notaMinPtr,
		NotaMax:     notaMaxPtr,
		Ano:         anoPtr,
		Dificuldade: params.Dificuldade,
		Pagina:      pagina,
		PorPagina:   porPagina,
	})
	if err != nil {
		return nil, fmt.Errorf("listar jogos zerados: %w", err)
	}

	var totalPaginas int
	if total > 0 {
		totalPaginas = int((total + int64(porPagina) - 1) / int64(porPagina))
	}

	return &ResultadoListagem{
		Jogos:        jogos,
		Total:        total,
		TotalPaginas: totalPaginas,
		Pagina:       pagina,
		PorPagina:    porPagina,
	}, nil
}

func escaparLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (s *JogosService) ObterOpcoesFiltros(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error) {
	filtros, err := s.repo.ObterFiltros(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter opcoes filtros: %w", err)
	}
	return filtros, nil
}

func (s *JogosService) ObterDetalhesJogoZerado(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
	jogo, err := s.repo.ObterPorID(ctx, id, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJogoNaoEncontrado
		}
		return nil, fmt.Errorf("obter detalhes jogo zerado: %w", err)
	}
	return jogo, nil
}
