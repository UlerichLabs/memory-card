//go:build integration

// Package integration contem os testes de integracao com banco real.
package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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

type preferenciasIntegrationEnv struct {
	pool    *pgxpool.Pool
	sqlDB   *sql.DB
	router  *gin.Engine
	tokens  *service.AuthToken
	secret  string
	connStr string
}

func setupPreferenciasIntegrationEnv(t *testing.T) *preferenciasIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_preferencias_test"),
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
		sqlDB.Close()
		pool.Close()
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
	preferenciasRepo := repository.NewPreferenciasRepository(queries, pool)
	preferenciasService := service.NewPreferenciasService(preferenciasRepo)
	preferenciasHandler := handler.NewPreferenciasHandler(preferenciasService)

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/me/preferencias", preferenciasHandler.ObterPreferencias)
	privadas.PUT("/me/preferencias", preferenciasHandler.AtualizarPreferencias)

	return &preferenciasIntegrationEnv{
		pool:    pool,
		sqlDB:   sqlDB,
		router:  router,
		tokens:  tokens,
		secret:  secret,
		connStr: connStr,
	}
}

func criarUsuarioPreferencias(t *testing.T, pool *pgxpool.Pool, idioma ...string) int32 {
	t.Helper()
	email := fmt.Sprintf("user_%s@example.com", uuid.NewString()[:8])
	lang := "pt-BR"
	if len(idioma) > 0 && idioma[0] != "" {
		lang = idioma[0]
	}
	var id int32
	err := pool.QueryRow(context.Background(), `
		INSERT INTO usuarios (nome, email, senha_hash, username, idioma)
		VALUES ($1, $2, 'hash_teste', $3, $4)
		RETURNING id
	`, "Usuario Teste", email, "user_"+uuid.NewString()[:8], lang).Scan(&id)
	if err != nil {
		t.Fatalf("falha ao criar usuario de teste: %v", err)
	}
	return id
}

func gerarTokenPreferencias(t *testing.T, secret string, usuarioID int32, idioma ...string) string {
	t.Helper()
	lang := "pt-BR"
	if len(idioma) > 0 && idioma[0] != "" {
		lang = idioma[0]
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: lang,
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
		t.Fatalf("falha ao gerar token: %v", err)
	}
	return raw
}

