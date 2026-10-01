//go:build integration

// Package integration contem os testes de integracao com banco real.
package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/UlerichLabs/memory-card/apps/api/internal/handler"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/migration"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

func init() {
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

type dashboardIntegrationEnv struct {
	pool    *pgxpool.Pool
	sqlDB   *sql.DB
	router  *gin.Engine
	tokens  *service.AuthToken
	secret  string
	connStr string
}

func setupDashboardIntegrationEnv(t *testing.T) *dashboardIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_dashboard_test"),
		tcpostgres.WithUsername("test_user"),
		tcpostgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("falha ao iniciar container postgres de integracao: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao obter connection string: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao interpretar config do pool: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao criar pool pgx: %v", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)

	if _, err := migration.Apply(pool); err != nil {
		pool.Close()
		sqlDB.Close()
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao aplicar migrations: %v", err)
	}

	t.Cleanup(func() {
		sqlDB.Close()
		pool.Close()
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer teardownCancel()
		if err := pgContainer.Terminate(teardownCtx); err != nil {
			t.Logf("aviso: falha ao terminar container: %v", err)
		}
	})

	queries := db.New(pool)
	dashRepo := repository.NewDashboardRepository(queries)
	dashSvc := service.NewDashboardService(dashRepo)
	dashH := handler.NewDashboardHandler(dashSvc)

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	privadas.GET("/dashboard/resumo", dashH.ObterResumo)
	privadas.GET("/dashboard/por-ano", dashH.ListarPorAno)
	privadas.GET("/dashboard/ranking-plataformas", dashH.ObterRankingPlataformas)
	privadas.GET("/dashboard/ranking-generos", dashH.ObterRankingGeneros)
	privadas.GET("/dashboard/breakdown-tipo", dashH.ObterBreakdownTipo)
	privadas.GET("/dashboard/recordes", dashH.ObterRecordes)
	privadas.GET("/dashboard/notas", dashH.ObterNotas)
	privadas.GET("/dashboard/dificuldade", dashH.ObterDificuldade)

	return &dashboardIntegrationEnv{
		pool:    pool,
		sqlDB:   sqlDB,
		router:  router,
		tokens:  tokens,
		secret:  secret,
		connStr: connStr,
	}
}

func criarUsuarioDashboard(t *testing.T, pool *pgxpool.Pool) int32 {
	t.Helper()
	uid := uuid.NewString()
	email := fmt.Sprintf("dash-%s@example.com", uid)
	nome := fmt.Sprintf("Dash %s", uid[:8])

	var id int32
	err := pool.QueryRow(context.Background(),
		"INSERT INTO usuarios (nome, email, senha_hash, username) VALUES ($1, $2, $3, $4) RETURNING id",
		nome, email, "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01", uid[:15],
	).Scan(&id)
	if err != nil {
		t.Fatalf("falha ao criar usuario de teste: %v", err)
	}
	return id
}

func gerarTokenDashboard(t *testing.T, secret string, usuarioID int32) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: "pt-BR",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(usuarioID)),
			Issuer:    "memory-card",
			Audience:  jwt.ClaimStrings{"access"},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

type jogoTesteParams struct {
	UsuarioID    int32
	Nome         string
	Console      string
	Genero       *string
	Tipo         *string
	FinalizadoEm time.Time
	TempoJogado  int32
	Nota         int32
	Dificuldade  string
	Destaque     bool
	IgdbCapaURL  *string
	DeletedAt    *time.Time
}

func inserirJogoZeradoTeste(t *testing.T, pool *pgxpool.Pool, p jogoTesteParams) int32 {
	t.Helper()
	var id int32
	err := pool.QueryRow(context.Background(), `
		INSERT INTO jogos_zerados (
			usuario_id, nome, console, genero, tipo, finalizado_em,
			tempo_jogado, nota, dificuldade, destaque, igdb_capa_url, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`,
		p.UsuarioID, p.Nome, p.Console, p.Genero, p.Tipo, p.FinalizadoEm,
		p.TempoJogado, p.Nota, p.Dificuldade, p.Destaque, p.IgdbCapaURL, p.DeletedAt,
	).Scan(&id)
	if err != nil {
		t.Fatalf("falha ao inserir jogo zerado de teste: %v", err)
	}
	return id
}

