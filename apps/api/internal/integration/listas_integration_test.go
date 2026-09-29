//go:build integration

// Package integration implementa testes de integracao com banco real via testcontainers.
package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/migration"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/UlerichLabs/memory-card/apps/api/migrations"
)

type listasIntegrationEnv struct {
	pool       *pgxpool.Pool
	router     *gin.Engine
	tokens     *service.AuthToken
	secret     string
	listasRepo repository.ListasRepository
	jogosRepo  repository.JogosRepository
}

type mockIntegrationIGDB struct {
	obterFranquiaFn            func(ctx context.Context, id int64) (*igdbclient.Franchise, error)
	jogosDaFranquiaFn          func(ctx context.Context, id int64) ([]igdbclient.Game, error)
	atualizarJogosDaFranquiaFn func(ctx context.Context, id int64) ([]igdbclient.Game, error)
}

func (m *mockIntegrationIGDB) ObterFranquia(ctx context.Context, id int64) (*igdbclient.Franchise, error) {
	if m.obterFranquiaFn != nil {
		return m.obterFranquiaFn(ctx, id)
	}
	return &igdbclient.Franchise{ID: id, Name: "Zelda"}, nil
}

func (m *mockIntegrationIGDB) JogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	if m.jogosDaFranquiaFn != nil {
		return m.jogosDaFranquiaFn(ctx, id)
	}
	return []igdbclient.Game{}, nil
}

func (m *mockIntegrationIGDB) AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	if m.atualizarJogosDaFranquiaFn != nil {
		return m.atualizarJogosDaFranquiaFn(ctx, id)
	}
	return []igdbclient.Game{}, nil
}

func setupListasIntegrationEnv(t *testing.T) *listasIntegrationEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_listas_integration"),
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

	if _, err := migration.Apply(pool); err != nil {
		pool.Close()
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao aplicar migrations: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer teardownCancel()
		_ = pgContainer.Terminate(teardownCtx)
	})

	queries := db.New(pool)
	jogosRepo := repository.NewJogosRepository(pool, queries)
	listasRepo := repository.NewListasRepository(pool, queries)
	mockIGDB := &mockIntegrationIGDB{}
	listasService := service.NewListasService(listasRepo, mockIGDB)
	jogosService := service.NewJogosService(jogosRepo)

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	listasHandler := handler.NewListasHandler(listasService)
	jogosHandler := handler.NewJogosHandler(jogosService)

	privadas.GET("/jogos", jogosHandler.ListarJogos)
	privadas.POST("/jogos", jogosHandler.CriarJogo)
	privadas.DELETE("/jogos/:id", jogosHandler.ExcluirJogo)

	privadas.GET("/listas", listasHandler.ListarListas)
	privadas.POST("/listas", listasHandler.CriarLista)
	privadas.GET("/listas/:id", listasHandler.ObterLista)
	privadas.PUT("/listas/:id", listasHandler.AtualizarLista)
	privadas.DELETE("/listas/:id", listasHandler.ExcluirLista)
	privadas.POST("/listas/:id/itens", listasHandler.AdicionarItem)
	privadas.DELETE("/listas/:id/itens/:itemId", listasHandler.ExcluirItem)
	privadas.PUT("/listas/:id/ordem", listasHandler.ReordenarItens)
	privadas.PUT("/listas/:id/itens/:itemId/zeramento", listasHandler.AssociarJogoZerado)
	privadas.DELETE("/listas/:id/itens/:itemId/zeramento", listasHandler.DesassociarJogoZerado)
	privadas.POST("/listas/:id/sincronizar", listasHandler.SincronizarFranquia)

	return &listasIntegrationEnv{
		pool:       pool,
		router:     router,
		tokens:     tokens,
		secret:     secret,
		listasRepo: listasRepo,
		jogosRepo:  jogosRepo,
	}
}

func criarUsuarioNoBanco(t *testing.T, pool *pgxpool.Pool, email string) int32 {
	t.Helper()
	var id int32
	err := pool.QueryRow(context.Background(),
		"INSERT INTO usuarios (nome, email, senha_hash) VALUES ($1, $2, $3) RETURNING id",
		"User Test", email, "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01",
	).Scan(&id)
	if err != nil {
		t.Fatalf("falha ao criar usuario de teste: %v", err)
	}
	return id
}