func fazerRequisicaoPreferencias(router *gin.Engine, metodo, caminho, token string, corpo any) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if corpo != nil {
		if r, ok := corpo.(io.Reader); ok {
			bodyReader = r
		} else {
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

func TestIntegration_Preferencias(t *testing.T) {
	env := setupPreferenciasIntegrationEnv(t)

	t.Run("Migration 0018 sobe e desce", func(t *testing.T) {
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

		var existeTabela bool
		err = env.sqlDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'preferencias_usuario'
			)
		`).Scan(&existeTabela)
		if err != nil || !existeTabela {
			t.Fatalf("esperava tabela preferencias_usuario apos migrate up, err=%v", err)
		}

		if err := m.Migrate(17); err != nil {
			t.Fatalf("falha ao reverter migration 0018: %v", err)
		}

		err = env.sqlDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'preferencias_usuario'
			)
		`).Scan(&existeTabela)
		if err != nil || existeTabela {
			t.Fatalf("esperava que tabela preferencias_usuario NAO existisse apos revert, err=%v", err)
		}

		if err := m.Up(); err != nil {
			t.Fatalf("falha ao reaplicar migration 0018: %v", err)
		}

		err = env.sqlDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'preferencias_usuario'
			)
		`).Scan(&existeTabela)
		if err != nil || !existeTabela {
			t.Fatalf("esperava tabela preferencias_usuario apos reaplicar, err=%v", err)
		}
	})

	t.Run("CHECKs rejeitam valor fora da lista no INSERT direto", func(t *testing.T) {
		usuarioID := criarUsuarioPreferencias(t, env.pool)

		testesCheck := []struct {
			nome  string
			query string
		}{
			{
				nome: "formato_data invalido",
				query: fmt.Sprintf(`INSERT INTO preferencias_usuario (usuario_id, formato_data) VALUES (%d, 'invalido')`, usuarioID),
			},
			{
				nome: "visual_biblioteca invalido",
				query: fmt.Sprintf(`INSERT INTO preferencias_usuario (usuario_id, visual_biblioteca) VALUES (%d, 'tabela')`, usuarioID),
			},
			{
				nome: "itens_por_pagina invalido",
				query: fmt.Sprintf(`INSERT INTO preferencias_usuario (usuario_id, itens_por_pagina) VALUES (%d, 10)`, usuarioID),
			},
			{
				nome: "modo_tema invalido",
				query: fmt.Sprintf(`INSERT INTO preferencias_usuario (usuario_id, modo_tema) VALUES (%d, 'neon')`, usuarioID),
			},
			{
				nome: "estilo_tema invalido",
				query: fmt.Sprintf(`INSERT INTO preferencias_usuario (usuario_id, estilo_tema) VALUES (%d, 'sega')`, usuarioID),
			},
		}

		for _, tc := range testesCheck {
			t.Run(tc.nome, func(t *testing.T) {
				_, err := env.pool.Exec(context.Background(), tc.query)
				if err == nil {
					t.Fatalf("esperava violacao de CHECK constraint para %s", tc.nome)
				}
			})
		}
	})

	t.Run("GET sem linha devolve defaults; PUT cria; segundo PUT atualiza sem duplicar", func(t *testing.T) {
		usuarioID := criarUsuarioPreferencias(t, env.pool, "pt-BR")
		token := gerarTokenPreferencias(t, env.secret, usuarioID, "pt-BR")

		recGet1 := fazerRequisicaoPreferencias(env.router, http.MethodGet, "/api/v1/me/preferencias", token, nil)
		if recGet1.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %s", recGet1.Code, recGet1.Body.String())
		}
		var respGet1 struct {
			Data repository.PreferenciasUsuario `json:"data"`
		}
		if err := json.Unmarshal(recGet1.Body.Bytes(), &respGet1); err != nil {
			t.Fatalf("erro decode get1: %v", err)
		}
		if respGet1.Data.Idioma != "pt-BR" ||
			respGet1.Data.FormatoData != "dmy" ||
			respGet1.Data.FusoHorario != "America/Sao_Paulo" ||
			respGet1.Data.VisualBiblioteca != "grade" ||
			respGet1.Data.ItensPorPagina != 24 ||
			respGet1.Data.ReduzirAnimacoes != false ||
			respGet1.Data.ModoTema != "escuro" ||
			respGet1.Data.EstiloTema != "padrao" {
			t.Fatalf("defaults inesperados: %+v", respGet1.Data)
		}

		var count int
		err := env.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM preferencias_usuario WHERE usuario_id = $1`, usuarioID).Scan(&count)
		if err != nil || count != 0 {
			t.Fatalf("nenhuma linha deveria ser criada no GET, count=%d err=%v", count, err)
		}

		put1 := map[string]any{
			"idioma":            "en",
			"formato_data":      "ymd",
			"fuso_horario":      "Europe/Lisbon",
			"visual_biblioteca": "lista",
			"itens_por_pagina":  50,
			"reduzir_animacoes": true,
			"modo_tema":         "claro",
			"estilo_tema":       "playstation",
		}
		recPut1 := fazerRequisicaoPreferencias(env.router, http.MethodPut, "/api/v1/me/preferencias", token, put1)
		if recPut1.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %s", recPut1.Code, recPut1.Body.String())
		}
		var respPut1 struct {
			Data repository.PreferenciasUsuario `json:"data"`
		}
		_ = json.Unmarshal(recPut1.Body.Bytes(), &respPut1)
		if respPut1.Data.Idioma != "en" || respPut1.Data.EstiloTema != "playstation" || respPut1.Data.ItensPorPagina != 50 {
			t.Fatalf("resposta do put1 incorreta: %+v", respPut1.Data)
		}

		err = env.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM preferencias_usuario WHERE usuario_id = $1`, usuarioID).Scan(&count)
		if err != nil || count != 1 {
			t.Fatalf("esperava 1 linha apos put1, count=%d err=%v", count, err)
		}

		var idiomaDB string
		err = env.pool.QueryRow(context.Background(), `SELECT idioma FROM usuarios WHERE id = $1`, usuarioID).Scan(&idiomaDB)
		if err != nil || idiomaDB != "en" {
			t.Fatalf("idioma em usuarios deveria ser 'en', obteve %s, err=%v", idiomaDB, err)
		}

		put2 := map[string]any{
			"idioma":            "en",
			"formato_data":      "mdy",
			"fuso_horario":      "America/New_York",
			"visual_biblioteca": "grade",
			"itens_por_pagina":  100,
			"reduzir_animacoes": false,
			"modo_tema":         "automatico",
			"estilo_tema":       "steam",
		}
		recPut2 := fazerRequisicaoPreferencias(env.router, http.MethodPut, "/api/v1/me/preferencias", token, put2)
		if recPut2.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %s", recPut2.Code, recPut2.Body.String())
		}
		var respPut2 struct {
			Data repository.PreferenciasUsuario `json:"data"`
		}
		_ = json.Unmarshal(recPut2.Body.Bytes(), &respPut2)
		if respPut2.Data.EstiloTema != "steam" || respPut2.Data.ItensPorPagina != 100 || respPut2.Data.ModoTema != "automatico" {
			t.Fatalf("resposta do put2 incorreta: %+v", respPut2.Data)
		}

		err = env.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM preferencias_usuario WHERE usuario_id = $1`, usuarioID).Scan(&count)
		if err != nil || count != 1 {
			t.Fatalf("segundo PUT duplicou linha, count=%d err=%v", count, err)
		}
	})

	t.Run("PUT altera usuarios.idioma e a preferencia na mesma transacao (rollback)", func(t *testing.T) {
		usuarioID := criarUsuarioPreferencias(t, env.pool, "pt-BR")

		queries := db.New(env.pool)
		repo := repository.NewPreferenciasRepository(queries, env.pool)

		ctx := context.Background()
		_, err := repo.AtualizarPreferencias(ctx, repository.AtualizarPreferenciasParams{
			UsuarioID:        usuarioID,
			Idioma:           "en",
			FormatoData:      "invalido_trigger_db_check",
			FusoHorario:      "America/Sao_Paulo",
			VisualBiblioteca: "grade",
			ItensPorPagina:   24,
			ReduzirAnimacoes: false,
			ModoTema:         "escuro",
			EstiloTema:       "padrao",
		})
		if err == nil {
			t.Fatal("esperava erro do banco na transacao")
		}

		var idiomaDB string
		err = env.pool.QueryRow(ctx, `SELECT idioma FROM usuarios WHERE id = $1`, usuarioID).Scan(&idiomaDB)
		if err != nil {
			t.Fatalf("falha ao consultar usuario: %v", err)
		}
		if idiomaDB != "pt-BR" {
			t.Fatalf("transacao deveria ter feito rollback do idioma, mas encontrou %s", idiomaDB)
		}

		var count int
		err = env.pool.QueryRow(ctx, `SELECT COUNT(*) FROM preferencias_usuario WHERE usuario_id = $1`, usuarioID).Scan(&count)
		if err != nil || count != 0 {
			t.Fatalf("transacao deveria ter feito rollback da preferencia, count=%d", count)
		}
	})

	t.Run("Isolamento entre usuarios: PUT do usuario A nao altera o B", func(t *testing.T) {
		userA := criarUsuarioPreferencias(t, env.pool, "pt-BR")
		userB := criarUsuarioPreferencias(t, env.pool, "pt-BR")
		tokenA := gerarTokenPreferencias(t, env.secret, userA, "pt-BR")
		tokenB := gerarTokenPreferencias(t, env.secret, userB, "pt-BR")

		putA := map[string]any{
			"idioma":            "en",
			"formato_data":      "ymd",
			"fuso_horario":      "Asia/Tokyo",
			"visual_biblioteca": "lista",
			"itens_por_pagina":  100,
			"reduzir_animacoes": true,
			"modo_tema":         "claro",
			"estilo_tema":       "nintendo",
		}
		recA := fazerRequisicaoPreferencias(env.router, http.MethodPut, "/api/v1/me/preferencias", tokenA, putA)
		if recA.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %s", recA.Code, recA.Body.String())
		}

		recB := fazerRequisicaoPreferencias(env.router, http.MethodGet, "/api/v1/me/preferencias", tokenB, nil)
		if recB.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body = %s", recB.Code, recB.Body.String())
		}
		var respB struct {
			Data repository.PreferenciasUsuario `json:"data"`
		}
		_ = json.Unmarshal(recB.Body.Bytes(), &respB)
		if respB.Data.Idioma != "pt-BR" || respB.Data.EstiloTema != "padrao" || respB.Data.ItensPorPagina != 24 {
			t.Fatalf("preferencias do usuario B foram alteradas indevidamente: %+v", respB.Data)
		}
	})

	t.Run("ON DELETE CASCADE ao remover o usuario", func(t *testing.T) {
		usuarioID := criarUsuarioPreferencias(t, env.pool, "pt-BR")
		token := gerarTokenPreferencias(t, env.secret, usuarioID, "pt-BR")

		payload := map[string]any{
			"idioma":            "pt-BR",
			"formato_data":      "dmy",
			"fuso_horario":      "America/Sao_Paulo",
			"visual_biblioteca": "grade",
			"itens_por_pagina":  24,
			"reduzir_animacoes": false,
			"modo_tema":         "escuro",
			"estilo_tema":       "padrao",
		}
		rec := fazerRequisicaoPreferencias(env.router, http.MethodPut, "/api/v1/me/preferencias", token, payload)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}

		var count int
		err := env.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM preferencias_usuario WHERE usuario_id = $1`, usuarioID).Scan(&count)
		if err != nil || count != 1 {
			t.Fatalf("esperava 1 linha antes do delete, count=%d err=%v", count, err)
		}

		_, err = env.pool.Exec(context.Background(), `DELETE FROM usuarios WHERE id = $1`, usuarioID)
		if err != nil {
			t.Fatalf("falha ao deletar usuario: %v", err)
		}

		err = env.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM preferencias_usuario WHERE usuario_id = $1`, usuarioID).Scan(&count)
		if err != nil || count != 0 {
			t.Fatalf("registro de preferencias_usuario deveria ter sido removido por CASCADE, count=%d err=%v", count, err)
		}
	})
}
