//go:build integration

// Package integration contem os testes de integracao com banco real.
package integration

import (
	"bytes"
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
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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
	migfs "github.com/UlerichLabs/memory-card/apps/api/migrations"
)

func init() {
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

type perfilIntegrationEnv struct {
	pool    *pgxpool.Pool
	sqlDB   *sql.DB
	router  *gin.Engine
	tokens  *service.AuthToken
	secret  string
	connStr string
}

func setupPerfilIntegrationEnv(t *testing.T) *perfilIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_perfil_test"),
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
	usuarioRepo := repository.NewUsuarioRepository(queries, pool)
	perfilSvc := service.NewPerfilService(usuarioRepo)
	perfilH := handler.NewPerfilHandler(perfilSvc)

	jogosRepo := repository.NewJogosRepository(pool, queries)
	jogosSvc := service.NewJogosService(jogosRepo)
	jogosH := handler.NewJogosHandler(jogosSvc)

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	privadas.GET("/me/perfil", perfilH.ObterPerfil)
	privadas.PUT("/me/perfil", perfilH.AtualizarPerfil)
	privadas.POST("/jogos", jogosH.CriarJogo)
	privadas.DELETE("/jogos/:id", jogosH.ExcluirJogo)

	return &perfilIntegrationEnv{
		pool:    pool,
		sqlDB:   sqlDB,
		router:  router,
		tokens:  tokens,
		secret:  secret,
		connStr: connStr,
	}
}

func criarUsuarioPerfil(t *testing.T, pool *pgxpool.Pool) int32 {
	t.Helper()
	uid := uuid.NewString()
	email := fmt.Sprintf("user-%s@example.com", uid)
	nome := fmt.Sprintf("User %s", uid[:8])

	var usuarioID int32
	err := pool.QueryRow(context.Background(),
		"INSERT INTO usuarios (nome, email, senha_hash) VALUES ($1, $2, $3) RETURNING id",
		nome, email, "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01",
	).Scan(&usuarioID)
	if err != nil {
		t.Fatalf("falha ao criar usuario de teste: %v", err)
	}
	return usuarioID
}

func gerarTokenPerfil(t *testing.T, secret string, usuarioID int32) string {
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
		t.Fatalf("falha ao gerar token de teste: %v", err)
	}
	return token
}

