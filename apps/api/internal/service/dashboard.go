// Package service contem as regras de negocio da aplicacao.
package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrDashboardLimiteInvalido    = errors.New("dashboard.limite_invalido")
	ErrDashboardGeneroObrigatorio = errors.New("dashboard.genero_obrigatorio")
)

type ResumoResponse struct {
	TotalJogos           int64      `json:"total_jogos"`
	TotalSegundos        int64      `json:"total_segundos"`
	MediaSegundosPorJogo int64      `json:"media_segundos_por_jogo"`
	NotaMedia            float64    `json:"nota_media"`
	JogosNoAnoAtual      int64      `json:"jogos_no_ano_atual"`
	PrimeiroZeramentoEm  *time.Time `json:"primeiro_zeramento_em"`
	DiasDesdePrimeiro    int        `json:"dias_desde_primeiro"`
	AnosDesdePrimeiro    int        `json:"anos_desde_primeiro"`
}

type GameDoAnoItem struct {
	ID          int32  `json:"id"`
	Nome        string `json:"nome"`
	Console     string `json:"console"`
	IgdbCapaURL string `json:"igdb_capa_url"`
	Nota        int32  `json:"nota"`
}

type PorAnoItem struct {
	Ano           int            `json:"ano"`
	TotalJogos    int64          `json:"total_jogos"`
	TotalSegundos int64          `json:"total_segundos"`
	GameDoAno     *GameDoAnoItem `json:"game_do_ano"`
}

type RankingPlataformaItem struct {
	Console            string  `json:"console"`
	TotalJogos         int64   `json:"total_jogos"`
	TotalSegundos      int64   `json:"total_segundos"`
	PercentualJogos    float64 `json:"percentual_jogos"`
	PercentualSegundos float64 `json:"percentual_segundos"`
}

type RankingGeneroItem struct {
	Genero             string  `json:"genero"`
	TotalJogos         int64   `json:"total_jogos"`
	TotalSegundos      int64   `json:"total_segundos"`
	PercentualJogos    float64 `json:"percentual_jogos"`
	PercentualSegundos float64 `json:"percentual_segundos"`
}

type BreakdownTipoItem struct {
	Tipo       string `json:"tipo"`
	TotalJogos int64  `json:"total_jogos"`
}

type RecordeJogoItem struct {
	ID                  int32  `json:"id"`
	Nome                string `json:"nome"`
	Console             string `json:"console"`
	Ano                 int    `json:"ano"`
	IgdbCapaURL         string `json:"igdb_capa_url"`
	TempoJogadoSegundos int32  `json:"tempo_jogado_segundos"`
}

type RecordesResponse struct {
	MaisLongo *RecordeJogoItem `json:"mais_longo"`
	MaisCurto *RecordeJogoItem `json:"mais_curto"`
}

type HistogramaItem struct {
	Nota  int   `json:"nota"`
	Total int64 `json:"total"`
}

type NotasResponse struct {
	Histograma     []HistogramaItem `json:"histograma"`
	NotaMedia      float64          `json:"nota_media"`
	TotalAvaliados int64            `json:"total_avaliados"`
}

type DificuldadeItem struct {
	Dificuldade string  `json:"dificuldade"`
	TotalJogos  int64   `json:"total_jogos"`
	Percentual  float64 `json:"percentual"`
}

type DashboardRepository interface {
	ObterResumo(ctx context.Context, usuarioID int32, anoAtual int32) (*repository.ResumoDados, error)
	ListarEstatisticasPorAno(ctx context.Context, usuarioID int32) ([]*repository.EstatisticaAnoDados, error)
	ObterRankingPlataformas(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error)
	ObterRankingGeneros(ctx context.Context, usuarioID int32) ([]*repository.RankingItemDados, error)
	ObterBreakdownTipo(ctx context.Context, usuarioID int32, genero string) ([]*repository.BreakdownTipoDados, error)
	ObterRecordeMaisLongo(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error)
	ObterRecordeMaisCurto(ctx context.Context, usuarioID int32) (*repository.RecordeJogoDados, error)
	ObterDistribuicaoNotas(ctx context.Context, usuarioID int32) ([]*repository.NotaDistribuicaoDados, error)
	ObterDistribuicaoDificuldade(ctx context.Context, usuarioID int32) ([]*repository.DificuldadeDistribuicaoDados, error)
}

