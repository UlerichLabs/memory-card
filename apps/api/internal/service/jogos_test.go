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
	criarFn            func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error)
	atualizarFn        func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error)
	excluirFn          func(ctx context.Context, id int32, usuarioID int32) error
	listarFn           func(ctx context.Context, params repository.ListarJogosZeradosParams) ([]*repository.JogoZerado, int64, error)
	obterFiltrosFn     func(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error)
	obterPorIDFn       func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error)
	obterResumoFn      func(ctx context.Context, usuarioID int32) ([]*repository.ItemResumoGameDoAno, error)
	definirGameDoAnoFn func(ctx context.Context, id int32, usuarioID int32) (*repository.DefinirGameDoAnoResultado, error)
	removerGameDoAnoFn func(ctx context.Context, id int32, usuarioID int32) error
}

func (m *mockJogosRepo) ObterPorID(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
	if m.obterPorIDFn != nil {
		return m.obterPorIDFn(ctx, id, usuarioID)
	}
	return nil, errors.New("nao implementado")
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

func (m *mockJogosRepo) Listar(ctx context.Context, params repository.ListarJogosZeradosParams) ([]*repository.JogoZerado, int64, error) {
	if m.listarFn != nil {
		return m.listarFn(ctx, params)
	}
	return nil, 0, errors.New("nao implementado")
}

func (m *mockJogosRepo) ObterFiltros(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error) {
	if m.obterFiltrosFn != nil {
		return m.obterFiltrosFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockJogosRepo) ObterResumoGameDoAno(ctx context.Context, usuarioID int32) ([]*repository.ItemResumoGameDoAno, error) {
	if m.obterResumoFn != nil {
		return m.obterResumoFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockJogosRepo) DefinirGameDoAno(ctx context.Context, id int32, usuarioID int32) (*repository.DefinirGameDoAnoResultado, error) {
	if m.definirGameDoAnoFn != nil {
		return m.definirGameDoAnoFn(ctx, id, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockJogosRepo) RemoverGameDoAno(ctx context.Context, id int32, usuarioID int32) error {
	if m.removerGameDoAnoFn != nil {
		return m.removerGameDoAnoFn(ctx, id, usuarioID)
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
			name: "review maior que 5000",
			modify: func(p *repository.CriarJogoZeradoParams) {
				p.Review = strings.Repeat("a", 5001)
			},
			expectedErr: ErrReviewMuitoLongo,
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
	tests := []struct {
		name      string
		id        int32
		usuarioID int32
	}{
		{
			name:      "inexistente",
			id:        999,
			usuarioID: 42,
		},
		{
			name:      "outro usuario",
			id:        1,
			usuarioID: 99,
		},
		{
			name:      "ja excluido",
			id:        2,
			usuarioID: 42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockJogosRepo{
				atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
					return nil, pgx.ErrNoRows
				},
			}

			svc := NewJogosService(repo)
			_, err := svc.AtualizarJogoZerado(context.Background(), repository.AtualizarJogoZeradoParams{
				ID:           tc.id,
				UsuarioID:    tc.usuarioID,
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
		})
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
			name: "review maior que 5000",
			modify: func(p *repository.AtualizarJogoZeradoParams) {
				p.Review = strings.Repeat("a", 5001)
			},
			expectedErr: ErrReviewMuitoLongo,
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

func TestJogosService_CamposMuitoLongos_CriarEAtualizar(t *testing.T) {
	now := time.Now()
	criarBase := repository.CriarJogoZeradoParams{
		Nome: "Jogo", Console: "Console", Genero: "Genero", Tipo: "Tipo",
		FinalizadoEm: now, TempoJogado: 1, Nota: 10, Dificuldade: "A",
	}
	atualizarBase := repository.AtualizarJogoZeradoParams{
		ID: 1, UsuarioID: 42, Nome: "Jogo", Console: "Console", Genero: "Genero", Tipo: "Tipo",
		FinalizadoEm: now, TempoJogado: 1, Nota: 10, Dificuldade: "A",
	}
	tests := []struct {
		name         string
		modifyCreate func(*repository.CriarJogoZeradoParams)
		modifyUpdate func(*repository.AtualizarJogoZeradoParams)
		expected     error
	}{
		{"nome", func(p *repository.CriarJogoZeradoParams) { p.Nome = strings.Repeat("a", 201) }, func(p *repository.AtualizarJogoZeradoParams) { p.Nome = strings.Repeat("a", 201) }, ErrNomeMuitoLongo},
		{"console", func(p *repository.CriarJogoZeradoParams) { p.Console = strings.Repeat("a", 101) }, func(p *repository.AtualizarJogoZeradoParams) { p.Console = strings.Repeat("a", 101) }, ErrConsoleMuitoLongo},
		{"genero", func(p *repository.CriarJogoZeradoParams) { p.Genero = strings.Repeat("a", 151) }, func(p *repository.AtualizarJogoZeradoParams) { p.Genero = strings.Repeat("a", 151) }, ErrGeneroMuitoLongo},
		{"tipo", func(p *repository.CriarJogoZeradoParams) { p.Tipo = strings.Repeat("a", 51) }, func(p *repository.AtualizarJogoZeradoParams) { p.Tipo = strings.Repeat("a", 51) }, ErrTipoMuitoLongo},
		{"review", func(p *repository.CriarJogoZeradoParams) { p.Review = strings.Repeat("a", 5001) }, func(p *repository.AtualizarJogoZeradoParams) { p.Review = strings.Repeat("a", 5001) }, ErrReviewMuitoLongo},
	}
	for _, tc := range tests {
		t.Run("criar_"+tc.name, func(t *testing.T) {
			params := criarBase
			tc.modifyCreate(&params)
			_, err := NewJogosService(&mockJogosRepo{}).CriarJogoZerado(context.Background(), params)
			if !errors.Is(err, tc.expected) {
				t.Fatalf("esperava %v, obteve %v", tc.expected, err)
			}
		})
		t.Run("atualizar_"+tc.name, func(t *testing.T) {
			params := atualizarBase
			tc.modifyUpdate(&params)
			_, err := NewJogosService(&mockJogosRepo{}).AtualizarJogoZerado(context.Background(), params)
			if !errors.Is(err, tc.expected) {
				t.Fatalf("esperava %v, obteve %v", tc.expected, err)
			}
		})
	}
}

func TestJogosService_CamposAmpliados_Sucesso(t *testing.T) {
	repo := &mockJogosRepo{criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
		return &repository.JogoZerado{Nome: params.Nome, Console: params.Console, Genero: params.Genero}, nil
	}}
	params := repository.CriarJogoZeradoParams{
		Nome: "Jogo", Console: strings.Repeat("c", 60), Genero: strings.Repeat("g", 100),
		FinalizadoEm: time.Now(), TempoJogado: 1, Nota: 10, Dificuldade: "A",
	}
	if _, err := NewJogosService(repo).CriarJogoZerado(context.Background(), params); err != nil {
		t.Fatalf("esperava sucesso, obteve %v", err)
	}
}

func TestJogosService_ReviewLimiteESemValor(t *testing.T) {
	repo := &mockJogosRepo{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			if len([]rune(params.Review)) > 5000 {
				t.Fatal("review excedeu o limite")
			}
			return &repository.JogoZerado{Review: params.Review}, nil
		},
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			return &repository.JogoZerado{Review: params.Review}, nil
		},
	}
	base := repository.CriarJogoZeradoParams{
		Nome: "Jogo", Console: "Console", FinalizadoEm: time.Now(), TempoJogado: 1, Nota: 10, Dificuldade: "A",
	}
	base.Review = strings.Repeat("r", 5000)
	if _, err := NewJogosService(repo).CriarJogoZerado(context.Background(), base); err != nil {
		t.Fatalf("review com 5000 caracteres falhou: %v", err)
	}
	base.Review = ""
	if _, err := NewJogosService(repo).CriarJogoZerado(context.Background(), base); err != nil {
		t.Fatalf("review vazia falhou: %v", err)
	}
	update := repository.AtualizarJogoZeradoParams{
		ID: 1, UsuarioID: 42, Nome: base.Nome, Console: base.Console, FinalizadoEm: base.FinalizadoEm,
		TempoJogado: 1, Nota: 10, Dificuldade: "A", Review: strings.Repeat("r", 5000),
	}
	if _, err := NewJogosService(repo).AtualizarJogoZerado(context.Background(), update); err != nil {
		t.Fatalf("review de edição com 5000 caracteres falhou: %v", err)
	}
}

func TestJogosService_SQLSTATE22001_ViraErroDeCampo(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected error
	}{
		{"nome", "nome", ErrNomeMuitoLongo},
		{"console", "console", ErrConsoleMuitoLongo},
		{"genero", "genero", ErrGeneroMuitoLongo},
		{"tipo", "tipo", ErrTipoMuitoLongo},
		{"review", "review", ErrReviewMuitoLongo},
		{"desconhecido", "outro", ErrReviewMuitoLongo},
	}

	for _, tc := range tests {
		t.Run("criar_"+tc.name, func(t *testing.T) {
			repo := &mockJogosRepo{criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
				return nil, &pgconn.PgError{Code: "22001", ColumnName: tc.column}
			}}
			params := repository.CriarJogoZeradoParams{Nome: "Jogo", Console: "Console", FinalizadoEm: time.Now(), TempoJogado: 1, Nota: 10, Dificuldade: "A"}
			_, err := NewJogosService(repo).CriarJogoZerado(context.Background(), params)
			if !errors.Is(err, tc.expected) {
				t.Fatalf("esperava %v, obteve %v", tc.expected, err)
			}
		})

		t.Run("atualizar_"+tc.name, func(t *testing.T) {
			repo := &mockJogosRepo{atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
				return nil, &pgconn.PgError{Code: "22001", ColumnName: tc.column}
			}}
			params := repository.AtualizarJogoZeradoParams{ID: 1, UsuarioID: 42, Nome: "Jogo", Console: "Console", FinalizadoEm: time.Now(), TempoJogado: 1, Nota: 10, Dificuldade: "A"}
			_, err := NewJogosService(repo).AtualizarJogoZerado(context.Background(), params)
			if !errors.Is(err, tc.expected) {
				t.Fatalf("esperava %v, obteve %v", tc.expected, err)
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
	tests := []struct {
		name      string
		id        int32
		usuarioID int32
	}{
		{
			name:      "inexistente",
			id:        999,
			usuarioID: 42,
		},
		{
			name:      "outro usuario",
			id:        1,
			usuarioID: 99,
		},
		{
			name:      "ja excluido",
			id:        2,
			usuarioID: 42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockJogosRepo{
				excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
					return pgx.ErrNoRows
				},
			}

			svc := NewJogosService(repo)
			err := svc.ExcluirJogoZerado(context.Background(), tc.id, tc.usuarioID)
			if !errors.Is(err, ErrJogoNaoEncontrado) {
				t.Fatalf("esperava ErrJogoNaoEncontrado, obteve: %v", err)
			}
		})
	}
}

func TestJogosService_ConflitoDestaque_CodigoString23505(t *testing.T) {
	now := time.Now()
	repoCriar := &mockJogosRepo{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, &pgconn.PgError{Code: "23505"}
		},
	}
	svcCriar := NewJogosService(repoCriar)
	_, err := svcCriar.CriarJogoZerado(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID:    1,
		Nome:         "Zelda",
		Console:      "NES",
		FinalizadoEm: now,
		TempoJogado:  100,
		Nota:         10,
		Dificuldade:  "A",
		Destaque:     true,
	})
	if !errors.Is(err, ErrDestaqueAnoConflito) {
		t.Fatalf("esperava ErrDestaqueAnoConflito, obteve: %v", err)
	}

	repoAtualizar := &mockJogosRepo{
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, &pgconn.PgError{Code: "23505"}
		},
	}
	svcAtualizar := NewJogosService(repoAtualizar)
	_, err = svcAtualizar.AtualizarJogoZerado(context.Background(), repository.AtualizarJogoZeradoParams{
		ID:           1,
		UsuarioID:    1,
		Nome:         "Zelda",
		Console:      "NES",
		FinalizadoEm: now,
		TempoJogado:  100,
		Nota:         10,
		Dificuldade:  "A",
		Destaque:     true,
	})
	if !errors.Is(err, ErrDestaqueAnoConflito) {
		t.Fatalf("esperava ErrDestaqueAnoConflito, obteve: %v", err)
	}
}

func TestJogosService_ErroInesperadoRepositorio(t *testing.T) {
	dbErr := errors.New("db error")
	repo := &mockJogosRepo{
		criarFn: func(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, dbErr
		},
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error) {
			return nil, dbErr
		},
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			return dbErr
		},
	}
	svc := NewJogosService(repo)
	now := time.Now()

	_, err := svc.CriarJogoZerado(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID: 1, Nome: "Jogo", Console: "NES", FinalizadoEm: now, TempoJogado: 1, Nota: 10, Dificuldade: "A",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("esperava dbErr, obteve: %v", err)
	}

	_, err = svc.AtualizarJogoZerado(context.Background(), repository.AtualizarJogoZeradoParams{
		ID: 1, UsuarioID: 1, Nome: "Jogo", Console: "NES", FinalizadoEm: now, TempoJogado: 1, Nota: 10, Dificuldade: "A",
	})
	if !errors.Is(err, dbErr) {
		t.Fatalf("esperava dbErr, obteve: %v", err)
	}

	err = svc.ExcluirJogoZerado(context.Background(), 1, 1)
	if !errors.Is(err, dbErr) {
		t.Fatalf("esperava dbErr, obteve: %v", err)
	}
}

