package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type mockJogosRepo struct {
	criarFn     func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error)
	atualizarFn func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error)
	excluirFn   func(ctx context.Context, id int32, usuarioID int32) error
}

func (m *mockJogosRepo) Criar(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
	if m.criarFn != nil {
		return m.criarFn(ctx, params)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockJogosRepo) Atualizar(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
	if m.atualizarFn != nil {
		return m.atualizarFn(ctx, params)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockJogosRepo) Excluir(ctx context.Context, id int32, usuarioID int32) error {
	if m.excluirFn != nil {
		return m.excluirFn(ctx, id, usuarioID)
	}
	return errors.New("nao implementado")
}

func TestJogosService_CriarJogoZerado_Sucesso(t *testing.T) {
	now := time.Now()
	repo := &mockJogosRepo{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			return &repository.JogoZerado{
				ID:           1,
				UsuarioID:    params.UsuarioID,
				Nome:         params.Nome,
				Console:      params.Console,
				TempoJogado:  params.TempoJogado,
				Nota:         params.Nota,
				Dificuldade:  params.Dificuldade,
				FinalizadoEm: params.FinalizadoEm,
			}, nil
		},
	}

	svc := NewJogosService(repo)
	jogo, err := svc.CriarJogoZerado(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID:    1,
		Nome:         "Chrono Trigger",
		Console:      "SNES",
		FinalizadoEm: now,
		TempoJogado:  36000,
		Nota:         10,
		Dificuldade:  "A",
	})
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if jogo.Nome != "Chrono Trigger" || jogo.Nota != 10 {
		t.Fatalf("jogo inesperado: %+v", jogo)
	}
}

func TestJogosService_CriarJogoZerado_Validacoes(t *testing.T) {
	now := time.Now()
	baseParams := repository.CriarJogoZeradoParams{
		UsuarioID:    1,
		Nome:         "Chrono Trigger",
		Console:      "SNES",
		FinalizadoEm: now,
		TempoJogado:  1000,
		Nota:         10,
		Dificuldade:  "A",
	}

	tests := []struct {
		name        string
		modify      func(p *repository.CriarJogoZeradoParams)
		expectedErr error
	}{
		{
			name:        "nome vazio",
			modify:      func(p *repository.CriarJogoZeradoParams) { p.Nome = "   " },
			expectedErr: ErrNomeObrigatorio,
		},
		{
			name:        "console vazio",
			modify:      func(p *repository.CriarJogoZeradoParams) { p.Console = "" },
			expectedErr: ErrConsoleObrigatorio,
		},
		{
			name:        "finalizado_em zero",
			modify:      func(p *repository.CriarJogoZeradoParams) { p.FinalizadoEm = time.Time{} },
			expectedErr: ErrFinalizadoEmObrigatorio,
		},
		{
			name:        "tempo_jogado negativo",
			modify:      func(p *repository.CriarJogoZeradoParams) { p.TempoJogado = -1 },
			expectedErr: ErrTempoJogadoInvalido,
		},
		{
			name:        "nota menor que 1",
			modify:      func(p *repository.CriarJogoZeradoParams) { p.Nota = 0 },
			expectedErr: ErrNotaInvalida,
		},
		{
			name:        "nota maior que 11",
			modify:      func(p *repository.CriarJogoZeradoParams) { p.Nota = 12 },
			expectedErr: ErrNotaInvalida,
		},
		{
			name:        "dificuldade invalida",
			modify:      func(p *repository.CriarJogoZeradoParams) { p.Dificuldade = "INVALID" },
			expectedErr: ErrDificuldadeInvalida,
		},
		{
			name: "condicao zeramento maior que 500",
			modify: func(p *repository.CriarJogoZeradoParams) {
				p.CondicaoZeramento = strings.Repeat("a", 501)
			},
			expectedErr: ErrCondicaoZeramentoInvalida,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := baseParams
			tc.modify(&params)
			svc := NewJogosService(&mockJogosRepo{})
			_, err := svc.CriarJogoZerado(context.Background(), params)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("esperava erro %v, obteve: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestJogosService_CriarJogoZerado_ConflitoDestaque(t *testing.T) {
	now := time.Now()
	repo := &mockJogosRepo{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, &pgconn.PgError{
				Code:           pgerrcode.UniqueViolation,
				ConstraintName: "idx_destaque_por_ano",
			}
		},
	}

	svc := NewJogosService(repo)
	_, err := svc.CriarJogoZerado(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID:    1,
		Nome:         "Chrono Trigger",
		Console:      "SNES",
		FinalizadoEm: now,
		TempoJogado:  3600,
		Nota:         10,
		Dificuldade:  "AA",
		Destaque:     true,
	})
	if !errors.Is(err, ErrDestaqueAnoConflito) {
		t.Fatalf("esperava ErrDestaqueAnoConflito, obteve: %v", err)
	}
}

func TestJogosService_AtualizarJogoZerado_Sucesso(t *testing.T) {
	now := time.Now()
	repo := &mockJogosRepo{
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			if params.ID != 1 || params.UsuarioID != 42 {
				t.Fatalf("parametros incorretos: id=%d, usuarioID=%d", params.ID, params.UsuarioID)
			}
			return &repository.JogoZerado{
				ID:           params.ID,
				UsuarioID:    params.UsuarioID,
				Nome:         params.Nome,
				Console:      params.Console,
				TempoJogado:  params.TempoJogado,
				Nota:         params.Nota,
				Dificuldade:  params.Dificuldade,
				FinalizadoEm: params.FinalizadoEm,
			}, nil
		},
	}

	svc := NewJogosService(repo)
	jogo, err := svc.AtualizarJogoZerado(context.Background(), repository.AtualizarJogoZeradoParams{
		ID:           1,
		UsuarioID:    42,
		Nome:         "Chrono Trigger Remake",
		Console:      "Nintendo Switch",
		FinalizadoEm: now,
		TempoJogado:  40000,
		Nota:         11,
		Dificuldade:  "AAA",
	})
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if jogo.Nome != "Chrono Trigger Remake" || jogo.Nota != 11 || jogo.Console != "Nintendo Switch" {
		t.Fatalf("jogo inesperado: %+v", jogo)
	}
}

