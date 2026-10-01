package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type mockDashboardService struct {
	obterResumoFn             func(ctx context.Context, usuarioID int32) (*service.ResumoResponse, error)
	listarPorAnoFn            func(ctx context.Context, usuarioID int32) ([]service.PorAnoItem, error)
	obterRankingPlataformasFn func(ctx context.Context, usuarioID int32, limite int) ([]service.RankingPlataformaItem, error)
	obterRankingGenerosFn     func(ctx context.Context, usuarioID int32, limite int) ([]service.RankingGeneroItem, error)
	obterBreakdownTipoFn      func(ctx context.Context, usuarioID int32, genero string) ([]service.BreakdownTipoItem, error)
	obterRecordesFn           func(ctx context.Context, usuarioID int32) (*service.RecordesResponse, error)
	obterNotasFn              func(ctx context.Context, usuarioID int32) (*service.NotasResponse, error)
	obterDificuldadeFn        func(ctx context.Context, usuarioID int32) ([]service.DificuldadeItem, error)
}

func (m *mockDashboardService) ObterResumo(ctx context.Context, usuarioID int32) (*service.ResumoResponse, error) {
	if m.obterResumoFn != nil {
		return m.obterResumoFn(ctx, usuarioID)
	}
	return nil, nil
}

func (m *mockDashboardService) ListarPorAno(ctx context.Context, usuarioID int32) ([]service.PorAnoItem, error) {
	if m.listarPorAnoFn != nil {
		return m.listarPorAnoFn(ctx, usuarioID)
	}
	return nil, nil
}

func (m *mockDashboardService) ObterRankingPlataformas(ctx context.Context, usuarioID int32, limite int) ([]service.RankingPlataformaItem, error) {
	if m.obterRankingPlataformasFn != nil {
		return m.obterRankingPlataformasFn(ctx, usuarioID, limite)
	}
	return nil, nil
}

func (m *mockDashboardService) ObterRankingGeneros(ctx context.Context, usuarioID int32, limite int) ([]service.RankingGeneroItem, error) {
	if m.obterRankingGenerosFn != nil {
		return m.obterRankingGenerosFn(ctx, usuarioID, limite)
	}
	return nil, nil
}

func (m *mockDashboardService) ObterBreakdownTipo(ctx context.Context, usuarioID int32, genero string) ([]service.BreakdownTipoItem, error) {
	if m.obterBreakdownTipoFn != nil {
		return m.obterBreakdownTipoFn(ctx, usuarioID, genero)
	}
	return nil, nil
}

func (m *mockDashboardService) ObterRecordes(ctx context.Context, usuarioID int32) (*service.RecordesResponse, error) {
	if m.obterRecordesFn != nil {
		return m.obterRecordesFn(ctx, usuarioID)
	}
	return nil, nil
}

func (m *mockDashboardService) ObterNotas(ctx context.Context, usuarioID int32) (*service.NotasResponse, error) {
	if m.obterNotasFn != nil {
		return m.obterNotasFn(ctx, usuarioID)
	}
	return nil, nil
}

func (m *mockDashboardService) ObterDificuldade(ctx context.Context, usuarioID int32) ([]service.DificuldadeItem, error) {
	if m.obterDificuldadeFn != nil {
		return m.obterDificuldadeFn(ctx, usuarioID)
	}
	return nil, nil
}

func setupDashboardTestRouter(svc DashboardServicer, secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	tokens, _ := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	h := NewDashboardHandler(svc)
	privadas.GET("/dashboard/resumo", h.ObterResumo)
	privadas.GET("/dashboard/por-ano", h.ListarPorAno)
	privadas.GET("/dashboard/ranking-plataformas", h.ObterRankingPlataformas)
	privadas.GET("/dashboard/ranking-generos", h.ObterRankingGeneros)
	privadas.GET("/dashboard/breakdown-tipo", h.ObterBreakdownTipo)
	privadas.GET("/dashboard/recordes", h.ObterRecordes)
	privadas.GET("/dashboard/notas", h.ObterNotas)
	privadas.GET("/dashboard/dificuldade", h.ObterDificuldade)
	return router
}