func TestJogosService_ListarJogosZerados_SucessoEDefaults(t *testing.T) {
	chamouRepo := false
	var capturado repository.ListarJogosZeradosParams

	repo := &mockJogosRepo{
		listarFn: func(ctx context.Context, params repository.ListarJogosZeradosParams) ([]*repository.JogoZerado, int64, error) {
			chamouRepo = true
			capturado = params
			return []*repository.JogoZerado{
				{ID: 1, Nome: "Chrono Trigger"},
			}, 1, nil
		},
	}

	svc := NewJogosService(repo)
	res, err := svc.ListarJogosZerados(context.Background(), ListarJogosParams{
		UsuarioID: 42,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !chamouRepo {
		t.Fatal("esperava que o repositorio fosse chamado")
	}
	if capturado.UsuarioID != 42 {
		t.Fatalf("esperava usuarioID 42, obteve %d", capturado.UsuarioID)
	}
	if capturado.Pagina != 1 || capturado.PorPagina != 24 {
		t.Fatalf("esperava pagina=1 por_pagina=24, obteve %d, %d", capturado.Pagina, capturado.PorPagina)
	}
	if res.Pagina != 1 || res.PorPagina != 24 || res.Total != 1 || res.TotalPaginas != 1 {
		t.Fatalf("meta incorreto: %+v", res)
	}
	if len(res.Jogos) != 1 {
		t.Fatalf("esperava 1 jogo, obteve: %+v", res.Jogos)
	}
}

func TestJogosService_ListarJogosZerados_PaginacaoTotal(t *testing.T) {
	testes := []struct {
		nome            string
		total           int64
		porPagina       int
		esperadoPaginas int
	}{
		{"zero registros", 0, 24, 0},
		{"menos de uma pagina", 10, 24, 1},
		{"exatamente uma pagina", 24, 24, 1},
		{"uma pagina e um item", 25, 24, 2},
		{"duas paginas exatas", 48, 24, 2},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			repo := &mockJogosRepo{
				listarFn: func(ctx context.Context, params repository.ListarJogosZeradosParams) ([]*repository.JogoZerado, int64, error) {
					return []*repository.JogoZerado{}, tt.total, nil
				},
			}
			svc := NewJogosService(repo)
			res, err := svc.ListarJogosZerados(context.Background(), ListarJogosParams{
				UsuarioID: 1,
				PorPagina: &tt.porPagina,
			})
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if res.TotalPaginas != tt.esperadoPaginas {
				t.Fatalf("esperava %d paginas, obteve %d", tt.esperadoPaginas, res.TotalPaginas)
			}
		})
	}
}

