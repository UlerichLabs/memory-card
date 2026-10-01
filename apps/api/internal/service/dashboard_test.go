package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type mockDashboardRepo struct {
	obterResumoFn                 func(ctx context.Context, usuarioID int32, anoAtual int32) (*repository.ResumoDados, error)
	listarEstatisticasPorAnoFn    func(ctx context.Context, usuarioID int32) ([]*repository.EstatisticaAnoDados, error)
	obterRankingPlataformasFn     func(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error)
	obterRankingGenerosFn         func(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error)
	obterBreakdownTipoFn          func(ctx context.Context, usuarioID int32, genero string) ([]*repository.BreakdownTipoDados, error)
	obterRecordeMaisLongoFn       func(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error)
	obterRecordeMaisCurtoFn       func(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error)
	obterDistribuicaoNotasFn      func(ctx context.Context, usuarioID int32) ([]*repository.NotaDistribuicaoDados, error)
	obterDistribuicaoDificuldadeFn func(ctx context.Context, usuarioID int32) ([]*repository.DificuldadeDistribuicaoDados, error)
}

func (m *mockDashboardRepo) ObterResumo(ctx context.Context, usuarioID int32, anoAtual int32) (*repository.ResumoDados, error) {
	if m.obterResumoFn != nil {
		return m.obterResumoFn(ctx, usuarioID, anoAtual)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockDashboardRepo) ListarEstatisticasPorAno(ctx context.Context, usuarioID int32) ([]*repository.EstatisticaAnoDados, error) {
	if m.listarEstatisticasPorAnoFn != nil {
		return m.listarEstatisticasPorAnoFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockDashboardRepo) ObterRankingPlataformas(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error) {
	if m.obterRankingPlataformasFn != nil {
		return m.obterRankingPlataformasFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockDashboardRepo) ObterRankingGeneros(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error) {
	if m.obterRankingGenerosFn != nil {
		return m.obterRankingGenerosFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockDashboardRepo) ObterBreakdownTipo(ctx context.Context, usuarioID int32, genero string) ([]*repository.BreakdownTipoDados, error) {
	if m.obterBreakdownTipoFn != nil {
		return m.obterBreakdownTipoFn(ctx, usuarioID, genero)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockDashboardRepo) ObterRecordeMaisLongo(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error) {
	if m.obterRecordeMaisLongoFn != nil {
		return m.obterRecordeMaisLongoFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockDashboardRepo) ObterRecordeMaisCurto(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error) {
	if m.obterRecordeMaisCurtoFn != nil {
		return m.obterRecordeMaisCurtoFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockDashboardRepo) ObterDistribuicaoNotas(ctx context.Context, usuarioID int32) ([]*repository.NotaDistribuicaoDados, error) {
	if m.obterDistribuicaoNotasFn != nil {
		return m.obterDistribuicaoNotasFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func (m *mockDashboardRepo) ObterDistribuicaoDificuldade(ctx context.Context, usuarioID int32) ([]*repository.DificuldadeDistribuicaoDados, error) {
	if m.obterDistribuicaoDificuldadeFn != nil {
		return m.obterDistribuicaoDificuldadeFn(ctx, usuarioID)
	}
	return nil, errors.New("nao implementado")
}

func TestFuncoesPuras_Calculos(t *testing.T) {
	t.Run("CalcularPercentual_TotalZero", func(t *testing.T) {
		got := CalcularPercentual(10, 0)
		if got != 0.0 {
			t.Fatalf("esperado 0.0, obteve %f", got)
		}
	})

	t.Run("CalcularPercentual_ValorZero", func(t *testing.T) {
		got := CalcularPercentual(0, 100)
		if got != 0.0 {
			t.Fatalf("esperado 0.0, obteve %f", got)
		}
	})

	t.Run("CalcularPercentual_ArredondamentoUmaCasa", func(t *testing.T) {
		got := CalcularPercentual(38, 152)
		if got != 25.0 {
			t.Fatalf("esperado 25.0, obteve %f", got)
		}

		got2 := CalcularPercentual(1, 3)
		if got2 != 33.3 {
			t.Fatalf("esperado 33.3, obteve %f", got2)
		}

		got3 := CalcularPercentual(2, 3)
		if got3 != 66.7 {
			t.Fatalf("esperado 66.7, obteve %f", got3)
		}
	})

	t.Run("CalcularMedia_TotalZero", func(t *testing.T) {
		got := CalcularMedia(100, 0)
		if got != 0.0 {
			t.Fatalf("esperado 0.0, obteve %f", got)
		}
	})

	t.Run("CalcularMedia_ArredondamentoUmaCasa", func(t *testing.T) {
		got := CalcularMedia(1246, 152)
		if got != 8.2 {
			t.Fatalf("esperado 8.2, obteve %f", got)
		}
	})

	t.Run("CalcularMediaSegundos", func(t *testing.T) {
		got := CalcularMediaSegundos(10584000, 152)
		if got != 69632 {
			t.Fatalf("esperado 69632, obteve %d", got)
		}

		gotZero := CalcularMediaSegundos(0, 0)
		if gotZero != 0 {
			t.Fatalf("esperado 0, obteve %d", gotZero)
		}
	})

	t.Run("CalcularDiasEAnosDesde", func(t *testing.T) {
		primeiro := time.Date(2019, 3, 14, 0, 0, 0, 0, time.UTC)
		hoje := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

		dias := CalcularDiasDesde(primeiro, hoje)
		if dias != 2757 {
			t.Fatalf("esperado 2757 dias, obteve %d", dias)
		}

		anos := CalcularAnosDesde(primeiro, hoje)
		if anos != 7 {
			t.Fatalf("esperado 7 anos, obteve %d", anos)
		}

		hojeAntesDoAniversario := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
		anosAntes := CalcularAnosDesde(primeiro, hojeAntesDoAniversario)
		if anosAntes != 6 {
			t.Fatalf("esperado 6 anos, obteve %d", anosAntes)
		}
	})
}

func TestDashboardService_ObterResumo(t *testing.T) {
	dataBase := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

	t.Run("SucessoComJogos", func(t *testing.T) {
		primeiro := time.Date(2019, 3, 14, 0, 0, 0, 0, time.UTC)
		repoChamado := false

		repo := &mockDashboardRepo{
			obterResumoFn: func(ctx context.Context, usuarioID int32, anoAtual int32) (*repository.ResumoDados, error) {
				repoChamado = true
				if usuarioID != 42 {
					t.Errorf("usuarioID incorreto: %d", usuarioID)
				}
				if anoAtual != 2026 {
					t.Errorf("anoAtual incorreto: %d", anoAtual)
				}
				return &repository.ResumoDados{
					TotalJogos:          152,
					TotalSegundos:       10584000,
					SomaNotas:           1246,
					TotalAvaliados:      152,
					JogosNoAnoAtual:     24,
					PrimeiroZeramentoEm: &primeiro,
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		svc.SetNow(func() time.Time { return dataBase })

		res, err := svc.ObterResumo(context.Background(), 42)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if !repoChamado {
			t.Fatal("repo nao foi chamado")
		}
		if res.TotalJogos != 152 {
			t.Errorf("total jogos: %d", res.TotalJogos)
		}
		if res.TotalSegundos != 10584000 {
			t.Errorf("total segundos: %d", res.TotalSegundos)
		}
		if res.MediaSegundosPorJogo != 69632 {
			t.Errorf("media segundos: %d", res.MediaSegundosPorJogo)
		}
		if res.NotaMedia != 8.2 {
			t.Errorf("nota media: %f", res.NotaMedia)
		}
		if res.JogosNoAnoAtual != 24 {
			t.Errorf("jogos ano atual: %d", res.JogosNoAnoAtual)
		}
		if res.DiasDesdePrimeiro != 2757 {
			t.Errorf("dias desde primeiro: %d", res.DiasDesdePrimeiro)
		}
		if res.AnosDesdePrimeiro != 7 {
			t.Errorf("anos desde primeiro: %d", res.AnosDesdePrimeiro)
		}
	})

	t.Run("SemJogosRetornaZeros", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterResumoFn: func(ctx context.Context, usuarioID int32, anoAtual int32) (*repository.ResumoDados, error) {
				return &repository.ResumoDados{
					TotalJogos:          0,
					TotalSegundos:       0,
					SomaNotas:           0,
					TotalAvaliados:      0,
					JogosNoAnoAtual:     0,
					PrimeiroZeramentoEm: nil,
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		svc.SetNow(func() time.Time { return dataBase })

		res, err := svc.ObterResumo(context.Background(), 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if res.TotalJogos != 0 || res.TotalSegundos != 0 || res.MediaSegundosPorJogo != 0 || res.NotaMedia != 0.0 {
			t.Fatalf("esperado tudo zero, obteve %+v", res)
		}
		if res.PrimeiroZeramentoEm != nil || res.DiasDesdePrimeiro != 0 || res.AnosDesdePrimeiro != 0 {
			t.Fatalf("esperado primeiro zeramento nil e 0 dias/anos, obteve %+v", res)
		}
	})
}

func TestDashboardService_ListarPorAno(t *testing.T) {
	dataBase := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

	t.Run("SucessoPreencheAnosVaziosEOrdena", func(t *testing.T) {
		notaGda := int32(11)
		idGda := int32(10)

		repo := &mockDashboardRepo{
			listarEstatisticasPorAnoFn: func(ctx context.Context, usuarioID int32) ([]*repository.EstatisticaAnoDados, error) {
				if usuarioID != 7 {
					t.Errorf("usuarioID incorreto: %d", usuarioID)
				}
				return []*repository.EstatisticaAnoDados{
					{
						Ano:                 2022,
						TotalJogos:          9,
						TotalSegundos:       1200000,
						DestaqueID:          &idGda,
						DestaqueNome:        "Hollow Knight",
						DestaqueConsole:     "PC",
						DestaqueIgdbCapaURL: "https://capa.png",
						DestaqueNota:        &notaGda,
					},
					{
						Ano:           2025,
						TotalJogos:    3,
						TotalSegundos: 500000,
					},
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		svc.SetNow(func() time.Time { return dataBase })

		itens, err := svc.ListarPorAno(context.Background(), 7)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if len(itens) != 5 {
			t.Fatalf("esperado 5 anos (2022 a 2026), obteve %d", len(itens))
		}

		if itens[0].Ano != 2022 || itens[0].TotalJogos != 9 || itens[0].GameDoAno == nil || itens[0].GameDoAno.Nome != "Hollow Knight" {
			t.Errorf("ano 2022 incorreto: %+v", itens[0])
		}

		if itens[1].Ano != 2023 || itens[1].TotalJogos != 0 || itens[1].TotalSegundos != 0 || itens[1].GameDoAno != nil {
			t.Errorf("ano 2023 vazio incorreto: %+v", itens[1])
		}

		if itens[2].Ano != 2024 || itens[2].TotalJogos != 0 || itens[2].TotalSegundos != 0 || itens[2].GameDoAno != nil {
			t.Errorf("ano 2024 vazio incorreto: %+v", itens[2])
		}

		if itens[3].Ano != 2025 || itens[3].TotalJogos != 3 || itens[3].TotalSegundos != 500000 || itens[3].GameDoAno != nil {
			t.Errorf("ano 2025 incorreto: %+v", itens[3])
		}

		if itens[4].Ano != 2026 || itens[4].TotalJogos != 0 || itens[4].TotalSegundos != 0 || itens[4].GameDoAno != nil {
			t.Errorf("ano 2026 corrente vazio incorreto: %+v", itens[4])
		}
	})

	t.Run("SemJogosRetornaListaVazia", func(t *testing.T) {
		repo := &mockDashboardRepo{
			listarEstatisticasPorAnoFn: func(ctx context.Context, usuarioID int32) ([]*repository.EstatisticaAnoDados, error) {
				return []*repository.EstatisticaAnoDados{}, nil
			},
		}

		svc := NewDashboardService(repo)
		itens, err := svc.ListarPorAno(context.Background(), 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if len(itens) != 0 {
			t.Fatalf("esperado lista vazia, obteve %d itens", len(itens))
		}
	})
}

func TestDashboardService_RankingPlataformas(t *testing.T) {
	t.Run("SucessoCalculaPercentualSobreTotalETruncaLimite", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterRankingPlataformasFn: func(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error) {
				if usuarioID != 15 {
					t.Errorf("usuarioID incorreto: %d", usuarioID)
				}
				return []*repository.RankingItemDados{
					{Nome: "PlayStation 5", TotalJogos: 50, TotalSegundos: 5000},
					{Nome: "PC", TotalJogos: 30, TotalSegundos: 3000},
					{Nome: "Nintendo Switch", TotalJogos: 20, TotalSegundos: 2000},
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		itens, err := svc.ObterRankingPlataformas(context.Background(), 15, 2)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if len(itens) != 2 {
			t.Fatalf("esperado 2 itens com limite=2, obteve %d", len(itens))
		}

		if itens[0].Console != "PlayStation 5" || itens[0].TotalJogos != 50 || itens[0].PercentualJogos != 50.0 || itens[0].PercentualSegundos != 50.0 {
			t.Errorf("item 0 incorreto: %+v", itens[0])
		}

		if itens[1].Console != "PC" || itens[1].TotalJogos != 30 || itens[1].PercentualJogos != 30.0 || itens[1].PercentualSegundos != 30.0 {
			t.Errorf("item 1 incorreto: %+v", itens[1])
		}
	})

	t.Run("LimiteInvalido", func(t *testing.T) {
		svc := NewDashboardService(&mockDashboardRepo{})

		_, err0 := svc.ObterRankingPlataformas(context.Background(), 1, 0)
		if !errors.Is(err0, ErrDashboardLimiteInvalido) {
			t.Errorf("esperado ErrDashboardLimiteInvalido para 0, obteve %v", err0)
		}

		_, err21 := svc.ObterRankingPlataformas(context.Background(), 1, 21)
		if !errors.Is(err21, ErrDashboardLimiteInvalido) {
			t.Errorf("esperado ErrDashboardLimiteInvalido para 21, obteve %v", err21)
		}
	})

	t.Run("SemJogosRetornaListaVazia", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterRankingPlataformasFn: func(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error) {
				return []*repository.RankingItemDados{}, nil
			},
		}

		svc := NewDashboardService(repo)
		itens, err := svc.ObterRankingPlataformas(context.Background(), 1, 5)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if len(itens) != 0 {
			t.Fatalf("esperado lista vazia, obteve %d", len(itens))
		}
	})
}

func TestDashboardService_RankingGeneros(t *testing.T) {
	t.Run("SucessoCalculaPercentualSobreTotalETruncaLimite", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterRankingGenerosFn: func(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error) {
				return []*repository.RankingItemDados{
					{Nome: "RPG", TotalJogos: 40, TotalSegundos: 4000},
					{Nome: "Ação", TotalJogos: 10, TotalSegundos: 1000},
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		itens, err := svc.ObterRankingGeneros(context.Background(), 1, 5)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if len(itens) != 2 {
			t.Fatalf("esperado 2 itens, obteve %d", len(itens))
		}

		if itens[0].Genero != "RPG" || itens[0].PercentualJogos != 80.0 {
			t.Errorf("item 0 incorreto: %+v", itens[0])
		}
	})

	t.Run("LimiteInvalido", func(t *testing.T) {
		svc := NewDashboardService(&mockDashboardRepo{})
		_, err := svc.ObterRankingGeneros(context.Background(), 1, -1)
		if !errors.Is(err, ErrDashboardLimiteInvalido) {
			t.Errorf("esperado ErrDashboardLimiteInvalido, obteve %v", err)
		}
	})
}

func TestDashboardService_BreakdownTipo(t *testing.T) {
	t.Run("SucessoRetornaTiposOrdenados", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterBreakdownTipoFn: func(ctx context.Context, usuarioID int32, genero string) ([]*repository.BreakdownTipoDados, error) {
				if genero != "RPG" {
					t.Errorf("genero inesperado: %s", genero)
				}
				return []*repository.BreakdownTipoDados{
					{Tipo: "Principal", TotalJogos: 31},
					{Tipo: "Completo", TotalJogos: 5},
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		itens, err := svc.ObterBreakdownTipo(context.Background(), 1, "RPG")
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if len(itens) != 2 || itens[0].Tipo != "Principal" || itens[0].TotalJogos != 31 {
			t.Errorf("resultado inesperado: %+v", itens)
		}
	})

	t.Run("GeneroVazioRetornaErro", func(t *testing.T) {
		svc := NewDashboardService(&mockDashboardRepo{})
		_, err := svc.ObterBreakdownTipo(context.Background(), 1, "   ")
		if !errors.Is(err, ErrDashboardGeneroObrigatorio) {
			t.Errorf("esperado ErrDashboardGeneroObrigatorio, obteve %v", err)
		}
	})

	t.Run("SemJogosRetornaListaVazia", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterBreakdownTipoFn: func(ctx context.Context, usuarioID int32, genero string) ([]*repository.BreakdownTipoDados, error) {
				return []*repository.BreakdownTipoDados{}, nil
			},
		}

		svc := NewDashboardService(repo)
		itens, err := svc.ObterBreakdownTipo(context.Background(), 1, "Corrida")
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if len(itens) != 0 {
			t.Fatalf("esperado lista vazia, obteve %d", len(itens))
		}
	})
}

func TestDashboardService_Recordes(t *testing.T) {
	t.Run("SucessoComMaisLongoEMaisCurto", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterRecordeMaisLongoFn: func(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error) {
				return &repository.RecordeJogoDados{
					ID:                  1,
					Nome:                "Persona 5 Royal",
					Console:             "PlayStation 4",
					Ano:                 2022,
					IgdbCapaURL:         "https://p5r.png",
					TempoJogadoSegundos: 511200,
				}, nil
			},
			obterRecordeMaisCurtoFn: func(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error) {
				return &repository.RecordeJogoDados{
					ID:                  2,
					Nome:                "A Short Hike",
					Console:             "PC",
					Ano:                 2020,
					IgdbCapaURL:         "https://ashort.png",
					TempoJogadoSegundos: 7200,
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		rec, err := svc.ObterRecordes(context.Background(), 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if rec.MaisLongo == nil || rec.MaisLongo.Nome != "Persona 5 Royal" || rec.MaisLongo.TempoJogadoSegundos != 511200 {
			t.Errorf("mais longo incorreto: %+v", rec.MaisLongo)
		}
		if rec.MaisCurto == nil || rec.MaisCurto.Nome != "A Short Hike" || rec.MaisCurto.TempoJogadoSegundos != 7200 {
			t.Errorf("mais curto incorreto: %+v", rec.MaisCurto)
		}
	})

	t.Run("SemElegiveisRetornaCamposNull", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterRecordeMaisLongoFn: func(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error) {
				return nil, nil
			},
			obterRecordeMaisCurtoFn: func(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error) {
				return nil, nil
			},
		}

		svc := NewDashboardService(repo)
		rec, err := svc.ObterRecordes(context.Background(), 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if rec.MaisLongo != nil || rec.MaisCurto != nil {
			t.Fatalf("esperado ambos null, obteve %+v", rec)
		}
	})
}

func TestDashboardService_Notas(t *testing.T) {
	t.Run("SucessoHistogramaSempre11Posicoes", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterDistribuicaoNotasFn: func(ctx context.Context, usuarioID int32) ([]*repository.NotaDistribuicaoDados, error) {
				return []*repository.NotaDistribuicaoDados{
					{Nota: 8, Total: 10},
					{Nota: 11, Total: 5},
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		resp, err := svc.ObterNotas(context.Background(), 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if len(resp.Histograma) != 11 {
			t.Fatalf("esperado exatamente 11 posicoes no histograma, obteve %d", len(resp.Histograma))
		}

		for i, h := range resp.Histograma {
			expectedNota := i + 1
			if h.Nota != expectedNota {
				t.Errorf("posicao %d: nota esperada %d, obteve %d", i, expectedNota, h.Nota)
			}
			if expectedNota == 8 && h.Total != 10 {
				t.Errorf("nota 8: total esperado 10, obteve %d", h.Total)
			} else if expectedNota == 11 && h.Total != 5 {
				t.Errorf("nota 11: total esperado 5, obteve %d", h.Total)
			} else if expectedNota != 8 && expectedNota != 11 && h.Total != 0 {
				t.Errorf("nota %d: esperado 0, obteve %d", expectedNota, h.Total)
			}
		}

		if resp.TotalAvaliados != 15 {
			t.Errorf("total avaliados: esperado 15, obteve %d", resp.TotalAvaliados)
		}

		expectedMedia := CalcularMedia(8*10+11*5, 15)
		if resp.NotaMedia != expectedMedia {
			t.Errorf("nota media: esperado %f, obteve %f", expectedMedia, resp.NotaMedia)
		}
	})

	t.Run("SemJogosAvaliadosRetornaZerosE11Posicoes", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterDistribuicaoNotasFn: func(ctx context.Context, usuarioID int32) ([]*repository.NotaDistribuicaoDados, error) {
				return []*repository.NotaDistribuicaoDados{}, nil
			},
		}

		svc := NewDashboardService(repo)
		resp, err := svc.ObterNotas(context.Background(), 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if len(resp.Histograma) != 11 {
			t.Fatalf("esperado 11 posicoes, obteve %d", len(resp.Histograma))
		}
		if resp.NotaMedia != 0.0 || resp.TotalAvaliados != 0 {
			t.Errorf("esperado nota media 0 e total avaliados 0, obteve %+v", resp)
		}
	})
}

func TestDashboardService_Dificuldade(t *testing.T) {
	t.Run("SucessoSempre5ItensOrdemFixaEPercentuais", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterDistribuicaoDificuldadeFn: func(ctx context.Context, usuarioID int32) ([]*repository.DificuldadeDistribuicaoDados, error) {
				return []*repository.DificuldadeDistribuicaoDados{
					{Dificuldade: "A", TotalJogos: 50},
					{Dificuldade: "C", TotalJogos: 50},
				}, nil
			},
		}

		svc := NewDashboardService(repo)
		itens, err := svc.ObterDificuldade(context.Background(), 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if len(itens) != 5 {
			t.Fatalf("esperado 5 itens, obteve %d", len(itens))
		}

		ordemEsperada := []string{"C", "B", "A", "AA", "AAA"}
		for i, dif := range ordemEsperada {
			if itens[i].Dificuldade != dif {
				t.Errorf("posicao %d: esperado %s, obteve %s", i, dif, itens[i].Dificuldade)
			}
		}

		if itens[0].TotalJogos != 50 || itens[0].Percentual != 50.0 {
			t.Errorf("C incorreto: %+v", itens[0])
		}
		if itens[1].TotalJogos != 0 || itens[1].Percentual != 0.0 {
			t.Errorf("B incorreto: %+v", itens[1])
		}
		if itens[2].TotalJogos != 50 || itens[2].Percentual != 50.0 {
			t.Errorf("A incorreto: %+v", itens[2])
		}
		if itens[3].TotalJogos != 0 || itens[3].Percentual != 0.0 {
			t.Errorf("AA incorreto: %+v", itens[3])
		}
		if itens[4].TotalJogos != 0 || itens[4].Percentual != 0.0 {
			t.Errorf("AAA incorreto: %+v", itens[4])
		}
	})

	t.Run("SemJogosRetorna5ItensZerados", func(t *testing.T) {
		repo := &mockDashboardRepo{
			obterDistribuicaoDificuldadeFn: func(ctx context.Context, usuarioID int32) ([]*repository.DificuldadeDistribuicaoDados, error) {
				return []*repository.DificuldadeDistribuicaoDados{}, nil
			},
		}

		svc := NewDashboardService(repo)
		itens, err := svc.ObterDificuldade(context.Background(), 1)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}

		if len(itens) != 5 {
			t.Fatalf("esperado 5 itens, obteve %d", len(itens))
		}
		for _, item := range itens {
			if item.TotalJogos != 0 || item.Percentual != 0.0 {
				t.Errorf("item nao zerado: %+v", item)
			}
		}
	})
}
