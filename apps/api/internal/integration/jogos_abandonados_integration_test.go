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

type abandonadosIntegrationEnv struct {
	pool    *pgxpool.Pool
	sqlDB   *sql.DB
	router  *gin.Engine
	tokens  *service.AuthToken
	secret  string
	connStr string
}

func setupAbandonadosIntegrationEnv(t *testing.T) *abandonadosIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_abandonados_test"),
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

	jogosRepo := repository.NewJogosRepository(pool, queries)
	jogosSvc := service.NewJogosService(jogosRepo)
	jogosH := handler.NewJogosHandler(jogosSvc)

	abandonadosRepo := repository.NewJogosAbandonadosRepository(pool, queries)
	abandonadosSvc := service.NewJogosAbandonadosService(abandonadosRepo)
	abandonadosH := handler.NewJogosAbandonadosHandler(abandonadosSvc)

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	privadas.GET("/jogos", jogosH.ListarJogos)
	privadas.GET("/jogos/filtros", jogosH.ObterFiltros)
	privadas.GET("/jogos/game-do-ano", jogosH.ObterGameDoAnoResumo)
	privadas.POST("/jogos", jogosH.CriarJogo)

	privadas.GET("/jogos-abandonados", abandonadosH.Listar)
	privadas.GET("/jogos-abandonados/filtros", abandonadosH.ObterFiltros)
	privadas.GET("/jogos-abandonados/total", abandonadosH.ObterTotal)
	privadas.GET("/jogos-abandonados/:id", abandonadosH.ObterDetalhes)
	privadas.POST("/jogos-abandonados", abandonadosH.Criar)
	privadas.PUT("/jogos-abandonados/:id", abandonadosH.Atualizar)
	privadas.DELETE("/jogos-abandonados/:id", abandonadosH.Excluir)

	return &abandonadosIntegrationEnv{
		pool:    pool,
		sqlDB:   sqlDB,
		router:  router,
		tokens:  tokens,
		secret:  secret,
		connStr: connStr,
	}
}

func criarUsuarioAbandonados(t *testing.T, pool *pgxpool.Pool) int32 {
	t.Helper()
	uid := uuid.NewString()
	email := fmt.Sprintf("user-%s@example.com", uid)
	nome := fmt.Sprintf("User %s", uid[:8])

	var id int32
	err := pool.QueryRow(context.Background(),
		"INSERT INTO usuarios (nome, email, senha_hash) VALUES ($1, $2, $3) RETURNING id",
		nome, email, "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01",
	).Scan(&id)
	if err != nil {
		t.Fatalf("falha ao criar usuario de teste: %v", err)
	}
	return id
}

