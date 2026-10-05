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
	"strings"
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

type jogandoAgoraIntegrationEnv struct {
	pool    *pgxpool.Pool
	sqlDB   *sql.DB
	router  *gin.Engine
	tokens  *service.AuthToken
	secret  string
	connStr string
}

type respostaItemJogandoAgora struct {
	Data repository.JogoEmAndamento `json:"data"`
}

type respostaListaJogandoAgora struct {
	Data []repository.JogoEmAndamento `json:"data"`
}

type respostaErroJogandoAgora struct {
	Error struct {
		Codigo   string `json:"codigo"`
		Mensagem string `json:"mensagem"`
	} `json:"error"`
}

func setupJogandoAgoraIntegrationEnv(t *testing.T) *jogandoAgoraIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_jogando_agora_test"),
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

	jogandoRepo := repository.NewJogosEmAndamentoRepository(queries)
	jogandoSvc := service.NewJogosEmAndamentoService(jogandoRepo)
	jogandoH := handler.NewJogosEmAndamentoHandler(jogandoSvc)

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)

	privadas.GET("/jogando", jogandoH.Listar)
	privadas.POST("/jogando", jogandoH.Criar)
	privadas.DELETE("/jogando/:id", jogandoH.Excluir)

	privadas.POST("/jogos", jogosH.CriarJogo)
	privadas.POST("/jogos-abandonados", abandonadosH.Criar)

	return &jogandoAgoraIntegrationEnv{
		pool:    pool,
		sqlDB:   sqlDB,
		router:  router,
		tokens:  tokens,
		secret:  secret,
		connStr: connStr,
	}
}