func generateDashboardToken(t *testing.T, secret, subject, lang string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: lang,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    "memory-card",
			Audience:  jwt.ClaimStrings{"access"},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestDashboardHandler_RotasSucesso200(t *testing.T) {
	secret := uuid.NewString()
	token := generateDashboardToken(t, secret, "10", "pt-BR")

	svc := &mockDashboardService{
		obterResumoFn: func(ctx context.Context, usuarioID int32) (*service.ResumoResponse, error) {
			return &service.ResumoResponse{
				TotalJogos:           152,
				TotalSegundos:        10584000,
				MediaSegundosPorJogo: 69630,
				NotaMedia:            8.2,
				JogosNoAnoAtual:      24,
			}, nil
		},
		listarPorAnoFn: func(ctx context.Context, usuarioID int32) ([]service.PorAnoItem, error) {
			return []service.PorAnoItem{
				{Ano: 2019, TotalJogos: 9, TotalSegundos: 1200000},
			}, nil
		},
		obterRankingPlataformasFn: func(ctx context.Context, usuarioID int32, limite int) ([]service.RankingPlataformaItem, error) {
			return []service.RankingPlataformaItem{
				{Console: "PlayStation 5", TotalJogos: 38, TotalSegundos: 2484000, PercentualJogos: 25.0, PercentualSegundos: 23.4},
			}, nil
		},
		obterRankingGenerosFn: func(ctx context.Context, usuarioID int32, limite int) ([]service.RankingGeneroItem, error) {
			return []service.RankingGeneroItem{
				{Genero: "RPG", TotalJogos: 38, TotalSegundos: 2484000, PercentualJogos: 25.0, PercentualSegundos: 23.4},
			}, nil
		},
		obterBreakdownTipoFn: func(ctx context.Context, usuarioID int32, genero string) ([]service.BreakdownTipoItem, error) {
			return []service.BreakdownTipoItem{
				{Tipo: "Principal", TotalJogos: 31},
			}, nil
		},
		obterRecordesFn: func(ctx context.Context, usuarioID int32) (*service.RecordesResponse, error) {
			return &service.RecordesResponse{
				MaisLongo: &service.RecordeJogoItem{ID: 1, Nome: "Persona 5 Royal"},
			}, nil
		},
		obterNotasFn: func(ctx context.Context, usuarioID int32) (*service.NotasResponse, error) {
			hist := make([]service.HistogramaItem, 11)
			for i := 1; i <= 11; i++ {
				hist[i-1] = service.HistogramaItem{Nota: i, Total: 0}
			}
			return &service.NotasResponse{Histograma: hist, NotaMedia: 8.2, TotalAvaliados: 152}, nil
		},
		obterDificuldadeFn: func(ctx context.Context, usuarioID int32) ([]service.DificuldadeItem, error) {
			return []service.DificuldadeItem{
				{Dificuldade: "C", TotalJogos: 22, Percentual: 14.5},
			}, nil
		},
	}

	router := setupDashboardTestRouter(svc, secret)

	endpoints := []string{
		"/api/v1/dashboard/resumo",
		"/api/v1/dashboard/por-ano",
		"/api/v1/dashboard/ranking-plataformas",
		"/api/v1/dashboard/ranking-generos",
		"/api/v1/dashboard/breakdown-tipo?genero=RPG",
		"/api/v1/dashboard/recordes",
		"/api/v1/dashboard/notas",
		"/api/v1/dashboard/dificuldade",
	}

	for _, ep := range endpoints {
		req, _ := http.NewRequest(http.MethodGet, ep, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("endpoint %s retornou %d, esperado 200", ep, w.Code)
		}
	}
}

func TestDashboardHandler_UsuarioSemJogosRetorna200(t *testing.T) {
	secret := uuid.NewString()
	token := generateDashboardToken(t, secret, "99", "pt-BR")

	svc := &mockDashboardService{
		obterResumoFn: func(ctx context.Context, usuarioID int32) (*service.ResumoResponse, error) {
			return &service.ResumoResponse{TotalJogos: 0, TotalSegundos: 0, MediaSegundosPorJogo: 0, NotaMedia: 0.0, JogosNoAnoAtual: 0}, nil
		},
		listarPorAnoFn: func(ctx context.Context, usuarioID int32) ([]service.PorAnoItem, error) {
			return []service.PorAnoItem{}, nil
		},
		obterRankingPlataformasFn: func(ctx context.Context, usuarioID int32, limite int) ([]service.RankingPlataformaItem, error) {
			return []service.RankingPlataformaItem{}, nil
		},
		obterRankingGenerosFn: func(ctx context.Context, usuarioID int32, limite int) ([]service.RankingGeneroItem, error) {
			return []service.RankingGeneroItem{}, nil
		},
		obterBreakdownTipoFn: func(ctx context.Context, usuarioID int32, genero string) ([]service.BreakdownTipoItem, error) {
			return []service.BreakdownTipoItem{}, nil
		},
		obterRecordesFn: func(ctx context.Context, usuarioID int32) (*service.RecordesResponse, error) {
			return &service.RecordesResponse{MaisLongo: nil, MaisCurto: nil}, nil
		},
		obterNotasFn: func(ctx context.Context, usuarioID int32) (*service.NotasResponse, error) {
			hist := make([]service.HistogramaItem, 11)
			for i := 1; i <= 11; i++ {
				hist[i-1] = service.HistogramaItem{Nota: i, Total: 0}
			}
			return &service.NotasResponse{Histograma: hist, NotaMedia: 0.0, TotalAvaliados: 0}, nil
		},
		obterDificuldadeFn: func(ctx context.Context, usuarioID int32) ([]service.DificuldadeItem, error) {
			ordem := []string{"C", "B", "A", "AA", "AAA"}
			itens := make([]service.DificuldadeItem, 5)
			for i, d := range ordem {
				itens[i] = service.DificuldadeItem{Dificuldade: d, TotalJogos: 0, Percentual: 0.0}
			}
			return itens, nil
		},
	}

	router := setupDashboardTestRouter(svc, secret)

	endpoints := []string{
		"/api/v1/dashboard/resumo",
		"/api/v1/dashboard/por-ano",
		"/api/v1/dashboard/ranking-plataformas",
		"/api/v1/dashboard/ranking-generos",
		"/api/v1/dashboard/breakdown-tipo?genero=RPG",
		"/api/v1/dashboard/recordes",
		"/api/v1/dashboard/notas",
		"/api/v1/dashboard/dificuldade",
	}

	for _, ep := range endpoints {
		req, _ := http.NewRequest(http.MethodGet, ep, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("endpoint %s sem jogos retornou %d, esperado 200", ep, w.Code)
		}
	}
}

func TestDashboardHandler_Validacoes400EI18n(t *testing.T) {
	secret := uuid.NewString()
	token := generateDashboardToken(t, secret, "10", "pt-BR")

	svc := &mockDashboardService{
		obterRankingPlataformasFn: func(ctx context.Context, usuarioID int32, limite int) ([]service.RankingPlataformaItem, error) {
			if limite < 1 || limite > 20 {
				return nil, service.ErrDashboardLimiteInvalido
			}
			return []service.RankingPlataformaItem{}, nil
		},
		obterRankingGenerosFn: func(ctx context.Context, usuarioID int32, limite int) ([]service.RankingGeneroItem, error) {
			if limite < 1 || limite > 20 {
				return nil, service.ErrDashboardLimiteInvalido
			}
			return []service.RankingGeneroItem{}, nil
		},
		obterBreakdownTipoFn: func(ctx context.Context, usuarioID int32, genero string) ([]service.BreakdownTipoItem, error) {
			if genero == "" {
				return nil, service.ErrDashboardGeneroObrigatorio
			}
			return []service.BreakdownTipoItem{}, nil
		},
	}

	router := setupDashboardTestRouter(svc, secret)

	t.Run("RankingPlataformas_LimiteInvalido_ptBR", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/ranking-plataformas?limite=abc", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "pt-BR")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperado 400, obteve %d", w.Code)
		}

		var resp map[string]map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"]["codigo"] != "dashboard.limite_invalido" {
			t.Errorf("codigo incorreto: %s", resp["error"]["codigo"])
		}
		if resp["error"]["mensagem"] != "Limite inválido. Informe um número inteiro entre 1 e 20." {
			t.Errorf("mensagem pt-BR incorreta: %s", resp["error"]["mensagem"])
		}
	})

	t.Run("RankingGeneros_LimiteInvalido_en", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/ranking-generos?limite=25", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperado 400, obteve %d", w.Code)
		}

		var resp map[string]map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"]["codigo"] != "dashboard.limite_invalido" {
			t.Errorf("codigo incorreto: %s", resp["error"]["codigo"])
		}
		if resp["error"]["mensagem"] != "Invalid limit. Provide an integer between 1 and 20." {
			t.Errorf("mensagem en incorreta: %s", resp["error"]["mensagem"])
		}
	})

	t.Run("BreakdownTipo_GeneroAusente_ptBR", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/breakdown-tipo", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "pt-BR")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperado 400, obteve %d", w.Code)
		}

		var resp map[string]map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"]["codigo"] != "dashboard.genero_obrigatorio" {
			t.Errorf("codigo incorreto: %s", resp["error"]["codigo"])
		}
		if resp["error"]["mensagem"] != "O gênero é obrigatório." {
			t.Errorf("mensagem pt-BR incorreta: %s", resp["error"]["mensagem"])
		}
	})

	t.Run("BreakdownTipo_GeneroVazio_en", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/breakdown-tipo?genero=   ", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperado 400, obteve %d", w.Code)
		}

		var resp map[string]map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["error"]["codigo"] != "dashboard.genero_obrigatorio" {
			t.Errorf("codigo incorreto: %s", resp["error"]["codigo"])
		}
		if resp["error"]["mensagem"] != "Genre is required." {
			t.Errorf("mensagem en incorreta: %s", resp["error"]["mensagem"])
		}
	})
}

func TestDashboardHandler_NaoAutenticado401(t *testing.T) {
	svc := &mockDashboardService{}
	router := setupDashboardTestRouter(svc, uuid.NewString())

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/resumo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obteve %d", w.Code)
	}

	var resp map[string]map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"]["codigo"] != "auth.session.unauthorized" {
		t.Errorf("codigo incorreto: %s", resp["error"]["codigo"])
	}
}
