//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

type postgresIntegrationEnv struct {
	pool    *pgxpool.Pool
	router  *gin.Engine
	tokens  *service.AuthToken
	repo    repository.JogosRepository
	service *service.JogosService
	handler *handler.JogosHandler
	secret  string
}

type criarJogoInput struct {
	Nome                string  `json:"nome"`
	Console             string  `json:"console"`
	Genero              string  `json:"genero,omitempty"`
	Tipo                string  `json:"tipo,omitempty"`
	IniciadoEm          *string `json:"iniciado_em,omitempty"`
	FinalizadoEm        string  `json:"finalizado_em"`
	TempoJogadoHoras    int     `json:"tempo_jogado_horas,omitempty"`
	TempoJogadoMinutos  int     `json:"tempo_jogado_minutos,omitempty"`
	TempoJogadoSegundos int     `json:"tempo_jogado_segundos,omitempty"`
	TempoJogado         *int32  `json:"tempo_jogado,omitempty"`
	Nota                int32   `json:"nota"`
	Dificuldade         string  `json:"dificuldade"`
	Review              string  `json:"review,omitempty"`
	Destaque            bool    `json:"destaque"`
	IgdbDescricao       string  `json:"igdb_descricao,omitempty"`
	IgdbCapaURL         string  `json:"igdb_capa_url,omitempty"`
}

type jogoEnvelope struct {
	Data repository.JogoZerado `json:"data"`
}

type listagemMeta struct {
	Pagina       int   `json:"pagina"`
	PorPagina    int   `json:"por_pagina"`
	Total        int64 `json:"total"`
	TotalPaginas int   `json:"total_paginas"`
}

type listagemEnvelope struct {
	Data []repository.JogoZerado `json:"data"`
	Meta listagemMeta            `json:"meta"`
}

type filtrosEnvelope struct {
	Data repository.OpcoesFiltros `json:"data"`
}

type erroEnvelope struct {
	Error struct {
		Codigo   string `json:"codigo"`
		Mensagem string `json:"mensagem"`
	} `json:"error"`
}

func setupIntegrationEnv(t *testing.T) *postgresIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_integration"),
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
		if err := pgContainer.Terminate(teardownCtx); err != nil {
			t.Logf("aviso: falha ao terminar container: %v", err)
		}
	})

	queries := db.New(pool)
	repo := repository.NewJogosRepository(queries)
	svc := service.NewJogosService(repo)
	h := handler.NewJogosHandler(svc)

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.POST("/jogos", h.CriarJogo)
	privadas.PUT("/jogos/:id", h.AtualizarJogo)
	privadas.DELETE("/jogos/:id", h.ExcluirJogo)
	privadas.GET("/jogos", h.ListarJogos)
	privadas.GET("/jogos/filtros", h.ObterFiltros)
	privadas.GET("/jogos/:id", h.ObterDetalhesJogo)

	return &postgresIntegrationEnv{
		pool:    pool,
		router:  router,
		tokens:  tokens,
		repo:    repo,
		service: svc,
		handler: h,
		secret:  secret,
	}
}