type DashboardService struct {
	repo DashboardRepository
	now  func() time.Time
}

func NewDashboardService(repo DashboardRepository) *DashboardService {
	return &DashboardService{
		repo: repo,
		now:  time.Now,
	}
}

func (s *DashboardService) SetNow(nowFn func() time.Time) {
	s.now = nowFn
}

func (s *DashboardService) getSaoPauloLocation() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("America/Sao_Paulo", -3*3600)
	}
	return loc
}

func CalcularPercentual(valor, total int64) float64 {
	if total <= 0 || valor <= 0 {
		return 0.0
	}
	p := (float64(valor) / float64(total)) * 100
	return math.Round(p*10) / 10
}

func CalcularMedia(soma, total int64) float64 {
	if total <= 0 {
		return 0.0
	}
	m := float64(soma) / float64(total)
	return math.Round(m*10) / 10
}

func CalcularMediaSegundos(totalSegundos, totalJogos int64) int64 {
	if totalJogos <= 0 || totalSegundos <= 0 {
		return 0
	}
	return int64(math.Round(float64(totalSegundos) / float64(totalJogos)))
}

func CalcularDiasDesde(primeiro, hoje time.Time) int {
	p := time.Date(primeiro.Year(), primeiro.Month(), primeiro.Day(), 0, 0, 0, 0, time.UTC)
	h := time.Date(hoje.Year(), hoje.Month(), hoje.Day(), 0, 0, 0, 0, time.UTC)
	diff := h.Sub(p)
	dias := int(math.Round(diff.Hours() / 24))
	if dias < 0 {
		return 0
	}
	return dias
}

func CalcularAnosDesde(primeiro, hoje time.Time) int {
	anos := hoje.Year() - primeiro.Year()
	if hoje.Month() < primeiro.Month() || (hoje.Month() == primeiro.Month() && hoje.Day() < primeiro.Day()) {
		anos--
	}
	if anos < 0 {
		return 0
	}
	return anos
}

func (s *DashboardService) ObterResumo(ctx context.Context, usuarioID int32) (*ResumoResponse, error) {
	loc := s.getSaoPauloLocation()
	nowSP := s.now().In(loc)
	anoAtual := int32(nowSP.Year())

	dados, err := s.repo.ObterResumo(ctx, usuarioID, anoAtual)
	if err != nil {
		return nil, fmt.Errorf("obter resumo dashboard: %w", err)
	}

	if dados == nil || dados.TotalJogos == 0 || dados.PrimeiroZeramentoEm == nil {
		return &ResumoResponse{
			TotalJogos:           0,
			TotalSegundos:        0,
			MediaSegundosPorJogo: 0,
			NotaMedia:            0.0,
			JogosNoAnoAtual:      0,
			PrimeiroZeramentoEm:  nil,
			DiasDesdePrimeiro:    0,
			AnosDesdePrimeiro:    0,
		}, nil
	}

	dias := CalcularDiasDesde(*dados.PrimeiroZeramentoEm, nowSP)
	anos := CalcularAnosDesde(*dados.PrimeiroZeramentoEm, nowSP)

	return &ResumoResponse{
		TotalJogos:           dados.TotalJogos,
		TotalSegundos:        dados.TotalSegundos,
		MediaSegundosPorJogo: CalcularMediaSegundos(dados.TotalSegundos, dados.TotalJogos),
		NotaMedia:            CalcularMedia(dados.SomaNotas, dados.TotalAvaliados),
		JogosNoAnoAtual:      dados.JogosNoAnoAtual,
		PrimeiroZeramentoEm:  dados.PrimeiroZeramentoEm,
		DiasDesdePrimeiro:    dias,
		AnosDesdePrimeiro:    anos,
	}, nil
}