func gerarTokenIntegracao(t *testing.T, secret string, usuarioID int32) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: "pt-BR",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", usuarioID),
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
	return token
}

func TestIntegration_MigrationUpDown(t *testing.T) {
	env := setupListasIntegrationEnv(t)

	dbConn := stdlib.OpenDBFromPool(env.pool)
	defer dbConn.Close()

	driver, err := pgxmigrate.WithInstance(dbConn, &pgxmigrate.Config{})
	if err != nil {
		t.Fatalf("criar driver migrate: %v", err)
	}
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("carregar source migrate: %v", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		t.Fatalf("instanciar migrate: %v", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Steps(-1); err != nil {
		t.Fatalf("falha ao executar migration down: %v", err)
	}

	var tableExists bool
	err = env.pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'listas')",
	).Scan(&tableExists)
	if err != nil {
		t.Fatalf("consulta informacao de tabela: %v", err)
	}
	if tableExists {
		t.Fatal("esperava que a tabela listas fosse removida apos migration down")
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("falha ao reaplicar migration up: %v", err)
	}

	err = env.pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'listas')",
	).Scan(&tableExists)
	if err != nil || !tableExists {
		t.Fatalf("esperava que a tabela listas existisse apos migration up, erro: %v", err)
	}
}

func TestIntegration_Progresso_ZeramentoAnteriorEInsensivel(t *testing.T) {
	env := setupListasIntegrationEnv(t)
	userAID := criarUsuarioNoBanco(t, env.pool, "user.a.progresso@example.com")
	tokenA := gerarTokenIntegracao(t, env.secret, userAID)

	igdbID := int32(1001)
	_, err := env.jogosRepo.Criar(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID:    userAID,
		IgdbID:       &igdbID,
		Nome:         "Super Mario World",
		Console:      "SNES",
		Genero:       "Platform",
		FinalizadoEm: time.Now().Add(-24 * time.Hour),
		Nota:         10,
		Dificuldade:  "A",
	})
	if err != nil {
		t.Fatalf("falha ao criar jogo zerado anterior por igdb_id: %v", err)
	}

	_, err = env.jogosRepo.Criar(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID:    userAID,
		Nome:         "Pokémon Red",
		Console:      "Game Boy",
		Genero:       "RPG",
		FinalizadoEm: time.Now().Add(-12 * time.Hour),
		Nota:         9,
		Dificuldade:  "B",
	})
	if err != nil {
		t.Fatalf("falha ao criar jogo zerado anterior por nome acentuado: %v", err)
	}

	criarListaBody := `{"tipo":"desafio","nome":"Desafio Classicos","regra":{"tipo":"manual"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listas", bytes.NewBufferString(criarListaBody))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("falha ao criar lista: %d - %s", w.Code, w.Body.String())
	}

	var criada struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &criada)
	listaID := criada.Data.ID

	item1Body := `{"nome":"Super Mario World","igdb_id":1001}`
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/listas/%d/itens", listaID), bytes.NewBufferString(item1Body))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("falha ao adicionar item 1: %d", w.Code)
	}

	item2Body := `{"nome":"pokemon red"}`
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/listas/%d/itens", listaID), bytes.NewBufferString(item2Body))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("falha ao adicionar item 2: %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/listas/%d", listaID), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("falha ao obter lista detalhada: %d", w.Code)
	}

	var listaDetalhe struct {
		Data service.ListaDetalhada `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &listaDetalhe)

	if listaDetalhe.Data.TotalItens != 2 {
		t.Fatalf("esperava 2 itens, obteve %d", listaDetalhe.Data.TotalItens)
	}
	if len(listaDetalhe.Data.Itens) != 2 {
		t.Fatalf("esperava 2 itens detalhados, obteve %d", len(listaDetalhe.Data.Itens))
	}
	if !listaDetalhe.Data.Itens[0].Zerado || listaDetalhe.Data.Itens[0].JogoZerado == nil {
		t.Fatalf("item 1 (igdb_id) deveria constar como zerado")
	}
	if !listaDetalhe.Data.Itens[1].Zerado || listaDetalhe.Data.Itens[1].JogoZerado == nil {
		t.Fatalf("item 2 (nome normalizado pokemon red vs Pokémon Red) deveria constar como zerado")
	}
	if listaDetalhe.Data.Progresso.Feitos != 2 || listaDetalhe.Data.Progresso.Percentual != 100 || !listaDetalhe.Data.Progresso.Concluido {
		t.Fatalf("progresso incorreto: %+v", listaDetalhe.Data.Progresso)
	}
}