func TestJogosService_ListarJogosZerados_Validacoes(t *testing.T) {
	intPtr := func(i int) *int { return &i }
	anoFuturoMaisDois := time.Now().Year() + 2

	testes := []struct {
		nome         string
		params       ListarJogosParams
		erroEsperado error
	}{
		{"pagina zero", ListarJogosParams{Pagina: intPtr(0)}, ErrPaginaInvalida},
		{"pagina negativa", ListarJogosParams{Pagina: intPtr(-1)}, ErrPaginaInvalida},
		{"por_pagina zero", ListarJogosParams{PorPagina: intPtr(0)}, ErrPorPaginaInvalido},
		{"por_pagina 101", ListarJogosParams{PorPagina: intPtr(101)}, ErrPorPaginaInvalido},
		{"por_pagina negativa", ListarJogosParams{PorPagina: intPtr(-5)}, ErrPorPaginaInvalido},
		{"nota_min zero", ListarJogosParams{NotaMin: intPtr(0)}, ErrNotaFiltroInvalida},
		{"nota_min 12", ListarJogosParams{NotaMin: intPtr(12)}, ErrNotaFiltroInvalida},
		{"nota_max zero", ListarJogosParams{NotaMax: intPtr(0)}, ErrNotaFiltroInvalida},
		{"nota_max 12", ListarJogosParams{NotaMax: intPtr(12)}, ErrNotaFiltroInvalida},
		{"nota_min maior que nota_max", ListarJogosParams{NotaMin: intPtr(10), NotaMax: intPtr(5)}, ErrNotaFaixaInvalida},
		{"ano anterior a 1970", ListarJogosParams{Ano: intPtr(1969)}, ErrAnoInvalido},
		{"ano alem do limite futuro", ListarJogosParams{Ano: &anoFuturoMaisDois}, ErrAnoInvalido},
		{"dificuldade invalida", ListarJogosParams{Dificuldade: "Z"}, ErrDificuldadeInvalida},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			svc := NewJogosService(&mockJogosRepo{})
			_, err := svc.ListarJogosZerados(context.Background(), tt.params)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("esperava erro %v, obteve %v", tt.erroEsperado, err)
			}
		})
	}
}