func (s *DashboardService) ListarPorAno(ctx context.Context, usuarioID int32) ([]PorAnoItem, error) {
	rows, err := s.repo.ListarEstatisticasPorAno(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar por ano: %w", err)
	}

	if len(rows) == 0 {
		return []PorAnoItem{}, nil
	}

	loc := s.getSaoPauloLocation()
	anoAtual := s.now().In(loc).Year()

	minAno := rows[0].Ano
	maxAno := anoAtual
	for _, r := range rows {
		if r.Ano > maxAno {
			maxAno = r.Ano
		}
	}

	mapa := make(map[int]*repository.EstatisticaAnoDados, len(rows))
	for _, r := range rows {
		mapa[r.Ano] = r
	}

	resultado := make([]PorAnoItem, 0, maxAno-minAno+1)
	for ano := minAno; ano <= maxAno; ano++ {
		if d, ok := mapa[ano]; ok {
			var gda *GameDoAnoItem
			if d.DestaqueID != nil && d.DestaqueNota != nil {
				gda = &GameDoAnoItem{
					ID:          *d.DestaqueID,
					Nome:        d.DestaqueNome,
					Console:     d.DestaqueConsole,
					IgdbCapaURL: d.DestaqueIgdbCapaURL,
					Nota:        *d.DestaqueNota,
				}
			}
			resultado = append(resultado, PorAnoItem{
				Ano:           ano,
				TotalJogos:    d.TotalJogos,
				TotalSegundos: d.TotalSegundos,
				GameDoAno:     gda,
			})
		} else {
			resultado = append(resultado, PorAnoItem{
				Ano:           ano,
				TotalJogos:    0,
				TotalSegundos: 0,
				GameDoAno:     nil,
			})
		}
	}

	return resultado, nil
}

func (s *DashboardService) ObterRankingPlataformas(ctx context.Context, usuarioID int32, limite int) ([]RankingPlataformaItem, error) {
	if limite < 1 || limite > 20 {
		return nil, ErrDashboardLimiteInvalido
	}

	rows, err := s.repo.ObterRankingPlataformas(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter ranking plataformas: %w", err)
	}

	if len(rows) == 0 {
		return []RankingPlataformaItem{}, nil
	}

	var totalJogosGeral int64
	var totalSegundosGeral int64
	for _, r := range rows {
		totalJogosGeral += r.TotalJogos
		totalSegundosGeral += r.TotalSegundos
	}

	n := limite
	if len(rows) < n {
		n = len(rows)
	}

	resultado := make([]RankingPlataformaItem, 0, n)
	for i := 0; i < n; i++ {
		r := rows[i]
		resultado = append(resultado, RankingPlataformaItem{
			Console:            r.Nome,
			TotalJogos:         r.TotalJogos,
			TotalSegundos:      r.TotalSegundos,
			PercentualJogos:    CalcularPercentual(r.TotalJogos, totalJogosGeral),
			PercentualSegundos: CalcularPercentual(r.TotalSegundos, totalSegundosGeral),
		})
	}

	return resultado, nil
}

func (s *DashboardService) ObterRankingGeneros(ctx context.Context, usuarioID int32, limite int) ([]RankingGeneroItem, error) {
	if limite < 1 || limite > 20 {
		return nil, ErrDashboardLimiteInvalido
	}

	rows, err := s.repo.ObterRankingGeneros(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter ranking generos: %w", err)
	}

	if len(rows) == 0 {
		return []RankingGeneroItem{}, nil
	}

	var totalJogosGeral int64
	var totalSegundosGeral int64
	for _, r := range rows {
		totalJogosGeral += r.TotalJogos
		totalSegundosGeral += r.TotalSegundos
	}

	n := limite
	if len(rows) < n {
		n = len(rows)
	}

	resultado := make([]RankingGeneroItem, 0, n)
	for i := 0; i < n; i++ {
		r := rows[i]
		resultado = append(resultado, RankingGeneroItem{
			Genero:             r.Nome,
			TotalJogos:         r.TotalJogos,
			TotalSegundos:      r.TotalSegundos,
			PercentualJogos:    CalcularPercentual(r.TotalJogos, totalJogosGeral),
			PercentualSegundos: CalcularPercentual(r.TotalSegundos, totalSegundosGeral),
		})
	}

	return resultado, nil
}

