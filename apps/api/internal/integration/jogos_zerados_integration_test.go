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
	"strings"
	"sync"
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

type itemResumoResposta struct {
	Ano        int                    `json:"ano"`
	TotalJogos int64                  `json:"total_jogos"`
	GameDoAno  *repository.JogoZerado `json:"game_do_ano"`
}

type resumoEnvelope struct {
	Data []itemResumoResposta `json:"data"`
}

type definirGameDoAnoEnvelope struct {
	Data struct {
		Ano        int                   `json:"ano"`
		GameDoAno  repository.JogoZerado `json:"game_do_ano"`
		AnteriorID *int32                `json:"anterior_id"`
	} `json:"data"`
}

func criarJogoViaAPI(t *testing.T, env *postgresIntegrationEnv, token string, input criarJogoInput) int32 {
	t.Helper()
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("falha ao criar jogo via API (%d): %s", w.Code, w.Body.String())
	}
	var resp jogoEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("falha ao decodificar resposta do jogo: %v", err)
	}
	return resp.Data.ID
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
	repo := repository.NewJogosRepository(pool, queries)
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
	privadas.PUT("/jogos/:id/game-do-ano", h.DefinirGameDoAno)
	privadas.DELETE("/jogos/:id", h.ExcluirJogo)
	privadas.DELETE("/jogos/:id/game-do-ano", h.RemoverGameDoAno)
	privadas.GET("/jogos", h.ListarJogos)
	privadas.GET("/jogos/filtros", h.ObterFiltros)
	privadas.GET("/jogos/game-do-ano", h.ObterGameDoAnoResumo)
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

	t.Run("18_QA_Busca_Lacunas", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		now := time.Now()
		jZelda, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "The Legend of Zelda: Ocarina of Time", Console: "N64", FinalizadoEm: now, TempoJogado: 1000, Nota: 10, Dificuldade: "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		jMarioUnderscore, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Super_Mario_Bros", Console: "NES", FinalizadoEm: now, TempoJogado: 1000, Nota: 9, Dificuldade: "A",
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

		fazerBusca := func(termo string) listagemEnvelope {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?busca="+termo, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("esperava 200 na busca %q, obteve %d", termo, w.Code)
			}
			var resp listagemEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			return resp
		}

		rMaiusculo := fazerBusca("ZELDA")
		rMinusculo := fazerBusca("zelda")
		rMisto := fazerBusca("Zelda")
		if len(rMaiusculo.Data) != 1 || len(rMinusculo.Data) != 1 || len(rMisto.Data) != 1 {
			t.Fatalf("esperava 1 resultado para todas as variacoes de caixa, obteve ZELDA=%d, zelda=%d, Zelda=%d",
				len(rMaiusculo.Data), len(rMinusculo.Data), len(rMisto.Data))
		}
		if rMaiusculo.Data[0].ID != jZelda.ID || rMinusculo.Data[0].ID != jZelda.ID || rMisto.Data[0].ID != jZelda.ID {
			t.Fatalf("IDs divergiram entre buscas de caixa: %d, %d, %d",
				rMaiusculo.Data[0].ID, rMinusculo.Data[0].ID, rMisto.Data[0].ID)
		}

		rMeio := fazerBusca("ocarina")
		if len(rMeio.Data) != 1 || rMeio.Data[0].ID != jZelda.ID {
			t.Fatalf("busca parcial no meio do nome falhou: %+v", rMeio.Data)
		}

		rEspacos := fazerBusca("%20%20zelda%20%20")
		if len(rEspacos.Data) != 1 || rEspacos.Data[0].ID != jZelda.ID {
			t.Fatalf("busca com espacos nas pontas falhou: %+v", rEspacos.Data)
		}

		rUnderscore := fazerBusca("_")
		if len(rUnderscore.Data) != 1 || rUnderscore.Data[0].ID != jMarioUnderscore.ID {
			t.Fatalf("busca por '_' literal falhou, deveria retornar apenas Super_Mario_Bros: %+v", rUnderscore.Data)
		}

		rInexistente := fazerBusca("TermoQueNaoExisteEmNenhumJogo999")
		if len(rInexistente.Data) != 0 || rInexistente.Meta.Total != 0 || rInexistente.Meta.TotalPaginas != 0 {
			t.Fatalf("termo sem resultado deveria ter data vazia e total 0: %+v", rInexistente)
		}
	})

	t.Run("19_QA_Filtros_Isolados", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		d2024 := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
		d2025Inicio := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		d2025Fim := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
		d2026 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

		j1, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Jogo SNES", Console: "SNES", Genero: "Ação, Aventura", Tipo: "Campanha",
			FinalizadoEm: d2024, TempoJogado: 1000, Nota: 8, Dificuldade: "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		j2, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Jogo Mega", Console: "Mega Drive", Genero: "Plataforma, Ação", Tipo: "100%",
			FinalizadoEm: d2025Inicio, TempoJogado: 1000, Nota: 10, Dificuldade: "AA",
		})
		if err != nil {
			t.Fatal(err)
		}

		j3, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Jogo PS1", Console: "PlayStation", Genero: "RPG, Ficção Científica", Tipo: "DLC",
			FinalizadoEm: d2025Fim, TempoJogado: 1000, Nota: 5, Dificuldade: "B",
		})
		if err != nil {
			t.Fatal(err)
		}

		j4, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Jogo N64", Console: "N64", Genero: "Corrida", Tipo: "Speedrun",
			FinalizadoEm: d2026, TempoJogado: 1000, Nota: 11, Dificuldade: "AAA",
		})
		if err != nil {
			t.Fatal(err)
		}

		consultar := func(queryString string) listagemEnvelope {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?"+queryString, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("filtro %q falhou com status %d: %s", queryString, w.Code, w.Body.String())
			}
			var resp listagemEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			return resp
		}

		rConsole := consultar("console=SNES")
		if len(rConsole.Data) != 1 || rConsole.Data[0].ID != j1.ID {
			t.Fatalf("filtro isolado console falhou: %+v", rConsole.Data)
		}

		rGeneroSemAcento := consultar("genero=ficcao")
		if len(rGeneroSemAcento.Data) != 1 || rGeneroSemAcento.Data[0].ID != j3.ID {
			t.Fatalf("filtro isolado genero sem acento falhou: %+v", rGeneroSemAcento.Data)
		}
		rGeneroMaiusculo := consultar("genero=AVENTURA")
		if len(rGeneroMaiusculo.Data) != 1 || rGeneroMaiusculo.Data[0].ID != j1.ID {
			t.Fatalf("filtro isolado genero maiusculo falhou: %+v", rGeneroMaiusculo.Data)
		}

		rTipo := consultar("tipo=100%25")
		if len(rTipo.Data) != 1 || rTipo.Data[0].ID != j2.ID {
			t.Fatalf("filtro isolado tipo falhou: %+v", rTipo.Data)
		}
		rTipoCase := consultar("tipo=campanha")
		if len(rTipoCase.Data) != 1 || rTipoCase.Data[0].ID != j1.ID {
			t.Fatalf("filtro isolado tipo case insensitive falhou: %+v", rTipoCase.Data)
		}

		rNotaMin := consultar("nota_min=10")
		if len(rNotaMin.Data) != 2 || rNotaMin.Data[0].ID != j4.ID || rNotaMin.Data[1].ID != j2.ID {
			t.Fatalf("filtro isolado nota_min falhou: %+v", rNotaMin.Data)
		}
		rNotaMax := consultar("nota_max=8")
		if len(rNotaMax.Data) != 2 || rNotaMax.Data[0].ID != j3.ID || rNotaMax.Data[1].ID != j1.ID {
			t.Fatalf("filtro isolado nota_max falhou: %+v", rNotaMax.Data)
		}
		rNotaFaixa := consultar("nota_min=8&nota_max=10")
		if len(rNotaFaixa.Data) != 2 || rNotaFaixa.Data[0].ID != j2.ID || rNotaFaixa.Data[1].ID != j1.ID {
			t.Fatalf("filtro isolado faixa de nota incluindo bordas falhou: %+v", rNotaFaixa.Data)
		}

		rAno2024 := consultar("ano=2024")
		if len(rAno2024.Data) != 1 || rAno2024.Data[0].ID != j1.ID {
			t.Fatalf("filtro isolado ano 2024 com 31/12 falhou: %+v", rAno2024.Data)
		}
		rAno2025 := consultar("ano=2025")
		if len(rAno2025.Data) != 2 || rAno2025.Data[0].ID != j3.ID || rAno2025.Data[1].ID != j2.ID {
			t.Fatalf("filtro isolado ano 2025 com bordas 01/01 e 31/12 falhou: %+v", rAno2025.Data)
		}

		rDif := consultar("dificuldade=AAA")
		if len(rDif.Data) != 1 || rDif.Data[0].ID != j4.ID {
			t.Fatalf("filtro isolado dificuldade falhou: %+v", rDif.Data)
		}
	})

	t.Run("20_QA_Combinacoes_Lacunas", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		now := time.Now()
		j1, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Metroid Fusion", Console: "GBA", Genero: "Metroidvania", Tipo: "Campanha",
			FinalizadoEm: now, TempoJogado: 1000, Nota: 9, Dificuldade: "A",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "Castlevania Aria of Sorrow", Console: "GBA", Genero: "Metroidvania", Tipo: "Campanha",
			FinalizadoEm: now, TempoJogado: 1000, Nota: 10, Dificuldade: "AA",
		})
		if err != nil {
			t.Fatal(err)
		}

		consultar := func(queryString string) listagemEnvelope {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?"+queryString, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			var resp listagemEnvelope
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			return resp
		}

		rBuscaEFiltro := consultar("busca=Metroid&console=GBA")
		if len(rBuscaEFiltro.Data) != 1 || rBuscaEFiltro.Data[0].ID != j1.ID {
			t.Fatalf("busca + filtro falhou: %+v", rBuscaEFiltro.Data)
		}

		rDoisFiltros := consultar("console=GBA&dificuldade=A")
		if len(rDoisFiltros.Data) != 1 || rDoisFiltros.Data[0].ID != j1.ID {
			t.Fatalf("dois filtros falhou: %+v", rDoisFiltros.Data)
		}

		rSemIntersecao := consultar("console=GBA&dificuldade=AAA")
		if len(rSemIntersecao.Data) != 0 || rSemIntersecao.Meta.Total != 0 {
			t.Fatalf("combinacao sem intersecao deveria retornar data vazia: %+v", rSemIntersecao)
		}
	})

	t.Run("21_QA_SoftDelete_ComFiltros", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		d2025 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		jAtivo, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "GameBoy Ativo", Console: "GameBoy", FinalizadoEm: d2025, TempoJogado: 1000, Nota: 8, Dificuldade: "B",
		})
		if err != nil {
			t.Fatal(err)
		}

		jExcluido, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "GameBoy Excluido", Console: "GameBoy", FinalizadoEm: d2025, TempoJogado: 1000, Nota: 8, Dificuldade: "B",
		})
		if err != nil {
			t.Fatal(err)
		}

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jExcluido.ID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("status delete: %d", wDel.Code)
		}

		reqListar := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?console=GameBoy&ano=2025", nil)
		reqListar.Header.Set("Authorization", "Bearer "+token)
		wListar := httptest.NewRecorder()
		env.router.ServeHTTP(wListar, reqListar)
		var respListar listagemEnvelope
		_ = json.Unmarshal(wListar.Body.Bytes(), &respListar)
		if len(respListar.Data) != 1 || respListar.Data[0].ID != jAtivo.ID {
			t.Fatalf("registro excluido apareceu em listagem com filtros: %+v", respListar.Data)
		}

		jUnicoVirtualBoy, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID: usuarioID, Nome: "VirtualBoy Unico", Console: "VirtualBoyExclusivo", FinalizadoEm: d2025, TempoJogado: 1000, Nota: 7, Dificuldade: "C",
		})
		if err != nil {
			t.Fatal(err)
		}

		reqDelVB := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jUnicoVirtualBoy.ID), nil)
		reqDelVB.Header.Set("Authorization", "Bearer "+token)
		wDelVB := httptest.NewRecorder()
		env.router.ServeHTTP(wDelVB, reqDelVB)
		if wDelVB.Code != http.StatusNoContent {
			t.Fatalf("status delete VB: %d", wDelVB.Code)
		}

		reqListarVB := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?console=VirtualBoyExclusivo", nil)
		reqListarVB.Header.Set("Authorization", "Bearer "+token)
		wListarVB := httptest.NewRecorder()
		env.router.ServeHTTP(wListarVB, reqListarVB)
		var respListarVB listagemEnvelope
		_ = json.Unmarshal(wListarVB.Body.Bytes(), &respListarVB)
		if len(respListarVB.Data) != 0 || respListarVB.Meta.Total != 0 {
			t.Fatalf("console exclusivo excluido retornou itens: %+v", respListarVB)
		}

		reqFiltros := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/filtros", nil)
		reqFiltros.Header.Set("Authorization", "Bearer "+token)
		wFiltros := httptest.NewRecorder()
		env.router.ServeHTTP(wFiltros, reqFiltros)
		var respFiltros filtrosEnvelope
		_ = json.Unmarshal(wFiltros.Body.Bytes(), &respFiltros)
		for _, c := range respFiltros.Data.Consoles {
			if c == "VirtualBoyExclusivo" {
				t.Fatal("console de registro unico excluido apareceu nos filtros")
			}
		}

		reqDetalhe := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/jogos/%d", jUnicoVirtualBoy.ID), nil)
		reqDetalhe.Header.Set("Authorization", "Bearer "+token)
		wDetalhe := httptest.NewRecorder()
		env.router.ServeHTTP(wDetalhe, reqDetalhe)
		if wDetalhe.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 no detalhe de registro excluido, obteve %d", wDetalhe.Code)
		}
	})

	t.Run("22_QA_Paginacao_Limites_E_FiltroAtivo", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		var idsCriadosSNES []int32
		for i := 1; i <= 5; i++ {
			j, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
				UsuarioID: usuarioID, Nome: fmt.Sprintf("SNES Game %d", i), Console: "SNES",
				FinalizadoEm: baseDate.Add(time.Duration(i) * time.Hour), TempoJogado: 1000, Nota: 8, Dificuldade: "A",
			})
			if err != nil {
				t.Fatal(err)
			}
			idsCriadosSNES = append([]int32{j.ID}, idsCriadosSNES...)
		}

		for i := 1; i <= 3; i++ {
			_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
				UsuarioID: usuarioID, Nome: fmt.Sprintf("Genesis Game %d", i), Console: "Genesis",
				FinalizadoEm: baseDate.Add(time.Duration(i) * time.Hour), TempoJogado: 1000, Nota: 8, Dificuldade: "A",
			})
			if err != nil {
				t.Fatal(err)
			}
		}

		reqP100 := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?por_pagina=100", nil)
		reqP100.Header.Set("Authorization", "Bearer "+token)
		wP100 := httptest.NewRecorder()
		env.router.ServeHTTP(wP100, reqP100)
		if wP100.Code != http.StatusOK {
			t.Fatalf("esperava 200 em por_pagina=100, obteve %d", wP100.Code)
		}

		reqP1 := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?por_pagina=1", nil)
		reqP1.Header.Set("Authorization", "Bearer "+token)
		wP1 := httptest.NewRecorder()
		env.router.ServeHTTP(wP1, reqP1)
		if wP1.Code != http.StatusOK {
			t.Fatalf("esperava 200 em por_pagina=1, obteve %d", wP1.Code)
		}
		var respP1 listagemEnvelope
		_ = json.Unmarshal(wP1.Body.Bytes(), &respP1)
		if len(respP1.Data) != 1 {
			t.Fatalf("esperava 1 item em por_pagina=1, obteve %d", len(respP1.Data))
		}

		reqP0 := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?por_pagina=0", nil)
		reqP0.Header.Set("Authorization", "Bearer "+token)
		wP0 := httptest.NewRecorder()
		env.router.ServeHTTP(wP0, reqP0)
		if wP0.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 em por_pagina=0, obteve %d", wP0.Code)
		}

		reqP101 := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?por_pagina=101", nil)
		reqP101.Header.Set("Authorization", "Bearer "+token)
		wP101 := httptest.NewRecorder()
		env.router.ServeHTTP(wP101, reqP101)
		if wP101.Code != http.StatusBadRequest {
			t.Fatalf("esperava 400 em por_pagina=101, obteve %d", wP101.Code)
		}

		var idsRecebidos []int32
		for pagina := 1; pagina <= 3; pagina++ {
			url := fmt.Sprintf("/api/v1/jogos?console=SNES&por_pagina=2&pagina=%d", pagina)
			reqPag := httptest.NewRequest(http.MethodGet, url, nil)
			reqPag.Header.Set("Authorization", "Bearer "+token)
			wPag := httptest.NewRecorder()
			env.router.ServeHTTP(wPag, reqPag)
			if wPag.Code != http.StatusOK {
				t.Fatalf("esperava 200 na pagina %d com filtro, obteve %d", pagina, wPag.Code)
			}
			var respPag listagemEnvelope
			_ = json.Unmarshal(wPag.Body.Bytes(), &respPag)
			if respPag.Meta.Total != 5 || respPag.Meta.TotalPaginas != 3 {
				t.Fatalf("meta incorreto na pagina %d: %+v", pagina, respPag.Meta)
			}
			for _, jogo := range respPag.Data {
				idsRecebidos = append(idsRecebidos, jogo.ID)
			}
		}

		if len(idsRecebidos) != 5 {
			t.Fatalf("esperava 5 registros paginados com filtro, obteve %d", len(idsRecebidos))
		}
		for i, id := range idsRecebidos {
			if id != idsCriadosSNES[i] {
				t.Fatalf("ordem incorreta ou pulo/repeticao no index %d: esperava %d, obteve %d", i, idsCriadosSNES[i], id)
			}
		}
	})

	t.Run("23_QA_Explain_1000_Registros", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		baseDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		const totalRegistros = 1000
		for i := 1; i <= totalRegistros; i++ {
			_, err := env.pool.Exec(context.Background(),
				`INSERT INTO jogos_zerados (usuario_id, nome, console, genero, tipo, finalizado_em, tempo_jogado, nota, dificuldade)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
				usuarioID, fmt.Sprintf("Jogo Teste %04d", i), "SNES", "RPG", "Campanha",
				baseDate.Add(time.Duration(i)*time.Hour), 3600, 10, "A",
			)
			if err != nil {
				t.Fatalf("falha ao inserir registro %d: %v", i, err)
			}
		}

		_, err := env.pool.Exec(context.Background(), "ANALYZE jogos_zerados;")
		if err != nil {
			t.Fatalf("falha ao executar ANALYZE: %v", err)
		}

		rows, err := env.pool.Query(context.Background(),
			`EXPLAIN (FORMAT TEXT)
			SELECT * FROM jogos_zerados
			WHERE usuario_id = $1
			  AND deleted_at IS NULL
			  AND ($2::text IS NULL OR unaccent(nome) ILIKE unaccent('%' || $2::text || '%'))
			  AND ($3::varchar IS NULL OR console = $3)
			  AND ($4::text IS NULL OR unaccent(genero) ILIKE unaccent('%' || $4::text || '%'))
			  AND ($5::varchar IS NULL OR LOWER(tipo) = LOWER($5))
			  AND ($6::int IS NULL OR nota >= $6)
			  AND ($7::int IS NULL OR nota <= $7)
			  AND ($8::int IS NULL OR EXTRACT(YEAR FROM finalizado_em) = $8)
			  AND ($9::varchar IS NULL OR dificuldade = $9::dificuldade)
			ORDER BY finalizado_em DESC, id DESC
			LIMIT 24 OFFSET 0;`,
			usuarioID, nil, nil, nil, nil, nil, nil, nil, nil,
		)
		if err != nil {
			t.Fatalf("falha ao rodar EXPLAIN: %v", err)
		}
		defer rows.Close()

		var explainOutput strings.Builder
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				t.Fatal(err)
			}
			explainOutput.WriteString(line + "\n")
		}

		plan := explainOutput.String()
		t.Logf("EXPLAIN PLAN:\n%s", plan)
	})

	t.Run("24_GameDoAno_1_ResumoTresAnosComESemDestaque", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo 2026 A", Console: "PS5", FinalizadoEm: "2026-03-01T00:00:00Z",
			Nota: 8, Dificuldade: "B", Destaque: false,
		})
		id2026Destaque := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo 2026 Destaque", Console: "PS5", FinalizadoEm: "2026-06-01T00:00:00Z",
			Nota: 11, Dificuldade: "AAA", Destaque: true,
		})
		criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo 2026 B", Console: "PS5", FinalizadoEm: "2026-09-01T00:00:00Z",
			Nota: 9, Dificuldade: "A", Destaque: false,
		})

		criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo 2025 A", Console: "Switch", FinalizadoEm: "2025-04-01T00:00:00Z",
			Nota: 7, Dificuldade: "C", Destaque: false,
		})
		criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo 2025 B", Console: "Switch", FinalizadoEm: "2025-07-01T00:00:00Z",
			Nota: 8, Dificuldade: "B", Destaque: false,
		})

		id2024Destaque := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo 2024 Destaque", Console: "PC", FinalizadoEm: "2024-11-01T00:00:00Z",
			Nota: 10, Dificuldade: "AA", Destaque: true,
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/game-do-ano", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}

		var resp resumoEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("falha ao decodificar resumo: %v", err)
		}

		if len(resp.Data) != 3 {
			t.Fatalf("esperava 3 anos no resumo, obteve %d", len(resp.Data))
		}

		if resp.Data[0].Ano != 2026 || resp.Data[0].TotalJogos != 3 || resp.Data[0].GameDoAno == nil || resp.Data[0].GameDoAno.ID != id2026Destaque {
			t.Fatalf("item 2026 incorreto: %+v", resp.Data[0])
		}
		if resp.Data[1].Ano != 2025 || resp.Data[1].TotalJogos != 2 || resp.Data[1].GameDoAno != nil {
			t.Fatalf("item 2025 incorreto: %+v", resp.Data[1])
		}
		if resp.Data[2].Ano != 2024 || resp.Data[2].TotalJogos != 1 || resp.Data[2].GameDoAno == nil || resp.Data[2].GameDoAno.ID != id2024Destaque {
			t.Fatalf("item 2024 incorreto: %+v", resp.Data[2])
		}
	})

	t.Run("25_GameDoAno_2_ResumoIgnoraExcluidos", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo Ativo 2026", Console: "PS5", FinalizadoEm: "2026-03-01T00:00:00Z",
			Nota: 8, Dificuldade: "B", Destaque: false,
		})
		idDestaque2026 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo Destaque Excluido 2026", Console: "PS5", FinalizadoEm: "2026-06-01T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})
		reqDel1 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", idDestaque2026), nil)
		reqDel1.Header.Set("Authorization", "Bearer "+token)
		wDel1 := httptest.NewRecorder()
		env.router.ServeHTTP(wDel1, reqDel1)
		if wDel1.Code != http.StatusNoContent {
			t.Fatalf("falha ao excluir jogo 2026: %d", wDel1.Code)
		}

		idExcluido2023 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo Unico 2023 Excluido", Console: "PC", FinalizadoEm: "2023-05-01T00:00:00Z",
			Nota: 9, Dificuldade: "AA", Destaque: true,
		})
		reqDel2 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", idExcluido2023), nil)
		reqDel2.Header.Set("Authorization", "Bearer "+token)
		wDel2 := httptest.NewRecorder()
		env.router.ServeHTTP(wDel2, reqDel2)
		if wDel2.Code != http.StatusNoContent {
			t.Fatalf("falha ao excluir jogo 2023: %d", wDel2.Code)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/game-do-ano", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		var resp resumoEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("falha ao decodificar resumo: %v", err)
		}

		if len(resp.Data) != 1 {
			t.Fatalf("esperava apenas 1 ano no resumo, obteve %d", len(resp.Data))
		}
		if resp.Data[0].Ano != 2026 || resp.Data[0].TotalJogos != 1 || resp.Data[0].GameDoAno != nil {
			t.Fatalf("dados do ano 2026 incorretos: %+v", resp.Data[0])
		}
	})

	t.Run("26_GameDoAno_3_ResumoIsolamentoOutroUsuario", func(t *testing.T) {
		userA := criarUsuarioTeste(t, env.pool)
		userB := criarUsuarioTeste(t, env.pool)
		tokenA := gerarAccessTokenTeste(t, env.secret, userA)
		tokenB := gerarAccessTokenTeste(t, env.secret, userB)

		idDestaqueA := criarJogoViaAPI(t, env, tokenA, criarJogoInput{
			Nome: "Jogo A 2026", Console: "PC", FinalizadoEm: "2026-05-01T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})
		criarJogoViaAPI(t, env, tokenA, criarJogoInput{
			Nome: "Jogo A 2025", Console: "PC", FinalizadoEm: "2025-05-01T00:00:00Z",
			Nota: 8, Dificuldade: "B", Destaque: false,
		})

		criarJogoViaAPI(t, env, tokenB, criarJogoInput{
			Nome: "Jogo B 2026", Console: "PS5", FinalizadoEm: "2026-07-01T00:00:00Z",
			Nota: 11, Dificuldade: "AAA", Destaque: true,
		})
		criarJogoViaAPI(t, env, tokenB, criarJogoInput{
			Nome: "Jogo B 2024", Console: "PS5", FinalizadoEm: "2024-07-01T00:00:00Z",
			Nota: 9, Dificuldade: "A", Destaque: true,
		})

		reqA := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/game-do-ano", nil)
		reqA.Header.Set("Authorization", "Bearer "+tokenA)
		wA := httptest.NewRecorder()
		env.router.ServeHTTP(wA, reqA)

		var respA resumoEnvelope
		if err := json.Unmarshal(wA.Body.Bytes(), &respA); err != nil {
			t.Fatalf("falha ao decodificar: %v", err)
		}

		if len(respA.Data) != 2 {
			t.Fatalf("esperava 2 anos para userA, obteve %d", len(respA.Data))
		}
		if respA.Data[0].Ano != 2026 || respA.Data[0].TotalJogos != 1 || respA.Data[0].GameDoAno == nil || respA.Data[0].GameDoAno.ID != idDestaqueA {
			t.Fatalf("dados do ano 2026 para userA incorretos: %+v", respA.Data[0])
		}
		if respA.Data[1].Ano != 2025 || respA.Data[1].TotalJogos != 1 {
			t.Fatalf("dados do ano 2025 para userA incorretos: %+v", respA.Data[1])
		}
	})

	t.Run("27_Listagem_4_OrdenarPorNotaComEmpateEAno", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j1 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Game 1", Console: "SNES", FinalizadoEm: "2025-05-10T00:00:00Z",
			Nota: 8, Dificuldade: "B",
		})
		j2 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Game 2", Console: "SNES", FinalizadoEm: "2025-06-10T00:00:00Z",
			Nota: 10, Dificuldade: "A",
		})
		j3 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Game 3", Console: "SNES", FinalizadoEm: "2025-06-10T00:00:00Z",
			Nota: 10, Dificuldade: "A",
		})
		j4 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Game 4", Console: "SNES", FinalizadoEm: "2025-04-10T00:00:00Z",
			Nota: 10, Dificuldade: "A",
		})
		j5 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Game 5", Console: "SNES", FinalizadoEm: "2026-01-10T00:00:00Z",
			Nota: 11, Dificuldade: "AAA",
		})

		reqNota := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?ordenar=nota", nil)
		reqNota.Header.Set("Authorization", "Bearer "+token)
		wNota := httptest.NewRecorder()
		env.router.ServeHTTP(wNota, reqNota)

		if wNota.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", wNota.Code)
		}
		var respNota listagemEnvelope
		_ = json.Unmarshal(wNota.Body.Bytes(), &respNota)

		esperadosNota := []int32{j5, j3, j2, j4, j1}
		if len(respNota.Data) != len(esperadosNota) {
			t.Fatalf("esperava %d jogos, obteve %d", len(esperadosNota), len(respNota.Data))
		}
		for i, id := range esperadosNota {
			if respNota.Data[i].ID != id {
				t.Fatalf("posicao %d: esperava id %d, obteve %d", i, id, respNota.Data[i].ID)
			}
		}

		reqAnoNota := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?ano=2025&ordenar=nota", nil)
		reqAnoNota.Header.Set("Authorization", "Bearer "+token)
		wAnoNota := httptest.NewRecorder()
		env.router.ServeHTTP(wAnoNota, reqAnoNota)

		if wAnoNota.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", wAnoNota.Code)
		}
		var respAnoNota listagemEnvelope
		_ = json.Unmarshal(wAnoNota.Body.Bytes(), &respAnoNota)

		esperadosAnoNota := []int32{j3, j2, j4, j1}
		if len(respAnoNota.Data) != len(esperadosAnoNota) {
			t.Fatalf("esperava %d jogos, obteve %d", len(esperadosAnoNota), len(respAnoNota.Data))
		}
		for i, id := range esperadosAnoNota {
			if respAnoNota.Data[i].ID != id {
				t.Fatalf("posicao %d filtrado por ano: esperava id %d, obteve %d", i, id, respAnoNota.Data[i].ID)
			}
		}
	})

	t.Run("28_GameDoAno_5_PutSemDestaqueAnterior", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j1 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo Sem Destaque", Console: "PS4", FinalizadoEm: "2025-08-20T00:00:00Z",
			Nota: 9, Dificuldade: "A", Destaque: false,
		})

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", j1), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}

		var resp definirGameDoAnoEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("falha ao decodificar: %v", err)
		}

		if resp.Data.Ano != 2025 || resp.Data.GameDoAno.ID != j1 || !resp.Data.GameDoAno.Destaque || resp.Data.AnteriorID != nil {
			t.Fatalf("resposta inesperada no PUT sem anterior: %+v", resp.Data)
		}

		var destaqueNoBanco bool
		err := env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j1).Scan(&destaqueNoBanco)
		if err != nil || !destaqueNoBanco {
			t.Fatalf("esperava destaque true no banco, err=%v, val=%v", err, destaqueNoBanco)
		}
	})

	t.Run("29_GameDoAno_6_PutTrocandoDestaqueExistente", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j1 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Destaque Antigo", Console: "NES", FinalizadoEm: "2025-03-10T00:00:00Z",
			Nota: 8, Dificuldade: "B", Destaque: true,
		})
		j2 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Novo Candidato", Console: "NES", FinalizadoEm: "2025-10-10T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: false,
		})

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", j2), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}

		var resp definirGameDoAnoEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("falha ao decodificar: %v", err)
		}

		if resp.Data.Ano != 2025 || resp.Data.GameDoAno.ID != j2 || !resp.Data.GameDoAno.Destaque {
			t.Fatalf("jogo marcado incorreto: %+v", resp.Data)
		}
		if resp.Data.AnteriorID == nil || *resp.Data.AnteriorID != j1 {
			t.Fatalf("esperava anterior_id=%d, obteve %v", j1, resp.Data.AnteriorID)
		}

		var countDestaques int
		err := env.pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM jogos_zerados WHERE usuario_id = $1 AND EXTRACT(YEAR FROM finalizado_em) = 2025 AND destaque = true AND deleted_at IS NULL",
			user,
		).Scan(&countDestaques)
		if err != nil || countDestaques != 1 {
			t.Fatalf("esperava exatamente 1 destaque no ano, obteve count=%d, err=%v", countDestaques, err)
		}

		var j1Destaque, j2Destaque bool
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j1).Scan(&j1Destaque)
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j2).Scan(&j2Destaque)
		if j1Destaque != false || j2Destaque != true {
			t.Fatalf("estados incorretos: j1=%v, j2=%v", j1Destaque, j2Destaque)
		}
	})

	t.Run("30_GameDoAno_7_PutIdempotenteMesmoRegistro", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j1 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Destaque Ja Marcado", Console: "GBA", FinalizadoEm: "2025-05-01T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", j1), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
		}

		var resp definirGameDoAnoEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("falha ao decodificar: %v", err)
		}
		if resp.Data.Ano != 2025 || resp.Data.GameDoAno.ID != j1 || resp.Data.AnteriorID != nil {
			t.Fatalf("esperava anterior_id nil em chamada idempotente, obteve: %+v", resp.Data)
		}

		var j1Destaque bool
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j1).Scan(&j1Destaque)
		if !j1Destaque {
			t.Fatal("jogo deveria permanecer como destaque")
		}
	})

	t.Run("31_GameDoAno_8_PutExcluidoOuOutroUsuario404", func(t *testing.T) {
		userA := criarUsuarioTeste(t, env.pool)
		userB := criarUsuarioTeste(t, env.pool)
		tokenA := gerarAccessTokenTeste(t, env.secret, userA)
		tokenB := gerarAccessTokenTeste(t, env.secret, userB)

		jDestaqueA := criarJogoViaAPI(t, env, tokenA, criarJogoInput{
			Nome: "Destaque Real A", Console: "PC", FinalizadoEm: "2025-01-01T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})
		jExcluidoA := criarJogoViaAPI(t, env, tokenA, criarJogoInput{
			Nome: "A Excluir", Console: "PC", FinalizadoEm: "2025-02-01T00:00:00Z",
			Nota: 8, Dificuldade: "B", Destaque: false,
		})
		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jExcluidoA), nil)
		reqDel.Header.Set("Authorization", "Bearer "+tokenA)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("falha na exclusao: %d", wDel.Code)
		}

		jB := criarJogoViaAPI(t, env, tokenB, criarJogoInput{
			Nome: "Jogo B", Console: "PC", FinalizadoEm: "2025-03-01T00:00:00Z",
			Nota: 9, Dificuldade: "A", Destaque: false,
		})

		reqOutro := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", jB), nil)
		reqOutro.Header.Set("Authorization", "Bearer "+tokenA)
		wOutro := httptest.NewRecorder()
		env.router.ServeHTTP(wOutro, reqOutro)
		if wOutro.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 para jogo de outro usuario, obteve %d", wOutro.Code)
		}

		reqExcluido := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", jExcluidoA), nil)
		reqExcluido.Header.Set("Authorization", "Bearer "+tokenA)
		wExcluido := httptest.NewRecorder()
		env.router.ServeHTTP(wExcluido, reqExcluido)
		if wExcluido.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 para jogo excluido, obteve %d", wExcluido.Code)
		}

		var jDestaqueAinda bool
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", jDestaqueA).Scan(&jDestaqueAinda)
		if !jDestaqueAinda {
			t.Fatal("destaque atual do usuario A nao deveria ter sido alterado")
		}
	})

	t.Run("32_GameDoAno_9_DeleteSucessoEIdempotente", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j1 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo a Desmarcar", Console: "N64", FinalizadoEm: "2025-05-01T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", j1), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)

		if wDel.Code != http.StatusNoContent {
			t.Fatalf("esperava 204 no primeiro DELETE, obteve %d", wDel.Code)
		}

		var noBanco bool
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j1).Scan(&noBanco)
		if noBanco {
			t.Fatal("esperava destaque false no banco apos DELETE")
		}

		reqResumo := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/game-do-ano", nil)
		reqResumo.Header.Set("Authorization", "Bearer "+token)
		wResumo := httptest.NewRecorder()
		env.router.ServeHTTP(wResumo, reqResumo)
		var respResumo resumoEnvelope
		_ = json.Unmarshal(wResumo.Body.Bytes(), &respResumo)
		if len(respResumo.Data) != 1 || respResumo.Data[0].GameDoAno != nil {
			t.Fatalf("esperava game_do_ano null no resumo, obteve: %+v", respResumo.Data)
		}

		reqDel2 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", j1), nil)
		reqDel2.Header.Set("Authorization", "Bearer "+token)
		wDel2 := httptest.NewRecorder()
		env.router.ServeHTTP(wDel2, reqDel2)
		if wDel2.Code != http.StatusNoContent {
			t.Fatalf("esperava 204 no segundo DELETE (idempotente), obteve %d", wDel2.Code)
		}

		reqDel404 := httptest.NewRequest(http.MethodDelete, "/api/v1/jogos/999999/game-do-ano", nil)
		reqDel404.Header.Set("Authorization", "Bearer "+token)
		wDel404 := httptest.NewRecorder()
		env.router.ServeHTTP(wDel404, reqDel404)
		if wDel404.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 para id inexistente no DELETE, obteve %d", wDel404.Code)
		}
	})

	t.Run("33_GameDoAno_10_TrocaNaoAfetaOutrosAnosNemOutrosUsuarios", func(t *testing.T) {
		userA := criarUsuarioTeste(t, env.pool)
		userB := criarUsuarioTeste(t, env.pool)
		tokenA := gerarAccessTokenTeste(t, env.secret, userA)
		tokenB := gerarAccessTokenTeste(t, env.secret, userB)

		j2025A := criarJogoViaAPI(t, env, tokenA, criarJogoInput{
			Nome: "Destaque 2025 A", Console: "PC", FinalizadoEm: "2025-06-01T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})
		j2026A1 := criarJogoViaAPI(t, env, tokenA, criarJogoInput{
			Nome: "Destaque 2026 A1", Console: "PC", FinalizadoEm: "2026-06-01T00:00:00Z",
			Nota: 9, Dificuldade: "B", Destaque: true,
		})
		j2026A2 := criarJogoViaAPI(t, env, tokenA, criarJogoInput{
			Nome: "Destaque 2026 A2", Console: "PC", FinalizadoEm: "2026-07-01T00:00:00Z",
			Nota: 11, Dificuldade: "AAA", Destaque: false,
		})

		j2026B := criarJogoViaAPI(t, env, tokenB, criarJogoInput{
			Nome: "Destaque 2026 B", Console: "PS5", FinalizadoEm: "2026-06-01T00:00:00Z",
			Nota: 10, Dificuldade: "AA", Destaque: true,
		})

		reqTroca := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", j2026A2), nil)
		reqTroca.Header.Set("Authorization", "Bearer "+tokenA)
		wTroca := httptest.NewRecorder()
		env.router.ServeHTTP(wTroca, reqTroca)
		if wTroca.Code != http.StatusOK {
			t.Fatalf("troca esperava 200, obteve %d", wTroca.Code)
		}

		var j2026A2Destaque, j2026A1Destaque, j2025ADestaque, j2026BDestaque bool
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j2026A2).Scan(&j2026A2Destaque)
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j2026A1).Scan(&j2026A1Destaque)
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j2025A).Scan(&j2025ADestaque)
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j2026B).Scan(&j2026BDestaque)

		if !j2026A2Destaque {
			t.Fatal("j2026A2 deveria ser destaque")
		}
		if j2026A1Destaque {
			t.Fatal("j2026A1 nao deveria mais ser destaque")
		}
		if !j2025ADestaque {
			t.Fatal("j2025A de outro ano deveria permanecer como destaque")
		}
		if !j2026BDestaque {
			t.Fatal("j2026B de outro usuario deveria permanecer como destaque")
		}
	})

	t.Run("34_GameDoAno_11_ConcorrenciaMesmoAno", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j1 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Concorrente 1", Console: "PC", FinalizadoEm: "2025-04-10T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: false,
		})
		j2 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Concorrente 2", Console: "PC", FinalizadoEm: "2025-08-20T00:00:00Z",
			Nota: 9, Dificuldade: "B", Destaque: false,
		})

		var wg sync.WaitGroup
		wg.Add(2)
		start := make(chan struct{})
		codigos := make([]int, 2)

		executar := func(idx int, jogoID int32) {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d/game-do-ano", jogoID), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			codigos[idx] = w.Code
		}

		go executar(0, j1)
		go executar(1, j2)

		close(start)
		wg.Wait()

		for i, code := range codigos {
			if code != http.StatusOK && code != http.StatusConflict {
				t.Fatalf("goroutine %d retornou status inesperado %d", i, code)
			}
		}

		var countDestaques int
		err := env.pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM jogos_zerados WHERE usuario_id = $1 AND EXTRACT(YEAR FROM finalizado_em) = 2025 AND destaque = true AND deleted_at IS NULL",
			user,
		).Scan(&countDestaques)
		if err != nil {
			t.Fatalf("falha ao consultar destaques: %v", err)
		}
		if countDestaques != 1 {
			t.Fatalf("esperava exatamente 1 destaque para 2025, obteve %d", countDestaques)
		}
	})

	t.Run("35_GameDoAno_12_MudancaDeAnoEdicaoConflito409", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j2024 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Destaque 2024", Console: "SNES", FinalizadoEm: "2024-05-10T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})
		j2025 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Destaque 2025", Console: "SNES", FinalizadoEm: "2025-07-20T00:00:00Z",
			Nota: 9, Dificuldade: "B", Destaque: true,
		})

		inputEdicao := criarJogoInput{
			Nome:         "Destaque 2025 Tentando Mover Para 2024",
			Console:      "SNES",
			FinalizadoEm: "2024-08-15T00:00:00Z",
			Nota:         9,
			Dificuldade:  "B",
			Destaque:     true,
		}
		body, _ := json.Marshal(inputEdicao)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", j2025), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("esperava 409, obteve %d: %s", w.Code, w.Body.String())
		}

		var errResp erroEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("falha ao decodificar erro: %v", err)
		}
		if errResp.Error.Codigo != "jogos.destaque_ano_conflito" {
			t.Fatalf("esperava jogos.destaque_ano_conflito, obteve %s", errResp.Error.Codigo)
		}

		var j2025Ano int
		var j2025Nome string
		var j2025Destaque bool
		err := env.pool.QueryRow(context.Background(),
			"SELECT EXTRACT(YEAR FROM finalizado_em)::int, nome, destaque FROM jogos_zerados WHERE id = $1",
			j2025,
		).Scan(&j2025Ano, &j2025Nome, &j2025Destaque)
		if err != nil {
			t.Fatalf("falha ao consultar j2025: %v", err)
		}
		if j2025Ano != 2025 || j2025Nome != "Destaque 2025" || !j2025Destaque {
			t.Fatalf("j2025 foi modificado indevidamente: ano=%d, nome=%s, destaque=%v", j2025Ano, j2025Nome, j2025Destaque)
		}

		var j2024Destaque bool
		_ = env.pool.QueryRow(context.Background(), "SELECT destaque FROM jogos_zerados WHERE id = $1", j2024).Scan(&j2024Destaque)
		if !j2024Destaque {
			t.Fatal("j2024 deveria continuar como destaque")
		}
	})

	t.Run("36_GameDoAno_13_MudancaDeAnoEdicaoSucessoSemConflito", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		jDestaque := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Destaque Original 2025", Console: "GBA", FinalizadoEm: "2025-06-01T00:00:00Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})
		criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Outro Jogo 2025", Console: "GBA", FinalizadoEm: "2025-08-01T00:00:00Z",
			Nota: 8, Dificuldade: "B", Destaque: false,
		})

		inputEdicao := criarJogoInput{
			Nome:         "Destaque Original 2025 Movido Para 2024",
			Console:      "GBA",
			FinalizadoEm: "2024-05-10T00:00:00Z",
			Nota:         10,
			Dificuldade:  "A",
			Destaque:     true,
		}
		body, _ := json.Marshal(inputEdicao)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jDestaque), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("esperava 200 na edicao, obteve %d: %s", w.Code, w.Body.String())
		}

		reqResumo := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/game-do-ano", nil)
		reqResumo.Header.Set("Authorization", "Bearer "+token)
		wResumo := httptest.NewRecorder()
		env.router.ServeHTTP(wResumo, reqResumo)

		if wResumo.Code != http.StatusOK {
			t.Fatalf("esperava 200 no resumo, obteve %d", wResumo.Code)
		}
		var respResumo resumoEnvelope
		if err := json.Unmarshal(wResumo.Body.Bytes(), &respResumo); err != nil {
			t.Fatalf("falha ao decodificar resumo: %v", err)
		}

		if len(respResumo.Data) != 2 {
			t.Fatalf("esperava 2 anos no resumo, obteve %d", len(respResumo.Data))
		}

		if respResumo.Data[0].Ano != 2025 || respResumo.Data[0].TotalJogos != 1 || respResumo.Data[0].GameDoAno != nil {
			t.Fatalf("ano 2025 deveria ter total 1 e game_do_ano null, obteve: %+v", respResumo.Data[0])
		}
		if respResumo.Data[1].Ano != 2024 || respResumo.Data[1].TotalJogos != 1 || respResumo.Data[1].GameDoAno == nil || respResumo.Data[1].GameDoAno.ID != jDestaque {
			t.Fatalf("ano 2024 deveria ter total 1 e game_do_ano id %d, obteve: %+v", jDestaque, respResumo.Data[1])
		}
	})

	t.Run("37_GameDoAno_14_BordasDeAnoVirada", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j2024 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Fim de 2024", Console: "PS5", FinalizadoEm: "2024-12-31T23:59:59Z",
			Nota: 10, Dificuldade: "A", Destaque: true,
		})
		j2025 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Inicio de 2025", Console: "PS5", FinalizadoEm: "2025-01-01T00:00:00Z",
			Nota: 11, Dificuldade: "AAA", Destaque: true,
		})

		reqResumo := httptest.NewRequest(http.MethodGet, "/api/v1/jogos/game-do-ano", nil)
		reqResumo.Header.Set("Authorization", "Bearer "+token)
		wResumo := httptest.NewRecorder()
		env.router.ServeHTTP(wResumo, reqResumo)

		if wResumo.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", wResumo.Code)
		}
		var resp resumoEnvelope
		if err := json.Unmarshal(wResumo.Body.Bytes(), &resp); err != nil {
			t.Fatalf("falha ao decodificar resumo: %v", err)
		}

		if len(resp.Data) != 2 {
			t.Fatalf("esperava 2 anos no resumo, obteve %d", len(resp.Data))
		}
		if resp.Data[0].Ano != 2025 || resp.Data[0].GameDoAno == nil || resp.Data[0].GameDoAno.ID != j2025 {
			t.Fatalf("ano 2025 incorreto: %+v", resp.Data[0])
		}
		if resp.Data[1].Ano != 2024 || resp.Data[1].GameDoAno == nil || resp.Data[1].GameDoAno.ID != j2024 {
			t.Fatalf("ano 2024 incorreto: %+v", resp.Data[1])
		}
	})

	t.Run("38_GamesDaVida_1_ListagemNota11Edicao", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		jA := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo Nota 11 Original", Console: "PS4", FinalizadoEm: "2025-02-10T00:00:00Z",
			Nota: 11, Dificuldade: "AAA",
		})
		jB := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Jogo Nota 10 Original", Console: "PS4", FinalizadoEm: "2025-03-10T00:00:00Z",
			Nota: 10, Dificuldade: "A",
		})

		consultarNota11 := func() []repository.JogoZerado {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?nota_min=11&por_pagina=100", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("esperava 200, obteve %d", w.Code)
			}
			var resp listagemEnvelope
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("falha ao decodificar: %v", err)
			}
			return resp.Data
		}

		res1 := consultarNota11()
		if len(res1) != 1 || res1[0].ID != jA {
			t.Fatalf("inicialmente esperava apenas jA na lista de nota 11, obteve: %+v", res1)
		}

		inputEditA := criarJogoInput{
			Nome:         "Jogo Nota 11 Original",
			Console:      "PS4",
			FinalizadoEm: "2025-02-10T00:00:00Z",
			Nota:         10,
			Dificuldade:  "AAA",
		}
		bodyA, _ := json.Marshal(inputEditA)
		reqEditA := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jA), bytes.NewReader(bodyA))
		reqEditA.Header.Set("Authorization", "Bearer "+token)
		reqEditA.Header.Set("Content-Type", "application/json")
		wEditA := httptest.NewRecorder()
		env.router.ServeHTTP(wEditA, reqEditA)
		if wEditA.Code != http.StatusOK {
			t.Fatalf("falha ao editar jA: %d", wEditA.Code)
		}

		res2 := consultarNota11()
		if len(res2) != 0 {
			t.Fatalf("esperava lista vazia apos jA passar para nota 10, obteve: %+v", res2)
		}

		inputEditB := criarJogoInput{
			Nome:         "Jogo Nota 10 Original",
			Console:      "PS4",
			FinalizadoEm: "2025-03-10T00:00:00Z",
			Nota:         11,
			Dificuldade:  "A",
		}
		bodyB, _ := json.Marshal(inputEditB)
		reqEditB := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jB), bytes.NewReader(bodyB))
		reqEditB.Header.Set("Authorization", "Bearer "+token)
		reqEditB.Header.Set("Content-Type", "application/json")
		wEditB := httptest.NewRecorder()
		env.router.ServeHTTP(wEditB, reqEditB)
		if wEditB.Code != http.StatusOK {
			t.Fatalf("falha ao editar jB: %d", wEditB.Code)
		}

		res3 := consultarNota11()
		if len(res3) != 1 || res3[0].ID != jB {
			t.Fatalf("esperava jB na lista apos passar para nota 11, obteve: %+v", res3)
		}
	})

	t.Run("39_GamesDaVida_2_ExcluidoNota11NaoAparece", func(t *testing.T) {
		user := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, user)

		j11 := criarJogoViaAPI(t, env, token, criarJogoInput{
			Nome: "Obra Prima a Excluir", Console: "Switch", FinalizadoEm: "2025-04-01T00:00:00Z",
			Nota: 11, Dificuldade: "A",
		})

		reqAntes := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?nota_min=11", nil)
		reqAntes.Header.Set("Authorization", "Bearer "+token)
		wAntes := httptest.NewRecorder()
		env.router.ServeHTTP(wAntes, reqAntes)
		var respAntes listagemEnvelope
		_ = json.Unmarshal(wAntes.Body.Bytes(), &respAntes)
		if len(respAntes.Data) != 1 || respAntes.Data[0].ID != j11 {
			t.Fatalf("esperava j11 presente antes da exclusao, obteve: %+v", respAntes.Data)
		}

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", j11), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("esperava 204 no delete, obteve %d", wDel.Code)
		}

		reqDepois := httptest.NewRequest(http.MethodGet, "/api/v1/jogos?nota_min=11", nil)
		reqDepois.Header.Set("Authorization", "Bearer "+token)
		wDepois := httptest.NewRecorder()
		env.router.ServeHTTP(wDepois, reqDepois)
		if wDepois.Code != http.StatusOK {
			t.Fatalf("esperava 200, obteve %d", wDepois.Code)
		}
		var respDepois listagemEnvelope
		_ = json.Unmarshal(wDepois.Body.Bytes(), &respDepois)
		if len(respDepois.Data) != 0 || respDepois.Meta.Total != 0 {
			t.Fatalf("jogo excluido com nota 11 nao deveria aparecer na listagem nota_min=11: %+v", respDepois)
		}
	})
}