func TestJogosService_ListarJogosZerados_TratamentoStringsEEscape(t *testing.T) {
	var capturado repository.ListarJogosZeradosParams

	repo := &mockJogosRepo{
		listarFn: func(ctx context.Context, params repository.ListarJogosZeradosParams) ([]*repository.JogoZerado, int64, error) {
			capturado = params
			return []*repository.JogoZerado{}, 0, nil
		},
	}

	svc := NewJogosService(repo)
	_, err := svc.ListarJogosZerados(context.Background(), ListarJogosParams{
		UsuarioID:   1,
		Busca:       "  Zelda%_\\  ",
		Console:     "  SNES  ",
		Genero:      "  Action%_  ",
		Tipo:        "  Campanha  ",
		Dificuldade: "AA",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if capturado.Busca != `Zelda\%\_\\` {
		t.Fatalf("esperava busca escapada, obteve %q", capturado.Busca)
	}
	if capturado.Genero != `Action\%\_` {
		t.Fatalf("esperava genero escapado, obteve %q", capturado.Genero)
	}
	if capturado.Console != "SNES" {
		t.Fatalf("esperava console sem espacos, obteve %q", capturado.Console)
	}
	if capturado.Tipo != "Campanha" {
		t.Fatalf("esperava tipo sem espacos, obteve %q", capturado.Tipo)
	}
	if capturado.Dificuldade != "AA" {
		t.Fatalf("esperava dificuldade AA, obteve %q", capturado.Dificuldade)
	}
}

func TestJogosService_ListarJogosZerados_ErroRepositorio(t *testing.T) {
	dbErr := errors.New("db listar error")
	repo := &mockJogosRepo{
		listarFn: func(ctx context.Context, params repository.ListarJogosZeradosParams) ([]*repository.JogoZerado, int64, error) {
			return nil, 0, dbErr
		},
	}
	svc := NewJogosService(repo)
	_, err := svc.ListarJogosZerados(context.Background(), ListarJogosParams{UsuarioID: 1})
	if !errors.Is(err, dbErr) {
		t.Fatalf("esperava erro %v, obteve %v", dbErr, err)
	}
}

func TestJogosService_ObterOpcoesFiltros_Sucesso(t *testing.T) {
	esperado := &repository.OpcoesFiltros{
		Consoles: []string{"NES", "SNES"},
		Generos:  []string{"Action", "RPG"},
		Tipos:    []string{"Campanha", "DLC"},
		Anos:     []int{2025, 2024},
	}
	repo := &mockJogosRepo{
		obterFiltrosFn: func(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error) {
			if usuarioID != 99 {
				t.Fatalf("esperava usuarioID 99, obteve %d", usuarioID)
			}
			return esperado, nil
		},
	}
	svc := NewJogosService(repo)
	res, err := svc.ObterOpcoesFiltros(context.Background(), 99)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(res.Consoles) != 2 || len(res.Generos) != 2 || len(res.Tipos) != 2 || len(res.Anos) != 2 {
		t.Fatalf("resultado inesperado: %+v", res)
	}
}

func TestJogosService_ObterOpcoesFiltros_ErroRepositorio(t *testing.T) {
	dbErr := errors.New("db filtros error")
	repo := &mockJogosRepo{
		obterFiltrosFn: func(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error) {
			return nil, dbErr
		},
	}
	svc := NewJogosService(repo)
	_, err := svc.ObterOpcoesFiltros(context.Background(), 1)
	if !errors.Is(err, dbErr) {
		t.Fatalf("esperava erro %v, obteve %v", dbErr, err)
	}
}

func TestJogosService_ObterDetalhesJogoZerado_Sucesso(t *testing.T) {
	esperado := &repository.JogoZerado{
		ID:            10,
		Numero:        2,
		UsuarioID:     42,
		Nome:          "Super Mario World",
		Console:       "SNES",
		TempoJogado:   18000,
		Nota:          10,
		Dificuldade:   "B",
		IgdbDescricao: "Mario saves dinosaur land.",
	}
	repo := &mockJogosRepo{
		obterPorIDFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
			if id != 10 {
				t.Fatalf("esperava id 10, obteve %d", id)
			}
			if usuarioID != 42 {
				t.Fatalf("esperava usuarioID 42, obteve %d", usuarioID)
			}
			return esperado, nil
		},
	}
	svc := NewJogosService(repo)
	jogo, err := svc.ObterDetalhesJogoZerado(context.Background(), 10, 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if jogo.ID != 10 || jogo.Numero != 2 || jogo.UsuarioID != 42 || jogo.IgdbDescricao != "Mario saves dinosaur land." {
		t.Fatalf("jogo inesperado: %+v", jogo)
	}
}

func TestJogosService_ObterDetalhesJogoZerado_NaoEncontrado(t *testing.T) {
	repo := &mockJogosRepo{
		obterPorIDFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
			return nil, pgx.ErrNoRows
		},
	}
	svc := NewJogosService(repo)
	_, err := svc.ObterDetalhesJogoZerado(context.Background(), 999, 42)
	if !errors.Is(err, ErrJogoNaoEncontrado) {
		t.Fatalf("esperava ErrJogoNaoEncontrado, obteve %v", err)
	}
}

func TestJogosService_ObterDetalhesJogoZerado_ErroRepositorio(t *testing.T) {
	dbErr := errors.New("db error")
	repo := &mockJogosRepo{
		obterPorIDFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error) {
			return nil, dbErr
		},
	}
	svc := NewJogosService(repo)
	_, err := svc.ObterDetalhesJogoZerado(context.Background(), 10, 42)
	if !errors.Is(err, dbErr) {
		t.Fatalf("esperava erro %v, obteve %v", dbErr, err)
	}
}