func strPtr(s string) *string {
	return &s
}

func TestIntegration_Dashboard(t *testing.T) {
	env := setupDashboardIntegrationEnv(t)

	t.Run("IsolamentoEntreUsuariosESoftDelete", func(t *testing.T) {
		userA := criarUsuarioDashboard(t, env.pool)
		tokenA := gerarTokenDashboard(t, env.secret, userA)

		userB := criarUsuarioDashboard(t, env.pool)

		dataFinalizacaoA := time.Date(2023, 5, 10, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    userA,
			Nome:         "Jogo A Ativo",
			Console:      "PlayStation 5",
			Genero:       strPtr("RPG"),
			Tipo:         strPtr("Principal"),
			FinalizadoEm: dataFinalizacaoA,
			TempoJogado:  3600,
			Nota:         9,
			Dificuldade:  "A",
		})

		dataDelete := time.Now()
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    userA,
			Nome:         "Jogo A Deletado",
			Console:      "Xbox Series",
			Genero:       strPtr("Aventura"),
			Tipo:         strPtr("Extra"),
			FinalizadoEm: dataFinalizacaoA,
			TempoJogado:  7200,
			Nota:         10,
			Dificuldade:  "AAA",
			DeletedAt:    &dataDelete,
		})

		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    userB,
			Nome:         "Jogo B",
			Console:      "Nintendo Switch",
			Genero:       strPtr("Plataforma"),
			Tipo:         strPtr("Completo"),
			FinalizadoEm: dataFinalizacaoA,
			TempoJogado:  50000,
			Nota:         11,
			Dificuldade:  "C",
		})

		reqResumo, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/resumo", nil)
		reqResumo.Header.Set("Authorization", "Bearer "+tokenA)
		wResumo := httptest.NewRecorder()
		env.router.ServeHTTP(wResumo, reqResumo)
		if wResumo.Code != http.StatusOK {
			t.Fatalf("status resumo=%d", wResumo.Code)
		}
		var resumo service.ResumoResponse
		_ = json.Unmarshal(wResumo.Body.Bytes(), &resumo)
		if resumo.TotalJogos != 1 {
			t.Errorf("esperado 1 jogo para user A, obteve %d", resumo.TotalJogos)
		}
		if resumo.TotalSegundos != 3600 {
			t.Errorf("esperado 3600s para user A, obteve %d", resumo.TotalSegundos)
		}

		reqRankPlat, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/ranking-plataformas", nil)
		reqRankPlat.Header.Set("Authorization", "Bearer "+tokenA)
		wRankPlat := httptest.NewRecorder()
		env.router.ServeHTTP(wRankPlat, reqRankPlat)
		var rankPlat []service.RankingPlataformaItem
		_ = json.Unmarshal(wRankPlat.Body.Bytes(), &rankPlat)
		if len(rankPlat) != 1 || rankPlat[0].Console != "PlayStation 5" {
			t.Errorf("ranking plataformas deve conter apenas PS5 do user A: %+v", rankPlat)
		}

		reqRankGen, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/ranking-generos", nil)
		reqRankGen.Header.Set("Authorization", "Bearer "+tokenA)
		wRankGen := httptest.NewRecorder()
		env.router.ServeHTTP(wRankGen, reqRankGen)
		var rankGen []service.RankingGeneroItem
		_ = json.Unmarshal(wRankGen.Body.Bytes(), &rankGen)
		if len(rankGen) != 1 || rankGen[0].Genero != "RPG" {
			t.Errorf("ranking generos deve conter apenas RPG do user A: %+v", rankGen)
		}

		reqBreakdown, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/breakdown-tipo?genero=RPG", nil)
		reqBreakdown.Header.Set("Authorization", "Bearer "+tokenA)
		wBreakdown := httptest.NewRecorder()
		env.router.ServeHTTP(wBreakdown, reqBreakdown)
		var breakdown []service.BreakdownTipoItem
		_ = json.Unmarshal(wBreakdown.Body.Bytes(), &breakdown)
		if len(breakdown) != 1 || breakdown[0].Tipo != "Principal" {
			t.Errorf("breakdown tipo deve conter apenas Principal do user A: %+v", breakdown)
		}

		reqRec, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/recordes", nil)
		reqRec.Header.Set("Authorization", "Bearer "+tokenA)
		wRec := httptest.NewRecorder()
		env.router.ServeHTTP(wRec, reqRec)
		var rec service.RecordesResponse
		_ = json.Unmarshal(wRec.Body.Bytes(), &rec)
		if rec.MaisLongo == nil || rec.MaisLongo.Nome != "Jogo A Ativo" {
			t.Errorf("mais longo deve ser Jogo A Ativo: %+v", rec.MaisLongo)
		}
		if rec.MaisCurto == nil || rec.MaisCurto.Nome != "Jogo A Ativo" {
			t.Errorf("mais curto deve ser Jogo A Ativo: %+v", rec.MaisCurto)
		}
	})

	t.Run("ConsoleGeneroTipoNulosOuVaziosIgnorados", func(t *testing.T) {
		user := criarUsuarioDashboard(t, env.pool)
		token := gerarTokenDashboard(t, env.secret, user)

		data := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Jogo Sem Genero E Tipo",
			Console:      "PC",
			Genero:       nil,
			Tipo:         nil,
			FinalizadoEm: data,
			TempoJogado:  1000,
			Nota:         8,
			Dificuldade:  "B",
		})

		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Jogo Genero Vazio",
			Console:      "PC",
			Genero:       strPtr(""),
			Tipo:         strPtr(""),
			FinalizadoEm: data,
			TempoJogado:  2000,
			Nota:         7,
			Dificuldade:  "B",
		})

		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Jogo Console Vazio",
			Console:      "",
			Genero:       strPtr("Acao"),
			Tipo:         strPtr("Principal"),
			FinalizadoEm: data,
			TempoJogado:  3000,
			Nota:         9,
			Dificuldade:  "A",
		})

		reqPlat, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/ranking-plataformas", nil)
		reqPlat.Header.Set("Authorization", "Bearer "+token)
		wPlat := httptest.NewRecorder()
		env.router.ServeHTTP(wPlat, reqPlat)
		var rankPlat []service.RankingPlataformaItem
		_ = json.Unmarshal(wPlat.Body.Bytes(), &rankPlat)
		if len(rankPlat) != 1 || rankPlat[0].Console != "PC" {
			t.Errorf("esperado apenas PC no ranking plataformas, obteve: %+v", rankPlat)
		}

		reqGen, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/ranking-generos", nil)
		reqGen.Header.Set("Authorization", "Bearer "+token)
		wGen := httptest.NewRecorder()
		env.router.ServeHTTP(wGen, reqGen)
		var rankGen []service.RankingGeneroItem
		_ = json.Unmarshal(wGen.Body.Bytes(), &rankGen)
		if len(rankGen) != 1 || rankGen[0].Genero != "Acao" {
			t.Errorf("esperado apenas Acao no ranking generos, obteve: %+v", rankGen)
		}
	})

	t.Run("PercentuaisSobreTotalComLimiteMenorQueGrupos", func(t *testing.T) {
		user := criarUsuarioDashboard(t, env.pool)
		token := gerarTokenDashboard(t, env.secret, user)

		data := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 5; i++ {
			inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
				UsuarioID:    user,
				Nome:         fmt.Sprintf("SNES %d", i),
				Console:      "SNES",
				Genero:       strPtr("RPG"),
				FinalizadoEm: data,
				TempoJogado:  200,
				Nota:         8,
				Dificuldade:  "B",
			})
		}
		for i := 0; i < 3; i++ {
			inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
				UsuarioID:    user,
				Nome:         fmt.Sprintf("PS5 %d", i),
				Console:      "PS5",
				Genero:       strPtr("Acao"),
				FinalizadoEm: data,
				TempoJogado:  200,
				Nota:         8,
				Dificuldade:  "B",
			})
		}
		for i := 0; i < 2; i++ {
			inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
				UsuarioID:    user,
				Nome:         fmt.Sprintf("PC %d", i),
				Console:      "PC",
				Genero:       strPtr("Aventura"),
				FinalizadoEm: data,
				TempoJogado:  200,
				Nota:         8,
				Dificuldade:  "B",
			})
		}

		reqPlat, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/ranking-plataformas?limite=1", nil)
		reqPlat.Header.Set("Authorization", "Bearer "+token)
		wPlat := httptest.NewRecorder()
		env.router.ServeHTTP(wPlat, reqPlat)
		var rankPlat []service.RankingPlataformaItem
		_ = json.Unmarshal(wPlat.Body.Bytes(), &rankPlat)

		if len(rankPlat) != 1 {
			t.Fatalf("esperado 1 item, obteve %d", len(rankPlat))
		}
		if rankPlat[0].Console != "SNES" || rankPlat[0].TotalJogos != 5 || rankPlat[0].PercentualJogos != 50.0 {
			t.Errorf("SNES com percentual incorreto: %+v", rankPlat[0])
		}
	})

	t.Run("PorAnoComLacunasEGameDoAno", func(t *testing.T) {
		user := criarUsuarioDashboard(t, env.pool)
		token := gerarTokenDashboard(t, env.secret, user)

		data2020 := time.Date(2020, 4, 15, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "GOTY 2020",
			Console:      "PC",
			FinalizadoEm: data2020,
			TempoJogado:  10000,
			Nota:         11,
			Dificuldade:  "AAA",
			Destaque:     true,
			IgdbCapaURL:  strPtr("https://capa2020.jpg"),
		})

		data2023 := time.Date(2023, 8, 20, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Jogo 2023 Comum",
			Console:      "PS5",
			FinalizadoEm: data2023,
			TempoJogado:  5000,
			Nota:         8,
			Dificuldade:  "A",
			Destaque:     false,
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/por-ano", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status por-ano=%d", w.Code)
		}

		var anos []service.PorAnoItem
		_ = json.Unmarshal(w.Body.Bytes(), &anos)

		if len(anos) < 4 {
			t.Fatalf("esperava ao menos 4 anos de 2020 a 2023+, obteve %d", len(anos))
		}

		if anos[0].Ano != 2020 || anos[0].TotalJogos != 1 || anos[0].GameDoAno == nil || anos[0].GameDoAno.Nome != "GOTY 2020" {
			t.Errorf("ano 2020 incorreto: %+v", anos[0])
		}

		var ano2021 *service.PorAnoItem
		var ano2023 *service.PorAnoItem
		for i := range anos {
			if anos[i].Ano == 2021 {
				ano2021 = &anos[i]
			}
			if anos[i].Ano == 2023 {
				ano2023 = &anos[i]
			}
		}

		if ano2021 == nil || ano2021.TotalJogos != 0 || ano2021.GameDoAno != nil {
			t.Errorf("ano 2021 lacuna incorreto: %+v", ano2021)
		}
		if ano2023 == nil || ano2023.TotalJogos != 1 || ano2023.GameDoAno != nil {
			t.Errorf("ano 2023 sem destaque incorreto: %+v", ano2023)
		}
	})

	t.Run("RecordesComEmpateETempoZero", func(t *testing.T) {
		user := criarUsuarioDashboard(t, env.pool)
		token := gerarTokenDashboard(t, env.secret, user)

		d1 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Jogo Tempo Zero",
			Console:      "PC",
			FinalizadoEm: d1,
			TempoJogado:  0,
			Nota:         5,
			Dificuldade:  "C",
		})

		d2 := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Curto Antigo",
			Console:      "PC",
			FinalizadoEm: d2,
			TempoJogado:  100,
			Nota:         6,
			Dificuldade:  "C",
		})

		d3 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Curto Recente",
			Console:      "PC",
			FinalizadoEm: d3,
			TempoJogado:  100,
			Nota:         7,
			Dificuldade:  "C",
		})

		d4 := time.Date(2021, 5, 1, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Longo Antigo",
			Console:      "PS5",
			FinalizadoEm: d4,
			TempoJogado:  50000,
			Nota:         9,
			Dificuldade:  "AA",
		})

		d5 := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Longo Recente",
			Console:      "PS5",
			FinalizadoEm: d5,
			TempoJogado:  50000,
			Nota:         10,
			Dificuldade:  "AAA",
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/recordes", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		var rec service.RecordesResponse
		_ = json.Unmarshal(w.Body.Bytes(), &rec)

		if rec.MaisCurto == nil || rec.MaisCurto.Nome != "Curto Recente" || rec.MaisCurto.TempoJogadoSegundos != 100 {
			t.Errorf("mais curto deveria ignorar tempo 0 e desempatar por finalizado_em DESC: %+v", rec.MaisCurto)
		}
		if rec.MaisLongo == nil || rec.MaisLongo.Nome != "Longo Recente" || rec.MaisLongo.TempoJogadoSegundos != 50000 {
			t.Errorf("mais longo deveria desempatar por finalizado_em DESC: %+v", rec.MaisLongo)
		}
	})

	t.Run("BreakdownTipoCaseInsensitiveESemAcento", func(t *testing.T) {
		user := criarUsuarioDashboard(t, env.pool)
		token := gerarTokenDashboard(t, env.secret, user)

		d := time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC)
		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Jogo RPG 1",
			Console:      "PC",
			Genero:       strPtr("RPG"),
			Tipo:         strPtr("Principal"),
			FinalizadoEm: d,
			TempoJogado:  1000,
			Nota:         8,
			Dificuldade:  "B",
		})

		inserirJogoZeradoTeste(t, env.pool, jogoTesteParams{
			UsuarioID:    user,
			Nome:         "Jogo Ação 1",
			Console:      "PS5",
			Genero:       strPtr("Ação e Aventura"),
			Tipo:         strPtr("Campanha"),
			FinalizadoEm: d,
			TempoJogado:  2000,
			Nota:         9,
			Dificuldade:  "A",
		})

		reqRPG, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/breakdown-tipo?genero=rpg", nil)
		reqRPG.Header.Set("Authorization", "Bearer "+token)
		wRPG := httptest.NewRecorder()
		env.router.ServeHTTP(wRPG, reqRPG)
		var resRPG []service.BreakdownTipoItem
		_ = json.Unmarshal(wRPG.Body.Bytes(), &resRPG)
		if len(resRPG) != 1 || resRPG[0].Tipo != "Principal" || resRPG[0].TotalJogos != 1 {
			t.Errorf("breakdown tipo com genero=rpg falhou: %+v", resRPG)
		}

		reqAcao, _ := http.NewRequest(http.MethodGet, "/api/v1/dashboard/breakdown-tipo?genero=acao%20e%20aventura", nil)
		reqAcao.Header.Set("Authorization", "Bearer "+token)
		wAcao := httptest.NewRecorder()
		env.router.ServeHTTP(wAcao, reqAcao)
		var resAcao []service.BreakdownTipoItem
		_ = json.Unmarshal(wAcao.Body.Bytes(), &resAcao)
		if len(resAcao) != 1 || resAcao[0].Tipo != "Campanha" || resAcao[0].TotalJogos != 1 {
			t.Errorf("breakdown tipo com acao sem acento falhou: %+v", resAcao)
		}
	})
}