func TestIntegration_ZeramentoSoftDeleted_NaoConta(t *testing.T) {
	env := setupListasIntegrationEnv(t)
	userID := criarUsuarioNoBanco(t, env.pool, "user.softdelete@example.com")
	token := gerarTokenIntegracao(t, env.secret, userID)

	igdbID := int32(500)
	jogo, err := env.jogosRepo.Criar(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID:    userID,
		IgdbID:       &igdbID,
		Nome:         "Metroid",
		Console:      "NES",
		Genero:       "Adventure",
		FinalizadoEm: time.Now(),
		Nota:         8,
		Dificuldade:  "A",
	})
	if err != nil {
		t.Fatalf("falha ao criar jogo zerado: %v", err)
	}

	criarListaBody := `{"tipo":"fila","nome":"Fila Metroid"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listas", bytes.NewBufferString(criarListaBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	var criada struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &criada)
	listaID := criada.Data.ID

	itemBody := `{"nome":"Metroid","igdb_id":500}`
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/listas/%d/itens", listaID), bytes.NewBufferString(itemBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/listas/%d", listaID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	var respAtiva struct {
		Data service.ListaDetalhada `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respAtiva)
	if !respAtiva.Data.Itens[0].Zerado {
		t.Fatalf("esperava item zerado")
	}

	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogo.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("falha ao excluir jogo zerado via API: %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/listas/%d", listaID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	var respAposDelete struct {
		Data service.ListaDetalhada `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respAposDelete)

	if respAposDelete.Data.Itens[0].Zerado {
		t.Fatalf("jogo excluido logicamente nao deveria contar como zerado")
	}
	if respAposDelete.Data.ItensPendentes != 1 {
		t.Fatalf("esperava 1 item pendente, obteve %d", respAposDelete.Data.ItensPendentes)
	}
}

func TestIntegration_ContagemPlataforma_CaseInsensitive(t *testing.T) {
	env := setupListasIntegrationEnv(t)
	userID := criarUsuarioNoBanco(t, env.pool, "user.platform@example.com")
	token := gerarTokenIntegracao(t, env.secret, userID)

	_, err := env.jogosRepo.Criar(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID:    userID,
		Nome:         "Alex Kidd",
		Console:      "Master System",
		Genero:       "Platform",
		FinalizadoEm: time.Now(),
		Nota:         9,
		Dificuldade:  "A",
		IgdbCapaURL:  "//images.igdb.com/capa_alex_kidd.jpg",
	})
	if err != nil {
		t.Fatalf("falha ao criar jogo zerado: %v", err)
	}

	criarListaBody := `{"tipo":"desafio","nome":"Desafio Master System","regra":{"tipo":"plataforma","valor":"master system"},"meta":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listas", bytes.NewBufferString(criarListaBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("falha ao criar desafio plataforma: %d - %s", w.Code, w.Body.String())
	}

	var criada struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &criada)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/listas/%d", criada.Data.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("falha ao consultar desafio: %d", w.Code)
	}

	var detalhe struct {
		Data service.ListaDetalhada `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &detalhe)

	if len(detalhe.Data.Itens) != 1 {
		t.Fatalf("esperava 1 item gerado por regra, obteve %d", len(detalhe.Data.Itens))
	}
	if detalhe.Data.Itens[0].Origem != "regra" || !detalhe.Data.Itens[0].Zerado {
		t.Fatalf("item deveria ter origem 'regra' e zerado=true")
	}
	if detalhe.Data.Itens[0].IgdbCapaURL == nil || *detalhe.Data.Itens[0].IgdbCapaURL != "//images.igdb.com/capa_alex_kidd.jpg" {
		t.Fatalf("esperava igdb_capa_url preenchido na regra, obteve %v", detalhe.Data.Itens[0].IgdbCapaURL)
	}
	if !detalhe.Data.Progresso.Concluido || detalhe.Data.Progresso.Feitos != 1 {
		t.Fatalf("desafio deveria estar concluido: %+v", detalhe.Data.Progresso)
	}
}

func TestIntegration_Progresso_ConcluidoEm_PosicaoMeta(t *testing.T) {
	env := setupListasIntegrationEnv(t)
	userID := criarUsuarioNoBanco(t, env.pool, "user.concluidoem@example.com")
	token := gerarTokenIntegracao(t, env.secret, userID)

	t1 := time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 5, 20, 15, 30, 0, 0, time.UTC)
	t3 := time.Date(2025, 8, 5, 18, 0, 0, 0, time.UTC)

	datas := []time.Time{t3, t1, t2}
	for i, d := range datas {
		_, err := env.jogosRepo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    userID,
			Nome:         fmt.Sprintf("RPG %d", i+1),
			Console:      "GBA",
			Genero:       "JRPG",
			FinalizadoEm: d,
			Nota:         8,
			Dificuldade:  "A",
		})
		if err != nil {
			t.Fatalf("falha ao criar jogo zerado %d: %v", i, err)
		}
	}

	criarListaBody := `{"tipo":"desafio","nome":"JRPGs de GBA","regra":{"tipo":"genero","valor":"jrpg"},"meta":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listas", bytes.NewBufferString(criarListaBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("falha ao criar desafio: %d - %s", w.Code, w.Body.String())
	}

	var criada struct {
		Data service.ListaDetalhada `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &criada)

	if !criada.Data.Progresso.Concluido {
		t.Fatal("esperava concluido=true")
	}
	if criada.Data.Progresso.ConcluidoEm == nil {
		t.Fatal("esperava concluido_em preenchido")
	}
	if !criada.Data.Progresso.ConcluidoEm.Equal(t2) {
		t.Fatalf("esperava concluido_em = %v (t2), obteve %v", t2, criada.Data.Progresso.ConcluidoEm)
	}
}

func TestIntegration_CascataEOnDeleteSetNull(t *testing.T) {
	env := setupListasIntegrationEnv(t)
	userID := criarUsuarioNoBanco(t, env.pool, "user.cascade@example.com")
	token := gerarTokenIntegracao(t, env.secret, userID)

	jogo, err := env.jogosRepo.Criar(context.Background(), repository.CriarJogoZeradoParams{
		UsuarioID:    userID,
		Nome:         "Castlevania",
		Console:      "GBA",
		Genero:       "Action",
		FinalizadoEm: time.Now(),
		Nota:         9,
		Dificuldade:  "AA",
	})
	if err != nil {
		t.Fatalf("falha ao criar jogo: %v", err)
	}

	criarListaBody := `{"tipo":"fila","nome":"Fila Castlevania"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listas", bytes.NewBufferString(criarListaBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	var criada struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &criada)
	listaID := criada.Data.ID

	itemBody := `{"nome":"Castlevania Aria of Sorrow"}`
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/listas/%d/itens", listaID), bytes.NewBufferString(itemBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	var itemCriado struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &itemCriado)
	itemID := itemCriado.Data.ID

	vincularBody := fmt.Sprintf(`{"jogo_zerado_id":%d}`, jogo.ID)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/listas/%d/itens/%d/zeramento", listaID, itemID), bytes.NewBufferString(vincularBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("falha ao vincular zeramento: %d", w.Code)
	}

	var jogoZeradoIDVinculado sql.NullInt32
	err = env.pool.QueryRow(context.Background(), "SELECT jogo_zerado_id FROM lista_itens WHERE id = $1", itemID).Scan(&jogoZeradoIDVinculado)
	if err != nil || !jogoZeradoIDVinculado.Valid || jogoZeradoIDVinculado.Int32 != jogo.ID {
		t.Fatalf("vinculo no banco invalido: %v", err)
	}

	_, err = env.pool.Exec(context.Background(), "DELETE FROM jogos_zerados WHERE id = $1", jogo.ID)
	if err != nil {
		t.Fatalf("falha ao excluir fisico em jogos_zerados: %v", err)
	}

	err = env.pool.QueryRow(context.Background(), "SELECT jogo_zerado_id FROM lista_itens WHERE id = $1", itemID).Scan(&jogoZeradoIDVinculado)
	if err != nil {
		t.Fatalf("falha ao consultar lista_itens: %v", err)
	}
	if jogoZeradoIDVinculado.Valid {
		t.Fatal("esperava que jogo_zerado_id ficasse NULL apos ON DELETE SET NULL")
	}

	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/listas/%d", listaID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("falha ao excluir lista: %d", w.Code)
	}

	var itemRestante bool
	err = env.pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM lista_itens WHERE id = $1)", itemID).Scan(&itemRestante)
	if err != nil || itemRestante {
		t.Fatalf("esperava que lista_itens sofresse cascade no delete da lista, restante: %v", itemRestante)
	}
}