func TestJogosService_ObterResumoGameDoAno_SucessoEErro(t *testing.T) {
	chamouRepo := false
	esperado := []*repository.ItemResumoGameDoAno{
		{
			Ano:        2026,
			TotalJogos: 9,
			GameDoAno:  &repository.JogoZerado{ID: 20, Nome: "Elden Ring", Destaque: true},
		},
		{
			Ano:        2025,
			TotalJogos: 7,
			GameDoAno:  nil,
		},
	}

	repo := &mockJogosRepo{
		obterResumoFn: func(ctx context.Context, usuarioID int32) ([]*repository.ItemResumoGameDoAno, error) {
			chamouRepo = true
			if usuarioID != 42 {
				t.Fatalf("esperava usuarioID 42, obteve %d", usuarioID)
			}
			return esperado, nil
		},
	}

	svc := NewJogosService(repo)
	res, err := svc.ObterResumoGameDoAno(context.Background(), 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !chamouRepo || len(res) != 2 {
		t.Fatalf("resultado inesperado: %+v", res)
	}
	if res[0].Ano != 2026 || res[0].GameDoAno.ID != 20 || res[1].GameDoAno != nil {
		t.Fatalf("campos incorretos: %+v", res)
	}

	dbErr := errors.New("db error")
	repoErro := &mockJogosRepo{
		obterResumoFn: func(ctx context.Context, usuarioID int32) ([]*repository.ItemResumoGameDoAno, error) {
			return nil, dbErr
		},
	}
	_, err = NewJogosService(repoErro).ObterResumoGameDoAno(context.Background(), 42)
	if !errors.Is(err, dbErr) {
		t.Fatalf("esperava dbErr, obteve %v", err)
	}
}

func TestJogosService_ListarJogosZerados_Ordenacao(t *testing.T) {
	testes := []struct {
		nome           string
		ordenarEntrada string
		esperadoParam  string
		esperaErro     error
	}{
		{"padrao vazio", "", "recentes", nil},
		{"recentes explicito", "recentes", "recentes", nil},
		{"nota explicito", "nota", "nota", nil},
		{"com espacos", "  nota  ", "nota", nil},
		{"invalido", "invalido", "", ErrOrdenacaoInvalida},
		{"invalido numero", "123", "", ErrOrdenacaoInvalida},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			var capturado repository.ListarJogosZeradosParams
			repo := &mockJogosRepo{
				listarFn: func(ctx context.Context, params repository.ListarJogosZeradosParams) ([]*repository.JogoZerado, int64, error) {
					capturado = params
					return []*repository.JogoZerado{}, 0, nil
				},
			}
			svc := NewJogosService(repo)
			_, err := svc.ListarJogosZerados(context.Background(), ListarJogosParams{
				UsuarioID: 42,
				Ordenar:   tt.ordenarEntrada,
			})
			if tt.esperaErro != nil {
				if !errors.Is(err, tt.esperaErro) {
					t.Fatalf("esperava erro %v, obteve %v", tt.esperaErro, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if capturado.Ordenar != tt.esperadoParam {
				t.Fatalf("esperava ordenar=%q, obteve %q", tt.esperadoParam, capturado.Ordenar)
			}
		})
	}
}

func TestJogosService_DefinirGameDoAno_SucessoETroca(t *testing.T) {
	antID := int32(14)
	esperado := &repository.DefinirGameDoAnoResultado{
		Ano:        2025,
		GameDoAno:  &repository.JogoZerado{ID: 20, Nome: "Novo Destaque", Destaque: true},
		AnteriorID: &antID,
	}

	repo := &mockJogosRepo{
		definirGameDoAnoFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.DefinirGameDoAnoResultado, error) {
			if id != 20 || usuarioID != 42 {
				t.Fatalf("parametros incorretos: id=%d, usuarioID=%d", id, usuarioID)
			}
			return esperado, nil
		},
	}

	svc := NewJogosService(repo)
	res, err := svc.DefinirGameDoAno(context.Background(), 20, 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if res.Ano != 2025 || res.GameDoAno.ID != 20 || res.AnteriorID == nil || *res.AnteriorID != 14 {
		t.Fatalf("resultado inesperado: %+v", res)
	}
}

func TestJogosService_DefinirGameDoAno_Idempotente(t *testing.T) {
	esperado := &repository.DefinirGameDoAnoResultado{
		Ano:        2025,
		GameDoAno:  &repository.JogoZerado{ID: 20, Nome: "Ja Destaque", Destaque: true},
		AnteriorID: nil,
	}

	repo := &mockJogosRepo{
		definirGameDoAnoFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.DefinirGameDoAnoResultado, error) {
			return esperado, nil
		},
	}

	svc := NewJogosService(repo)
	res, err := svc.DefinirGameDoAno(context.Background(), 20, 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if res.AnteriorID != nil {
		t.Fatalf("esperava anterior_id nil em chamada idempotente, obteve %v", res.AnteriorID)
	}
}

func TestJogosService_DefinirGameDoAno_NaoEncontrado(t *testing.T) {
	repo := &mockJogosRepo{
		definirGameDoAnoFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.DefinirGameDoAnoResultado, error) {
			return nil, pgx.ErrNoRows
		},
	}

	svc := NewJogosService(repo)
	_, err := svc.DefinirGameDoAno(context.Background(), 999, 42)
	if !errors.Is(err, ErrJogoNaoEncontrado) {
		t.Fatalf("esperava ErrJogoNaoEncontrado, obteve %v", err)
	}
}

func TestJogosService_DefinirGameDoAno_ConflitoUnicidade(t *testing.T) {
	repo := &mockJogosRepo{
		definirGameDoAnoFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.DefinirGameDoAnoResultado, error) {
			return nil, &pgconn.PgError{Code: pgerrcode.UniqueViolation, ConstraintName: "idx_destaque_por_ano"}
		},
	}

	svc := NewJogosService(repo)
	_, err := svc.DefinirGameDoAno(context.Background(), 20, 42)
	if !errors.Is(err, ErrDestaqueAnoConflito) {
		t.Fatalf("esperava ErrDestaqueAnoConflito, obteve %v", err)
	}
}

func TestJogosService_RemoverGameDoAno_SucessoENaoEncontrado(t *testing.T) {
	chamouRepo := false
	repo := &mockJogosRepo{
		removerGameDoAnoFn: func(ctx context.Context, id int32, usuarioID int32) error {
			chamouRepo = true
			if id != 20 || usuarioID != 42 {
				t.Fatalf("parametros incorretos: id=%d, usuarioID=%d", id, usuarioID)
			}
			return nil
		},
	}

	svc := NewJogosService(repo)
	err := svc.RemoverGameDoAno(context.Background(), 20, 42)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !chamouRepo {
		t.Fatal("esperava chamada ao repo.RemoverGameDoAno")
	}

	repo404 := &mockJogosRepo{
		removerGameDoAnoFn: func(ctx context.Context, id int32, usuarioID int32) error {
			return pgx.ErrNoRows
		},
	}
	err = NewJogosService(repo404).RemoverGameDoAno(context.Background(), 999, 42)
	if !errors.Is(err, ErrJogoNaoEncontrado) {
		t.Fatalf("esperava ErrJogoNaoEncontrado, obteve %v", err)
	}
}