func gerarTokenAbandonados(t *testing.T, secret string, usuarioID int32) string {
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

func TestIntegration_JogosAbandonados(t *testing.T) {
	env := setupAbandonadosIntegrationEnv(t)

	t.Run("1_Migration_0012_Aplica_E_Reverte", func(t *testing.T) {
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
				WHERE table_schema = 'public' AND table_name = 'jogos_abandonados'
			)
		`).Scan(&existeTabela)
		if err != nil || !existeTabela {
			t.Fatalf("esperava que a tabela jogos_abandonados existisse apos migrate up, err=%v", err)
		}

		if err := m.Migrate(11); err != nil {
			t.Fatalf("falha ao reverter migration 0012: %v", err)
		}

		err = env.sqlDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'jogos_abandonados'
			)
		`).Scan(&existeTabela)
		if err != nil || existeTabela {
			t.Fatalf("esperava que a tabela jogos_abandonados NAO existisse apos revert, err=%v", err)
		}

		if err := m.Up(); err != nil {
			t.Fatalf("falha ao reaplicar migration 0012: %v", err)
		}

		err = env.sqlDB.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'jogos_abandonados'
			)
		`).Scan(&existeTabela)
		if err != nil || !existeTabela {
			t.Fatalf("esperava que a tabela jogos_abandonados existisse apos reup, err=%v", err)
		}
	})

	t.Run("2_Fluxo_Completo_Criar_Listar_Obter_Atualizar_Excluir", func(t *testing.T) {
		usuarioID := criarUsuarioAbandonados(t, env.pool)
		token := gerarTokenAbandonados(t, env.secret, usuarioID)

		bodyCriar := []byte(`{
			"nome": "Dark Souls II",
			"console": "PS3",
			"tempo_jogado_horas": 15,
			"tempo_jogado_minutos": 30,
			"tempo_jogado_segundos": 0,
			"motivo": "Frustrante demais na Shrine of Amana"
		}`)
		reqCriar := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(bodyCriar))
		reqCriar.Header.Set("Authorization", "Bearer "+token)
		reqCriar.Header.Set("Content-Type", "application/json")
		wCriar := httptest.NewRecorder()
		env.router.ServeHTTP(wCriar, reqCriar)

		if wCriar.Code != http.StatusCreated {
			t.Fatalf("status criacao=%d, body=%s", wCriar.Code, wCriar.Body.String())
		}
		var respCriar struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wCriar.Body.Bytes(), &respCriar)
		jogoID := respCriar.Data.ID
		if jogoID <= 0 || respCriar.Data.Nome != "Dark Souls II" || respCriar.Data.TempoJogado != 55800 {
			t.Fatalf("dados do jogo criado incorretos: %+v", respCriar.Data)
		}

		reqObter := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID), nil)
		reqObter.Header.Set("Authorization", "Bearer "+token)
		wObter := httptest.NewRecorder()
		env.router.ServeHTTP(wObter, reqObter)
		if wObter.Code != http.StatusOK {
			t.Fatalf("status obter=%d, body=%s", wObter.Code, wObter.Body.String())
		}
		var respObter struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wObter.Body.Bytes(), &respObter)
		if respObter.Data.ID != jogoID {
			t.Fatalf("id inesperado: %d", respObter.Data.ID)
		}

		bodyAtualizar := []byte(`{
			"nome": "Dark Souls II: Scholar of the First Sin",
			"console": "PS4",
			"tempo_jogado_horas": 20,
			"tempo_jogado_minutos": 0,
			"tempo_jogado_segundos": 0,
			"motivo": "Larguei novamente"
		}`)
		reqAtualizar := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID),
			bytes.NewReader(bodyAtualizar),
		)
		reqAtualizar.Header.Set("Authorization", "Bearer "+token)
		reqAtualizar.Header.Set("Content-Type", "application/json")
		wAtualizar := httptest.NewRecorder()
		env.router.ServeHTTP(wAtualizar, reqAtualizar)
		if wAtualizar.Code != http.StatusOK {
			t.Fatalf("status atualizar=%d, body=%s", wAtualizar.Code, wAtualizar.Body.String())
		}
		var respAtualizar struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wAtualizar.Body.Bytes(), &respAtualizar)
		if respAtualizar.Data.Nome != "Dark Souls II: Scholar of the First Sin" ||
			respAtualizar.Data.TempoJogado != 72000 {
			t.Fatalf("dados atualizados incorretos: %+v", respAtualizar.Data)
		}

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("status delete=%d, body=%s", wDel.Code, wDel.Body.String())
		}

		reqObterExcluido := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID),
			nil,
		)
		reqObterExcluido.Header.Set("Authorization", "Bearer "+token)
		wObterExcluido := httptest.NewRecorder()
		env.router.ServeHTTP(wObterExcluido, reqObterExcluido)
		if wObterExcluido.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em GET de excluido, obteve %d", wObterExcluido.Code)
		}

		reqAtualizarExcluido := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID),
			bytes.NewReader(bodyAtualizar),
		)
		reqAtualizarExcluido.Header.Set("Authorization", "Bearer "+token)
		reqAtualizarExcluido.Header.Set("Content-Type", "application/json")
		wAtualizarExcluido := httptest.NewRecorder()
		env.router.ServeHTTP(wAtualizarExcluido, reqAtualizarExcluido)
		if wAtualizarExcluido.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em PUT de excluido, obteve %d", wAtualizarExcluido.Code)
		}

		reqDelExcluido := httptest.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID),
			nil,
		)
		reqDelExcluido.Header.Set("Authorization", "Bearer "+token)
		wDelExcluido := httptest.NewRecorder()
		env.router.ServeHTTP(wDelExcluido, reqDelExcluido)
		if wDelExcluido.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em DELETE de excluido, obteve %d", wDelExcluido.Code)
		}

		reqListar := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados", nil)
		reqListar.Header.Set("Authorization", "Bearer "+token)
		wListar := httptest.NewRecorder()
		env.router.ServeHTTP(wListar, reqListar)
		var respListar struct {
			Data []repository.JogoAbandonado `json:"data"`
			Meta struct {
				Total int64 `json:"total"`
			} `json:"meta"`
		}
		_ = json.Unmarshal(wListar.Body.Bytes(), &respListar)
		if respListar.Meta.Total != 0 || len(respListar.Data) != 0 {
			t.Fatalf("registro excluido continua aparecendo na listagem: %+v", respListar)
		}
	})

	t.Run("3_Isolamento_Entre_Usuarios", func(t *testing.T) {
		userA := criarUsuarioAbandonados(t, env.pool)
		userB := criarUsuarioAbandonados(t, env.pool)
		tokenA := gerarTokenAbandonados(t, env.secret, userA)
		tokenB := gerarTokenAbandonados(t, env.secret, userB)

		bodyCriar := []byte(`{"nome":"Jogo do Usuario A","console":"PS5"}`)
		reqCriar := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(bodyCriar))
		reqCriar.Header.Set("Authorization", "Bearer "+tokenA)
		reqCriar.Header.Set("Content-Type", "application/json")
		wCriar := httptest.NewRecorder()
		env.router.ServeHTTP(wCriar, reqCriar)
		if wCriar.Code != http.StatusCreated {
			t.Fatalf("status criacao=%d", wCriar.Code)
		}
		var respCriar struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wCriar.Body.Bytes(), &respCriar)
		jogoIDA := respCriar.Data.ID

		reqGetB := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoIDA), nil)
		reqGetB.Header.Set("Authorization", "Bearer "+tokenB)
		wGetB := httptest.NewRecorder()
		env.router.ServeHTTP(wGetB, reqGetB)
		if wGetB.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 quando usuario B obtem registro de A, obteve %d", wGetB.Code)
		}

		reqPutB := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoIDA),
			bytes.NewReader([]byte(`{"nome":"Invasor","console":"X"}`)),
		)
		reqPutB.Header.Set("Authorization", "Bearer "+tokenB)
		reqPutB.Header.Set("Content-Type", "application/json")
		wPutB := httptest.NewRecorder()
		env.router.ServeHTTP(wPutB, reqPutB)
		if wPutB.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 quando usuario B edita registro de A, obteve %d", wPutB.Code)
		}

		reqDelB := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoIDA), nil)
		reqDelB.Header.Set("Authorization", "Bearer "+tokenB)
		wDelB := httptest.NewRecorder()
		env.router.ServeHTTP(wDelB, reqDelB)
		if wDelB.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 quando usuario B exclui registro de A, obteve %d", wDelB.Code)
		}

		reqListB := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados", nil)
		reqListB.Header.Set("Authorization", "Bearer "+tokenB)
		wListB := httptest.NewRecorder()
		env.router.ServeHTTP(wListB, reqListB)
		var respListB struct {
			Data []repository.JogoAbandonado `json:"data"`
			Meta struct {
				Total int64 `json:"total"`
			} `json:"meta"`
		}
		_ = json.Unmarshal(wListB.Body.Bytes(), &respListB)
		if respListB.Meta.Total != 0 || len(respListB.Data) != 0 {
			t.Fatalf("usuario B nao deveria ver registros de A: %+v", respListB)
		}
	})

	t.Run("4_Busca_Sem_Acento_Filtros_Ordenacao_Paginacao_Total", func(t *testing.T) {
		usuarioID := criarUsuarioAbandonados(t, env.pool)
		token := gerarTokenAbandonados(t, env.secret, usuarioID)

		jogos := []struct {
			nome         string
			console      string
			tempo        int
			abandonadoEm string
		}{
			{"Pokémon Red", "Game Boy", 1000, "2026-01-01T10:00:00Z"},
			{"Pokemon Blue", "Game Boy", 2000, "2026-02-01T10:00:00Z"},
			{"Chrono Trigger", "SNES", 5000, "2026-03-01T10:00:00Z"},
			{"Ás do Espaço", "SNES", 500, "2026-04-01T10:00:00Z"},
		}

		for _, j := range jogos {
			body, _ := json.Marshal(map[string]any{
				"nome":          j.nome,
				"console":       j.console,
				"tempo_jogado":  j.tempo,
				"abandonado_em": j.abandonadoEm,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			if w.Code != http.StatusCreated {
				t.Fatalf("falha ao criar jogo %s: %s", j.nome, w.Body.String())
			}
		}

		reqBusca := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados?busca=pokemon", nil)
		reqBusca.Header.Set("Authorization", "Bearer "+token)
		wBusca := httptest.NewRecorder()
		env.router.ServeHTTP(wBusca, reqBusca)
		var respBusca struct {
			Data []repository.JogoAbandonado `json:"data"`
			Meta struct {
				Total int64 `json:"total"`
			} `json:"meta"`
		}
		_ = json.Unmarshal(wBusca.Body.Bytes(), &respBusca)
		if respBusca.Meta.Total != 2 {
			t.Fatalf("esperava 2 resultados na busca por 'pokemon', obteve %d", respBusca.Meta.Total)
		}

		reqBuscaAcento := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados?busca=as", nil)
		reqBuscaAcento.Header.Set("Authorization", "Bearer "+token)
		wBuscaAcento := httptest.NewRecorder()
		env.router.ServeHTTP(wBuscaAcento, reqBuscaAcento)
		var respBuscaAcento struct {
			Data []repository.JogoAbandonado `json:"data"`
			Meta struct {
				Total int64 `json:"total"`
			} `json:"meta"`
		}
		_ = json.Unmarshal(wBuscaAcento.Body.Bytes(), &respBuscaAcento)
		if respBuscaAcento.Meta.Total != 1 || respBuscaAcento.Data[0].Nome != "Ás do Espaço" {
			t.Fatalf("esperava encontrar 'Ás do Espaço' buscando 'as', obteve: %+v", respBuscaAcento)
		}

		reqConsole := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados?console=Game%20Boy", nil)
		reqConsole.Header.Set("Authorization", "Bearer "+token)
		wConsole := httptest.NewRecorder()
		env.router.ServeHTTP(wConsole, reqConsole)
		var respConsole struct {
			Meta struct {
				Total int64 `json:"total"`
			} `json:"meta"`
		}
		_ = json.Unmarshal(wConsole.Body.Bytes(), &respConsole)
		if respConsole.Meta.Total != 2 {
			t.Fatalf("esperava 2 jogos para Game Boy, obteve %d", respConsole.Meta.Total)
		}

		reqFiltros := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados/filtros", nil)
		reqFiltros.Header.Set("Authorization", "Bearer "+token)
		wFiltros := httptest.NewRecorder()
		env.router.ServeHTTP(wFiltros, reqFiltros)
		var respFiltros struct {
			Data struct {
				Consoles []string `json:"consoles"`
			} `json:"data"`
		}
		_ = json.Unmarshal(wFiltros.Body.Bytes(), &respFiltros)
		if len(respFiltros.Data.Consoles) != 2 ||
			respFiltros.Data.Consoles[0] != "Game Boy" ||
			respFiltros.Data.Consoles[1] != "SNES" {
			t.Fatalf("consoles inesperados: %+v", respFiltros.Data.Consoles)
		}

		reqTotal := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados/total", nil)
		reqTotal.Header.Set("Authorization", "Bearer "+token)
		wTotal := httptest.NewRecorder()
		env.router.ServeHTTP(wTotal, reqTotal)
		var respTotal struct {
			Data struct {
				Total int64 `json:"total"`
			} `json:"data"`
		}
		_ = json.Unmarshal(wTotal.Body.Bytes(), &respTotal)
		if respTotal.Data.Total != 4 {
			t.Fatalf("esperava total 4, obteve %d", respTotal.Data.Total)
		}

		reqTempo := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados?ordenar=tempo", nil)
		reqTempo.Header.Set("Authorization", "Bearer "+token)
		wTempo := httptest.NewRecorder()
		env.router.ServeHTTP(wTempo, reqTempo)
		var respTempo struct {
			Data []repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wTempo.Body.Bytes(), &respTempo)
		if len(respTempo.Data) != 4 || respTempo.Data[0].Nome != "Chrono Trigger" {
			t.Fatalf("ordenacao por tempo incorreta: %+v", respTempo.Data)
		}

		reqNome := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados?ordenar=nome", nil)
		reqNome.Header.Set("Authorization", "Bearer "+token)
		wNome := httptest.NewRecorder()
		env.router.ServeHTTP(wNome, reqNome)
		var respNome struct {
			Data []repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wNome.Body.Bytes(), &respNome)
		if len(respNome.Data) != 4 || respNome.Data[0].Nome != "Ás do Espaço" {
			t.Fatalf("ordenacao por nome incorreta: %+v", respNome.Data)
		}

		reqAntigos := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados?ordenar=antigos", nil)
		reqAntigos.Header.Set("Authorization", "Bearer "+token)
		wAntigos := httptest.NewRecorder()
		env.router.ServeHTTP(wAntigos, reqAntigos)
		var respAntigos struct {
			Data []repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wAntigos.Body.Bytes(), &respAntigos)
		if len(respAntigos.Data) != 4 || respAntigos.Data[0].Nome != "Pokémon Red" {
			t.Fatalf("ordenacao antigos incorreta: %+v", respAntigos.Data)
		}

		reqRecentes := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados?ordenar=recentes", nil)
		reqRecentes.Header.Set("Authorization", "Bearer "+token)
		wRecentes := httptest.NewRecorder()
		env.router.ServeHTTP(wRecentes, reqRecentes)
		var respRecentes struct {
			Data []repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wRecentes.Body.Bytes(), &respRecentes)
		if len(respRecentes.Data) != 4 || respRecentes.Data[0].Nome != "Ás do Espaço" {
			t.Fatalf("ordenacao recentes incorreta: %+v", respRecentes.Data)
		}

		reqPag := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados?pagina=1&por_pagina=2", nil)
		reqPag.Header.Set("Authorization", "Bearer "+token)
		wPag := httptest.NewRecorder()
		env.router.ServeHTTP(wPag, reqPag)
		var respPag struct {
			Data []repository.JogoAbandonado `json:"data"`
			Meta struct {
				Pagina       int   `json:"pagina"`
				PorPagina    int   `json:"por_pagina"`
				Total        int64 `json:"total"`
				TotalPaginas int   `json:"total_paginas"`
			} `json:"meta"`
		}
		_ = json.Unmarshal(wPag.Body.Bytes(), &respPag)
		if len(respPag.Data) != 2 || respPag.Meta.Total != 4 || respPag.Meta.TotalPaginas != 2 ||
			respPag.Meta.Pagina != 1 || respPag.Meta.PorPagina != 2 {
			t.Fatalf("paginacao incorreta: %+v", respPag.Meta)
		}
	})

	t.Run("5_Prova_De_Isolamento_Das_Estatisticas", func(t *testing.T) {
		usuarioID := criarUsuarioAbandonados(t, env.pool)
		token := gerarTokenAbandonados(t, env.secret, usuarioID)

		bodyJogoZerado := []byte(`{
			"nome": "Super Metroid",
			"console": "SNES",
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado": 3600,
			"nota": 10,
			"dificuldade": "A"
		}`)
		reqJogo := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyJogoZerado))
		reqJogo.Header.Set("Authorization", "Bearer "+token)
		reqJogo.Header.Set("Content-Type", "application/json")
		wJogo := httptest.NewRecorder()
		env.router.ServeHTTP(wJogo, reqJogo)
		if wJogo.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo zerado: %s", wJogo.Body.String())
		}

		fazerChamada := func(url string) string {
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("falha em GET %s (%d): %s", url, w.Code, w.Body.String())
			}
			return w.Body.String()
		}

		jogosAntes := fazerChamada("/api/v1/jogos")
		filtrosAntes := fazerChamada("/api/v1/jogos/filtros")
		gameDoAnoAntes := fazerChamada("/api/v1/jogos/game-do-ano")

		for i := 1; i <= 3; i++ {
			bodyAbandonado := []byte(fmt.Sprintf(`{
				"nome": "Jogo Abandonado %d",
				"console": "Mega Drive",
				"tempo_jogado": %d
			}`, i, i*1000))
			reqAb := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(bodyAbandonado))
			reqAb.Header.Set("Authorization", "Bearer "+token)
			reqAb.Header.Set("Content-Type", "application/json")
			wAb := httptest.NewRecorder()
			env.router.ServeHTTP(wAb, reqAb)
			if wAb.Code != http.StatusCreated {
				t.Fatalf("falha ao criar abandonado: %s", wAb.Body.String())
			}
		}

		jogosDepois := fazerChamada("/api/v1/jogos")
		filtrosDepois := fazerChamada("/api/v1/jogos/filtros")
		gameDoAnoDepois := fazerChamada("/api/v1/jogos/game-do-ano")

		if jogosAntes != jogosDepois {
			t.Fatalf("GET /jogos mudou apos criar abandonados!\nAntes: %s\nDepois: %s", jogosAntes, jogosDepois)
		}
		if filtrosAntes != filtrosDepois {
			t.Fatalf("GET /jogos/filtros mudou apos criar abandonados!\nAntes: %s\nDepois: %s", filtrosAntes, filtrosDepois)
		}
		if gameDoAnoAntes != gameDoAnoDepois {
			t.Fatalf("GET /jogos/game-do-ano mudou apos criar abandonados!\nAntes: %s\nDepois: %s", gameDoAnoAntes, gameDoAnoDepois)
		}
	})

	t.Run("6_SoftDelete_Mantem_Linha_No_Banco", func(t *testing.T) {
		usuarioID := criarUsuarioAbandonados(t, env.pool)
		token := gerarTokenAbandonados(t, env.secret, usuarioID)

		bodyCriar := []byte(`{
			"nome": "Castlevania: Symphony of the Night",
			"console": "PS1",
			"tempo_jogado_horas": 5,
			"tempo_jogado_minutos": 0,
			"tempo_jogado_segundos": 0,
			"motivo": "Perdi o save"
		}`)
		reqCriar := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(bodyCriar))
		reqCriar.Header.Set("Authorization", "Bearer "+token)
		reqCriar.Header.Set("Content-Type", "application/json")
		wCriar := httptest.NewRecorder()
		env.router.ServeHTTP(wCriar, reqCriar)

		if wCriar.Code != http.StatusCreated {
			t.Fatalf("status criacao=%d, body=%s", wCriar.Code, wCriar.Body.String())
		}
		var respCriar struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wCriar.Body.Bytes(), &respCriar)
		jogoID := respCriar.Data.ID
		if jogoID <= 0 {
			t.Fatalf("id invalido apos criacao: %d", jogoID)
		}

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("status delete=%d, body=%s", wDel.Code, wDel.Body.String())
		}

		var (
			dbID      int32
			deletedAt sql.NullTime
		)
		err := env.pool.QueryRow(context.Background(),
			"SELECT id, deleted_at FROM jogos_abandonados WHERE id = $1",
			jogoID,
		).Scan(&dbID, &deletedAt)
		if err != nil {
			t.Fatalf("falha ao consultar jogo abandonado no banco apos delete: %v", err)
		}
		if dbID != jogoID {
			t.Fatalf("id no banco (%d) diferente do esperado (%d)", dbID, jogoID)
		}
		if !deletedAt.Valid || deletedAt.Time.IsZero() {
			t.Fatalf("esperava deleted_at preenchido no banco, obteve: %+v", deletedAt)
		}
	})

	t.Run("7_Excluido_Some_De_Listagem_Total_Filtros", func(t *testing.T) {
		usuarioID := criarUsuarioAbandonados(t, env.pool)
		token := gerarTokenAbandonados(t, env.secret, usuarioID)

		bodyJogo1 := []byte(`{"nome":"Final Fantasy VII","console":"PlayStation"}`)
		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(bodyJogo1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		env.router.ServeHTTP(w1, req1)
		if w1.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo 1: %s", w1.Body.String())
		}
		var resp1 struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(w1.Body.Bytes(), &resp1)
		jogoID1 := resp1.Data.ID

		bodyJogo2 := []byte(`{"nome":"Halo Combat Evolved","console":"Xbox"}`)
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(bodyJogo2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		env.router.ServeHTTP(w2, req2)
		if w2.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo 2: %s", w2.Body.String())
		}
		var resp2 struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(w2.Body.Bytes(), &resp2)
		jogoID2 := resp2.Data.ID

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID2), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("status delete=%d, body=%s", wDel.Code, wDel.Body.String())
		}

		reqListar := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados", nil)
		reqListar.Header.Set("Authorization", "Bearer "+token)
		wListar := httptest.NewRecorder()
		env.router.ServeHTTP(wListar, reqListar)
		if wListar.Code != http.StatusOK {
			t.Fatalf("status listar=%d, body=%s", wListar.Code, wListar.Body.String())
		}
		var respListar struct {
			Data []repository.JogoAbandonado `json:"data"`
			Meta struct {
				Total int64 `json:"total"`
			} `json:"meta"`
		}
		_ = json.Unmarshal(wListar.Body.Bytes(), &respListar)
		if respListar.Meta.Total != 1 || len(respListar.Data) != 1 {
			t.Fatalf("esperava total 1 na listagem, obteve meta=%d, len=%d", respListar.Meta.Total, len(respListar.Data))
		}
		if respListar.Data[0].ID != jogoID1 {
			t.Fatalf("esperava apenas jogo 1 (%d) na listagem, obteve id=%d", jogoID1, respListar.Data[0].ID)
		}

		reqTotal := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados/total", nil)
		reqTotal.Header.Set("Authorization", "Bearer "+token)
		wTotal := httptest.NewRecorder()
		env.router.ServeHTTP(wTotal, reqTotal)
		if wTotal.Code != http.StatusOK {
			t.Fatalf("status total=%d, body=%s", wTotal.Code, wTotal.Body.String())
		}
		var respTotal struct {
			Data struct {
				Total int64 `json:"total"`
			} `json:"data"`
		}
		_ = json.Unmarshal(wTotal.Body.Bytes(), &respTotal)
		if respTotal.Data.Total != 1 {
			t.Fatalf("esperava total 1, obteve %d", respTotal.Data.Total)
		}

		reqFiltros := httptest.NewRequest(http.MethodGet, "/api/v1/jogos-abandonados/filtros", nil)
		reqFiltros.Header.Set("Authorization", "Bearer "+token)
		wFiltros := httptest.NewRecorder()
		env.router.ServeHTTP(wFiltros, reqFiltros)
		if wFiltros.Code != http.StatusOK {
			t.Fatalf("status filtros=%d, body=%s", wFiltros.Code, wFiltros.Body.String())
		}
		var respFiltros struct {
			Data struct {
				Consoles []string `json:"consoles"`
			} `json:"data"`
		}
		_ = json.Unmarshal(wFiltros.Body.Bytes(), &respFiltros)
		if len(respFiltros.Data.Consoles) != 1 || respFiltros.Data.Consoles[0] != "PlayStation" {
			t.Fatalf("esperava filtros apenas com PlayStation, obteve %+v", respFiltros.Data.Consoles)
		}
	})

	t.Run("8_Excluido_Retorna_404", func(t *testing.T) {
		usuarioID := criarUsuarioAbandonados(t, env.pool)
		token := gerarTokenAbandonados(t, env.secret, usuarioID)

		bodyCriar := []byte(`{"nome":"Silent Hill 2","console":"PS2"}`)
		reqCriar := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(bodyCriar))
		reqCriar.Header.Set("Authorization", "Bearer "+token)
		reqCriar.Header.Set("Content-Type", "application/json")
		wCriar := httptest.NewRecorder()
		env.router.ServeHTTP(wCriar, reqCriar)
		if wCriar.Code != http.StatusCreated {
			t.Fatalf("status criacao=%d, body=%s", wCriar.Code, wCriar.Body.String())
		}
		var respCriar struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wCriar.Body.Bytes(), &respCriar)
		jogoID := respCriar.Data.ID

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("status delete=%d, body=%s", wDel.Code, wDel.Body.String())
		}

		type erroPayload struct {
			Error struct {
				Codigo   string `json:"codigo"`
				Mensagem string `json:"mensagem"`
			} `json:"error"`
		}

		reqGet := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID), nil)
		reqGet.Header.Set("Authorization", "Bearer "+token)
		wGet := httptest.NewRecorder()
		env.router.ServeHTTP(wGet, reqGet)
		if wGet.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em GET de excluido, obteve %d", wGet.Code)
		}
		var errGet erroPayload
		_ = json.Unmarshal(wGet.Body.Bytes(), &errGet)
		if errGet.Error.Codigo != "abandonados.nao_encontrado" {
			t.Fatalf("codigo erro GET inesperado: %s", errGet.Error.Codigo)
		}

		bodyPut := []byte(`{"nome":"Silent Hill 2 Director's Cut","console":"PS2"}`)
		reqPut := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID),
			bytes.NewReader(bodyPut),
		)
		reqPut.Header.Set("Authorization", "Bearer "+token)
		reqPut.Header.Set("Content-Type", "application/json")
		wPut := httptest.NewRecorder()
		env.router.ServeHTTP(wPut, reqPut)
		if wPut.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em PUT de excluido, obteve %d", wPut.Code)
		}
		var errPut erroPayload
		_ = json.Unmarshal(wPut.Body.Bytes(), &errPut)
		if errPut.Error.Codigo != "abandonados.nao_encontrado" {
			t.Fatalf("codigo erro PUT inesperado: %s", errPut.Error.Codigo)
		}

		reqDel2 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos-abandonados/%d", jogoID), nil)
		reqDel2.Header.Set("Authorization", "Bearer "+token)
		wDel2 := httptest.NewRecorder()
		env.router.ServeHTTP(wDel2, reqDel2)
		if wDel2.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em segundo DELETE de excluido, obteve %d", wDel2.Code)
		}
		var errDel2 erroPayload
		_ = json.Unmarshal(wDel2.Body.Bytes(), &errDel2)
		if errDel2.Error.Codigo != "abandonados.nao_encontrado" {
			t.Fatalf("codigo erro DELETE inesperado: %s", errDel2.Error.Codigo)
		}
	})

	t.Run("9_Excluir_Abandonado_Nao_Afeta_Zerados", func(t *testing.T) {
		usuarioID := criarUsuarioAbandonados(t, env.pool)
		token := gerarTokenAbandonados(t, env.secret, usuarioID)

		bodyJogoZerado := []byte(`{
			"nome": "Super Mario World",
			"console": "SNES",
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado": 7200,
			"nota": 10,
			"dificuldade": "A"
		}`)
		reqJogo := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyJogoZerado))
		reqJogo.Header.Set("Authorization", "Bearer "+token)
		reqJogo.Header.Set("Content-Type", "application/json")
		wJogo := httptest.NewRecorder()
		env.router.ServeHTTP(wJogo, reqJogo)
		if wJogo.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo zerado: %s", wJogo.Body.String())
		}

		bodyAbandonado := []byte(`{
			"nome": "Donkey Kong Country",
			"console": "SNES",
			"tempo_jogado": 1800
		}`)
		reqAb := httptest.NewRequest(http.MethodPost, "/api/v1/jogos-abandonados", bytes.NewReader(bodyAbandonado))
		reqAb.Header.Set("Authorization", "Bearer "+token)
		reqAb.Header.Set("Content-Type", "application/json")
		wAb := httptest.NewRecorder()
		env.router.ServeHTTP(wAb, reqAb)
		if wAb.Code != http.StatusCreated {
			t.Fatalf("falha ao criar abandonado: %s", wAb.Body.String())
		}
		var respAb struct {
			Data repository.JogoAbandonado `json:"data"`
		}
		_ = json.Unmarshal(wAb.Body.Bytes(), &respAb)
		abandonadoID := respAb.Data.ID

		fazerChamada := func(url string) string {
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("falha em GET %s (%d): %s", url, w.Code, w.Body.String())
			}
			return w.Body.String()
		}

		jogosAntes := fazerChamada("/api/v1/jogos")
		gameDoAnoAntes := fazerChamada("/api/v1/jogos/game-do-ano")

		reqDel := httptest.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("/api/v1/jogos-abandonados/%d", abandonadoID),
			nil,
		)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("status delete abandonado=%d, body=%s", wDel.Code, wDel.Body.String())
		}

		jogosDepois := fazerChamada("/api/v1/jogos")
		gameDoAnoDepois := fazerChamada("/api/v1/jogos/game-do-ano")

		if jogosAntes != jogosDepois {
			t.Fatalf("GET /jogos mudou apos excluir abandonado!\nAntes: %s\nDepois: %s", jogosAntes, jogosDepois)
		}
		if gameDoAnoAntes != gameDoAnoDepois {
			t.Fatalf("GET /jogos/game-do-ano mudou apos excluir abandonado!\nAntes: %s\nDepois: %s", gameDoAnoAntes, gameDoAnoDepois)
		}
	})
}
