// Package service contem as regras de negocio da aplicacao.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrJogandoNomeObrigatorio     = errors.New("jogando.nome_obrigatorio")
	ErrJogandoNomeMuitoLongo      = errors.New("jogando.nome_muito_longo")
	ErrJogandoIniciadoObrigatorio = errors.New("jogando.iniciado_em_obrigatorio")
	ErrJogandoIniciadoInvalido    = errors.New("jogando.iniciado_em_invalido")
	ErrJogandoIniciadoFuturo      = errors.New("jogando.iniciado_em_futuro")
	ErrJogandoIdInvalido          = errors.New("jogando.id_invalido")
	ErrJogandoNaoEncontrado       = errors.New("jogando.nao_encontrado")
	ErrJogandoEntradaInvalida     = errors.New("jogando.entrada_invalida")
)

type SalvarJogoEmAndamentoInput struct {
	UsuarioID   int32
	Nome        string
	IgdbID      *int32
	IgdbCapaURL *string
	IniciadoEm  *time.Time
}

type JogosEmAndamentoService struct {
	repo repository.JogosEmAndamentoRepository
	now  func() time.Time
}

func NewJogosEmAndamentoService(repo repository.JogosEmAndamentoRepository) *JogosEmAndamentoService {
	return &JogosEmAndamentoService{repo: repo, now: time.Now}
}

func (s *JogosEmAndamentoService) SetNow(nowFn func() time.Time) {
	s.now = nowFn
}

func (s *JogosEmAndamentoService) Criar(
	ctx context.Context,
	input SalvarJogoEmAndamentoInput,
) (*repository.JogoEmAndamento, error) {
	nome := strings.TrimSpace(input.Nome)
	if nome == "" {
		return nil, ErrJogandoNomeObrigatorio
	}
	if len([]rune(nome)) > 200 {
		return nil, ErrJogandoNomeMuitoLongo
	}
	if input.IniciadoEm == nil {
		return nil, ErrJogandoIniciadoObrigatorio
	}
	if input.IniciadoEm.IsZero() {
		return nil, ErrJogandoIniciadoInvalido
	}
	if input.IniciadoEm.After(s.now()) {
		return nil, ErrJogandoIniciadoFuturo
	}

	var capaURL *string
	if input.IgdbCapaURL != nil {
		valor := strings.TrimSpace(*input.IgdbCapaURL)
		if valor != "" {
			capaURL = &valor
		}
	}
	jogo, err := s.repo.Criar(ctx, repository.CriarJogoEmAndamentoParams{
		UsuarioID:   input.UsuarioID,
		Nome:        nome,
		IgdbID:      input.IgdbID,
		IgdbCapaURL: capaURL,
		IniciadoEm:  *input.IniciadoEm,
	})
	if err != nil {
		return nil, fmt.Errorf("criar jogo em andamento: %w", err)
	}
	return jogo, nil
}

func (s *JogosEmAndamentoService) Listar(ctx context.Context, usuarioID int32) ([]*repository.JogoEmAndamento, error) {
	jogos, err := s.repo.Listar(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar jogos em andamento: %w", err)
	}
	if jogos == nil {
		return []*repository.JogoEmAndamento{}, nil
	}
	return jogos, nil
}

func (s *JogosEmAndamentoService) Excluir(ctx context.Context, id int32, usuarioID int32) error {
	if id <= 0 {
		return ErrJogandoIdInvalido
	}
	if err := s.repo.Excluir(ctx, id, usuarioID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrJogandoNaoEncontrado
		}
		return fmt.Errorf("excluir jogo em andamento: %w", err)
	}
	return nil
}
