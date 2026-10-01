package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type mockJogosEmAndamentoRepo struct {
	criarFn   func(context.Context, repository.CriarJogoEmAndamentoParams) (*repository.JogoEmAndamento, error)
	listarFn  func(context.Context, int32) ([]*repository.JogoEmAndamento, error)
	excluirFn func(context.Context, int32, int32) error
}

func (m *mockJogosEmAndamentoRepo) Criar(ctx context.Context, p repository.CriarJogoEmAndamentoParams) (*repository.JogoEmAndamento, error) {
	return m.criarFn(ctx, p)
}

func (m *mockJogosEmAndamentoRepo) Listar(ctx context.Context, id int32) ([]*repository.JogoEmAndamento, error) {
	return m.listarFn(ctx, id)
}

func (m *mockJogosEmAndamentoRepo) Excluir(ctx context.Context, id int32, usuarioID int32) error {
	return m.excluirFn(ctx, id, usuarioID)
}

func TestJogosEmAndamentoService_Criar(t *testing.T) {
	agora := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	iniciado := agora.Add(-24 * time.Hour)
	var recebido repository.CriarJogoEmAndamentoParams
	svc := NewJogosEmAndamentoService(&mockJogosEmAndamentoRepo{
		criarFn: func(ctx context.Context, p repository.CriarJogoEmAndamentoParams) (*repository.JogoEmAndamento, error) {
			recebido = p
			return &repository.JogoEmAndamento{ID: 1, Nome: p.Nome}, nil
		},
	})
	svc.SetNow(func() time.Time { return agora })

	jogo, err := svc.Criar(context.Background(), SalvarJogoEmAndamentoInput{
		UsuarioID:  7,
		Nome:       "  Elden Ring  ",
		IniciadoEm: &iniciado,
	})
	if err != nil || jogo.ID != 1 {
		t.Fatalf("criação inesperada: jogo=%v erro=%v", jogo, err)
	}
	if recebido.UsuarioID != 7 || recebido.Nome != "Elden Ring" || !recebido.IniciadoEm.Equal(iniciado) {
		t.Fatalf("parâmetros inesperados: %+v", recebido)
	}
}

func TestJogosEmAndamentoService_Validacoes(t *testing.T) {
	agora := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	svc := NewJogosEmAndamentoService(&mockJogosEmAndamentoRepo{})
	svc.SetNow(func() time.Time { return agora })
	testes := []struct {
		nome     string
		input    SalvarJogoEmAndamentoInput
		esperado error
	}{
		{"nome vazio", SalvarJogoEmAndamentoInput{Nome: "  "}, ErrJogandoNomeObrigatorio},
		{"nome longo", SalvarJogoEmAndamentoInput{Nome: string(make([]byte, 201))}, ErrJogandoNomeMuitoLongo},
		{"data ausente", SalvarJogoEmAndamentoInput{Nome: "Jogo"}, ErrJogandoIniciadoObrigatorio},
		{"data zero", SalvarJogoEmAndamentoInput{Nome: "Jogo", IniciadoEm: func() *time.Time { valor := time.Time{}; return &valor }()}, ErrJogandoIniciadoInvalido},
		{"data futura", SalvarJogoEmAndamentoInput{Nome: "Jogo", IniciadoEm: func() *time.Time { v := agora.Add(time.Hour); return &v }()}, ErrJogandoIniciadoFuturo},
	}
	for _, teste := range testes {
		t.Run(teste.nome, func(t *testing.T) {
			_, err := svc.Criar(context.Background(), teste.input)
			if !errors.Is(err, teste.esperado) {
				t.Fatalf("esperava %v, obteve %v", teste.esperado, err)
			}
		})
	}
}

func TestJogosEmAndamentoService_ListarVazioEExcluirIdInvalido(t *testing.T) {
	svc := NewJogosEmAndamentoService(&mockJogosEmAndamentoRepo{
		listarFn:  func(context.Context, int32) ([]*repository.JogoEmAndamento, error) { return nil, nil },
		excluirFn: func(context.Context, int32, int32) error { return nil },
	})
	jogos, err := svc.Listar(context.Background(), 42)
	if err != nil || jogos == nil || len(jogos) != 0 {
		t.Fatalf("esperava lista vazia não nula, obteve jogos=%v erro=%v", jogos, err)
	}
	if err := svc.Excluir(context.Background(), 0, 42); !errors.Is(err, ErrJogandoIdInvalido) {
		t.Fatalf("esperava id inválido, obteve %v", err)
	}
}

func TestJogosEmAndamentoService_Excluir(t *testing.T) {
	svc := NewJogosEmAndamentoService(&mockJogosEmAndamentoRepo{
		excluirFn: func(context.Context, int32, int32) error { return pgx.ErrNoRows },
	})
	if err := svc.Excluir(context.Background(), 4, 8); !errors.Is(err, ErrJogandoNaoEncontrado) {
		t.Fatalf("esperava não encontrado, obteve %v", err)
	}
}