func fazerRequisicaoPerfil(
	router *gin.Engine,
	metodo string,
	caminho string,
	token string,
	corpo any,
) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if corpo != nil {
		switch valor := corpo.(type) {
		case []byte:
			bodyReader = bytes.NewReader(valor)
		case string:
			bodyReader = bytes.NewReader([]byte(valor))
		default:
			bytesCorpo, _ := json.Marshal(corpo)
			bodyReader = bytes.NewReader(bytesCorpo)
		}
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	request := httptest.NewRequest(metodo, caminho, bodyReader)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if corpo != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestIntegration_Perfil(t *testing.T) {
	env := setupPerfilIntegrationEnv(t)

	t.Run("Migration 0016 sobe e desce", func(t *testing.T) {
		driver, err := pgxmigrate.WithInstance(env.sqlDB, &pgxmigrate.Config{})
		if err != nil {
			t.Fatalf("criar driver migrate: %v", err)
		}
		sourceDriver, err := iofs.New(migfs.FS, ".")
		if err != nil {
			t.Fatalf("carregar migrations iofs: %v", err)
		}
		m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
		if err != nil {
			t.Fatalf("instanciar migrate: %v", err)
		}
		defer func() { _, _ = m.Close() }()

		var existeColuna bool
		err = env.sqlDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'usuarios' AND column_name = 'jogo_favorito_id'
			)
		`).Scan(&existeColuna)
		if err != nil || !existeColuna {
			t.Fatalf("esperava coluna jogo_favorito_id apos migrate up, err=%v", err)
		}

		if err := m.Steps(-1); err != nil {
			t.Fatalf("falha ao reverter migration 0016: %v", err)
		}

		err = env.sqlDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'usuarios' AND column_name = 'jogo_favorito_id'
			)
		`).Scan(&existeColuna)
		if err != nil || existeColuna {
			t.Fatalf("esperava que coluna jogo_favorito_id NAO existisse apos revert, err=%v", err)
		}

		if err := m.Up(); err != nil {
			t.Fatalf("falha ao reaplicar migration 0016: %v", err)
		}

		err = env.sqlDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'usuarios' AND column_name = 'jogo_favorito_id'
			)
		`).Scan(&existeColuna)
		if err != nil || !existeColuna {
			t.Fatalf("esperava coluna jogo_favorito_id apos reaplicar, err=%v", err)
		}
	})

	t.Run("Unicidade case-insensitive do username", func(t *testing.T) {
		usuario1 := criarUsuarioPerfil(t, env.pool)
		usuario2 := criarUsuarioPerfil(t, env.pool)
		token1 := gerarTokenPerfil(t, env.secret, usuario1)
		token2 := gerarTokenPerfil(t, env.secret, usuario2)

		rec1 := fazerRequisicaoPerfil(env.router, http.MethodPut, "/api/v1/me/perfil", token1, map[string]any{
			"nome":     "Lucas Primeiro",
			"username": "GamerMaster",
		})
		if rec1.Code != http.StatusOK {
			t.Fatalf("falha ao definir username para usuario 1: %d body=%s", rec1.Code, rec1.Body.String())
		}

		rec2 := fazerRequisicaoPerfil(env.router, http.MethodPut, "/api/v1/me/perfil", token2, map[string]any{
			"nome":     "Lucas Segundo",
			"username": "gamermaster",
		})
		if rec2.Code != http.StatusConflict {
			t.Fatalf("esperava 409 para username duplicado case-insensitive, obteve %d body=%s", rec2.Code, rec2.Body.String())
		}

		var respErro struct {
			Error struct {
				Codigo string `json:"codigo"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec2.Body.Bytes(), &respErro)
		if respErro.Error.Codigo != "perfil.username_em_uso" {
			t.Fatalf("codigo inesperado: %s", respErro.Error.Codigo)
		}

		rec3 := fazerRequisicaoPerfil(env.router, http.MethodPut, "/api/v1/me/perfil", token2, map[string]any{
			"nome":     "Lucas Segundo",
			"username": "GAMERMASTER",
		})
		if rec3.Code != http.StatusConflict {
			t.Fatalf("esperava 409 em uppercase, obteve %d", rec3.Code)
		}

		recProprio := fazerRequisicaoPerfil(env.router, http.MethodPut, "/api/v1/me/perfil", token1, map[string]any{
			"nome":     "Lucas Primeiro Atualizado",
			"username": "gamermaster",
		})
		if recProprio.Code != http.StatusOK {
			t.Fatalf("atualizar proprio username com caixa diferente deveria passar, obteve %d", recProprio.Code)
		}
	})

	t.Run("Jogo favorito deletado soft delete volta como null no GET", func(t *testing.T) {
		usuarioID := criarUsuarioPerfil(t, env.pool)
		token := gerarTokenPerfil(t, env.secret, usuarioID)

		corpoJogo := map[string]any{
			"nome":          "Chrono Trigger",
			"console":       "SNES",
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  7200,
			"nota":          10,
			"dificuldade":   "A",
			"igdb_capa_url": "https://images.igdb.com/cover.jpg",
		}
		recJogo := fazerRequisicaoPerfil(env.router, http.MethodPost, "/api/v1/jogos", token, corpoJogo)
		if recJogo.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo zerado: %d body=%s", recJogo.Code, recJogo.Body.String())
		}
		var respJogo struct {
			Data struct {
				ID int32 `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(recJogo.Body.Bytes(), &respJogo)
		jogoID := respJogo.Data.ID

		recPut := fazerRequisicaoPerfil(env.router, http.MethodPut, "/api/v1/me/perfil", token, map[string]any{
			"nome":             "Lucas Fan",
			"jogo_favorito_id": jogoID,
		})
		if recPut.Code != http.StatusOK {
			t.Fatalf("falha ao definir jogo favorito: %d body=%s", recPut.Code, recPut.Body.String())
		}

		recGet1 := fazerRequisicaoPerfil(env.router, http.MethodGet, "/api/v1/me/perfil", token, nil)
		if recGet1.Code != http.StatusOK {
			t.Fatalf("falha GET perfil: %d", recGet1.Code)
		}
		var respGet1 struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		_ = json.Unmarshal(recGet1.Body.Bytes(), &respGet1)
		if respGet1.Data.JogoFavorito == nil || respGet1.Data.JogoFavorito.ID != jogoID || respGet1.Data.JogoFavorito.Nome != "Chrono Trigger" {
			t.Fatalf("jogo favorito esperado Chrono Trigger, obteve: %+v", respGet1.Data.JogoFavorito)
		}

		recDel := fazerRequisicaoPerfil(env.router, http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), token, nil)
		if recDel.Code != http.StatusNoContent {
			t.Fatalf("falha ao deletar jogo zerado: %d", recDel.Code)
		}

		recGet2 := fazerRequisicaoPerfil(env.router, http.MethodGet, "/api/v1/me/perfil", token, nil)
		if recGet2.Code != http.StatusOK {
			t.Fatalf("falha GET perfil apos soft delete: %d", recGet2.Code)
		}
		var respGet2 struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		_ = json.Unmarshal(recGet2.Body.Bytes(), &respGet2)
		if respGet2.Data.JogoFavorito != nil {
			t.Fatalf("esperava jogo favorito null apos soft delete, obteve: %+v", respGet2.Data.JogoFavorito)
		}
	})

	t.Run("jogando_desde_efetivo calculado do primeiro zerado real", func(t *testing.T) {
		usuarioID := criarUsuarioPerfil(t, env.pool)
		token := gerarTokenPerfil(t, env.secret, usuarioID)

		recVazio := fazerRequisicaoPerfil(env.router, http.MethodGet, "/api/v1/me/perfil", token, nil)
		if recVazio.Code != http.StatusOK {
			t.Fatalf("falha GET perfil vazio: %d", recVazio.Code)
		}
		var respVazio struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		_ = json.Unmarshal(recVazio.Body.Bytes(), &respVazio)
		if respVazio.Data.JogandoDesde != nil || respVazio.Data.JogandoDesdeEfetivo != nil {
			t.Fatalf("esperava jogando desde nulo sem jogos, obteve: %+v", respVazio.Data)
		}

		jogo2021 := map[string]any{
			"nome":          "Metroid Dread",
			"console":       "Switch",
			"finalizado_em": "2021-11-01T10:00:00Z",
			"tempo_jogado":  36000,
			"nota":          9,
			"dificuldade":   "A",
		}
		fazerRequisicaoPerfil(env.router, http.MethodPost, "/api/v1/jogos", token, jogo2021)

		jogo2018 := map[string]any{
			"nome":          "Celeste",
			"console":       "PC",
			"finalizado_em": "2018-05-01T10:00:00Z",
			"tempo_jogado":  18000,
			"nota":          10,
			"dificuldade":   "A",
		}
		fazerRequisicaoPerfil(env.router, http.MethodPost, "/api/v1/jogos", token, jogo2018)

		jogo2015Deletado := map[string]any{
			"nome":          "The Witcher 3",
			"console":       "PC",
			"finalizado_em": "2015-06-01T10:00:00Z",
			"tempo_jogado":  100000,
			"nota":          10,
			"dificuldade":   "A",
		}
		rec2015 := fazerRequisicaoPerfil(env.router, http.MethodPost, "/api/v1/jogos", token, jogo2015Deletado)
		var resp2015 struct {
			Data struct {
				ID int32 `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(rec2015.Body.Bytes(), &resp2015)
		fazerRequisicaoPerfil(env.router, http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", resp2015.Data.ID), token, nil)

		recEfetivo := fazerRequisicaoPerfil(env.router, http.MethodGet, "/api/v1/me/perfil", token, nil)
		var respEfetivo struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		_ = json.Unmarshal(recEfetivo.Body.Bytes(), &respEfetivo)
		if respEfetivo.Data.JogandoDesde != nil {
			t.Fatalf("jogando_desde salvo deveria ser nulo, obteve %d", *respEfetivo.Data.JogandoDesde)
		}
		if respEfetivo.Data.JogandoDesdeEfetivo == nil || *respEfetivo.Data.JogandoDesdeEfetivo != 2018 {
			t.Fatalf("jogando_desde_efetivo esperado 2018 (menor ano nao deletado), obteve: %+v", respEfetivo.Data.JogandoDesdeEfetivo)
		}

		fazerRequisicaoPerfil(env.router, http.MethodPut, "/api/v1/me/perfil", token, map[string]any{
			"nome":          "Lucas Jogador",
			"jogando_desde": 1990,
		})
		recSalvo := fazerRequisicaoPerfil(env.router, http.MethodGet, "/api/v1/me/perfil", token, nil)
		var respSalvo struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		_ = json.Unmarshal(recSalvo.Body.Bytes(), &respSalvo)
		if respSalvo.Data.JogandoDesde == nil || *respSalvo.Data.JogandoDesde != 1990 {
			t.Fatalf("jogando_desde salvo esperado 1990, obteve: %+v", respSalvo.Data.JogandoDesde)
		}
		if respSalvo.Data.JogandoDesdeEfetivo == nil || *respSalvo.Data.JogandoDesdeEfetivo != 1990 {
			t.Fatalf("jogando_desde_efetivo esperado 1990 quando valor salvo existe, obteve: %+v", respSalvo.Data.JogandoDesdeEfetivo)
		}
	})

	t.Run("Validacao de jogo favorito de outro usuario retorna 404", func(t *testing.T) {
		usuarioA := criarUsuarioPerfil(t, env.pool)
		usuarioB := criarUsuarioPerfil(t, env.pool)
		tokenA := gerarTokenPerfil(t, env.secret, usuarioA)
		tokenB := gerarTokenPerfil(t, env.secret, usuarioB)

		recJogoA := fazerRequisicaoPerfil(env.router, http.MethodPost, "/api/v1/jogos", tokenA, map[string]any{
			"nome":          "Super Mario 64",
			"console":       "N64",
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  15000,
			"nota":          10,
			"dificuldade":   "A",
		})
		var respJogoA struct {
			Data struct {
				ID int32 `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(recJogoA.Body.Bytes(), &respJogoA)
		jogoIDA := respJogoA.Data.ID

		recPutB := fazerRequisicaoPerfil(env.router, http.MethodPut, "/api/v1/me/perfil", tokenB, map[string]any{
			"nome":             "Invasor B",
			"jogo_favorito_id": jogoIDA,
		})
		if recPutB.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 para jogo de outro usuario, obteve %d body=%s", recPutB.Code, recPutB.Body.String())
		}
		var respErro struct {
			Error struct {
				Codigo string `json:"codigo"`
			} `json:"error"`
		}
		_ = json.Unmarshal(recPutB.Body.Bytes(), &respErro)
		if respErro.Error.Codigo != "perfil.jogo_favorito_nao_encontrado" {
			t.Fatalf("codigo inesperado: %s", respErro.Error.Codigo)
		}
	})
}