func TestJogosService_AtualizarJogoZerado_NaoEncontrado(t *testing.T) {
	repo := &mockJogosRepo{
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, pgx.ErrNoRows
		},
	}

	svc := NewJogosService(repo)
	_, err := svc.AtualizarJogoZerado(context.Background(), repository.AtualizarJogoZeradoParams{
		ID:           999,
		UsuarioID:    42,
		Nome:         "Inexistente",
		Console:      "SNES",
		FinalizadoEm: time.Now(),
		TempoJogado:  100,
		Nota:         8,
		Dificuldade:  "B",
	})
	if !errors.Is(err, ErrJogoNaoEncontrado) {
		t.Fatalf("esperava ErrJogoNaoEncontrado, obteve: %v", err)
	}
}

func TestJogosService_AtualizarJogoZerado_ConflitoDestaque(t *testing.T) {
	repo := &mockJogosRepo{
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, &pgconn.PgError{
				Code:           pgerrcode.UniqueViolation,
				ConstraintName: "idx_destaque_por_ano",
			}
		},
	}

	svc := NewJogosService(repo)
	_, err := svc.AtualizarJogoZerado(context.Background(), repository.AtualizarJogoZeradoParams{
		ID:           1,
		UsuarioID:    42,
		Nome:         "Zelda",
		Console:      "NES",
		FinalizadoEm: time.Now(),
		TempoJogado:  100,
		Nota:         10,
		Dificuldade:  "A",
		Destaque:     true,
	})
	if !errors.Is(err, ErrDestaqueAnoConflito) {
		t.Fatalf("esperava ErrDestaqueAnoConflito, obteve: %v", err)
	}
}

func TestJogosService_AtualizarJogoZerado_Validacoes(t *testing.T) {
	now := time.Now()
	baseParams := repository.AtualizarJogoZeradoParams{
		ID:           1,
		UsuarioID:    42,
		Nome:         "Chrono Trigger",
		Console:      "SNES",
		FinalizadoEm: now,
		TempoJogado:  1000,
		Nota:         10,
		Dificuldade:  "A",
	}

	tests := []struct {
		name        string
		modify      func(p *repository.AtualizarJogoZeradoParams)
		expectedErr error
	}{
		{
			name:        "nome vazio",
			modify:      func(p *repository.AtualizarJogoZeradoParams) { p.Nome = "   " },
			expectedErr: ErrNomeObrigatorio,
		},
		{
			name:        "console vazio",
			modify:      func(p *repository.AtualizarJogoZeradoParams) { p.Console = "" },
			expectedErr: ErrConsoleObrigatorio,
		},
		{
			name:        "finalizado_em zero",
			modify:      func(p *repository.AtualizarJogoZeradoParams) { p.FinalizadoEm = time.Time{} },
			expectedErr: ErrFinalizadoEmObrigatorio,
		},
		{
			name:        "tempo_jogado negativo",
			modify:      func(p *repository.AtualizarJogoZeradoParams) { p.TempoJogado = -1 },
			expectedErr: ErrTempoJogadoInvalido,
		},
		{
			name:        "nota menor que 1",
			modify:      func(p *repository.AtualizarJogoZeradoParams) { p.Nota = 0 },
			expectedErr: ErrNotaInvalida,
		},
		{
			name:        "nota maior que 11",
			modify:      func(p *repository.AtualizarJogoZeradoParams) { p.Nota = 12 },
			expectedErr: ErrNotaInvalida,
		},
		{
			name:        "dificuldade invalida",
			modify:      func(p *repository.AtualizarJogoZeradoParams) { p.Dificuldade = "INVALID" },
			expectedErr: ErrDificuldadeInvalida,
		},
		{
			name: "condicao zeramento maior que 500",
			modify: func(p *repository.AtualizarJogoZeradoParams) {
				p.CondicaoZeramento = strings.Repeat("a", 501)
			},
			expectedErr: ErrCondicaoZeramentoInvalida,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := baseParams
			tc.modify(&params)
			svc := NewJogosService(&mockJogosRepo{})
			_, err := svc.AtualizarJogoZerado(context.Background(), params)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("esperava erro %v, obteve: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestJogosService_ExcluirJogoZerado_Sucesso(t *testing.T) {
	var excluirChamado bool
	repo := &mockJogosRepo{
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			if id == 1 && usuarioID == 42 {
				excluirChamado = true
				return nil
			}
			return errors.New("parametros invalidos")
		},
	}

	svc := NewJogosService(repo)
	err := svc.ExcluirJogoZerado(context.Background(), 1, 42)
	if err != nil {
		t.Fatalf("esperava sucesso, obteve: %v", err)
	}
	if !excluirChamado {
		t.Fatal("esperava que repo.Excluir fosse chamado")
	}
}

func TestJogosService_ExcluirJogoZerado_NaoEncontrado(t *testing.T) {
	repo := &mockJogosRepo{
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			return pgx.ErrNoRows
		},
	}

	svc := NewJogosService(repo)
	err := svc.ExcluirJogoZerado(context.Background(), 999, 42)
	if !errors.Is(err, ErrJogoNaoEncontrado) {
		t.Fatalf("esperava ErrJogoNaoEncontrado, obteve: %v", err)
	}
}