func TestIntegration_IsolamentoEntreUsuarios(t *testing.T) {
	env := setupListasIntegrationEnv(t)

	userAID := criarUsuarioNoBanco(t, env.pool, "user.a.isolamento@example.com")
	userBID := criarUsuarioNoBanco(t, env.pool, "user.b.isolamento@example.com")

	tokenA := gerarTokenIntegracao(t, env.secret, userAID)
	tokenB := gerarTokenIntegracao(t, env.secret, userBID)

	criarListaBody := `{"tipo":"fila","nome":"Lista Exclusiva do Usuario A"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/listas", bytes.NewBufferString(criarListaBody))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	var criada struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &criada)
	listaIDA := criada.Data.ID

	itemBody := `{"nome":"Item A"}`
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/listas/%d/itens", listaIDA), bytes.NewBufferString(itemBody))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	var itemCriado struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &itemCriado)
	itemIDA := itemCriado.Data.ID

	req = httptest.NewRequest(http.MethodGet, "/api/v1/listas", nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	w = httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("falha ao listar para user B: %d", w.Code)
	}
	var listasB struct {
		Data []service.ListaResumo `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &listasB)
	if len(listasB.Data) != 0 {
		t.Fatalf("usuario B nao deveria ver listas de usuario A, obteve %d listas", len(listasB.Data))
	}

	tentativas := []struct {
		metodo string
		url    string
		body   string
	}{
		{"GET", fmt.Sprintf("/api/v1/listas/%d", listaIDA), ""},
		{"PUT", fmt.Sprintf("/api/v1/listas/%d", listaIDA), `{"nome":"Invasao"}`},
		{"DELETE", fmt.Sprintf("/api/v1/listas/%d", listaIDA), ""},
		{"POST", fmt.Sprintf("/api/v1/listas/%d/itens", listaIDA), `{"nome":"Item Invasao"}`},
		{"DELETE", fmt.Sprintf("/api/v1/listas/%d/itens/%d", listaIDA, itemIDA), ""},
		{"PUT", fmt.Sprintf("/api/v1/listas/%d/ordem", listaIDA), fmt.Sprintf(`{"item_ids":[%d]}`, itemIDA)},
		{"PUT", fmt.Sprintf("/api/v1/listas/%d/itens/%d/zeramento", listaIDA, itemIDA), `{"jogo_zerado_id":1}`},
		{"DELETE", fmt.Sprintf("/api/v1/listas/%d/itens/%d/zeramento", listaIDA, itemIDA), ""},
		{"POST", fmt.Sprintf("/api/v1/listas/%d/sincronizar", listaIDA), ""},
	}

	for _, tt := range tentativas {
		t.Run(tt.metodo+" "+tt.url, func(t *testing.T) {
			r := httptest.NewRequest(tt.metodo, tt.url, bytes.NewBufferString(tt.body))
			r.Header.Set("Authorization", "Bearer "+tokenB)
			r.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			env.router.ServeHTTP(rec, r)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("esperava 404 para acesso indevido do usuario B em %s %s, obteve: %d", tt.metodo, tt.url, rec.Code)
			}
		})
	}
}