func criarUsuarioTeste(t *testing.T, pool *pgxpool.Pool) int32 {
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

func gerarAccessTokenTeste(t *testing.T, secret string, usuarioID int32) string {
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

func TestIntegration_JogosZerados(t *testing.T) {
	env := setupIntegrationEnv(t)

	t.Run("1_Criar_Editar_Excluir", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		inputCriar := criarJogoInput{
			Nome:                "Super Mario World",
			Console:             "SNES",
			Genero:              "Platformer",
			Tipo:                "Zeramento Normal",
			FinalizadoEm:        "2026-05-10T14:30:00Z",
			TempoJogadoHoras:    10,
			TempoJogadoMinutos:  30,
			TempoJogadoSegundos: 0,
			Nota:                10,
			Dificuldade:         "A",
			Review:              "Campanha inicial excelente.",
			Destaque:            false,
		}
		bodyCriar, _ := json.Marshal(inputCriar)

		reqCriar := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyCriar))
		reqCriar.Header.Set("Authorization", "Bearer "+token)
		reqCriar.Header.Set("Content-Type", "application/json")
		wCriar := httptest.NewRecorder()
		env.router.ServeHTTP(wCriar, reqCriar)

		if wCriar.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 na criacao, obteve %d: %s", wCriar.Code, wCriar.Body.String())
		}

		var respCriar jogoEnvelope
		if err := json.Unmarshal(wCriar.Body.Bytes(), &respCriar); err != nil {
			t.Fatalf("falha ao decodificar resposta da criacao: %v", err)
		}
		jogoID := respCriar.Data.ID
		if jogoID <= 0 {
			t.Fatalf("ID de jogo invalido retornado: %d", jogoID)
		}

		inputEditar := criarJogoInput{
			Nome:                "Super Mario World Editado",
			Console:             "SNES Classic",
			Genero:              "Platformer 2D",
			Tipo:                "100%",
			FinalizadoEm:        "2026-05-15T18:00:00Z",
			TempoJogadoHoras:    15,
			TempoJogadoMinutos:  45,
			TempoJogadoSegundos: 30,
			Nota:                11,
			Dificuldade:         "AA",
			Review:              "Review atualizada com todos os 96 segredos.",
			Destaque:            false,
		}
		bodyEditar, _ := json.Marshal(inputEditar)

		reqEditar := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jogoID), bytes.NewReader(bodyEditar))
		reqEditar.Header.Set("Authorization", "Bearer "+token)
		reqEditar.Header.Set("Content-Type", "application/json")
		wEditar := httptest.NewRecorder()
		env.router.ServeHTTP(wEditar, reqEditar)

		if wEditar.Code != http.StatusOK {
			t.Fatalf("esperava status 200 na edicao, obteve %d: %s", wEditar.Code, wEditar.Body.String())
		}

		var respEditar jogoEnvelope
		if err := json.Unmarshal(wEditar.Body.Bytes(), &respEditar); err != nil {
			t.Fatalf("falha ao decodificar resposta da edicao: %v", err)
		}
		if respEditar.Data.Nome != "Super Mario World Editado" || respEditar.Data.Nota != 11 {
			t.Fatalf("valores da resposta da edicao incorretos: %+v", respEditar.Data)
		}

		var dbNome, dbReview string
		var dbNota int32
		var dbDeletedAt *time.Time
		err := env.pool.QueryRow(context.Background(),
			"SELECT nome, review, nota, deleted_at FROM jogos_zerados WHERE id = $1",
			jogoID,
		).Scan(&dbNome, &dbReview, &dbNota, &dbDeletedAt)
		if err != nil {
			t.Fatalf("falha ao consultar banco apos edicao: %v", err)
		}
		if dbNome != "Super Mario World Editado" || dbReview != "Review atualizada com todos os 96 segredos." || dbNota != 11 {
			t.Fatalf("banco nao reflete edicao: nome=%s, review=%s, nota=%d", dbNome, dbReview, dbNota)
		}
		if dbDeletedAt != nil {
			t.Fatalf("deleted_at nao deveria estar preenchido antes do delete: %v", dbDeletedAt)
		}

		reqExcluir := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqExcluir.Header.Set("Authorization", "Bearer "+token)
		wExcluir := httptest.NewRecorder()
		env.router.ServeHTTP(wExcluir, reqExcluir)

		if wExcluir.Code != http.StatusNoContent {
			t.Fatalf("esperava status 204 na exclusao, obteve %d: %s", wExcluir.Code, wExcluir.Body.String())
		}

		var dbDeletedAtAfter *time.Time
		err = env.pool.QueryRow(context.Background(),
			"SELECT deleted_at FROM jogos_zerados WHERE id = $1",
			jogoID,
		).Scan(&dbDeletedAtAfter)
		if err != nil {
			t.Fatalf("registro nao encontrado no banco apos soft delete: %v", err)
		}
		if dbDeletedAtAfter == nil {
			t.Fatal("deleted_at deveria estar preenchido apos soft delete")
		}
	})

	t.Run("2_Registro_Excluido_Retorna_404_Em_PUT_E_DELETE", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		input := criarJogoInput{
			Nome:         "Castlevania",
			Console:      "NES",
			FinalizadoEm: "2026-01-10T12:00:00Z",
			Nota:         9,
			Dificuldade:  "B",
		}
		body, _ := json.Marshal(input)

		reqCriar := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body))
		reqCriar.Header.Set("Authorization", "Bearer "+token)
		reqCriar.Header.Set("Content-Type", "application/json")
		wCriar := httptest.NewRecorder()
		env.router.ServeHTTP(wCriar, reqCriar)

		if wCriar.Code != http.StatusCreated {
			t.Fatalf("status criacao=%d", wCriar.Code)
		}
		var respCriar jogoEnvelope
		_ = json.Unmarshal(wCriar.Body.Bytes(), &respCriar)
		jogoID := respCriar.Data.ID

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("status delete=%d", wDel.Code)
		}

		reqPut := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jogoID), bytes.NewReader(body))
		reqPut.Header.Set("Authorization", "Bearer "+token)
		reqPut.Header.Set("Content-Type", "application/json")
		wPut := httptest.NewRecorder()
		env.router.ServeHTTP(wPut, reqPut)

		if wPut.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em PUT de registro excluido, obteve %d", wPut.Code)
		}
		var errPut erroEnvelope
		_ = json.Unmarshal(wPut.Body.Bytes(), &errPut)
		if errPut.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo jogos.not_found, obteve %s", errPut.Error.Codigo)
		}

		reqDel2 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqDel2.Header.Set("Authorization", "Bearer "+token)
		wDel2 := httptest.NewRecorder()
		env.router.ServeHTTP(wDel2, reqDel2)

		if wDel2.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em DELETE de registro excluido, obteve %d", wDel2.Code)
		}
		var errDel erroEnvelope
		_ = json.Unmarshal(wDel2.Body.Bytes(), &errDel)
		if errDel.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo jogos.not_found, obteve %s", errDel.Error.Codigo)
		}
	})

	t.Run("3_Destaque_Conflito_Mesmo_Ano_Mesmo_Usuario", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		input1 := criarJogoInput{
			Nome:         "Zelda I",
			Console:      "NES",
			FinalizadoEm: "2026-03-01T12:00:00Z",
			Nota:         10,
			Dificuldade:  "A",
			Destaque:     true,
		}
		body1, _ := json.Marshal(input1)

		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		env.router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("esperava 201 no primeiro destaque, obteve %d: %s", w1.Code, w1.Body.String())
		}

		input2 := criarJogoInput{
			Nome:         "Zelda II",
			Console:      "NES",
			FinalizadoEm: "2026-11-20T12:00:00Z",
			Nota:         9,
			Dificuldade:  "AA",
			Destaque:     true,
		}
		body2, _ := json.Marshal(input2)

		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		env.router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusConflict {
			t.Fatalf("esperava 409 no segundo destaque do mesmo ano, obteve %d: %s", w2.Code, w2.Body.String())
		}

		var resp erroEnvelope
		if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
			t.Fatalf("falha ao decodificar erro 409: %v", err)
		}
		if resp.Error.Codigo != "jogos.destaque_ano_conflito" {
			t.Fatalf("esperava codigo jogos.destaque_ano_conflito, obteve %s", resp.Error.Codigo)
		}
	})

	t.Run("4_Destaque_Anos_Diferentes_Ambos_201", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		input1 := criarJogoInput{
			Nome:         "Chrono Trigger",
			Console:      "SNES",
			FinalizadoEm: "2024-06-10T12:00:00Z",
			Nota:         10,
			Dificuldade:  "A",
			Destaque:     true,
		}
		body1, _ := json.Marshal(input1)

		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		env.router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("esperava 201 no destaque de 2024, obteve %d", w1.Code)
		}

		input2 := criarJogoInput{
			Nome:         "Final Fantasy VI",
			Console:      "SNES",
			FinalizadoEm: "2025-06-10T12:00:00Z",
			Nota:         10,
			Dificuldade:  "A",
			Destaque:     true,
		}
		body2, _ := json.Marshal(input2)

		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		env.router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusCreated {
			t.Fatalf("esperava 201 no destaque de 2025, obteve %d", w2.Code)
		}
	})

	t.Run("5_Destaque_Mesmo_Ano_Usuarios_Diferentes_Ambos_201", func(t *testing.T) {
		userA := criarUsuarioTeste(t, env.pool)
		userB := criarUsuarioTeste(t, env.pool)
		tokenA := gerarAccessTokenTeste(t, env.secret, userA)
		tokenB := gerarAccessTokenTeste(t, env.secret, userB)

		inputA := criarJogoInput{
			Nome:         "Metroid Prime",
			Console:      "GameCube",
			FinalizadoEm: "2026-08-01T12:00:00Z",
			Nota:         10,
			Dificuldade:  "AA",
			Destaque:     true,
		}
		bodyA, _ := json.Marshal(inputA)

		reqA := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyA))
		reqA.Header.Set("Authorization", "Bearer "+tokenA)
		reqA.Header.Set("Content-Type", "application/json")
		wA := httptest.NewRecorder()
		env.router.ServeHTTP(wA, reqA)

		if wA.Code != http.StatusCreated {
			t.Fatalf("esperava 201 para usuario A, obteve %d", wA.Code)
		}

		inputB := criarJogoInput{
			Nome:         "Resident Evil 4",
			Console:      "GameCube",
			FinalizadoEm: "2026-09-01T12:00:00Z",
			Nota:         11,
			Dificuldade:  "AAA",
			Destaque:     true,
		}
		bodyB, _ := json.Marshal(inputB)

		reqB := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyB))
		reqB.Header.Set("Authorization", "Bearer "+tokenB)
		reqB.Header.Set("Content-Type", "application/json")
		wB := httptest.NewRecorder()
		env.router.ServeHTTP(wB, reqB)

		if wB.Code != http.StatusCreated {
			t.Fatalf("esperava 201 para usuario B no mesmo ano, obteve %d", wB.Code)
		}
	})

	t.Run("6_Excluir_Destaque_Libera_Ano_Para_Novo_Destaque", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		input1 := criarJogoInput{
			Nome:         "Pokemon Red",
			Console:      "Game Boy",
			FinalizadoEm: "2023-02-10T12:00:00Z",
			Nota:         9,
			Dificuldade:  "B",
			Destaque:     true,
		}
		body1, _ := json.Marshal(input1)

		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		env.router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("esperava 201 no primeiro destaque, obteve %d", w1.Code)
		}
		var resp1 jogoEnvelope
		_ = json.Unmarshal(w1.Body.Bytes(), &resp1)
		jogoID := resp1.Data.ID

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)

		if wDel.Code != http.StatusNoContent {
			t.Fatalf("esperava 204 na exclusao do destaque, obteve %d", wDel.Code)
		}

		input2 := criarJogoInput{
			Nome:         "Pokemon Blue",
			Console:      "Game Boy",
			FinalizadoEm: "2023-11-20T12:00:00Z",
			Nota:         10,
			Dificuldade:  "B",
			Destaque:     true,
		}
		body2, _ := json.Marshal(input2)

		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		env.router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusCreated {
			t.Fatalf("esperava 201 ao cadastrar novo destaque no mesmo ano apos exclusao, obteve %d: %s", w2.Code, w2.Body.String())
		}
	})

	t.Run("7_Isolamento_Usuario_B_Nao_Edita_Nem_Exclui_Registro_De_Usuario_A", func(t *testing.T) {
		userA := criarUsuarioTeste(t, env.pool)
		userB := criarUsuarioTeste(t, env.pool)
		tokenA := gerarAccessTokenTeste(t, env.secret, userA)
		tokenB := gerarAccessTokenTeste(t, env.secret, userB)

		inputA := criarJogoInput{
			Nome:         "Silent Hill",
			Console:      "PlayStation",
			FinalizadoEm: "2026-07-04T12:00:00Z",
			Nota:         10,
			Dificuldade:  "AA",
		}
		bodyA, _ := json.Marshal(inputA)

		reqA := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyA))
		reqA.Header.Set("Authorization", "Bearer "+tokenA)
		reqA.Header.Set("Content-Type", "application/json")
		wA := httptest.NewRecorder()
		env.router.ServeHTTP(wA, reqA)

		if wA.Code != http.StatusCreated {
			t.Fatalf("esperava 201 na criacao de user A, obteve %d", wA.Code)
		}
		var respA jogoEnvelope
		_ = json.Unmarshal(wA.Body.Bytes(), &respA)
		jogoID := respA.Data.ID

		inputHack := criarJogoInput{
			Nome:         "Silent Hill Modificado por B",
			Console:      "PlayStation",
			FinalizadoEm: "2026-07-04T12:00:00Z",
			Nota:         1,
			Dificuldade:  "C",
		}
		bodyHack, _ := json.Marshal(inputHack)

		reqPutB := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jogoID), bytes.NewReader(bodyHack))
		reqPutB.Header.Set("Authorization", "Bearer "+tokenB)
		reqPutB.Header.Set("Content-Type", "application/json")
		wPutB := httptest.NewRecorder()
		env.router.ServeHTTP(wPutB, reqPutB)

		if wPutB.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 quando usuario B tenta editar registro de A, obteve %d", wPutB.Code)
		}
		var errPut erroEnvelope
		_ = json.Unmarshal(wPutB.Body.Bytes(), &errPut)
		if errPut.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo jogos.not_found, obteve %s", errPut.Error.Codigo)
		}

		reqDelB := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqDelB.Header.Set("Authorization", "Bearer "+tokenB)
		wDelB := httptest.NewRecorder()
		env.router.ServeHTTP(wDelB, reqDelB)

		if wDelB.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 quando usuario B tenta excluir registro de A, obteve %d", wDelB.Code)
		}
		var errDel erroEnvelope
		_ = json.Unmarshal(wDelB.Body.Bytes(), &errDel)
		if errDel.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo jogos.not_found, obteve %s", errDel.Error.Codigo)
		}

		var dbNome string
		var dbDeletedAt *time.Time
		err := env.pool.QueryRow(context.Background(),
			"SELECT nome, deleted_at FROM jogos_zerados WHERE id = $1",
			jogoID,
		).Scan(&dbNome, &dbDeletedAt)
		if err != nil {
			t.Fatalf("falha ao consultar registro original de A: %v", err)
		}
		if dbNome != "Silent Hill" || dbDeletedAt != nil {
			t.Fatalf("registro de A foi alterado: nome=%s, deleted_at=%v", dbNome, dbDeletedAt)
		}
	})

	t.Run("8_Nota_12_Rejeitada_Pelo_CHECK_Do_Banco_Via_Repository", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)

		_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo Nota Invalida DB",
			Console:      "Mega Drive",
			FinalizadoEm: time.Now(),
			TempoJogado:  3600,
			Nota:         12,
			Dificuldade:  "A",
		})
		if err == nil {
			t.Fatal("esperava erro de violacao de CHECK constraint ao enviar nota 12 ao repository")
		}

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			t.Fatalf("esperava erro do tipo *pgconn.PgError, obteve: %v", err)
		}
		if pgErr.Code != "23514" {
			t.Fatalf("esperava codigo postgres 23514 (check_violation), obteve: %s", pgErr.Code)
		}
	})

	t.Run("9_Ordenacao_FinalizadoEm_Desc_Id_Desc", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		mesmoMomento := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
		momentoAnterior := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

		j1, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo Antigo",
			Console:      "SNES",
			FinalizadoEm: momentoAnterior,
			TempoJogado:  3600,
			Nota:         8,
			Dificuldade:  "B",
		})
		if err != nil {
			t.Fatal(err)
		}

		j2, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo Mesmo Momento 1",
			Console:      "SNES",
			FinalizadoEm: mesmoMomento,
			TempoJogado:  3600,
			Nota:         9,
			Dificuldade:  "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		j3, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo Mesmo Momento 2",
			Console:      "SNES",
			FinalizadoEm: mesmoMomento,
			TempoJogado:  3600,
			Nota:         10,
			Dificuldade:  "AA",
		})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}

		var resp listagemEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}

		if len(resp.Data) != 3 {
			t.Fatalf("esperava 3 jogos, obteve %d", len(resp.Data))
		}
		if resp.Data[0].ID != j3.ID || resp.Data[1].ID != j2.ID || resp.Data[2].ID != j1.ID {
			t.Fatalf("ordem incorreta: obteve [%d, %d, %d], esperava [%d, %d, %d]",
				resp.Data[0].ID, resp.Data[1].ID, resp.Data[2].ID, j3.ID, j2.ID, j1.ID)
		}
	})

	t.Run("10_Paginacao_30_Registros", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		for i := 1; i <= 30; i++ {
			_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
				UsuarioID:    usuarioID,
				Nome:         fmt.Sprintf("Jogo Paginado %02d", i),
				Console:      "SNES",
				FinalizadoEm: baseTime.Add(time.Duration(i) * time.Hour),
				TempoJogado:  1000,
				Nota:         8,
				Dificuldade:  "A",
			})
			if err != nil {
				t.Fatalf("falha ao criar jogo %d: %v", i, err)
			}
		}

		reqP1 := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?pagina=1&por_pagina=24", nil)
		reqP1.Header.Set("Authorization", "Bearer "+token)
		wP1 := httptest.NewRecorder()
		env.router.ServeHTTP(wP1, reqP1)
		if wP1.Code != http.StatusOK {
			t.Fatalf("pagina 1 esperava 200, obteve %d: %s", wP1.Code, wP1.Body.String())
		}
		var respP1 listagemEnvelope
		_ = json.Unmarshal(wP1.Body.Bytes(), &respP1)
		if len(respP1.Data) != 24 || respP1.Meta.Total != 30 || respP1.Meta.TotalPaginas != 2 || respP1.Meta.Pagina != 1 || respP1.Meta.PorPagina != 24 {
			t.Fatalf("pagina 1 com dados ou meta incorreto: %+v", respP1.Meta)
		}

		reqP2 := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?pagina=2&por_pagina=24", nil)
		reqP2.Header.Set("Authorization", "Bearer "+token)
		wP2 := httptest.NewRecorder()
		env.router.ServeHTTP(wP2, reqP2)
		if wP2.Code != http.StatusOK {
			t.Fatalf("pagina 2 esperava 200, obteve %d: %s", wP2.Code, wP2.Body.String())
		}
		var respP2 listagemEnvelope
		_ = json.Unmarshal(wP2.Body.Bytes(), &respP2)
		if len(respP2.Data) != 6 || respP2.Meta.Total != 30 || respP2.Meta.TotalPaginas != 2 || respP2.Meta.Pagina != 2 || respP2.Meta.PorPagina != 24 {
			t.Fatalf("pagina 2 com dados ou meta incorreto: len=%d meta=%+v", len(respP2.Data), respP2.Meta)
		}

		reqP3 := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?pagina=3&por_pagina=24", nil)
		reqP3.Header.Set("Authorization", "Bearer "+token)
		wP3 := httptest.NewRecorder()
		env.router.ServeHTTP(wP3, reqP3)
		if wP3.Code != http.StatusOK {
			t.Fatalf("pagina 3 esperava 200, obteve %d: %s", wP3.Code, wP3.Body.String())
		}
		var respP3 listagemEnvelope
		_ = json.Unmarshal(wP3.Body.Bytes(), &respP3)
		if len(respP3.Data) != 0 || respP3.Meta.Total != 30 || respP3.Meta.TotalPaginas != 2 || respP3.Meta.Pagina != 3 || respP3.Meta.PorPagina != 24 {
			t.Fatalf("pagina 3 com dados ou meta incorreto: len=%d meta=%+v", len(respP3.Data), respP3.Meta)
		}
	})

	t.Run("11_SoftDelete_ExcluidoNaoApareceEmListagemNemFiltros", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		jExcluir, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo a Excluir",
			Console:      "ConsoleUnicoExcluido",
			Genero:       "GeneroUnicoExcluido",
			Tipo:         "TipoUnicoExcluido",
			FinalizadoEm: time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			TempoJogado:  1000,
			Nota:         7,
			Dificuldade:  "B",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo Que Fica",
			Console:      "ConsoleAtivo",
			Genero:       "GeneroAtivo",
			Tipo:         "TipoAtivo",
			FinalizadoEm: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			TempoJogado:  2000,
			Nota:         9,
			Dificuldade:  "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jExcluir.ID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("esperava 204 ao excluir, obteve %d", wDel.Code)
		}

		reqListar := httptest.NewRequest(http.MethodGet, "/api/v1/jogos", nil)
		reqListar.Header.Set("Authorization", "Bearer "+token)
		wListar := httptest.NewRecorder()
		env.router.ServeHTTP(wListar, reqListar)
		var respListar listagemEnvelope
		_ = json.Unmarshal(wListar.Body.Bytes(), &respListar)
		if len(respListar.Data) != 1 || respListar.Data[0].Nome != "Jogo Que Fica" {
			t.Fatalf("listagem contem itens excluidos ou incorretos: %+v", respListar.Data)
		}

		reqFiltros := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/filtros", nil)
		reqFiltros.Header.Set("Authorization", "Bearer "+token)
		wFiltros := httptest.NewRecorder()
		env.router.ServeHTTP(wFiltros, reqFiltros)
		var respFiltros filtrosEnvelope
		_ = json.Unmarshal(wFiltros.Body.Bytes(), &respFiltros)

		for _, c := range respFiltros.Data.Consoles {
			if c == "ConsoleUnicoExcluido" {
				t.Fatal("console de registro excluido apareceu nos filtros")
			}
		}
		for _, g := range respFiltros.Data.Generos {
			if g == "GeneroUnicoExcluido" {
				t.Fatal("genero de registro excluido apareceu nos filtros")
			}
		}
		for _, tip := range respFiltros.Data.Tipos {
			if tip == "TipoUnicoExcluido" {
				t.Fatal("tipo de registro excluido apareceu nos filtros")
			}
		}
		for _, a := range respFiltros.Data.Anos {
			if a == 2023 {
				t.Fatal("ano de registro excluido apareceu nos filtros")
			}
		}
	})

	t.Run("12_Isolamento_UsuarioB_NaoVe_RegistrosUsuarioA", func(t *testing.T) {
		usuarioA := criarUsuarioTeste(t, env.pool)
		usuarioB := criarUsuarioTeste(t, env.pool)
		tokenB := gerarAccessTokenTeste(t, env.secret, usuarioB)

		_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioA,
			Nome:         "Exclusivo Usuario A",
			Console:      "Console A",
			Genero:       "Genero A",
			Tipo:         "Tipo A",
			FinalizadoEm: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			TempoJogado:  1000,
			Nota:         10,
			Dificuldade:  "AAA",
		})
		if err != nil {
			t.Fatal(err)
		}

		reqListar := httptest.NewRequest(http.MethodGet, "/api/v1/jogos", nil)
		reqListar.Header.Set("Authorization", "Bearer "+tokenB)
		wListar := httptest.NewRecorder()
		env.router.ServeHTTP(wListar, reqListar)
		var respListar listagemEnvelope
		_ = json.Unmarshal(wListar.Body.Bytes(), &respListar)
		if len(respListar.Data) != 0 || respListar.Meta.Total != 0 {
			t.Fatalf("usuario B viu dados do usuario A na listagem: %+v", respListar)
		}

		reqFiltros := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/filtros", nil)
		reqFiltros.Header.Set("Authorization", "Bearer "+tokenB)
		wFiltros := httptest.NewRecorder()
		env.router.ServeHTTP(wFiltros, reqFiltros)
		var respFiltros filtrosEnvelope
		_ = json.Unmarshal(wFiltros.Body.Bytes(), &respFiltros)
		if len(respFiltros.Data.Consoles) != 0 || len(respFiltros.Data.Generos) != 0 || len(respFiltros.Data.Tipos) != 0 || len(respFiltros.Data.Anos) != 0 {
			t.Fatalf("usuario B viu filtros do usuario A: %+v", respFiltros.Data)
		}
	})

	t.Run("13_Unaccent_BuscaCaseInsensitiveESemAcento", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		now := time.Now()
		_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Pokémon Red", Console: "Game Boy", FinalizadoEm: now, TempoJogado: 1000, Nota: 10, Dificuldade: "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Pokemon Stadium", Console: "N64", FinalizadoEm: now, TempoJogado: 1000, Nota: 8, Dificuldade: "B",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "The Legend of Zelda", Console: "NES", FinalizadoEm: now, TempoJogado: 1000, Nota: 9, Dificuldade: "AA",
		})
		if err != nil {
			t.Fatal(err)
		}

		verificarBusca := func(termo string, quantidadeEsperada int) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?busca="+termo, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("busca %s retornou %d: %s", termo, w.Code, w.Body.String())
			}
			var resp listagemEnvelope
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if len(resp.Data) != quantidadeEsperada {
				t.Fatalf("busca %q esperava %d resultados, obteve %d", termo, quantidadeEsperada, len(resp.Data))
			}
		}

		verificarBusca("pokemon", 2)
		verificarBusca("Pokémon", 2)
		verificarBusca("ZELDA", 1)
	})

	t.Run("14_Escape_Curinga_BuscaLiteralPorPorcento", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		now := time.Now()
		_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "100% Orange Juice", Console: "PC", FinalizadoEm: now, TempoJogado: 1000, Nota: 8, Dificuldade: "B",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Super Mario World", Console: "SNES", FinalizadoEm: now, TempoJogado: 1000, Nota: 10, Dificuldade: "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?busca=%25", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}
		var resp listagemEnvelope
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Data) != 1 || resp.Data[0].Nome != "100% Orange Juice" {
			t.Fatalf("busca por '%%' deveria retornar apenas '100%% Orange Juice', obteve: %+v", resp.Data)
		}
	})

	t.Run("15_GeneroComposto", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		now := time.Now()
		_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Smash Bros", Console: "N64", Genero: "Fighting, Platform, Hack and slash", FinalizadoEm: now, TempoJogado: 1000, Nota: 9, Dificuldade: "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?genero=platform", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}
		var resp listagemEnvelope
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Data) != 1 || resp.Data[0].Nome != "Smash Bros" {
			t.Fatalf("filtro por 'platform' nao encontrou genero composto: %+v", resp.Data)
		}
	})

	t.Run("16_CombinacaoDeFiltros", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		data2025 := time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)
		data2024 := time.Date(2024, 7, 10, 0, 0, 0, 0, time.UTC)

		jMatch, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Match Total", Console: "SNES", Tipo: "Campanha", FinalizadoEm: data2025, TempoJogado: 1000, Nota: 10, Dificuldade: "AA",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, _ = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Console Errado", Console: "NES", Tipo: "Campanha", FinalizadoEm: data2025, TempoJogado: 1000, Nota: 10, Dificuldade: "AA",
		})

		_, _ = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Nota Baixa", Console: "SNES", Tipo: "Campanha", FinalizadoEm: data2025, TempoJogado: 1000, Nota: 6, Dificuldade: "AA",
		})

		_, _ = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Ano Errado", Console: "SNES", Tipo: "Campanha", FinalizadoEm: data2024, TempoJogado: 1000, Nota: 10, Dificuldade: "AA",
		})

		_, _ = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Dificuldade Errada", Console: "SNES", Tipo: "Campanha", FinalizadoEm: data2025, TempoJogado: 1000, Nota: 10, Dificuldade: "C",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?console=SNES&nota_min=8&nota_max=10&ano=2025&dificuldade=AA", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}
		var resp listagemEnvelope
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Data) != 1 || resp.Data[0].ID != jMatch.ID {
			t.Fatalf("combinacao de filtros retornou dados inesperados: %+v", resp.Data)
		}
	})

	t.Run("17_Filtros_Unicos_Limpos_Ordenados", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo 1",
			Console:      "SNES",
			Genero:       "Action, Adventure",
			Tipo:         "Campanha",
			FinalizadoEm: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			TempoJogado:  1000,
			Nota:         8,
			Dificuldade:  "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo 2",
			Console:      "PlayStation 5",
			Genero:       "Adventure, RPG",
			Tipo:         "DLC",
			FinalizadoEm: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			TempoJogado:  1000,
			Nota:         9,
			Dificuldade:  "AA",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo 3",
			Console:      "Game Boy Advance",
			Genero:       "RPG",
			Tipo:         "100%",
			FinalizadoEm: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			TempoJogado:  1000,
			Nota:         10,
			Dificuldade:  "AAA",
		})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/filtros", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}
		var resp filtrosEnvelope
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		if len(resp.Data.Consoles) != 3 || resp.Data.Consoles[0] != "Game Boy Advance" || resp.Data.Consoles[1] != "PlayStation 5" || resp.Data.Consoles[2] != "SNES" {
			t.Fatalf("consoles ordenados incorretamente: %+v", resp.Data.Consoles)
		}

		if len(resp.Data.Generos) != 3 || resp.Data.Generos[0] != "Action" || resp.Data.Generos[1] != "Adventure" || resp.Data.Generos[2] != "RPG" {
			t.Fatalf("generos deduplicados e ordenados incorretamente: %+v", resp.Data.Generos)
		}

		if len(resp.Data.Tipos) != 3 || resp.Data.Tipos[0] != "100%" || resp.Data.Tipos[1] != "Campanha" || resp.Data.Tipos[2] != "DLC" {
			t.Fatalf("tipos ordenados incorretamente: %+v", resp.Data.Tipos)
		}

		if len(resp.Data.Anos) != 3 || resp.Data.Anos[0] != 2026 || resp.Data.Anos[1] != 2025 || resp.Data.Anos[2] != 2024 {
			t.Fatalf("anos ordenados descendentemente incorretamente: %+v", resp.Data.Anos)
		}
	})

	t.Run("10_DetalheJogo_MEMOR96", func(t *testing.T) {
		usuarioA := criarUsuarioTeste(t, env.pool)
		tokenA := gerarAccessTokenTeste(t, env.secret, usuarioA)
		usuarioB := criarUsuarioTeste(t, env.pool)
		tokenB := gerarAccessTokenTeste(t, env.secret, usuarioB)

		fazerCriar := func(token string, nome string, review string, igdbDescricao string) repository.JogoZerado {
			input := criarJogoInput{
				Nome:          nome,
				Console:       "SNES",
				FinalizadoEm:  "2026-05-10T14:30:00Z",
				TempoJogado:   func(v int32) *int32 { return &v }(3600),
				Nota:          10,
				Dificuldade:   "A",
				Review:        review,
				IgdbDescricao: igdbDescricao,
			}
			body, _ := json.Marshal(input)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			if w.Code != http.StatusCreated {
				t.Fatalf("esperava 201 ao criar jogo, obteve %d: %s", w.Code, w.Body.String())
			}
			var resp jogoEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			return resp.Data
		}

		fazerGetDetalhe := func(token string, id int32) (*http.Response, repository.JogoZerado, erroEnvelope) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/jogos/%d", id), nil)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			var resp jogoEnvelope
			var errResp erroEnvelope
			if w.Code == http.StatusOK {
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
			} else {
				_ = json.Unmarshal(w.Body.Bytes(), &errResp)
			}
			return w.Result(), resp.Data, errResp
		}

		fazerExcluir := func(token string, id int32) int {
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", id), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			return w.Code
		}

		jogoA1 := fazerCriar(tokenA, "Jogo A1", "Review A1 completo", "Descricao IGDB A1 completa")
		time.Sleep(20 * time.Millisecond)
		jogoB1 := fazerCriar(tokenB, "Jogo B1", "Review B1", "Descricao IGDB B1")
		time.Sleep(20 * time.Millisecond)
		jogoA2 := fazerCriar(tokenA, "Jogo A2", "Review A2", "Descricao IGDB A2")
		time.Sleep(20 * time.Millisecond)
		jogoA3 := fazerCriar(tokenA, "Jogo A3", "Review A3", "Descricao IGDB A3")

		resA1, detalheA1, _ := fazerGetDetalhe(tokenA, jogoA1.ID)
		if resA1.StatusCode != http.StatusOK {
			t.Fatalf("esperava 200 para A1, obteve %d", resA1.StatusCode)
		}
		if detalheA1.IgdbDescricao != "Descricao IGDB A1 completa" || detalheA1.Review != "Review A1 completo" {
			t.Fatalf("campos completos esperados para A1, obteve igdb_descricao=%q review=%q", detalheA1.IgdbDescricao, detalheA1.Review)
		}

		if detalheA1.Numero != 1 {
			t.Fatalf("esperava numero 1 para A1, obteve %d", detalheA1.Numero)
		}

		resA2, detalheA2, _ := fazerGetDetalhe(tokenA, jogoA2.ID)
		if resA2.StatusCode != http.StatusOK || detalheA2.Numero != 2 {
			t.Fatalf("esperava numero 2 para A2, obteve %d", detalheA2.Numero)
		}

		resA3, detalheA3, _ := fazerGetDetalhe(tokenA, jogoA3.ID)
		if resA3.StatusCode != http.StatusOK || detalheA3.Numero != 3 {
			t.Fatalf("esperava numero 3 para A3, obteve %d", detalheA3.Numero)
		}

		resB1, detalheB1, _ := fazerGetDetalhe(tokenB, jogoB1.ID)
		if resB1.StatusCode != http.StatusOK || detalheB1.Numero != 1 {
			t.Fatalf("esperava numero 1 para B1, obteve %d", detalheB1.Numero)
		}

		statusExclusao := fazerExcluir(tokenA, jogoA2.ID)
		if statusExclusao != http.StatusNoContent {
			t.Fatalf("esperava 204 ao excluir A2, obteve %d", statusExclusao)
		}

		resA3AposExclusao, detalheA3AposExclusao, _ := fazerGetDetalhe(tokenA, jogoA3.ID)
		if resA3AposExclusao.StatusCode != http.StatusOK || detalheA3AposExclusao.Numero != 3 {
			t.Fatalf("esperava que A3 continuasse com numero 3 apos exclusao de A2, obteve %d", detalheA3AposExclusao.Numero)
		}

		resA1AposExclusao, detalheA1AposExclusao, _ := fazerGetDetalhe(tokenA, jogoA1.ID)
		if resA1AposExclusao.StatusCode != http.StatusOK || detalheA1AposExclusao.Numero != 1 {
			t.Fatalf("esperava que A1 continuasse com numero 1, obteve %d", detalheA1AposExclusao.Numero)
		}

		resA2Excluido, _, errA2 := fazerGetDetalhe(tokenA, jogoA2.ID)
		if resA2Excluido.StatusCode != http.StatusNotFound || errA2.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava 404 jogos.not_found para jogo excluido, obteve %d: %s", resA2Excluido.StatusCode, errA2.Error.Codigo)
		}

		resOutroUsuario, _, errOutroUsuario := fazerGetDetalhe(tokenA, jogoB1.ID)
		if resOutroUsuario.StatusCode != http.StatusNotFound || errOutroUsuario.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava 404 jogos.not_found para jogo de outro usuario, obteve %d: %s", resOutroUsuario.StatusCode, errOutroUsuario.Error.Codigo)
		}

		resBtentandoA, _, errBtentandoA := fazerGetDetalhe(tokenB, jogoA1.ID)
		if resBtentandoA.StatusCode != http.StatusNotFound || errBtentandoA.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava 404 jogos.not_found para usuario B tentando ver A1, obteve %d: %s", resBtentandoA.StatusCode, errBtentandoA.Error.Codigo)
		}

		jogoDesc := fazerCriar(tokenA, "Jogo Com Descricao", "Review", "Descricao persistida com sucesso")
		resDesc, detalheDesc, _ := fazerGetDetalhe(tokenA, jogoDesc.ID)
		if resDesc.StatusCode != http.StatusOK || detalheDesc.IgdbDescricao != "Descricao persistida com sucesso" {
			t.Fatalf("esperava igdb_descricao persistida, obteve status %d descricao=%q", resDesc.StatusCode, detalheDesc.IgdbDescricao)
		}
	})
}