func criarUsuarioJogandoAgora(t *testing.T, pool *pgxpool.Pool) int32 {
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

func gerarTokenJogandoAgora(t *testing.T, secret string, usuarioID int32) string {
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

func fazerRequisicaoJogandoAgora(
	router *gin.Engine,
	metodo string,
	caminho string,
	token string,
	corpo any,
	idiomas ...string,
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
	if len(idiomas) > 0 && idiomas[0] != "" {
		request.Header.Set("Accept-Language", idiomas[0])
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestIntegration_JogandoAgora(t *testing.T) {
	env := setupJogandoAgoraIntegrationEnv(t)

	t.Run("Fluxo basico de criacao e listagem com campos e ordenacao", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoItem1 := map[string]any{
			"nome":          "Chrono Trigger",
			"igdb_id":       1010,
			"igdb_capa_url": "https://images.igdb.com/cover1.jpg",
			"iniciado_em":   "2026-01-10T12:00:00Z",
		}
		rec1 := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoItem1)
		if rec1.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 na criacao, obteve %d body=%s", rec1.Code, rec1.Body.String())
		}

		var payloadBrutoCriacao struct {
			Data map[string]any `json:"data"`
		}
		if err := json.Unmarshal(rec1.Body.Bytes(), &payloadBrutoCriacao); err != nil {
			t.Fatalf("falha ao decodificar json da resposta: %v", err)
		}
		for _, chave := range []string{"id", "nome", "igdb_id", "igdb_capa_url", "iniciado_em"} {
			if _, ok := payloadBrutoCriacao.Data[chave]; !ok {
				t.Fatalf("campo %s ausente no corpo retornado na criacao", chave)
			}
		}

		var respCriar1 respostaItemJogandoAgora
		if err := json.Unmarshal(rec1.Body.Bytes(), &respCriar1); err != nil {
			t.Fatalf("falha ao interpretar struct criada: %v", err)
		}
		if respCriar1.Data.ID <= 0 || respCriar1.Data.Nome != "Chrono Trigger" ||
			respCriar1.Data.IgdbID == nil || *respCriar1.Data.IgdbID != 1010 ||
			respCriar1.Data.IgdbCapaURL == nil || *respCriar1.Data.IgdbCapaURL != "https://images.igdb.com/cover1.jpg" {
			t.Fatalf("dados inesperados no item criado: %+v", respCriar1.Data)
		}

		corpoItem2 := map[string]any{
			"nome":        "Super Metroid",
			"iniciado_em": "2026-02-10T12:00:00Z",
		}
		rec2 := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoItem2)
		if rec2.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 para item 2, obteve %d body=%s", rec2.Code, rec2.Body.String())
		}
		var respCriar2 respostaItemJogandoAgora
		if err := json.Unmarshal(rec2.Body.Bytes(), &respCriar2); err != nil {
			t.Fatalf("falha ao interpretar struct 2: %v", err)
		}

		corpoItem3 := map[string]any{
			"nome":        "Castlevania Symphony of the Night",
			"iniciado_em": "2026-02-10T12:00:00Z",
		}
		rec3 := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoItem3)
		if rec3.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 para item 3, obteve %d body=%s", rec3.Code, rec3.Body.String())
		}
		var respCriar3 respostaItemJogandoAgora
		if err := json.Unmarshal(rec3.Body.Bytes(), &respCriar3); err != nil {
			t.Fatalf("falha ao interpretar struct 3: %v", err)
		}

		recListar := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		if recListar.Code != http.StatusOK {
			t.Fatalf("esperava status 200 na listagem, obteve %d body=%s", recListar.Code, recListar.Body.String())
		}

		var respLista respostaListaJogandoAgora
		if err := json.Unmarshal(recListar.Body.Bytes(), &respLista); err != nil {
			t.Fatalf("falha ao interpretar lista: %v", err)
		}
		if len(respLista.Data) != 3 {
			t.Fatalf("esperava 3 itens na listagem, obteve %d", len(respLista.Data))
		}

		if respLista.Data[0].ID != respCriar3.Data.ID {
			t.Fatalf("primeiro item deveria ser o item 3 (iniciado_em mais recente e id maior), obteve id=%d", respLista.Data[0].ID)
		}
		if respLista.Data[1].ID != respCriar2.Data.ID {
			t.Fatalf("segundo item deveria ser o item 2 (mesmo iniciado_em e id menor), obteve id=%d", respLista.Data[1].ID)
		}
		if respLista.Data[2].ID != respCriar1.Data.ID {
			t.Fatalf("terceiro item deveria ser o item 1 (iniciado_em mais antigo), obteve id=%d", respLista.Data[2].ID)
		}
	})

	t.Run("Exclusao logica com soft delete no banco e segundo delete retorna 404", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpo := map[string]any{
			"nome":        "Jogo Para Soft Delete",
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recCriar := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpo)
		if recCriar.Code != http.StatusCreated {
			t.Fatalf("status criacao=%d body=%s", recCriar.Code, recCriar.Body.String())
		}
		var respCriar respostaItemJogandoAgora
		_ = json.Unmarshal(recCriar.Body.Bytes(), &respCriar)
		jogoID := respCriar.Data.ID

		recDel := fazerRequisicaoJogandoAgora(
			env.router,
			http.MethodDelete,
			fmt.Sprintf("/api/v1/jogando/%d", jogoID),
			token,
			nil,
		)
		if recDel.Code != http.StatusNoContent {
			t.Fatalf("esperava status 204 no delete, obteve %d body=%s", recDel.Code, recDel.Body.String())
		}

		recListar := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		if recListar.Code != http.StatusOK {
			t.Fatalf("esperava status 200, obteve %d", recListar.Code)
		}
		var respLista respostaListaJogandoAgora
		_ = json.Unmarshal(recListar.Body.Bytes(), &respLista)
		if len(respLista.Data) != 0 {
			t.Fatalf("esperava listagem vazia apos exclusao, obteve %d itens", len(respLista.Data))
		}

		var (
			bancoID   int32
			deletedAt sql.NullTime
		)
		err := env.pool.QueryRow(context.Background(),
			"SELECT id, deleted_at FROM jogos_em_andamento WHERE id = $1",
			jogoID,
		).Scan(&bancoID, &deletedAt)
		if err != nil {
			t.Fatalf("falha ao consultar banco diretamente: %v", err)
		}
		if bancoID != jogoID {
			t.Fatalf("id do banco diferente: %d vs %d", bancoID, jogoID)
		}
		if !deletedAt.Valid || deletedAt.Time.IsZero() {
			t.Fatalf("esperava soft delete com deleted_at preenchido, obteve: %+v", deletedAt)
		}

		recDel2 := fazerRequisicaoJogandoAgora(
			env.router,
			http.MethodDelete,
			fmt.Sprintf("/api/v1/jogando/%d", jogoID),
			token,
			nil,
		)
		if recDel2.Code != http.StatusNotFound {
			t.Fatalf("esperava status 404 no segundo delete, obteve %d body=%s", recDel2.Code, recDel2.Body.String())
		}
		var respErro respostaErroJogandoAgora
		_ = json.Unmarshal(recDel2.Body.Bytes(), &respErro)
		if respErro.Error.Codigo != "jogando.nao_encontrado" {
			t.Fatalf("codigo de erro inesperado: %s", respErro.Error.Codigo)
		}
	})

	t.Run("Baixa automatica ao registrar jogo zerado com mesmo igdb_id", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoJogando := map[string]any{
			"nome":        "Final Fantasy VII",
			"igdb_id":     1001,
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogando := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogando)
		if recJogando.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo em andamento: %d", recJogando.Code)
		}

		corpoZerado := map[string]any{
			"nome":          "FFVII Remake",
			"console":       "PS4",
			"igdb_id":       1001,
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  7200,
			"nota":          10,
			"dificuldade":   "A",
		}
		recZerado := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos", token, corpoZerado)
		if recZerado.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 no cadastro de zerado, obteve %d body=%s", recZerado.Code, recZerado.Body.String())
		}

		recListar := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista respostaListaJogandoAgora
		_ = json.Unmarshal(recListar.Body.Bytes(), &respLista)
		if len(respLista.Data) != 0 {
			t.Fatalf("esperava baixa automatica e lista vazia, obteve %d itens", len(respLista.Data))
		}
	})

	t.Run("Baixa automatica ao registrar jogo zerado sem igdb_id por nome normalizado", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoJogando := map[string]any{
			"nome":        "  Pokémon Esmeralda  ",
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogando := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogando)
		if recJogando.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo em andamento: %d", recJogando.Code)
		}

		corpoZerado := map[string]any{
			"nome":          "pokemon esmeralda",
			"console":       "GBA",
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  36000,
			"nota":          9,
			"dificuldade":   "A",
		}
		recZerado := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos", token, corpoZerado)
		if recZerado.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 no cadastro de zerado, obteve %d body=%s", recZerado.Code, recZerado.Body.String())
		}

		recListar := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista respostaListaJogandoAgora
		_ = json.Unmarshal(recListar.Body.Bytes(), &respLista)
		if len(respLista.Data) != 0 {
			t.Fatalf("esperava baixa automatica com nome acentuado/espacos diferentes, obteve %d itens", len(respLista.Data))
		}
	})

	t.Run("Nao da baixa com nome parecido mas diferente no jogo zerado", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoJogando := map[string]any{
			"nome":        "Celeste",
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogando := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogando)
		if recJogando.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo em andamento: %d", recJogando.Code)
		}

		corpoZerado := map[string]any{
			"nome":          "Celeste (Steam)",
			"console":       "PC",
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  18000,
			"nota":          10,
			"dificuldade":   "A",
		}
		recZerado := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos", token, corpoZerado)
		if recZerado.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 no cadastro de zerado, obteve %d", recZerado.Code)
		}

		recListar := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista respostaListaJogandoAgora
		_ = json.Unmarshal(recListar.Body.Bytes(), &respLista)
		if len(respLista.Data) != 1 || respLista.Data[0].Nome != "Celeste" {
			t.Fatalf("nao deveria ter dado baixa em nome parecido: %+v", respLista.Data)
		}
	})

	t.Run("Nao da baixa com mesmo nome quando igdb_ids sao diferentes no jogo zerado", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoJogando := map[string]any{
			"nome":        "Tetris",
			"igdb_id":     2001,
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogando := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogando)
		if recJogando.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo em andamento: %d", recJogando.Code)
		}

		corpoZerado := map[string]any{
			"nome":          "Tetris",
			"console":       "Game Boy",
			"igdb_id":       2002,
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  3600,
			"nota":          8,
			"dificuldade":   "A",
		}
		recZerado := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos", token, corpoZerado)
		if recZerado.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 no cadastro de zerado, obteve %d", recZerado.Code)
		}

		recListar := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista respostaListaJogandoAgora
		_ = json.Unmarshal(recListar.Body.Bytes(), &respLista)
		if len(respLista.Data) != 1 || respLista.Data[0].Nome != "Tetris" {
			t.Fatalf("nao deveria ter dado baixa quando igdb_ids divergem: %+v", respLista.Data)
		}
	})

	t.Run("Baixa apenas o item mais antigo quando ha dois itens correspondentes", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoItem1 := map[string]any{
			"nome":        "Metroid Prime",
			"igdb_id":     3001,
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		rec1 := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoItem1)
		if rec1.Code != http.StatusCreated {
			t.Fatalf("falha item 1: %d", rec1.Code)
		}
		var resp1 respostaItemJogandoAgora
		_ = json.Unmarshal(rec1.Body.Bytes(), &resp1)

		corpoItem2 := map[string]any{
			"nome":        "Metroid Prime",
			"igdb_id":     3001,
			"iniciado_em": "2026-03-01T00:00:00Z",
		}
		rec2 := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoItem2)
		if rec2.Code != http.StatusCreated {
			t.Fatalf("falha item 2: %d", rec2.Code)
		}
		var resp2 respostaItemJogandoAgora
		_ = json.Unmarshal(rec2.Body.Bytes(), &resp2)

		corpoZerado := map[string]any{
			"nome":          "Metroid Prime",
			"console":       "GameCube",
			"igdb_id":       3001,
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  50000,
			"nota":          10,
			"dificuldade":   "A",
		}
		recZerado := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos", token, corpoZerado)
		if recZerado.Code != http.StatusCreated {
			t.Fatalf("falha ao registrar zerado: %d", recZerado.Code)
		}

		recListar := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista respostaListaJogandoAgora
		_ = json.Unmarshal(recListar.Body.Bytes(), &respLista)
		if len(respLista.Data) != 1 {
			t.Fatalf("esperava exatamente 1 item restante, obteve %d", len(respLista.Data))
		}
		if respLista.Data[0].ID != resp2.Data.ID {
			t.Fatalf("deveria ter mantido o mais recente (id=%d), mas manteve id=%d", resp2.Data.ID, respLista.Data[0].ID)
		}

		var deletedAtItem1 sql.NullTime
		err := env.pool.QueryRow(context.Background(),
			"SELECT deleted_at FROM jogos_em_andamento WHERE id = $1",
			resp1.Data.ID,
		).Scan(&deletedAtItem1)
		if err != nil || !deletedAtItem1.Valid {
			t.Fatalf("item mais antigo deveria ter deleted_at preenchido: err=%v deletedAt=%+v", err, deletedAtItem1)
		}
	})

	t.Run("Nao altera lista de jogando quando jogo zerado nao tem correspondente", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoJogando := map[string]any{
			"nome":        "Super Mario Odyssey",
			"igdb_id":     4001,
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogando := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogando)
		if recJogando.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo em andamento: %d", recJogando.Code)
		}

		corpoZerado := map[string]any{
			"nome":          "Zelda Breath of the Wild",
			"console":       "Switch",
			"igdb_id":       4002,
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  100000,
			"nota":          10,
			"dificuldade":   "A",
		}
		recZerado := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos", token, corpoZerado)
		if recZerado.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 no cadastro de zerado sem correspondente, obteve %d", recZerado.Code)
		}

		recListar := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista respostaListaJogandoAgora
		_ = json.Unmarshal(recListar.Body.Bytes(), &respLista)
		if len(respLista.Data) != 1 || respLista.Data[0].Nome != "Super Mario Odyssey" {
			t.Fatalf("lista de jogando deveria permanecer intacta: %+v", respLista.Data)
		}
	})

	t.Run("Baixa automatica ao registrar jogo abandonado nos cenarios principais", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoJogandoIGDB := map[string]any{
			"nome":        "Dark Souls II",
			"igdb_id":     5001,
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogandoIGDB := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogandoIGDB)
		if recJogandoIGDB.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogando igdb: %d", recJogandoIGDB.Code)
		}

		corpoAbandonadoIGDB := map[string]any{
			"nome":         "Dark Souls 2 Scholar",
			"console":      "PS3",
			"igdb_id":      5001,
			"tempo_jogado": 5000,
		}
		recAb1 := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos-abandonados", token, corpoAbandonadoIGDB)
		if recAb1.Code != http.StatusCreated {
			t.Fatalf("falha ao criar abandonado igdb: %d body=%s", recAb1.Code, recAb1.Body.String())
		}

		recListar1 := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista1 respostaListaJogandoAgora
		_ = json.Unmarshal(recListar1.Body.Bytes(), &respLista1)
		if len(respLista1.Data) != 0 {
			t.Fatalf("esperava baixa por igdb_id no abandonado, restou %d itens", len(respLista1.Data))
		}

		corpoJogandoNome := map[string]any{
			"nome":        "  Ás do Espaço  ",
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogandoNome := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogandoNome)
		if recJogandoNome.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogando por nome: %d", recJogandoNome.Code)
		}

		corpoAbandonadoNome := map[string]any{
			"nome":         "as do espaco",
			"console":      "SNES",
			"tempo_jogado": 1200,
		}
		recAb2 := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos-abandonados", token, corpoAbandonadoNome)
		if recAb2.Code != http.StatusCreated {
			t.Fatalf("falha ao criar abandonado por nome: %d body=%s", recAb2.Code, recAb2.Body.String())
		}

		recListar2 := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista2 respostaListaJogandoAgora
		_ = json.Unmarshal(recListar2.Body.Bytes(), &respLista2)
		if len(respLista2.Data) != 0 {
			t.Fatalf("esperava baixa por nome normalizado no abandonado, restou %d itens", len(respLista2.Data))
		}

		corpoJogandoFica := map[string]any{
			"nome":        "Final Fantasy IX",
			"igdb_id":     6001,
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogandoFica := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogandoFica)
		if recJogandoFica.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogando fica: %d", recJogandoFica.Code)
		}

		corpoAbandonadoOutro := map[string]any{
			"nome":         "Dragon Quest XI",
			"console":      "PC",
			"igdb_id":      6002,
			"tempo_jogado": 3000,
		}
		recAb3 := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos-abandonados", token, corpoAbandonadoOutro)
		if recAb3.Code != http.StatusCreated {
			t.Fatalf("falha ao criar abandonado outro: %d body=%s", recAb3.Code, recAb3.Body.String())
		}

		recListar3 := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista3 respostaListaJogandoAgora
		_ = json.Unmarshal(recListar3.Body.Bytes(), &respLista3)
		if len(respLista3.Data) != 1 || respLista3.Data[0].Nome != "Final Fantasy IX" {
			t.Fatalf("item nao correspondente deveria continuar listado: %+v", respLista3.Data)
		}
	})

	t.Run("Falha de validacao no registro de zerado ou abandonado nao da baixa", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoJogando := map[string]any{
			"nome":        "Elden Ring",
			"igdb_id":     7001,
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogando := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoJogando)
		if recJogando.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo em andamento: %d", recJogando.Code)
		}

		corpoZeradoInvalido := map[string]any{
			"nome":               "Elden Ring",
			"igdb_id":            7001,
			"console":            "PC",
			"finalizado_em":      "",
			"tempo_jogado_horas": 10,
		}
		recZerado := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos", token, corpoZeradoInvalido)
		if recZerado.Code != http.StatusBadRequest {
			t.Fatalf("esperava status 400 em zerado invalido, obteve %d", recZerado.Code)
		}

		recListar1 := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista1 respostaListaJogandoAgora
		_ = json.Unmarshal(recListar1.Body.Bytes(), &respLista1)
		if len(respLista1.Data) != 1 || respLista1.Data[0].Nome != "Elden Ring" {
			t.Fatalf("jogo em andamento nao deveria sumir apos erro 400 de zerado: %+v", respLista1.Data)
		}

		corpoAbandonadoInvalido := map[string]any{
			"nome":               "Elden Ring",
			"igdb_id":            7001,
			"console":            "PC",
			"tempo_jogado_horas": -5,
		}
		recAbandonado := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos-abandonados", token, corpoAbandonadoInvalido)
		if recAbandonado.Code != http.StatusBadRequest {
			t.Fatalf("esperava status 400 em abandonado invalido, obteve %d", recAbandonado.Code)
		}

		recListar2 := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", token, nil)
		var respLista2 respostaListaJogandoAgora
		_ = json.Unmarshal(recListar2.Body.Bytes(), &respLista2)
		if len(respLista2.Data) != 1 || respLista2.Data[0].Nome != "Elden Ring" {
			t.Fatalf("jogo em andamento nao deveria sumir apos erro 400 de abandonado: %+v", respLista2.Data)
		}
	})

	t.Run("Isolamento entre usuarios na listagem, exclusao e baixa", func(t *testing.T) {
		usuarioA := criarUsuarioJogandoAgora(t, env.pool)
		usuarioB := criarUsuarioJogandoAgora(t, env.pool)
		tokenA := gerarTokenJogandoAgora(t, env.secret, usuarioA)
		tokenB := gerarTokenJogandoAgora(t, env.secret, usuarioB)

		corpoJogandoA := map[string]any{
			"nome":        "Sekiro Shadows Die Twice",
			"igdb_id":     8001,
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recJogandoA := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", tokenA, corpoJogandoA)
		if recJogandoA.Code != http.StatusCreated {
			t.Fatalf("falha ao criar jogo para usuario A: %d", recJogandoA.Code)
		}
		var respA respostaItemJogandoAgora
		_ = json.Unmarshal(recJogandoA.Body.Bytes(), &respA)
		jogoIDA := respA.Data.ID

		recListarB := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", tokenB, nil)
		if recListarB.Code != http.StatusOK {
			t.Fatalf("status listar B=%d", recListarB.Code)
		}
		var respListaB respostaListaJogandoAgora
		_ = json.Unmarshal(recListarB.Body.Bytes(), &respListaB)
		if len(respListaB.Data) != 0 {
			t.Fatalf("usuario B nao deveria ver jogo em andamento de A: %+v", respListaB.Data)
		}

		recDelB := fazerRequisicaoJogandoAgora(
			env.router,
			http.MethodDelete,
			fmt.Sprintf("/api/v1/jogando/%d", jogoIDA),
			tokenB,
			nil,
		)
		if recDelB.Code != http.StatusNotFound {
			t.Fatalf("esperava status 404 quando B tenta deletar jogo de A, obteve %d", recDelB.Code)
		}

		recListarA := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", tokenA, nil)
		var respListaA respostaListaJogandoAgora
		_ = json.Unmarshal(recListarA.Body.Bytes(), &respListaA)
		if len(respListaA.Data) != 1 || respListaA.Data[0].ID != jogoIDA {
			t.Fatalf("jogo de A deveria continuar listado apos tentativa indevida de delete: %+v", respListaA.Data)
		}

		corpoZeradoB := map[string]any{
			"nome":          "Sekiro Shadows Die Twice",
			"console":       "PC",
			"igdb_id":       8001,
			"finalizado_em": "2026-05-01T10:00:00Z",
			"tempo_jogado":  120000,
			"nota":          10,
			"dificuldade":   "A",
		}
		recZeradoB := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos", tokenB, corpoZeradoB)
		if recZeradoB.Code != http.StatusCreated {
			t.Fatalf("falha ao registrar zerado de B: %d", recZeradoB.Code)
		}

		recListarA2 := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", tokenA, nil)
		var respListaA2 respostaListaJogandoAgora
		_ = json.Unmarshal(recListarA2.Body.Bytes(), &respListaA2)
		if len(respListaA2.Data) != 1 || respListaA2.Data[0].ID != jogoIDA {
			t.Fatalf("jogo de A nao deveria receber baixa por zerado de B: %+v", respListaA2.Data)
		}

		corpoAbandonadoB := map[string]any{
			"nome":         "Sekiro Shadows Die Twice",
			"console":      "PS5",
			"igdb_id":      8001,
			"tempo_jogado": 5000,
		}
		recAbandonadoB := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogos-abandonados", tokenB, corpoAbandonadoB)
		if recAbandonadoB.Code != http.StatusCreated {
			t.Fatalf("falha ao registrar abandonado de B: %d", recAbandonadoB.Code)
		}

		recListarA3 := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", tokenA, nil)
		var respListaA3 respostaListaJogandoAgora
		_ = json.Unmarshal(recListarA3.Body.Bytes(), &respListaA3)
		if len(respListaA3.Data) != 1 || respListaA3.Data[0].ID != jogoIDA {
			t.Fatalf("jogo de A nao deveria receber baixa por abandonado de B: %+v", respListaA3.Data)
		}
	})

	t.Run("Validacoes de nome no cadastro de jogo em andamento", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoNomeVazio := map[string]any{
			"nome":        "",
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recVazio := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoNomeVazio)
		if recVazio.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 para nome vazio, obteve %d", recVazio.Code)
		}
		var errVazio respostaErroJogandoAgora
		_ = json.Unmarshal(recVazio.Body.Bytes(), &errVazio)
		if errVazio.Error.Codigo != "jogando.nome_obrigatorio" {
			t.Fatalf("codigo inesperado: %s", errVazio.Error.Codigo)
		}

		corpoNomeEspacos := map[string]any{
			"nome":        "     ",
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recEspacos := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoNomeEspacos)
		if recEspacos.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 para nome com apenas espacos, obteve %d", recEspacos.Code)
		}
		var errEspacos respostaErroJogandoAgora
		_ = json.Unmarshal(recEspacos.Body.Bytes(), &errEspacos)
		if errEspacos.Error.Codigo != "jogando.nome_obrigatorio" {
			t.Fatalf("codigo inesperado: %s", errEspacos.Error.Codigo)
		}

		corpoNomeLongo := map[string]any{
			"nome":        strings.Repeat("A", 201),
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recLongo := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoNomeLongo)
		if recLongo.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 para nome maior que 200 caracteres, obteve %d", recLongo.Code)
		}
		var errLongo respostaErroJogandoAgora
		_ = json.Unmarshal(recLongo.Body.Bytes(), &errLongo)
		if errLongo.Error.Codigo != "jogando.nome_muito_longo" {
			t.Fatalf("codigo inesperado: %s", errLongo.Error.Codigo)
		}
	})

	t.Run("Validacoes da data de inicio ausente, invalida, futura e hoje", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoSemData := map[string]any{
			"nome": "Super Mario World",
		}
		recSemData := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoSemData)
		if recSemData.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 para data ausente, obteve %d", recSemData.Code)
		}
		var errSemData respostaErroJogandoAgora
		_ = json.Unmarshal(recSemData.Body.Bytes(), &errSemData)
		if errSemData.Error.Codigo != "jogando.iniciado_em_obrigatorio" {
			t.Fatalf("codigo inesperado: %s", errSemData.Error.Codigo)
		}

		corpoDataVazia := map[string]any{
			"nome":        "Super Mario World",
			"iniciado_em": "   ",
		}
		recDataVazia := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoDataVazia)
		if recDataVazia.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 para data em branco, obteve %d", recDataVazia.Code)
		}
		var errDataVazia respostaErroJogandoAgora
		_ = json.Unmarshal(recDataVazia.Body.Bytes(), &errDataVazia)
		if errDataVazia.Error.Codigo != "jogando.iniciado_em_obrigatorio" {
			t.Fatalf("codigo inesperado: %s", errDataVazia.Error.Codigo)
		}

		corpoDataInvalida := map[string]any{
			"nome":        "Super Mario World",
			"iniciado_em": "data-formato-errado",
		}
		recInvalida := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoDataInvalida)
		if recInvalida.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 para data invalida, obteve %d", recInvalida.Code)
		}
		var errInvalida respostaErroJogandoAgora
		_ = json.Unmarshal(recInvalida.Body.Bytes(), &errInvalida)
		if errInvalida.Error.Codigo != "jogando.iniciado_em_invalido" {
			t.Fatalf("codigo inesperado: %s", errInvalida.Error.Codigo)
		}

		dataFutura := time.Now().Add(48 * time.Hour).Format("2006-01-02")
		corpoDataFutura := map[string]any{
			"nome":        "Super Mario World",
			"iniciado_em": dataFutura,
		}
		recFutura := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoDataFutura)
		if recFutura.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 para data futura, obteve %d body=%s", recFutura.Code, recFutura.Body.String())
		}
		var errFutura respostaErroJogandoAgora
		_ = json.Unmarshal(recFutura.Body.Bytes(), &errFutura)
		if errFutura.Error.Codigo != "jogando.iniciado_em_futuro" {
			t.Fatalf("codigo inesperado: %s", errFutura.Error.Codigo)
		}

		dataHoje := time.Now().UTC().Format("2006-01-02")
		corpoDataHoje := map[string]any{
			"nome":        "Super Mario World",
			"iniciado_em": dataHoje,
		}
		recHoje := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoDataHoje)
		if recHoje.Code != http.StatusCreated {
			t.Fatalf("esperava 201 para data de hoje (%s), obteve %d body=%s", dataHoje, recHoje.Code, recHoje.Body.String())
		}
	})

	t.Run("Internacionalizacao de mensagens de erro em pt-BR e en", func(t *testing.T) {
		usuarioID := criarUsuarioJogandoAgora(t, env.pool)
		token := gerarTokenJogandoAgora(t, env.secret, usuarioID)

		corpoInvalido := map[string]any{
			"nome":        "",
			"iniciado_em": "2026-01-01T00:00:00Z",
		}

		recPT := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoInvalido, "pt-BR")
		if recPT.Code != http.StatusBadRequest {
			t.Fatalf("status pt=%d", recPT.Code)
		}
		var errPT respostaErroJogandoAgora
		_ = json.Unmarshal(recPT.Body.Bytes(), &errPT)
		if errPT.Error.Codigo != "jogando.nome_obrigatorio" || errPT.Error.Mensagem != "O nome do jogo é obrigatório." {
			t.Fatalf("mensagem pt-BR inesperada: %+v", errPT.Error)
		}

		recEN := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", token, corpoInvalido, "en")
		if recEN.Code != http.StatusBadRequest {
			t.Fatalf("status en=%d", recEN.Code)
		}
		var errEN respostaErroJogandoAgora
		_ = json.Unmarshal(recEN.Body.Bytes(), &errEN)
		if errEN.Error.Codigo != "jogando.nome_obrigatorio" || errEN.Error.Mensagem != "Game name is required." {
			t.Fatalf("mensagem en inesperada: %+v", errEN.Error)
		}
	})

	t.Run("Autenticacao obrigatoria nos endpoints de jogos em andamento", func(t *testing.T) {
		recGet := fazerRequisicaoJogandoAgora(env.router, http.MethodGet, "/api/v1/jogando", "", nil)
		if recGet.Code != http.StatusUnauthorized {
			t.Fatalf("esperava 401 no GET sem token, obteve %d", recGet.Code)
		}
		var errGet respostaErroJogandoAgora
		_ = json.Unmarshal(recGet.Body.Bytes(), &errGet)
		if errGet.Error.Codigo != "auth.session.unauthorized" {
			t.Fatalf("codigo inesperado no GET sem token: %s", errGet.Error.Codigo)
		}

		corpoCriar := map[string]any{
			"nome":        "Jogo Nao Autorizado",
			"iniciado_em": "2026-01-01T00:00:00Z",
		}
		recPost := fazerRequisicaoJogandoAgora(env.router, http.MethodPost, "/api/v1/jogando", "", corpoCriar)
		if recPost.Code != http.StatusUnauthorized {
			t.Fatalf("esperava 401 no POST sem token, obteve %d", recPost.Code)
		}
		var errPost respostaErroJogandoAgora
		_ = json.Unmarshal(recPost.Body.Bytes(), &errPost)
		if errPost.Error.Codigo != "auth.session.unauthorized" {
			t.Fatalf("codigo inesperado no POST sem token: %s", errPost.Error.Codigo)
		}

		recDel := fazerRequisicaoJogandoAgora(env.router, http.MethodDelete, "/api/v1/jogando/1", "", nil)
		if recDel.Code != http.StatusUnauthorized {
			t.Fatalf("esperava 401 no DELETE sem token, obteve %d", recDel.Code)
		}
		var errDel respostaErroJogandoAgora
		_ = json.Unmarshal(recDel.Body.Bytes(), &errDel)
		if errDel.Error.Codigo != "auth.session.unauthorized" {
			t.Fatalf("codigo inesperado no DELETE sem token: %s", errDel.Error.Codigo)
		}
	})
}
