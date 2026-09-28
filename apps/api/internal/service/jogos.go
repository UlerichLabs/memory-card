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
)

type JogosRepository interface {
	Criar(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error)
	Atualizar(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error)
	Excluir(ctx context.Context, id int32, usuarioID int32) error
}

type JogosService struct {
	repo JogosRepository
}

func NewJogosService(repo JogosRepository) *JogosService {
	return &JogosService{repo: repo}
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