func (s *DashboardService) ObterBreakdownTipo(ctx context.Context, usuarioID int32, genero string) ([]BreakdownTipoItem, error) {
	genero = strings.TrimSpace(genero)
	if genero == "" {
		return nil, ErrDashboardGeneroObrigatorio
	}

	rows, err := s.repo.ObterBreakdownTipo(ctx, usuarioID, genero)
	if err != nil {
		return nil, fmt.Errorf("obter breakdown tipo: %w", err)
	}

	if len(rows) == 0 {
		return []BreakdownTipoItem{}, nil
	}

	resultado := make([]BreakdownTipoItem, 0, len(rows))
	for _, r := range rows {
		resultado = append(resultado, BreakdownTipoItem{
			Tipo:       r.Tipo,
			TotalJogos: r.TotalJogos,
		})
	}

	return resultado, nil
}

func (s *DashboardService) ObterRecordes(ctx context.Context, usuarioID int32) (*RecordesResponse, error) {
	maisLongo, err := s.repo.ObterRecordeMaisLongo(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter recorde mais longo: %w", err)
	}

	maisCurto, err := s.repo.ObterRecordeMaisCurto(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter recorde mais curto: %w", err)
	}

	resp := &RecordesResponse{}
	if maisLongo != nil {
		resp.MaisLongo = &RecordeJogoItem{
			ID:                  maisLongo.ID,
			Nome:                maisLongo.Nome,
			Console:             maisLongo.Console,
			Ano:                 maisLongo.Ano,
			IgdbCapaURL:         maisLongo.IgdbCapaURL,
			TempoJogadoSegundos: maisLongo.TempoJogadoSegundos,
		}
	}
	if maisCurto != nil {
		resp.MaisCurto = &RecordeJogoItem{
			ID:                  maisCurto.ID,
			Nome:                maisCurto.Nome,
			Console:             maisCurto.Console,
			Ano:                 maisCurto.Ano,
			IgdbCapaURL:         maisCurto.IgdbCapaURL,
			TempoJogadoSegundos: maisCurto.TempoJogadoSegundos,
		}
	}

	return resp, nil
}

func (s *DashboardService) ObterNotas(ctx context.Context, usuarioID int32) (*NotasResponse, error) {
	rows, err := s.repo.ObterDistribuicaoNotas(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter notas: %w", err)
	}

	histograma := make([]HistogramaItem, 11)
	for i := 1; i <= 11; i++ {
		histograma[i-1] = HistogramaItem{Nota: i, Total: 0}
	}

	var totalAvaliados int64
	var somaNotas int64
	for _, r := range rows {
		if r.Nota >= 1 && r.Nota <= 11 {
			histograma[r.Nota-1].Total = r.Total
			totalAvaliados += r.Total
			somaNotas += int64(r.Nota) * r.Total
		}
	}

	return &NotasResponse{
		Histograma:     histograma,
		NotaMedia:      CalcularMedia(somaNotas, totalAvaliados),
		TotalAvaliados: totalAvaliados,
	}, nil
}

func (s *DashboardService) ObterDificuldade(ctx context.Context, usuarioID int32) ([]DificuldadeItem, error) {
	rows, err := s.repo.ObterDistribuicaoDificuldade(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obter dificuldade: %w", err)
	}

	ordem := []string{"C", "B", "A", "AA", "AAA"}
	mapa := make(map[string]int64, len(rows))
	var totalGeral int64
	for _, r := range rows {
		mapa[r.Dificuldade] = r.TotalJogos
		totalGeral += r.TotalJogos
	}

	resultado := make([]DificuldadeItem, 0, 5)
	for _, dif := range ordem {
		total := mapa[dif]
		resultado = append(resultado, DificuldadeItem{
			Dificuldade: dif,
			TotalJogos:  total,
			Percentual:  CalcularPercentual(total, totalGeral),
		})
	}

	return resultado, nil
}
