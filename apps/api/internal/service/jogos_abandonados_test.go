package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type mockAbandonadosRepo struct {
	criarFn         func(ctx context.Context, params repository.CriarJogoAbandonadoParams) (*repository.JogoAbandonado, error)
	atualizarFn     func(ctx context.Context, params repository.AtualizarJogoAbandonadoParams) (*repository.JogoAbandonado, error)
	excluirFn       func(ctx context.Context, id int32, usuarioID int32) error
	listarFn        func(ctx context.Context, params repository.ListarJogosAbandonadosParams) ([]*repository.JogoAbandonado, int64, error)
	obterPorIDFn    func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoAbandonado, error)
	obterConsolesFn func(ctx context.Context, usuarioID int32) ([]string, error)
	obterTotalFn    func(ctx context.Context, usuarioID int32) (int64, error)
}

func (m *mockAbandonadosRepo) Criar(
	ctx context.Context,
	params repository.CriarJogoAbandonadoParams,
) (*repository.JogoAbandonado, error) {
	if m.criarFn != nil {
		return m.criarFn(ctx, params)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockAbandonadosRepo) Atualizar(
	ctx context.Context,
	params repository.AtualizarJogoAbandonadoParams,
) (*repository.JogoAbandonado, error) {
	if m.atualizarFn != nil {
		return m.atualizarFn(ctx, params)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockAbandonadosRepo) Excluir(ctx context.Context, id int32, usuarioID int32) error {
	if m.excluirFn != nil {
		return m.excluirFn(ctx, id, usuarioID)
	}
	return errors.New("nao implementado")
}

func (m *mockAbandonadosRepo) Listar(
	ctx context.Context,
	params repository.ListarJogosAbandonadosParams,
) ([]*repository.JogoAbandonado, int64, error) {
	if m.listarFn != nil {
		return m.listarFn(ctx, params)
	}
	return nil, 0, errors.New("nao implementado")
}

func (m *mockAbandonadosRepo) ObterPorID(
	ctx context.Context,
	id int32,
	usuarioID int32,
) (*repository.JogoAbandonado, error) {
	if m.obterPorIDFn != nil {
		return m.obterPorIDFn(ctx, id, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockAbandonadosRepo) ObterConsoles(ctx context.Context, usuarioID int32) ([]string, error) {
	if m.obterConsolesFn != nil {
		return m.obterConsolesFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockAbandonadosRepo) ObterTotal(ctx context.Context, usuarioID int32) (int64, error) {
	if m.obterTotalFn != nil {
		return m.obterTotalFn(ctx, usuarioID)
	}
	return 0, errors.New("nao implementado")
}

func TestJogosAbandonadosService_Criar_Sucesso(t *testing.T) {
	refTime := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	repo := &mockAbandonadosRepo{
		criarFn: func(ctx context.Context, params repository.CriarJogoAbandonadoParams) (*repository.JogoAbandonado, error) {
			if params.UsuarioID != 1 {
				t.Fatalf("esperava usuarioID 1, obteve %d", params.UsuarioID)
			}
			if params.Nome != "Dark Souls" {
				t.Fatalf("esperava nome Dark Souls, obteve %s", params.Nome)
			}
			if params.Console != "PS3" {
				t.Fatalf("esperava console PS3, obteve %s", params.Console)
			}
			if params.TempoJogado != 7200 {
				t.Fatalf("esperava tempo 7200s, obteve %d", params.TempoJogado)
			}
			if params.Motivo == nil || *params.Motivo != "Muitas mortes" {
				t.Fatalf("motivo inesperado: %v", params.Motivo)
			}
			return &repository.JogoAbandonado{
				ID:           10,
				UsuarioID:    params.UsuarioID,
				Nome:         params.Nome,
				Console:      params.Console,
				TempoJogado:  params.TempoJogado,
				Motivo:       params.Motivo,
				AbandonadoEm: params.AbandonadoEm,
			}, nil
		},
	}

	svc := NewJogosAbandonadosService(repo)
	svc.SetNow(func() time.Time { return refTime })

	motivo := "  Muitas mortes  "
	jogo, err := svc.CriarJogoAbandonado(context.Background(), SalvarJogoAbandonadoInput{
		UsuarioID:          1,
		Nome:               "  Dark Souls  ",
		Console:            "  PS3  ",
		TempoJogadoHoras:   2,
		TempoJogadoMinutos: 0,
		Motivo:             &motivo,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if jogo.ID != 10 {
		t.Fatalf("esperava ID 10, obteve %d", jogo.ID)
	}
	if !jogo.AbandonadoEm.Equal(refTime) {
		t.Fatalf("esperava abandonado_em %v, obteve %v", refTime, jogo.AbandonadoEm)
	}
}

func TestJogosAbandonadosService_ConversaoTempo(t *testing.T) {
	tests := []struct {
		name     string
		input    SalvarJogoAbandonadoInput
		esperado int32
	}{
		{
			name: "horas_minutos_segundos",
			input: SalvarJogoAbandonadoInput{
				TempoJogadoHoras:    1,
				TempoJogadoMinutos:  30,
				TempoJogadoSegundos: 45,
			},
			esperado: 5445,
		},
		{
			name: "tempo_jogado_direto",
			input: SalvarJogoAbandonadoInput{
				TempoJogado: func() *int32 { v := int32(3600); return &v }(),
			},
			esperado: 3600,
		},
		{
			name:     "todos_ausentes",
			input:    SalvarJogoAbandonadoInput{},
			esperado: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var tempoRecebido int32
			repo := &mockAbandonadosRepo{
				criarFn: func(ctx context.Context, params repository.CriarJogoAbandonadoParams) (*repository.JogoAbandonado, error) {
					tempoRecebido = params.TempoJogado
					return &repository.JogoAbandonado{ID: 1}, nil
				},
			}
			svc := NewJogosAbandonadosService(repo)
			tc.input.Nome = "Jogo"
			tc.input.Console = "SNES"
			_, err := svc.CriarJogoAbandonado(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if tempoRecebido != tc.esperado {
				t.Fatalf("esperava %d segundos, obteve %d", tc.esperado, tempoRecebido)
			}
		})
	}
}

func TestJogosAbandonadosService_Validacoes(t *testing.T) {
	agora := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	svc := NewJogosAbandonadosService(&mockAbandonadosRepo{})
	svc.SetNow(func() time.Time { return agora })

	motivoLongo := strings.Repeat("a", 501)
	nomeLongo := strings.Repeat("b", 201)
	consoleLongo := strings.Repeat("c", 101)
	dataFutura := agora.Add(24 * time.Hour)
	dataZero := time.Time{}
	tempoNegativo := int32(-1)

	tests := []struct {
		name        string
		input       SalvarJogoAbandonadoInput
		errEsperado error
	}{
		{
			name:        "nome_vazio",
			input:       SalvarJogoAbandonadoInput{Nome: "   ", Console: "SNES"},
			errEsperado: ErrAbandonadoNomeObrigatorio,
		},
		{
			name:        "nome_muito_longo",
			input:       SalvarJogoAbandonadoInput{Nome: nomeLongo, Console: "SNES"},
			errEsperado: ErrAbandonadoNomeMuitoLongo,
		},
		{
			name:        "console_vazio",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: "   "},
			errEsperado: ErrAbandonadoConsoleObrigatorio,
		},
		{
			name:        "console_muito_longo",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: consoleLongo},
			errEsperado: ErrAbandonadoConsoleMuitoLongo,
		},
		{
			name:        "tempo_horas_negativo",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: "SNES", TempoJogadoHoras: -1},
			errEsperado: ErrAbandonadoTempoInvalido,
		},
		{
			name:        "tempo_minutos_invalido",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: "SNES", TempoJogadoMinutos: 60},
			errEsperado: ErrAbandonadoTempoInvalido,
		},
		{
			name:        "tempo_segundos_invalido",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: "SNES", TempoJogadoSegundos: 60},
			errEsperado: ErrAbandonadoTempoInvalido,
		},
		{
			name:        "tempo_direto_negativo",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: "SNES", TempoJogado: &tempoNegativo},
			errEsperado: ErrAbandonadoTempoInvalido,
		},
		{
			name:        "motivo_muito_longo",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: "SNES", Motivo: &motivoLongo},
			errEsperado: ErrAbandonadoMotivoMuitoLongo,
		},
		{
			name:        "data_futura",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: "SNES", AbandonadoEm: &dataFutura},
			errEsperado: ErrAbandonadoDataFutura,
		},
		{
			name:        "data_invalida_zero",
			input:       SalvarJogoAbandonadoInput{Nome: "Jogo", Console: "SNES", AbandonadoEm: &dataZero},
			errEsperado: ErrAbandonadoDataInvalida,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CriarJogoAbandonado(context.Background(), tc.input)
			if !errors.Is(err, tc.errEsperado) {
				t.Fatalf("esperava erro %v, obteve %v", tc.errEsperado, err)
			}
			_, err = svc.AtualizarJogoAbandonado(context.Background(), tc.input)
			if !errors.Is(err, tc.errEsperado) {
				t.Fatalf("esperava erro %v na atualizacao, obteve %v", tc.errEsperado, err)
			}
		})
	}
}

