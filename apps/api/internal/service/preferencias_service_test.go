package service

import (
	"context"
	"errors"
	"testing"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type preferenciasRepoMock struct {
	buscarPorUsuarioID    func(context.Context, int32) (*repository.PreferenciasUsuario, error)
	atualizarPreferencias func(context.Context, repository.AtualizarPreferenciasParams) (*repository.PreferenciasUsuario, error)
}

func (m preferenciasRepoMock) BuscarPorUsuarioID(ctx context.Context, usuarioID int32) (*repository.PreferenciasUsuario, error) {
	if m.buscarPorUsuarioID != nil {
		return m.buscarPorUsuarioID(ctx, usuarioID)
	}
	return nil, nil
}

func (m preferenciasRepoMock) AtualizarPreferencias(ctx context.Context, params repository.AtualizarPreferenciasParams) (*repository.PreferenciasUsuario, error) {
	if m.atualizarPreferencias != nil {
		return m.atualizarPreferencias(ctx, params)
	}
	return nil, nil
}

func helperStringPtr(s string) *string {
	return &s
}

func helperIntPtr(i int) *int {
	return &i
}

func helperBoolPtr(b bool) *bool {
	return &b
}

func payloadValidoPreferencias() AtualizarPreferenciasInput {
	return AtualizarPreferenciasInput{
		Idioma:           helperStringPtr("pt-BR"),
		FormatoData:      helperStringPtr("dmy"),
		FusoHorario:      helperStringPtr("America/Sao_Paulo"),
		VisualBiblioteca: helperStringPtr("grade"),
		ItensPorPagina:   helperIntPtr(24),
		ReduzirAnimacoes: helperBoolPtr(false),
		ModoTema:         helperStringPtr("escuro"),
		EstiloTema:       helperStringPtr("padrao"),
	}
}

func TestPreferenciasService_ObterPreferencias_DefaultsQuandoNaoHaRegistro(t *testing.T) {
	ctx := context.Background()
	usuarioID := int32(42)

	defaults := &repository.PreferenciasUsuario{
		Idioma:           "pt-BR",
		FormatoData:      DefaultFormatoData,
		FusoHorario:      DefaultFusoHorario,
		VisualBiblioteca: DefaultVisualBiblioteca,
		ItensPorPagina:   DefaultItensPorPagina,
		ReduzirAnimacoes: DefaultReduzirAnimacoes,
		ModoTema:         DefaultModoTema,
		EstiloTema:       DefaultEstiloTema,
	}

	called := false
	mock := preferenciasRepoMock{
		buscarPorUsuarioID: func(gotCtx context.Context, id int32) (*repository.PreferenciasUsuario, error) {
			called = true
			if gotCtx != ctx || id != usuarioID {
				t.Fatalf("parametros incorretos: got %d, want %d", id, usuarioID)
			}
			return defaults, nil
		},
	}

	svc := NewPreferenciasService(mock)
	res, err := svc.ObterPreferencias(ctx, "42")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !called {
		t.Fatal("repositorio nao foi chamado")
	}
	if res.Idioma != "pt-BR" ||
		res.FormatoData != "dmy" ||
		res.FusoHorario != "America/Sao_Paulo" ||
		res.VisualBiblioteca != "grade" ||
		res.ItensPorPagina != 24 ||
		res.ReduzirAnimacoes != false ||
		res.ModoTema != "escuro" ||
		res.EstiloTema != "padrao" {
		t.Fatalf("resultado com defaults incorretos: %+v", res)
	}
}

func TestPreferenciasService_AtualizarPreferencias_SucessoEParametros(t *testing.T) {
	ctx := context.Background()
	usuarioID := int32(42)

	input := AtualizarPreferenciasInput{
		Idioma:           helperStringPtr("en"),
		FormatoData:      helperStringPtr("ymd"),
		FusoHorario:      helperStringPtr("Europe/Lisbon"),
		VisualBiblioteca: helperStringPtr("lista"),
		ItensPorPagina:   helperIntPtr(50),
		ReduzirAnimacoes: helperBoolPtr(true),
		ModoTema:         helperStringPtr("claro"),
		EstiloTema:       helperStringPtr("playstation"),
	}

	esperadoRetorno := &repository.PreferenciasUsuario{
		Idioma:           "en",
		FormatoData:      "ymd",
		FusoHorario:      "Europe/Lisbon",
		VisualBiblioteca: "lista",
		ItensPorPagina:   50,
		ReduzirAnimacoes: true,
		ModoTema:         "claro",
		EstiloTema:       "playstation",
	}

	called := false
	mock := preferenciasRepoMock{
		atualizarPreferencias: func(gotCtx context.Context, params repository.AtualizarPreferenciasParams) (*repository.PreferenciasUsuario, error) {
			called = true
			if params.UsuarioID != usuarioID {
				t.Fatalf("usuario_id incorreto: got %d, want %d", params.UsuarioID, usuarioID)
			}
			if params.Idioma != "en" ||
				params.FormatoData != "ymd" ||
				params.FusoHorario != "Europe/Lisbon" ||
				params.VisualBiblioteca != "lista" ||
				params.ItensPorPagina != 50 ||
				params.ReduzirAnimacoes != true ||
				params.ModoTema != "claro" ||
				params.EstiloTema != "playstation" {
				t.Fatalf("parametros repassados incorretos: %+v", params)
			}
			return esperadoRetorno, nil
		},
	}

	svc := NewPreferenciasService(mock)
	res, err := svc.AtualizarPreferencias(ctx, "42", input)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !called {
		t.Fatal("repositorio nao foi chamado")
	}
	if res != esperadoRetorno {
		t.Fatalf("retorno inesperado: %+v", res)
	}
}

func TestPreferenciasService_AtualizarPreferencias_CamposInvalidos(t *testing.T) {
	testCases := []struct {
		nome    string
		ajuste  func(*AtualizarPreferenciasInput)
		wantErr error
	}{
		{
			nome:    "idioma invalido",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.Idioma = helperStringPtr("fr") },
			wantErr: ErrPreferenciasIdiomaInvalido,
		},
		{
			nome:    "formato_data invalido",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.FormatoData = helperStringPtr("invalid") },
			wantErr: ErrPreferenciasFormatoDataInvalido,
		},
		{
			nome:    "fuso_horario Foo/Bar",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.FusoHorario = helperStringPtr("Foo/Bar") },
			wantErr: ErrPreferenciasFusoHorarioInvalido,
		},
		{
			nome:    "fuso_horario Local",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.FusoHorario = helperStringPtr("Local") },
			wantErr: ErrPreferenciasFusoHorarioInvalido,
		},
		{
			nome:    "fuso_horario vazio",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.FusoHorario = helperStringPtr("") },
			wantErr: ErrPreferenciasFusoHorarioInvalido,
		},
		{
			nome:    "fuso_horario somente espacos",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.FusoHorario = helperStringPtr("   ") },
			wantErr: ErrPreferenciasFusoHorarioInvalido,
		},
		{
			nome:    "visual_biblioteca invalido",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.VisualBiblioteca = helperStringPtr("tabela") },
			wantErr: ErrPreferenciasVisualBibliotecaInvalido,
		},
		{
			nome:    "itens_por_pagina = 10",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.ItensPorPagina = helperIntPtr(10) },
			wantErr: ErrPreferenciasItensPorPaginaInvalido,
		},
		{
			nome:    "itens_por_pagina = 0",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.ItensPorPagina = helperIntPtr(0) },
			wantErr: ErrPreferenciasItensPorPaginaInvalido,
		},
		{
			nome:    "modo_tema invalido",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.ModoTema = helperStringPtr("neon") },
			wantErr: ErrPreferenciasModoTemaInvalido,
		},
		{
			nome:    "estilo_tema invalido",
			ajuste:  func(in *AtualizarPreferenciasInput) { in.EstiloTema = helperStringPtr("sega") },
			wantErr: ErrPreferenciasEstiloTemaInvalido,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.nome, func(t *testing.T) {
			input := payloadValidoPreferencias()
			tc.ajuste(&input)

			mock := preferenciasRepoMock{
				atualizarPreferencias: func(context.Context, repository.AtualizarPreferenciasParams) (*repository.PreferenciasUsuario, error) {
					t.Fatal("repositorio nao deveria ser chamado para input invalido")
					return nil, nil
				},
			}

			svc := NewPreferenciasService(mock)
			_, err := svc.AtualizarPreferencias(context.Background(), "42", input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("erro = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestPreferenciasService_AtualizarPreferencias_CamposAusentes(t *testing.T) {
	testCases := []struct {
		nome   string
		ajuste func(*AtualizarPreferenciasInput)
	}{
		{"idioma ausente", func(in *AtualizarPreferenciasInput) { in.Idioma = nil }},
		{"formato_data ausente", func(in *AtualizarPreferenciasInput) { in.FormatoData = nil }},
		{"fuso_horario ausente", func(in *AtualizarPreferenciasInput) { in.FusoHorario = nil }},
		{"visual_biblioteca ausente", func(in *AtualizarPreferenciasInput) { in.VisualBiblioteca = nil }},
		{"itens_por_pagina ausente", func(in *AtualizarPreferenciasInput) { in.ItensPorPagina = nil }},
		{"reduzir_animacoes ausente", func(in *AtualizarPreferenciasInput) { in.ReduzirAnimacoes = nil }},
		{"modo_tema ausente", func(in *AtualizarPreferenciasInput) { in.ModoTema = nil }},
		{"estilo_tema ausente", func(in *AtualizarPreferenciasInput) { in.EstiloTema = nil }},
	}

	for _, tc := range testCases {
		t.Run(tc.nome, func(t *testing.T) {
			input := payloadValidoPreferencias()
			tc.ajuste(&input)

			mock := preferenciasRepoMock{
				atualizarPreferencias: func(context.Context, repository.AtualizarPreferenciasParams) (*repository.PreferenciasUsuario, error) {
					t.Fatal("repositorio nao deveria ser chamado para campo ausente")
					return nil, nil
				},
			}

			svc := NewPreferenciasService(mock)
			_, err := svc.AtualizarPreferencias(context.Background(), "42", input)
			if !errors.Is(err, ErrPreferenciasEntradaInvalida) {
				t.Fatalf("erro = %v, want %v", err, ErrPreferenciasEntradaInvalida)
			}
		})
	}
}

func TestPreferenciasService_IdentidadeInvalida(t *testing.T) {
	svc := NewPreferenciasService(preferenciasRepoMock{
		buscarPorUsuarioID: func(context.Context, int32) (*repository.PreferenciasUsuario, error) {
			t.Fatal("nao deveria consultar repositorio")
			return nil, nil
		},
	})

	for _, subject := range []string{"", "abc", "0", "-1", "2147483648"} {
		t.Run("obter "+subject, func(t *testing.T) {
			res, err := svc.ObterPreferencias(context.Background(), subject)
			if res != nil || !errors.Is(err, ErrTokenInvalido) {
				t.Fatalf("erro = %v, want %v", err, ErrTokenInvalido)
			}
		})

		t.Run("atualizar "+subject, func(t *testing.T) {
			res, err := svc.AtualizarPreferencias(context.Background(), subject, payloadValidoPreferencias())
			if res != nil || !errors.Is(err, ErrTokenInvalido) {
				t.Fatalf("erro = %v, want %v", err, ErrTokenInvalido)
			}
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.ObterPreferencias(ctx, "42"); !errors.Is(err, context.Canceled) {
		t.Fatalf("esperava context.Canceled, obteve %v", err)
	}
	if _, err := svc.AtualizarPreferencias(ctx, "42", payloadValidoPreferencias()); !errors.Is(err, context.Canceled) {
		t.Fatalf("esperava context.Canceled, obteve %v", err)
	}
}

func TestPreferenciasService_ErroRepositorioPropagado(t *testing.T) {
	dbErr := errors.New("db connection failure")

	mock := preferenciasRepoMock{
		buscarPorUsuarioID: func(context.Context, int32) (*repository.PreferenciasUsuario, error) {
			return nil, dbErr
		},
		atualizarPreferencias: func(context.Context, repository.AtualizarPreferenciasParams) (*repository.PreferenciasUsuario, error) {
			return nil, dbErr
		},
	}

	svc := NewPreferenciasService(mock)

	_, errObter := svc.ObterPreferencias(context.Background(), "42")
	if !errors.Is(errObter, dbErr) {
		t.Fatalf("erro de obter nao propagou erro do repo: %v", errObter)
	}

	_, errAtualizar := svc.AtualizarPreferencias(context.Background(), "42", payloadValidoPreferencias())
	if !errors.Is(errAtualizar, dbErr) {
		t.Fatalf("erro de atualizar nao propagou erro do repo: %v", errAtualizar)
	}
}

func TestPreferenciasService_UsuarioNaoEncontradoNoRepo(t *testing.T) {
	mock := preferenciasRepoMock{
		buscarPorUsuarioID: func(context.Context, int32) (*repository.PreferenciasUsuario, error) {
			return nil, nil
		},
		atualizarPreferencias: func(context.Context, repository.AtualizarPreferenciasParams) (*repository.PreferenciasUsuario, error) {
			return nil, nil
		},
	}

	svc := NewPreferenciasService(mock)

	_, errObter := svc.ObterPreferencias(context.Background(), "42")
	if !errors.Is(errObter, ErrTokenInvalido) {
		t.Fatalf("esperava ErrTokenInvalido ao obter usuario inexistente, obteve %v", errObter)
	}

	_, errAtualizar := svc.AtualizarPreferencias(context.Background(), "42", payloadValidoPreferencias())
	if !errors.Is(errAtualizar, ErrTokenInvalido) {
		t.Fatalf("esperava ErrTokenInvalido ao atualizar usuario inexistente, obteve %v", errAtualizar)
	}
}