func TestJogosAbandonadosService_Atualizar_Sucesso(t *testing.T) {
	repo := &mockAbandonadosRepo{
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoAbandonadoParams) (*repository.JogoAbandonado, error) {
			if params.ID != 5 || params.UsuarioID != 2 {
				t.Fatalf("parametros incorretos: ID=%d, UsuarioID=%d", params.ID, params.UsuarioID)
			}
			return &repository.JogoAbandonado{
				ID:        params.ID,
				UsuarioID: params.UsuarioID,
				Nome:      params.Nome,
			}, nil
		},
	}
	svc := NewJogosAbandonadosService(repo)
	jogo, err := svc.AtualizarJogoAbandonado(context.Background(), SalvarJogoAbandonadoInput{
		ID:        5,
		UsuarioID: 2,
		Nome:      "Bloodborne",
		Console:   "PS4",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if jogo.ID != 5 {
		t.Fatalf("esperava ID 5, obteve %d", jogo.ID)
	}
}

func TestJogosAbandonadosService_NaoEncontrado(t *testing.T) {
	repo := &mockAbandonadosRepo{
		obterPorIDFn: func(ctx context.Context, id int32, usuarioID int32) (*repository.JogoAbandonado, error) {
			return nil, pgx.ErrNoRows
		},
		atualizarFn: func(ctx context.Context, params repository.AtualizarJogoAbandonadoParams) (*repository.JogoAbandonado, error) {
			return nil, pgx.ErrNoRows
		},
		excluirFn: func(ctx context.Context, id int32, usuarioID int32) error {
			return pgx.ErrNoRows
		},
	}
	svc := NewJogosAbandonadosService(repo)

	_, err := svc.ObterJogoAbandonado(context.Background(), 99, 1)
	if !errors.Is(err, ErrAbandonadoNaoEncontrado) {
		t.Fatalf("esperava ErrAbandonadoNaoEncontrado em obter, obteve %v", err)
	}

	_, err = svc.AtualizarJogoAbandonado(context.Background(), SalvarJogoAbandonadoInput{
		ID:        99,
		UsuarioID: 1,
		Nome:      "Jogo",
		Console:   "Console",
	})
	if !errors.Is(err, ErrAbandonadoNaoEncontrado) {
		t.Fatalf("esperava ErrAbandonadoNaoEncontrado em atualizar, obteve %v", err)
	}

	err = svc.ExcluirJogoAbandonado(context.Background(), 99, 1)
	if !errors.Is(err, ErrAbandonadoNaoEncontrado) {
		t.Fatalf("esperava ErrAbandonadoNaoEncontrado em excluir, obteve %v", err)
	}
}

func TestJogosAbandonadosService_Listar_Filtros_Ordenacao(t *testing.T) {
	var paramsRecebidos repository.ListarJogosAbandonadosParams
	repo := &mockAbandonadosRepo{
		listarFn: func(ctx context.Context, params repository.ListarJogosAbandonadosParams) ([]*repository.JogoAbandonado, int64, error) {
			paramsRecebidos = params
			return []*repository.JogoAbandonado{{ID: 1}, {ID: 2}}, 25, nil
		},
	}
	svc := NewJogosAbandonadosService(repo)

	pagina := 2
	porPagina := 10
	res, err := svc.ListarJogosAbandonados(context.Background(), ListarJogosAbandonadosInput{
		UsuarioID: 7,
		Busca:     "zelda",
		Console:   "Switch",
		Ordenar:   "tempo",
		Pagina:    &pagina,
		PorPagina: &porPagina,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if res.Total != 25 || res.TotalPaginas != 3 || res.Pagina != 2 || res.PorPagina != 10 {
		t.Fatalf("resultado inesperado: %+v", res)
	}
	if paramsRecebidos.UsuarioID != 7 || paramsRecebidos.Busca != "zelda" || paramsRecebidos.Console != "Switch" ||
		paramsRecebidos.Ordenar != "tempo" || paramsRecebidos.Pagina != 2 || paramsRecebidos.PorPagina != 10 {
		t.Fatalf("parametros repassados ao repository incorretos: %+v", paramsRecebidos)
	}
}

func TestJogosAbandonadosService_Listar_ErrosValidacao(t *testing.T) {
	svc := NewJogosAbandonadosService(&mockAbandonadosRepo{})

	paginaInvalida := 0
	_, err := svc.ListarJogosAbandonados(context.Background(), ListarJogosAbandonadosInput{
		Pagina: &paginaInvalida,
	})
	if !errors.Is(err, ErrAbandonadoPaginaInvalida) {
		t.Fatalf("esperava ErrAbandonadoPaginaInvalida, obteve %v", err)
	}

	porPaginaInvalidaZero := 0
	_, err = svc.ListarJogosAbandonados(context.Background(), ListarJogosAbandonadosInput{
		PorPagina: &porPaginaInvalidaZero,
	})
	if !errors.Is(err, ErrAbandonadoPorPaginaInvalida) {
		t.Fatalf("esperava ErrAbandonadoPorPaginaInvalida, obteve %v", err)
	}

	porPaginaInvalidaMax := 101
	_, err = svc.ListarJogosAbandonados(context.Background(), ListarJogosAbandonadosInput{
		PorPagina: &porPaginaInvalidaMax,
	})
	if !errors.Is(err, ErrAbandonadoPorPaginaInvalida) {
		t.Fatalf("esperava ErrAbandonadoPorPaginaInvalida, obteve %v", err)
	}

	_, err = svc.ListarJogosAbandonados(context.Background(), ListarJogosAbandonadosInput{
		Ordenar: "invalido",
	})
	if !errors.Is(err, ErrAbandonadoOrdenarInvalido) {
		t.Fatalf("esperava ErrAbandonadoOrdenarInvalido, obteve %v", err)
	}
}

func TestJogosAbandonadosService_ConsolesETotal(t *testing.T) {
	repo := &mockAbandonadosRepo{
		obterConsolesFn: func(ctx context.Context, usuarioID int32) ([]string, error) {
			if usuarioID != 3 {
				t.Fatalf("esperava usuarioID 3, obteve %d", usuarioID)
			}
			return []string{"GBA", "SNES"}, nil
		},
		obterTotalFn: func(ctx context.Context, usuarioID int32) (int64, error) {
			if usuarioID != 3 {
				t.Fatalf("esperava usuarioID 3, obteve %d", usuarioID)
			}
			return 42, nil
		},
	}
	svc := NewJogosAbandonadosService(repo)

	consoles, err := svc.ObterConsoles(context.Background(), 3)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(consoles) != 2 || consoles[0] != "GBA" || consoles[1] != "SNES" {
		t.Fatalf("consoles inesperados: %v", consoles)
	}

	total, err := svc.ObterTotal(context.Background(), 3)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if total != 42 {
		t.Fatalf("esperava total 42, obteve %d", total)
	}
}

func TestJogosAbandonadosService_ErrosPostgres(t *testing.T) {
	repo := &mockAbandonadosRepo{
		criarFn: func(ctx context.Context, params repository.CriarJogoAbandonadoParams) (*repository.JogoAbandonado, error) {
			return nil, &pgconn.PgError{Code: "22001", ColumnName: "nome"}
		},
	}
	svc := NewJogosAbandonadosService(repo)
	_, err := svc.CriarJogoAbandonado(context.Background(), SalvarJogoAbandonadoInput{
		Nome:    "Jogo",
		Console: "SNES",
	})
	if !errors.Is(err, ErrAbandonadoNomeMuitoLongo) {
		t.Fatalf("esperava ErrAbandonadoNomeMuitoLongo, obteve %v", err)
	}

	repoCheck := &mockAbandonadosRepo{
		criarFn: func(ctx context.Context, params repository.CriarJogoAbandonadoParams) (*repository.JogoAbandonado, error) {
			return nil, &pgconn.PgError{Code: "23514", ConstraintName: "jogos_abandonados_tempo_jogado_check"}
		},
	}
	svcCheck := NewJogosAbandonadosService(repoCheck)
	_, err = svcCheck.CriarJogoAbandonado(context.Background(), SalvarJogoAbandonadoInput{
		Nome:    "Jogo",
		Console: "SNES",
	})
	if !errors.Is(err, ErrAbandonadoTempoInvalido) {
		t.Fatalf("esperava ErrAbandonadoTempoInvalido, obteve %v", err)
	}
}
